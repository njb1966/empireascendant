package game

import "testing"

func TestEmpireScoreIncludesPopulationMilitaryMoneyAndTech(t *testing.T) {
	state := NewStartingEconomy("empire-1")
	state.Empire.Population = 100
	state.Empire.Money = 10000
	state.Empire.MoneyBank = 5000
	state.Military.NormalSoldiers = 10
	state.Tech = map[string]bool{
		TechEnergyFossil:  true,
		TechSoldierNormal: true,
	}
	want := 100 + 100 + 1500 + 1000
	if got := EmpireScore(state); got != want {
		t.Fatalf("score = %d, want %d", got, want)
	}
}

func TestEmpireScoreDoesNotPenalizeBanking(t *testing.T) {
	state := NewStartingEconomy("empire-1")
	state.Empire.Population = DefaultPopulation
	state.Empire.Money = DefaultMoney

	before := EmpireScore(state)
	if err := Bank(&state.Empire, 7500); err != nil {
		t.Fatal(err)
	}
	after := EmpireScore(state)
	if after != before {
		t.Fatalf("banking changed score: before=%d after=%d empire=%+v", before, after, state.Empire)
	}
}

func TestEmpireScoreRepresentativeStates(t *testing.T) {
	starting := NewStartingEconomy("starting")
	starting.Empire = Empire{Population: DefaultPopulation, Money: DefaultMoney}
	startingScore := EmpireScore(starting)

	builder := NewStartingEconomy("builder")
	builder.Empire = Empire{Population: 30000, Money: 250000, MoneyBank: 250000}
	builder.Buildings.ResearchLabs = 10
	builder.Buildings.ConstructionFactories = 8
	builder.Tech[TechEnergyFission] = true
	builder.Tech[TechEnergyFusion] = true
	builderScore := EmpireScore(builder)

	raider := NewStartingEconomy("raider")
	raider.Empire = Empire{Population: 18000, Money: 20000}
	raider.Military.NormalSoldiers = 1500
	raider.Military.Tanks = 40
	raider.Tech[TechVehicleTank] = true
	raiderScore := EmpireScore(raider)

	if startingScore != 23000 {
		t.Fatalf("starting score = %d", startingScore)
	}
	if builderScore <= startingScore {
		t.Fatalf("builder score %d should exceed starting score %d", builderScore, startingScore)
	}
	if raiderScore <= startingScore {
		t.Fatalf("raider score %d should exceed starting score %d", raiderScore, startingScore)
	}
	if builderScore <= raiderScore {
		t.Fatalf("builder score %d should exceed raider score %d so stored wealth remains leaderboard-relevant", builderScore, raiderScore)
	}
}
