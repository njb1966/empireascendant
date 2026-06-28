package game

import "testing"

func TestApplyDailyProduction(t *testing.T) {
	e := Empire{}
	ApplyDefaults(&e, 1)
	state := NewStartingEconomy("empire1")
	state.Empire = e
	ag := state.Regions[RegionAgricultural]
	ag.Activated = 5
	state.Regions[RegionAgricultural] = ag
	river := state.Regions[RegionRiver]
	river.Quantity = 1
	river.Activated = 1
	state.Regions[RegionRiver] = river
	state.Buildings.FishersAssigned = 2
	gold := state.Mines[MineGold]
	gold.MinersAssigned = 3
	state.Mines[MineGold] = gold

	result := ApplyDailyProduction(&state)

	if result.FoodProduced != 2900 {
		t.Fatalf("food produced = %d", result.FoodProduced)
	}
	if result.FoodConsumed != 2000 {
		t.Fatalf("food consumed = %d", result.FoodConsumed)
	}
	if state.Empire.Population != 20200 {
		t.Fatalf("population = %d", state.Empire.Population)
	}
	if state.Empire.Energy != 400 {
		t.Fatalf("energy = %d", state.Empire.Energy)
	}
	if state.Empire.ResearchPts != 20 {
		t.Fatalf("research = %d", state.Empire.ResearchPts)
	}
	if state.Empire.BuildingPts != 5 {
		t.Fatalf("building = %d", state.Empire.BuildingPts)
	}
	if state.Mines[MineGold].StoredMinerals != 3 || state.Mines[MineGold].MineralLeft != 997 {
		t.Fatalf("gold mine = %+v", state.Mines[MineGold])
	}
}

func TestApplyDailyProductionFoodDeficit(t *testing.T) {
	state := NewStartingEconomy("empire1")
	state.Empire = Empire{Population: 20000, Food: 100}
	result := ApplyDailyProduction(&state)
	if result.PopulationDiff != -100 {
		t.Fatalf("population diff = %d", result.PopulationDiff)
	}
	if state.Empire.Population != 19900 {
		t.Fatalf("population = %d", state.Empire.Population)
	}
	if state.Empire.Food != 0 {
		t.Fatalf("food = %d", state.Empire.Food)
	}
}
