// This work is largely based on the AMAZING work done in the abs-lang/abs repos for their REPL
// using Bubble Tea. I used them as a source for generating this implementation for this library.
// Large swathes of this code are copied from their implementation and therefore hold the same
// license as that project (MIT).
//
// Source: https://github.com/abs-lang/abs/tree/2.7.0/terminal
package repllib

import (
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

type SuggestionType int

const (
	SuggestionFunction SuggestionType = iota
	SuggestionIdentifier
	SuggestionProperty
	SuggestionKeyword
)

type Suggestion struct {
	Value   string
	Comment string
	Type    SuggestionType
}

func NewSuggestion(value string, suggType SuggestionType, comment string) Suggestion {
	return Suggestion{
		Value:   value,
		Type:    suggType,
		Comment: comment,
	}
}

// SuggestionProvider interface for getting suggestions
type SuggestionProvider interface {
	GetSuggestions(input string, cursorPos int) ([]Suggestion, string, error)
}

// Default suggestion provider that uses simple prefix matching
type defaultSuggestionProvider struct {
	functions   map[string]string // function name -> description
	identifiers map[string]string // identifier name -> value
	keywords    []string
}

func NewDefaultSuggestionProvider() *defaultSuggestionProvider {
	return &defaultSuggestionProvider{
		functions:   make(map[string]string),
		identifiers: make(map[string]string),
		keywords:    []string{},
	}
}

func (p *defaultSuggestionProvider) AddFunction(name, description string) *defaultSuggestionProvider {
	p.functions[name] = description
	return p
}

func (p *defaultSuggestionProvider) AddIdentifier(name, value string) *defaultSuggestionProvider {
	p.identifiers[name] = value
	return p
}

func (p *defaultSuggestionProvider) AddKeyword(keyword string) *defaultSuggestionProvider {
	p.keywords = append(p.keywords, keyword)
	return p
}

func (p *defaultSuggestionProvider) GetSuggestions(input string, cursorPos int) ([]Suggestion, string, error) {
	// Find the word at cursor position
	wordStart, wordEnd := findWordBounds(input, cursorPos)
	word := input[wordStart:wordEnd]

	if len(word) == 0 {
		return []Suggestion{}, "", nil
	}

	var suggestions []Suggestion

	// Match functions
	for name, desc := range p.functions {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(word)) {
			comment := desc
			if len(comment) > 50 {
				comment = comment[:50] + "..."
			}
			suggestions = append(suggestions, NewSuggestion(name, SuggestionFunction, comment))
		}
	}

	// Match identifiers
	for name, value := range p.identifiers {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(word)) {
			comment := value
			if len(comment) > 50 {
				comment = comment[:50] + "..."
			}
			suggestions = append(suggestions, NewSuggestion(name, SuggestionIdentifier, comment))
		}
	}

	// Match keywords
	for _, keyword := range p.keywords {
		if strings.HasPrefix(strings.ToLower(keyword), strings.ToLower(word)) {
			suggestions = append(suggestions, NewSuggestion(keyword, SuggestionKeyword, ""))
		}
	}

	// Sort suggestions by type priority (functions first, then identifiers, then keywords)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Type < suggestions[j].Type
	})

	return suggestions, word, nil
}

// Find word boundaries around cursor position
func findWordBounds(input string, cursorPos int) (start, end int) {
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	// Find start of word (go backwards from cursor)
	start = cursorPos
	for start > 0 && isWordChar(rune(input[start-1])) {
		start--
	}

	// Find end of word (go forwards from cursor)
	end = cursorPos
	for end < len(input) && isWordChar(rune(input[end])) {
		end++
	}

	return start, end
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Apply suggestion to input string
func applySuggestion(input, textToReplace, suggestion string, cursorPos int) (string, int) {
	// Find where the word to replace starts
	wordStart, _ := findWordBounds(input, cursorPos)

	// Replace the word
	head := input[:wordStart]
	tail := input[wordStart+len(textToReplace):]
	newInput := head + suggestion + tail
	newCursorPos := wordStart + len(suggestion)

	return newInput, newCursorPos
}

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

func renderSuggestions(suggestions []Suggestion, selectedIndex int) string {
	if len(suggestions) == 0 {
		return ""
	}

	var lines []string

	for i, sugg := range suggestions {
		style := styleSuggestions[sugg.Type]
		prefix := "   "
		text := style.Render(sugg.Value)

		if selectedIndex == i {
			prefix = styleSelectedPrefix.Render(" → ")
			text = styleSelectedSuggestion.Render(sugg.Value)

			if sugg.Comment != "" {
				text += " " + styleComment.Render("# "+sugg.Comment)
			}
		}

		lines = append(lines, prefix+text)
	}

	return styleSuggestionBox.Render(strings.Join(lines, "\n"))
}
