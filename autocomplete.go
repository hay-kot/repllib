// This work is largely based on the AMAZING work done in the abs-lang/abs repos for their REPL
// using Bubble Tea. I used them as a source for generating this implementation for this library.
// Large swathes of this code are copied from their implementation and therefore hold the same
// license as that project (MIT).
//
// Source: https://github.com/abs-lang/abs/tree/2.7.0/terminal
package repllib

import (
	"slices"
	"sort"
	"strings"

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
	optMatchAny bool
	functions   map[string]string
	identifiers map[string]string
	keywords    []string
	delegate    map[string]SuggestionProvider
}

func NewDefaultSuggestionProvider() *defaultSuggestionProvider {
	return &defaultSuggestionProvider{
		functions:   make(map[string]string),
		identifiers: make(map[string]string),
		keywords:    []string{},
		delegate:    make(map[string]SuggestionProvider),
	}
}

func (p *defaultSuggestionProvider) SetMatchAny(match bool) *defaultSuggestionProvider {
	p.optMatchAny = match
	return p
}

func (p *defaultSuggestionProvider) AddDelegate(name string, provider SuggestionProvider) *defaultSuggestionProvider {
	p.delegate[name] = provider
	return p
}

func (p *defaultSuggestionProvider) RemoveDelegate(name string) *defaultSuggestionProvider {
	delete(p.delegate, name)
	return p
}

func (p *defaultSuggestionProvider) AddFunction(name, description string) *defaultSuggestionProvider {
	p.functions[name] = description
	return p
}

func (p *defaultSuggestionProvider) RemoveFunction(name string) *defaultSuggestionProvider {
	delete(p.functions, name)
	return p
}

func (p *defaultSuggestionProvider) AddIdentifier(name, value string) *defaultSuggestionProvider {
	p.identifiers[name] = value
	return p
}

func (p *defaultSuggestionProvider) RemoveIdentifier(name string) *defaultSuggestionProvider {
	delete(p.identifiers, name)
	return p
}

func (p *defaultSuggestionProvider) AddKeyword(keyword string) *defaultSuggestionProvider {
	p.keywords = append(p.keywords, keyword)
	return p
}

func (p *defaultSuggestionProvider) RemoveKeyword(keyword string) *defaultSuggestionProvider {
	for i, k := range p.keywords {
		if k == keyword {
			p.keywords = slices.Delete(p.keywords, i, i+1)
			break
		}
	}
	return p
}

func (p *defaultSuggestionProvider) GetSuggestions(input string, cursorPos int) ([]Suggestion, string, error) {
	for name, provider := range p.delegate {
		if strings.HasPrefix(input, name+" ") {
			trimmed := strings.TrimSpace(strings.TrimPrefix(input, name+" "))
			return provider.GetSuggestions(trimmed, cursorPos)
		}
	}

	// Find the word at cursor position
	wordStart, wordEnd := findWordBounds(input, cursorPos)
	word := input[wordStart:wordEnd]

	if len(word) == 0 {
		if p.optMatchAny {
			suggestions := []Suggestion{}

			for name, desc := range p.functions {
				suggestions = append(suggestions, NewSuggestion(name, SuggestionFunction, trimString(desc, 50)))
			}

			for name, value := range p.identifiers {
				suggestions = append(suggestions, NewSuggestion(name, SuggestionIdentifier, trimString(value, 50)))
			}

			for _, keyword := range p.keywords {
				suggestions = append(suggestions, NewSuggestion(keyword, SuggestionKeyword, ""))
			}

			return atMostN(suggestions, 8), "", nil
		}
		return []Suggestion{}, "", nil
	}

	var suggestions []Suggestion

	// Match functions
	for name, desc := range p.functions {
		if prefixMatch(name, word) {
			suggestions = append(suggestions, NewSuggestion(name, SuggestionFunction, trimString(desc, 50)))
		}
	}

	// Match identifiers
	for name, value := range p.identifiers {
		if prefixMatch(name, word) {
			suggestions = append(suggestions, NewSuggestion(name, SuggestionIdentifier, trimString(value, 50)))
		}
	}

	// Match keywords
	for _, keyword := range p.keywords {
		if prefixMatch(keyword, word) {
			suggestions = append(suggestions, NewSuggestion(keyword, SuggestionKeyword, ""))
		}
	}

	// Sort suggestions by type priority (functions first, then identifiers, then keywords)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Type < suggestions[j].Type
	})

	return atMostN(suggestions, 8), word, nil
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
