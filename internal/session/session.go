package session

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"empireascendant/internal/data"
	"empireascendant/internal/game"
	"empireascendant/internal/interdoor"
)

type Store interface {
	CreateEmpire(context.Context, data.CreateEmpireInput) (game.Empire, error)
	FindByUsername(context.Context, string) (game.Empire, error)
	ListTargetEmpires(context.Context, string) ([]game.Empire, error)
	LoadEconomy(context.Context, game.Empire) (game.EconomyState, error)
	SaveEconomy(context.Context, game.EconomyState) error
	UpdateTurns(context.Context, game.Empire) error
	EmitEvent(context.Context, string, string, map[string]any, time.Time) (interdoor.Event, error)
	LocalRankings(context.Context) ([]data.RankingEntry, error)
	RemoteRoster(context.Context, time.Time) ([]data.RankingEntry, error)
	News(context.Context, int) ([]data.NewsItem, error)
	CheckAttackLimit(context.Context, string, string, string, int) error
	RecordAttack(context.Context, string, string, string, int) error
	RecordOutboundPvP(context.Context, string, string, string, string, string, time.Time) error
	ExportTravelSnapshot(context.Context, string, string) (data.TravelSnapshotExport, error)
	MarkTravelSubmitted(context.Context, string, string, string, string, string, time.Time) error
	EmpireTravelStatus(context.Context, string) (data.TravelStatus, error)
	RefreshLifecycle(context.Context, string, time.Time) (data.Lifecycle, error)
	TouchEmpireLogin(context.Context, string, time.Time) error
	AddDispatch(context.Context, string, string, string) (game.Dispatch, error)
	ListDispatches(context.Context, string, int) ([]game.Dispatch, error)
}

type PvPQueuer interface {
	QueuePvP(context.Context, interdoor.PvPQueueRequest) (interdoor.PvPQueueResponse, error)
}

type TravelSubmitter interface {
	SubmitTravel(context.Context, interdoor.TravelSubmitRequest) (interdoor.TravelSubmitResponse, error)
}

var errRemotePvPUnavailable = errors.New("remote pvp requires node id, hub url, and api key")
var errTravelUnavailable = errors.New("hyperdrive travel requires node id, hub url, and api key")

type Runner struct {
	store           Store
	nodeID          string
	pvpQueuer       PvPQueuer
	travelSubmitter TravelSubmitter
	ansi            bool
}

type Options struct {
	NodeID          string
	PvPQueuer       PvPQueuer
	TravelSubmitter TravelSubmitter
	ANSI            bool
}

func New(store Store, options ...Options) Runner {
	runner := Runner{store: store}
	if len(options) > 0 {
		runner.nodeID = options[0].NodeID
		runner.pvpQueuer = options[0].PvPQueuer
		runner.travelSubmitter = options[0].TravelSubmitter
		runner.ansi = options[0].ANSI
	}
	return runner
}

func (r Runner) Run(ctx context.Context, input io.Reader, output io.Writer) error {
	s := &session{
		ctx:             ctx,
		in:              bufio.NewReader(input),
		out:             output,
		store:           r.store,
		now:             time.Now,
		nodeID:          r.nodeID,
		pvpQueuer:       r.pvpQueuer,
		travelSubmitter: r.travelSubmitter,
		presenter:       newPresenter(r.ansi),
	}
	return s.mainMenu()
}

type session struct {
	ctx             context.Context
	in              *bufio.Reader
	out             io.Writer
	store           Store
	now             func() time.Time
	nodeID          string
	pvpQueuer       PvPQueuer
	travelSubmitter TravelSubmitter
	presenter       presenter
	loginNotice     string
}

func (s *session) mainMenu() error {
	for {
		notice := s.loginNotice
		s.loginNotice = ""
		if notice != "" && s.presenter.ansi {
			s.presenter.section(s.out, "EMPIRE LINK CLOSED")
			fmt.Fprintln(s.out, notice)
		} else {
			s.presenter.title(s.out)
			if notice != "" {
				fmt.Fprintln(s.out, notice)
			}
		}
		s.presenter.menu(s.out, "MAIN MENU", []menuItem{
			{Key: "E", Label: "Enter Your Empire"},
			{Key: "R", Label: "Rankings -- Top Empires"},
			{Key: "N", Label: "Galactic News"},
			{Key: "S", Label: "Story / Instructions"},
			{Key: "Q", Label: "Quit"},
		})
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "E":
			if err := s.enterEmpire(); err != nil {
				return err
			}
		case "R":
			if err := s.rankingsMenu(); err != nil {
				return err
			}
			if err := s.pauseAfterANSIScreen(); err != nil {
				return err
			}
		case "N":
			if err := s.newsMenu(); err != nil {
				return err
			}
			if err := s.pauseAfterANSIScreen(); err != nil {
				return err
			}
		case "S":
			s.presenter.storyInstructions(s.out)
			if err := s.pauseAfterANSIScreen(); err != nil {
				return err
			}
		case "Q":
			fmt.Fprintln(s.out, "Goodbye.")
			return nil
		default:
			fmt.Fprintln(s.out, "Unknown command.")
		}
	}
}

