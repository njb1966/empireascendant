package game

func ApplyDailyProduction(state *EconomyState) ProductionResult {
	result := ProductionResult{MineralsMined: make(map[string]int)}

	agricultural := state.Regions[RegionAgricultural].Activated
	river := state.Regions[RegionRiver].Activated
	result.FoodProduced = agricultural*FoodPerAgriculturalRegion + river*FoodPerRiverRegion + state.Buildings.FishersAssigned*FoodPerFisher
	state.Empire.Food += result.FoodProduced

	result.FoodConsumed = state.Empire.Population / 10
	state.Empire.Food -= result.FoodConsumed
	if state.Empire.Food > 0 {
		result.PopulationDiff = state.Empire.Population / 100
		state.Empire.Population += result.PopulationDiff
	} else if state.Empire.Food < 0 {
		result.PopulationDiff = -(state.Empire.Population / 200)
		state.Empire.Population += result.PopulationDiff
		state.Empire.Food = 0
	}

	result.EnergyProduced = state.Buildings.FossilPlants*FossilEnergyPerPlant +
		state.Buildings.FissionPlants*FissionEnergyPerPlant +
		state.Buildings.FusionPlants*FusionEnergyPerPlant
	state.Empire.Energy += result.EnergyProduced

	result.ResearchGained = state.Buildings.ResearchLabs * ResearchPerLab
	state.Empire.ResearchPts += result.ResearchGained

	result.BuildingGained = state.Buildings.ConstructionFactories * BuildingPerConstruction
	state.Empire.BuildingPts += result.BuildingGained

	for mineType, mine := range state.Mines {
		mined := min(mine.MinersAssigned, mine.MineralLeft)
		if mined < 0 {
			mined = 0
		}
		mine.MineralLeft -= mined
		mine.StoredMinerals += mined
		state.Mines[mineType] = mine
		result.MineralsMined[mineType] = mined
	}

	return result
}
