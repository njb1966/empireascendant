package federation

import (
	"context"
	"errors"
	"time"

	"empireascendant/internal/config"
	"empireascendant/internal/data"
	"empireascendant/internal/interdoor"
)

type Store interface {
	SetFederationValue(context.Context, string, string) error
	FederationValue(context.Context, string) (string, error)
	SetHubCursor(context.Context, int64) error
	HubCursor(context.Context) (int64, error)
	OutboundEvents(context.Context, string, int) ([]interdoor.Event, error)
	MarkEventsPushed(context.Context, []interdoor.Event, int64) error
	StorePulledEvents(context.Context, []interdoor.FeedEvent) error
	ApplyPvPResolvedEvents(context.Context, string, []interdoor.FeedEvent, time.Time) (int, error)
	ResolveInboundPvP(context.Context, string, interdoor.PvPPending, time.Time) (interdoor.Event, bool, error)
	ImportTravelArrival(context.Context, string, interdoor.TravelPending, time.Time) (interdoor.Event, bool, error)
	LocalRosterEntries(context.Context, string, time.Time) ([]interdoor.RosterEntry, error)
	UpsertRemoteRoster(context.Context, []interdoor.RosterEntry) error
	LocalPlayerCount(context.Context) (int, error)
}

type Syncer struct {
	store  Store
	client interdoor.Client
	cfg    config.Config
	now    func() time.Time
	start  time.Time
}

type SyncResult struct {
	PushedEvents      int
	PulledEvents      int
	PushedRoster      int
	PulledRoster      int
	HubSeqHead        int64
	Pending           interdoor.PendingCounts
	AppliedPvPResults int
	ResolvedPvP       int
	BlockedPvP        int
	ArrivedTravel     int
	BlockedTravel     int
}

func New(store Store, client interdoor.Client, cfg config.Config) Syncer {
	now := time.Now
	return Syncer{
		store:  store,
		client: client,
		cfg:    cfg,
		now:    now,
		start:  now(),
	}
}

func (s Syncer) Register(ctx context.Context) (interdoor.RegisterResponse, error) {
	if s.cfg.NodeID == "" || s.cfg.HubURL == "" || s.cfg.RegistrationToken == "" {
		return interdoor.RegisterResponse{}, errors.New("node id, hub url, and registration token are required")
	}
	resp, err := s.client.Register(ctx, interdoor.RegisterRequest{
		NodeID:            s.cfg.NodeID,
		RegistrationToken: s.cfg.RegistrationToken,
		GameID:            s.cfg.GameID,
		GameTitle:         s.cfg.GameTitle,
		GameVersion:       s.cfg.GameVersion,
		ProtocolVersion:   s.cfg.ProtocolVersion,
		AdvertiseAddr:     s.cfg.AdvertiseAddr,
	})
	if err != nil {
		return interdoor.RegisterResponse{}, err
	}
	if err := s.store.SetFederationValue(ctx, data.StateAPIKey, resp.APIKey); err != nil {
		return interdoor.RegisterResponse{}, err
	}
	if err := s.store.SetHubCursor(ctx, resp.HubSeqHead); err != nil {
		return interdoor.RegisterResponse{}, err
	}
	return resp, nil
}

func (s Syncer) HeartbeatOnce(ctx context.Context) (interdoor.HeartbeatResponse, error) {
	playerCount, err := s.store.LocalPlayerCount(ctx)
	if err != nil {
		return interdoor.HeartbeatResponse{}, err
	}
	resp, err := s.client.Heartbeat(ctx, interdoor.HeartbeatRequest{
		NodeID:      s.cfg.NodeID,
		PlayerCount: playerCount,
		UptimeS:     int(s.now().Sub(s.start).Seconds()),
		GameVersion: s.cfg.GameVersion,
	})
	if err != nil {
		return interdoor.HeartbeatResponse{}, err
	}
	return resp, s.store.SetHubCursor(ctx, resp.HubSeqHead)
}