func (s *session) enterEmpire() error {
	username, err := s.prompt("Username")
	if err != nil {
		return err
	}
	if username == "" {
		fmt.Fprintln(s.out, "Username is required.")
		return nil
	}

	e, err := s.store.FindByUsername(s.ctx, username)
	if err == nil {
		password, err := s.prompt("Password")
		if err != nil {
			return err
		}
		if !game.CheckPassword(e.PasswordHash, password) {
			fmt.Fprintln(s.out, "Invalid password.")
			return nil
		}
		lifecycle, err := s.store.RefreshLifecycle(s.ctx, e.ID, s.now())
		if err != nil {
			return err
		}
		printLifecycleNotice(s.out, e.EmpireName, lifecycle.Status)
		if err := s.store.TouchEmpireLogin(s.ctx, e.ID, s.now()); err != nil {
			return err
		}
		travel, err := s.store.EmpireTravelStatus(s.ctx, e.ID)
		if err != nil {
			return err
		}
		if travel.Status == data.TravelStatusTraveling || travel.Status == data.TravelStatusAway {
			fmt.Fprintf(s.out, "%s is away from this node via Hyperdrive toward %s.\n", e.EmpireName, travel.DestNode)
			return nil
		}
		state, err := s.store.LoadEconomy(s.ctx, e)
		if err != nil {
			return err
		}
		if game.ApplyDailyTurnReset(&state.Empire, s.now()) {
			result := game.ApplyDailyProduction(&state)
			if err := s.store.SaveEconomy(s.ctx, state); err != nil {
				return err
			}
			fmt.Fprintf(s.out, "Daily production complete: food +%d, energy +%d, research +%d, building +%d.\n",
				result.FoodProduced, result.EnergyProduced, result.ResearchGained, result.BuildingGained)
		}
		return s.empireHQ(state)
	}
	if !errors.Is(err, data.ErrNotFound) {
		return err
	}

	create, err := s.prompt("No empire found. Create one? [Y/n]")
	if err != nil {
		return err
	}
	if strings.EqualFold(create, "n") {
		return nil
	}

	password, err := s.prompt("Choose password")
	if err != nil {
		return err
	}
	world, err := s.prompt("Name Your World")
	if err != nil {
		return err
	}
	empireName, err := s.prompt("Name Your Empire")
	if err != nil {
		return err
	}

	e, err = s.store.CreateEmpire(s.ctx, data.CreateEmpireInput{
		Username:     username,
		PasswordHash: game.HashPassword(password),
		WorldName:    world,
		EmpireName:   empireName,
		Now:          s.now(),
	})
	if errors.Is(err, data.ErrDuplicateUser) || errors.Is(err, data.ErrDuplicateWorld) || errors.Is(err, data.ErrDuplicateEmpire) {
		fmt.Fprintln(s.out, err)
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(s.out, "New World %s discovered.\n", e.WorldName)
	fmt.Fprintf(s.out, "Empire %s founded.\n", e.EmpireName)
	if s.nodeID != "" {
		if _, err := s.store.EmitEvent(s.ctx, s.nodeID, "empireascendant.empire_founded", map[string]any{
			"global_id":   data.GlobalEmpireID(s.nodeID, e.ID),
			"empire_name": e.EmpireName,
			"world_name":  e.WorldName,
			"node":        s.nodeID,
			"created_at":  s.now().Unix(),
		}, s.now()); err != nil {
			return err
		}
	}
	state, err := s.store.LoadEconomy(s.ctx, e)
	if err != nil {
		return err
	}
	return s.empireHQ(state)
}

func printLifecycleNotice(out io.Writer, empireName, status string) {
	switch status {
	case game.LifecycleWarned:
		fmt.Fprintf(out, "%s has been quiet for a while. Your empire is active again.\n", empireName)
	case game.LifecycleHidden:
		fmt.Fprintf(out, "%s was hidden from public lists due to inactivity. Your empire is active again.\n", empireName)
	case game.LifecyclePurgeEligible:
		fmt.Fprintf(out, "%s was marked purge-eligible due to inactivity. Your empire is active again; no deletion was performed.\n", empireName)
	}
}

func (s *session) empireHQ(state game.EconomyState) error {
	for {
		s.presenter.menu(s.out, fmt.Sprintf("EMPIRE HQ [%d turns remaining]", state.Empire.TurnsLeft), []menuItem{
			{Key: "T", Label: "Your Empire Report"},
			{Key: "D", Label: "Develop Empire"},
			{Key: "B", Label: "Bank"},
			{Key: "A", Label: "Attack Menu"},
			{Key: "I", Label: "Intel Report"},
			{Key: "M", Label: "Galactic Dispatches"},
			{Key: "W", Label: "Wanderers"},
			{Key: "H", Label: "Hyperdrive"},
			{Key: "Q", Label: "Quit to Main"},
		})
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "T":
			if s.presenter.ansi {
				if err := s.reportMenu(state); err != nil {
					return err
				}
			} else {
				s.presenter.empireReport(s.out, state)
			}
		case "D":
			if err := s.developMenu(&state); err != nil {
				return err
			}
		case "B":
			if err := s.bankMenu(&state); err != nil {
				fmt.Fprintln(s.out, err)
			}
		case "A", "I", "M", "W", "H":
			switch strings.ToUpper(choice) {
			case "A":
				if err := s.attackMenu(&state); err != nil {
					return err
				}
			case "M":
				if err := s.dispatchMenu(state.Empire.ID); err != nil {
					return err
				}
				if err := s.pauseAfterANSIScreen(); err != nil {
					return err
				}
			case "W":
				if err := s.wanderersMenu(); err != nil {
					return err
				}
				if err := s.pauseAfterANSIScreen(); err != nil {
					return err
				}
			case "H":
				done, err := s.hyperdriveMenu(state)
				if err != nil {
					s.presenter.section(s.out, "HYPERDRIVE")
					fmt.Fprintln(s.out, err)
					if pauseErr := s.pauseAfterANSIScreen(); pauseErr != nil {
						return pauseErr
					}
				}
				if done {
					return nil
				}
			default:
				s.intelReport(state)
				if err := s.pauseAfterANSIScreen(); err != nil {
					return err
				}
			}
		case "Q":
			fmt.Fprintln(s.out, "Returning to login menu.")
			s.loginNotice = "Returned from empire command to the login menu."
			return nil
		default:
			fmt.Fprintln(s.out, "Unknown command.")
		}
	}
}

