package game

func EmpireScore(state EconomyState) int {
	techScore := 0
	for _, unlocked := range state.Tech {
		if unlocked {
			techScore++
		}
	}
	return state.Empire.Population +
		AttackPower(state.Military)*10 +
		state.Empire.Money/100 +
		techScore*500
}
