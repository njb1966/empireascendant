package game

import "errors"

var (
	ErrInvalidUnit       = errors.New("invalid unit")
	ErrInvalidAction     = errors.New("invalid action")
	ErrAttackLimit       = errors.New("attack limit reached for this target today")
	ErrNoTarget          = errors.New("target not found")
	ErrInsufficientUnits = errors.New("not enough units")
)

type Military struct {
	EmpireID           string
	NormalSoldiers     int
	SuperSoldiers      int
	MegaSoldiers       int
	Tanks              int
	Hovercraft         int
	NuclearMissiles    int
	AntimatterMissiles int
	ReconDrones        int
	Spies              int
	Terrorists         int
	GroundTurrets      int
	OrbitalSatellites  int
	GlobalShields      int
}

type Combatant struct {
	Empire   Empire
	Military Military
}

type CombatResult struct {
	AttackerWon        bool
	AttackPower        int
	DefensePower       int
	AdjustedDefense    int
	Loot               int
	AttackerCasualties int
	DefenderCasualties int
	Message            string
}

type Dispatch struct {
	ID        string
	EmpireID  string
	Kind      string
	Message   string
	CreatedAt string
	ReadAt    string
}

func StartingMilitary(empireID string) Military {
	return Military{
		EmpireID:          empireID,
		NormalSoldiers:    DefaultNormalSoldiers,
		ReconDrones:       DefaultReconDrones,
		GroundTurrets:     DefaultGroundTurrets,
		OrbitalSatellites: DefaultOrbitalSatellites,
	}
}

func AddStartingMilitaryTech(tech map[string]bool) {
	tech[TechSoldierNormal] = true
	tech[TechSoldierSuper] = false
	tech[TechSoldierMega] = false
	tech[TechVehicleTank] = false
	tech[TechVehicleHovercraft] = false
	tech[TechBallisticNuclear] = false
	tech[TechBallisticAnti] = false
	tech[TechDefenseTurret] = false
	tech[TechDefenseSatellite] = false
	tech[TechDefenseGlobalShield] = false
	tech[TechEspionageIntel] = false
}

func AttackPower(m Military) int {
	return m.NormalSoldiers*NormalSoldierStrength +
		m.SuperSoldiers*SuperSoldierStrength +
		m.MegaSoldiers*MegaSoldierStrength +
		m.Tanks*TankStrength +
		m.Hovercraft*HovercraftStrength
}

func DefensePower(m Military) int {
	return AttackPower(m) +
		m.OrbitalSatellites*SatelliteDefenseStrength +
		m.GroundTurrets*TurretDefenseStrength +
		m.GlobalShields*ShieldDefenseStrength
}

func AttackLimitForAction(action string) (int, error) {
	switch action {
	case ActionGroundAttack:
		return GroundAttackLimit, nil
	case ActionMissile:
		return MissileLimit, nil
	case ActionSpy:
		return SpyLimit, nil
	default:
		return 0, ErrInvalidAction
	}
}