func (s *session) pauseAfterANSIScreen() error {
	if !s.presenter.ansi {
		return nil
	}
	_, err := s.prompt("Press Enter")
	return err
}

func (s *session) reportMenu(state game.EconomyState) error {
	items := []menuItem{
		{Key: "1", Label: "General Information"},
		{Key: "2", Label: "Financial Information"},
		{Key: "3", Label: "Army Status Information"},
		{Key: "4", Label: "Ballistic Weapons Information"},
		{Key: "5", Label: "Global Defence Information"},
		{Key: "6", Label: "Region Information"},
		{Key: "7", Label: "Mine Information"},
		{Key: "8", Label: "Empire Energy Information"},
	}
	for {
		s.presenter.reportMenu(s.out, items)
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "1":
			s.presenter.reportDetail(s.out, "General Information", generalReportRows(state), ansiCyan+ansiBright)
		case "2":
			s.presenter.reportDetail(s.out, "Financial Information", financialReportRows(state), ansiGreen+ansiBright)
		case "3":
			s.presenter.reportDetail(s.out, "Army Status Information", armyReportRows(state), ansiRed+ansiBright)
		case "4":
			s.presenter.reportDetail(s.out, "Ballistic Weapons Information", ballisticReportRows(state), ansiRed+ansiBright)
		case "5":
			s.presenter.reportDetail(s.out, "Global Defence Information", defenceReportRows(state), ansiGreen+ansiBright)
		case "6":
			s.presenter.reportDetail(s.out, "Region Information", regionReportRows(state), ansiGreen+ansiBright)
		case "7":
			s.presenter.reportDetail(s.out, "Mine Information", mineReportRows(state), ansiCyan+ansiBright)
		case "8":
			s.presenter.reportDetail(s.out, "Empire Energy Information", energyReportRows(state), ansiYellow+ansiBright)
		case "Q":
			return nil
		default:
			fmt.Fprintln(s.out, "Unknown report command.")
		}
	}
}

func (s *session) attackMenu(state *game.EconomyState) error {
	notice := ""
	for {
		items := []menuItem{
			{Key: "1", Label: "Recruit Units"},
			{Key: "2", Label: "Build Defenses"},
			{Key: "3", Label: "Research Military Tech"},
			{Key: "4", Label: "Buy Missile"},
			{Key: "5", Label: "Ground Assault"},
			{Key: "6", Label: "Ballistic Strike"},
			{Key: "7", Label: "Spy Mission"},
			{Key: "Q", Label: "Return"},
		}
		if s.presenter.ansi {
			s.presenter.attackMenu(s.out, *state, items, notice)
		} else {
			if notice != "" {
				fmt.Fprintln(s.out, notice)
			}
			s.presenter.menu(s.out, "ATTACK MENU", items)
		}
		notice = ""
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "1":
			if err := s.recruitUnits(state); err != nil {
				notice = "Recruit Units blocked: " + err.Error()
			} else {
				notice = "Units recruited."
			}
		case "2":
			if err := s.buildDefenses(state); err != nil {
				notice = "Build Defenses blocked: " + err.Error()
			} else {
				notice = "Defense built."
			}
		case "3":
			if err := s.researchMilitary(state); err != nil {
				notice = "Military Research blocked: " + err.Error()
			} else {
				notice = "Military research complete."
			}
		case "4":
			if err := s.buyMissile(state); err != nil {
				notice = "Buy Missile blocked: " + err.Error()
			} else {
				notice = "Missile purchased."
			}
		case "5":
			if err := s.groundAssault(state); err != nil {
				notice = "Ground Assault blocked: " + err.Error()
			}
		case "6":
			if err := s.ballisticStrike(state); err != nil {
				notice = "Ballistic Strike blocked: " + err.Error()
			}
		case "7":
			if err := s.spyMission(state); err != nil {
				notice = "Spy Mission blocked: " + err.Error()
			}
		case "Q":
			return nil
		default:
			notice = "Unknown attack command."
		}
	}
}

