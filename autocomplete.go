// This work is largely based on the AMAZING work done in the abs-lang/abs repos for their REPL
// using Bubble Tea. I used them as a source for generating this implementation for this library.
// Large swathes of this code are copied from their implementation and therefore hold the same
// license as that project (MIT).
//
// Source: https://github.com/abs-lang/abs/tree/2.7.0/terminal
package repllib

import (
	"cmp"
	"slices"
	"sort"
	"strings"
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

// Suggester interface for getting suggestions
type Suggester interface {
	GetSuggestions(input string, cursorPos int) ([]Suggestion, string, error)
}

// Default suggestion provider that uses simple prefix matching
type SuggestionProvider struct {
	optMatchAny bool
	functions   map[string]string
	identifiers map[string]string
	keywords    []string
	delegate    map[string]Suggester
}

func NewSuggestionProvider() *SuggestionProvider {
	return &SuggestionProvider{
		functions:   make(map[string]string),
		identifiers: make(map[string]string),
		keywords:    []string{},
		delegate:    make(map[string]Suggester),
	}
}

func (p *SuggestionProvider) SetMatchAny(match bool) *SuggestionProvider {
	p.optMatchAny = match
	return p
}

func (p *SuggestionProvider) AddDelegate(name string, provider Suggester) *SuggestionProvider {
	p.delegate[name] = provider
	return p
}

func (p *SuggestionProvider) RemoveDelegate(name string) *SuggestionProvider {
	delete(p.delegate, name)
	return p
}

func (p *SuggestionProvider) AddFunction(name, description string) *SuggestionProvider {
	p.functions[name] = description
	return p
}

func (p *SuggestionProvider) RemoveFunction(name string) *SuggestionProvider {
	delete(p.functions, name)
	return p
}

func (p *SuggestionProvider) AddIdentifier(name, value string) *SuggestionProvider {
	p.identifiers[name] = value
	return p
}

func (p *SuggestionProvider) RemoveIdentifier(name string) *SuggestionProvider {
	delete(p.identifiers, name)
	return p
}

func (p *SuggestionProvider) AddKeyword(keyword string) *SuggestionProvider {
	p.keywords = append(p.keywords, keyword)
	return p
}

func (p *SuggestionProvider) RemoveKeyword(keyword string) *SuggestionProvider {
	for i, k := range p.keywords {
		if k == keyword {
			p.keywords = slices.Delete(p.keywords, i, i+1)
			break
		}
	}
	return p
}

func (p *SuggestionProvider) GetSuggestions(input string, cursorPos int) ([]Suggestion, string, error) {
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
			funcSugs := []Suggestion{}
			for name, desc := range p.functions {
				funcSugs = append(funcSugs, NewSuggestion(name, SuggestionFunction, trimString(desc, 50)))
			}

			slices.SortFunc(funcSugs, func(a, b Suggestion) int {
				return cmp.Compare(a.Value, b.Value)
			})

			idSugs := []Suggestion{}
			for name, value := range p.identifiers {
				idSugs = append(idSugs, NewSuggestion(name, SuggestionIdentifier, trimString(value, 50)))
			}

			slices.SortFunc(idSugs, func(a, b Suggestion) int {
				return cmp.Compare(a.Value, b.Value)
			})

			kwSugs := []Suggestion{}
			for _, keyword := range p.keywords {
				kwSugs = append(kwSugs, NewSuggestion(keyword, SuggestionKeyword, ""))
			}

			suggestions := []Suggestion{}
			suggestions = slices.Concat(suggestions, funcSugs)
			suggestions = slices.Concat(suggestions, idSugs)
			suggestions = slices.Concat(suggestions, kwSugs)

			return atMostN(suggestions, 8), "", nil
		}
		return []Suggestion{}, "", nil
	}

	var suggestions []Suggestion

	for name, desc := range p.functions {
		sugs := []Suggestion{}
		if autocompleteMatch(name, word) {
			sugs = append(sugs, NewSuggestion(name, SuggestionFunction, trimString(desc, 50)))
		}

		slices.SortFunc(sugs, func(a, b Suggestion) int {
			return cmp.Compare(a.Value, b.Value)
		})
	}

	for name, value := range p.identifiers {
		sugs := []Suggestion{}
		if autocompleteMatch(name, word) {
			sugs = append(sugs, NewSuggestion(name, SuggestionIdentifier, trimString(value, 50)))
		}

		slices.SortFunc(sugs, func(a, b Suggestion) int {
			return cmp.Compare(a.Value, b.Value)
		})
	}

	for _, keyword := range p.keywords {
		sugs := []Suggestion{}
		if autocompleteMatch(keyword, word) {
			sugs = append(sugs, NewSuggestion(keyword, SuggestionKeyword, ""))
		}

		slices.SortFunc(sugs, func(a, b Suggestion) int {
			return cmp.Compare(a.Value, b.Value)
		})
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Type < suggestions[j].Type
	})

	return atMostN(suggestions, 8), word, nil
}

// Apply suggestion to input string
func applySuggestion(input, textToReplace, suggestion string, cursorPos int) (string, int) {
	// Find where the word to replace starts
	wordStart, wordEnd := findWordBounds(input, cursorPos)

	// Calculate the actual end position for replacement
	replaceEnd := max(wordEnd, min(wordStart+len(textToReplace), len(input)))

	// Replace the word
	head := input[:wordStart]
	tail := input[replaceEnd:]
	newInput := head + suggestion + tail
	newCursorPos := wordStart + len(suggestion)

	return newInput, newCursorPos
}

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
