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
	stateLevel0              // ctf source / box platform
	stateLevel1              // ctf category
	stateLeaf                // challenge / box
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
)

// Custom item delegate

type item struct {
	name string
	// ghost marks a configured default that is not on disk yet. It is listed
	// so it can be picked, and created only when it is.
	ghost bool
}

func (i item) FilterValue() string { return i.name }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	selected := index == m.Index()

	gutter := strings.Repeat(" ", gutterCol)
	style := normalItemStyle
	switch {
	case selected:
		gutter = markerStyle.Render(marker) + strings.Repeat(" ", gutterCol-lipgloss.Width(marker))
		style = selectedItemStyle
	case m.FilterState() == list.Filtering:
		style = dimItemStyle
	case i.ghost:
		style = ghostItemStyle
	}

	line := strings.Repeat(" ", padCol) + gutter + style.Render(i.name)

	if i.ghost {
		gap := m.Width() - padCol - textCol - lipgloss.Width(i.name) - len(ghostTag)
		if gap > 0 {
			line += strings.Repeat(" ", gap) + ghostTagStyle.Render(ghostTag)
		}
	}

	fmt.Fprint(w, line)
}

// History

type histEntry struct {
	state      appState
	currentDir string
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
	m := Model{
		mode:   mode,
		width:  80,
		height: 24,
		cfg:    cfg,
	}
	if mode == "" {
		m.state = stateModeSelect
		m = m.loadList()
		return m
	}

	baseDir := m.baseDirForMode(mode)
	if err := fs.MkdirAll(baseDir); err != nil {
		m.Err = err
		m.state = stateDone
		return m
	}
	m.baseDir = baseDir
	m.currentDir = baseDir
	m.state = stateLevel0
	return m.loadList()
}

func (m Model) baseDirForMode(mode string) string {
	if mode == "ctf" {
		return filepath.Join(m.cfg.BaseDir, "challenges")
	}
	return filepath.Join(m.cfg.BaseDir, "boxes")
}

// levelLabel names the thing being picked at the current level.
func (m Model) levelLabel() string {
	switch m.state {
	case stateLevel0:
		if m.mode == "ctf" {
			return "source"
		}
		return "platform"
	case stateLevel1:
		return "category"
	case stateLeaf:
		if m.mode == "ctf" {
			return "challenge"
		}
		return "box"
	}
	return ""
}

// defaultsForLevel returns the configured entries offered at this level. They
// are listed whether or not they exist and created only on selection.
func (m Model) defaultsForLevel() []string {
	switch {
	case m.state == stateLevel0 && m.mode == "ctf":
		return m.cfg.CTFSources
	case m.state == stateLevel0 && m.mode == "box":
		return m.cfg.BoxPlatforms
	case m.state == stateLevel1 && m.mode == "ctf":
		return m.cfg.CTFCategories
	}
	return nil
}

// List helpers

func (m Model) listHeight() int {
	// header, blank, label row, blank, status, footer
	return max(m.height-6, 3)
}

func (m Model) initList(items []item) Model {
	li := make([]list.Item, len(items))
	for i, it := range items {
		li[i] = it
	}

	l := list.New(li, itemDelegate{}, m.width, m.listHeight())
	l.SetShowTitle(false)
	l.SetShowFilter(false) // the filter prompt is drawn on the status row
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(true)
	l.FilterInput.Prompt = "/"
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(accent)
	m.list = l
	return m
}

func (m Model) loadList() Model {
	m.label = m.levelLabel()

	if m.state == stateModeSelect {
		return m.initList([]item{{name: "ctf"}, {name: "box"}})
	}

	dirs, err := fs.ListDirs(m.currentDir)
	if err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error reading directory: %v", err))
		dirs = nil
	}

	items := make([]item, 0, len(dirs))
	seen := make(map[string]bool, len(dirs))
	for _, d := range dirs {
		seen[d] = true
		items = append(items, item{name: d})
	}
	for _, d := range m.defaultsForLevel() {
		if !seen[d] {
			items = append(items, item{name: d, ghost: true})
		}
	}

	return m.initList(items)
}

