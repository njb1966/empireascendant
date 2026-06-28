package game

import (
	"errors"
	"fmt"
)

var (
	ErrInsufficientTurns     = errors.New("not enough turns")
	ErrInsufficientMoney     = errors.New("not enough money")
	ErrInsufficientBuilding  = errors.New("not enough building points")
	ErrInsufficientResearch  = errors.New("not enough research points")
	ErrInvalidRegion         = errors.New("invalid region type")
	ErrInvalidMine           = errors.New("invalid mine type")
	ErrInvalidTech           = errors.New("invalid tech")
	ErrNoAvailableRegion     = errors.New("no inactive region available")
	ErrNoMineralsToSell      = errors.New("no stored minerals to sell")
	ErrGuildRequired         = errors.New("required guild is not built")
	ErrTechnologyUnavailable = errors.New("technology is not available")
)

type Region struct {
	EmpireID     string
	Type         string
	Quantity     int
	Activated    int
	ActivateCost int
}

type Buildings struct {
	EmpireID              string
	MinersGuild           bool
	MinersAvailable       int
	FishingGuild          bool
	FishersAssigned       int
	FishStock             int
	ConstructionFactories int
	ResearchLabs          int
	FossilPlants          int
	FissionPlants         int
	FusionPlants          int
}

type Mine struct {
	EmpireID       string
	Type           string
	NumMines       int
	MinersAssigned int
	MineralLeft    int
	StoredMinerals int
}

type Tech struct {
	EmpireID string
	Key      string
	Unlocked bool
}

type EconomyState struct {
	Empire    Empire
	Regions   map[string]Region
	Buildings Buildings
	Mines     map[string]Mine
	Tech      map[string]bool
	Military  Military
}

type ProductionResult struct {
	FoodProduced   int
	FoodConsumed   int
	PopulationDiff int
	EnergyProduced int
	ResearchGained int
	BuildingGained int
	MineralsMined  map[string]int
}

type ActionResult struct {
	Message string
}

func NewStartingEconomy(empireID string) EconomyState {
	return EconomyState{
		Regions:   StartingRegions(empireID),
		Buildings: StartingBuildings(empireID),
		Mines:     StartingMines(empireID),
		Tech:      StartingTech(),
		Military:  StartingMilitary(empireID),
	}
}

func StartingRegions(empireID string) map[string]Region {
	quantities := map[string]int{
		RegionAgricultural: DefaultAgriculturalRegions,
		RegionIndustrial:   DefaultIndustrialRegions,
		RegionDesert:       DefaultDesertRegions,
		RegionUrban:        DefaultUrbanRegions,
		RegionRiver:        DefaultRiverRegions,
		RegionOcean:        DefaultOceanRegions,
		RegionVolcanic:     DefaultVolcanicRegions,
		RegionWasteland:    DefaultWastelandRegions,
	}
	regions := make(map[string]Region, len(RegionTypes))
	for _, regionType := range RegionTypes {
		quantity := quantities[regionType]
		regions[regionType] = Region{
			EmpireID:     empireID,
			Type:         regionType,
			Quantity:     quantity,
			Activated:    0,
			ActivateCost: RegionActivationCost(regionType, quantity),
		}
	}
	return regions
}

func StartingBuildings(empireID string) Buildings {
	return Buildings{
		EmpireID:              empireID,
		ConstructionFactories: DefaultConstructionFactories,
		ResearchLabs:          DefaultResearchLabs,
		FossilPlants:          DefaultFossilPlants,
	}
}

func StartingMines(empireID string) map[string]Mine {
	counts := map[string]int{
		MineGold:   DefaultGoldMines,
		MineSilver: DefaultSilverMines,
		MineIron:   DefaultIronMines,
		MineCopper: DefaultCopperMines,
		MineNickel: DefaultNickelMines,
	}
	mines := make(map[string]Mine, len(MineTypes))
	for _, mineType := range MineTypes {
		numMines := counts[mineType]
		mines[mineType] = Mine{
			EmpireID:    empireID,
			Type:        mineType,
			NumMines:    numMines,
			MineralLeft: numMines * DefaultMineralLeftPerMine,
		}
	}
	return mines
}

func StartingTech() map[string]bool {
	tech := map[string]bool{
		TechEnergyFossil:  true,
		TechEnergyFission: false,
		TechEnergyFusion:  false,
	}
	AddStartingMilitaryTech(tech)
	return tech
}

func RegionActivationCost(regionType string, quantity int) int {
	base := map[string]int{
		RegionAgricultural: 1000,
		RegionIndustrial:   1700,
		RegionDesert:       500,
		RegionUrban:        1500,
		RegionVolcanic:     1300,
		RegionRiver:        1300,
		RegionOcean:        1000,
		RegionWasteland:    0,
	}[regionType]
	if base == 0 {
		return 0
	}
	return base + (base*quantity)/ActivateRegionAdditionalScale
}

func RegionValid(regionType string) bool {
	for _, known := range RegionTypes {
		if known == regionType {
			return true
		}
	}
	return false
}

func MineValid(mineType string) bool {
	for _, known := range MineTypes {
		if known == mineType {
			return true
		}
	}
	return false
}

func SellValue(mine Mine) int {
	return mine.StoredMinerals * MineralPrices[mine.Type]
}

func RequireTurn(e Empire) error {
	if e.TurnsLeft <= 0 {
		return ErrInsufficientTurns
	}
	return nil
}

func SpendTurn(e *Empire) error {
	if err := RequireTurn(*e); err != nil {
		return err
	}
	e.TurnsLeft--
	return nil
}

func formatResourceDelta(label string, amount int) string {
	if amount >= 0 {
		return fmt.Sprintf("%s +%d", label, amount)
	}
	return fmt.Sprintf("%s %d", label, amount)
}
