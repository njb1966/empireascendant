package federation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"empireascendant/internal/config"
	"empireascendant/internal/data"
	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

func TestRegisterStoresAPIKeyAndCursor(t *testing.T) {
	store := newStore(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/register" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var req interdoor.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.NodeID != "ascendant" || req.RegistrationToken != "token" {
			t.Fatalf("register request = %+v", req)
		}
		_ = json.NewEncoder(w).Encode(interdoor.RegisterResponse{APIKey: "api-secret", HubSeqHead: 12})
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "ascendant"
	cfg.HubURL = server.URL
	cfg.RegistrationToken = "token"
	syncer := New(store, interdoor.NewClient(server.URL, ""), cfg)
	if _, err := syncer.Register(t.Context()); err != nil {
		t.Fatal(err)
	}
	apiKey, err := store.FederationValue(t.Context(), data.StateAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := store.HubCursor(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if apiKey != "api-secret" || cursor != 12 {
		t.Fatalf("api=%q cursor=%d", apiKey, cursor)
	}
}

func TestSyncOncePushesAndPullsEventsAndRoster(t *testing.T) {
	store := newStore(t)
	ctx := t.Context()
	_, err := store.CreateEmpire(ctx, data.CreateEmpireInput{
		Username:     "player1",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.EmitEvent(ctx, "ascendant", "empireascendant.empire_founded", map[string]any{
		"empire_name": "Solar Crown",
		"world_name":  "New Terra",
		"node":        "ascendant",
	}, time.Unix(100, 0)); err != nil {
		t.Fatal(err)
	}

	var sawEvents bool
	var sawRoster bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer api-secret" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			_ = json.NewEncoder(w).Encode(interdoor.HeartbeatResponse{HubSeqHead: 5})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/events":
			var req interdoor.PushEventsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Events) != 1 || req.Events[0].EventID != "ascendant:1" {
				t.Fatalf("events = %+v", req.Events)
			}
			sawEvents = true
			_ = json.NewEncoder(w).Encode(interdoor.PushEventsResponse{Accepted: 1, LastHubSeq: 10})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			var req interdoor.PushRosterRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Entries) != 1 || req.Entries[0].Name != "Solar Crown" {
				t.Fatalf("roster = %+v", req.Entries)
			}
			sawRoster = true
			_ = json.NewEncoder(w).Encode(interdoor.PushRosterResponse{Updated: len(req.Entries)})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			if r.URL.Query().Get("after") != "5" {
				t.Fatalf("after = %q", r.URL.Query().Get("after"))
			}
			_ = json.NewEncoder(w).Encode(interdoor.PullEventsResponse{Head: 11, Events: []interdoor.FeedEvent{{
				HubSeq: 11,
				Event: interdoor.Event{
					EventID:    "remote:1",
					SourceNode: "remote",
					Seq:        1,
					Type:       "empireascendant.galactic_news",
					TS:         101,
					Payload:    []byte(`{"headline":"Remote signal."}`),
				},
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingPvPResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingTravelResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PullRosterResponse{Entries: []interdoor.RosterEntry{{
				GlobalID: "remote:p_1",
				NodeID:   "remote",
				Name:     "Far Crown",
				Level:    120,
				Status:   "active",
				LastSeen: time.Unix(100, 0).Unix(),
			}}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "ascendant"
	cfg.HubURL = server.URL
	cfg.APIKey = "api-secret"
	result, err := New(store, interdoor.NewClient(server.URL, cfg.APIKey), cfg).SyncOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !sawEvents || !sawRoster {
		t.Fatalf("saw events=%t roster=%t", sawEvents, sawRoster)
	}
	if result.PushedEvents != 1 || result.PulledEvents != 1 || result.PushedRoster != 1 || result.PulledRoster != 1 || result.HubSeqHead != 11 {
		t.Fatalf("result = %+v", result)
	}
	news, err := store.News(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(news) != 2 || news[0].Headline != "Remote signal." {
		t.Fatalf("news = %+v", news)
	}
	remote, err := store.RemoteRoster(ctx, time.Unix(150, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(remote) != 1 || remote[0].EmpireName != "Far Crown" {
		t.Fatalf("remote roster = %+v", remote)
	}
}

func TestSyncOnceResolvesPendingPvPAndPushesResult(t *testing.T) {
	store := newStore(t)
	ctx := t.Context()
	victim, err := store.CreateEmpire(ctx, data.CreateEmpireInput{
		Username:     "victim",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	victimState, err := store.LoadEconomy(ctx, victim)
	if err != nil {
		t.Fatal(err)
	}
	victimState.Military.NormalSoldiers = 100
	if err := store.SaveEconomy(ctx, victimState); err != nil {
		t.Fatal(err)
	}

	attacker := game.NewStartingEconomy("remote-attacker")
	attacker.Empire.ID = "remote-attacker"
	attacker.Empire.WorldName = "Far World"
	attacker.Empire.EmpireName = "Far Crown"
	attacker.Military.NormalSoldiers = 1000
	payload, err := json.Marshal(game.NewRemoteAttackPayload(attacker, game.ActionGroundAttack, "", "remote:p_a", data.GlobalEmpireID("ascendant", victim.ID)))
	if err != nil {
		t.Fatal(err)
	}

	var completed bool
	var pushedResult bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer api-secret" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			_ = json.NewEncoder(w).Encode(interdoor.HeartbeatResponse{HubSeqHead: 5, Pending: interdoor.PendingCounts{PVP: 1}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PushRosterResponse{Updated: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			_ = json.NewEncoder(w).Encode(interdoor.PullEventsResponse{Head: 5})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingPvPResponse{Requests: []interdoor.PvPPending{{
				RequestID:  "req-1",
				AttackerID: "remote:p_a",
				VictimID:   data.GlobalEmpireID("ascendant", victim.ID),
				Attacker:   payload,
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/pvp/req-1/result":
			completed = true
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingTravelResponse{})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/events":
			var req interdoor.PushEventsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Events) != 1 || req.Events[0].Type != "pvp.resolved" {
				t.Fatalf("events = %+v", req.Events)
			}
			pushedResult = true
			_ = json.NewEncoder(w).Encode(interdoor.PushEventsResponse{Accepted: 1, LastHubSeq: 6})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PullRosterResponse{})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "ascendant"
	cfg.HubURL = server.URL
	cfg.APIKey = "api-secret"
	result, err := New(store, interdoor.NewClient(server.URL, cfg.APIKey), cfg).SyncOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result.ResolvedPvP != 1 || result.PushedEvents != 1 || !completed || !pushedResult {
		t.Fatalf("result=%+v completed=%t pushed=%t", result, completed, pushedResult)
	}
	got, err := store.LoadEconomy(ctx, victim)
	if err != nil {
		t.Fatal(err)
	}
	if got.Military.NormalSoldiers >= victimState.Military.NormalSoldiers {
		t.Fatalf("victim soldiers did not change: %+v", got.Military)
	}
}

func TestSyncOnceBlocksMalformedPvP(t *testing.T) {
	store := newStore(t)
	var blocked bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			_ = json.NewEncoder(w).Encode(interdoor.HeartbeatResponse{HubSeqHead: 5, Pending: interdoor.PendingCounts{PVP: 1}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PushRosterResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			_ = json.NewEncoder(w).Encode(interdoor.PullEventsResponse{Head: 5})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingPvPResponse{Requests: []interdoor.PvPPending{{
				RequestID:  "bad-1",
				AttackerID: "remote:p_a",
				VictimID:   "ascendant:p_missing",
				Attacker:   []byte(`{"action":"ground"}`),
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/pvp/bad-1/blocked":
			blocked = true
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingTravelResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PullRosterResponse{})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "ascendant"
	cfg.HubURL = server.URL
	cfg.APIKey = "api-secret"
	result, err := New(store, interdoor.NewClient(server.URL, cfg.APIKey), cfg).SyncOnce(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if result.BlockedPvP != 1 || !blocked {
		t.Fatalf("result=%+v blocked=%t", result, blocked)
	}
}

func TestSyncOnceImportsTravelArrivalAndPushesEvent(t *testing.T) {
	home := newStore(t)
	ctx := t.Context()
	empire, err := home.CreateEmpire(ctx, data.CreateEmpireInput{
		Username:     "traveler",
		PasswordHash: game.HashPassword("secret"),
		WorldName:    "New Terra",
		EmpireName:   "Solar Crown",
	})
	if err != nil {
		t.Fatal(err)
	}
	export, err := home.ExportTravelSnapshot(ctx, "home", empire.ID)
	if err != nil {
		t.Fatal(err)
	}

	dest := newStore(t)
	var arrived bool
	var pushedTravelEvent bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			_ = json.NewEncoder(w).Encode(interdoor.HeartbeatResponse{HubSeqHead: 5, Pending: interdoor.PendingCounts{Travel: 1}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PushRosterResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			_ = json.NewEncoder(w).Encode(interdoor.PullEventsResponse{Head: 5})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingPvPResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingTravelResponse{Arrivals: []interdoor.TravelPending{{
				TravelID: "travel-1",
				GlobalID: export.GlobalID,
				HomeNode: "home",
				FromNode: "home",
				DestNode: "remote",
				Snapshot: export.Snapshot,
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/travel/travel-1/arrived":
			arrived = true
		case r.Method == http.MethodPost && r.URL.Path == "/v1/events":
			var req interdoor.PushEventsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Events) != 1 || req.Events[0].Type != "player.traveled" {
				t.Fatalf("events = %+v", req.Events)
			}
			pushedTravelEvent = true
			_ = json.NewEncoder(w).Encode(interdoor.PushEventsResponse{Accepted: 1, LastHubSeq: 6})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PullRosterResponse{})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "remote"
	cfg.HubURL = server.URL
	cfg.APIKey = "api-secret"
	result, err := New(dest, interdoor.NewClient(server.URL, cfg.APIKey), cfg).SyncOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result.ArrivedTravel != 1 || result.PushedEvents != 1 || !arrived || !pushedTravelEvent {
		t.Fatalf("result=%+v arrived=%t pushed=%t", result, arrived, pushedTravelEvent)
	}
}

func TestSyncOnceBlocksMalformedTravel(t *testing.T) {
	store := newStore(t)
	var blocked bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			_ = json.NewEncoder(w).Encode(interdoor.HeartbeatResponse{HubSeqHead: 5, Pending: interdoor.PendingCounts{Travel: 1}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PushRosterResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			_ = json.NewEncoder(w).Encode(interdoor.PullEventsResponse{Head: 5})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingPvPResponse{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(interdoor.PendingTravelResponse{Arrivals: []interdoor.TravelPending{{
				TravelID: "bad-travel",
				GlobalID: "home:p_1",
				HomeNode: "home",
				FromNode: "home",
				DestNode: "remote",
				Snapshot: []byte(`{"version":1}`),
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/travel/bad-travel/blocked":
			blocked = true
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(interdoor.PullRosterResponse{})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	cfg := config.Default()
	cfg.NodeID = "remote"
	cfg.HubURL = server.URL
	cfg.APIKey = "api-secret"
	result, err := New(store, interdoor.NewClient(server.URL, cfg.APIKey), cfg).SyncOnce(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if result.BlockedTravel != 1 || !blocked {
		t.Fatalf("result=%+v blocked=%t", result, blocked)
	}
}

func newStore(t *testing.T) *data.Store {
	t.Helper()
	store, err := data.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Init(t.Context()); err != nil {
		t.Fatal(err)
	}
	return store
}
