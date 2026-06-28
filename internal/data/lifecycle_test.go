package data

import (
	"testing"
	"time"

	"empireascendant/internal/game"
)

func TestLifecycleTransitionsAndReactivation(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	empire := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	now := time.Unix(10_000_000, 0)
	oldLogin := now.Add(-game.LifecycleHideAfter - time.Hour).Unix()
	if _, err := s.db.ExecContext(ctx, `UPDATE empire_lifecycle SET last_login_at = ?, status = ? WHERE empire_id = ?`,
		oldLogin, game.LifecycleActive, empire.ID); err != nil {
		t.Fatal(err)
	}

	lifecycle, err := s.RefreshLifecycle(ctx, empire.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if lifecycle.Status != game.LifecycleHidden || lifecycle.HiddenAt == 0 {
		t.Fatalf("lifecycle = %+v", lifecycle)
	}
	if err := s.TouchEmpireLogin(ctx, empire.ID, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	lifecycle, err = s.EmpireLifecycle(ctx, empire.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if lifecycle.Status != game.LifecycleActive || lifecycle.WarnedAt != 0 || lifecycle.HiddenAt != 0 || lifecycle.PurgeEligibleAt != 0 {
		t.Fatalf("reactivated lifecycle = %+v", lifecycle)
	}
}

func TestLifecyclePurgeEligibleIsNonDestructive(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	empire := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	now := time.Unix(10_000_000, 0)
	oldLogin := now.Add(-game.LifecyclePurgeEligibleAfter - time.Hour).Unix()
	if _, err := s.db.ExecContext(ctx, `UPDATE empire_lifecycle SET last_login_at = ?, status = ? WHERE empire_id = ?`,
		oldLogin, game.LifecycleActive, empire.ID); err != nil {
		t.Fatal(err)
	}

	lifecycle, err := s.RefreshLifecycle(ctx, empire.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if lifecycle.Status != game.LifecyclePurgeEligible || lifecycle.PurgeEligibleAt == 0 {
		t.Fatalf("lifecycle = %+v", lifecycle)
	}
	if _, err := s.FindEmpireByID(ctx, empire.ID); err != nil {
		t.Fatalf("purge eligible should not delete empire: %v", err)
	}
}

func TestHiddenLifecycleFiltersPublicVisibility(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	active := createTestEmpire(t, s, "active", "New Terra", "Solar Crown")
	hidden := createTestEmpire(t, s, "hidden", "Old Terra", "Old Crown")
	now := time.Now()
	oldLogin := now.Add(-game.LifecycleHideAfter - time.Hour).Unix()
	if _, err := s.db.ExecContext(ctx, `UPDATE empire_lifecycle SET last_login_at = ?, status = ? WHERE empire_id = ?`,
		oldLogin, game.LifecycleActive, hidden.ID); err != nil {
		t.Fatal(err)
	}

	rankings, err := s.LocalRankings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rankings) != 1 || rankings[0].EmpireName != active.EmpireName {
		t.Fatalf("rankings = %+v", rankings)
	}
	roster, err := s.LocalRosterEntries(ctx, "ascendant", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(roster) != 1 || roster[0].Name != active.EmpireName {
		t.Fatalf("roster = %+v", roster)
	}
	targets, err := s.ListTargetEmpires(ctx, active.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatalf("targets = %+v", targets)
	}
	count, err := s.LocalPlayerCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("player count = %d", count)
	}
}
