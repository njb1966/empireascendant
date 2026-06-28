package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

var ErrMalformedPvP = errors.New("malformed pvp payload")

func LocalEmpireIDFromGlobal(nodeID, globalID string) (string, bool) {
	prefix := nodeID + ":p_"
	if nodeID == "" || !strings.HasPrefix(globalID, prefix) {
		return "", false
	}
	id := strings.TrimPrefix(globalID, prefix)
	return id, id != ""
}

func (s *Store) RecordOutboundPvP(ctx context.Context, requestID, attackerEmpireID, attackerGlobalID, victimGlobalID, action string, now time.Time) error {
	if requestID == "" || attackerEmpireID == "" || attackerGlobalID == "" || victimGlobalID == "" || action == "" {
		return ErrMalformedPvP
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO pvp_outbound
		(request_id, attacker_empire_id, attacker_global_id, victim_global_id, action, status, created_at)
		VALUES (?, ?, ?, ?, ?, 'queued', ?)`,
		requestID, attackerEmpireID, attackerGlobalID, victimGlobalID, action, now.Unix())
	return err
}

func (s *Store) ResolveInboundPvP(ctx context.Context, nodeID string, pending interdoor.PvPPending, now time.Time) (interdoor.Event, bool, error) {
	if pending.RequestID == "" || pending.AttackerID == "" || pending.VictimID == "" || len(pending.Attacker) == 0 {
		return interdoor.Event{}, false, ErrMalformedPvP
	}
	if already, err := s.inboundPvPProcessed(ctx, pending.RequestID); err != nil {
		return interdoor.Event{}, false, err
	} else if already {
		return interdoor.Event{}, true, nil
	}

	var payload game.RemoteAttackPayload
	if err := json.Unmarshal(pending.Attacker, &payload); err != nil {
		return interdoor.Event{}, false, fmt.Errorf("%w: %v", ErrMalformedPvP, err)
	}
	if payload.AttackerGlobalID != pending.AttackerID || payload.VictimGlobalID != pending.VictimID ||
		payload.AttackerEmpireName == "" || payload.Action == "" {
		return interdoor.Event{}, false, ErrMalformedPvP
	}
	victimID, ok := LocalEmpireIDFromGlobal(nodeID, pending.VictimID)
	if !ok {
		return interdoor.Event{}, false, ErrMalformedPvP
	}

	victimEmpire, err := s.FindEmpireByID(ctx, victimID)
	if err != nil {
		return interdoor.Event{}, false, err
	}
	victim, err := s.LoadEconomy(ctx, victimEmpire)
	if err != nil {
		return interdoor.Event{}, false, err
	}

	result, err := resolvePendingAttack(&victim, payload)
	if err != nil {
		if errors.Is(err, game.ErrInvalidAction) || errors.Is(err, game.ErrInvalidUnit) {
			return interdoor.Event{}, false, fmt.Errorf("%w: %v", ErrMalformedPvP, err)
		}
		return interdoor.Event{}, false, err
	}
	winnerGlobalID := pending.VictimID
	if result.AttackerWon {
		winnerGlobalID = pending.AttackerID
	}

	if now.IsZero() {
		now = time.Now()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return interdoor.Event{}, false, err
	}
	defer tx.Rollback()
	if already, err := inboundPvPProcessedTx(ctx, tx, pending.RequestID); err != nil {
		return interdoor.Event{}, false, err
	} else if already {
		return interdoor.Event{}, true, nil
	}
	if err := saveCombatState(ctx, tx, victim); err != nil {
		return interdoor.Event{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO dispatches (id, empire_id, kind, message)
		VALUES (?, ?, ?, ?)`, newID(), victim.Empire.ID, "pvp", result.Message); err != nil {
		return interdoor.Event{}, false, err
	}
	event, err := appendLocalEvent(ctx, tx, nodeID, "pvp.resolved", map[string]any{
		"request_id":           pending.RequestID,
		"attacker_global_id":   pending.AttackerID,
		"victim_global_id":     pending.VictimID,
		"winner_global_id":     winnerGlobalID,
		"result_text":          result.Message,
		"resolved_at":          now.Unix(),
		"action":               payload.Action,
		"missile_type":         payload.MissileType,
		"attacker_empire_name": payload.AttackerEmpireName,
		"victim_empire_name":   victim.Empire.EmpireName,
		"attacker_won":         result.AttackerWon,
		"loot":                 result.Loot,
		"attacker_casualties":  result.AttackerCasualties,
		"defender_casualties":  result.DefenderCasualties,
	}, now.Unix())
	if err != nil {
		return interdoor.Event{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO pvp_inbound_processed (request_id, event_id, processed_at)
		VALUES (?, ?, ?)`, pending.RequestID, event.EventID, now.Unix()); err != nil {
		return interdoor.Event{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return interdoor.Event{}, false, err
	}
	return event, true, nil
}

func (s *Store) ApplyPvPResolvedEvents(ctx context.Context, nodeID string, events []interdoor.FeedEvent, now time.Time) (int, error) {
	if now.IsZero() {
		now = time.Now()
	}
	applied := 0
	for _, event := range events {
		if event.Type != "pvp.resolved" {
			continue
		}
		changed, err := s.applyPvPResolvedEvent(ctx, nodeID, event, now)
		if err != nil {
			return applied, err
		}
		if changed {
			applied++
		}
	}
	return applied, nil
}

func (s *Store) applyPvPResolvedEvent(ctx context.Context, nodeID string, event interdoor.FeedEvent, now time.Time) (bool, error) {
	var payload struct {
		RequestID          string `json:"request_id"`
		AttackerGlobalID   string `json:"attacker_global_id"`
		WinnerGlobalID     string `json:"winner_global_id"`
		ResultText         string `json:"result_text"`
		Action             string `json:"action"`
		Loot               int    `json:"loot"`
		AttackerCasualties int    `json:"attacker_casualties"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return false, err
	}
	attackerID, ok := LocalEmpireIDFromGlobal(nodeID, payload.AttackerGlobalID)
	if !ok || payload.RequestID == "" {
		return false, nil
	}
	var outboundStatus string
	err := s.db.QueryRowContext(ctx, `SELECT status FROM pvp_outbound
		WHERE request_id = ? AND attacker_global_id = ?`, payload.RequestID, payload.AttackerGlobalID).Scan(&outboundStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if outboundStatus == "resolved" {
		return false, nil
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pvp_results_applied WHERE event_id = ?`, event.EventID).Scan(&exists); err != nil {
		return false, err
	}
	if exists > 0 {
		return false, nil
	}

	empire, err := s.FindEmpireByID(ctx, attackerID)
	if err != nil {
		return false, err
	}
	state, err := s.LoadEconomy(ctx, empire)
	if err != nil {
		return false, err
	}
	if payload.AttackerCasualties > state.Military.NormalSoldiers {
		state.Military.NormalSoldiers = 0
	} else {
		state.Military.NormalSoldiers -= payload.AttackerCasualties
	}
	if payload.Action == game.ActionGroundAttack && payload.WinnerGlobalID == payload.AttackerGlobalID {
		state.Empire.Money += payload.Loot
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if err := saveCombatState(ctx, tx, state); err != nil {
		return false, err
	}
	message := payload.ResultText
	if message == "" {
		message = "Cross-node battle resolved."
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO dispatches (id, empire_id, kind, message)
		VALUES (?, ?, ?, ?)`, newID(), state.Empire.ID, "pvp", message); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pvp_outbound SET status = 'resolved', resolved_at = ?
		WHERE request_id = ?`, now.Unix(), payload.RequestID); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO pvp_results_applied (event_id, request_id, applied_at)
		VALUES (?, ?, ?)`, event.EventID, payload.RequestID, now.Unix()); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func resolvePendingAttack(victim *game.EconomyState, payload game.RemoteAttackPayload) (game.CombatResult, error) {
	switch payload.Action {
	case game.ActionGroundAttack:
		attacker := game.Combatant{
			Empire: game.Empire{
				ID:         payload.AttackerGlobalID,
				WorldName:  payload.AttackerWorldName,
				EmpireName: payload.AttackerEmpireName,
			},
			Military: payload.Military.Military(payload.AttackerGlobalID),
		}
		defender := game.Combatant{Empire: victim.Empire, Military: victim.Military}
		result := game.ResolveGroundAttack(&attacker, &defender, game.FixedDefenseMultiplier(1.0))
		victim.Empire = defender.Empire
		victim.Military = defender.Military
		return result, nil
	case game.ActionMissile:
		return game.ResolveRemoteMissileStrike(payload.AttackerEmpireName, victim, payload.MissileType)
	default:
		return game.CombatResult{}, game.ErrInvalidAction
	}
}

func (s *Store) inboundPvPProcessed(ctx context.Context, requestID string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pvp_inbound_processed WHERE request_id = ?`, requestID).Scan(&exists)
	return exists > 0, err
}

func inboundPvPProcessedTx(ctx context.Context, tx *sql.Tx, requestID string) (bool, error) {
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pvp_inbound_processed WHERE request_id = ?`, requestID).Scan(&exists)
	return exists > 0, err
}

func saveCombatState(ctx context.Context, execer sqlExecer, state game.EconomyState) error {
	if _, err := execer.ExecContext(ctx, `UPDATE empires SET
		turns_left = ?, turn_day = ?, money = ?, money_bank = ?, population = ?,
		food = ?, food_storage = ?, energy = ?, research_pts = ?, building_pts = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		state.Empire.TurnsLeft, state.Empire.TurnDay, state.Empire.Money, state.Empire.MoneyBank,
		state.Empire.Population, state.Empire.Food, state.Empire.FoodStorage, state.Empire.Energy,
		state.Empire.ResearchPts, state.Empire.BuildingPts, state.Empire.ID); err != nil {
		return err
	}
	return saveMilitary(ctx, execer, state.Military)
}
