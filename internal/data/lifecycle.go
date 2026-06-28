package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"empireascendant/internal/game"
)

type Lifecycle struct {
	EmpireID        string
	LastLoginAt     int64
	WarnedAt        int64
	HiddenAt        int64
	PurgeEligibleAt int64
	Status          string
	UpdatedAt       int64
}

func (s *Store) TouchEmpireLogin(ctx context.Context, empireID string, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO empire_lifecycle
		(empire_id, last_login_at, warned_at, hidden_at, purge_eligible_at, status, updated_at)
		VALUES (?, ?, 0, 0, 0, ?, ?)
		ON CONFLICT(empire_id) DO UPDATE SET
			last_login_at = excluded.last_login_at,
			warned_at = 0,
			hidden_at = 0,
			purge_eligible_at = 0,
			status = excluded.status,
			updated_at = excluded.updated_at`,
		empireID, now.Unix(), game.LifecycleActive, now.Unix())
	return err
}

func (s *Store) EmpireLifecycle(ctx context.Context, empireID string, now time.Time) (Lifecycle, error) {
	lifecycle, err := s.lifecycle(ctx, empireID)
	if errors.Is(err, sql.ErrNoRows) {
		return s.seedLifecycle(ctx, empireID, now)
	}
	if err != nil {
		return Lifecycle{}, err
	}
	return lifecycle, nil
}

func (s *Store) RefreshLifecycle(ctx context.Context, empireID string, now time.Time) (Lifecycle, error) {
	lifecycle, err := s.EmpireLifecycle(ctx, empireID, now)
	if err != nil {
		return Lifecycle{}, err
	}
	refreshed := refreshLifecycleStatus(lifecycle, now)
	if refreshed == lifecycle {
		return refreshed, nil
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE empire_lifecycle SET
		warned_at = ?, hidden_at = ?, purge_eligible_at = ?, status = ?, updated_at = ?
		WHERE empire_id = ?`,
		refreshed.WarnedAt, refreshed.HiddenAt, refreshed.PurgeEligibleAt, refreshed.Status,
		refreshed.UpdatedAt, refreshed.EmpireID); err != nil {
		return Lifecycle{}, err
	}
	return refreshed, nil
}

func (s *Store) RefreshAllLifecycle(ctx context.Context, now time.Time) error {
	empires, err := s.allEmpires(ctx)
	if err != nil {
		return err
	}
	for _, empire := range empires {
		if _, err := s.RefreshLifecycle(ctx, empire.ID, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) visibleEmpireIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT e.id
		FROM empires e
		LEFT JOIN empire_lifecycle l ON l.empire_id = e.id
		WHERE COALESCE(l.status, ?) NOT IN (?, ?)`,
		game.LifecycleActive, game.LifecycleHidden, game.LifecyclePurgeEligible)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}

func (s *Store) lifecycle(ctx context.Context, empireID string) (Lifecycle, error) {
	var lifecycle Lifecycle
	err := s.db.QueryRowContext(ctx, `SELECT
		empire_id, last_login_at, warned_at, hidden_at, purge_eligible_at, status, updated_at
		FROM empire_lifecycle WHERE empire_id = ?`, empireID).Scan(
		&lifecycle.EmpireID, &lifecycle.LastLoginAt, &lifecycle.WarnedAt, &lifecycle.HiddenAt,
		&lifecycle.PurgeEligibleAt, &lifecycle.Status, &lifecycle.UpdatedAt)
	return lifecycle, err
}

func (s *Store) seedLifecycle(ctx context.Context, empireID string, now time.Time) (Lifecycle, error) {
	if now.IsZero() {
		now = time.Now()
	}
	var createdAt int64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(unixepoch(created_at), ?) FROM empires WHERE id = ?`,
		now.Unix(), empireID).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Lifecycle{}, ErrNotFound
	}
	if err != nil {
		return Lifecycle{}, err
	}
	lifecycle := Lifecycle{
		EmpireID:    empireID,
		LastLoginAt: createdAt,
		Status:      game.LifecycleActive,
		UpdatedAt:   now.Unix(),
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO empire_lifecycle
		(empire_id, last_login_at, warned_at, hidden_at, purge_eligible_at, status, updated_at)
		VALUES (?, ?, 0, 0, 0, ?, ?)`,
		lifecycle.EmpireID, lifecycle.LastLoginAt, lifecycle.Status, lifecycle.UpdatedAt); err != nil {
		return Lifecycle{}, err
	}
	return lifecycle, nil
}

func refreshLifecycleStatus(lifecycle Lifecycle, now time.Time) Lifecycle {
	if now.IsZero() {
		now = time.Now()
	}
	elapsed := now.Sub(time.Unix(lifecycle.LastLoginAt, 0))
	next := lifecycle
	next.UpdatedAt = now.Unix()
	switch {
	case elapsed >= game.LifecyclePurgeEligibleAfter:
		next.Status = game.LifecyclePurgeEligible
		if next.WarnedAt == 0 {
			next.WarnedAt = now.Unix()
		}
		if next.HiddenAt == 0 {
			next.HiddenAt = now.Unix()
		}
		if next.PurgeEligibleAt == 0 {
			next.PurgeEligibleAt = now.Unix()
		}
	case elapsed >= game.LifecycleHideAfter:
		next.Status = game.LifecycleHidden
		if next.WarnedAt == 0 {
			next.WarnedAt = now.Unix()
		}
		if next.HiddenAt == 0 {
			next.HiddenAt = now.Unix()
		}
		next.PurgeEligibleAt = 0
	case elapsed >= game.LifecycleWarnAfter:
		next.Status = game.LifecycleWarned
		if next.WarnedAt == 0 {
			next.WarnedAt = now.Unix()
		}
		next.HiddenAt = 0
		next.PurgeEligibleAt = 0
	default:
		next.Status = game.LifecycleActive
		next.WarnedAt = 0
		next.HiddenAt = 0
		next.PurgeEligibleAt = 0
	}
	return next
}
