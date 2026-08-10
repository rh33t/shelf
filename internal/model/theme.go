package model

import "github.com/charmbracelet/lipgloss"

// ANSI 0-15 resolve against the terminal's own palette, so shelf inherits
// whatever colorscheme is already configured.
const (
	accent = lipgloss.Color("4")
	dim    = lipgloss.Color("8")
	red    = lipgloss.Color("1")
	green  = lipgloss.Color("2")
	yellow = lipgloss.Color("3")
)

// Every region lines up on one text column. The selection marker hangs in the
// gutter to its left, so names never shift as the cursor moves.
const (
	padCol    = 2
	gutterCol = 3
	textCol   = padCol + gutterCol
)

const (
	marker   = "▍"
	ghostTag = "not created"
)

var (
	// Header
	appNameStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	crumbStyle   = lipgloss.NewStyle().Foreground(dim)

	// Label row
	sectionStyle = lipgloss.NewStyle().Foreground(dim)
	counterStyle = lipgloss.NewStyle().Foreground(dim)

	// Items
	markerStyle       = lipgloss.NewStyle().Foreground(accent)
	selectedItemStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	normalItemStyle   = lipgloss.NewStyle()
	ghostItemStyle    = lipgloss.NewStyle().Foreground(dim)
	ghostTagStyle     = lipgloss.NewStyle().Foreground(dim).Faint(true)
	dimItemStyle      = lipgloss.NewStyle().Foreground(dim).Faint(true)
	noItemsStyle      = lipgloss.NewStyle().Foreground(dim).PaddingLeft(textCol)

	// Footer
	footerStyle    = lipgloss.NewStyle().Foreground(dim)
	footerKeyStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)

	// Status
	successStyle = lipgloss.NewStyle().Foreground(green)
	errorStyle   = lipgloss.NewStyle().Foreground(red)
	warnStyle    = lipgloss.NewStyle().Foreground(yellow)
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	deleteStyle  = lipgloss.NewStyle().Bold(true).Foreground(red)
	helpStyle    = lipgloss.NewStyle().Foreground(dim)

	// Modal
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3)
)
