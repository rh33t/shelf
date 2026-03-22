package model

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shelf/internal/fs"
)

// States

type appState int

const (
	stateModeSelect appState = iota
	statePlatform
	stateCategory
	stateItem
	stateInput
	stateRename
	stateConfirm
	stateDone
	stateSearch
)

type confirmKind int

const (
	confirmSlugify confirmKind = iota
	confirmDelete
	confirmGenCategories
)

// Custom item delegate

type item struct{ name string }

func (i item) FilterValue() string { return i.name }
func (i item) Title() string       { return i.name }
func (i item) Description() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	if m.IsFiltered() && index != m.Index() {
		fmt.Fprint(w, dimItemStyle.Render("     "+i.name))
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("  ▶  "+i.name))
	} else {
		fmt.Fprint(w, normalItemStyle.Render("     "+i.name))
	}
}

// History

type histEntry struct {
	state      appState
	currentDir string
	label      string
}

// Model

type Model struct {
	mode    string
	baseDir string

	state      appState
	history    []histEntry
	currentDir string
	label      string

	list  list.Model
	input textinput.Model

	cKind    confirmKind
	cMsg     string
	cPending string

	inputLabel string
	renameOld  string
	searchBase string

	width  int
	height int

	SelectedPath string
	Err          error

	statusMsg string

	cfg *fs.Config
}

func New(mode string, cfg *fs.Config) Model {
	initStyles(cfg.PrimaryColor, cfg.SecondaryColor)
	m := Model{
		mode:   mode,
		width:  80,
		height: 24,
		cfg:    cfg,
	}
	if mode == "" {
		m.state = stateModeSelect
		m.label = "mode"
		m = m.initList([]string{"ctf", "box"}, "Select Mode")
	} else {
		baseDir := baseDirForMode(mode)
		if err := fs.MkdirAll(baseDir); err != nil {
			m.Err = err
			m.state = stateDone
			return m
		}
		m.baseDir = baseDir
		m.currentDir = baseDir
		m.state = statePlatform
		m.label = "platform"
		m = m.loadList()
	}
	return m
}

func baseDirForMode(mode string) string {
	if mode == "ctf" {
		return filepath.Join(labDir, "training", "challenges")
	}
	return filepath.Join(labDir, "training", "boxes")
}

// List helpers

func (m Model) initList(names []string, title string) Model {
	items := make([]list.Item, len(names))
	for i, n := range names {
		items[i] = item{n}
	}

	listHeight := max(m.height-4, 4)

	l := list.New(items, itemDelegate{}, m.width, listHeight)
	l.Title = title
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color(primary)).PaddingLeft(4).PaddingTop(2)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(5)
	m.list = l
	return m
}

func (m Model) loadList() Model {
	if m.state == stateModeSelect {
		return m.initList([]string{"ctf", "box"}, m.listTitle())
	}

	dirs, err := fs.ListDirs(m.currentDir)
	if err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error reading directory: %v", err))
		dirs = nil
	}

	var names []string
	if m.mode == "box" && m.state == statePlatform {
		seen := map[string]bool{}
		for _, p := range boxPlatforms {
			seen[p] = true
			names = append(names, p)
		}
		for _, d := range dirs {
			if !seen[d] {
				names = append(names, d)
			}
		}
	} else {
		names = dirs
	}

	return m.initList(names, m.listTitle())
}

func (m Model) listTitle() string {
	switch m.state {
	case stateModeSelect:
		return "Training Manager"
	case statePlatform:
		if m.mode == "ctf" {
			return "CTF — Select Platform"
		}
		return "Box — Select Platform"
	case stateCategory:
		return "CTF — Select Category"
	case stateItem:
		if m.mode == "ctf" {
			return "CTF — Select Challenge"
		}
		return "Box — Select Box"
	case stateSearch:
		return "Jump To"
	}
	return ""
}

// tea.Model

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := max(m.height-4, 4)
		m.list.SetSize(msg.Width, listHeight)
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m.delegateUpdate(msg)
}