func (s *session) intelReport(state game.EconomyState) {
	s.presenter.section(s.out, "INTEL REPORT")
	fmt.Fprintf(s.out, "Empire: %s\n", state.Empire.EmpireName)
	fmt.Fprintln(s.out, "No dedicated intel reports are available yet.")
	fmt.Fprintln(s.out, "Use Wanderers for remote sightings and Galactic Dispatches for empire notices.")
}

func (s *session) recruitUnits(state *game.EconomyState) error {
	unit, err := s.choose("Unit", []string{"Normal Soldier", "Super Soldier", "Mega Soldier", "Tank", "Hovercraft"})
	if err != nil {
		return err
	}
	count, err := s.promptInt("Count")
	if err != nil {
		return err
	}
	if err := game.RecruitMilitaryUnit(state, unit, count); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Units recruited.")
	return nil
}

func (s *session) buildDefenses(state *game.EconomyState) error {
	defense, err := s.choose("Defense", []string{"Ground Turret", "Orbital Satellite", "Global Shield"})
	if err != nil {
		return err
	}
	if err := game.BuildDefense(state, defense); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Defense built.")
	return nil
}

func (s *session) researchMilitary(state *game.EconomyState) error {
	tech, err := s.choose("Military Tech", []string{
		game.TechSoldierSuper,
		game.TechSoldierMega,
		game.TechVehicleTank,
		game.TechVehicleHovercraft,
		game.TechBallisticNuclear,
		game.TechBallisticAnti,
		game.TechDefenseTurret,
		game.TechDefenseSatellite,
		game.TechDefenseGlobalShield,
		game.TechEspionageIntel,
	})
	if err != nil {
		return err
	}
	if err := game.ResearchMilitaryTech(state, tech); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Military research complete.")
	return nil
}

func (s *session) buyMissile(state *game.EconomyState) error {
	missile, err := s.choose("Missile", []string{"Nuclear Missile", "Antimatter Missile"})
	if err != nil {
		return err
	}
	if err := game.BuyMissile(state, missile); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Missile purchased.")
	return nil
}

