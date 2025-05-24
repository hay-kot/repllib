package repllib

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styling
var (
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render // Red

	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Gray
)

// EvalFunc is the required evaluation function
type EvalFunc func(buffer string) string

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
	output      []string
	historyIdx  int
	promptCount int // Track the prompt number like IPython
	quitting    bool
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
	return textinput.Blink
}

func (r *Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			r.quitting = true
			return r, tea.Quit

		case tea.KeyEnter:
			input := strings.TrimSpace(r.textInput.Value())
			if input != "" {
				result := r.handler.Eval(input)

				// Ensure extra empty line
				if result != "" && !strings.HasSuffix(result, "\n") {
					result += "\n"
				}

				// Add to output with current prompt
				r.output = append(r.output, r.handler.Prompt(r.promptCount)+input)
				if result != "" {
					r.output = append(r.output, result)
				}

				// Add to history
				_ = r.history.Push(input) // TO-DO: log error

				// Increment prompt counter and reset input
				r.promptCount++
				r.textInput.SetValue("")
				r.textInput.Prompt = r.handler.Prompt(r.promptCount)
				r.historyIdx = -1
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
			// Tab completion
			current := r.textInput.Value()
			completed := r.handler.Tab(current)
			if completed != current {
				r.textInput.SetValue(completed)
				r.textInput.CursorEnd()
			}
			return r, nil
		default:
			// Do Nothing
		}
	}

	// Update text input
	r.textInput, cmd = r.textInput.Update(msg)
	return r, cmd
}

func (r *Repl) View() string {
	if r.quitting {
		return "Goodbye!\n"
	}

	var view strings.Builder

	// Show recent output (last 20 lines to avoid infinite scroll)
	outputStart := 0
	if len(r.output) > 20 {
		outputStart = len(r.output) - 20
	}

	for i := outputStart; i < len(r.output); i++ {
		view.WriteString(r.output[i])
		view.WriteString("\n")
	}

	// Show current input
	view.WriteString(r.textInput.View())

	// Help text
	view.WriteString("\n\n")
	view.WriteString(helpStyle.Render("↑/↓: history • tab: complete • ctrl+c/esc: quit"))

	return view.String()
}

// Loop starts the REPL - this is the main entry point
func (r *Repl) Run(opts ...tea.ProgramOption) error {
	p := tea.NewProgram(r, opts...)
	_, err := p.Run()
	return err
}
