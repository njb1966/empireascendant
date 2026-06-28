package interdoor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type Error struct {
	StatusCode int
	Message    string
}

func (e Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("interdoor: status %d", e.StatusCode)
	}
	return fmt.Sprintf("interdoor: status %d: %s", e.StatusCode, e.Message)
}

func NewClient(baseURL, apiKey string) Client {
	return Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c Client) WithHTTPClient(client *http.Client) Client {
	c.httpClient = client
	return c
}

func (c Client) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	var resp RegisterResponse
	err := c.do(ctx, http.MethodPost, "/v1/register", "", req, &resp)
	return resp, err
}

func (c Client) Heartbeat(ctx context.Context, req HeartbeatRequest) (HeartbeatResponse, error) {
	var resp HeartbeatResponse
	err := c.do(ctx, http.MethodPost, "/v1/heartbeat", c.apiKey, req, &resp)
	return resp, err
}

func (c Client) PushEvents(ctx context.Context, events []Event) (PushEventsResponse, error) {
	var resp PushEventsResponse
	err := c.do(ctx, http.MethodPost, "/v1/events", c.apiKey, PushEventsRequest{Events: events}, &resp)
	return resp, err
}

func (c Client) PullEvents(ctx context.Context, after int64, limit int, excludeSelf bool) (PullEventsResponse, error) {
	values := url.Values{}
	values.Set("after", fmt.Sprintf("%d", after))
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	if excludeSelf {
		values.Set("exclude_self", "true")
	}
	var resp PullEventsResponse
	err := c.do(ctx, http.MethodGet, "/v1/events?"+values.Encode(), c.apiKey, nil, &resp)
	return resp, err
}

func (c Client) PushRoster(ctx context.Context, entries []RosterEntry) (PushRosterResponse, error) {
	var resp PushRosterResponse
	err := c.do(ctx, http.MethodPost, "/v1/roster", c.apiKey, PushRosterRequest{Entries: entries}, &resp)
	return resp, err
}

func (c Client) PullRoster(ctx context.Context, excludeSelf bool) (PullRosterResponse, error) {
	path := "/v1/roster"
	if excludeSelf {
		path += "?exclude_self=true"
	}
	var resp PullRosterResponse
	err := c.do(ctx, http.MethodGet, path, c.apiKey, nil, &resp)
	return resp, err
}

func (c Client) QueuePvP(ctx context.Context, req PvPQueueRequest) (PvPQueueResponse, error) {
	var resp PvPQueueResponse
	err := c.do(ctx, http.MethodPost, "/v1/pvp", c.apiKey, req, &resp)
	return resp, err
}

func (c Client) PendingPvP(ctx context.Context) (PendingPvPResponse, error) {
	var resp PendingPvPResponse
	err := c.do(ctx, http.MethodGet, "/v1/pvp/pending", c.apiKey, nil, &resp)
	return resp, err
}

func (c Client) ResolvePvP(ctx context.Context, requestID string) error {
	return c.do(ctx, http.MethodPost, "/v1/pvp/"+url.PathEscape(requestID)+"/result", c.apiKey, struct{}{}, nil)
}

func (c Client) BlockPvP(ctx context.Context, requestID, reason string) error {
	return c.do(ctx, http.MethodPost, "/v1/pvp/"+url.PathEscape(requestID)+"/blocked", c.apiKey, BlockPvPRequest{Error: reason}, nil)
}

func (c Client) SubmitTravel(ctx context.Context, req TravelSubmitRequest) (TravelSubmitResponse, error) {
	var resp TravelSubmitResponse
	err := c.do(ctx, http.MethodPost, "/v1/travel", c.apiKey, req, &resp)
	return resp, err
}

func (c Client) PendingTravel(ctx context.Context) (PendingTravelResponse, error) {
	var resp PendingTravelResponse
	err := c.do(ctx, http.MethodGet, "/v1/travel/pending", c.apiKey, nil, &resp)
	return resp, err
}

func (c Client) MarkTravelArrived(ctx context.Context, travelID string) error {
	return c.do(ctx, http.MethodPost, "/v1/travel/"+url.PathEscape(travelID)+"/arrived", c.apiKey, struct{}{}, nil)
}

func (c Client) BlockTravel(ctx context.Context, travelID, reason string) error {
	return c.do(ctx, http.MethodPost, "/v1/travel/"+url.PathEscape(travelID)+"/blocked", c.apiKey, BlockTravelRequest{Error: reason}, nil)
}

func (c Client) do(ctx context.Context, method, path, apiKey string, body any, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("interdoor: hub url is required")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func decodeError(resp *http.Response) error {
	var payload struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload.Error == "" {
		payload.Error = http.StatusText(resp.StatusCode)
	}
	return Error{StatusCode: resp.StatusCode, Message: payload.Error}
}
