package session

import (
	"bytes"
	"strings"
	"testing"

	"empireascendant/internal/game"
)

func TestPrintMenuIncludesTitleAndCommands(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(false)
	p.menu(&output, "TEST MENU", []menuItem{
		{Key: "1", Label: "First"},
		{Key: "Q", Label: "Return"},
	})

	out := output.String()
	for _, want := range []string{"TEST MENU", "[1] First", "[Q] Return"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("plain menu contains ANSI escape:\n%q", out)
	}
}

func TestANSIPrintMenuIncludesEscapes(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.menu(&output, "TEST MENU", []menuItem{
		{Key: "1", Label: "First"},
	})

	out := output.String()
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("ANSI menu missing escape sequence:\n%q", out)
	}
	if !strings.Contains(out, "1") || !strings.Contains(out, "First") {
		t.Fatalf("ANSI menu missing command text:\n%s", out)
	}
	for _, want := range []string{"┌", "│", "└", "┐", "┘"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ANSI menu missing frame glyph %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "╔") || strings.Contains(out, "═") {
		t.Fatalf("ANSI menu still uses rejected double-line frame:\n%s", out)
	}
}

func TestANSITitleUsesDoorGameScreen(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.title(&output)
	p.menu(&output, "MAIN MENU", []menuItem{
		{Key: "E", Label: "Enter Your Empire"},
		{Key: "R", Label: "Rankings -- Top Empires"},
		{Key: "N", Label: "Galactic News"},
		{Key: "S", Label: "Story / Instructions"},
		{Key: "Q", Label: "Quit"},
	})

	out := output.String()
	plain := stripANSI(out)
	if !strings.Contains(out, "\x1b[2J") {
		t.Fatalf("ANSI title missing clear sequence:\n%s", out)
	}
	for _, want := range []string{"InterDoor Empire Node", "Dominion-inspired Strategy Game", "Empire Menu", "Enter Your Empire", "Rankings Top Empires", "Story / Instructions", "Galactic News", "Your Command <E,R,N,S,Q>", "┌", "┐", "│"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("ANSI asset menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, out)
		}
	}
}

func TestANSIPromptUsesCommandBox(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.prompt(&output, "Username")

	out := output.String()
	for _, want := range []string{"\x1b[", "Username", "┌─", "│", ">"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ANSI prompt missing %q:\n%s", want, out)
		}
	}
}

func TestANSIAssetPromptSuppressionOnlyAppliesOnce(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.menu(&output, "MAIN MENU", []menuItem{{Key: "Q", Label: "Quit"}})
	before := output.Len()
	p.prompt(&output, "Command")
	mainPrompt := output.String()[before:]
	if !strings.Contains(mainPrompt, ">") || strings.Contains(mainPrompt, "Your Command") {
		t.Fatalf("main asset live prompt wrong:\n%s", mainPrompt)
	}

	p.menu(&output, "EMPIRE HQ [15 turns remaining]", []menuItem{{Key: "Q", Label: "Quit to Main"}})
	before = output.Len()
	p.prompt(&output, "Command")
	hqPrompt := output.String()[before:]
	if !strings.Contains(hqPrompt, ">") || strings.Contains(hqPrompt, "Your Command") {
		t.Fatalf("HQ asset live prompt wrong:\n%s", hqPrompt)
	}

	before = output.Len()
	p.prompt(&output, "Username")
	if output.Len() == before {
		t.Fatalf("non-command prompt was suppressed after asset:\n%s", output.String())
	}
}

