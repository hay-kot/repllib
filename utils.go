package repllib

import (
	"strings"
	"unicode"
)

// trimString trims the string to a maximum length of maxLen.
//
// Example:
//
// //   trimString("HelloWorld", 5) // "Hello"
func trimString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// autocompleteMatch checks if the name starts with the given word, case-insensitively.
//
// Examples:
//
//	//   autocompleteMatch("HelloWorld", "hello") // true
//	//   autocompleteMatch("HelloWorld", "world") // false
//	//   autocompleteMatch("HelloWorld", "HelloWorld") // false
func autocompleteMatch(name, word string) bool {
	lowerName := strings.ToLower(name)
	lowerWord := strings.ToLower(word)

	return strings.HasPrefix(lowerName, lowerWord) && !strings.EqualFold(lowerName, lowerWord)
}

// atMostN returns a slice containing at most n elements from vals.
func atMostN[T any](vals []T, n int) []T {
	if len(vals) <= n {
		return vals
	}
	return vals[:n]
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
