package interdoor

import "encoding/json"

const ProtocolVersion = "1"

type RegisterRequest struct {
	NodeID            string `json:"node_id"`
	RegistrationToken string `json:"registration_token"`
	GameID            string `json:"game_id"`
	GameTitle         string `json:"game_title"`
	GameVersion       string `json:"game_version"`
	ProtocolVersion   string `json:"protocol_version"`
	AdvertiseAddr     string `json:"advertise_addr"`
}

type RegisterResponse struct {
	APIKey     string `json:"api_key"`
	HubSeqHead int64  `json:"hub_seq_head"`
}

type HeartbeatRequest struct {
	NodeID      string `json:"node_id"`
	PlayerCount int    `json:"player_count"`
	UptimeS     int    `json:"uptime_s"`
	GameVersion string `json:"game_version"`
}

type PendingCounts struct {
	Events int `json:"events"`
	PVP    int `json:"pvp"`
	Travel int `json:"travel"`
}

type HeartbeatResponse struct {
	HubSeqHead int64         `json:"hub_seq_head"`
	Pending    PendingCounts `json:"pending"`
}

type Event struct {
	EventID    string          `json:"event_id"`
	SourceNode string          `json:"source_node"`
	Seq        int64           `json:"seq"`
	Type       string          `json:"type"`
	TS         int64           `json:"ts"`
	Payload    json.RawMessage `json:"payload"`
}

type FeedEvent struct {
	HubSeq int64 `json:"hub_seq"`
	Event
}

type PushEventsRequest struct {
	Events []Event `json:"events"`
}

type PushEventsResponse struct {
	Accepted   int   `json:"accepted"`
	Duplicates int   `json:"duplicates"`
	LastHubSeq int64 `json:"last_hub_seq"`
}

type PullEventsResponse struct {
	Head   int64       `json:"head"`
	Events []FeedEvent `json:"events"`
}

type RosterEntry struct {
	GlobalID string `json:"global_id"`
	NodeID   string `json:"node_id,omitempty"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Status   string `json:"status"`
	LastSeen int64  `json:"last_seen"`
}

type PushRosterRequest struct {
	Entries []RosterEntry `json:"entries"`
}

type PushRosterResponse struct {
	Updated int `json:"updated"`
}

type PullRosterResponse struct {
	Entries []RosterEntry `json:"entries"`
}

type PvPQueueRequest struct {
	AttackerID string          `json:"attacker_id"`
	VictimID   string          `json:"victim_id"`
	Attacker   json.RawMessage `json:"attacker"`
}

type PvPQueueResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

type PvPPending struct {
	RequestID  string          `json:"request_id"`
	AttackerID string          `json:"attacker_id"`
	VictimID   string          `json:"victim_id"`
	VictimNode string          `json:"victim_node,omitempty"`
	Attacker   json.RawMessage `json:"attacker"`
	Status     string          `json:"status,omitempty"`
	Error      string          `json:"error,omitempty"`
	UpdatedAt  int64           `json:"updated_at,omitempty"`
	CreatedAt  int64           `json:"created_at,omitempty"`
}

type PendingPvPResponse struct {
	Requests []PvPPending `json:"requests"`
}

type BlockPvPRequest struct {
	Error string `json:"error"`
}

type TravelSubmitRequest struct {
	GlobalID string          `json:"global_id"`
	HomeNode string          `json:"home_node,omitempty"`
	DestNode string          `json:"dest_node"`
	Snapshot json.RawMessage `json:"snapshot"`
}

type TravelSubmitResponse struct {
	TravelID string `json:"travel_id"`
	Status   string `json:"status"`
}

type TravelPending struct {
	TravelID  string          `json:"travel_id"`
	GlobalID  string          `json:"global_id"`
	HomeNode  string          `json:"home_node"`
	FromNode  string          `json:"from_node"`
	DestNode  string          `json:"dest_node"`
	Snapshot  json.RawMessage `json:"snapshot"`
	Status    string          `json:"status,omitempty"`
	Error     string          `json:"error,omitempty"`
	UpdatedAt int64           `json:"updated_at,omitempty"`
	CreatedAt int64           `json:"created_at,omitempty"`
}

type PendingTravelResponse struct {
	Arrivals []TravelPending `json:"arrivals"`
}

type BlockTravelRequest struct {
	Error string `json:"error"`
}
