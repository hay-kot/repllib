package repllib

import (
	"fmt"
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

// Core interfaces
type Handler interface {
	Prompt(count int) string
	Eval(buffer string) string
	Tab(buffer string) string
}

// EvalResult allows handlers to return both output and errors
type EvalResult struct {
	Output string
	Error  error
}

// EvalResultHandler is an alternative handler interface that supports error reporting
type EvalResultHandler interface {
	Prompt(count int) string
	EvalWithResult(buffer string) EvalResult
	Tab(buffer string) string
}

type ReplHistory interface {
	Push(buffer string) error
	GetAll() ([]string, error)
}

// EvalFunc is the required evaluation function
type EvalFunc func(buffer string) string

// EvalResultFunc is an alternative evaluation function that can return errors
type EvalResultFunc func(buffer string) EvalResult

// Optional function types for builder
type (
	PromptFunc func(count int) string
	TabFunc    func(buffer string) string
)

// Default implementations
type defaultHandler struct {
	evalFunc   EvalFunc
	promptFunc PromptFunc
	tabFunc    TabFunc
}

func (h *defaultHandler) Prompt(count int) string {
	if h.promptFunc != nil {
		return h.promptFunc(count)
	}
	return fmt.Sprintf("In [%d]: ", count) // IPython-style default prompt
}

func (h *defaultHandler) Eval(buffer string) string {
	return h.evalFunc(buffer)
}

func (h *defaultHandler) Tab(buffer string) string {
	if h.tabFunc != nil {
		return h.tabFunc(buffer)
	}
	return buffer // default: no completion
}

// Simple in-memory history implementation
type memoryHistory struct {
	commands []string
}

func (h *memoryHistory) Push(buffer string) error {
	h.commands = append(h.commands, buffer)
	return nil
}

func (h *memoryHistory) GetAll() ([]string, error) {
	return h.commands, nil
}

// Builder struct
type ReplBuilder struct {
	evalFunc   EvalFunc
	promptFunc PromptFunc
	tabFunc    TabFunc
	history    ReplHistory
}

// NewRepl creates a new REPL builder - requires an evaluation function
func NewRepl(evalFunc EvalFunc) *ReplBuilder {
	history := &memoryHistory{}
	return &ReplBuilder{
		evalFunc: evalFunc,
		history:  history,
		promptFunc: func(count int) string {
			return promptStyle.Render(fmt.Sprintf("In [%d]: ", count))
		},
		tabFunc: func(buffer string) string {
			// Simple tab completion
			commands, err := history.GetAll()
			if err != nil {
				return ""
			}

			for _, cmd := range commands {
				if strings.HasPrefix(cmd, buffer) {
					return cmd
				}
			}
			return buffer
		},
	}
}

// Builder methods
func (b *ReplBuilder) WithPrompt(promptFunc PromptFunc) *ReplBuilder {
	b.promptFunc = promptFunc
	return b
}

func (b *ReplBuilder) WithTab(tabFunc TabFunc) *ReplBuilder {
	b.tabFunc = tabFunc
	return b
}

func (b *ReplBuilder) WithHistory(history ReplHistory) *ReplBuilder {
	b.history = history
	return b
}

// Build creates the final Repl
func (b *ReplBuilder) Build() *Repl {
	handler := &defaultHandler{
		evalFunc:   b.evalFunc,
		promptFunc: b.promptFunc,
		tabFunc:    b.tabFunc,
	}

	return &Repl{
		handler: handler,
		history: b.history,
	}
}

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
func (r *Repl) Loop() error {
	p := tea.NewProgram(r)
	_, err := p.Run()
	return err
}