func (s *session) groundAssault(state *game.EconomyState) error {
	target, err := s.chooseCombatTarget(state.Empire.ID)
	if err != nil {
		return err
	}
	if target.Remote {
		return s.queueRemoteGroundAssault(state, target.RemoteEntry)
	}
	defender, err := s.store.LoadEconomy(s.ctx, target.LocalEmpire)
	if err != nil {
		return err
	}
	if err := s.store.CheckAttackLimit(s.ctx, state.Empire.ID, target.LocalEmpire.ID, game.ActionGroundAttack, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	result, err := game.GroundAssault(state, &defender, game.FixedDefenseMultiplier(1.0))
	if err != nil {
		return err
	}
	if err := s.store.RecordAttack(s.ctx, state.Empire.ID, target.LocalEmpire.ID, game.ActionGroundAttack, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, defender); err != nil {
		return err
	}
	if _, err := s.store.AddDispatch(s.ctx, target.LocalEmpire.ID, "ground", result.Message); err != nil {
		return err
	}
	if s.nodeID != "" {
		if _, err := s.store.EmitEvent(s.ctx, s.nodeID, "empireascendant.attack_resolved", map[string]any{
			"attacker": state.Empire.EmpireName,
			"defender": defender.Empire.EmpireName,
			"outcome":  result.Message,
			"loot":     result.Loot,
			"message":  result.Message,
		}, s.now()); err != nil {
			return err
		}
	}
	fmt.Fprintln(s.out, result.Message)
	return nil
}

func (s *session) ballisticStrike(state *game.EconomyState) error {
	target, err := s.chooseCombatTarget(state.Empire.ID)
	if err != nil {
		return err
	}
	missile, err := s.choose("Missile", []string{"Nuclear Missile", "Antimatter Missile"})
	if err != nil {
		return err
	}
	if target.Remote {
		return s.queueRemoteMissileStrike(state, target.RemoteEntry, missile)
	}
	defender, err := s.store.LoadEconomy(s.ctx, target.LocalEmpire)
	if err != nil {
		return err
	}
	if err := s.store.CheckAttackLimit(s.ctx, state.Empire.ID, target.LocalEmpire.ID, game.ActionMissile, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	var result game.CombatResult
	switch missile {
	case "Nuclear Missile":
		result, err = game.FireNuclearMissile(state, &defender)
	case "Antimatter Missile":
		result, err = game.FireAntimatterMissile(state, &defender)
	}
	if err != nil {
		return err
	}
	if err := s.store.RecordAttack(s.ctx, state.Empire.ID, target.LocalEmpire.ID, game.ActionMissile, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, defender); err != nil {
		return err
	}
	if _, err := s.store.AddDispatch(s.ctx, target.LocalEmpire.ID, "missile", result.Message); err != nil {
		return err
	}
	if s.nodeID != "" {
		if _, err := s.store.EmitEvent(s.ctx, s.nodeID, "empireascendant.missile_strike", map[string]any{
			"attacker":     state.Empire.EmpireName,
			"target":       defender.Empire.EmpireName,
			"missile_type": missile,
			"damage":       result.DefenderCasualties,
			"message":      result.Message,
		}, s.now()); err != nil {
			return err
		}
	}
	fmt.Fprintln(s.out, result.Message)
	return nil
}

func (s *session) spyMission(state *game.EconomyState) error {
	target, defender, err := s.chooseTarget(state.Empire.ID)
	if err != nil {
		return err
	}
	mission, err := s.choose("Mission", []string{"Scout", "Sabotage"})
	if err != nil {
		return err
	}
	if err := s.store.CheckAttackLimit(s.ctx, state.Empire.ID, target.ID, game.ActionSpy, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	var message string
	switch mission {
	case "Scout":
		message, err = game.SpyScout(state, &defender)
	case "Sabotage":
		var result game.CombatResult
		result, err = game.SpySabotage(state, &defender)
		message = result.Message
	}
	if err != nil {
		return err
	}
	if err := s.store.RecordAttack(s.ctx, state.Empire.ID, target.ID, game.ActionSpy, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, defender); err != nil {
		return err
	}
	if _, err := s.store.AddDispatch(s.ctx, target.ID, "spy", fmt.Sprintf("Spy activity reported from %s.", state.Empire.EmpireName)); err != nil {
		return err
	}
	fmt.Fprintln(s.out, message)
	return nil
}

func (s *session) chooseTarget(attackerID string) (game.Empire, game.EconomyState, error) {
	targets, err := s.store.ListTargetEmpires(s.ctx, attackerID)
	if err != nil {
		return game.Empire{}, game.EconomyState{}, err
	}
	if len(targets) == 0 {
		return game.Empire{}, game.EconomyState{}, game.ErrNoTarget
	}
	options := make([]string, len(targets))
	for i, target := range targets {
		options[i] = fmt.Sprintf("%s / %s", target.EmpireName, target.WorldName)
	}
	selected, err := s.choose("Target", options)
	if err != nil {
		return game.Empire{}, game.EconomyState{}, err
	}
	for i, option := range options {
		if option == selected {
			state, err := s.store.LoadEconomy(s.ctx, targets[i])
			return targets[i], state, err
		}
	}
	return game.Empire{}, game.EconomyState{}, game.ErrNoTarget
}

type combatTarget struct {
	Remote      bool
	LocalEmpire game.Empire
	RemoteEntry data.RankingEntry
}

func (s *session) chooseCombatTarget(attackerID string) (combatTarget, error) {
	localTargets, err := s.store.ListTargetEmpires(s.ctx, attackerID)
	if err != nil {
		return combatTarget{}, err
	}
	remoteTargets, err := s.store.RemoteRoster(s.ctx, s.now())
	if err != nil {
		return combatTarget{}, err
	}
	options := make([]string, 0, len(localTargets)+len(remoteTargets))
	choices := make([]combatTarget, 0, len(localTargets)+len(remoteTargets))
	for _, target := range localTargets {
		options = append(options, fmt.Sprintf("%s / %s", target.EmpireName, target.WorldName))
		choices = append(choices, combatTarget{LocalEmpire: target})
	}
	for _, target := range remoteTargets {
		if target.GlobalID == "" {
			continue
		}
		stale := ""
		if target.Stale {
			stale = " stale"
		}
		options = append(options, fmt.Sprintf("%s [%s%s]", target.EmpireName, target.NodeID, stale))
		choices = append(choices, combatTarget{Remote: true, RemoteEntry: target})
	}
	if len(options) == 0 {
		return combatTarget{}, game.ErrNoTarget
	}
	selected, err := s.choose("Target", options)
	if err != nil {
		return combatTarget{}, err
	}
	for i, option := range options {
		if option == selected {
			return choices[i], nil
		}
	}
	return combatTarget{}, game.ErrNoTarget
}

func (s *session) queueRemoteGroundAssault(state *game.EconomyState, target data.RankingEntry) error {
	attackerGlobalID := data.GlobalEmpireID(s.nodeID, state.Empire.ID)
	if err := s.store.CheckAttackLimit(s.ctx, attackerGlobalID, target.GlobalID, game.ActionGroundAttack, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	queued := *state
	if err := game.PrepareRemoteGroundAttack(&queued); err != nil {
		return err
	}
	return s.queueRemotePvP(state, queued, target, game.ActionGroundAttack, "")
}

func (s *session) queueRemoteMissileStrike(state *game.EconomyState, target data.RankingEntry, missile string) error {
	attackerGlobalID := data.GlobalEmpireID(s.nodeID, state.Empire.ID)
	if err := s.store.CheckAttackLimit(s.ctx, attackerGlobalID, target.GlobalID, game.ActionMissile, int(state.Empire.TurnDay)); err != nil {
		return err
	}
	queued := *state
	if err := game.PrepareRemoteMissile(&queued, missile); err != nil {
		return err
	}
	return s.queueRemotePvP(state, queued, target, game.ActionMissile, missile)
}

func (s *session) queueRemotePvP(current *game.EconomyState, queued game.EconomyState, target data.RankingEntry, action, missile string) error {
	if s.nodeID == "" || s.pvpQueuer == nil {
		return errRemotePvPUnavailable
	}
	attackerGlobalID := data.GlobalEmpireID(s.nodeID, current.Empire.ID)
	payload := game.NewRemoteAttackPayload(*current, action, missile, attackerGlobalID, target.GlobalID)
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := s.pvpQueuer.QueuePvP(s.ctx, interdoor.PvPQueueRequest{
		AttackerID: attackerGlobalID,
		VictimID:   target.GlobalID,
		Attacker:   raw,
	})
	if err != nil {
		return err
	}
	if resp.RequestID == "" {
		return errors.New("hub returned empty pvp request id")
	}
	if err := s.store.RecordAttack(s.ctx, attackerGlobalID, target.GlobalID, action, int(current.Empire.TurnDay)); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, queued); err != nil {
		return err
	}
	if err := s.store.RecordOutboundPvP(s.ctx, resp.RequestID, current.Empire.ID, attackerGlobalID, target.GlobalID, action, s.now()); err != nil {
		return err
	}
	*current = queued
	fmt.Fprintf(s.out, "Cross-node attack queued against %s [%s]. Request %s.\n", target.EmpireName, target.NodeID, resp.RequestID)
	return nil
}

func (s *session) dispatchMenu(empireID string) error {
	dispatches, err := s.store.ListDispatches(s.ctx, empireID, 10)
	if err != nil {
		return err
	}
	s.presenter.section(s.out, "GALACTIC DISPATCHES")
	if len(dispatches) == 0 {
		fmt.Fprintln(s.out, "No dispatches.")
		return nil
	}
	for _, d := range dispatches {
		fmt.Fprintf(s.out, "%s: %s\n", strings.ToUpper(d.Kind), d.Message)
	}
	return nil
}

func (s *session) rankingsMenu() error {
	local, err := s.store.LocalRankings(s.ctx)
	if err != nil {
		return err
	}
	remote, err := s.store.RemoteRoster(s.ctx, s.now())
	if err != nil {
		return err
	}
	s.presenter.section(s.out, "RANKINGS -- TOP EMPIRES")
	if len(local) == 0 && len(remote) == 0 {
		fmt.Fprintln(s.out, "No empires ranked.")
		return nil
	}
	rank := 1
	for _, entry := range local {
		fmt.Fprintf(s.out, "%d. %s / %s -- %d\n", rank, entry.EmpireName, entry.WorldName, entry.Score)
		rank++
	}
	for _, entry := range remote {
		stale := ""
		if entry.Stale {
			stale = " stale"
		}
		fmt.Fprintf(s.out, "%d. %s [%s%s] -- %d\n", rank, entry.EmpireName, entry.NodeID, stale, entry.Score)
		rank++
	}
	return nil
}

func (s *session) newsMenu() error {
	items, err := s.store.News(s.ctx, 10)
	if err != nil {
		return err
	}
	s.presenter.section(s.out, "GALACTIC NEWS")
	if len(items) == 0 {
		fmt.Fprintln(s.out, "No news.")
		return nil
	}
	for _, item := range items {
		fmt.Fprintf(s.out, "[%s] %s\n", item.SourceNode, item.Headline)
	}
	return nil
}

func (s *session) wanderersMenu() error {
	entries, err := s.store.RemoteRoster(s.ctx, s.now())
	if err != nil {
		return err
	}
	s.presenter.section(s.out, "WANDERERS")
	if len(entries) == 0 {
		fmt.Fprintln(s.out, "No remote empires seen.")
		return nil
	}
	for _, entry := range entries {
		stale := ""
		if entry.Stale {
			stale = " stale"
		}
		fmt.Fprintf(s.out, "%s [%s%s] score %d\n", entry.EmpireName, entry.NodeID, stale, entry.Score)
	}
	return nil
}

func (s *session) hyperdriveMenu(state game.EconomyState) (bool, error) {
	if s.nodeID == "" || s.travelSubmitter == nil {
		return false, errTravelUnavailable
	}
	nodes, err := s.remoteNodeOptions()
	if err != nil {
		return false, err
	}
	if len(nodes) == 0 {
		return false, errors.New("no remote nodes available")
	}
	destNode, err := s.choose("Destination Node", nodes)
	if err != nil {
		return false, err
	}
	export, err := s.store.ExportTravelSnapshot(s.ctx, s.nodeID, state.Empire.ID)
	if err != nil {
		return false, err
	}
	resp, err := s.travelSubmitter.SubmitTravel(s.ctx, interdoor.TravelSubmitRequest{
		GlobalID: export.GlobalID,
		HomeNode: export.HomeNode,
		DestNode: destNode,
		Snapshot: export.Snapshot,
	})
	if err != nil {
		return false, err
	}
	if resp.TravelID == "" {
		return false, errors.New("hub returned empty travel id")
	}
	if err := s.store.MarkTravelSubmitted(s.ctx, state.Empire.ID, export.GlobalID, export.HomeNode, destNode, resp.TravelID, s.now()); err != nil {
		return false, err
	}
	fmt.Fprintf(s.out, "%s entered Hyperdrive toward %s. Travel %s pending.\n", state.Empire.EmpireName, destNode, resp.TravelID)
	return true, nil
}

func (s *session) remoteNodeOptions() ([]string, error) {
	entries, err := s.store.RemoteRoster(s.ctx, s.now())
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var nodes []string
	for _, entry := range entries {
		if entry.NodeID == "" || entry.NodeID == s.nodeID || seen[entry.NodeID] {
			continue
		}
		seen[entry.NodeID] = true
		nodes = append(nodes, entry.NodeID)
	}
	return nodes, nil
}

func (s *session) developMenu(state *game.EconomyState) error {
	notice := ""
	for {
		items := []menuItem{
			{Key: "1", Label: "Activate Region"},
			{Key: "2", Label: "Build Structure"},
			{Key: "3", Label: "Research Energy Tech"},
			{Key: "4", Label: "Buy Mine"},
			{Key: "5", Label: "Hire/Assign Workers"},
			{Key: "6", Label: "Sell Minerals"},
			{Key: "Q", Label: "Return"},
		}
		if s.presenter.ansi {
			s.presenter.developMenu(s.out, *state, items, notice)
		} else {
			if notice != "" {
				fmt.Fprintln(s.out, notice)
			}
			s.presenter.menu(s.out, "DEVELOP EMPIRE", items)
		}
		notice = ""
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "1":
			if err := s.activateRegion(state); err != nil {
				notice = "Activate Region blocked: " + err.Error()
			} else {
				notice = "Region activated."
			}
		case "2":
			if err := s.buildStructure(state); err != nil {
				notice = "Build Structure blocked: " + err.Error()
			} else {
				notice = "Structure built."
			}
		case "3":
			if err := s.researchEnergy(state); err != nil {
				notice = "Research Energy blocked: " + err.Error()
			} else {
				notice = "Research complete."
			}
		case "4":
			if err := s.buyMine(state); err != nil {
				notice = "Buy Mine blocked: " + err.Error()
			} else {
				notice = "Mine purchased."
			}
		case "5":
			if err := s.workersMenu(state); err != nil {
				notice = "Workers blocked: " + err.Error()
			}
		case "6":
			if err := s.sellMinerals(state); err != nil {
				notice = "Sell Minerals blocked: " + err.Error()
			} else {
				notice = "Minerals sold."
			}
		case "Q":
			return nil
		default:
			notice = "Unknown develop command."
		}
	}
}

func (s *session) activateRegion(state *game.EconomyState) error {
	regionType, err := s.choose("Region", []string{
		game.RegionAgricultural,
		game.RegionIndustrial,
		game.RegionDesert,
		game.RegionUrban,
		game.RegionRiver,
		game.RegionOcean,
		game.RegionVolcanic,
	})
	if err != nil {
		return err
	}
	if err := game.ActivateRegion(state, regionType); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Region activated.")
	return nil
}

func (s *session) buildStructure(state *game.EconomyState) error {
	structure, err := s.choose("Structure", []string{"Research Lab", "Construction Factory", "Miners Guild", "Fishing Guild"})
	if err != nil {
		return err
	}
	if err := game.BuildStructure(state, structure); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Structure built.")
	return nil
}

func (s *session) researchEnergy(state *game.EconomyState) error {
	tech, err := s.choose("Energy Tech", []string{game.TechEnergyFission, game.TechEnergyFusion})
	if err != nil {
		return err
	}
	if err := game.ResearchEnergyTech(state, tech); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Research complete.")
	return nil
}

func (s *session) buyMine(state *game.EconomyState) error {
	mineType, err := s.choose("Mine", game.MineTypes)
	if err != nil {
		return err
	}
	if err := game.BuyMine(state, mineType); err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintln(s.out, "Mine purchased.")
	return nil
}

func (s *session) workersMenu(state *game.EconomyState) error {
	items := []menuItem{
		{Key: "1", Label: "Hire Miner"},
		{Key: "2", Label: "Assign Miner"},
		{Key: "3", Label: "Hire Fisher"},
		{Key: "Q", Label: "Return"},
	}
	notice := ""
	for {
		s.presenter.workerMenu(s.out, *state, items, notice)
		notice = ""
		choice, err := s.prompt("Command")
		if err != nil {
			return err
		}
		switch strings.ToUpper(choice) {
		case "1":
			if err := game.HireMiner(state); err != nil {
				notice = workerActionMessage("Hire Miner", err, *state)
				continue
			}
			if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
				return err
			}
			notice = fmt.Sprintf("Miner hired. Available miners: %d.", state.Buildings.MinersAvailable)
		case "2":
			if state.Buildings.MinersAvailable <= 0 {
				notice = "Assign Miner blocked: no available miners. Hire a miner first."
				continue
			}
			mineType, err := s.choose("Mine", game.MineTypes)
			if err != nil {
				notice = err.Error()
				continue
			}
			if err := game.AssignMiner(state, mineType); err != nil {
				notice = workerActionMessage("Assign Miner", err, *state)
				continue
			}
			if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
				return err
			}
			notice = fmt.Sprintf("Miner assigned to %s.", mineType)
		case "3":
			if err := game.HireFisher(state); err != nil {
				notice = workerActionMessage("Hire Fisher", err, *state)
				continue
			}
			if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
				return err
			}
			notice = fmt.Sprintf("Fisher hired. Assigned fishers: %d.", state.Buildings.FishersAssigned)
		case "Q":
			return nil
		default:
			notice = "Unknown worker command."
		}
	}
}

