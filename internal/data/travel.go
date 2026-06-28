package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

const (
	TravelStatusActive    = "active"
	TravelStatusTraveling = "traveling"
	TravelStatusAway      = "away"
	TravelStatusVisiting  = "visiting"
)

var ErrMalformedTravel = errors.New("malformed travel snapshot")

type TravelStatus struct {
	EmpireID    string
	GlobalID    string
	HomeNode    string
	CurrentNode string
	DestNode    string
	Status      string
	TravelID    string
	UpdatedAt   int64
}

type TravelSnapshotExport struct {
	GlobalID string
	HomeNode string
	Snapshot json.RawMessage
}

func (s *Store) ExportTravelSnapshot(ctx context.Context, nodeID, empireID string) (TravelSnapshotExport, error) {
	if nodeID == "" || empireID == "" {
		return TravelSnapshotExport{}, ErrMalformedTravel
	}
	empire, err := s.FindEmpireByID(ctx, empireID)
	if err != nil {
		return TravelSnapshotExport{}, err
	}
	state, err := s.LoadEconomy(ctx, empire)
	if err != nil {
		return TravelSnapshotExport{}, err
	}
	globalID := GlobalEmpireID(nodeID, empireID)
	snapshot := game.NewTravelSnapshot(state, globalID, nodeID, nodeID)
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return TravelSnapshotExport{}, err
	}
	return TravelSnapshotExport{GlobalID: globalID, HomeNode: nodeID, Snapshot: raw}, nil
}

