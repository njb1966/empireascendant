package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func OpenMemory() (*Store, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empires (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			username_key TEXT NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL,
			world_name TEXT NOT NULL,
			world_name_key TEXT NOT NULL UNIQUE,
			empire_name TEXT NOT NULL,
			empire_name_key TEXT NOT NULL UNIQUE,
			turns_left INTEGER NOT NULL,
			turn_day INTEGER NOT NULL,
			money INTEGER NOT NULL,
			money_bank INTEGER NOT NULL,
			population INTEGER NOT NULL,
			food INTEGER NOT NULL,
			food_storage INTEGER NOT NULL,
			energy INTEGER NOT NULL,
			research_pts INTEGER NOT NULL,
			building_pts INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_empires_username ON empires(username)`,
		`CREATE TABLE IF NOT EXISTS empire_regions (
			empire_id TEXT NOT NULL,
			region_type TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			activated INTEGER NOT NULL,
			activate_cost INTEGER NOT NULL,
			PRIMARY KEY (empire_id, region_type)
		)`,
		`CREATE TABLE IF NOT EXISTS empire_buildings (
			empire_id TEXT PRIMARY KEY,
			miners_guild INTEGER NOT NULL,
			miners_available INTEGER NOT NULL,
			fishing_guild INTEGER NOT NULL,
			fishers_assigned INTEGER NOT NULL,
			fish_stock INTEGER NOT NULL,
			construction_factories INTEGER NOT NULL,
			research_labs INTEGER NOT NULL,
			fossil_plants INTEGER NOT NULL,
			fission_plants INTEGER NOT NULL,
			fusion_plants INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS empire_mines (
			empire_id TEXT NOT NULL,
			mine_type TEXT NOT NULL,
			num_mines INTEGER NOT NULL,
			miners_assigned INTEGER NOT NULL,
			mineral_left INTEGER NOT NULL,
			stored_minerals INTEGER NOT NULL,
			PRIMARY KEY (empire_id, mine_type)
		)`,
		`CREATE TABLE IF NOT EXISTS empire_tech (
			empire_id TEXT NOT NULL,
			tech_key TEXT NOT NULL,
			unlocked INTEGER NOT NULL,
			PRIMARY KEY (empire_id, tech_key)
		)`,
		`CREATE TABLE IF NOT EXISTS empire_military (
			empire_id TEXT PRIMARY KEY,
			normal_soldiers INTEGER NOT NULL,
			super_soldiers INTEGER NOT NULL,
			mega_soldiers INTEGER NOT NULL,
			tanks INTEGER NOT NULL,
			hovercraft INTEGER NOT NULL,
			nuclear_missiles INTEGER NOT NULL,
			antimatter_missiles INTEGER NOT NULL,
			recon_drones INTEGER NOT NULL,
			spies INTEGER NOT NULL,
			terrorists INTEGER NOT NULL,
			ground_turrets INTEGER NOT NULL,
			orbital_satellites INTEGER NOT NULL,
			global_shields INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS attack_limits (
			attacker_id TEXT NOT NULL,
			defender_id TEXT NOT NULL,
			action TEXT NOT NULL,
			turn_day INTEGER NOT NULL,
			count INTEGER NOT NULL,
			PRIMARY KEY (attacker_id, defender_id, action, turn_day)
		)`,
		`CREATE TABLE IF NOT EXISTS dispatches (
			id TEXT PRIMARY KEY,
			empire_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			read_at TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_dispatches_empire_created ON dispatches(empire_id, created_at)`,
		`CREATE TABLE IF NOT EXISTS federation_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS federation_events (
			event_id TEXT PRIMARY KEY,
			source_node TEXT NOT NULL,
			seq INTEGER NOT NULL,
			hub_seq INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			ts INTEGER NOT NULL,
			payload TEXT NOT NULL,
			pushed INTEGER NOT NULL,
			applied INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_federation_events_outbound ON federation_events(source_node, pushed, seq)`,
		`CREATE INDEX IF NOT EXISTS idx_federation_events_hub_seq ON federation_events(hub_seq)`,
		`CREATE TABLE IF NOT EXISTS remote_roster (
			global_id TEXT PRIMARY KEY,
			node_id TEXT NOT NULL,
			name TEXT NOT NULL,
			level INTEGER NOT NULL,
			status TEXT NOT NULL,
			last_seen INTEGER NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_remote_roster_node ON remote_roster(node_id)`,
		`CREATE TABLE IF NOT EXISTS federation_news (
			event_id TEXT PRIMARY KEY,
			source_node TEXT NOT NULL,
			event_type TEXT NOT NULL,
			headline TEXT NOT NULL,
			ts INTEGER NOT NULL,
			hub_seq INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_federation_news_ts ON federation_news(ts)`,
		`CREATE TABLE IF NOT EXISTS pvp_outbound (
			request_id TEXT PRIMARY KEY,
			attacker_empire_id TEXT NOT NULL,
			attacker_global_id TEXT NOT NULL,
			victim_global_id TEXT NOT NULL,
			action TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			resolved_at INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pvp_outbound_attacker ON pvp_outbound(attacker_global_id, status)`,
		`CREATE TABLE IF NOT EXISTS pvp_inbound_processed (
			request_id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL,
			processed_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS pvp_results_applied (
			event_id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS travel_state (
			empire_id TEXT PRIMARY KEY,
			global_id TEXT NOT NULL,
			home_node TEXT NOT NULL,
			current_node TEXT NOT NULL,
			dest_node TEXT NOT NULL,
			status TEXT NOT NULL,
			travel_id TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_travel_state_global ON travel_state(global_id)`,
		`CREATE TABLE IF NOT EXISTS travel_visitors (
			global_id TEXT PRIMARY KEY,
			home_node TEXT NOT NULL,
			current_node TEXT NOT NULL,
			from_node TEXT NOT NULL,
			empire_name TEXT NOT NULL,
			world_name TEXT NOT NULL,
			snapshot TEXT NOT NULL,
			status TEXT NOT NULL,
			arrived_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS travel_arrivals_processed (
			travel_id TEXT PRIMARY KEY,
			global_id TEXT NOT NULL,
			event_id TEXT NOT NULL,
			processed_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS empire_lifecycle (
			empire_id TEXT PRIMARY KEY,
			last_login_at INTEGER NOT NULL,
			warned_at INTEGER NOT NULL DEFAULT 0,
			hidden_at INTEGER NOT NULL DEFAULT 0,
			purge_eligible_at INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_empire_lifecycle_status ON empire_lifecycle(status)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("execute schema: %w", err)
		}
	}
	if err := s.migrateUsernameKeys(ctx); err != nil {
		return err
	}
	return nil
}
