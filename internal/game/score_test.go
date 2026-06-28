package game

import "testing"

func TestEmpireScoreIncludesPopulationMilitaryMoneyAndTech(t *testing.T) {
	state := NewStartingEconomy("empire-1")
	state.Empire.Population = 100
	state.Empire.Money = 10000
	state.Military.NormalSoldiers = 10
	state.Tech = map[string]bool{
		TechEnergyFossil:  true,
		TechSoldierNormal: true,
	}
	want := 100 + 100 + 100 + 1000
	if got := EmpireScore(state); got != want {
		t.Fatalf("score = %d, want %d", got, want)
	}
}
