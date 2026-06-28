package game

import (
	"testing"
	"time"
)

func TestApplyDailyTurnReset(t *testing.T) {
	now := time.Unix(86400*10, 0)
	e := Empire{TurnsLeft: 3, TurnDay: UnixDay(now) - 1}
	if !ApplyDailyTurnReset(&e, now) {
		t.Fatal("expected reset")
	}
	if e.TurnsLeft != DefaultTurns {
		t.Fatalf("turns = %d", e.TurnsLeft)
	}
	if e.TurnDay != UnixDay(now) {
		t.Fatalf("turn day = %d", e.TurnDay)
	}
	if ApplyDailyTurnReset(&e, now) {
		t.Fatal("did not expect same-day reset")
	}
}