func (s *Store) MarkTravelSubmitted(ctx context.Context, empireID, globalID, homeNode, destNode, travelID string, now time.Time) error {
	if empireID == "" || globalID == "" || homeNode == "" || destNode == "" || travelID == "" {
		return ErrMalformedTravel
	}
	if now.IsZero() {
		now = time.Now()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO travel_state
		(empire_id, global_id, home_node, current_node, dest_node, status, travel_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(empire_id) DO UPDATE SET
			global_id = excluded.global_id,
			home_node = excluded.home_node,
			current_node = excluded.current_node,
			dest_node = excluded.dest_node,
			status = excluded.status,
			travel_id = excluded.travel_id,
			updated_at = excluded.updated_at`,
		empireID, globalID, homeNode, homeNode, destNode, TravelStatusTraveling, travelID, now.Unix())
	return err
}

func (s *Store) EmpireTravelStatus(ctx context.Context, empireID string) (TravelStatus, error) {
	var status TravelStatus
	err := s.db.QueryRowContext(ctx, `SELECT empire_id, global_id, home_node, current_node, dest_node, status, travel_id, updated_at
		FROM travel_state WHERE empire_id = ?`, empireID).Scan(
		&status.EmpireID, &status.GlobalID, &status.HomeNode, &status.CurrentNode, &status.DestNode,
		&status.Status, &status.TravelID, &status.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TravelStatus{EmpireID: empireID, Status: TravelStatusActive}, nil
	}
	return status, err
}

func (s *Store) ImportTravelArrival(ctx context.Context, nodeID string, pending interdoor.TravelPending, now time.Time) (interdoor.Event, bool, error) {
	if pending.TravelID == "" || pending.GlobalID == "" || pending.HomeNode == "" || pending.FromNode == "" ||
		pending.DestNode == "" || len(pending.Snapshot) == 0 {
		return interdoor.Event{}, false, ErrMalformedTravel
	}
	if pending.DestNode != nodeID {
		return interdoor.Event{}, false, ErrMalformedTravel
	}
	if already, err := s.travelArrivalProcessed(ctx, pending.TravelID); err != nil {
		return interdoor.Event{}, false, err
	} else if already {
		return interdoor.Event{}, true, nil
	}

	var snapshot game.TravelSnapshot
	if err := json.Unmarshal(pending.Snapshot, &snapshot); err != nil {
		return interdoor.Event{}, false, fmt.Errorf("%w: %v", ErrMalformedTravel, err)
	}
	if err := validateTravelSnapshot(snapshot, pending); err != nil {
		return interdoor.Event{}, false, err
	}
	if now.IsZero() {
		now = time.Now()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return interdoor.Event{}, false, err
	}
	defer tx.Rollback()
	if already, err := travelArrivalProcessedTx(ctx, tx, pending.TravelID); err != nil {
		return interdoor.Event{}, false, err
	} else if already {
		return interdoor.Event{}, true, nil
	}

	state := snapshot.EconomyState()
	if pending.HomeNode == nodeID {
		if err := upsertFullEconomyState(ctx, tx, state); err != nil {
			return interdoor.Event{}, false, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO travel_state
			(empire_id, global_id, home_node, current_node, dest_node, status, travel_id, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(empire_id) DO UPDATE SET
				global_id = excluded.global_id,
				home_node = excluded.home_node,
				current_node = excluded.current_node,
				dest_node = excluded.dest_node,
				status = excluded.status,
				travel_id = excluded.travel_id,
				updated_at = excluded.updated_at`,
			state.Empire.ID, pending.GlobalID, pending.HomeNode, nodeID, "", TravelStatusActive, "", now.Unix()); err != nil {
			return interdoor.Event{}, false, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `INSERT INTO travel_visitors
			(global_id, home_node, current_node, from_node, empire_name, world_name, snapshot, status, arrived_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(global_id) DO UPDATE SET
				home_node = excluded.home_node,
				current_node = excluded.current_node,
				from_node = excluded.from_node,
				empire_name = excluded.empire_name,
				world_name = excluded.world_name,
				snapshot = excluded.snapshot,
				status = excluded.status,
				arrived_at = excluded.arrived_at,
				updated_at = excluded.updated_at`,
			pending.GlobalID, pending.HomeNode, nodeID, pending.FromNode, snapshot.Empire.EmpireName,
			snapshot.Empire.WorldName, string(pending.Snapshot), TravelStatusVisiting, now.Unix(), now.Unix()); err != nil {
			return interdoor.Event{}, false, err
		}
	}

	hash := sha256.Sum256(pending.Snapshot)
	event, err := appendLocalEvent(ctx, tx, nodeID, "player.traveled", map[string]any{
		"global_id":     pending.GlobalID,
		"src_node":      pending.FromNode,
		"dest_node":     pending.DestNode,
		"snapshot_hash": hex.EncodeToString(hash[:]),
		"timestamp":     now.Unix(),
		"empire_name":   snapshot.Empire.EmpireName,
		"world_name":    snapshot.Empire.WorldName,
	}, now.Unix())
	if err != nil {
		return interdoor.Event{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO travel_arrivals_processed (travel_id, global_id, event_id, processed_at)
		VALUES (?, ?, ?, ?)`, pending.TravelID, pending.GlobalID, event.EventID, now.Unix()); err != nil {
		return interdoor.Event{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return interdoor.Event{}, false, err
	}
	return event, true, nil
}

func validateTravelSnapshot(snapshot game.TravelSnapshot, pending interdoor.TravelPending) error {
	if snapshot.Version != game.TravelSnapshotVersion || snapshot.GlobalID != pending.GlobalID ||
		snapshot.HomeNode != pending.HomeNode || snapshot.FromNode != pending.FromNode ||
		snapshot.Empire.ID == "" || snapshot.Empire.EmpireName == "" || snapshot.Empire.WorldName == "" ||
		len(snapshot.Regions) == 0 || len(snapshot.Mines) == 0 || len(snapshot.Tech) == 0 {
		return ErrMalformedTravel
	}
	return nil
}

func (s *Store) travelArrivalProcessed(ctx context.Context, travelID string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM travel_arrivals_processed WHERE travel_id = ?`, travelID).Scan(&exists)
	return exists > 0, err
}

func travelArrivalProcessedTx(ctx context.Context, tx *sql.Tx, travelID string) (bool, error) {
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM travel_arrivals_processed WHERE travel_id = ?`, travelID).Scan(&exists)
	return exists > 0, err
}

func upsertFullEconomyState(ctx context.Context, execer sqlExecer, state game.EconomyState) error {
	key, err := usernameKey(state.Empire.Username)
	if err != nil {
		return err
	}
	if _, err := execer.ExecContext(ctx, `INSERT INTO empires (
		id, username, username_key, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		username = excluded.username,
		username_key = excluded.username_key,
		password_hash = excluded.password_hash,
		world_name = excluded.world_name,
		world_name_key = excluded.world_name_key,
		empire_name = excluded.empire_name,
		empire_name_key = excluded.empire_name_key,
		turns_left = excluded.turns_left,
		turn_day = excluded.turn_day,
		money = excluded.money,
		money_bank = excluded.money_bank,
		population = excluded.population,
		food = excluded.food,
		food_storage = excluded.food_storage,
		energy = excluded.energy,
		research_pts = excluded.research_pts,
		building_pts = excluded.building_pts,
		updated_at = CURRENT_TIMESTAMP`,
		state.Empire.ID, state.Empire.Username, key, state.Empire.PasswordHash, state.Empire.WorldName,
		state.Empire.WorldNameKey, state.Empire.EmpireName, state.Empire.EmpireNameKey,
		state.Empire.TurnsLeft, state.Empire.TurnDay, state.Empire.Money, state.Empire.MoneyBank,
		state.Empire.Population, state.Empire.Food, state.Empire.FoodStorage, state.Empire.Energy,
		state.Empire.ResearchPts, state.Empire.BuildingPts); err != nil {
		return err
	}
	if err := saveRegions(ctx, execer, state.Regions); err != nil {
		return err
	}
	if err := saveBuildings(ctx, execer, state.Buildings); err != nil {
		return err
	}
	if err := saveMines(ctx, execer, state.Mines); err != nil {
		return err
	}
	if err := saveTech(ctx, execer, state.Empire.ID, state.Tech); err != nil {
		return err
	}
	return saveMilitary(ctx, execer, state.Military)
}
