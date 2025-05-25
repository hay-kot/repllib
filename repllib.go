// Package repllib provides a simple REPL (Read-Eval-Print Loop) implementation with built in
// features for evaluation, history, and autocomplete to simplify building REPLs
package repllib

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// EvalFunc is the required evaluation function
type EvalFunc func(buffer string) string

// EvalMiddleware is an optional middleware function that can modify the evaluation process.
type EvalMiddleware func(buffer string) (string, tea.Cmd)

func WithExitMiddleware() EvalMiddleware {
	return func(buffer string) (string, tea.Cmd) {
		kw := [...]string{
			"exit",
			"quit",
			":exit",
			":quit",
		}

		for _, kw := range kw {
			if strings.EqualFold(strings.TrimSpace(buffer), kw) {
				return "Exiting REPL...", tea.Quit
			}
		}

		return buffer, nil // No exit command found
	}
}

// Optional function types for builder
type (
	PromptFunc func(count int) string
	TabFunc    func(buffer string) string
)

// Main Repl struct that implements tea.Model
type Repl struct {
	handler     Handler
	history     ReplHistory
	textInput   textinput.Model
	historyIdx  int
	promptCount int // Track the prompt number like IPython
	quitting    bool

	// Autocomplete state
	suggestionProvider Suggester
	suggestions        []Suggestion
	selectedSuggestion int
	textToReplace      string
	dirtyInput         string // Input before showing suggestions
	isSuggesting       bool
}

// Bubble Tea Model implementation
func (r *Repl) Init() tea.Cmd {
	// Initialize text input
	r.textInput = textinput.New()
	r.textInput.Focus()
	r.promptCount = 1 // Start at 1 like IPython
	r.textInput.Prompt = r.handler.Prompt(r.promptCount)
	r.textInput.Width = 80

	r.historyIdx = -1
	r.selectedSuggestion = -1
	return textinput.Blink
}

func (r *Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		// Handle autocomplete navigation first
		if r.isSuggesting {
			switch msg.Type {
			case tea.KeyEnter:
				return r.selectSuggestion(), nil
			case tea.KeyTab, tea.KeyDown:
				return r.navigateSuggestions(1), nil
			case tea.KeyUp:
				return r.navigateSuggestions(-1), nil
			case tea.KeyEsc:
				return r.exitSuggestions(), nil
			default:
				// Any other key exits suggestions and processes normally
				r = r.exitSuggestions()
			}
		}

		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(r.textInput.Value())

			nextPrompt := func() {
				// Increment prompt counter and reset input
				r.promptCount++
				r.textInput.SetValue("")
				r.textInput.Prompt = r.handler.Prompt(r.promptCount)
				r.historyIdx = -1
			}

			switch input {
			case "exit", "quit", ":exit", ":quit":
				r.quitting = true
				nextPrompt()
				return r, tea.Quit // Exit the REPL
			case "clear":
				nextPrompt()
				return r, tea.ClearScreen // Clear the screen
			}

			if input != "" {
				result := r.handler.Eval(input)

				// Ensure extra empty line
				if result != "" && !strings.HasSuffix(result, "\n") {
					result += "\n"
				}

				nextPrompt()

				// Add to history
				_ = r.history.Push(input) // TO-DO: log error

				// Use tea.Println to display the result
				result = r.handler.Prompt(r.promptCount-1) + input + "\n" + result
				return r, tea.Println(result)
			}
			return r, nil

		case tea.KeyUp:
			// Navigate history backwards
			history, err := r.history.GetAll()
			if err == nil && len(history) > 0 {
				if r.historyIdx == -1 {
					r.historyIdx = len(history) - 1
				} else if r.historyIdx > 0 {
					r.historyIdx--
				}
				if r.historyIdx >= 0 && r.historyIdx < len(history) {
					r.textInput.SetValue(history[r.historyIdx])
					r.textInput.CursorEnd()
				}
			}
			return r, nil

		case tea.KeyDown:
			// Navigate history forwards
			history, err := r.history.GetAll()
			if err == nil && len(history) > 0 {
				if r.historyIdx < len(history)-1 {
					r.historyIdx++
					r.textInput.SetValue(history[r.historyIdx])
					r.textInput.CursorEnd()
				} else {
					r.historyIdx = -1
					r.textInput.SetValue("")
				}
			}
			return r, nil

		case tea.KeyTab:
			// Tab completion - trigger autocomplete
			return r.triggerAutocomplete(), nil
		default:
			// Do Nothing
		}
	}

	// Update text input
	r.textInput, cmd = r.textInput.Update(msg)
	return r, cmd
}

