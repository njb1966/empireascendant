package session

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"empireascendant/internal/data"
	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

func TestSessionCreateReportAndQuit(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nT\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{"EMPIRE ASCENDANT", "New World New Terra discovered.", "Empire Solar Crown founded.", "Empire Report", "Turns Left: 15", "Money: 10000", "Goodbye."} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionStoryInstructions(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("S\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{
		"Story",
		"How To Play",
		"Turns are the main daily limit.",
		"Build labs and factories",
		"Buy mines, hire miners",
		"Current Limits",
		"Visitor gameplay after travel is not implemented yet.",
		"Goodbye.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Not available yet.") {
		t.Fatalf("story still unavailable:\n%s", out)
	}
}

func TestSessionANSIMainMenuReadOnlyScreensPause(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("S\n\nR\n\nN\n\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Story / Instructions",
		"Turns limit daily action.",
		"Build labs/factories first",
		"Buy mines, hire miners",
		"RANKINGS -- TOP EMPIRES",
		"No empires ranked.",
		"GALACTIC NEWS",
		"No news.",
		"Press Enter",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionLoginUsernameIsCaseInsensitive(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nCalvusRex\nY\nsecret\nNew Terra\nSolar Crown\nQ\nE\ncalvusrex\nsecret\nT\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	if strings.Contains(out, "No empire found. Create one?") && strings.Count(out, "No empire found. Create one?") > 1 {
		t.Fatalf("second login offered create:\n%s", out)
	}
	if !strings.Contains(out, "Empire Report") || !strings.Contains(out, "Solar Crown") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestSessionLoginReactivatesHiddenEmpire(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	empire, err := store.CreateEmpire(context.Background(), data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
		Now:          time.Now().Add(-game.LifecycleHideAfter - time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nsecret\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	if !strings.Contains(out, "was hidden from public lists due to inactivity") {
		t.Fatalf("output:\n%s", out)
	}
	lifecycle, err := store.EmpireLifecycle(context.Background(), empire.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if lifecycle.Status != game.LifecycleActive {
		t.Fatalf("lifecycle = %+v", lifecycle)
	}
}

func TestSessionBankDoesNotSpendTurn(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nB\n1\n100\nQ\nT\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{"Bank updated: deposit 100.", "Turns Left: 15", "Money: 9900", "Bank: 100"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionInvalidBankChoiceDoesNotExit(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nB\nD\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{"Invalid bank action.", "EMPIRE HQ", "Goodbye."} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionANSIBankShowsStatusAndUpdates(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nB\n1\n100\n2\n50\nD\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Financial Status",
		"Bank updated: deposit 100.",
		"Bank updated: withdraw 50.",
		"Cash 9950",
		"Bank 50",
		"Invalid bank action.",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionANSIActivateRegionShowsUpdatedDevelopStatus(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nD\n1\n1\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{"Develop Menu", "Region", "Agricultural", "Agr 1/10", "Goodbye."} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "[1] Agricultural") {
		t.Fatalf("ANSI region picker still uses plain bracket list:\n%s", plain)
	}
}

func TestSessionANSIDevelopBlockedActionsShowNotice(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nD\n2\n3\n6\n1\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Action Result",
		"Build Structure blocked: not enough building points",
		"Sell Minerals blocked: no stored minerals to sell",
		"Develop Menu",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionANSIWorkerMenuShowsBlockedRequirements(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nD\n5\n1\nQ\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Worker Readiness",
		"Hire Miner blocked: build Miners Guild first.",
		"No  Build Miners Guild first (15 building)",
		"No  No available miners",
		"No  Build Fishing Guild first (15 building)",
		"Develop Menu",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionANSIWorkerHireUpdatesStatus(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	empire, err := store.CreateEmpire(context.Background(), data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.LoadEconomy(context.Background(), empire)
	if err != nil {
		t.Fatal(err)
	}
	state.Buildings.MinersGuild = true
	if err := store.SaveEconomy(context.Background(), state); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nsecret\nD\n5\n1\nQ\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Worker Readiness",
		"Miner hired. Available miners: 1.",
		"Yes Costs 100 money and 1 turn",
		"Workers   Miners 1 available",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionANSIReportUsesIndividualStatusScreens(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nT\n1\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Empire Status",
		"General Information",
		"Financial Information",
		"Army Status Information",
		"Your Command <1-8,Q>",
		"Empire : Solar Crown",
		"World  : New Terra",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "Region / Mine / Energy Information") {
		t.Fatalf("ANSI report still dumps combined report:\n%s", plain)
	}
}

func TestSessionANSIHQCommandsShowVisibleFeedback(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nA\n1\n1\n1\nQ\nI\n\nM\n\nW\n\nH\n\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Attack Menu",
		"Army Status",
		"Unit",
		"Normal Soldier",
		"Units recruited.",
		"Soldiers 101",
		"INTEL REPORT",
		"No dedicated intel reports are available yet.",
		"GALACTIC DISPATCHES",
		"No dispatches.",
		"WANDERERS",
		"No remote empires seen.",
		"HYPERDRIVE",
		"hyperdrive travel requires node id, hub url, and api key",
		"Returning to login menu.",
		"Returned from empire command to the login menu.",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionANSIAttackBlockedActionShowsNotice(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nA\n1\n5\n1\nQ\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{ANSI: true})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(output.String())
	for _, want := range []string{
		"Action Result",
		"Recruit Units blocked: technology is not available",
		"Army Status",
		"Hovercraft 0",
		"Goodbye.",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("output missing %q:\n%s", want, plain)
		}
	}
}

func TestSessionRecruitFromAttackMenu(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateEmpire(context.Background(), data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nsecret\nA\n1\n1\n5\nQ\nT\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store)
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{"ATTACK MENU", "Units recruited.", "Turns Left: 14", "Military: Soldiers 105"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionRankingsNewsAndWanderers(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertRemoteRoster(context.Background(), []interdoor.RosterEntry{{
		GlobalID: "remote:p_1",
		NodeID:   "remote",
		Name:     "Far Crown",
		Level:    123,
		Status:   "active",
		LastSeen: time.Now().Unix(),
	}}); err != nil {
		t.Fatal(err)
	}

	input := strings.NewReader("E\nplayer1\nY\nsecret\nNew Terra\nSolar Crown\nW\nQ\nR\nN\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{NodeID: "ascendant"})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	out := output.String()
	for _, want := range []string{"WANDERERS", "Far Crown", "RANKINGS -- TOP EMPIRES", "Solar Crown", "GALACTIC NEWS", "Solar Crown has risen on New Terra."} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSessionQueuesRemoteGroundAttack(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	empire, err := store.CreateEmpire(context.Background(), data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertRemoteRoster(context.Background(), []interdoor.RosterEntry{{
		GlobalID: "remote:p_1",
		NodeID:   "remote",
		Name:     "Far Crown",
		Level:    123,
		Status:   "active",
		LastSeen: time.Now().Unix(),
	}}); err != nil {
		t.Fatal(err)
	}

	queue := &fakePvPQueuer{requestID: "req-1"}
	input := strings.NewReader("E\nplayer1\nsecret\nA\n5\n1\nQ\nT\nQ\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{NodeID: "ascendant", PvPQueuer: queue})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	if queue.request.AttackerID != data.GlobalEmpireID("ascendant", empire.ID) || queue.request.VictimID != "remote:p_1" {
		t.Fatalf("queue request = %+v", queue.request)
	}
	empire, err = store.FindEmpireByID(context.Background(), empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.LoadEconomy(context.Background(), empire)
	if err != nil {
		t.Fatal(err)
	}
	if state.Empire.TurnsLeft != game.DefaultTurns-1 {
		t.Fatalf("turns = %d", state.Empire.TurnsLeft)
	}
	if !strings.Contains(output.String(), "Cross-node attack queued against Far Crown [remote]. Request req-1.") {
		t.Fatalf("output:\n%s", output.String())
	}
}

func TestSessionHyperdriveMarksEmpireAway(t *testing.T) {
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	empire, err := store.CreateEmpire(context.Background(), data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertRemoteRoster(context.Background(), []interdoor.RosterEntry{{
		GlobalID: "remote:p_1",
		NodeID:   "remote",
		Name:     "Far Crown",
		Level:    123,
		Status:   "active",
		LastSeen: time.Now().Unix(),
	}}); err != nil {
		t.Fatal(err)
	}

	travel := &fakeTravelSubmitter{travelID: "travel-1"}
	input := strings.NewReader("E\nplayer1\nsecret\nH\n1\nE\nplayer1\nsecret\nQ\n")
	var output bytes.Buffer
	runner := New(store, Options{NodeID: "ascendant", TravelSubmitter: travel})
	if err := runner.Run(context.Background(), input, &output); err != nil {
		t.Fatal(err)
	}
	if travel.request.GlobalID != data.GlobalEmpireID("ascendant", empire.ID) || travel.request.DestNode != "remote" || len(travel.request.Snapshot) == 0 {
		t.Fatalf("travel request = %+v", travel.request)
	}
	status, err := store.EmpireTravelStatus(context.Background(), empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != data.TravelStatusTraveling || status.DestNode != "remote" {
		t.Fatalf("status = %+v", status)
	}
	out := output.String()
	for _, want := range []string{"Solar Crown entered Hyperdrive toward remote. Travel travel-1 pending.", "Solar Crown is away from this node via Hyperdrive toward remote."} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

type fakePvPQueuer struct {
	requestID string
	request   interdoor.PvPQueueRequest
}

func (f *fakePvPQueuer) QueuePvP(_ context.Context, req interdoor.PvPQueueRequest) (interdoor.PvPQueueResponse, error) {
	f.request = req
	return interdoor.PvPQueueResponse{RequestID: f.requestID, Status: "queued"}, nil
}

type fakeTravelSubmitter struct {
	travelID string
	request  interdoor.TravelSubmitRequest
}

func (f *fakeTravelSubmitter) SubmitTravel(_ context.Context, req interdoor.TravelSubmitRequest) (interdoor.TravelSubmitResponse, error) {
	f.request = req
	return interdoor.TravelSubmitResponse{TravelID: f.travelID, Status: "pending"}, nil
}
