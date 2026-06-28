package game

import "time"

func UnixDay(t time.Time) int64 {
	return t.UTC().Unix() / 86400
}

func ApplyDailyTurnReset(e *Empire, now time.Time) bool {
	today := UnixDay(now)
	if e.TurnDay == today {
		return false
	}
	e.TurnDay = today
	e.TurnsLeft = DefaultTurns
	return true
}