func (m Model) sectionLabel() string {
	switch m.state {
	case stateModeSelect:
		return "MODE"
	case stateSearch:
		return "JUMP TO"
	}
	return strings.ToUpper(m.label)
}

// tea.Model

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, m.listHeight())
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m.delegateUpdate(msg)
}

func (m Model) delegateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case stateModeSelect, stateLevel0, stateLevel1, stateLeaf, stateSearch:
		m.list, cmd = m.list.Update(msg)
	case stateInput, stateRename:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m Model) isListState() bool {
	switch m.state {
	case stateModeSelect, stateLevel0, stateLevel1, stateLeaf, stateSearch:
		return true
	}
	return false
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isListState() && m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch {
	case m.isListState():
		return m.handleListKey(msg)
	case m.state == stateInput || m.state == stateRename:
		return m.handleInputKey(msg)
	case m.state == stateConfirm:
		return m.handleConfirmKey(msg)
	}
	return m, nil
}

// Key handlers

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	// left/h and right/l are bound to pagination by the list widget. Claiming
	// them here shadows that, so they walk the tree the way yazi does.
	case "esc", "left", "h":
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

	case "enter", " ", "right", "l":
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
			// Only a target is worth opening. Jumping to a level above one
			// browses into it instead, so esc still returns where ctrl+f began.
			if !m.isTarget(path) {
				m.currentDir = path
				m.state = m.currentListState()
				m.searchBase = ""
				return m.loadList(), nil
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
		sel, ok := m.editableSelection()
		if !ok {
			return m, nil
		}
		return m.startDelete(sel.name)

	case "r":
		sel, ok := m.editableSelection()
		if !ok {
			return m, nil
		}
		return m.startRename(sel.name)

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

// editableSelection returns the highlighted item if it can be renamed or
// deleted. A ghost has nothing on disk to act on.
func (m *Model) editableSelection() (item, bool) {
	if m.state == stateSearch || m.state == stateModeSelect {
		return item{}, false
	}
	sel := m.list.SelectedItem()
	if sel == nil {
		return item{}, false
	}
	it := sel.(item)
	if it.ghost {
		m.statusMsg = warnStyle.Render(fmt.Sprintf("'%s' does not exist yet.", it.name))
		return item{}, false
	}
	return it, true
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
	if m.state == stateModeSelect {
		m.mode = name
		m.baseDir = m.baseDirForMode(name)
		if err := fs.MkdirAll(m.baseDir); err != nil {
			m.Err = err
			m.state = stateDone
			return m, tea.Quit
		}
		m.pushHistory(stateModeSelect)
		m.currentDir = m.baseDir
		m.state = stateLevel0
		return m.loadList(), nil
	}

	dir := filepath.Join(m.currentDir, name)
	if err := fs.MkdirAll(dir); err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", err))
		return m, nil
	}

	if m.state == stateLeaf {
		m.writeNotes(dir, name)
		return m.finish(dir)
	}

	m.pushHistory(m.state)
	m.currentDir = dir
	if m.state == stateLevel0 && m.mode == "ctf" {
		m.state = stateLevel1
	} else {
		m.state = stateLeaf
	}
	return m.loadList(), nil
}

// writeNotes seeds notes.md for a newly picked target, leaving an existing one
// alone. currentDir is the parent level: the platform for a box, the category
// for a challenge.
func (m Model) writeNotes(dir, name string) {
	if _, err := os.Stat(filepath.Join(dir, "notes.md")); !os.IsNotExist(err) {
		return
	}
	if m.mode == "box" {
		_ = fs.WriteBoxNotes(dir, filepath.Base(m.currentDir), name)
		return
	}
	source := filepath.Base(filepath.Dir(m.currentDir))
	_ = fs.WriteChallengeNotes(dir, source, filepath.Base(m.currentDir), name)
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
	m.statusMsg = ""
	if entry.state == stateModeSelect {
		m.mode = ""
	}
	return m.loadList(), nil
}

