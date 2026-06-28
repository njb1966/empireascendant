package data

import (
	"encoding/json"
	"testing"
	"time"

	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

func TestResolveInboundPvPEmitsResultOnce(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	victim := createTestEmpire(t, s, "victim", "New Terra", "Solar Crown")
	state, err := s.LoadEconomy(ctx, victim)
	if err != nil {
		t.Fatal(err)
	}
	state.Military.NormalSoldiers = 100
	if err := s.SaveEconomy(ctx, state); err != nil {
		t.Fatal(err)
	}

	attacker := game.NewStartingEconomy("remote-attacker")
	attacker.Empire.ID = "remote-attacker"
	attacker.Empire.WorldName = "Far World"
	attacker.Empire.EmpireName = "Far Crown"
	attacker.Military.NormalSoldiers = 1000
	raw, err := json.Marshal(game.NewRemoteAttackPayload(attacker, game.ActionGroundAttack, "", "remote:p_a", GlobalEmpireID("ascendant", victim.ID)))
	if err != nil {
		t.Fatal(err)
	}

	event, complete, err := s.ResolveInboundPvP(ctx, "ascendant", interdoor.PvPPending{
		RequestID:  "req-1",
		AttackerID: "remote:p_a",
		VictimID:   GlobalEmpireID("ascendant", victim.ID),
		Attacker:   raw,
	}, time.Unix(500, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !complete || event.Type != "pvp.resolved" {
		t.Fatalf("complete=%t event=%+v", complete, event)
	}
	victim, err = s.FindEmpireByID(ctx, victim.ID)
	if err != nil {
		t.Fatal(err)
	}
	after, err := s.LoadEconomy(ctx, victim)
	if err != nil {
		t.Fatal(err)
	}
	if after.Military.NormalSoldiers >= state.Military.NormalSoldiers || after.Empire.Money >= state.Empire.Money {
		t.Fatalf("victim state did not change: before=%+v after=%+v", state.Military, after.Military)
	}

	event, complete, err = s.ResolveInboundPvP(ctx, "ascendant", interdoor.PvPPending{
		RequestID:  "req-1",
		AttackerID: "remote:p_a",
		VictimID:   GlobalEmpireID("ascendant", victim.ID),
		Attacker:   raw,
	}, time.Unix(501, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !complete || event.EventID != "" {
		t.Fatalf("duplicate complete=%t event=%+v", complete, event)
	}
	victim, err = s.FindEmpireByID(ctx, victim.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.LoadEconomy(ctx, victim)
	if err != nil {
		t.Fatal(err)
	}
	if again.Military.NormalSoldiers != after.Military.NormalSoldiers || again.Empire.Money != after.Empire.Money {
		t.Fatalf("duplicate changed victim: after=%+v again=%+v", after.Military, again.Military)
	}
}

func TestApplyPvPResolvedEventUpdatesAttackerOnce(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	attacker := createTestEmpire(t, s, "attacker", "New Terra", "Solar Crown")
	state, err := s.LoadEconomy(ctx, attacker)
	if err != nil {
		t.Fatal(err)
	}
	state.Military.NormalSoldiers = 100
	if err := s.SaveEconomy(ctx, state); err != nil {
		t.Fatal(err)
	}
	attackerGlobalID := GlobalEmpireID("ascendant", attacker.ID)
	if err := s.RecordOutboundPvP(ctx, "req-1", attacker.ID, attackerGlobalID, "remote:p_v", game.ActionGroundAttack, time.Unix(500, 0)); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"request_id":"req-1","attacker_global_id":"` + attackerGlobalID + `","victim_global_id":"remote:p_v","winner_global_id":"` + attackerGlobalID + `","result_text":"Solar Crown won.","resolved_at":501,"action":"ground","loot":250,"attacker_casualties":10}`)
	events := []interdoor.FeedEvent{{
		HubSeq: 10,
		Event: interdoor.Event{
			EventID:    "remote:1",
			SourceNode: "remote",
			Seq:        1,
			Type:       "pvp.resolved",
			TS:         501,
			Payload:    payload,
		},
	}}
	applied, err := s.ApplyPvPResolvedEvents(ctx, "ascendant", events, time.Unix(502, 0))
	if err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("applied = %d", applied)
	}
	attacker, err = s.FindEmpireByID(ctx, attacker.ID)
	if err != nil {
		t.Fatal(err)
	}
	after, err := s.LoadEconomy(ctx, attacker)
	if err != nil {
		t.Fatal(err)
	}
	if after.Empire.Money != state.Empire.Money+250 || after.Military.NormalSoldiers != 90 {
		t.Fatalf("attacker after = empire=%+v military=%+v", after.Empire, after.Military)
	}
	applied, err = s.ApplyPvPResolvedEvents(ctx, "ascendant", events, time.Unix(503, 0))
	if err != nil {
		t.Fatal(err)
	}
	if applied != 0 {
		t.Fatalf("duplicate applied = %d", applied)
	}
	attacker, err = s.FindEmpireByID(ctx, attacker.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.LoadEconomy(ctx, attacker)
	if err != nil {
		t.Fatal(err)
	}
	if again.Empire.Money != after.Empire.Money || again.Military.NormalSoldiers != after.Military.NormalSoldiers {
		t.Fatalf("duplicate changed attacker: after=%+v again=%+v", after.Military, again.Military)
	}
}