func workerActionMessage(action string, err error, state game.EconomyState) string {
	switch {
	case errors.Is(err, game.ErrGuildRequired) && action == "Hire Miner":
		return "Hire Miner blocked: build Miners Guild first."
	case errors.Is(err, game.ErrGuildRequired) && action == "Hire Fisher":
		return "Hire Fisher blocked: build Fishing Guild first."
	case errors.Is(err, game.ErrGuildRequired):
		return action + " blocked: no available worker."
	case errors.Is(err, game.ErrInsufficientMoney):
		return fmt.Sprintf("%s blocked: need %d money; have %d.", action, game.HireWorkerMoneyCost, state.Empire.Money)
	case errors.Is(err, game.ErrInsufficientTurns):
		return fmt.Sprintf("%s blocked: need 1 turn; have %d.", action, state.Empire.TurnsLeft)
	default:
		return action + " blocked: " + err.Error()
	}
}

func (s *session) sellMinerals(state *game.EconomyState) error {
	mineType, err := s.choose("Mine", game.MineTypes)
	if err != nil {
		return err
	}
	value, err := game.SellMinerals(state, mineType)
	if err != nil {
		return err
	}
	if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
		return err
	}
	fmt.Fprintf(s.out, "Sold minerals for %d.\n", value)
	return nil
}

