package data

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"empireascendant/internal/game"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCreateEmpirePersistsDefaults(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e, err := s.CreateEmpire(ctx, CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
		Now:          time.Unix(86400*5, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.WorldNameKey != "NEWTERRA" || e.EmpireNameKey != "SOLARCROWN" {
		t.Fatalf("keys = %q %q", e.WorldNameKey, e.EmpireNameKey)
	}
	got, err := s.FindByUsername(ctx, "player1")
	if err != nil {
		t.Fatal(err)
	}
	if got.TurnsLeft != game.DefaultTurns || got.Money != game.DefaultMoney || got.Population != game.DefaultPopulation || got.Food != game.DefaultFood || got.Energy != game.DefaultEnergy {
		t.Fatalf("defaults not persisted: %+v", got)
	}
}

func TestCreateEmpireSeedsEconomy(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e, err := s.CreateEmpire(ctx, CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	if state.Regions[game.RegionAgricultural].Quantity != game.DefaultAgriculturalRegions {
		t.Fatalf("ag regions = %+v", state.Regions[game.RegionAgricultural])
	}
	if state.Buildings.ResearchLabs != game.DefaultResearchLabs {
		t.Fatalf("buildings = %+v", state.Buildings)
	}
	if state.Mines[game.MineGold].NumMines != game.DefaultGoldMines {
		t.Fatalf("gold = %+v", state.Mines[game.MineGold])
	}
	if !state.Tech[game.TechEnergyFossil] {
		t.Fatalf("tech = %+v", state.Tech)
	}
}

func TestSaveEconomyPersistsChanges(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e, err := s.CreateEmpire(ctx, CreateEmpireInput{Username: "player1", PasswordHash: "x", WorldName: "New Terra", EmpireName: "Solar Crown"})
	if err != nil {
		t.Fatal(err)
	}
	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	state.Empire.Money = 123
	ag := state.Regions[game.RegionAgricultural]
	ag.Activated = 1
	state.Regions[game.RegionAgricultural] = ag
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
	if got.Empire.Money != 123 || got.Regions[game.RegionAgricultural].Activated != 1 {
		t.Fatalf("got = %+v", got)
	}
}

func TestLoadEconomyBackfillsD1Empire(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	e := game.Empire{
		ID:            "legacy-empire",
		Username:      "legacy",
		PasswordHash:  game.HashPassword("secret"),
		WorldName:     "Old World",
		WorldNameKey:  "OLDWORLD",
		EmpireName:    "Old Crown",
		EmpireNameKey: "OLDCROWN",
	}
	game.ApplyDefaults(&e, 1)
	_, err := s.db.ExecContext(ctx, `INSERT INTO empires (
		id, username, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Username, e.PasswordHash, e.WorldName, e.WorldNameKey, e.EmpireName, e.EmpireNameKey,
		e.TurnsLeft, e.TurnDay, e.Money, e.MoneyBank, e.Population, e.Food, e.FoodStorage, e.Energy,
		e.ResearchPts, e.BuildingPts)
	if err != nil {
		t.Fatal(err)
	}

	state, err := s.LoadEconomy(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	if state.Buildings.ResearchLabs != game.DefaultResearchLabs {
		t.Fatalf("buildings = %+v", state.Buildings)
	}
	if state.Regions[game.RegionAgricultural].Quantity != game.DefaultAgriculturalRegions {
		t.Fatalf("regions = %+v", state.Regions)
	}
	if state.Mines[game.MineGold].NumMines != game.DefaultGoldMines {
		t.Fatalf("mines = %+v", state.Mines)
	}
	if !state.Tech[game.TechEnergyFossil] {
		t.Fatalf("tech = %+v", state.Tech)
	}
}

func TestCreateEmpireRejectsDuplicateWorldAndEmpireNames(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	_, err := s.CreateEmpire(ctx, CreateEmpireInput{Username: "p1", PasswordHash: "x", WorldName: "New Terra", EmpireName: "Solar Crown"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CreateEmpire(ctx, CreateEmpireInput{Username: "p2", PasswordHash: "x", WorldName: "newterra", EmpireName: "Other"})
	if !errors.Is(err, ErrDuplicateWorld) {
		t.Fatalf("world err = %v", err)
	}
	_, err = s.CreateEmpire(ctx, CreateEmpireInput{Username: "p3", PasswordHash: "x", WorldName: "Other", EmpireName: "Solar  Crown"})
	if !errors.Is(err, ErrDuplicateEmpire) {
		t.Fatalf("empire err = %v", err)
	}
}

func TestUsernameLookupIsCaseInsensitive(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	created, err := s.CreateEmpire(ctx, CreateEmpireInput{
		Username:     "CalvusRex",
		PasswordHash: "x",
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.FindByUsername(ctx, "calvusrex")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.Username != "CalvusRex" {
		t.Fatalf("got = %+v created=%+v", got, created)
	}
	_, err = s.CreateEmpire(ctx, CreateEmpireInput{
		Username:     "CALVUSREX",
		PasswordHash: "x",
		WorldName:    "Other World",
		EmpireName:   "Other Crown",
	})
	if !errors.Is(err, ErrDuplicateUser) {
		t.Fatalf("duplicate user err = %v", err)
	}
}

func TestInitMigratesLegacyEmpireUsernameKeyBeforeIndex(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE empires (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
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
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO empires (
		id, username, password_hash, world_name, world_name_key, empire_name, empire_name_key,
		turns_left, turn_day, money, money_bank, population, food, food_storage, energy,
		research_pts, building_pts
	) VALUES (
		'empire-1', 'CalvusRex', 'hash', 'New Terra', 'NEWTERRA', 'Solar Crown', 'SOLARCROWN',
		15, 1, 10000, 0, 20000, 10000, 0, 300, 0, 0
	)`); err != nil {
		t.Fatal(err)
	}
	if err := s.Init(ctx); err != nil {
		t.Fatalf("init legacy db: %v", err)
	}
	empire, err := s.FindByUsername(ctx, "calvusrex")
	if err != nil {
		t.Fatalf("find migrated username: %v", err)
	}
	if empire.Username != "CalvusRex" {
		t.Fatalf("username = %q", empire.Username)
	}
}
