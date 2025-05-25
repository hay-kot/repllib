// builder.go - Enhanced version
package repllib

import (
	"strings"
)

// Builder struct
type ReplBuilder struct {
	evalFunc           EvalFunc
	promptFunc         PromptFunc
	tabFunc            TabFunc
	history            ReplHistory
	suggestionProvider Suggester
}

// New creates a new REPL builder - requires an evaluation function
func New(evalFunc EvalFunc) *ReplBuilder {
	history := &memoryHistory{}
	return &ReplBuilder{
		evalFunc:   evalFunc,
		history:    history,
		promptFunc: PromptIPython(),
		tabFunc: func(buffer string) string {
			// Simple tab completion fallback
			commands, err := history.GetAll()
			if err != nil {
				return buffer
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

// Build creates the final Repl
func (b *ReplBuilder) Build() *Repl {
	handler := &defaultHandler{
		evalFunc:   b.evalFunc,
		promptFunc: b.promptFunc,
		tabFunc:    b.tabFunc,
	}

	return &Repl{
		handler:            handler,
		history:            b.history,
		suggestionProvider: b.suggestionProvider,
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

// New autocomplete methods
func (b *ReplBuilder) WithSuggestions(provider Suggester) *ReplBuilder {
	b.suggestionProvider = provider
	return b
}
