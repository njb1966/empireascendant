package data

import (
	"testing"
	"time"

	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

func TestFederationEventLifecycleAndNews(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()

	event, err := s.EmitEvent(ctx, "ascendant", "empireascendant.empire_founded", map[string]any{
		"empire_name": "Solar Crown",
		"world_name":  "New Terra",
		"node":        "ascendant",
	}, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if event.EventID != "ascendant:1" || event.Seq != 1 {
		t.Fatalf("event = %+v", event)
	}
	outbound, err := s.OutboundEvents(ctx, "ascendant", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(outbound) != 1 || outbound[0].EventID != event.EventID {
		t.Fatalf("outbound = %+v", outbound)
	}
	if err := s.MarkEventsPushed(ctx, outbound, 5); err != nil {
		t.Fatal(err)
	}
	outbound, err = s.OutboundEvents(ctx, "ascendant", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(outbound) != 0 {
		t.Fatalf("outbound after push = %+v", outbound)
	}

	if err := s.StorePulledEvents(ctx, []interdoor.FeedEvent{{
		HubSeq: 6,
		Event: interdoor.Event{
			EventID:    "remote:1",
			SourceNode: "remote",
			Seq:        1,
			Type:       "empireascendant.galactic_news",
			TS:         101,
			Payload:    []byte(`{"headline":"Remote empire stirred."}`),
		},
	}}); err != nil {
		t.Fatal(err)
	}
	news, err := s.News(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(news) != 2 || news[0].Headline != "Remote empire stirred." || news[1].Headline != "Solar Crown has risen on New Terra." {
		t.Fatalf("news = %+v", news)
	}
}

func TestFederationRosterAndRankings(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	e := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	state.Military.NormalSoldiers = 200
	if err := s.SaveEconomy(ctx, state); err != nil {
		t.Fatal(err)
	}

	local, err := s.LocalRankings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(local) != 1 || local[0].EmpireName != "Solar Crown" || local[0].Score <= 0 {
		t.Fatalf("local rankings = %+v", local)
	}

	roster, err := s.LocalRosterEntries(ctx, "ascendant", time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(roster) != 1 || roster[0].GlobalID != GlobalEmpireID("ascendant", e.ID) || roster[0].Status != "active" {
		t.Fatalf("local roster = %+v", roster)
	}

	if err := s.UpsertRemoteRoster(ctx, []interdoor.RosterEntry{{
		GlobalID: "remote:p_1",
		NodeID:   "remote",
		Name:     "Far Crown",
		Level:    123,
		Status:   "active",
		LastSeen: time.Unix(100, 0).Unix(),
	}}); err != nil {
		t.Fatal(err)
	}
	remote, err := s.RemoteRoster(ctx, time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(remote) != 1 || remote[0].EmpireName != "Far Crown" || remote[0].Stale {
		t.Fatalf("remote roster = %+v", remote)
	}
}

func TestFederationValueAndCursor(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	if err := s.SetFederationValue(ctx, StateAPIKey, "secret"); err != nil {
		t.Fatal(err)
	}
	got, err := s.FederationValue(ctx, StateAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret" {
		t.Fatalf("api key = %q", got)
	}
	if err := s.SetHubCursor(ctx, 42); err != nil {
		t.Fatal(err)
	}
	cursor, err := s.HubCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cursor != 42 {
		t.Fatalf("cursor = %d", cursor)
	}
}

func TestGlobalEmpireID(t *testing.T) {
	if got := GlobalEmpireID("ascendant", "abc"); got != "ascendant:p_abc" {
		t.Fatalf("global id = %q", got)
	}
}

func TestLocalPlayerCount(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	createTestEmpire(t, s, "player2", "Iron March", "Iron Banner")
	count, err := s.LocalPlayerCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d", count)
	}
}

func TestScoreUsedForRosterLevel(t *testing.T) {
	state := game.NewStartingEconomy("empire-1")
	state.Empire.Population = 1
	if game.EmpireScore(state) <= 0 {
		t.Fatal("score should be positive")
	}
}
