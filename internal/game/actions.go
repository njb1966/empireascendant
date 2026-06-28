package game

import "fmt"

func Bank(e *Empire, amount int) error {
	if amount <= 0 {
		return ErrInsufficientMoney
	}
	if e.Money < amount {
		return ErrInsufficientMoney
	}
	e.Money -= amount
	e.MoneyBank += amount
	return nil
}

func Withdraw(e *Empire, amount int) error {
	if amount <= 0 || e.MoneyBank < amount {
		return ErrInsufficientMoney
	}
	e.MoneyBank -= amount
	e.Money += amount
	return nil
}

func ActivateRegion(state *EconomyState, regionType string) error {
	region, ok := state.Regions[regionType]
	if !ok || !RegionValid(regionType) || regionType == RegionWasteland {
		return ErrInvalidRegion
	}
	if region.Activated >= region.Quantity {
		return ErrNoAvailableRegion
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < region.ActivateCost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= region.ActivateCost
	region.Activated++
	region.ActivateCost = RegionActivationCost(region.Type, region.Quantity)
	state.Regions[regionType] = region
	return nil
}

func BuildStructure(state *EconomyState, structure string) error {
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	cost := 0
	switch structure {
	case "Research Lab":
		cost = BuildResearchLabCost
	case "Construction Factory":
		cost = BuildConstructionFactoryCost
	case "Miners Guild":
		cost = BuildMinersGuildCost
	case "Fishing Guild":
		cost = BuildFishingGuildCost
	default:
		state.Empire.TurnsLeft++
		return fmt.Errorf("unknown structure: %s", structure)
	}
	if state.Empire.BuildingPts < cost {
		state.Empire.TurnsLeft++
		return ErrInsufficientBuilding
	}
	state.Empire.BuildingPts -= cost
	switch structure {
	case "Research Lab":
		state.Buildings.ResearchLabs++
	case "Construction Factory":
		state.Buildings.ConstructionFactories++
	case "Miners Guild":
		state.Buildings.MinersGuild = true
	case "Fishing Guild":
		state.Buildings.FishingGuild = true
	}
	return nil
}

func ResearchEnergyTech(state *EconomyState, tech string) error {
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	cost := 0
	switch tech {
	case TechEnergyFission:
		cost = ResearchFissionCost
	case TechEnergyFusion:
		if !state.Tech[TechEnergyFission] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = ResearchFusionCost
	default:
		state.Empire.TurnsLeft++
		return ErrInvalidTech
	}
	if state.Empire.ResearchPts < cost {
		state.Empire.TurnsLeft++
		return ErrInsufficientResearch
	}
	state.Empire.ResearchPts -= cost
	state.Tech[tech] = true
	return nil
}

func BuyMine(state *EconomyState, mineType string) error {
	mine, ok := state.Mines[mineType]
	if !ok || !MineValid(mineType) {
		return ErrInvalidMine
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < BuyMineMoneyCost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= BuyMineMoneyCost
	mine.NumMines++
	mine.MineralLeft += DefaultMineralLeftPerMine
	state.Mines[mineType] = mine
	return nil
}

func HireMiner(state *EconomyState) error {
	if !state.Buildings.MinersGuild {
		return ErrGuildRequired
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < HireWorkerMoneyCost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= HireWorkerMoneyCost
	state.Buildings.MinersAvailable++
	return nil
}

func AssignMiner(state *EconomyState, mineType string) error {
	mine, ok := state.Mines[mineType]
	if !ok || !MineValid(mineType) {
		return ErrInvalidMine
	}
	if state.Buildings.MinersAvailable <= 0 {
		return ErrGuildRequired
	}
	state.Buildings.MinersAvailable--
	mine.MinersAssigned++
	state.Mines[mineType] = mine
	return nil
}

func HireFisher(state *EconomyState) error {
	if !state.Buildings.FishingGuild {
		return ErrGuildRequired
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < HireWorkerMoneyCost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= HireWorkerMoneyCost
	state.Buildings.FishersAssigned++
	return nil
}

func SellMinerals(state *EconomyState, mineType string) (int, error) {
	mine, ok := state.Mines[mineType]
	if !ok || !MineValid(mineType) {
		return 0, ErrInvalidMine
	}
	value := SellValue(mine)
	if value <= 0 {
		return 0, ErrNoMineralsToSell
	}
	state.Empire.Money += value
	mine.StoredMinerals = 0
	state.Mines[mineType] = mine
	return value, nil
}
