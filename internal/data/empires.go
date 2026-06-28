package data

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"empireascendant/internal/game"
)

var (
	ErrNotFound        = errors.New("empire not found")
	ErrDuplicateUser   = errors.New("username already exists")
	ErrDuplicateWorld  = errors.New("world name already exists")
	ErrDuplicateEmpire = errors.New("empire name already exists")
)

type CreateEmpireInput struct {
	Username     string
	PasswordHash string
	WorldName    string
	EmpireName   string
	Now          time.Time
}

func (s *Store) CreateEmpire(ctx context.Context, in CreateEmpireInput) (game.Empire, error) {
	userName, err := game.NormalizeName(in.Username)
	if err != nil {
		return game.Empire{}, fmt.Errorf("username: %w", err)
	}
	userKey := userName.Key
	world, err := game.NormalizeName(in.WorldName)
	if err != nil {
		return game.Empire{}, fmt.Errorf("world name: %w", err)
	}
	empireName, err := game.NormalizeName(in.EmpireName)
	if err != nil {
		return game.Empire{}, fmt.Errorf("empire name: %w", err)
	}

	if exists, err := s.exists(ctx, "username_key", userKey); err != nil {
		return game.Empire{}, err
	} else if exists {
		return game.Empire{}, ErrDuplicateUser
	}
	if exists, err := s.exists(ctx, "world_name_key", world.Key); err != nil {
		return game.Empire{}, err
	} else if exists {
		return game.Empire{}, ErrDuplicateWorld
	}
	if exists, err := s.exists(ctx, "empire_name_key", empireName.Key); err != nil {
		return game.Empire{}, err
	} else if exists {
		return game.Empire{}, ErrDuplicateEmpire
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}
	e := game.Empire{
		ID:            newID(),
		Username:      userName.Display,
		PasswordHash:  in.PasswordHash,
		WorldName:     world.Display,
		WorldNameKey:  world.Key,
		EmpireName:    empireName.Display,
		EmpireNameKey: empireName.Key,
	}
	game.ApplyDefaults(&e, game.UnixDay(now))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return game.Empire{}, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO empires (
		id, username, username_key, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Username, userKey, e.PasswordHash, e.WorldName, e.WorldNameKey, e.EmpireName, e.EmpireNameKey,
		e.TurnsLeft, e.TurnDay, e.Money, e.MoneyBank, e.Population, e.Food, e.FoodStorage, e.Energy,
		e.ResearchPts, e.BuildingPts,
	)
	if err != nil {
		return game.Empire{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO empire_lifecycle
		(empire_id, last_login_at, warned_at, hidden_at, purge_eligible_at, status, updated_at)
		VALUES (?, ?, 0, 0, 0, ?, ?)`,
		e.ID, now.Unix(), game.LifecycleActive, now.Unix()); err != nil {
		return game.Empire{}, err
	}
	if err := insertStartingEconomy(ctx, tx, e.ID); err != nil {
		return game.Empire{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.Empire{}, err
	}
	return e, nil
}

func (s *Store) FindByUsername(ctx context.Context, username string) (game.Empire, error) {
	key, err := usernameKey(username)
	if err != nil {
		return game.Empire{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `SELECT
		id, username, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
		FROM empires WHERE username_key = ?
		ORDER BY CASE WHEN username = ? THEN 0 ELSE 1 END, created_at ASC, id ASC
		LIMIT 1`, key, username)
	return scanEmpire(row)
}

func (s *Store) UpdateTurns(ctx context.Context, e game.Empire) error {
	_, err := s.db.ExecContext(ctx, `UPDATE empires SET turns_left = ?, turn_day = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, e.TurnsLeft, e.TurnDay, e.ID)
	return err
}

func (s *Store) SaveEmpire(ctx context.Context, e game.Empire) error {
	_, err := s.db.ExecContext(ctx, `UPDATE empires SET
		turns_left = ?, turn_day = ?, money = ?, money_bank = ?, population = ?,
		food = ?, food_storage = ?, energy = ?, research_pts = ?, building_pts = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		e.TurnsLeft, e.TurnDay, e.Money, e.MoneyBank, e.Population,
		e.Food, e.FoodStorage, e.Energy, e.ResearchPts, e.BuildingPts,
		e.ID,
	)
	return err
}

func (s *Store) exists(ctx context.Context, column, value string) (bool, error) {
	var one int
	query := fmt.Sprintf("SELECT 1 FROM empires WHERE %s = ? LIMIT 1", column)
	err := s.db.QueryRowContext(ctx, query, value).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func usernameKey(username string) (string, error) {
	name, err := game.NormalizeName(username)
	if err != nil {
		return "", err
	}
	return name.Key, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEmpire(row rowScanner) (game.Empire, error) {
	var e game.Empire
	err := row.Scan(
		&e.ID, &e.Username, &e.PasswordHash, &e.WorldName, &e.WorldNameKey, &e.EmpireName, &e.EmpireNameKey,
		&e.TurnsLeft, &e.TurnDay, &e.Money, &e.MoneyBank, &e.Population, &e.Food, &e.FoodStorage, &e.Energy,
		&e.ResearchPts, &e.BuildingPts,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return game.Empire{}, ErrNotFound
	}
	if err != nil {
		return game.Empire{}, err
	}
	return e, nil
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("empire-%d", time.Now().UnixNano())
}