func (s Syncer) SyncOnce(ctx context.Context) (SyncResult, error) {
	var result SyncResult
	heartbeat, err := s.HeartbeatOnce(ctx)
	if err != nil {
		return result, err
	}
	result.HubSeqHead = heartbeat.HubSeqHead
	result.Pending = heartbeat.Pending

	pushed, head, err := s.pushOutboundEvents(ctx)
	if err != nil {
		return result, err
	}
	result.PushedEvents += pushed
	if head > 0 {
		result.HubSeqHead = head
	}

	roster, err := s.store.LocalRosterEntries(ctx, s.cfg.NodeID, s.now())
	if err != nil {
		return result, err
	}
	pushedRoster, err := s.client.PushRoster(ctx, roster)
	if err != nil {
		return result, err
	}
	result.PushedRoster = pushedRoster.Updated

	cursor, err := s.store.HubCursor(ctx)
	if err != nil {
		return result, err
	}
	pulledEvents, err := s.client.PullEvents(ctx, cursor, 500, true)
	if err != nil {
		return result, err
	}
	if err := s.store.StorePulledEvents(ctx, pulledEvents.Events); err != nil {
		return result, err
	}
	applied, err := s.store.ApplyPvPResolvedEvents(ctx, s.cfg.NodeID, pulledEvents.Events, s.now())
	if err != nil {
		return result, err
	}
	result.AppliedPvPResults = applied
	if err := s.store.SetHubCursor(ctx, pulledEvents.Head); err != nil {
		return result, err
	}
	result.PulledEvents = len(pulledEvents.Events)
	result.HubSeqHead = pulledEvents.Head

	pendingPvP, err := s.client.PendingPvP(ctx)
	if err != nil {
		return result, err
	}
	for _, req := range pendingPvP.Requests {
		_, complete, err := s.store.ResolveInboundPvP(ctx, s.cfg.NodeID, req, s.now())
		if err != nil {
			if errors.Is(err, data.ErrMalformedPvP) {
				if blockErr := s.client.BlockPvP(ctx, req.RequestID, err.Error()); blockErr != nil {
					return result, blockErr
				}
				result.BlockedPvP++
				continue
			}
			return result, err
		}
		if complete {
			if err := s.client.ResolvePvP(ctx, req.RequestID); err != nil {
				return result, err
			}
			result.ResolvedPvP++
		}
	}
	pushed, head, err = s.pushOutboundEvents(ctx)
	if err != nil {
		return result, err
	}
	result.PushedEvents += pushed
	if head > 0 {
		result.HubSeqHead = head
	}

	pendingTravel, err := s.client.PendingTravel(ctx)
	if err != nil {
		return result, err
	}
	for _, arrival := range pendingTravel.Arrivals {
		_, complete, err := s.store.ImportTravelArrival(ctx, s.cfg.NodeID, arrival, s.now())
		if err != nil {
			if errors.Is(err, data.ErrMalformedTravel) {
				if blockErr := s.client.BlockTravel(ctx, arrival.TravelID, err.Error()); blockErr != nil {
					return result, blockErr
				}
				result.BlockedTravel++
				continue
			}
			return result, err
		}
		if complete {
			if err := s.client.MarkTravelArrived(ctx, arrival.TravelID); err != nil {
				return result, err
			}
			result.ArrivedTravel++
		}
	}
	pushed, head, err = s.pushOutboundEvents(ctx)
	if err != nil {
		return result, err
	}
	result.PushedEvents += pushed
	if head > 0 {
		result.HubSeqHead = head
	}

	pulledRoster, err := s.client.PullRoster(ctx, true)
	if err != nil {
		return result, err
	}
	if err := s.store.UpsertRemoteRoster(ctx, pulledRoster.Entries); err != nil {
		return result, err
	}
	result.PulledRoster = len(pulledRoster.Entries)
	return result, nil
}

func (s Syncer) pushOutboundEvents(ctx context.Context) (int, int64, error) {
	events, err := s.store.OutboundEvents(ctx, s.cfg.NodeID, 100)
	if err != nil {
		return 0, 0, err
	}
	if len(events) == 0 {
		return 0, 0, nil
	}
	pushed, err := s.client.PushEvents(ctx, events)
	if err != nil {
		return 0, 0, err
	}
	if err := s.store.MarkEventsPushed(ctx, events, pushed.LastHubSeq); err != nil {
		return 0, 0, err
	}
	return len(events), pushed.LastHubSeq, nil
}
