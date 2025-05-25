package repllib

import "github.com/charmbracelet/lipgloss"

// Styling
var (
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render // Red

	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Gray
)
