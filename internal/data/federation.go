package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

const (
	StateAPIKey    = "api_key"
	StateHubCursor = "hub_cursor"
)

type sqlQueryExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type RankingEntry struct {
	GlobalID   string
	EmpireName string
	WorldName  string
	NodeID     string
	Score      int
	Remote     bool
	Stale      bool
	LastSeen   int64
}

type NewsItem struct {
	EventID    string
	SourceNode string
	Type       string
	Headline   string
	TS         int64
}

func GlobalEmpireID(nodeID, empireID string) string {
	return fmt.Sprintf("%s:p_%s", nodeID, empireID)
}

func (s *Store) SetFederationValue(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO federation_state (key, value)
		VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`, key, value)
	return err
}

func (s *Store) FederationValue(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM federation_state WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (s *Store) SetHubCursor(ctx context.Context, cursor int64) error {
	return s.SetFederationValue(ctx, StateHubCursor, fmt.Sprintf("%d", cursor))
}

func (s *Store) HubCursor(ctx context.Context) (int64, error) {
	value, err := s.FederationValue(ctx, StateHubCursor)
	if err != nil || value == "" {
		return 0, err
	}
	var cursor int64
	_, err = fmt.Sscanf(value, "%d", &cursor)
	return cursor, err
}

func (s *Store) EmitEvent(ctx context.Context, nodeID, eventType string, payload map[string]any, now time.Time) (interdoor.Event, error) {
	if nodeID == "" {
		return interdoor.Event{}, errors.New("node id is required")
	}
	if now.IsZero() {
		now = time.Now()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return interdoor.Event{}, err
	}
	defer tx.Rollback()
	event, err := appendLocalEvent(ctx, tx, nodeID, eventType, payload, now.Unix())
	if err != nil {
		return interdoor.Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return interdoor.Event{}, err
	}
	return event, nil
}

func appendLocalEvent(ctx context.Context, execer sqlQueryExecer, nodeID, eventType string, payload map[string]any, ts int64) (interdoor.Event, error) {
	var seq int64
	if err := execer.QueryRowContext(ctx, `SELECT COALESCE(MAX(seq), 0) + 1 FROM federation_events WHERE source_node = ?`, nodeID).Scan(&seq); err != nil {
		return interdoor.Event{}, err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return interdoor.Event{}, err
	}
	event := interdoor.Event{
		EventID:    fmt.Sprintf("%s:%d", nodeID, seq),
		SourceNode: nodeID,
		Seq:        seq,
		Type:       eventType,
		TS:         ts,
		Payload:    raw,
	}
	if _, err := execer.ExecContext(ctx, `INSERT INTO federation_events
		(event_id, source_node, seq, hub_seq, event_type, ts, payload, pushed, applied)
		VALUES (?, ?, ?, 0, ?, ?, ?, 0, 1)`,
		event.EventID, event.SourceNode, event.Seq, event.Type, event.TS, string(event.Payload)); err != nil {
		return interdoor.Event{}, err
	}
	if err := insertNews(ctx, execer, event.EventID, event.SourceNode, event.Type, event.Payload, event.TS, 0); err != nil {
		return interdoor.Event{}, err
	}
	return event, nil
}

func (s *Store) OutboundEvents(ctx context.Context, nodeID string, limit int) ([]interdoor.Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT event_id, source_node, seq, event_type, ts, payload
		FROM federation_events
		WHERE source_node = ? AND pushed = 0
		ORDER BY seq
		LIMIT ?`, nodeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []interdoor.Event
	for rows.Next() {
		var event interdoor.Event
		var payload string
		if err := rows.Scan(&event.EventID, &event.SourceNode, &event.Seq, &event.Type, &event.TS, &payload); err != nil {
			return nil, err
		}
		event.Payload = json.RawMessage(payload)
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) MarkEventsPushed(ctx context.Context, events []interdoor.Event, hubSeq int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, event := range events {
		if _, err := tx.ExecContext(ctx, `UPDATE federation_events SET pushed = 1, hub_seq = ? WHERE event_id = ?`, hubSeq, event.EventID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) StorePulledEvents(ctx context.Context, events []interdoor.FeedEvent) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, event := range events {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO federation_events
			(event_id, source_node, seq, hub_seq, event_type, ts, payload, pushed, applied)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1, 1)`,
			event.EventID, event.SourceNode, event.Seq, event.HubSeq, event.Type, event.TS, string(event.Payload)); err != nil {
			return err
		}
		if err := insertNews(ctx, tx, event.EventID, event.SourceNode, event.Type, event.Payload, event.TS, event.HubSeq); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertNews(ctx context.Context, execer sqlExecer, eventID, sourceNode, eventType string, payload json.RawMessage, ts, hubSeq int64) error {
	headline := newsHeadline(eventType, payload)
	_, err := execer.ExecContext(ctx, `INSERT OR IGNORE INTO federation_news
		(event_id, source_node, event_type, headline, ts, hub_seq)
		VALUES (?, ?, ?, ?, ?, ?)`, eventID, sourceNode, eventType, headline, ts, hubSeq)
	return err
}

func newsHeadline(eventType string, payload json.RawMessage) string {
	var fields map[string]any
	_ = json.Unmarshal(payload, &fields)
	switch eventType {
	case "empireascendant.empire_founded":
		empire, _ := fields["empire_name"].(string)
		world, _ := fields["world_name"].(string)
		if empire != "" && world != "" {
			return fmt.Sprintf("%s has risen on %s.", empire, world)
		}
	case "empireascendant.attack_resolved", "empireascendant.missile_strike", "empireascendant.galactic_news":
		if text, _ := fields["message"].(string); text != "" {
			return text
		}
		if text, _ := fields["headline"].(string); text != "" {
			return text
		}
	case "pvp.resolved":
		if text, _ := fields["result_text"].(string); text != "" {
			return text
		}
	}
	return fmt.Sprintf("%s from %s", eventType, fields["node"])
}

func (s *Store) UpsertRemoteRoster(ctx context.Context, entries []interdoor.RosterEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, entry := range entries {
		if _, err := tx.ExecContext(ctx, `INSERT INTO remote_roster
			(global_id, node_id, name, level, status, last_seen)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(global_id) DO UPDATE SET
				node_id = excluded.node_id,
				name = excluded.name,
				level = excluded.level,
				status = excluded.status,
				last_seen = excluded.last_seen,
				updated_at = CURRENT_TIMESTAMP`,
			entry.GlobalID, entry.NodeID, entry.Name, entry.Level, entry.Status, entry.LastSeen); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) RemoteRoster(ctx context.Context, now time.Time) ([]RankingEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT global_id, node_id, name, level, status, last_seen
		FROM remote_roster ORDER BY level DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	staleBefore := now.Add(-15 * time.Minute).Unix()
	var entries []RankingEntry
	for rows.Next() {
		var globalID, nodeID, name, status string
		var level int
		var lastSeen int64
		if err := rows.Scan(&globalID, &nodeID, &name, &level, &status, &lastSeen); err != nil {
			return nil, err
		}
		entries = append(entries, RankingEntry{
			GlobalID:   globalID,
			EmpireName: name,
			NodeID:     nodeID,
			Score:      level,
			Remote:     true,
			Stale:      lastSeen < staleBefore,
			LastSeen:   lastSeen,
		})
	}
	return entries, rows.Err()
}

func (s *Store) LocalRankings(ctx context.Context) ([]RankingEntry, error) {
	if err := s.RefreshAllLifecycle(ctx, time.Now()); err != nil {
		return nil, err
	}
	empires, err := s.allEmpires(ctx)
	if err != nil {
		return nil, err
	}
	visible, err := s.visibleEmpireIDs(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]RankingEntry, 0, len(empires))
	for _, empire := range empires {
		if !visible[empire.ID] {
			continue
		}
		state, err := s.LoadEconomy(ctx, empire)
		if err != nil {
			return nil, err
		}
		entries = append(entries, RankingEntry{
			GlobalID:   empire.ID,
			EmpireName: empire.EmpireName,
			WorldName:  empire.WorldName,
			Score:      game.EmpireScore(state),
		})
	}
	sortRankings(entries)
	return entries, nil
}

func (s *Store) LocalRosterEntries(ctx context.Context, nodeID string, now time.Time) ([]interdoor.RosterEntry, error) {
	if err := s.RefreshAllLifecycle(ctx, now); err != nil {
		return nil, err
	}
	rankings, err := s.LocalRankings(ctx)
	if err != nil {
		return nil, err
	}
	empires, err := s.allEmpires(ctx)
	if err != nil {
		return nil, err
	}
	visible, err := s.visibleEmpireIDs(ctx)
	if err != nil {
		return nil, err
	}
	scoreByName := make(map[string]int, len(rankings))
	for _, entry := range rankings {
		scoreByName[entry.EmpireName] = entry.Score
	}
	entries := make([]interdoor.RosterEntry, 0, len(empires))
	for _, empire := range empires {
		if !visible[empire.ID] {
			continue
		}
		status := "active"
		travel, err := s.EmpireTravelStatus(ctx, empire.ID)
		if err != nil {
			return nil, err
		}
		if travel.Status == TravelStatusTraveling || travel.Status == TravelStatusAway {
			status = "traveling"
		}
		entries = append(entries, interdoor.RosterEntry{
			GlobalID: GlobalEmpireID(nodeID, empire.ID),
			Name:     empire.EmpireName,
			Level:    scoreByName[empire.EmpireName],
			Status:   status,
			LastSeen: now.Unix(),
		})
	}
	return entries, nil
}

func (s *Store) LocalPlayerCount(ctx context.Context) (int, error) {
	if err := s.RefreshAllLifecycle(ctx, time.Now()); err != nil {
		return 0, err
	}
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM empires e
		LEFT JOIN travel_state t ON t.empire_id = e.id
		LEFT JOIN empire_lifecycle l ON l.empire_id = e.id
		WHERE COALESCE(t.status, 'active') = 'active'
		AND COALESCE(l.status, ?) NOT IN (?, ?)`,
		game.LifecycleActive, game.LifecycleHidden, game.LifecyclePurgeEligible).Scan(&count)
	return count, err
}

func (s *Store) News(ctx context.Context, limit int) ([]NewsItem, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT event_id, source_node, event_type, headline, ts
		FROM federation_news ORDER BY ts DESC, event_id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NewsItem
	for rows.Next() {
		var item NewsItem
		if err := rows.Scan(&item.EventID, &item.SourceNode, &item.Type, &item.Headline, &item.TS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) allEmpires(ctx context.Context) ([]game.Empire, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT
		id, username, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
		FROM empires ORDER BY empire_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var empires []game.Empire
	for rows.Next() {
		empire, err := scanEmpire(rows)
		if err != nil {
			return nil, err
		}
		empires = append(empires, empire)
	}
	return empires, rows.Err()
}

func sortRankings(entries []RankingEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Score > entries[i].Score ||
				(entries[j].Score == entries[i].Score && entries[j].EmpireName < entries[i].EmpireName) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
