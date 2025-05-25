package repllib

import "fmt"

// Core interfaces
type Handler interface {
	Prompt(count int) string
	Eval(buffer string) string
	Tab(buffer string) string
}

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