func (m *Model) pushHistory(s appState) {
	m.history = append(m.history, histEntry{state: s, currentDir: m.currentDir})
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

	items := make([]item, len(dirs))
	for i, d := range dirs {
		items[i] = item{name: d}
	}
	m = m.initList(items)

	return m, func() tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	}
}

func (m Model) startCreate() (tea.Model, tea.Cmd) {
	m.inputLabel = "create"
	ti := textinput.New()
	ti.Placeholder = "name..."
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
	ti.Placeholder = "name..."
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
		m.statusMsg = errorStyle.Render("Invalid name, nothing left after slugifying.")
		return m, nil
	}

	if slug != raw {
		// Show slugify warning, enter will auto-confirm it.
		m.cKind = confirmSlugify
		m.cMsg = fmt.Sprintf("'%s'   →   '%s'", raw, slug)
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
	}
	return m.backToList()
}

func (m Model) cancelConfirm() (tea.Model, tea.Cmd) {
	if m.cKind == confirmSlugify {
		cmd := m.input.Focus()
		if m.inputLabel == "rename" {
			m.state = stateRename
		} else {
			m.state = stateInput
		}
		return m, cmd
	}
	return m.backToList()
}

func (m Model) doCreate(name string) (tea.Model, tea.Cmd) {
	dir := filepath.Join(m.currentDir, name)
	if err := fs.MkdirAll(dir); err != nil {
		m.statusMsg = errorStyle.Render(fmt.Sprintf("Error creating '%s': %v", name, err))
		return m.backToList()
	}
	m2, _ := m.backToList()
	return m2.(Model).selectItem(name)
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
	m.statusMsg = successStyle.Render(fmt.Sprintf("'%s' → '%s'", m.renameOld, newName))
	return m.backToList()
}

func (m Model) backToList() (tea.Model, tea.Cmd) {
	m.state = m.currentListState()
	return m.loadList(), nil
}

// isTarget reports whether path is a challenge or a box, the deepest level and
// the only one shelf opens a session on.
func (m Model) isTarget(path string) bool {
	rel, err := filepath.Rel(m.baseDir, path)
	if err != nil || rel == "." {
		return false
	}
	depth := len(strings.Split(rel, string(filepath.Separator)))
	if m.mode == "ctf" {
		return depth >= 3
	}
	return depth >= 2
}

// currentListState recovers the level from how deep currentDir sits under the
// mode root, for returning from a modal.
func (m Model) currentListState() appState {
	if m.mode == "" {
		return stateModeSelect
	}
	rel, err := filepath.Rel(m.baseDir, m.currentDir)
	if err != nil || rel == "." {
		return stateLevel0
	}
	if len(strings.Split(rel, string(filepath.Separator))) == 1 && m.mode == "ctf" {
		return stateLevel1
	}
	return stateLeaf
}

// View

func (m Model) View() string {
	if m.state == stateDone {
		if m.Err != nil {
			return errorStyle.Render(fmt.Sprintf("Error: %v\n", m.Err))
		}
		return ""
	}
	switch m.state {
	case stateInput, stateRename:
		return m.viewInput()
	case stateConfirm:
		return m.viewConfirm()
	}
	return m.viewList()
}

func (m Model) viewList() string {
	return strings.Join([]string{
		m.renderHeader(),
		"",
		m.renderLabel(),
		"",
		m.listBody(),
		m.renderStatus(),
		m.renderFooter(),
	}, "\n")
}

func (m Model) listBody() string {
	if len(m.list.VisibleItems()) == 0 && m.list.FilterState() != list.Filtering {
		return lipgloss.NewStyle().Height(m.listHeight()).Render(noItemsStyle.Render("nothing here yet"))
	}
	return m.list.View()
}

