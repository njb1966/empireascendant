package data

import (
	"context"
	"database/sql"
	"fmt"

	"empireascendant/internal/game"
)

type sqlExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func insertStartingEconomy(ctx context.Context, execer sqlExecer, empireID string) error {
	state := game.NewStartingEconomy(empireID)
	if err := saveRegions(ctx, execer, state.Regions); err != nil {
		return err
	}
	if err := saveBuildings(ctx, execer, state.Buildings); err != nil {
		return err
	}
	if err := saveMines(ctx, execer, state.Mines); err != nil {
		return err
	}
	if err := saveTech(ctx, execer, empireID, state.Tech); err != nil {
		return err
	}
	if err := saveMilitary(ctx, execer, state.Military); err != nil {
		return err
	}
	return nil
}

func (s *Store) LoadEconomy(ctx context.Context, empire game.Empire) (game.EconomyState, error) {
	if err := s.ensureEconomy(ctx, empire.ID); err != nil {
		return game.EconomyState{}, err
	}
	state := game.EconomyState{Empire: empire}
	regions, err := s.loadRegions(ctx, empire.ID)
	if err != nil {
		return game.EconomyState{}, err
	}
	buildings, err := s.loadBuildings(ctx, empire.ID)
	if err != nil {
		return game.EconomyState{}, err
	}
	mines, err := s.loadMines(ctx, empire.ID)
	if err != nil {
		return game.EconomyState{}, err
	}
	tech, err := s.loadTech(ctx, empire.ID)
	if err != nil {
		return game.EconomyState{}, err
	}
	military, err := s.loadMilitary(ctx, empire.ID)
	if err != nil {
		return game.EconomyState{}, err
	}
	state.Regions = regions
	state.Buildings = buildings
	state.Mines = mines
	state.Tech = tech
	state.Military = military
	return state, nil
}

func (s *Store) ensureEconomy(ctx context.Context, empireID string) error {
	var buildingsCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM empire_buildings WHERE empire_id = ?`, empireID).Scan(&buildingsCount); err != nil {
		return err
	}
	if buildingsCount == 0 {
		return insertStartingEconomy(ctx, s.db, empireID)
	}
	var militaryCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM empire_military WHERE empire_id = ?`, empireID).Scan(&militaryCount); err != nil {
		return err
	}
	if militaryCount == 0 {
		if err := saveMilitary(ctx, s.db, game.StartingMilitary(empireID)); err != nil {
			return err
		}
	}
	return ensureTechKeys(ctx, s.db, empireID)
}

func (s *Store) SaveEconomy(ctx context.Context, state game.EconomyState) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE empires SET
		turns_left = ?, turn_day = ?, money = ?, money_bank = ?, population = ?,
		food = ?, food_storage = ?, energy = ?, research_pts = ?, building_pts = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		state.Empire.TurnsLeft, state.Empire.TurnDay, state.Empire.Money, state.Empire.MoneyBank,
		state.Empire.Population, state.Empire.Food, state.Empire.FoodStorage, state.Empire.Energy,
		state.Empire.ResearchPts, state.Empire.BuildingPts, state.Empire.ID); err != nil {
		return err
	}
	if err := saveRegions(ctx, tx, state.Regions); err != nil {
		return err
	}
	if err := saveBuildings(ctx, tx, state.Buildings); err != nil {
		return err
	}
	if err := saveMines(ctx, tx, state.Mines); err != nil {
		return err
	}
	if err := saveTech(ctx, tx, state.Empire.ID, state.Tech); err != nil {
		return err
	}
	if err := saveMilitary(ctx, tx, state.Military); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) loadRegions(ctx context.Context, empireID string) (map[string]game.Region, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT empire_id, region_type, quantity, activated, activate_cost FROM empire_regions WHERE empire_id = ?`, empireID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	regions := make(map[string]game.Region)
	for rows.Next() {
		var region game.Region
		if err := rows.Scan(&region.EmpireID, &region.Type, &region.Quantity, &region.Activated, &region.ActivateCost); err != nil {
			return nil, err
		}
		regions[region.Type] = region
	}
	return regions, rows.Err()
}

func (s *Store) loadBuildings(ctx context.Context, empireID string) (game.Buildings, error) {
	var b game.Buildings
	var minersGuild, fishingGuild int
	err := s.db.QueryRowContext(ctx, `SELECT
		empire_id, miners_guild, miners_available, fishing_guild, fishers_assigned, fish_stock,
		construction_factories, research_labs, fossil_plants, fission_plants, fusion_plants
		FROM empire_buildings WHERE empire_id = ?`, empireID).Scan(
		&b.EmpireID, &minersGuild, &b.MinersAvailable, &fishingGuild, &b.FishersAssigned, &b.FishStock,
		&b.ConstructionFactories, &b.ResearchLabs, &b.FossilPlants, &b.FissionPlants, &b.FusionPlants,
	)
	if err != nil {
		return game.Buildings{}, err
	}
	b.MinersGuild = minersGuild != 0
	b.FishingGuild = fishingGuild != 0
	return b, nil
}

func (s *Store) loadMines(ctx context.Context, empireID string) (map[string]game.Mine, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT empire_id, mine_type, num_mines, miners_assigned, mineral_left, stored_minerals FROM empire_mines WHERE empire_id = ?`, empireID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mines := make(map[string]game.Mine)
	for rows.Next() {
		var mine game.Mine
		if err := rows.Scan(&mine.EmpireID, &mine.Type, &mine.NumMines, &mine.MinersAssigned, &mine.MineralLeft, &mine.StoredMinerals); err != nil {
			return nil, err
		}
		mines[mine.Type] = mine
	}
	return mines, rows.Err()
}

