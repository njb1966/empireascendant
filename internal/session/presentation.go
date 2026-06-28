package session

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"empireascendant/internal/game"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBright  = "\x1b[1m"
	ansiBlackBG = "\x1b[40m"
	ansiClear   = "\x1b[2J\x1b[H"
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[35m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiWhite   = "\x1b[37m"
	ansiDim     = "\x1b[2m"
)

type menuItem struct {
	Key   string
	Label string
}

type presenter struct {
	ansi                      bool
	suppressNextCommandPrompt bool
	appendNextANSIMenu        bool
}

func newPresenter(ansi bool) presenter {
	return presenter{ansi: ansi}
}

func (p *presenter) title(out io.Writer) {
	p.suppressNextCommandPrompt = false
	p.appendNextANSIMenu = false
	if p.ansi {
		p.ansiTitle(out)
		return
	}
	fmt.Fprintln(out, "")
	p.write(out, ansiCyan+ansiBright, "+------------------------------------------------------------------+")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+"  _____ __  __ ____ ___ ____  _____                              "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+" | ____|  \\/  |  _ \\_ _|  _ \\| ____|                             "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+" |  _| | |\\/| | |_) | || |_) |  _|                               "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+" | |___| |  | |  __/| ||  _ <| |___                              "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+" |_____|_|  |_|_|  |___|_| \\_\\_____|                             "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiMagenta+"                     A S C E N D A N T                          "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "|"+ansiWhite+"                       EMPIRE ASCENDANT                           "+ansiCyan+"|")
	p.write(out, ansiCyan+ansiBright, "+------------------------------------------------------------------+")
}

func (p *presenter) section(out io.Writer, title string) {
	p.suppressNextCommandPrompt = false
	p.appendNextANSIMenu = false
	if p.ansi {
		fmt.Fprint(out, ansiClear, ansiBlackBG)
		for _, line := range ansiBoxLines(title, nil, 76, ansiWhite+ansiBright, ansiCyan+ansiBright) {
			fmt.Fprintln(out, line)
		}
		p.appendNextANSIMenu = true
		return
	}
	fmt.Fprintln(out, "")
	p.write(out, ansiRed+ansiBright, "+-- "+title+" "+strings.Repeat("-", max(1, 61-len(title)))+"+")
}

func (p *presenter) menu(out io.Writer, title string, items []menuItem) {
	if p.ansi {
		p.ansiMenu(out, title, items)
		return
	}
	p.section(out, title)
	if len(items) == 0 {
		return
	}
	leftCount := (len(items) + 1) / 2
	for i := 0; i < leftCount; i++ {
		left := p.menuCell(items[i])
		rowItems := []menuItem{items[i]}
		right := ""
		if i+leftCount < len(items) {
			right = p.menuCell(items[i+leftCount])
			rowItems = append(rowItems, items[i+leftCount])
		}
		var line string
		if right == "" {
			line = fmt.Sprintf("| %-62s |", left)
		} else {
			line = fmt.Sprintf("| %-30s  %-30s |", left, right)
		}
		fmt.Fprintln(out, p.colorMenuLine(line, rowItems))
	}
	p.write(out, ansiRed+ansiBright, "+------------------------------------------------------------------+")
}

func (p *presenter) prompt(out io.Writer, label string) {
	if p.ansi {
		p.ansiPrompt(out, label)
		return
	}
	fmt.Fprintf(out, "%s: ", label)
}

func (p *presenter) storyInstructions(out io.Writer) {
	p.suppressNextCommandPrompt = false
	p.appendNextANSIMenu = false
	if p.ansi {
		p.ansiStoryInstructions(out)
		return
	}
	p.title(out)
	fmt.Fprintln(out, "")
	p.write(out, ansiMagenta+ansiBright, "Story")
	fmt.Fprintln(out, "The old empires have gone quiet, leaving scattered worlds,")
	fmt.Fprintln(out, "half-built fleets, abandoned mines, and relay routes that")
	fmt.Fprintln(out, "still whisper between InterDoor nodes. Your house begins with")
	fmt.Fprintln(out, "one world, a small treasury, and enough industry to decide")
	fmt.Fprintln(out, "whether it will build, trade, raid, or travel.")
	fmt.Fprintln(out, "")
	p.write(out, ansiMagenta+ansiBright, "How To Play")
	fmt.Fprintln(out, "- Enter Your Empire to create or resume an empire.")
	fmt.Fprintln(out, "- Turns are the main daily limit. Development, attacks, and travel")
	fmt.Fprintln(out, "  spend turns; banking and reports do not.")
	fmt.Fprintln(out, "- Start by developing regions. Regions feed population, industry,")
	fmt.Fprintln(out, "  energy, research, and construction.")
	fmt.Fprintln(out, "- Build labs and factories to increase daily research and building")
	fmt.Fprintln(out, "  points. Guilds unlock miners and fishers.")
	fmt.Fprintln(out, "- Buy mines, hire miners, then assign miners to extract minerals.")
	fmt.Fprintln(out, "  Sell stored minerals when you need money.")
	fmt.Fprintln(out, "- Use the Attack Menu for recruitment, defenses, missiles, local")
	fmt.Fprintln(out, "  attacks, cross-node attacks, and spy missions.")
	fmt.Fprintln(out, "- Rankings show local empires and known remote empires.")
	fmt.Fprintln(out, "- Galactic News shows local and InterDoor federation events.")
	fmt.Fprintln(out, "- Wanderers lists remote empires seen through the InterDoor hub.")
	fmt.Fprintln(out, "- Hyperdrive can send an empire toward another known node.")
	fmt.Fprintln(out, "")
	p.write(out, ansiMagenta+ansiBright, "Current Limits")
	fmt.Fprintln(out, "- Visitor gameplay after travel is not implemented yet.")
	fmt.Fprintln(out, "- Public deployment and live hub listing are separate later work.")
	fmt.Fprintln(out, "- Balance tuning is a later D5 slice; current numbers are unchanged.")
}