func TestANSIUsesEmpireStatusStyleForHQ(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.menu(&output, "EMPIRE HQ [15 turns remaining]", []menuItem{
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

	plain := stripANSI(output.String())
	for _, want := range []string{"Head Quarters", "Your Empire Report", "Develop Empire", "Bank", "Galactic Dispatch", "Hyperdrive", "Your Command <T,D,B,A,I,M,W,H,Q>"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("HQ asset menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
}

func TestANSIDevelopMenuUsesMenu8AndLiveStatus(t *testing.T) {
	state := game.NewStartingEconomy("empire-1")
	state.Empire = game.Empire{ID: "empire-1", Money: 9000, TurnsLeft: 14, ResearchPts: 2, BuildingPts: 3}
	ag := state.Regions[game.RegionAgricultural]
	ag.Activated = 1
	state.Regions[game.RegionAgricultural] = ag

	var output bytes.Buffer
	p := newPresenter(true)
	p.developMenu(&output, state, []menuItem{
		{Key: "1", Label: "Activate Region"},
		{Key: "2", Label: "Build Structure"},
		{Key: "3", Label: "Research Energy Tech"},
		{Key: "4", Label: "Buy Mine"},
		{Key: "5", Label: "Hire/Assign Workers"},
		{Key: "6", Label: "Sell Minerals"},
		{Key: "Q", Label: "Return"},
	}, "")

	plain := stripANSI(output.String())
	for _, want := range []string{
		"Current Status",
		"Agr 1/10",
		"Develop Menu",
		"Activate Region",
		"Build Structure",
		"Sell Minerals",
		"Your Command <1,2,3,4,5,6,Q>",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("develop menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
	if strings.Contains(plain, "Regions\nBuildings\nMines") {
		t.Fatalf("develop menu still shows placeholder status:\n%s", plain)
	}
	for _, want := range []string{
		ansiCyan + ansiBright + "Regions",
		ansiGreen + ansiBright + "Buildings",
		ansiYellow + ansiBright + "Workers",
		ansiMagenta + ansiBright + "Mines",
		ansiRed + ansiBright + "Empire",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("develop status missing section color %q:\nraw:\n%s", want, output.String())
		}
	}
	for _, want := range []string{
		"Labs " + ansiBlackBG + ansiGreen + ansiBright + "2",
		"Money " + ansiBlackBG + ansiRed + ansiBright + "9000",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("develop status missing value color %q:\nraw:\n%s", want, output.String())
		}
	}
}

func TestANSIWorkerMenuShowsReadiness(t *testing.T) {
	state := game.NewStartingEconomy("empire-1")
	state.Empire = game.Empire{ID: "empire-1", Money: 9000, TurnsLeft: 14}

	var output bytes.Buffer
	p := newPresenter(true)
	p.workerMenu(&output, state, []menuItem{
		{Key: "1", Label: "Hire Miner"},
		{Key: "2", Label: "Assign Miner"},
		{Key: "3", Label: "Hire Fisher"},
		{Key: "Q", Label: "Return"},
	}, "Hire Miner blocked: build Miners Guild first.")

	plain := stripANSI(output.String())
	for _, want := range []string{
		"Worker Readiness",
		"Hire Miner blocked: build Miners Guild first.",
		"Hire Miner",
		"No  Build Miners Guild first (15 building)",
		"Assign Miner",
		"No  No available miners",
		"Hire Fisher",
		"No  Build Fishing Guild first (15 building)",
		"Worker Action",
		"Your Command <1,2,3,Q>",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("worker menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
}

func TestANSIBankMenuShowsFinancialStatus(t *testing.T) {
	state := game.NewStartingEconomy("empire-1")
	state.Empire = game.Empire{ID: "empire-1", Money: 9900, MoneyBank: 100, TurnsLeft: 15}

	var output bytes.Buffer
	p := newPresenter(true)
	p.bankMenu(&output, state, []menuItem{
		{Key: "1", Label: "Deposit"},
		{Key: "2", Label: "Withdraw"},
		{Key: "Q", Label: "Return"},
	}, "Bank updated: deposit 100.")

	plain := stripANSI(output.String())
	for _, want := range []string{
		"Action Result",
		"Bank updated: deposit 100.",
		"Financial Status",
		"Cash 9900",
		"Bank 100",
		"Turns 15",
		"Bank",
		"Deposit",
		"Withdraw",
		"Your Command <1,2,Q>",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("bank menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
}

func TestANSIChoiceMenuUsesMenu8Prompt(t *testing.T) {
	var output bytes.Buffer
	p := newPresenter(true)
	p.choiceMenu(&output, "Region", []string{
		game.RegionAgricultural,
		game.RegionIndustrial,
		game.RegionDesert,
	})

	plain := stripANSI(output.String())
	for _, want := range []string{"Region", "Agricultural", "Industrial", "Desert", "Your Command <1,2,3>"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("choice menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
	if strings.Contains(plain, "[1] Agricultural") {
		t.Fatalf("choice menu still uses plain bracket list:\n%s", plain)
	}
}

func TestANSIAttackMenuUsesMenu8AndArmyStatus(t *testing.T) {
	state := game.NewStartingEconomy("empire-1")
	state.Empire = game.Empire{ID: "empire-1", Money: 8000, TurnsLeft: 12}
	state.Military.NormalSoldiers = 125
	state.Military.NuclearMissiles = 1

	var output bytes.Buffer
	p := newPresenter(true)
	p.attackMenu(&output, state, []menuItem{
		{Key: "1", Label: "Recruit Units"},
		{Key: "2", Label: "Build Defenses"},
		{Key: "3", Label: "Research Military Tech"},
		{Key: "4", Label: "Buy Missile"},
		{Key: "5", Label: "Ground Assault"},
		{Key: "6", Label: "Ballistic Strike"},
		{Key: "7", Label: "Spy Mission"},
		{Key: "Q", Label: "Return"},
	}, "")

	plain := stripANSI(output.String())
	for _, want := range []string{
		"Army Status",
		"Soldiers 125",
		"Nuclear 1",
		"Attack Menu",
		"Recruit Units",
		"Build Defenses",
		"Spy Mission",
		"Your Command <1,2,3,4,5,6,7,Q>",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("attack menu missing %q:\nplain:\n%s\nraw:\n%s", want, plain, output.String())
		}
	}
}