func (m Model) delegateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case stateModeSelect, statePlatform, stateCategory, stateItem, stateSearch:
		m.list, cmd = m.list.Update(msg)
	case stateInput, stateRename:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isListState := m.state == stateModeSelect || m.state == statePlatform ||
		m.state == stateCategory || m.state == stateItem || m.state == stateSearch

	if isListState && m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch m.state {
	case stateModeSelect, statePlatform, stateCategory, stateItem, stateSearch:
		return m.handleListKey(msg)
	case stateInput, stateRename:
		return m.handleInputKey(msg)
	case stateConfirm:
		return m.handleConfirmKey(msg)
	}
	return m, nil
}

// Key handlers

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		if m.list.FilterState() == list.FilterApplied {
			m.list.ResetFilter()
			return m, nil
		}
		return m.goBack()

	case "ctrl+f":
		if m.state == stateSearch || m.state == stateModeSelect {
			return m, nil
		}
		return m.startSearch()

	case "enter", " ":
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		m.statusMsg = ""
		if m.state == stateSearch {
			path := filepath.Join(m.searchBase, sel.(item).name)
			if err := fs.MkdirAll(path); err != nil {
				m.Err = err
				m.state = stateDone
				return m, tea.Quit
			}
			return m.finish(path)
		}
		return m.selectItem(sel.(item).name)

	case "up", "k":
		if m.list.Index() == 0 {
			if visible := m.list.VisibleItems(); len(visible) > 0 {
				m.list.Select(len(visible) - 1)
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case "down", "j":
		visible := m.list.VisibleItems()
		if len(visible) > 0 && m.list.Index() == len(visible)-1 {
			m.list.Select(0)
			return m, nil
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case "n":
		if m.state == stateSearch || m.state == stateModeSelect {
			return m, nil
		}
		return m.startCreate()

	case "d":
		if m.state == stateSearch || m.state == stateModeSelect {
			return m, nil
		}
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		return m.startDelete(sel.(item).name)

	case "r":
		if m.state == stateSearch || m.state == stateModeSelect {
			return m, nil
		}
		sel := m.list.SelectedItem()
		if sel == nil {
			return m, nil
		}
		return m.startRename(sel.(item).name)

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.cancelInput()
	case "enter":
		return m.confirmInput()
	default:
		m.statusMsg = ""
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "y", "Y", "enter":
		return m.confirmAction()
	case "n", "N", "esc":
		return m.cancelConfirm()
	}
	return m, nil
}

// Actions

func (m Model) selectItem(name string) (tea.Model, tea.Cmd) {
	dir := filepath.Join(m.currentDir, name)

	switch m.state {
	case stateModeSelect:
		m.mode = name
		m.baseDir = baseDirForMode(name)
		if err := fs.MkdirAll(m.baseDir); err != nil {
			m.Err = err
			m.state = stateDone
			return m, tea.Quit
		}
		m.pushHistory(stateModeSelect)
		m.currentDir = m.baseDir
		m.state = statePlatform
		m.label = "platform"
		m = m.loadList()
		return m, nil

	case statePlatform:
		if err := fs.MkdirAll(dir); err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", err))
			return m, nil
		}
		if m.mode == "ctf" {
			existing, _ := fs.ListDirs(dir)
			if len(existing) == 0 {
				m.pushHistory(statePlatform)
				m.cKind = confirmGenCategories
				m.cMsg = fmt.Sprintf("Platform '%s' is empty.\nGenerate default CTF categories?", name)
				m.cPending = dir
				m.currentDir = dir
				m.label = "category"
				m.state = stateConfirm
				return m, nil
			}
			m.pushHistory(statePlatform)
			m.currentDir = dir
			m.state = stateCategory
			m.label = "category"
		} else {
			m.pushHistory(statePlatform)
			m.currentDir = dir
			m.state = stateItem
			m.label = "box"
		}
		m = m.loadList()
		return m, nil

	case stateCategory:
		if err := fs.MkdirAll(dir); err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", err))
			return m, nil
		}
		m.pushHistory(stateCategory)
		m.currentDir = dir
		m.state = stateItem
		m.label = "challenge"
		m = m.loadList()
		return m, nil

	case stateItem:
		if err := fs.MkdirAll(dir); err != nil {
			m.Err = err
			m.state = stateDone
			return m, tea.Quit
		}
		if m.mode == "box" {
			notesPath := filepath.Join(dir, "notes.md")
			if _, statErr := os.Stat(notesPath); os.IsNotExist(statErr) {
				platform := filepath.Base(m.currentDir)
				_ = fs.WriteNotesTemplate(dir, platform, name)
			}
		}
		return m.finish(dir)
	}
	return m, nil
}

