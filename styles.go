package repllib

import "github.com/charmbracelet/lipgloss"

// Styling
var (
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render // Red

	stylePrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
	styleHelp   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Gray
)

// Styling for suggestions
var (
	styleSuggestions = map[SuggestionType]lipgloss.Style{
		SuggestionFunction:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
		SuggestionIdentifier: lipgloss.NewStyle().Foreground(lipgloss.Color("220")), // Yellow
		SuggestionProperty:   lipgloss.NewStyle().Foreground(lipgloss.Color("14")),  // Cyan
		SuggestionKeyword:    lipgloss.NewStyle().Foreground(lipgloss.Color("13")),  // Magenta
	}
	styleSelectedSuggestion = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Underline(true) // Bright green
	styleSelectedPrefix     = styleSelectedSuggestion.Underline(false)
	styleSuggestionBox      = lipgloss.NewStyle().PaddingLeft(1).PaddingTop(1)
	styleComment            = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Faint(true) // Gray
)
