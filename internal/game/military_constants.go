package game

const (
	TechSoldierNormal       = "soldier_normal"
	TechSoldierSuper        = "soldier_super"
	TechSoldierMega         = "soldier_mega"
	TechVehicleTank         = "vehicle_tank"
	TechVehicleHovercraft   = "vehicle_hovercraft"
	TechBallisticNuclear    = "ballistic_nuclear"
	TechBallisticAnti       = "ballistic_antimatter"
	TechDefenseTurret       = "defense_turret"
	TechDefenseSatellite    = "defense_satellite"
	TechDefenseGlobalShield = "defense_global_shield"
	TechEspionageIntel      = "espionage_intelligence"
)

var D3TechKeys = []string{
	TechSoldierNormal,
	TechSoldierSuper,
	TechSoldierMega,
	TechVehicleTank,
	TechVehicleHovercraft,
	TechBallisticNuclear,
	TechBallisticAnti,
	TechDefenseTurret,
	TechDefenseSatellite,
	TechDefenseGlobalShield,
	TechEspionageIntel,
}

const (
	DefaultNormalSoldiers    = 100
	DefaultReconDrones       = 5
	DefaultOrbitalSatellites = 10
	DefaultGroundTurrets     = 10
)

const (
	NormalSoldierStrength    = 1
	SuperSoldierStrength     = 3
	MegaSoldierStrength      = 6
	TankStrength             = 20
	HovercraftStrength       = 35
	TurretDefenseStrength    = 20
	SatelliteDefenseStrength = 50
	ShieldDefenseStrength    = 200
)

const (
	RecruitNormalSoldierCost = 10
	RecruitSuperSoldierCost  = 30
	RecruitMegaSoldierCost   = 60
	RecruitTankCost          = 250
	RecruitHovercraftCost    = 500
	BuildTurretCost          = 15
	BuildSatelliteCost       = 25
	BuildShieldCost          = 50
	NuclearMissileCost       = 1000
	AntimatterMissileCost    = 2000
	SpyCost                  = 500
)

const (
	ResearchSuperSoldierCost = 75
	ResearchMegaSoldierCost  = 150
	ResearchTankCost         = 50
	ResearchHovercraftCost   = 100
	ResearchNuclearCost      = 100
	ResearchAntimatterCost   = 175
	ResearchTurretCost       = 40
	ResearchSatelliteCost    = 75
	ResearchShieldCost       = 150
	ResearchIntelCost        = 60
)

const (
	ActionGroundAttack = "ground"
	ActionMissile      = "missile"
	ActionSpy          = "spy"

	GroundAttackLimit = 3
	MissileLimit      = 2
	SpyLimit          = 1
)