func (p *presenter) empireReport(out io.Writer, state game.EconomyState) {
	p.suppressNextCommandPrompt = false
	if !p.ansi {
		fmt.Fprintln(out, game.EmpireReport(state))
		return
	}
	fmt.Fprint(out, ansiClear, ansiBlackBG)
	reportRows := []string{
		fmt.Sprintf("Empire : %s", state.Empire.EmpireName),
		fmt.Sprintf("World  : %s", state.Empire.WorldName),
		fmt.Sprintf("Turns  : %d", state.Empire.TurnsLeft),
		"",
		fmt.Sprintf("Money %-10d  Bank %-10d  Population %-10d", state.Empire.Money, state.Empire.MoneyBank, state.Empire.Population),
		fmt.Sprintf("Food  %-10d  Storage %-7d  Energy %-10d", state.Empire.Food, state.Empire.FoodStorage, state.Empire.Energy),
		fmt.Sprintf("Research %-7d Building %-7d", state.Empire.ResearchPts, state.Empire.BuildingPts),
	}
	for _, line := range ansiBoxLines("General Information", reportRows, 76, ansiWhite+ansiBright, ansiCyan+ansiBright) {
		fmt.Fprintln(out, line)
	}
	assetRows := []string{
		fmt.Sprintf("Regions  Agr %d/%d  Ind %d/%d  River %d/%d",
			state.Regions[game.RegionAgricultural].Activated,
			state.Regions[game.RegionAgricultural].Quantity,
			state.Regions[game.RegionIndustrial].Activated,
			state.Regions[game.RegionIndustrial].Quantity,
			state.Regions[game.RegionRiver].Activated,
			state.Regions[game.RegionRiver].Quantity,
		),
		fmt.Sprintf("Buildings Labs %d  Factories %d  Fossil %d  Fission %d  Fusion %d",
			state.Buildings.ResearchLabs,
			state.Buildings.ConstructionFactories,
			state.Buildings.FossilPlants,
			state.Buildings.FissionPlants,
			state.Buildings.FusionPlants,
		),
		fmt.Sprintf("Mines    Gold %d  Silver %d  Iron %d  Copper %d  Nickel %d",
			state.Mines[game.MineGold].NumMines,
			state.Mines[game.MineSilver].NumMines,
			state.Mines[game.MineIron].NumMines,
			state.Mines[game.MineCopper].NumMines,
			state.Mines[game.MineNickel].NumMines,
		),
	}
	for _, line := range ansiBoxLines("Region / Mine / Energy Information", assetRows, 76, ansiWhite+ansiBright, ansiGreen+ansiBright) {
		fmt.Fprintln(out, line)
	}
	militaryRows := []string{
		fmt.Sprintf("Military Soldiers %d  Tanks %d  Hovercraft %d",
			state.Military.NormalSoldiers+state.Military.SuperSoldiers+state.Military.MegaSoldiers,
			state.Military.Tanks,
			state.Military.Hovercraft,
		),
		fmt.Sprintf("Defence  Turrets %d  Satellites %d  Shields %d",
			state.Military.GroundTurrets,
			state.Military.OrbitalSatellites,
			state.Military.GlobalShields,
		),
		fmt.Sprintf("Weapons  Nuclear %d  Antimatter %d  Spies %d",
			state.Military.NuclearMissiles,
			state.Military.AntimatterMissiles,
			state.Military.Spies,
		),
	}
	for _, line := range ansiBoxLines("Army / Defence Information", militaryRows, 76, ansiWhite+ansiBright, ansiRed+ansiBright) {
		fmt.Fprintln(out, line)
	}
	p.appendNextANSIMenu = true
}

