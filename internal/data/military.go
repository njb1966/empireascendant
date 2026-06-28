package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"empireascendant/internal/game"
)

func (s *Store) ListTargetEmpires(ctx context.Context, attackerID string) ([]game.Empire, error) {
	if err := s.RefreshAllLifecycle(ctx, time.Now()); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT
		e.id, e.username, e.password_hash, e.world_name, e.world_name_key, e.empire_name, e.empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
		FROM empires e
		LEFT JOIN empire_lifecycle l ON l.empire_id = e.id
		WHERE e.id <> ?
		AND COALESCE(l.status, ?) NOT IN (?, ?)
		ORDER BY e.empire_name`,
		attackerID, game.LifecycleActive, game.LifecycleHidden, game.LifecyclePurgeEligible)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var empires []game.Empire
	for rows.Next() {
		e, err := scanEmpire(rows)
		if err != nil {
			return nil, err
		}
		empires = append(empires, e)
	}
	return empires, rows.Err()
}

func (s *Store) FindEmpireByID(ctx context.Context, id string) (game.Empire, error) {
	row := s.db.QueryRowContext(ctx, `SELECT
		id, username, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
		FROM empires WHERE id = ?`, id)
	return scanEmpire(row)
}

func (s *Store) CheckAttackLimit(ctx context.Context, attackerID, defenderID, action string, turnDay int) error {
	limit, err := game.AttackLimitForAction(action)
	if err != nil {
		return err
	}
	count, err := s.attackCount(ctx, attackerID, defenderID, action, turnDay)
	if err != nil {
		return err
	}
	if count >= limit {
		return game.ErrAttackLimit
	}
	return nil
}

func (s *Store) RecordAttack(ctx context.Context, attackerID, defenderID, action string, turnDay int) error {
	if err := s.CheckAttackLimit(ctx, attackerID, defenderID, action, turnDay); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO attack_limits (attacker_id, defender_id, action, turn_day, count)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(attacker_id, defender_id, action, turn_day) DO UPDATE SET count = count + 1`,
		attackerID, defenderID, action, turnDay)
	return err
}

func (s *Store) attackCount(ctx context.Context, attackerID, defenderID, action string, turnDay int) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count FROM attack_limits
		WHERE attacker_id = ? AND defender_id = ? AND action = ? AND turn_day = ?`,
		attackerID, defenderID, action, turnDay).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) AddDispatch(ctx context.Context, empireID, kind, message string) (game.Dispatch, error) {
	d := game.Dispatch{
		ID:       newID(),
		EmpireID: empireID,
		Kind:     kind,
		Message:  message,
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO dispatches (id, empire_id, kind, message)
		VALUES (?, ?, ?, ?)`, d.ID, d.EmpireID, d.Kind, d.Message)
	if err != nil {
		return game.Dispatch{}, fmt.Errorf("add dispatch: %w", err)
	}
	return d, nil
}

func (s *Store) ListDispatches(ctx context.Context, empireID string, limit int) ([]game.Dispatch, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, empire_id, kind, message, created_at, read_at
		FROM dispatches WHERE empire_id = ? ORDER BY created_at DESC LIMIT ?`, empireID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dispatches []game.Dispatch
	for rows.Next() {
		var d game.Dispatch
		if err := rows.Scan(&d.ID, &d.EmpireID, &d.Kind, &d.Message, &d.CreatedAt, &d.ReadAt); err != nil {
			return nil, err
		}
		dispatches = append(dispatches, d)
	}
	return dispatches, rows.Err()
}
