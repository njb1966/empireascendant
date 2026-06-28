package interdoor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientRegisterHeartbeatEventsAndRoster(t *testing.T) {
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/register":
			var req RegisterRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.NodeID != "ascendant" || req.GameID != "empire_ascendant" || req.ProtocolVersion != ProtocolVersion {
				t.Fatalf("register request = %+v", req)
			}
			_ = json.NewEncoder(w).Encode(RegisterResponse{APIKey: "api-secret", HubSeqHead: 7})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/heartbeat":
			if r.Header.Get("Authorization") != "Bearer api-secret" {
				t.Fatalf("auth = %q", r.Header.Get("Authorization"))
			}
			sawAuth = true
			_ = json.NewEncoder(w).Encode(HeartbeatResponse{HubSeqHead: 8, Pending: PendingCounts{PVP: 1}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/events":
			var req PushEventsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Events) != 1 || req.Events[0].EventID != "ascendant:1" {
				t.Fatalf("events request = %+v", req)
			}
			_ = json.NewEncoder(w).Encode(PushEventsResponse{Accepted: 1, LastHubSeq: 9})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/events":
			if r.URL.Query().Get("after") != "8" || r.URL.Query().Get("exclude_self") != "true" {
				t.Fatalf("events query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(PullEventsResponse{Head: 9, Events: []FeedEvent{{
				HubSeq: 9,
				Event:  Event{EventID: "remote:1", SourceNode: "remote", Seq: 1, Type: "empireascendant.galactic_news", TS: 10, Payload: []byte(`{"headline":"remote"}`)},
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/roster":
			var req PushRosterRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if len(req.Entries) != 1 || req.Entries[0].GlobalID != "ascendant:p_1" {
				t.Fatalf("roster request = %+v", req)
			}
			_ = json.NewEncoder(w).Encode(PushRosterResponse{Updated: len(req.Entries)})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/roster":
			_ = json.NewEncoder(w).Encode(PullRosterResponse{Entries: []RosterEntry{{GlobalID: "remote:p_1", NodeID: "remote", Name: "Far Empire", Level: 100, Status: "active", LastSeen: 20}}})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	registerClient := NewClient(server.URL, "")
	reg, err := registerClient.Register(t.Context(), RegisterRequest{
		NodeID: "ascendant", RegistrationToken: "token", GameID: "empire_ascendant", ProtocolVersion: ProtocolVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	if reg.APIKey != "api-secret" || reg.HubSeqHead != 7 {
		t.Fatalf("register response = %+v", reg)
	}

	client := NewClient(server.URL, reg.APIKey)
	if _, err := client.Heartbeat(t.Context(), HeartbeatRequest{NodeID: "ascendant"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PushEvents(t.Context(), []Event{{EventID: "ascendant:1", SourceNode: "ascendant", Seq: 1, Type: "empireascendant.galactic_news", TS: 10, Payload: []byte(`{"headline":"local"}`)}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PullEvents(t.Context(), 8, 500, true); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PushRoster(t.Context(), []RosterEntry{{GlobalID: "ascendant:p_1", Name: "Solar Crown"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PullRoster(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	if !sawAuth {
		t.Fatal("heartbeat did not use bearer auth")
	}
}

func TestClientPvPEndpoints(t *testing.T) {
	var queued bool
	var resolved bool
	var blocked bool
	var travelSubmitted bool
	var travelArrived bool
	var travelBlocked bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer api-secret" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/pvp":
			var req PvPQueueRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.AttackerID != "ascendant:p_a" || req.VictimID != "remote:p_v" || string(req.Attacker) != `{"action":"ground"}` {
				t.Fatalf("queue request = %+v payload=%s", req, string(req.Attacker))
			}
			queued = true
			_ = json.NewEncoder(w).Encode(PvPQueueResponse{RequestID: "req-1", Status: "queued"})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/pvp/pending":
			_ = json.NewEncoder(w).Encode(PendingPvPResponse{Requests: []PvPPending{{
				RequestID:  "req-2",
				AttackerID: "remote:p_a",
				VictimID:   "ascendant:p_v",
				Attacker:   []byte(`{"action":"ground"}`),
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/pvp/req-2/result":
			resolved = true
		case r.Method == http.MethodPost && r.URL.Path == "/v1/pvp/req-3/blocked":
			var req BlockPvPRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Error == "" {
				t.Fatal("missing block reason")
			}
			blocked = true
		case r.Method == http.MethodPost && r.URL.Path == "/v1/travel":
			var req TravelSubmitRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.GlobalID != "ascendant:p_a" || req.HomeNode != "ascendant" || req.DestNode != "remote" || string(req.Snapshot) != `{"version":1}` {
				t.Fatalf("travel request = %+v snapshot=%s", req, string(req.Snapshot))
			}
			travelSubmitted = true
			_ = json.NewEncoder(w).Encode(TravelSubmitResponse{TravelID: "travel-1", Status: "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/travel/pending":
			_ = json.NewEncoder(w).Encode(PendingTravelResponse{Arrivals: []TravelPending{{
				TravelID: "travel-2",
				GlobalID: "remote:p_a",
				HomeNode: "remote",
				FromNode: "remote",
				DestNode: "ascendant",
				Snapshot: []byte(`{"version":1}`),
			}}})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/travel/travel-2/arrived":
			travelArrived = true
		case r.Method == http.MethodPost && r.URL.Path == "/v1/travel/travel-3/blocked":
			var req BlockTravelRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Error == "" {
				t.Fatal("missing travel block reason")
			}
			travelBlocked = true
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "api-secret")
	resp, err := client.QueuePvP(t.Context(), PvPQueueRequest{
		AttackerID: "ascendant:p_a",
		VictimID:   "remote:p_v",
		Attacker:   []byte(`{"action":"ground"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RequestID != "req-1" {
		t.Fatalf("queue response = %+v", resp)
	}
	pending, err := client.PendingPvP(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(pending.Requests) != 1 || pending.Requests[0].RequestID != "req-2" {
		t.Fatalf("pending = %+v", pending)
	}
	if err := client.ResolvePvP(t.Context(), "req-2"); err != nil {
		t.Fatal(err)
	}
	if err := client.BlockPvP(t.Context(), "req-3", "bad payload"); err != nil {
		t.Fatal(err)
	}
	travel, err := client.SubmitTravel(t.Context(), TravelSubmitRequest{
		GlobalID: "ascendant:p_a",
		HomeNode: "ascendant",
		DestNode: "remote",
		Snapshot: []byte(`{"version":1}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if travel.TravelID != "travel-1" {
		t.Fatalf("travel response = %+v", travel)
	}
	arrivals, err := client.PendingTravel(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(arrivals.Arrivals) != 1 || arrivals.Arrivals[0].TravelID != "travel-2" {
		t.Fatalf("arrivals = %+v", arrivals)
	}
	if err := client.MarkTravelArrived(t.Context(), "travel-2"); err != nil {
		t.Fatal(err)
	}
	if err := client.BlockTravel(t.Context(), "travel-3", "bad snapshot"); err != nil {
		t.Fatal(err)
	}
	if !queued || !resolved || !blocked || !travelSubmitted || !travelArrived || !travelBlocked {
		t.Fatalf("queued=%t resolved=%t blocked=%t travelSubmitted=%t travelArrived=%t travelBlocked=%t",
			queued, resolved, blocked, travelSubmitted, travelArrived, travelBlocked)
	}
}

func TestClientStructuredError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad token"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "").Heartbeat(t.Context(), HeartbeatRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "interdoor: status 401: bad token" {
		t.Fatalf("error = %q", got)
	}
}