func (r *Repl) View() string {
	var view strings.Builder

	// Show current input
	view.WriteString(r.textInput.View())

	// Show autocomplete suggestions if active
	if r.isSuggesting {
		view.WriteString("\n")
		view.WriteString(renderSuggestions(r.suggestions, r.selectedSuggestion))
	}

	// Help text
	view.WriteString("\n\n")
	helpText := "↑/↓: history • tab: complete • 'exit': quit"
	if r.isSuggesting {
		helpText = "↑/↓: navigate • enter: select • esc: cancel"
	}
	view.WriteString(styleHelp.Render(helpText))

	return view.String()
}

// Autocomplete methods
func (r *Repl) triggerAutocomplete() *Repl {
	if r.suggestionProvider == nil {
		// Fall back to legacy tab completion
		current := r.textInput.Value()
		completed := r.handler.Tab(current)
		if completed != current {
			r.textInput.SetValue(completed)
			r.textInput.CursorEnd()
		}
		return r
	}

	input := r.textInput.Value()
	cursorPos := r.textInput.Position()

	suggestions, textToReplace, err := r.suggestionProvider.GetSuggestions(input, cursorPos)
	if err != nil || len(suggestions) == 0 {
		return r
	}

	// If only one suggestion, apply it immediately
	if len(suggestions) == 1 {
		newInput, newCursorPos := applySuggestion(input, textToReplace, suggestions[0].Value, cursorPos)
		r.textInput.SetValue(newInput)
		r.textInput.SetCursor(newCursorPos)
		return r
	}

	// Multiple suggestions - show suggestion box
	r.dirtyInput = input
	r.suggestions = suggestions
	r.textToReplace = textToReplace
	r.selectedSuggestion = 0
	r.isSuggesting = true

	// Apply first suggestion as preview
	newInput, newCursorPos := applySuggestion(input, textToReplace, suggestions[0].Value, cursorPos)
	r.textInput.SetValue(newInput)
	r.textInput.SetCursor(newCursorPos)

	return r
}

func (r *Repl) navigateSuggestions(direction int) *Repl {
	if !r.isSuggesting || len(r.suggestions) == 0 {
		return r
	}

	r.selectedSuggestion += direction
	if r.selectedSuggestion < 0 {
		r.selectedSuggestion = len(r.suggestions) - 1
	} else if r.selectedSuggestion >= len(r.suggestions) {
		r.selectedSuggestion = 0
	}

	// Apply the selected suggestion as preview
	cursorPos := r.textInput.Position()
	newInput, newCursorPos := applySuggestion(r.dirtyInput, r.textToReplace, r.suggestions[r.selectedSuggestion].Value, cursorPos)
	r.textInput.SetValue(newInput)
	r.textInput.SetCursor(newCursorPos)

	return r
}

func (r *Repl) selectSuggestion() *Repl {
	r.isSuggesting = false
	r.suggestions = nil
	r.selectedSuggestion = -1
	r.dirtyInput = ""
	r.textToReplace = ""
	r.textInput.CursorEnd()
	return r
}

func (r *Repl) exitSuggestions() *Repl {
	if r.isSuggesting {
		// Restore original input
		r.textInput.SetValue(r.dirtyInput)
		r.textInput.CursorEnd()
	}

	r.isSuggesting = false
	r.suggestions = nil
	r.selectedSuggestion = -1
	r.dirtyInput = ""
	r.textToReplace = ""
	return r
}

// Loop starts the REPL - this is the main entry point
func (r *Repl) Run(opts ...tea.ProgramOption) error {
	p := tea.NewProgram(r, opts...)
	_, err := p.Run()
	return err
}