func (p *presenter) bankMenu(out io.Writer, state game.EconomyState, items []menuItem, notice string) {
	p.suppressNextCommandPrompt = false
	p.appendNextANSIMenu = false
	if !p.ansi {
		if notice != "" {
			fmt.Fprintln(out, notice)
		}
		p.menu(out, "BANK", items)
		return
	}
	fmt.Fprint(out, ansiClear, ansiBlackBG)
	ansiNoticeBox(out, notice)
	for _, line := range ansiBoxLines("Financial Status", bankStatusRows(state), 76, ansiWhite+ansiBright, ansiGreen+ansiBright) {
		fmt.Fprintln(out, line)
	}
	if screen, ok := ansiMenu8Asset("Bank", ansiMenu8ItemPatches(items), "<1,2,Q>"); ok {
		fmt.Fprint(out, screen)
		p.suppressNextCommandPrompt = true
		return
	}
	for _, line := range ansiBoxLines("Bank", ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiGreen+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func (p *presenter) developMenu(out io.Writer, state game.EconomyState, items []menuItem, notice string) {
	p.suppressNextCommandPrompt = false
	if !p.ansi {
		if notice != "" {
			fmt.Fprintln(out, notice)
		}
		p.menu(out, "DEVELOP EMPIRE", items)
		return
	}
	fmt.Fprint(out, p.ansiMenuPrefix())
	ansiNoticeBox(out, notice)
	for _, line := range ansiBoxLines("Current Status", developStatusRows(state), 76, ansiWhite+ansiBright, ansiGreen+ansiBright) {
		fmt.Fprintln(out, line)
	}
	if screen, ok := ansiMenu8Asset("Develop Menu", ansiMenu8ItemPatches(items), "<1,2,3,4,5,6,Q>"); ok {
		fmt.Fprint(out, screen)
		p.suppressNextCommandPrompt = true
		return
	}
	for _, line := range ansiBoxLines("Develop Menu", ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiGreen+ansiBright) {
		fmt.Fprintln(out, line)
	}
	ansiTraceBlock(out, "DEVELOP EMPIRE")
}

func (p *presenter) workerMenu(out io.Writer, state game.EconomyState, items []menuItem, notice string) {
	p.suppressNextCommandPrompt = false
	rows := workerReadinessRows(state, notice)
	if !p.ansi {
		p.section(out, "WORKERS")
		if notice != "" {
			fmt.Fprintln(out, notice)
		}
		for _, row := range rows {
			fmt.Fprintln(out, stripANSI(row))
		}
		p.menu(out, "WORKER ACTION", items)
		return
	}
	fmt.Fprint(out, p.ansiMenuPrefix())
	ansiNoticeBox(out, notice)
	for _, line := range ansiBoxLines("Worker Readiness", rows, 76, ansiWhite+ansiBright, ansiYellow+ansiBright) {
		fmt.Fprintln(out, line)
	}
	if screen, ok := ansiMenu8Asset("Worker Action", ansiMenu8ItemPatches(items), "<1,2,3,Q>"); ok {
		fmt.Fprint(out, screen)
		p.suppressNextCommandPrompt = true
		return
	}
	for _, line := range ansiBoxLines("Worker Action", ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiYellow+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func (p *presenter) attackMenu(out io.Writer, state game.EconomyState, items []menuItem, notice string) {
	p.suppressNextCommandPrompt = false
	if !p.ansi {
		if notice != "" {
			fmt.Fprintln(out, notice)
		}
		p.menu(out, "ATTACK MENU", items)
		return
	}
	fmt.Fprint(out, p.ansiMenuPrefix())
	ansiNoticeBox(out, notice)
	for _, line := range ansiBoxLines("Army Status", armyStatusRows(state), 76, ansiWhite+ansiBright, ansiRed+ansiBright) {
		fmt.Fprintln(out, line)
	}
	if screen, ok := ansiMenu8Asset("Attack Menu", ansiMenu8ItemPatches(items), "<1,2,3,4,5,6,7,Q>"); ok {
		fmt.Fprint(out, screen)
		p.suppressNextCommandPrompt = true
		return
	}
	for _, line := range ansiBoxLines("Attack Menu", ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiRed+ansiBright) {
		fmt.Fprintln(out, line)
	}
	ansiTraceBlock(out, "ATTACK MENU")
}

func (p *presenter) reportMenu(out io.Writer, items []menuItem) {
	p.suppressNextCommandPrompt = false
	if !p.ansi {
		p.menu(out, "EMPIRE REPORT", items)
		return
	}
	if screen, ok := ansiMenu8Asset("Empire Status", ansiMenu8ItemPatches(items), "<1-8,Q>"); ok {
		fmt.Fprint(out, p.ansiMenuPrefix(), screen)
		p.suppressNextCommandPrompt = true
		return
	}
	fmt.Fprint(out, p.ansiMenuPrefix())
	for _, line := range ansiBoxLines("Empire Status", ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiCyan+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func (p *presenter) reportDetail(out io.Writer, title string, rows []string, titleStyle string) {
	p.suppressNextCommandPrompt = false
	if p.ansi {
		fmt.Fprint(out, ansiClear, ansiBlackBG)
		for _, line := range ansiBoxLines(title, rows, 76, ansiWhite+ansiBright, titleStyle) {
			fmt.Fprintln(out, line)
		}
		p.appendNextANSIMenu = true
		return
	}
	for _, row := range rows {
		fmt.Fprintln(out, row)
	}
}

func (p *presenter) choiceMenu(out io.Writer, label string, options []string) {
	p.suppressNextCommandPrompt = false
	if !p.ansi {
		return
	}
	items := make([]menuItem, 0, len(options))
	for i, option := range options {
		items = append(items, menuItem{Key: strconv.Itoa(i + 1), Label: option})
	}
	if len(items) <= 8 {
		if screen, ok := ansiMenu8Asset(label, ansiMenu8ItemPatches(items), ansiPromptOptions(items)); ok {
			fmt.Fprint(out, ansiClear, ansiBlackBG, screen)
			p.suppressNextCommandPrompt = true
			return
		}
	}
	fmt.Fprint(out, ansiClear, ansiBlackBG)
	for _, line := range ansiBoxLines(label, ansiMenuRows(items), 76, ansiWhite+ansiBright, ansiCyan+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func (p *presenter) ansiTitle(out io.Writer) {
	fmt.Fprint(out, ansiClear, ansiBlackBG)
	if header, ok := ansiHeaderAsset(); ok {
		fmt.Fprint(out, header)
		if !strings.HasSuffix(header, "\n") {
			fmt.Fprintln(out)
		}
	}
	p.appendNextANSIMenu = true
}

func (p *presenter) ansiMenu(out io.Writer, title string, items []menuItem) {
	p.suppressNextCommandPrompt = false
	if title == "MAIN MENU" {
		if screen, ok := ansiMenu8Asset("Empire Menu", []ansiMenuPatch{
			{oldKey: "1", oldLabel: "General Information", newKey: "E", newLabel: "Enter Your Empire"},
			{oldKey: "2", oldLabel: "Financial Information", newKey: "R", newLabel: "Rankings Top Empires"},
			{oldKey: "3", oldLabel: "Army Status Information", newKey: "N", newLabel: "Galactic News"},
			{oldKey: "4", oldLabel: "Ballistic Weapons Information", newKey: "S", newLabel: "Story / Instructions"},
			{oldKey: "5", oldLabel: "Global Defence Information", newKey: "Q", newLabel: "Quit"},
			{oldKey: "6", oldLabel: "Region Information", newKey: " ", newLabel: ""},
			{oldKey: "7", oldLabel: "Mine Information", newKey: " ", newLabel: ""},
			{oldKey: "8", oldLabel: "Empire Energy Information", newKey: " ", newLabel: ""},
		}, "<E,R,N,S,Q>"); ok {
			fmt.Fprint(out, p.ansiMenuPrefix(), screen)
			p.suppressNextCommandPrompt = true
			return
		}
	}
	if strings.HasPrefix(title, "EMPIRE HQ") {
		if screen, ok := ansiMenu8Asset("Head Quarters", []ansiMenuPatch{
			{oldKey: "1", oldLabel: "General Information", newKey: "T", newLabel: "Your Empire Report"},
			{oldKey: "2", oldLabel: "Financial Information", newKey: "D", newLabel: "Develop Empire"},
			{oldKey: "3", oldLabel: "Army Status Information", newKey: "B", newLabel: "Bank"},
			{oldKey: "4", oldLabel: "Ballistic Weapons Information", newKey: "A", newLabel: "Attack Menu"},
			{oldKey: "5", oldLabel: "Global Defence Information", newKey: "I", newLabel: "Intel Report"},
			{oldKey: "6", oldLabel: "Region Information", newKey: "M", newLabel: "Galactic Dispatch"},
			{oldKey: "7", oldLabel: "Mine Information", newKey: "W", newLabel: "Wanderers"},
			{oldKey: "8", oldLabel: "Empire Energy Information", newKey: "H", newLabel: "Hyperdrive / Q=Quit"},
		}, "<T,D,B,A,I,M,W,H,Q>"); ok {
			fmt.Fprint(out, p.ansiMenuPrefix(), screen)
			p.suppressNextCommandPrompt = true
			return
		}
	}
	p.appendNextANSIMenu = false
	menuTitle := ansiMenuTitle(title)
	menuColor := ansiFrameColor(title)
	rows := ansiMenuRows(items)
	left := ansiBoxLines(menuTitle, rows, 58, ansiWhite+ansiBright, menuColor)
	right := ansiStatusBox(title)
	for i := 0; i < max(len(left), len(right)); i++ {
		leftLine := strings.Repeat(" ", 58)
		if i < len(left) {
			leftLine = left[i]
		}
		if len(right) == 0 {
			fmt.Fprintln(out, leftLine)
			continue
		}
		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}
		fmt.Fprintln(out, leftLine+"  "+rightLine)
	}
	ansiTraceBlock(out, title)
}

func (p *presenter) ansiMenuPrefix() string {
	if p.appendNextANSIMenu {
		p.appendNextANSIMenu = false
		return ""
	}
	return ansiClear + ansiBlackBG
}

func (p *presenter) ansiStoryInstructions(out io.Writer) {
	p.ansiTitle(out)
	story := []string{
		"Story",
		"The old empires have gone quiet. Scattered worlds,",
		"abandoned mines, and relay routes remain.",
		"Your house begins with one world, a small treasury,",
		"and a choice: build, trade, raid, or travel.",
		"",
		"Play",
		"Turns limit daily action. Development, attacks, and",
		"Hyperdrive spend turns; banking and reports are free.",
		"Develop regions for food, energy, research, and build points.",
		"Build labs/factories first; guilds unlock miners and fishers.",
		"Buy mines, hire miners, assign them, then sell stored minerals.",
		"Recruit and defend before raiding local or remote empires.",
		"Use reports often. Rankings, News, and Wanderers track the sector.",
		"",
		"Limits",
		"Visitor gameplay, public deployment, live listing, and",
		"balance tuning remain later slices.",
	}
	for _, line := range ansiBoxLines("Story / Instructions", story, 76, ansiWhite+ansiBright, ansiMagenta+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func (p *presenter) ansiPrompt(out io.Writer, label string) {
	if strings.EqualFold(label, "Command") && p.suppressNextCommandPrompt {
		p.suppressNextCommandPrompt = false
		fmt.Fprintf(out, "%s%s>%s ", ansiBlackBG, ansiYellow+ansiBright, ansiReset)
		return
	}
	p.suppressNextCommandPrompt = false
	promptLabel := label
	width := 38
	width = max(width, visibleLen("┌─ "+promptLabel+" ")+2)
	top := ansiWhite + ansiBright + "┌─ " + ansiCyan + ansiBright + promptLabel + " " + ansiWhite + ansiBright
	top += strings.Repeat("─", max(1, width-visibleLen("┌─ "+promptLabel+" ")-1)) + "┐" + ansiReset
	fmt.Fprintln(out)
	fmt.Fprintln(out, ansiBlackBG+top)
	fmt.Fprintf(out, "%s%s│ %s>%s ", ansiBlackBG, ansiWhite+ansiBright, ansiYellow+ansiBright, ansiReset)
}

func mainMenuPatches() []ansiPatch {
	return []ansiPatch{
		{old: "Empire Status", new: "Empire Control"},
		{old: "General Information", new: "Enter Your Empire"},
		{old: "Financial Information", new: "Rankings -- Top Empires"},
		{old: "Army Status Information", new: ""},
		{old: "Ballistic Weapons Information", new: "Story / Help"},
		{old: "Global Defence Information", new: ""},
		{old: "Region Information", new: ""},
		{old: "Mine Information", new: "Galactic News"},
		{old: "Empire Energy Information", new: "Quit"},
	}
}

func mainMenuKeyMap() []ansiKeyPatch {
	return []ansiKeyPatch{
		{old: "1", new: "E"},
		{old: "2", new: "R"},
		{old: "3", new: " "},
		{old: "4", new: "S"},
		{old: "5", new: " "},
		{old: "6", new: " "},
		{old: "7", new: " "},
		{old: "8", new: "Q"},
		{old: "9", new: "N"},
	}
}

func hqMenuPatches() []ansiPatch {
	return []ansiPatch{
		{old: "Empire Status", new: "Global Head Quarters"},
		{old: "General Information", new: "Your Empire Report"},
		{old: "Financial Information", new: "Develop Empire"},
		{old: "Army Status Information", new: "Bank"},
		{old: "Ballistic Weapons Information", new: "Attack Menu"},
		{old: "Global Defence Information", new: "Intel Report"},
		{old: "Region Information", new: "Galactic Dispatches"},
		{old: "Mine Information", new: "Wanderers"},
		{old: "Empire Energy Information", new: "Hyperdrive / Q=Quit"},
	}
}

func hqMenuKeyMap() []ansiKeyPatch {
	return []ansiKeyPatch{
		{old: "1", new: "T"},
		{old: "2", new: "D"},
		{old: "3", new: "B"},
		{old: "4", new: "A"},
		{old: "5", new: "I"},
		{old: "6", new: "M"},
		{old: "7", new: "W"},
		{old: "8", new: "H"},
	}
}

func ansiAsset(name string, patches []ansiPatch, keyPatches []ansiKeyPatch, promptOptions string) (string, bool) {
	data, err := readANSAsset("MENU0.ANS")
	if name != "MENU0.ANS" {
		data, err = readANSAsset(name)
	}
	if err != nil {
		return "", false
	}
	for _, patch := range patches {
		data = replaceANSLabel(data, patch.old, patch.new)
	}
	for _, patch := range keyPatches {
		dot := "."
		if patch.new == " " {
			dot = " "
		}
		data = bytes.ReplaceAll(data, []byte("\x1b[35m"+patch.old+"\x1b[36m."), []byte("\x1b[35m"+patch.new+"\x1b[36m"+dot))
		data = bytes.ReplaceAll(data, []byte("\x1b[1;35m"+patch.old+"\x1b[36m."), []byte("\x1b[1;35m"+patch.new+"\x1b[36m"+dot))
	}
	if promptOptions != "" {
		data = replacePromptOptions(data, promptOptions)
	}
	return cp437ANSIToUTF8(data), true
}

func ansiBankAsset() (string, bool) {
	data, err := readANSAsset("MENU5.ANS")
	if err != nil {
		return "", false
	}
	data = bytes.Replace(data, []byte("\x1b[37m3\x1b[0m."), []byte("\x1b[37mQ\x1b[0m."), 1)
	data = bytes.Replace(data, []byte("\x1b[1;36mI\x1b[34mnvestments"), []byte("\x1b[1;36mR\x1b[34meturn     "), 1)
	data = replacePromptOptions(data, "<1,2,Q,?>")
	return cp437ANSIToUTF8(data), true
}

func ansiHeaderAsset() (string, bool) {
	data, err := readANSAsset("HEADER1.ANS")
	if err != nil {
		return "", false
	}
	screen := cp437ANSIToUTF8(data)
	screen = strings.ReplaceAll(screen, "\r\n", "\n")
	screen = strings.ReplaceAll(screen, "\r", "\n")
	lines := strings.Split(screen, "\n")
	for len(lines) > 0 && strings.TrimSpace(stripANSI(lines[0])) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(stripANSI(lines[len(lines)-1])) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", false
	}
	return strings.Join(lines, "\n") + ansiReset + "\n", true
}

type ansiMenuPatch struct {
	oldKey   string
	oldLabel string
	newKey   string
	newLabel string
}

func ansiMenu8ItemPatches(items []menuItem) []ansiMenuPatch {
	oldLabels := []string{
		"General Information",
		"Financial Information",
		"Army Status Information",
		"Ballistic Weapons Information",
		"Global Defence Information",
		"Region Information",
		"Mine Information",
		"Empire Energy Information",
	}
	patches := make([]ansiMenuPatch, 0, len(oldLabels))
	for i, oldLabel := range oldLabels {
		newKey := " "
		newLabel := ""
		if i < len(items) {
			newKey = items[i].Key
			newLabel = items[i].Label
		}
		patches = append(patches, ansiMenuPatch{
			oldKey:   strconv.Itoa(i + 1),
			oldLabel: oldLabel,
			newKey:   newKey,
			newLabel: newLabel,
		})
	}
	return patches
}

func ansiPromptOptions(items []menuItem) string {
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}
	return "<" + strings.Join(keys, ",") + ">"
}

func developStatusRows(state game.EconomyState) []string {
	ag := state.Regions[game.RegionAgricultural]
	ind := state.Regions[game.RegionIndustrial]
	des := state.Regions[game.RegionDesert]
	urban := state.Regions[game.RegionUrban]
	river := state.Regions[game.RegionRiver]
	ocean := state.Regions[game.RegionOcean]
	volc := state.Regions[game.RegionVolcanic]
	gold := state.Mines[game.MineGold]
	silver := state.Mines[game.MineSilver]
	iron := state.Mines[game.MineIron]
	copper := state.Mines[game.MineCopper]
	nickel := state.Mines[game.MineNickel]
	regionStyle := ansiCyan + ansiBright
	buildingStyle := ansiGreen + ansiBright
	workerStyle := ansiYellow + ansiBright
	mineStyle := ansiMagenta + ansiBright
	empireStyle := ansiRed + ansiBright
	return []string{
		fmt.Sprintf("%s   Agr %s  Ind %s  Des %s  Urb %s",
			statusLabel("Regions", regionStyle),
			statusValue(regionStyle, "%d/%d", ag.Activated, ag.Quantity),
			statusValue(regionStyle, "%d/%d", ind.Activated, ind.Quantity),
			statusValue(regionStyle, "%d/%d", des.Activated, des.Quantity),
			statusValue(regionStyle, "%d/%d", urban.Activated, urban.Quantity),
		),
		fmt.Sprintf("          River %s  Ocean %s  Volc %s",
			statusValue(regionStyle, "%d/%d", river.Activated, river.Quantity),
			statusValue(regionStyle, "%d/%d", ocean.Activated, ocean.Quantity),
			statusValue(regionStyle, "%d/%d", volc.Activated, volc.Quantity),
		),
		fmt.Sprintf("%s Labs %s  Factories %s  Fossil %s  Fission %s  Fusion %s",
			statusLabel("Buildings", buildingStyle),
			statusValue(buildingStyle, "%d", state.Buildings.ResearchLabs),
			statusValue(buildingStyle, "%d", state.Buildings.ConstructionFactories),
			statusValue(buildingStyle, "%d", state.Buildings.FossilPlants),
			statusValue(buildingStyle, "%d", state.Buildings.FissionPlants),
			statusValue(buildingStyle, "%d", state.Buildings.FusionPlants),
		),
		fmt.Sprintf("%s   Miners %s available  Fishers %s assigned",
			statusLabel("Workers", workerStyle),
			statusValue(workerStyle, "%d", state.Buildings.MinersAvailable),
			statusValue(workerStyle, "%d", state.Buildings.FishersAssigned),
		),
		fmt.Sprintf("%s     Gold %s  Silver %s  Iron %s  Copper %s  Nickel %s",
			statusLabel("Mines", mineStyle),
			statusValue(mineStyle, "%d", gold.NumMines),
			statusValue(mineStyle, "%d", silver.NumMines),
			statusValue(mineStyle, "%d", iron.NumMines),
			statusValue(mineStyle, "%d", copper.NumMines),
			statusValue(mineStyle, "%d", nickel.NumMines),
		),
		fmt.Sprintf("%s    Money %s  Turns %s  Research %s  Building %s",
			statusLabel("Empire", empireStyle),
			statusValue(empireStyle, "%d", state.Empire.Money),
			statusValue(empireStyle, "%d", state.Empire.TurnsLeft),
			statusValue(empireStyle, "%d", state.Empire.ResearchPts),
			statusValue(empireStyle, "%d", state.Empire.BuildingPts),
		),
	}
}

func bankStatusRows(state game.EconomyState) []string {
	style := ansiGreen + ansiBright
	return []string{
		fmt.Sprintf("%s Cash %s  Bank %s  Turns %s",
			statusLabel("Funds", style),
			statusValue(style, "%d", state.Empire.Money),
			statusValue(style, "%d", state.Empire.MoneyBank),
			statusValue(style, "%d", state.Empire.TurnsLeft),
		),
		"Deposits and withdrawals are free actions.",
		"Enter an amount after choosing Deposit or Withdraw.",
	}
}

func statusLabel(label, style string) string {
	return ansiBlackBG + style + label + ansiBlackBG + ansiWhite + ansiBright
}

func statusValue(style, format string, args ...any) string {
	return ansiBlackBG + style + fmt.Sprintf(format, args...) + ansiBlackBG + ansiWhite + ansiBright
}

func workerReadinessRows(state game.EconomyState, notice string) []string {
	rows := make([]string, 0, 7)
	rows = append(rows,
		fmt.Sprintf("%s Money %s  Turns %s  Building %s  Worker Cost %s",
			statusLabel("Empire", ansiRed+ansiBright),
			statusValue(ansiRed+ansiBright, "%d", state.Empire.Money),
			statusValue(ansiRed+ansiBright, "%d", state.Empire.TurnsLeft),
			statusValue(ansiRed+ansiBright, "%d", state.Empire.BuildingPts),
			statusValue(ansiRed+ansiBright, "%d", game.HireWorkerMoneyCost),
		),
		fmt.Sprintf("%s %s",
			statusLabel("Hire Miner", ansiCyan+ansiBright),
			workerCapability(state, "miner"),
		),
		fmt.Sprintf("%s %s",
			statusLabel("Assign Miner", ansiMagenta+ansiBright),
			workerCapability(state, "assign"),
		),
		fmt.Sprintf("%s %s",
			statusLabel("Hire Fisher", ansiGreen+ansiBright),
			workerCapability(state, "fisher"),
		),
		fmt.Sprintf("%s Gold %s/%s  Silver %s/%s  Iron %s/%s",
			statusLabel("Assignments", ansiWhite+ansiBright),
			assignedValue(state, game.MineGold),
			mineCountValue(state, game.MineGold),
			assignedValue(state, game.MineSilver),
			mineCountValue(state, game.MineSilver),
			assignedValue(state, game.MineIron),
			mineCountValue(state, game.MineIron),
		),
		fmt.Sprintf("            Copper %s/%s  Nickel %s/%s  Fishers %s",
			assignedValue(state, game.MineCopper),
			mineCountValue(state, game.MineCopper),
			assignedValue(state, game.MineNickel),
			mineCountValue(state, game.MineNickel),
			statusValue(ansiGreen+ansiBright, "%d", state.Buildings.FishersAssigned),
		),
	)
	return rows
}

func workerCapability(state game.EconomyState, action string) string {
	switch action {
	case "miner":
		if !state.Buildings.MinersGuild {
			return statusValue(ansiRed+ansiBright, "No") + fmt.Sprintf("  Build Miners Guild first (%d building)", game.BuildMinersGuildCost)
		}
		if state.Empire.TurnsLeft <= 0 {
			return statusValue(ansiRed+ansiBright, "No") + "  Need 1 turn"
		}
		if state.Empire.Money < game.HireWorkerMoneyCost {
			return statusValue(ansiRed+ansiBright, "No") + fmt.Sprintf("  Need %d money", game.HireWorkerMoneyCost)
		}
		return statusValue(ansiGreen+ansiBright, "Yes") + fmt.Sprintf(" Costs %d money and 1 turn", game.HireWorkerMoneyCost)
	case "assign":
		if state.Buildings.MinersAvailable <= 0 {
			return statusValue(ansiRed+ansiBright, "No") + "  No available miners"
		}
		return statusValue(ansiGreen+ansiBright, "Yes") + fmt.Sprintf(" %d miner(s) available", state.Buildings.MinersAvailable)
	case "fisher":
		if !state.Buildings.FishingGuild {
			return statusValue(ansiRed+ansiBright, "No") + fmt.Sprintf("  Build Fishing Guild first (%d building)", game.BuildFishingGuildCost)
		}
		if state.Empire.TurnsLeft <= 0 {
			return statusValue(ansiRed+ansiBright, "No") + "  Need 1 turn"
		}
		if state.Empire.Money < game.HireWorkerMoneyCost {
			return statusValue(ansiRed+ansiBright, "No") + fmt.Sprintf("  Need %d money", game.HireWorkerMoneyCost)
		}
		return statusValue(ansiGreen+ansiBright, "Yes") + fmt.Sprintf(" Costs %d money and 1 turn", game.HireWorkerMoneyCost)
	default:
		return ""
	}
}

func assignedValue(state game.EconomyState, mineType string) string {
	mine := state.Mines[mineType]
	return statusValue(ansiMagenta+ansiBright, "%d", mine.MinersAssigned)
}

func mineCountValue(state game.EconomyState, mineType string) string {
	mine := state.Mines[mineType]
	return statusValue(ansiMagenta+ansiBright, "%d", mine.NumMines)
}

func generalReportRows(state game.EconomyState) []string {
	return []string{
		fmt.Sprintf("Empire : %s", state.Empire.EmpireName),
		fmt.Sprintf("World  : %s", state.Empire.WorldName),
		fmt.Sprintf("Turns  : %d", state.Empire.TurnsLeft),
		fmt.Sprintf("People : %d", state.Empire.Population),
		fmt.Sprintf("Food   : %d stored, %d in reserve", state.Empire.Food, state.Empire.FoodStorage),
		fmt.Sprintf("Energy : %d", state.Empire.Energy),
	}
}

func financialReportRows(state game.EconomyState) []string {
	return []string{
		fmt.Sprintf("Money           : %d", state.Empire.Money),
		fmt.Sprintf("Bank            : %d", state.Empire.MoneyBank),
		fmt.Sprintf("Research Points : %d", state.Empire.ResearchPts),
		fmt.Sprintf("Building Points : %d", state.Empire.BuildingPts),
		"Reports are free. Development, attacks, and travel spend turns.",
	}
}

func armyReportRows(state game.EconomyState) []string {
	m := state.Military
	return []string{
		fmt.Sprintf("Normal Soldiers : %d", m.NormalSoldiers),
		fmt.Sprintf("Super Soldiers  : %d", m.SuperSoldiers),
		fmt.Sprintf("Mega Soldiers   : %d", m.MegaSoldiers),
		fmt.Sprintf("Tanks           : %d", m.Tanks),
		fmt.Sprintf("Hovercraft      : %d", m.Hovercraft),
		fmt.Sprintf("Attack Power    : %d", game.AttackPower(m)),
	}
}

func ballisticReportRows(state game.EconomyState) []string {
	m := state.Military
	return []string{
		fmt.Sprintf("Nuclear Missiles     : %d", m.NuclearMissiles),
		fmt.Sprintf("Antimatter Missiles  : %d", m.AntimatterMissiles),
		fmt.Sprintf("Nuclear Tech         : %t", state.Tech[game.TechBallisticNuclear]),
		fmt.Sprintf("Antimatter Tech      : %t", state.Tech[game.TechBallisticAnti]),
	}
}

func defenceReportRows(state game.EconomyState) []string {
	m := state.Military
	return []string{
		fmt.Sprintf("Ground Turrets      : %d", m.GroundTurrets),
		fmt.Sprintf("Orbital Satellites  : %d", m.OrbitalSatellites),
		fmt.Sprintf("Global Shields      : %d", m.GlobalShields),
		fmt.Sprintf("Recon Drones        : %d", m.ReconDrones),
		fmt.Sprintf("Spies               : %d", m.Spies),
		fmt.Sprintf("Defence Power       : %d", game.DefensePower(m)),
	}
}

func regionReportRows(state game.EconomyState) []string {
	rows := make([]string, 0, len(game.RegionTypes))
	for _, regionType := range game.RegionTypes {
		if regionType == game.RegionWasteland {
			continue
		}
		region := state.Regions[regionType]
		rows = append(rows, fmt.Sprintf("%-14s %d/%d  next cost %d",
			region.Type,
			region.Activated,
			region.Quantity,
			region.ActivateCost,
		))
	}
	return rows
}

func mineReportRows(state game.EconomyState) []string {
	rows := make([]string, 0, len(game.MineTypes))
	for _, mineType := range game.MineTypes {
		mine := state.Mines[mineType]
		rows = append(rows, fmt.Sprintf("%-8s mines %d  assigned %d  stored %d",
			mine.Type,
			mine.NumMines,
			mine.MinersAssigned,
			mine.StoredMinerals,
		))
	}
	return rows
}

func energyReportRows(state game.EconomyState) []string {
	return []string{
		fmt.Sprintf("Current Energy : %d", state.Empire.Energy),
		fmt.Sprintf("Fossil Plants  : %d  tech %t", state.Buildings.FossilPlants, state.Tech[game.TechEnergyFossil]),
		fmt.Sprintf("Fission Plants : %d  tech %t", state.Buildings.FissionPlants, state.Tech[game.TechEnergyFission]),
		fmt.Sprintf("Fusion Plants  : %d  tech %t", state.Buildings.FusionPlants, state.Tech[game.TechEnergyFusion]),
	}
}

func armyStatusRows(state game.EconomyState) []string {
	m := state.Military
	unitStyle := ansiCyan + ansiBright
	vehicleStyle := ansiGreen + ansiBright
	defenceStyle := ansiYellow + ansiBright
	weaponStyle := ansiMagenta + ansiBright
	intelStyle := ansiBlue + ansiBright
	powerStyle := ansiRed + ansiBright
	return []string{
		fmt.Sprintf("%s     Soldiers %s  Super %s  Mega %s",
			statusLabel("Units", unitStyle),
			statusValue(unitStyle, "%d", m.NormalSoldiers),
			statusValue(unitStyle, "%d", m.SuperSoldiers),
			statusValue(unitStyle, "%d", m.MegaSoldiers),
		),
		fmt.Sprintf("%s  Tanks %s  Hovercraft %s",
			statusLabel("Vehicles", vehicleStyle),
			statusValue(vehicleStyle, "%d", m.Tanks),
			statusValue(vehicleStyle, "%d", m.Hovercraft),
		),
		fmt.Sprintf("%s   Turrets %s  Satellites %s  Shields %s",
			statusLabel("Defence", defenceStyle),
			statusValue(defenceStyle, "%d", m.GroundTurrets),
			statusValue(defenceStyle, "%d", m.OrbitalSatellites),
			statusValue(defenceStyle, "%d", m.GlobalShields),
		),
		fmt.Sprintf("%s   Nuclear %s  Antimatter %s",
			statusLabel("Weapons", weaponStyle),
			statusValue(weaponStyle, "%d", m.NuclearMissiles),
			statusValue(weaponStyle, "%d", m.AntimatterMissiles),
		),
		fmt.Sprintf("%s     Recon %s  Spies %s  Terrorists %s",
			statusLabel("Intel", intelStyle),
			statusValue(intelStyle, "%d", m.ReconDrones),
			statusValue(intelStyle, "%d", m.Spies),
			statusValue(intelStyle, "%d", m.Terrorists),
		),
		fmt.Sprintf("%s     Attack %s  Defence %s  Turns %s  Money %s",
			statusLabel("Power", powerStyle),
			statusValue(powerStyle, "%d", game.AttackPower(m)),
			statusValue(powerStyle, "%d", game.DefensePower(m)),
			statusValue(powerStyle, "%d", state.Empire.TurnsLeft),
			statusValue(powerStyle, "%d", state.Empire.Money),
		),
	}
}

func ansiNoticeBox(out io.Writer, notice string) {
	if notice == "" {
		return
	}
	for _, line := range ansiBoxLines("Action Result", []string{notice}, 76, ansiWhite+ansiBright, ansiRed+ansiBright) {
		fmt.Fprintln(out, line)
	}
}

func ansiMenu8Asset(title string, patches []ansiMenuPatch, promptOptions string) (string, bool) {
	data, err := readANSAsset("MENU8.ANS")
	if err != nil {
		return "", false
	}
	grid := parseANSGrid(data)
	grid.patchText("Empire Status", title)
	for _, patch := range patches {
		row, col, ok := grid.findText(patch.oldLabel)
		if !ok {
			continue
		}
		grid.patchTextAt(row, col, len(patch.oldLabel), patch.newLabel)
		keyCol := grid.findKeyBefore(row, col, patch.oldKey)
		if keyCol >= 0 {
			grid.patchTextAt(row, keyCol, len(patch.oldKey), patch.newKey)
			if patch.newKey == " " && keyCol+1 < len(grid.rows[row]) {
				grid.patchTextAt(row, keyCol+1, 1, " ")
			}
		}
	}
	grid.patchPromptOptions(promptOptions)
	return grid.render(), true
}

type ansiGrid struct {
	rows [][]ansiCell
}

type ansiCell struct {
	ch    string
	style string
}

type ansiStyle struct {
	bold bool
	dim  bool
	fg   int
	bg   int
}

func parseANSGrid(data []byte) ansiGrid {
	grid := ansiGrid{}
	style := ansiStyle{fg: -1, bg: 40}
	row, col := 0, 0
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if ch == '\x1b' && i+1 < len(data) && data[i+1] == '[' {
			j := i + 2
			for j < len(data) && (data[j] < '@' || data[j] > '~') {
				j++
			}
			if j >= len(data) {
				break
			}
			params := string(data[i+2 : j])
			switch data[j] {
			case 'm':
				style.applySGR(params)
			case 'C':
				col += ansiParam(params, 1)
			case 'D':
				col -= ansiParam(params, 1)
				if col < 0 {
					col = 0
				}
			}
			i = j
			continue
		}
		switch ch {
		case '\r':
			col = 0
		case '\n':
			row++
		default:
			text := string(ch)
			if ch >= 0x80 {
				text = cp437Rune(ch)
			}
			grid.set(row, col, text, style.sequence())
			col++
		}
	}
	return grid
}

func (s *ansiStyle) applySGR(params string) {
	if params == "" {
		params = "0"
	}
	for _, part := range strings.Split(params, ";") {
		code, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		switch {
		case code == 0:
			*s = ansiStyle{fg: -1, bg: 40}
		case code == 1:
			s.bold = true
			s.dim = false
		case code == 2:
			s.dim = true
			s.bold = false
		case code >= 30 && code <= 37:
			s.fg = code
		case code >= 40 && code <= 47:
			s.bg = code
		}
	}
}

func (s ansiStyle) sequence() string {
	var codes []string
	if s.bg >= 40 {
		codes = append(codes, strconv.Itoa(s.bg))
	}
	if s.bold {
		codes = append(codes, "1")
	}
	if s.dim {
		codes = append(codes, "2")
	}
	if s.fg >= 30 {
		codes = append(codes, strconv.Itoa(s.fg))
	}
	if len(codes) == 0 {
		return ansiReset
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func ansiParam(params string, fallback int) int {
	if params == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.Split(params, ";")[0])
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func (g *ansiGrid) set(row, col int, ch, style string) {
	for len(g.rows) <= row {
		g.rows = append(g.rows, nil)
	}
	for len(g.rows[row]) <= col {
		g.rows[row] = append(g.rows[row], ansiCell{ch: " ", style: ansiBlackBG})
	}
	g.rows[row][col] = ansiCell{ch: ch, style: style}
}

func (g ansiGrid) findText(text string) (int, int, bool) {
	for row, cells := range g.rows {
		plain := rowPlain(cells)
		byteCol := strings.Index(plain, text)
		if byteCol >= 0 {
			return row, utf8.RuneCountInString(plain[:byteCol]), true
		}
	}
	return 0, 0, false
}

func rowPlain(cells []ansiCell) string {
	var b strings.Builder
	for _, cell := range cells {
		if cell.ch == "" {
			b.WriteByte(' ')
			continue
		}
		b.WriteString(cell.ch)
	}
	return b.String()
}

func (g *ansiGrid) patchText(old, new string) {
	row, col, ok := g.findText(old)
	if !ok {
		return
	}
	g.patchTextAt(row, col, len(old), new)
}

func (g *ansiGrid) patchTextAt(row, col, width int, text string) {
	if row >= len(g.rows) {
		return
	}
	runes := []rune(text)
	if len(runes) > width {
		runes = runes[:width]
	}
	style := ansiWhite + ansiBright
	if col < len(g.rows[row]) && g.rows[row][col].style != "" {
		style = g.rows[row][col].style
	}
	for i := 0; i < width; i++ {
		ch := " "
		if i < len(runes) {
			ch = string(runes[i])
		}
		g.set(row, col+i, ch, style)
	}
}

func (g ansiGrid) findKeyBefore(row, labelCol int, oldKey string) int {
	if row >= len(g.rows) {
		return -1
	}
	cells := g.rows[row]
	for col := labelCol - 1; col >= 0 && col >= labelCol-8; col-- {
		if col < len(cells) && cells[col].ch == oldKey {
			return col
		}
	}
	return -1
}

func (g *ansiGrid) patchPromptOptions(options string) {
	row, col, ok := g.findText("Your Command ")
	if !ok {
		return
	}
	start := col + len("Your Command ")
	width := 0
	if row < len(g.rows) {
		width = len(g.rows[row]) - start
	}
	if width <= 0 {
		return
	}
	g.patchTextAt(row, start, width, options)
}

func (g ansiGrid) render() string {
	var b strings.Builder
	for _, cells := range g.rows {
		last := lastVisibleCell(cells)
		if last < 0 {
			b.WriteByte('\n')
			continue
		}
		currentStyle := ""
		for col := 0; col <= last; col++ {
			cell := ansiCell{ch: " ", style: ansiBlackBG}
			if col < len(cells) && cells[col].ch != "" {
				cell = cells[col]
			}
			if cell.style != currentStyle {
				b.WriteString(cell.style)
				currentStyle = cell.style
			}
			b.WriteString(cell.ch)
		}
		b.WriteString(ansiReset)
		b.WriteByte('\n')
	}
	return b.String()
}

func lastVisibleCell(cells []ansiCell) int {
	for i := len(cells) - 1; i >= 0; i-- {
		if cells[i].ch != "" && cells[i].ch != " " {
			return i
		}
	}
	return -1
}

func readANSAsset(name string) ([]byte, error) {
	paths := []string{
		"SCREENS/" + name,
		name,
		"../../SCREENS/" + name,
		"../../" + name,
	}
	var last error
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		last = err
	}
	return nil, last
}

type ansiPatch struct {
	old string
	new string
}

type ansiKeyPatch struct {
	old string
	new string
}

func replaceANSLabel(data []byte, old, new string) []byte {
	if len(new) > len(old) {
		new = new[:len(old)]
	}
	return bytes.ReplaceAll(data, []byte(old), []byte(new+strings.Repeat(" ", len(old)-len(new))))
}

func replacePromptOptions(data []byte, promptOptions string) []byte {
	label := []byte("Your Command ")
	start := bytes.Index(data, label)
	if start < 0 {
		return data
	}
	optionsStart := start + len(label)
	lineEnd := bytes.IndexByte(data[optionsStart:], '\r')
	if lineEnd < 0 {
		return data
	}
	lineEnd += optionsStart
	options := coloredPromptOptions(promptOptions)
	if len(options) < lineEnd-optionsStart {
		options = append(options, bytes.Repeat([]byte(" "), lineEnd-optionsStart-len(options))...)
	}
	replaced := make([]byte, 0, len(data))
	replaced = append(replaced, data[:optionsStart]...)
	replaced = append(replaced, options...)
	replaced = append(replaced, data[lineEnd:]...)
	return replaced
}

func coloredPromptOptions(promptOptions string) []byte {
	var out []byte
	for _, r := range promptOptions {
		switch {
		case r == '?':
			out = append(out, []byte("\x1b[33m?\x1b[30m")...)
		case r >= 'A' && r <= 'Z':
			out = append(out, []byte("\x1b[37m"+string(r)+"\x1b[30m")...)
		default:
			out = append(out, byte(r))
		}
	}
	return out
}

func cp437ANSIToUTF8(data []byte) string {
	var b strings.Builder
	for i := 0; i < len(data); i++ {
		if data[i] == '\x1b' {
			b.WriteByte(data[i])
			continue
		}
		if data[i] < 0x80 {
			b.WriteByte(data[i])
			continue
		}
		b.WriteString(cp437Rune(data[i]))
	}
	return b.String()
}

func cp437Rune(ch byte) string {
	switch ch {
	case 0xb0:
		return "░"
	case 0xb1:
		return "▒"
	case 0xb2:
		return "▓"
	case 0xb3:
		return "│"
	case 0xb4:
		return "┤"
	case 0xb5:
		return "╡"
	case 0xb6:
		return "╢"
	case 0xb7:
		return "╖"
	case 0xb8:
		return "╕"
	case 0xb9:
		return "╣"
	case 0xba:
		return "║"
	case 0xbb:
		return "╗"
	case 0xbc:
		return "╝"
	case 0xbd:
		return "╜"
	case 0xbe:
		return "╛"
	case 0xbf:
		return "┐"
	case 0xc0:
		return "└"
	case 0xc1:
		return "┴"
	case 0xc2:
		return "┬"
	case 0xc3:
		return "├"
	case 0xc4:
		return "─"
	case 0xc5:
		return "┼"
	case 0xc6:
		return "╞"
	case 0xc7:
		return "╟"
	case 0xc8:
		return "╚"
	case 0xc9:
		return "╔"
	case 0xca:
		return "╩"
	case 0xcb:
		return "╦"
	case 0xcc:
		return "╠"
	case 0xcd:
		return "═"
	case 0xce:
		return "╬"
	case 0xcf:
		return "╧"
	case 0xd0:
		return "╨"
	case 0xd1:
		return "╤"
	case 0xd2:
		return "╥"
	case 0xd3:
		return "╙"
	case 0xd4:
		return "╘"
	case 0xd5:
		return "╒"
	case 0xd6:
		return "╓"
	case 0xd7:
		return "╫"
	case 0xd8:
		return "╪"
	case 0xd9:
		return "┘"
	case 0xda:
		return "┌"
	case 0xdb:
		return "█"
	case 0xdc:
		return "▄"
	case 0xdd:
		return "▌"
	case 0xde:
		return "▐"
	case 0xdf:
		return "▀"
	default:
		return "?"
	}
}

func ansiMenuTitle(title string) string {
	switch {
	case title == "MAIN MENU":
		return "Empire Control"
	case strings.HasPrefix(title, "EMPIRE HQ"):
		return "Global Head Quarters"
	case title == "DEVELOP EMPIRE":
		return "Region Management Centre"
	case title == "ATTACK MENU":
		return "War Room"
	default:
		return title
	}
}

func ansiFrameColor(title string) string {
	switch {
	case title == "MAIN MENU":
		return ansiRed + ansiBright
	case strings.HasPrefix(title, "EMPIRE HQ"):
		return ansiBlue + ansiBright
	case title == "DEVELOP EMPIRE":
		return ansiGreen + ansiBright
	case title == "ATTACK MENU":
		return ansiRed + ansiBright
	default:
		return ansiMagenta + ansiBright
	}
}

func ansiMenuRows(items []menuItem) []string {
	if len(items) == 0 {
		return nil
	}
	leftCount := len(items)
	twoColumn := len(items) > 4
	if twoColumn {
		leftCount = (len(items) + 1) / 2
	}
	rows := []string{""}
	for i := 0; i < leftCount; i++ {
		if !twoColumn {
			rows = append(rows, "  "+ansiMenuCell(items[i], 48))
			continue
		}
		left := ansiMenuCell(items[i], 26)
		right := ""
		if i+leftCount < len(items) {
			right = ansiMenuCell(items[i+leftCount], 26)
		}
		rows = append(rows, "  "+left+"  "+right)
	}
	rows = append(rows, "")
	return rows
}

func ansiMenuCell(item menuItem, width int) string {
	plain := fmt.Sprintf("%s. %s", item.Key, item.Label)
	if visibleLen(plain) > width {
		runes := []rune(plain)
		plain = string(runes[:width])
	}
	text := fmt.Sprintf("%s%s%s.%s %s", ansiMagenta+ansiBright, item.Key, ansiCyan+ansiBright, ansiReset, ansiWhite+ansiBright)
	text += strings.TrimPrefix(plain, item.Key+". ")
	text += ansiReset
	return padANSI(text, width)
}

func ansiStatusBox(title string) []string {
	var rows []string
	boxTitle := "Status"
	switch {
	case title == "MAIN MENU":
		boxTitle = "Node Status"
		rows = []string{"  Local Realm", "  ANSI Link", "  InterDoor Ready"}
	case strings.HasPrefix(title, "EMPIRE HQ"):
		boxTitle = "Empire Status"
		rows = []string{"  " + strings.TrimPrefix(strings.TrimSuffix(title, "]"), "EMPIRE HQ ["), "  Reports Free", "  Actions Cost Turns"}
	case title == "DEVELOP EMPIRE":
		boxTitle = "Current Status"
		rows = []string{"  Regions", "  Buildings", "  Mines"}
	case title == "ATTACK MENU":
		boxTitle = "Army Status"
		rows = []string{"  Ground", "  Ballistic", "  Espionage"}
	default:
		return nil
	}
	return ansiBoxLines(boxTitle, rows, 20, ansiWhite+ansiBright, ansiGreen+ansiBright)
}

func ansiTraceBlock(out io.Writer, title string) {
	colorA, colorB, colorC := ansiCyan+ansiBright, ansiGreen+ansiBright, ansiRed+ansiBright
	if title == "ATTACK MENU" {
		colorB = ansiRed + ansiBright
		colorC = ansiMagenta + ansiBright
	}
	if title == "DEVELOP EMPIRE" {
		colorA = ansiGreen + ansiBright
		colorB = ansiCyan + ansiBright
		colorC = ansiMagenta + ansiBright
	}
	fmt.Fprintln(out, ansiBlackBG+colorA+"└──────┐"+strings.Repeat(" ", 7)+colorB+"┌────────────┐"+strings.Repeat(" ", 26)+colorC+"┌─────┘"+ansiReset)
	fmt.Fprintln(out, ansiBlackBG+strings.Repeat(" ", 7)+colorA+"└────┐"+strings.Repeat(" ", 5)+colorB+"└────┐"+strings.Repeat(" ", 7)+colorC+"┌────┘"+ansiReset)
}

func ansiBoxLines(title string, rows []string, width int, frameStyle, titleStyle string) []string {
	innerWidth := width - 2
	titleText := "─ " + title + " "
	top := frameStyle + "┌" + titleStyle + titleText + frameStyle + strings.Repeat("─", max(1, innerWidth-visibleLen(titleText))) + "┐" + ansiReset
	lines := []string{ansiBlackBG + top}
	for _, row := range rows {
		lines = append(lines, ansiBlackBG+frameStyle+"│"+ansiReset+padANSI(row, innerWidth)+ansiBlackBG+frameStyle+"│"+ansiReset)
	}
	lines = append(lines, ansiBlackBG+frameStyle+"└"+strings.Repeat("─", innerWidth)+"┘"+ansiReset)
	return lines
}

func (p presenter) menuCell(item menuItem) string {
	return fmt.Sprintf("[%s] %s", item.Key, item.Label)
}

func (p presenter) colorMenuLine(line string, items []menuItem) string {
	if !p.ansi {
		return line
	}
	colored := line
	for _, item := range items {
		key := "[" + item.Key + "]"
		colored = strings.ReplaceAll(colored, key, ansiMagenta+ansiBright+key+ansiReset)
	}
	return colored
}

func (p *presenter) write(out io.Writer, style, text string) {
	if p.ansi {
		fmt.Fprintln(out, style+text+ansiReset)
		return
	}
	fmt.Fprintln(out, stripANSI(text))
}

func stripANSI(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) {
				if text[i] >= '@' && text[i] <= '~' {
					break
				}
				i++
			}
			continue
		}
		b.WriteByte(text[i])
	}
	return b.String()
}

func padANSI(text string, width int) string {
	visible := visibleLen(text)
	if visible >= width {
		return text
	}
	return text + strings.Repeat(" ", width-visible)
}

func visibleLen(text string) int {
	return utf8.RuneCountInString(stripANSI(text))
}
