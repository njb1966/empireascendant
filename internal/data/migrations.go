package data

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) migrateUsernameKeys(ctx context.Context) error {
	exists, err := s.columnExists(ctx, "empires", "username_key")
	if err != nil {
		return err
	}
	if !exists {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE empires ADD COLUMN username_key TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("add username_key: %w", err)
		}
	}
	if _, err := s.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_empires_username_key ON empires(username_key)`); err != nil {
		return fmt.Errorf("index username_key: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, username FROM empires WHERE username_key = ''`)
	if err != nil {
		return err
	}
	type usernameBackfill struct {
		id       string
		username string
		key      string
	}
	var backfills []usernameBackfill
	for rows.Next() {
		var id, username string
		if err := rows.Scan(&id, &username); err != nil {
			_ = rows.Close()
			return err
		}
		key, err := usernameKey(username)
		if err != nil {
			_ = rows.Close()
			return fmt.Errorf("backfill username %q: %w", username, err)
		}
		backfills = append(backfills, usernameBackfill{id: id, username: username, key: key})
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, backfill := range backfills {
		if _, err := s.db.ExecContext(ctx, `UPDATE empires SET username_key = ? WHERE id = ?`, backfill.key, backfill.id); err != nil {
			return fmt.Errorf("update username_key for %q: %w", backfill.username, err)
		}
	}
	return nil
}

func (s *Store) columnExists(ctx context.Context, table, column string) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
