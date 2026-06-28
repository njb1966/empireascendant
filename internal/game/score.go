package game

const (
	ScoreMilitaryPowerMultiplier = 10
	ScoreMoneyDivisor            = 10
	ScoreTechValue               = 500
)

func EmpireScore(state EconomyState) int {
	techScore := 0
	for _, unlocked := range state.Tech {
		if unlocked {
			techScore++
		}
	}
	totalMoney := state.Empire.Money + state.Empire.MoneyBank
	return state.Empire.Population +
		AttackPower(state.Military)*ScoreMilitaryPowerMultiplier +
		totalMoney/ScoreMoneyDivisor +
		techScore*ScoreTechValue
}
