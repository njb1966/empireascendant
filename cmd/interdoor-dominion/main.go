package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"empireascendant/internal/config"
	"empireascendant/internal/data"
	"empireascendant/internal/federation"
	"empireascendant/internal/interdoor"
	"empireascendant/internal/session"
	"empireascendant/internal/sshserver"
)

func main() {
	cfg := config.Default()
	var registerOnce bool
	var heartbeatOnce bool
	var syncOnce bool
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path")
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "SSH listen address")
	flag.BoolVar(&cfg.Stdio, "stdio", cfg.Stdio, "run local terminal review mode")
	flag.BoolVar(&cfg.ANSI, "ansi", cfg.ANSI, "enable ANSI color terminal presentation")
	flag.StringVar(&cfg.SSHEncoding, "ssh-encoding", cfg.SSHEncoding, "SSH terminal encoding: auto, cp437, or utf8")
	flag.StringVar(&cfg.SSHHostKeyPath, "ssh-host-key", cfg.SSHHostKeyPath, "SSH host key path")
	flag.StringVar(&cfg.NodeID, "node-id", cfg.NodeID, "InterDoor node id")
	flag.StringVar(&cfg.HubURL, "hub-url", cfg.HubURL, "InterDoor hub URL")
	flag.StringVar(&cfg.RegistrationToken, "registration-token", cfg.RegistrationToken, "InterDoor one-time registration token")
	flag.StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "InterDoor node API key")
	flag.StringVar(&cfg.AdvertiseAddr, "advertise-addr", cfg.AdvertiseAddr, "public InterDoor advertise address")
	flag.StringVar(&cfg.GameVersion, "game-version", cfg.GameVersion, "Empire Ascendant game version")
	flag.BoolVar(&registerOnce, "register", false, "register this node with the InterDoor hub")
	flag.BoolVar(&heartbeatOnce, "heartbeat-once", false, "send one InterDoor heartbeat")
	flag.BoolVar(&syncOnce, "sync-once", false, "run one InterDoor event and roster sync")
	flag.Parse()

	store, err := data.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := store.Init(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "initialize database: %v\n", err)
		os.Exit(1)
	}

	if cfg.APIKey == "" {
		apiKey, err := store.FederationValue(ctx, data.StateAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load stored api key: %v\n", err)
			os.Exit(1)
		}
		cfg.APIKey = apiKey
	}

	if registerOnce || heartbeatOnce || syncOnce {
		client := interdoor.NewClient(cfg.HubURL, cfg.APIKey)
		syncer := federation.New(store, client, cfg)
		switch {
		case registerOnce:
			resp, err := syncer.Register(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "register node: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "registered node %s, hub cursor %d\n", cfg.NodeID, resp.HubSeqHead)
		case heartbeatOnce:
			resp, err := syncer.HeartbeatOnce(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "heartbeat: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "heartbeat ok: hub_seq_head=%d pending_events=%d pending_pvp=%d pending_travel=%d\n",
				resp.HubSeqHead, resp.Pending.Events, resp.Pending.PVP, resp.Pending.Travel)
		case syncOnce:
			result, err := syncer.SyncOnce(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "sync: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "sync ok: pushed_events=%d pulled_events=%d pushed_roster=%d pulled_roster=%d applied_pvp_results=%d resolved_pvp=%d blocked_pvp=%d arrived_travel=%d blocked_travel=%d hub_seq_head=%d\n",
				result.PushedEvents, result.PulledEvents, result.PushedRoster, result.PulledRoster,
				result.AppliedPvPResults, result.ResolvedPvP, result.BlockedPvP,
				result.ArrivedTravel, result.BlockedTravel, result.HubSeqHead)
		}
		return
	}

	var pvpQueuer session.PvPQueuer
	if cfg.HubURL != "" && cfg.APIKey != "" {
		pvpQueuer = interdoor.NewClient(cfg.HubURL, cfg.APIKey)
	}
	var travelSubmitter session.TravelSubmitter
	if cfg.HubURL != "" && cfg.APIKey != "" {
		travelSubmitter = interdoor.NewClient(cfg.HubURL, cfg.APIKey)
	}
	runner := session.New(store, session.Options{NodeID: cfg.NodeID, PvPQueuer: pvpQueuer, TravelSubmitter: travelSubmitter, ANSI: cfg.ANSI})
	if !cfg.Stdio {
		server := sshserver.Server{Addr: cfg.Addr, HostKeyPath: cfg.SSHHostKeyPath, TerminalEncoding: cfg.SSHEncoding, Runner: runner}
		fmt.Fprintf(os.Stdout, "Empire Ascendant SSH listener on %s\n", cfg.Addr)
		if err := server.ListenAndServe(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "ssh listener error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := runner.Run(ctx, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "session error: %v\n", err)
		os.Exit(1)
	}
}
