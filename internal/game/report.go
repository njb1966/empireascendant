package game

import "fmt"

func EmpireReport(state EconomyState) string {
	e := state.Empire
	return fmt.Sprintf(`Empire Report

Empire: %s
World: %s
Turns Left: %d

Money: %d
Bank: %d
Population: %d
Food: %d
Food Storage: %d
Energy: %d
Research Points: %d
Building Points: %d

Regions: Agricultural %d/%d, Industrial %d/%d, River %d/%d
Buildings: Labs %d, Construction Factories %d
Energy Plants: Fossil %d, Fission %d, Fusion %d
Mines: Gold %d, Silver %d, Iron %d, Copper %d, Nickel %d
Tech: Fossil %t, Fission %t, Fusion %t
Military: Soldiers %d, Tanks %d, Hovercraft %d
Defenses: Turrets %d, Satellites %d, Shields %d
Weapons: Nuclear %d, Antimatter %d, Spies %d
`,
		e.EmpireName,
		e.WorldName,
		e.TurnsLeft,
		e.Money,
		e.MoneyBank,
		e.Population,
		e.Food,
		e.FoodStorage,
		e.Energy,
		e.ResearchPts,
		e.BuildingPts,
		state.Regions[RegionAgricultural].Activated,
		state.Regions[RegionAgricultural].Quantity,
		state.Regions[RegionIndustrial].Activated,
		state.Regions[RegionIndustrial].Quantity,
		state.Regions[RegionRiver].Activated,
		state.Regions[RegionRiver].Quantity,
		state.Buildings.ResearchLabs,
		state.Buildings.ConstructionFactories,
		state.Buildings.FossilPlants,
		state.Buildings.FissionPlants,
		state.Buildings.FusionPlants,
		state.Mines[MineGold].NumMines,
		state.Mines[MineSilver].NumMines,
		state.Mines[MineIron].NumMines,
		state.Mines[MineCopper].NumMines,
		state.Mines[MineNickel].NumMines,
		state.Tech[TechEnergyFossil],
		state.Tech[TechEnergyFission],
		state.Tech[TechEnergyFusion],
		state.Military.NormalSoldiers+state.Military.SuperSoldiers+state.Military.MegaSoldiers,
		state.Military.Tanks,
		state.Military.Hovercraft,
		state.Military.GroundTurrets,
		state.Military.OrbitalSatellites,
		state.Military.GlobalShields,
		state.Military.NuclearMissiles,
		state.Military.AntimatterMissiles,
		state.Military.Spies,
	)
}