func (m Model) finish(path string) (tea.Model, tea.Cmd) {
	m.SelectedPath = path
	m.state = stateDone
	return m, tea.Quit
}

func (m Model) goBack() (tea.Model, tea.Cmd) {
	if len(m.history) == 0 {
		return m, tea.Quit
	}
	entry := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.state = entry.state
	m.currentDir = entry.currentDir
	m.label = entry.label
	m.statusMsg = ""
	if entry.state == stateModeSelect {
		m.mode = ""
	}
	m = m.loadList()
	return m, nil
}

func (m *Model) pushHistory(s appState) {
	m.history = append(m.history, histEntry{
		state:      s,
		currentDir: m.currentDir,
		label:      m.label,
	})
}

func (m Model) startSearch() (tea.Model, tea.Cmd) {
	dirs, err := fs.WalkAllDirs(m.currentDir)
	if err != nil || len(dirs) == 0 {
		m.statusMsg = warnStyle.Render("No directories found.")
		return m, nil
	}
	m.pushHistory(m.state)
	m.searchBase = m.currentDir
	m.state = stateSearch
	m = m.initList(dirs, "Jump To")
	return m, func() tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	}
}

func (m Model) startCreate() (tea.Model, tea.Cmd) {
	m.inputLabel = "create"
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("new %s name...", m.label)
	ti.CharLimit = 200
	cmd := ti.Focus()
	m.input = ti
	m.statusMsg = ""
	m.state = stateInput
	return m, cmd
}

func (m Model) startDelete(name string) (tea.Model, tea.Cmd) {
	m.cKind = confirmDelete
	m.cMsg = fmt.Sprintf("Delete '%s'?\nThis cannot be undone.", name)
	m.cPending = filepath.Join(m.currentDir, name)
	m.statusMsg = ""
	m.state = stateConfirm
	return m, nil
}

func (m Model) startRename(name string) (tea.Model, tea.Cmd) {
	m.renameOld = name
	m.inputLabel = "rename"
	ti := textinput.New()
	ti.Placeholder = "new name..."
	ti.CharLimit = 200
	ti.SetValue(name)
	cmd := ti.Focus()
	m.input = ti
	m.statusMsg = ""
	m.state = stateRename
	return m, cmd
}

func (m Model) confirmInput() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(m.input.Value())
	if raw == "" {
		m.statusMsg = errorStyle.Render("Name cannot be empty.")
		return m, nil
	}
	slug := fs.Slugify(raw)
	if slug == "" {
		m.statusMsg = errorStyle.Render("Invalid name — nothing left after slugifying.")
		return m, nil
	}

	if slug != raw {
		// Show slugify warning — enter will auto-confirm it.
		m.cKind = confirmSlugify
		m.cMsg = fmt.Sprintf("'%s'  →  '%s'", raw, slug)
		m.cPending = slug
		m.state = stateConfirm
		return m, nil
	}

	if m.state == stateRename {
		return m.doRename(slug)
	}
	return m.doCreate(slug)
}

func (m Model) cancelInput() (tea.Model, tea.Cmd) {
	return m.backToList()
}

