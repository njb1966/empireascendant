package game

import "fmt"

type RemoteMilitarySnapshot struct {
	NormalSoldiers     int `json:"normal_soldiers"`
	SuperSoldiers      int `json:"super_soldiers"`
	MegaSoldiers       int `json:"mega_soldiers"`
	Tanks              int `json:"tanks"`
	Hovercraft         int `json:"hovercraft"`
	NuclearMissiles    int `json:"nuclear_missiles"`
	AntimatterMissiles int `json:"antimatter_missiles"`
	ReconDrones        int `json:"recon_drones"`
	Spies              int `json:"spies"`
	Terrorists         int `json:"terrorists"`
	GroundTurrets      int `json:"ground_turrets"`
	OrbitalSatellites  int `json:"orbital_satellites"`
	GlobalShields      int `json:"global_shields"`
}

type RemoteAttackPayload struct {
	Action             string                 `json:"action"`
	MissileType        string                 `json:"missile_type,omitempty"`
	AttackerGlobalID   string                 `json:"attacker_global_id"`
	VictimGlobalID     string                 `json:"victim_global_id"`
	AttackerEmpireName string                 `json:"attacker_empire_name"`
	AttackerWorldName  string                 `json:"attacker_world_name"`
	TurnDay            int64                  `json:"turn_day"`
	Military           RemoteMilitarySnapshot `json:"military"`
}

func SnapshotMilitary(m Military) RemoteMilitarySnapshot {
	return RemoteMilitarySnapshot{
		NormalSoldiers:     m.NormalSoldiers,
		SuperSoldiers:      m.SuperSoldiers,
		MegaSoldiers:       m.MegaSoldiers,
		Tanks:              m.Tanks,
		Hovercraft:         m.Hovercraft,
		NuclearMissiles:    m.NuclearMissiles,
		AntimatterMissiles: m.AntimatterMissiles,
		ReconDrones:        m.ReconDrones,
		Spies:              m.Spies,
		Terrorists:         m.Terrorists,
		GroundTurrets:      m.GroundTurrets,
		OrbitalSatellites:  m.OrbitalSatellites,
		GlobalShields:      m.GlobalShields,
	}
}

func (s RemoteMilitarySnapshot) Military(empireID string) Military {
	return Military{
		EmpireID:           empireID,
		NormalSoldiers:     s.NormalSoldiers,
		SuperSoldiers:      s.SuperSoldiers,
		MegaSoldiers:       s.MegaSoldiers,
		Tanks:              s.Tanks,
		Hovercraft:         s.Hovercraft,
		NuclearMissiles:    s.NuclearMissiles,
		AntimatterMissiles: s.AntimatterMissiles,
		ReconDrones:        s.ReconDrones,
		Spies:              s.Spies,
		Terrorists:         s.Terrorists,
		GroundTurrets:      s.GroundTurrets,
		OrbitalSatellites:  s.OrbitalSatellites,
		GlobalShields:      s.GlobalShields,
	}
}

func NewRemoteAttackPayload(state EconomyState, action, missileType, attackerGlobalID, victimGlobalID string) RemoteAttackPayload {
	return RemoteAttackPayload{
		Action:             action,
		MissileType:        missileType,
		AttackerGlobalID:   attackerGlobalID,
		VictimGlobalID:     victimGlobalID,
		AttackerEmpireName: state.Empire.EmpireName,
		AttackerWorldName:  state.Empire.WorldName,
		TurnDay:            state.Empire.TurnDay,
		Military:           SnapshotMilitary(state.Military),
	}
}

func PrepareRemoteGroundAttack(state *EconomyState) error {
	if AttackPower(state.Military) <= 0 {
		return ErrInsufficientUnits
	}
	return SpendTurn(&state.Empire)
}

func PrepareRemoteMissile(state *EconomyState, missile string) error {
	switch missile {
	case "Nuclear Missile":
		if !state.Tech[TechBallisticNuclear] {
			return ErrTechnologyUnavailable
		}
	case "Antimatter Missile":
		if !state.Tech[TechBallisticAnti] {
			return ErrTechnologyUnavailable
		}
	default:
		return ErrInvalidUnit
	}
	if err := SpendTurn(&state.Empire); err != nil {
		return err
	}
	switch missile {
	case "Nuclear Missile":
		if state.Military.NuclearMissiles <= 0 {
			state.Empire.TurnsLeft++
			return ErrInsufficientUnits
		}
		state.Military.NuclearMissiles--
	case "Antimatter Missile":
		if state.Military.AntimatterMissiles <= 0 {
			state.Empire.TurnsLeft++
			return ErrInsufficientUnits
		}
		state.Military.AntimatterMissiles--
	}
	return nil
}

func ResolveRemoteMissileStrike(attackerName string, defender *EconomyState, missile string) (CombatResult, error) {
	baseDamage := 0
	switch missile {
	case "Nuclear Missile":
		baseDamage = 100
	case "Antimatter Missile":
		baseDamage = 250
	default:
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
		Message:            fmt.Sprintf("%s hit %s with a %s.", attackerName, defender.Empire.EmpireName, missile),
	}, nil
}
