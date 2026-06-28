package game

import "fmt"

type DefenseMultiplier func() float64

func FixedDefenseMultiplier(v float64) DefenseMultiplier {
	return func() float64 { return v }
}

func ResolveGroundAttack(attacker, defender *Combatant, multiplier DefenseMultiplier) CombatResult {
	attackPower := AttackPower(attacker.Military)
	defensePower := DefensePower(defender.Military)
	adjustedDefense := int(float64(defensePower) * multiplier())
	result := CombatResult{
		AttackPower:     attackPower,
		DefensePower:    defensePower,
		AdjustedDefense: adjustedDefense,
	}
	if attackPower > adjustedDefense {
		result.AttackerWon = true
		result.Loot = defender.Empire.Money / 4
		attacker.Empire.Money += result.Loot
		defender.Empire.Money -= result.Loot
		result.AttackerCasualties = attacker.Military.NormalSoldiers / 10
		result.DefenderCasualties = defender.Military.NormalSoldiers / 4
		attacker.Military.NormalSoldiers -= result.AttackerCasualties
		defender.Military.NormalSoldiers -= result.DefenderCasualties
		result.Message = fmt.Sprintf("%s defeated %s and looted %d.", attacker.Empire.EmpireName, defender.Empire.EmpireName, result.Loot)
		return result
	}
	result.AttackerCasualties = attacker.Military.NormalSoldiers
	result.DefenderCasualties = defender.Military.NormalSoldiers / 10
	attacker.Military.NormalSoldiers = 0
	defender.Military.NormalSoldiers -= result.DefenderCasualties
	result.Message = fmt.Sprintf("%s failed to conquer %s.", attacker.Empire.EmpireName, defender.Empire.EmpireName)
	return result
}

func ResolveGroundAttackState(attacker, defender *EconomyState, multiplier DefenseMultiplier) CombatResult {
	attackCombatant := Combatant{Empire: attacker.Empire, Military: attacker.Military}
	defenseCombatant := Combatant{Empire: defender.Empire, Military: defender.Military}
	result := ResolveGroundAttack(&attackCombatant, &defenseCombatant, multiplier)
	attacker.Empire = attackCombatant.Empire
	attacker.Military = attackCombatant.Military
	defender.Empire = defenseCombatant.Empire
	defender.Military = defenseCombatant.Military
	return result
}

func GroundAssault(attacker, defender *EconomyState, multiplier DefenseMultiplier) (CombatResult, error) {
	if AttackPower(attacker.Military) <= 0 {
		return CombatResult{}, ErrInsufficientUnits
	}
	if err := SpendTurn(&attacker.Empire); err != nil {
		return CombatResult{}, err
	}
	return ResolveGroundAttackState(attacker, defender, multiplier), nil
}