func (s *session) bankMenu(state *game.EconomyState) error {
	items := []menuItem{
		{Key: "1", Label: "Deposit"},
		{Key: "2", Label: "Withdraw"},
		{Key: "Q", Label: "Return"},
	}
	notice := ""
	for {
		action := ""
		if s.presenter.ansi {
			s.presenter.bankMenu(s.out, *state, items, notice)
			raw, err := s.prompt("Command")
			if err != nil {
				return err
			}
			notice = ""
			switch strings.ToUpper(raw) {
			case "1":
				action = "Deposit"
			case "2":
				action = "Withdraw"
			case "3", "Q":
				return nil
			default:
				notice = "Invalid bank action."
				continue
			}
		} else {
			if notice != "" {
				fmt.Fprintln(s.out, notice)
			}
			s.presenter.bankMenu(s.out, *state, items, "")
			raw, err := s.prompt("Bank Action")
			if err != nil {
				return err
			}
			notice = ""
			switch strings.ToUpper(raw) {
			case "1":
				action = "Deposit"
			case "2":
				action = "Withdraw"
			case "3", "Q":
				return nil
			default:
				notice = "Invalid bank action."
				continue
			}
		}
		amount, err := s.promptInt("Amount")
		if err != nil {
			notice = "Bank blocked: " + err.Error()
			continue
		}
		switch action {
		case "Deposit":
			err = game.Bank(&state.Empire, amount)
		case "Withdraw":
			err = game.Withdraw(&state.Empire, amount)
		}
		if err != nil {
			notice = action + " blocked: " + err.Error()
			continue
		}
		if err := s.store.SaveEconomy(s.ctx, *state); err != nil {
			return err
		}
		notice = fmt.Sprintf("Bank updated: %s %d.", strings.ToLower(action), amount)
	}
}

func (s *session) choose(label string, options []string) (string, error) {
	if s.presenter.ansi {
		s.presenter.choiceMenu(s.out, label, options)
		raw, err := s.prompt("Command")
		if err != nil {
			return "", err
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > len(options) {
			return "", fmt.Errorf("invalid %s", strings.ToLower(label))
		}
		return options[n-1], nil
	}
	for i, option := range options {
		fmt.Fprintf(s.out, "[%d] %s\n", i+1, option)
	}
	raw, err := s.prompt(label)
	if err != nil {
		return "", err
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > len(options) {
		return "", fmt.Errorf("invalid %s", strings.ToLower(label))
	}
	return options[n-1], nil
}

func (s *session) promptInt(label string) (int, error) {
	raw, err := s.prompt(label)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	return n, nil
}

func (s *session) prompt(label string) (string, error) {
	s.presenter.prompt(s.out, label)
	line, err := s.in.ReadString('\n')
	if errors.Is(err, io.EOF) && line != "" {
		return strings.TrimSpace(line), nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
