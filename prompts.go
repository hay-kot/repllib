package repllib

import "fmt"

// PromptIPython is an ipython style prompt for REPLs.
//
// Example:
//
// 	// In [1]: ...
// 	// In [2]: ...
func PromptIPython() PromptFunc {
	return func(count int) string {
		return stylePrompt.Render(fmt.Sprintf("In [%d]: ", count))
	}
}

// PromptPython is a simple carrot style prompt for REPLs.
//
// Example:
//
// 	// >
func PromptCarrot() PromptFunc {
	return func(count int) string {
		return stylePrompt.Render("> ")
	}
}
