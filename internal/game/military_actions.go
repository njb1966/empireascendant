package game

import "fmt"

func RecruitNormalSoldiers(state *EconomyState, count int) error {
	return RecruitMilitaryUnit(state, "Normal Soldier", count)
}

func RecruitMilitaryUnit(state *EconomyState, unit string, count int) error {
	if count <= 0 {
		return ErrInvalidUnit
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	cost := 0
	switch unit {
	case "Normal Soldier":
		cost = count * RecruitNormalSoldierCost
	case "Super Soldier":
		if !state.Tech[TechSoldierSuper] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = count * RecruitSuperSoldierCost
	case "Mega Soldier":
		if !state.Tech[TechSoldierMega] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = count * RecruitMegaSoldierCost
	case "Tank":
		if !state.Tech[TechVehicleTank] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = count * RecruitTankCost
	case "Hovercraft":
		if !state.Tech[TechVehicleHovercraft] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = count * RecruitHovercraftCost
	default:
		state.Empire.TurnsLeft++
		return ErrInvalidUnit
	}
	if state.Empire.Money < cost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= cost
	switch unit {
	case "Normal Soldier":
		state.Military.NormalSoldiers += count
	case "Super Soldier":
		state.Military.SuperSoldiers += count
	case "Mega Soldier":
		state.Military.MegaSoldiers += count
	case "Tank":
		state.Military.Tanks += count
	case "Hovercraft":
		state.Military.Hovercraft += count
	}
	return nil
}

func BuildGroundTurret(state *EconomyState) error {
	return BuildDefense(state, "Ground Turret")
}

func BuildDefense(state *EconomyState, defense string) error {
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	cost := 0
	switch defense {
	case "Ground Turret":
		if !state.Tech[TechDefenseTurret] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = BuildTurretCost
	case "Orbital Satellite":
		if !state.Tech[TechDefenseSatellite] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = BuildSatelliteCost
	case "Global Shield":
		if !state.Tech[TechDefenseGlobalShield] {
			state.Empire.TurnsLeft++
			return ErrTechnologyUnavailable
		}
		cost = BuildShieldCost
	default:
		state.Empire.TurnsLeft++
		return ErrInvalidUnit
	}
	if state.Empire.BuildingPts < cost {
		state.Empire.TurnsLeft++
		return ErrInsufficientBuilding
	}
	state.Empire.BuildingPts -= cost
	switch defense {
	case "Ground Turret":
		state.Military.GroundTurrets++
	case "Orbital Satellite":
		state.Military.OrbitalSatellites++
	case "Global Shield":
		state.Military.GlobalShields++
	}
	return nil
}

func ResearchMilitaryTech(state *EconomyState, tech string) error {
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	cost := map[string]int{
		TechSoldierSuper:        ResearchSuperSoldierCost,
		TechSoldierMega:         ResearchMegaSoldierCost,
		TechVehicleTank:         ResearchTankCost,
		TechVehicleHovercraft:   ResearchHovercraftCost,
		TechBallisticNuclear:    ResearchNuclearCost,
		TechBallisticAnti:       ResearchAntimatterCost,
		TechDefenseTurret:       ResearchTurretCost,
		TechDefenseSatellite:    ResearchSatelliteCost,
		TechDefenseGlobalShield: ResearchShieldCost,
		TechEspionageIntel:      ResearchIntelCost,
	}[tech]
	if cost == 0 {
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

func BuyNuclearMissile(state *EconomyState) error {
	return BuyMissile(state, "Nuclear Missile")
}

func BuyMissile(state *EconomyState, missile string) error {
	cost := 0
	switch missile {
	case "Nuclear Missile":
		if !state.Tech[TechBallisticNuclear] {
			return ErrTechnologyUnavailable
		}
		cost = NuclearMissileCost
	case "Antimatter Missile":
		if !state.Tech[TechBallisticAnti] {
			return ErrTechnologyUnavailable
		}
		cost = AntimatterMissileCost
	default:
		return ErrInvalidUnit
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < cost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= cost
	switch missile {
	case "Nuclear Missile":
		state.Military.NuclearMissiles++
	case "Antimatter Missile":
		state.Military.AntimatterMissiles++
	}
	return nil
}

func FireNuclearMissile(attacker, defender *EconomyState) (CombatResult, error) {
	if !attacker.Tech[TechBallisticNuclear] {
		return CombatResult{}, ErrTechnologyUnavailable
	}
	return fireMissile(attacker, defender, "Nuclear Missile", 100)
}

func FireAntimatterMissile(attacker, defender *EconomyState) (CombatResult, error) {
	if !attacker.Tech[TechBallisticAnti] {
		return CombatResult{}, ErrTechnologyUnavailable
	}
	return fireMissile(attacker, defender, "Antimatter Missile", 250)
}

func fireMissile(attacker, defender *EconomyState, missile string, baseDamage int) (CombatResult, error) {
	if err := SpendTurn(&attacker.Empire); err != nil {
		return CombatResult{}, err
	}
	switch missile {
	case "Nuclear Missile":
		if attacker.Military.NuclearMissiles <= 0 {
			attacker.Empire.TurnsLeft++
			return CombatResult{}, ErrInsufficientUnits
		}
		attacker.Military.NuclearMissiles--
	case "Antimatter Missile":
		if attacker.Military.AntimatterMissiles <= 0 {
			attacker.Empire.TurnsLeft++
			return CombatResult{}, ErrInsufficientUnits
		}
		attacker.Military.AntimatterMissiles--
	default:
		attacker.Empire.TurnsLeft++
		return CombatResult{}, ErrInvalidUnit
	}

	damage := baseDamage
	if defender.Military.GlobalShields > 0 {
		defender.Military.GlobalShields--
		damage /= 2
	}
	casualties := minInt(defender.Military.NormalSoldiers, damage)
	defender.Military.NormalSoldiers -= casualties
	moneyDamage := minInt(defender.Empire.Money/20, damage*2)
	defender.Empire.Money -= moneyDamage
	return CombatResult{
		AttackerWon:        true,
		AttackPower:        baseDamage,
		DefensePower:       DefensePower(defender.Military),
		DefenderCasualties: casualties,
		Loot:               moneyDamage,
		Message:            fmt.Sprintf("%s hit %s with a %s.", attacker.Empire.EmpireName, defender.Empire.EmpireName, missile),
	}, nil
}

func HireSpy(state *EconomyState) error {
	if !state.Tech[TechEspionageIntel] {
		return ErrTechnologyUnavailable
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	if state.Empire.Money < SpyCost {
		state.Empire.TurnsLeft++
		return ErrInsufficientMoney
	}
	state.Empire.Money -= SpyCost
	state.Military.Spies++
	return nil
}

func SpyScout(attacker, defender *EconomyState) (string, error) {
	if !attacker.Tech[TechEspionageIntel] {
		return "", ErrTechnologyUnavailable
	}
	if err := SpendTurn(&attacker.Empire); err != nil {
		return "", err
	}
	if attacker.Military.Spies <= 0 {
		attacker.Empire.TurnsLeft++
		return "", ErrInsufficientUnits
	}
	attacker.Military.Spies--
	return fmt.Sprintf("%s: soldiers %d, defenses %d, money %d.",
		defender.Empire.EmpireName,
		defender.Military.NormalSoldiers+defender.Military.SuperSoldiers+defender.Military.MegaSoldiers,
		DefensePower(defender.Military),
		defender.Empire.Money,
	), nil
}

func SpySabotage(attacker, defender *EconomyState) (CombatResult, error) {
	if !attacker.Tech[TechEspionageIntel] {
		return CombatResult{}, ErrTechnologyUnavailable
	}
	if err := SpendTurn(&attacker.Empire); err != nil {
		return CombatResult{}, err
	}
	if attacker.Military.Spies <= 0 {
		attacker.Empire.TurnsLeft++
		return CombatResult{}, ErrInsufficientUnits
	}
	attacker.Military.Spies--
	damage := minInt(defender.Empire.Money/10, 500)
	defender.Empire.Money -= damage
	return CombatResult{
		AttackerWon: true,
		Loot:        damage,
		Message:     fmt.Sprintf("%s sabotaged %s for %d money.", attacker.Empire.EmpireName, defender.Empire.EmpireName, damage),
	}, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
