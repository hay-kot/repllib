package repllib

import (
	"fmt"
	"strings"
)

// Builder struct
type ReplBuilder struct {
	evalFunc   EvalFunc
	promptFunc PromptFunc
	tabFunc    TabFunc
	history    ReplHistory
	mw         []EvalMiddleware
}

// New creates a new REPL builder - requires an evaluation function
func New(evalFunc EvalFunc) *ReplBuilder {
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
		mw: []EvalMiddleware{
			WithExitMiddleware(),
		},
	}
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
		mw:      b.mw,
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

func (b *ReplBuilder) WithMiddleware(mw ...EvalMiddleware) *ReplBuilder {
	b.mw = append(b.mw, mw...)
	return b
}
