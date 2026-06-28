package game

import (
	"strings"
	"testing"
)

func TestEmpireReportIncludesD1Resources(t *testing.T) {
	e := Empire{EmpireName: "Solar Crown", WorldName: "New Terra"}
	ApplyDefaults(&e, 1)
	state := NewStartingEconomy("empire1")
	state.Empire = e
	report := EmpireReport(state)
	for _, want := range []string{"Solar Crown", "New Terra", "Turns Left: 15", "Money: 10000", "Population: 20000", "Food: 10000", "Energy: 300", "Agricultural 0/10", "Labs 2"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report missing %q:\n%s", want, report)
		}
	}
}
