package game

import "testing"

func TestBankAndWithdrawAreFree(t *testing.T) {
	e := Empire{Money: 1000, MoneyBank: 100, TurnsLeft: 5}
	if err := Bank(&e, 250); err != nil {
		t.Fatal(err)
	}
	if e.Money != 750 || e.MoneyBank != 350 || e.TurnsLeft != 5 {
		t.Fatalf("after bank = %+v", e)
	}
	if err := Withdraw(&e, 50); err != nil {
		t.Fatal(err)
	}
	if e.Money != 800 || e.MoneyBank != 300 || e.TurnsLeft != 5 {
		t.Fatalf("after withdraw = %+v", e)
	}
}

func TestActivateRegionCostsOneTurn(t *testing.T) {
	state := NewStartingEconomy("empire1")
	state.Empire = Empire{Money: 10000, TurnsLeft: 3}
	if err := ActivateRegion(&state, RegionAgricultural); err != nil {
		t.Fatal(err)
	}
	if state.Empire.TurnsLeft != 2 {
		t.Fatalf("turns = %d", state.Empire.TurnsLeft)
	}
	if state.Regions[RegionAgricultural].Activated != 1 {
		t.Fatalf("region = %+v", state.Regions[RegionAgricultural])
	}
}

func TestSellMineralsIsFree(t *testing.T) {
	state := NewStartingEconomy("empire1")
	state.Empire = Empire{Money: 100, TurnsLeft: 4}
	gold := state.Mines[MineGold]
	gold.StoredMinerals = 2
	state.Mines[MineGold] = gold
	value, err := SellMinerals(&state, MineGold)
	if err != nil {
		t.Fatal(err)
	}
	if value != 200 || state.Empire.Money != 300 || state.Empire.TurnsLeft != 4 {
		t.Fatalf("value=%d state=%+v", value, state.Empire)
	}
}
