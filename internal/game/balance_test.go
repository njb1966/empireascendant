package game

import "testing"

func TestBalanceEarlyMiningPathIsViable(t *testing.T) {
	state := NewStartingEconomy("empire-1")
	state.Empire = Empire{Money: DefaultMoney, Food: DefaultFood, Population: DefaultPopulation, TurnsLeft: DefaultTurns}

	if err := ActivateRegion(&state, RegionAgricultural); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		ApplyDailyProduction(&state)
	}
	if state.Empire.BuildingPts < BuildMinersGuildCost {
		t.Fatalf("building points after 3 production ticks = %d, want at least %d", state.Empire.BuildingPts, BuildMinersGuildCost)
	}

	if err := BuildStructure(&state, "Miners Guild"); err != nil {
		t.Fatal(err)
	}
	if err := HireMiner(&state); err != nil {
		t.Fatal(err)
	}
	if err := AssignMiner(&state, MineGold); err != nil {
		t.Fatal(err)
	}
	ApplyDailyProduction(&state)
	value, err := SellMinerals(&state, MineGold)
	if err != nil {
		t.Fatal(err)
	}
	if value != MineralPrices[MineGold] {
		t.Fatalf("first gold miner sale = %d, want %d", value, MineralPrices[MineGold])
	}
	if state.Empire.TurnsLeft != DefaultTurns-3 {
		t.Fatalf("turns left = %d, want %d", state.Empire.TurnsLeft, DefaultTurns-3)
	}
}

func TestBalanceCombatThresholdsRequireCommittedRecruitment(t *testing.T) {
	defender := NewStartingEconomy("defender")
	defender.Empire = Empire{EmpireName: "Defender", Money: DefaultMoney}

	modestAttacker := NewStartingEconomy("modest")
	modestAttacker.Empire = Empire{EmpireName: "Modest", Money: DefaultMoney, TurnsLeft: DefaultTurns}
	if err := RecruitMilitaryUnit(&modestAttacker, "Normal Soldier", 500); err != nil {
		t.Fatal(err)
	}
	modestResult, err := GroundAssault(&modestAttacker, &defender, FixedDefenseMultiplier(1.0))
	if err != nil {
		t.Fatal(err)
	}
	if modestResult.AttackerWon {
		t.Fatalf("modest attacker unexpectedly won: %+v", modestResult)
	}

	committedAttacker := NewStartingEconomy("committed")
	committedAttacker.Empire = Empire{EmpireName: "Committed", Money: DefaultMoney, TurnsLeft: DefaultTurns}
	if err := RecruitMilitaryUnit(&committedAttacker, "Normal Soldier", 800); err != nil {
		t.Fatal(err)
	}
	defender = NewStartingEconomy("defender")
	defender.Empire = Empire{EmpireName: "Defender", Money: DefaultMoney}
	committedResult, err := GroundAssault(&committedAttacker, &defender, FixedDefenseMultiplier(1.0))
	if err != nil {
		t.Fatal(err)
	}
	if !committedResult.AttackerWon {
		t.Fatalf("committed attacker should win against default defender: %+v", committedResult)
	}
	wantMoney := DefaultMoney - 800*RecruitNormalSoldierCost + DefaultMoney/4
	if committedAttacker.Empire.Money != wantMoney {
		t.Fatalf("committed attacker money = %d, want %d", committedAttacker.Empire.Money, wantMoney)
	}
	if committedAttacker.Military.NormalSoldiers >= DefaultNormalSoldiers+800 {
		t.Fatalf("committed attacker should take casualties: soldiers=%d", committedAttacker.Military.NormalSoldiers)
	}
}
