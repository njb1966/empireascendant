package data

import (
	"testing"
	"time"

	"empireascendant/internal/interdoor"
)

func TestTravelSubmitMarksEmpireAway(t *testing.T) {
	s := newTestStore(t)
	ctx := t.Context()
	empire := createTestEmpire(t, s, "traveler", "New Terra", "Solar Crown")

	export, err := s.ExportTravelSnapshot(ctx, "ascendant", empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if export.GlobalID != GlobalEmpireID("ascendant", empire.ID) || len(export.Snapshot) == 0 {
		t.Fatalf("export = %+v", export)
	}
	if err := s.MarkTravelSubmitted(ctx, empire.ID, export.GlobalID, export.HomeNode, "remote", "travel-1", time.Unix(700, 0)); err != nil {
		t.Fatal(err)
	}
	status, err := s.EmpireTravelStatus(ctx, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != TravelStatusTraveling || status.DestNode != "remote" || status.TravelID != "travel-1" {
		t.Fatalf("status = %+v", status)
	}
}

func TestImportTravelArrivalStoresVisitorOnce(t *testing.T) {
	home := newTestStore(t)
	ctx := t.Context()
	empire := createTestEmpire(t, home, "traveler", "New Terra", "Solar Crown")
	export, err := home.ExportTravelSnapshot(ctx, "home", empire.ID)
	if err != nil {
		t.Fatal(err)
	}

	dest := newTestStore(t)
	arrival := interdoor.TravelPending{
		TravelID: "travel-1",
		GlobalID: export.GlobalID,
		HomeNode: "home",
		FromNode: "home",
		DestNode: "remote",
		Snapshot: export.Snapshot,
	}
	event, complete, err := dest.ImportTravelArrival(ctx, "remote", arrival, time.Unix(701, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !complete || event.Type != "player.traveled" {
		t.Fatalf("complete=%t event=%+v", complete, event)
	}
	event, complete, err = dest.ImportTravelArrival(ctx, "remote", arrival, time.Unix(702, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !complete || event.EventID != "" {
		t.Fatalf("duplicate complete=%t event=%+v", complete, event)
	}
}

func TestImportTravelArrivalRejectsMalformedSnapshot(t *testing.T) {
	s := newTestStore(t)
	_, _, err := s.ImportTravelArrival(t.Context(), "remote", interdoor.TravelPending{
		TravelID: "bad-1",
		GlobalID: "home:p_1",
		HomeNode: "home",
		FromNode: "home",
		DestNode: "remote",
		Snapshot: []byte(`{"version":1}`),
	}, time.Unix(703, 0))
	if err == nil {
		t.Fatal("expected malformed travel error")
	}
}
