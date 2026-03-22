package model

import "github.com/charmbracelet/lipgloss"

var primary, secondary string

var (
	// Header
	headerLeftStyle  lipgloss.Style
	headerRightStyle lipgloss.Style
	headerFillStyle  lipgloss.Style

	// List title
	titleStyle lipgloss.Style

	// Item delegate
	selectedItemStyle lipgloss.Style
	normalItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#c0c0c0"))
	dimItemStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	// Footer
	footerStyle    lipgloss.Style
	footerKeyStyle lipgloss.Style

	// Status / feedback
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3fb950"))
	errorStyle   lipgloss.Style
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#e3b341"))
	labelStyle   lipgloss.Style
	deleteStyle  lipgloss.Style
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	// Modal
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3)
)

func initStyles(p, s string) {
	primary = p
	secondary = s

	pc := lipgloss.Color(p)
	sc := lipgloss.Color(s)

	headerLeftStyle = lipgloss.NewStyle().
		Background(pc).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)
	headerRightStyle = lipgloss.NewStyle().
		Background(pc).
		Foreground(lipgloss.Color("#ffcccc"))
	headerFillStyle = lipgloss.NewStyle().
		Background(pc)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(pc).
		PaddingLeft(2).
		PaddingTop(2).
		PaddingBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
		Foreground(pc).
		Bold(true)

	footerStyle = lipgloss.NewStyle().Background(sc).Foreground(lipgloss.Color("#888888"))
	footerKeyStyle = lipgloss.NewStyle().Background(sc).Foreground(pc).Bold(true)

	errorStyle = lipgloss.NewStyle().Foreground(pc)
	labelStyle = lipgloss.NewStyle().Bold(true).Foreground(pc)
	deleteStyle = lipgloss.NewStyle().Bold(true).Foreground(pc)
}