func (m Model) confirmAction() (tea.Model, tea.Cmd) {
	switch m.cKind {
	case confirmSlugify:
		if m.inputLabel == "rename" {
			return m.doRename(m.cPending)
		}
		return m.doCreate(m.cPending)

	case confirmDelete:
		if err := fs.DeleteDir(m.cPending); err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("Error deleting: %v", err))
			return m.backToList()
		}
		m.statusMsg = successStyle.Render(fmt.Sprintf("Deleted '%s'.", filepath.Base(m.cPending)))
		return m.backToList()

	case confirmGenCategories:
		for _, cat := range ctfCategories {
			_ = fs.MkdirAll(filepath.Join(m.cPending, cat))
		}
		m.statusMsg = successStyle.Render("Default categories created.")
		m.state = stateCategory
		m = m.loadList()
		return m, nil
	}
	return m.backToList()
}

func (m Model) cancelConfirm() (tea.Model, tea.Cmd) {
	switch m.cKind {
	case confirmSlugify:
		cmd := m.input.Focus()
		if m.inputLabel == "rename" {
			m.state = stateRename
		} else {
			m.state = stateInput
		}
		return m, cmd
	case confirmGenCategories:
		return m.goBack()
	default:
		return m.backToList()
	}
}

func (m Model) doCreate(name string) (tea.Model, tea.Cmd) {
	dir := filepath.Join(m.currentDir, name)
	if err := fs.MkdirAll(dir); err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error creating '%s': %v", name, err))
		return m.backToList()
	}
	m2, _ := m.backToList()
	mm := m2.(Model)
	return mm.selectItem(name)
}

func (m Model) doRename(newName string) (tea.Model, tea.Cmd) {
	if newName == m.renameOld {
		return m.backToList()
	}
	oldPath := filepath.Join(m.currentDir, m.renameOld)
	newPath := filepath.Join(m.currentDir, newName)
	if err := fs.RenameDir(oldPath, newPath); err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error renaming: %v", err))
		return m.backToList()
	}
	m.statusMsg = successStyle.Render(fmt.Sprintf("'%s'  →  '%s'", m.renameOld, newName))
	return m.backToList()
}

func (m Model) backToList() (tea.Model, tea.Cmd) {
	m.state = m.currentListState()
	m = m.loadList()
	return m, nil
}

func (m Model) currentListState() appState {
	if m.mode == "" {
		return stateModeSelect
	}
	rel, err := filepath.Rel(m.baseDir, m.currentDir)
	if err != nil || rel == "." {
		return statePlatform
	}
	parts := strings.Split(rel, string(filepath.Separator))
	switch len(parts) {
	case 1:
		if m.mode == "ctf" {
			return stateCategory
		}
		return stateItem
	default:
		return stateItem
	}
}

// View

func (m Model) View() string {
	if m.state == stateDone {
		if m.Err != nil {
			return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.Err))
		}
		return successStyle.Render("✓ "+m.SelectedPath) + "\n"
	}
	switch m.state {
	case stateModeSelect, statePlatform, stateCategory, stateItem, stateSearch:
		return m.viewList()
	case stateInput, stateRename:
		return m.viewInput()
	case stateConfirm:
		return m.viewConfirm()
	}
	return ""
}

func (m Model) viewList() string {
	statusLine := "  " + m.statusMsg // always one line, blank when empty
	return m.renderHeader() + "\n" +
		m.list.View() + "\n" +
		statusLine + "\n" +
		m.renderFooter()
}

