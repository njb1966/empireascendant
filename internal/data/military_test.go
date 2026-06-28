package data

import (
	"context"
	"errors"
	"strings"
	"testing"

	"empireascendant/internal/game"
)

func createTestEmpire(t *testing.T, s *Store, username, world, empire string) game.Empire {
	t.Helper()
	e, err := s.CreateEmpire(context.Background(), CreateEmpireInput{
		Username:     username,
		PasswordHash: game.HashPassword("secret"),
		WorldName:    world,
		EmpireName:   empire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestCreateEmpireSeedsMilitary(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")

	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	if state.Military.NormalSoldiers != game.DefaultNormalSoldiers {
		t.Fatalf("military = %+v", state.Military)
	}
	if !state.Tech[game.TechSoldierNormal] || state.Tech[game.TechBallisticNuclear] {
		t.Fatalf("tech = %+v", state.Tech)
	}
}

func TestSaveEconomyPersistsMilitaryAndTech(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")

	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	state.Military.NormalSoldiers = 250
	state.Military.NuclearMissiles = 2
	state.Tech[game.TechBallisticNuclear] = true
	if err := s.SaveEconomy(ctx, state); err != nil {
		t.Fatal(err)
	}

	gotEmpire, err := s.FindByUsername(ctx, "player1")
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadEconomy(ctx, gotEmpire)
	if err != nil {
		t.Fatal(err)
	}
	if got.Military.NormalSoldiers != 250 || got.Military.NuclearMissiles != 2 || !got.Tech[game.TechBallisticNuclear] {
		t.Fatalf("got = %+v", got)
	}
}

func TestLoadEconomyBackfillsD3MilitaryWithoutResettingTech(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	state.Tech[game.TechBallisticNuclear] = true
	if err := s.SaveEconomy(ctx, state); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM empire_military WHERE empire_id = ?`, e.ID); err != nil {
		t.Fatal(err)
	}

	got, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	if got.Military.NormalSoldiers != game.DefaultNormalSoldiers {
		t.Fatalf("military = %+v", got.Military)
	}
	if !got.Tech[game.TechBallisticNuclear] {
		t.Fatalf("researched tech was reset: %+v", got.Tech)
	}
}

func TestAttackLimitsAndDispatches(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	attacker := createTestEmpire(t, s, "player1", "New Terra", "Solar Crown")
	defender := createTestEmpire(t, s, "player2", "Iron March", "Iron Banner")

	for i := 0; i < game.GroundAttackLimit; i++ {
		if err := s.RecordAttack(ctx, attacker.ID, defender.ID, game.ActionGroundAttack, int(attacker.TurnDay)); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	if err := s.CheckAttackLimit(ctx, attacker.ID, defender.ID, game.ActionGroundAttack, int(attacker.TurnDay)); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("limit err = %v", err)
	}
	if err := s.CheckAttackLimit(ctx, attacker.ID, defender.ID, game.ActionGroundAttack, int(attacker.TurnDay)+1); err != nil {
		t.Fatalf("next day err = %v", err)
	}

	if _, err := s.AddDispatch(ctx, defender.ID, "ground", "Solar Crown attacked Iron Banner."); err != nil {
		t.Fatal(err)
	}
	dispatches, err := s.ListDispatches(ctx, defender.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(dispatches) != 1 || !strings.Contains(dispatches[0].Message, "attacked") {
		t.Fatalf("dispatches = %+v", dispatches)
	}
}
