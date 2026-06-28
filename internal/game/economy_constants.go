package game

const (
	RegionAgricultural = "Agricultural"
	RegionIndustrial   = "Industrial"
	RegionDesert       = "Desert"
	RegionUrban        = "Urban"
	RegionRiver        = "River"
	RegionOcean        = "Ocean"
	RegionVolcanic     = "Volcanic"
	RegionWasteland    = "Wasteland"
)

var RegionTypes = []string{
	RegionAgricultural,
	RegionIndustrial,
	RegionDesert,
	RegionUrban,
	RegionRiver,
	RegionOcean,
	RegionVolcanic,
	RegionWasteland,
}

const (
	MineGold   = "Gold"
	MineSilver = "Silver"
	MineIron   = "Iron"
	MineCopper = "Copper"
	MineNickel = "Nickel"
)

var MineTypes = []string{
	MineGold,
	MineSilver,
	MineIron,
	MineCopper,
	MineNickel,
}

const (
	TechEnergyFossil  = "energy_fossil"
	TechEnergyFission = "energy_fission"
	TechEnergyFusion  = "energy_fusion"
)

var D2TechKeys = []string{
	TechEnergyFossil,
	TechEnergyFission,
	TechEnergyFusion,
}

func AllTechKeys() []string {
	keys := make([]string, 0, len(D2TechKeys)+len(D3TechKeys))
	keys = append(keys, D2TechKeys...)
	keys = append(keys, D3TechKeys...)
	return keys
}

const (
	DefaultAgriculturalRegions = 10
	DefaultIndustrialRegions   = 1
	DefaultDesertRegions       = 0
	DefaultUrbanRegions        = 0
	DefaultRiverRegions        = 0
	DefaultOceanRegions        = 0
	DefaultVolcanicRegions     = 0
	DefaultWastelandRegions    = 0

	DefaultGoldMines   = 1
	DefaultSilverMines = 1
	DefaultIronMines   = 1
	DefaultCopperMines = 0
	DefaultNickelMines = 0

	DefaultConstructionFactories = 1
	DefaultResearchLabs          = 2
	DefaultFossilPlants          = 1

	DefaultMineralLeftPerMine = 1000
)

const (
	FoodPerAgriculturalRegion = 500
	FoodPerRiverRegion        = 300
	FoodPerFisher             = 50

	FossilEnergyPerPlant  = 100
	FissionEnergyPerPlant = 500
	FusionEnergyPerPlant  = 2000

	ResearchPerLab          = 10
	BuildingPerConstruction = 5
)

const (
	BuildResearchLabCost          = 20
	BuildConstructionFactoryCost  = 25
	BuildMinersGuildCost          = 15
	BuildFishingGuildCost         = 15
	ResearchFissionCost           = 50
	ResearchFusionCost            = 100
	BuyMineMoneyCost              = 1000
	HireWorkerMoneyCost           = 100
	ActivateRegionAdditionalScale = 10000
)

var MineralPrices = map[string]int{
	MineGold:   100,
	MineSilver: 50,
	MineIron:   20,
	MineCopper: 15,
	MineNickel: 10,
}
