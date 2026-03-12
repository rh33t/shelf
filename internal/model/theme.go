package model

import "github.com/charmbracelet/lipgloss"

const primary = "#7a0e13"

var (
	// Header
	headerLeftStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(primary)).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)
	headerRightStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(primary)).
				Foreground(lipgloss.Color("#ffcccc"))
	headerFillStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(primary))

	// List title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(primary)).
			PaddingLeft(2).
			PaddingTop(2).
			PaddingBottom(1)

	// Item delegate
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(primary)).
				Bold(true)
	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0c0c0"))
	dimItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	// Footer
	footerBg       = lipgloss.Color("#1a1a1a")
	footerStyle    = lipgloss.NewStyle().Background(footerBg).Foreground(lipgloss.Color("#888888"))
	footerKeyStyle = lipgloss.NewStyle().Background(footerBg).Foreground(lipgloss.Color(primary)).Bold(true)

	// Status / feedback
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3fb950"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(primary))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#e3b341"))
	labelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(primary))
	deleteStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(primary))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	// Modal
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3)
)