func (m Model) viewInput() string {
	action := "NEW"
	if m.state == stateRename {
		action = "RENAME"
	}

	status := ""
	if m.statusMsg != "" {
		status = "\n" + m.statusMsg
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(action+" "+strings.ToUpper(m.label)),
		"",
		m.input.View(),
		status,
		"",
		helpStyle.Render("enter: confirm   esc: cancel"),
	)

	return m.centered(accent, content)
}

func (m Model) viewConfirm() string {
	var borderColor lipgloss.Color
	var heading, body string

	switch m.cKind {
	case confirmDelete:
		borderColor = red
		heading = deleteStyle.Render("DELETE")
		body = errorStyle.Render(m.cMsg)
	case confirmSlugify:
		borderColor = yellow
		heading = warnStyle.Render("RENAMING TO A SLUG")
		body = m.cMsg
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		heading,
		"",
		body,
		"",
		helpStyle.Render("enter / y: confirm   n / esc: cancel"),
	)

	return m.centered(borderColor, content)
}

func (m Model) centered(border lipgloss.Color, content string) string {
	box := modalStyle.
		BorderForeground(border).
		Width(clamp(m.width-20, 40, 70)).
		Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// UI components

// spread renders left and right on one line, pushed to opposite edges.
func (m Model) spread(leftIndent int, left, right string) string {
	l := strings.Repeat(" ", leftIndent) + left
	r := right + strings.Repeat(" ", padCol)
	gap := max(m.width-lipgloss.Width(l)-lipgloss.Width(r), 0)
	return l + strings.Repeat(" ", gap) + r
}

func (m Model) renderHeader() string {
	return m.spread(padCol, appNameStyle.Render("shelf"), crumbStyle.Render(m.breadcrumb()))
}

func (m Model) renderLabel() string {
	return m.spread(textCol, sectionStyle.Render(m.sectionLabel()), counterStyle.Render(m.counter()))
}

func (m Model) counter() string {
	n := len(m.list.VisibleItems())
	if n == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", m.list.Index()+1, n)
}

func (m Model) renderStatus() string {
	if m.list.FilterState() == list.Filtering {
		return strings.Repeat(" ", textCol) + m.list.FilterInput.View()
	}
	if m.statusMsg == "" {
		return ""
	}
	return strings.Repeat(" ", textCol) + m.statusMsg
}

type keyHint struct{ k, v string }

func (m Model) renderFooter() string {
	var hints []keyHint
	switch {
	case m.list.FilterState() == list.Filtering:
		hints = []keyHint{{"⏎", "apply"}, {"esc", "cancel"}}
	case m.state == stateSearch:
		hints = []keyHint{{"↑↓", "move"}, {"⏎", "open"}, {"esc", "back"}}
	case m.state == stateModeSelect:
		hints = []keyHint{{"↑↓", "move"}, {"⏎", "select"}, {"q", "quit"}}
	default:
		// No navigation hint here: the full set overflows 80 columns, and the
		// mode screen already shows it.
		hints = []keyHint{
			{"⏎", "open"}, {"n", "new"}, {"r", "rename"}, {"d", "delete"},
			{"/", "filter"}, {"^f", "jump"}, {"esc", "back"}, {"q", "quit"},
		}
	}

	// Drop hints from the right until the line fits the terminal.
	for len(hints) > 1 && hintsWidth(hints) > m.width-2*padCol {
		hints = hints[:len(hints)-1]
	}

	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = footerKeyStyle.Render(h.k) + footerStyle.Render(" "+h.v)
	}
	return strings.Repeat(" ", padCol) + strings.Join(parts, footerStyle.Render("  "))
}

func hintsWidth(hints []keyHint) int {
	w := 2 * (len(hints) - 1) // separators
	for _, h := range hints {
		w += lipgloss.Width(h.k) + 1 + lipgloss.Width(h.v)
	}
	return w
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
	return strings.Join(parts, " › ")
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