func (s *Store) loadTech(ctx context.Context, empireID string) (map[string]bool, error) {
	tech := game.StartingTech()
	rows, err := s.db.QueryContext(ctx, `SELECT tech_key, unlocked FROM empire_tech WHERE empire_id = ?`, empireID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var unlocked int
		if err := rows.Scan(&key, &unlocked); err != nil {
			return nil, err
		}
		tech[key] = unlocked != 0
	}
	return tech, rows.Err()
}

func (s *Store) loadMilitary(ctx context.Context, empireID string) (game.Military, error) {
	var m game.Military
	err := s.db.QueryRowContext(ctx, `SELECT
		empire_id, normal_soldiers, super_soldiers, mega_soldiers, tanks, hovercraft,
		nuclear_missiles, antimatter_missiles, recon_drones, spies, terrorists,
		ground_turrets, orbital_satellites, global_shields
		FROM empire_military WHERE empire_id = ?`, empireID).Scan(
		&m.EmpireID, &m.NormalSoldiers, &m.SuperSoldiers, &m.MegaSoldiers, &m.Tanks, &m.Hovercraft,
		&m.NuclearMissiles, &m.AntimatterMissiles, &m.ReconDrones, &m.Spies, &m.Terrorists,
		&m.GroundTurrets, &m.OrbitalSatellites, &m.GlobalShields,
	)
	if err != nil {
		return game.Military{}, err
	}
	return m, nil
}

func saveRegions(ctx context.Context, execer sqlExecer, regions map[string]game.Region) error {
	for _, region := range regions {
		if _, err := execer.ExecContext(ctx, `INSERT INTO empire_regions (empire_id, region_type, quantity, activated, activate_cost)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(empire_id, region_type) DO UPDATE SET quantity = excluded.quantity, activated = excluded.activated, activate_cost = excluded.activate_cost`,
			region.EmpireID, region.Type, region.Quantity, region.Activated, region.ActivateCost); err != nil {
			return fmt.Errorf("save region %s: %w", region.Type, err)
		}
	}
	return nil
}

