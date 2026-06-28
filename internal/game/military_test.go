package game

import (
	"errors"
	"strings"
	"testing"
)

func testMilitaryState(id, name string) EconomyState {
	state := NewStartingEconomy(id)
	state.Empire = Empire{
		ID:         id,
		EmpireName: name,
		WorldName:  name + " World",
		TurnsLeft:  DefaultTurns,
		TurnDay:    7,
		Money:      DefaultMoney,
	}
	return state
}

func TestGroundAssaultWinPersistsCasualtiesAndLoot(t *testing.T) {
	attacker := testMilitaryState("a", "Attacker")
	defender := testMilitaryState("d", "Defender")
	attacker.Military.NormalSoldiers = 1000
	defender.Military = Military{EmpireID: "d", NormalSoldiers: 100}

	result, err := GroundAssault(&attacker, &defender, FixedDefenseMultiplier(1.0))
	if err != nil {
		t.Fatal(err)
	}
	if !result.AttackerWon {
		t.Fatalf("expected attacker win: %+v", result)
	}
	if attacker.Empire.TurnsLeft != DefaultTurns-1 {
		t.Fatalf("turns = %d", attacker.Empire.TurnsLeft)
	}
	if attacker.Empire.Money != DefaultMoney+DefaultMoney/4 || defender.Empire.Money != DefaultMoney-DefaultMoney/4 {
		t.Fatalf("money attacker=%d defender=%d", attacker.Empire.Money, defender.Empire.Money)
	}
	if attacker.Military.NormalSoldiers != 900 || defender.Military.NormalSoldiers != 75 {
		t.Fatalf("military attacker=%+v defender=%+v", attacker.Military, defender.Military)
	}
}

func TestGroundAssaultLossRemovesAttackerSoldiers(t *testing.T) {
	attacker := testMilitaryState("a", "Attacker")
	defender := testMilitaryState("d", "Defender")

	result, err := GroundAssault(&attacker, &defender, FixedDefenseMultiplier(1.0))
	if err != nil {
		t.Fatal(err)
	}
	if result.AttackerWon {
		t.Fatalf("expected attacker loss: %+v", result)
	}
	if attacker.Military.NormalSoldiers != 0 {
		t.Fatalf("attacker soldiers = %d", attacker.Military.NormalSoldiers)
	}
}

func TestMilitaryActionsConsumeResources(t *testing.T) {
	state := testMilitaryState("a", "Attacker")
	if err := RecruitMilitaryUnit(&state, "Normal Soldier", 5); err != nil {
		t.Fatal(err)
	}
	if state.Military.NormalSoldiers != DefaultNormalSoldiers+5 || state.Empire.Money != DefaultMoney-5*RecruitNormalSoldierCost {
		t.Fatalf("state = %+v", state)
	}

	state.Empire.ResearchPts = ResearchNuclearCost
	if err := ResearchMilitaryTech(&state, TechBallisticNuclear); err != nil {
		t.Fatal(err)
	}
	if !state.Tech[TechBallisticNuclear] {
		t.Fatalf("nuclear tech not unlocked")
	}
	if err := BuyNuclearMissile(&state); err != nil {
		t.Fatal(err)
	}
	if state.Military.NuclearMissiles != 1 {
		t.Fatalf("missiles = %+v", state.Military)
	}
}

func TestMissileAndSpyActionsConsumeRequiredUnits(t *testing.T) {
	attacker := testMilitaryState("a", "Attacker")
	defender := testMilitaryState("d", "Defender")
	attacker.Tech[TechBallisticNuclear] = true
	attacker.Military.NuclearMissiles = 1

	result, err := FireNuclearMissile(&attacker, &defender)
	if err != nil {
		t.Fatal(err)
	}
	if result.DefenderCasualties == 0 || attacker.Military.NuclearMissiles != 0 {
		t.Fatalf("result=%+v attacker=%+v", result, attacker.Military)
	}

	attacker.Tech[TechEspionageIntel] = true
	attacker.Military.Spies = 1
	report, err := SpyScout(&attacker, &defender)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report, "Defender") || attacker.Military.Spies != 0 {
		t.Fatalf("report=%q spies=%d", report, attacker.Military.Spies)
	}
	_, err = SpyScout(&attacker, &defender)
	if !errors.Is(err, ErrInsufficientUnits) {
		t.Fatalf("err = %v", err)
	}
}