func (m Model) viewInput() string {
	action := "create  " + m.label
	if m.state == stateRename {
		action = "rename  " + m.label
	}

	status := ""
	if m.statusMsg != "" {
		status = "\n" + m.statusMsg
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(action),
		"",
		m.input.View(),
		status,
		"",
		helpStyle.Render("enter: confirm  ·  esc: cancel"),
	)

	box := modalStyle.
		BorderForeground(lipgloss.Color(primary)).
		Width(clamp(m.width-20, 40, 70)).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) viewConfirm() string {
	var borderColor lipgloss.Color
	var heading, body string

	switch m.cKind {
	case confirmDelete:
		borderColor = lipgloss.Color(primary)
		heading = deleteStyle.Render("DELETE")
		body = errorStyle.Render(m.cMsg)
	case confirmSlugify:
		borderColor = lipgloss.Color("#e3b341")
		heading = warnStyle.Render("NAME WILL BE CHANGED")
		body = m.cMsg
	case confirmGenCategories:
		borderColor = lipgloss.Color(primary)
		heading = labelStyle.Render("GENERATE CATEGORIES")
		body = m.cMsg
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		heading,
		"",
		body,
		"",
		helpStyle.Render("enter / y: confirm  ·  n / esc: cancel"),
	)

	box := modalStyle.
		BorderForeground(borderColor).
		Width(clamp(m.width-20, 40, 70)).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// UI components

func (m Model) renderHeader() string {
	left := " CTF TOOL "
	right := " " + m.breadcrumb() + " "

	leftR := headerLeftStyle.Render(left)
	rightR := headerRightStyle.Render(right)

	gapW := max(m.width-lipgloss.Width(leftR)-lipgloss.Width(rightR), 0)
	gap := headerFillStyle.Render(strings.Repeat(" ", gapW))
	return leftR + gap + rightR
}

func (m Model) renderFooter() string {
	type kv struct{ k, v string }
	var pairs []kv
	if m.list.FilterState() == list.Filtering {
		pairs = []kv{{"/", "filter"}, {"esc", "cancel"}}
	} else if m.state == stateSearch {
		pairs = []kv{
			{"↑↓", "navigate"}, {"enter", "select"},
			{"/", "filter"}, {"esc", "back"},
		}
	} else if m.state == stateModeSelect {
		pairs = []kv{
			{"↑↓", "navigate"}, {"enter", "select"},
			{"/", "filter"}, {"q", "quit"},
		}
	} else {
		pairs = []kv{
			{"↑↓", "navigate"}, {"enter", "select"},
			{"n", "new"}, {"/", "filter"},
			{"d", "delete"}, {"r", "rename"},
			{"ctrl+f", "search"}, {"esc", "back"}, {"q", "quit"},
		}
	}

	// Render each element and measure its display width.
	elems := make([]string, len(pairs))
	totalElemW := 0
	for i, p := range pairs {
		e := footerKeyStyle.Render(p.k) + footerStyle.Render(" "+p.v)
		elems[i] = e
		totalElemW += lipgloss.Width(footerKeyStyle.Render(p.k)) + lipgloss.Width(" "+p.v)
	}

	// Distribute remaining space as equal gaps between (and around) elements.
	n := len(elems)
	slots := n + 1 // gaps: before first, between each pair, after last
	available := max(m.width-totalElemW, slots)
	baseGap := available / slots
	extra := available % slots

	var sb strings.Builder
	for i, e := range elems {
		g := baseGap
		if i < extra {
			g++
		}
		sb.WriteString(footerStyle.Render(strings.Repeat(" ", g)))
		sb.WriteString(e)
	}
	// trailing gap
	lastGap := baseGap
	if n < extra {
		lastGap++
	}
	sb.WriteString(footerStyle.Render(strings.Repeat(" ", lastGap)))
	return sb.String()
}

func (m Model) breadcrumb() string {
	if m.mode == "" {
		return ""
	}
	parts := []string{m.mode}
	if m.baseDir != "" && m.currentDir != m.baseDir {
		rel, err := filepath.Rel(m.baseDir, m.currentDir)
		if err == nil && rel != "." {
			for _, p := range strings.Split(rel, string(filepath.Separator)) {
				if p != "" {
					parts = append(parts, p)
				}
			}
		}
	}
	return strings.Join(parts, "  ›  ")
}

// Utilities

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		if len(path) == 1 || path[1] == '/' {
			home, _ := os.UserHomeDir()
			if len(path) == 1 {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