func saveBuildings(ctx context.Context, execer sqlExecer, b game.Buildings) error {
	minersGuild := 0
	if b.MinersGuild {
		minersGuild = 1
	}
	fishingGuild := 0
	if b.FishingGuild {
		fishingGuild = 1
	}
	_, err := execer.ExecContext(ctx, `INSERT INTO empire_buildings (
		empire_id, miners_guild, miners_available, fishing_guild, fishers_assigned, fish_stock,
		construction_factories, research_labs, fossil_plants, fission_plants, fusion_plants
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(empire_id) DO UPDATE SET
		miners_guild = excluded.miners_guild,
		miners_available = excluded.miners_available,
		fishing_guild = excluded.fishing_guild,
		fishers_assigned = excluded.fishers_assigned,
		fish_stock = excluded.fish_stock,
		construction_factories = excluded.construction_factories,
		research_labs = excluded.research_labs,
		fossil_plants = excluded.fossil_plants,
		fission_plants = excluded.fission_plants,
		fusion_plants = excluded.fusion_plants`,
		b.EmpireID, minersGuild, b.MinersAvailable, fishingGuild, b.FishersAssigned, b.FishStock,
		b.ConstructionFactories, b.ResearchLabs, b.FossilPlants, b.FissionPlants, b.FusionPlants)
	return err
}

func saveMines(ctx context.Context, execer sqlExecer, mines map[string]game.Mine) error {
	for _, mine := range mines {
		if _, err := execer.ExecContext(ctx, `INSERT INTO empire_mines (empire_id, mine_type, num_mines, miners_assigned, mineral_left, stored_minerals)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(empire_id, mine_type) DO UPDATE SET
				num_mines = excluded.num_mines,
				miners_assigned = excluded.miners_assigned,
				mineral_left = excluded.mineral_left,
				stored_minerals = excluded.stored_minerals`,
			mine.EmpireID, mine.Type, mine.NumMines, mine.MinersAssigned, mine.MineralLeft, mine.StoredMinerals); err != nil {
			return fmt.Errorf("save mine %s: %w", mine.Type, err)
		}
	}
	return nil
}

func saveTech(ctx context.Context, execer sqlExecer, empireID string, tech map[string]bool) error {
	for _, key := range game.AllTechKeys() {
		unlocked := 0
		if tech[key] {
			unlocked = 1
		}
		if _, err := execer.ExecContext(ctx, `INSERT INTO empire_tech (empire_id, tech_key, unlocked)
			VALUES (?, ?, ?)
			ON CONFLICT(empire_id, tech_key) DO UPDATE SET unlocked = excluded.unlocked`,
			empireID, key, unlocked); err != nil {
			return fmt.Errorf("save tech %s: %w", key, err)
		}
	}
	return nil
}

func ensureTechKeys(ctx context.Context, execer sqlExecer, empireID string) error {
	defaults := game.StartingTech()
	for _, key := range game.AllTechKeys() {
		unlocked := 0
		if defaults[key] {
			unlocked = 1
		}
		if _, err := execer.ExecContext(ctx, `INSERT OR IGNORE INTO empire_tech (empire_id, tech_key, unlocked)
			VALUES (?, ?, ?)`, empireID, key, unlocked); err != nil {
			return fmt.Errorf("ensure tech %s: %w", key, err)
		}
	}
	return nil
}

func saveMilitary(ctx context.Context, execer sqlExecer, m game.Military) error {
	_, err := execer.ExecContext(ctx, `INSERT INTO empire_military (
		empire_id, normal_soldiers, super_soldiers, mega_soldiers, tanks, hovercraft,
		nuclear_missiles, antimatter_missiles, recon_drones, spies, terrorists,
		ground_turrets, orbital_satellites, global_shields
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(empire_id) DO UPDATE SET
		normal_soldiers = excluded.normal_soldiers,
		super_soldiers = excluded.super_soldiers,
		mega_soldiers = excluded.mega_soldiers,
		tanks = excluded.tanks,
		hovercraft = excluded.hovercraft,
		nuclear_missiles = excluded.nuclear_missiles,
		antimatter_missiles = excluded.antimatter_missiles,
		recon_drones = excluded.recon_drones,
		spies = excluded.spies,
		terrorists = excluded.terrorists,
		ground_turrets = excluded.ground_turrets,
		orbital_satellites = excluded.orbital_satellites,
		global_shields = excluded.global_shields`,
		m.EmpireID, m.NormalSoldiers, m.SuperSoldiers, m.MegaSoldiers, m.Tanks, m.Hovercraft,
		m.NuclearMissiles, m.AntimatterMissiles, m.ReconDrones, m.Spies, m.Terrorists,
		m.GroundTurrets, m.OrbitalSatellites, m.GlobalShields)
	return err
}
