package repllib

import (
	"testing"
)

func TestTrimString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than maxLen",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "string equal to maxLen",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "string longer than maxLen",
			input:    "HelloWorld",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "maxLen is zero",
			input:    "Hello",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "single character",
			input:    "H",
			maxLen:   1,
			expected: "H",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("trimString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestAutocompleteMatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		word     string
		expected bool
	}{
		{
			name:     "case insensitive prefix match",
			input:    "HelloWorld",
			word:     "hello",
			expected: true,
		},
		{
			name:     "no prefix match",
			input:    "HelloWorld",
			word:     "world",
			expected: false,
		},
		{
			name:     "exact match (should return false)",
			input:    "HelloWorld",
			word:     "HelloWorld",
			expected: false,
		},
		{
			name:     "exact match case insensitive (should return false)",
			input:    "HelloWorld",
			word:     "helloworld",
			expected: false,
		},
		{
			name:     "partial prefix match",
			input:    "HelloWorld",
			word:     "Hell",
			expected: true,
		},
		{
			name:     "empty word",
			input:    "HelloWorld",
			word:     "",
			expected: true,
		},
		{
			name:     "empty input",
			input:    "",
			word:     "hello",
			expected: false,
		},
		{
			name:     "both empty",
			input:    "",
			word:     "",
			expected: false,
		},
		{
			name:     "word longer than input",
			input:    "Hi",
			word:     "Hello",
			expected: false,
		},
		{
			name:     "case variations",
			input:    "HELLOWORLD",
			word:     "hello",
			expected: true,
		},
		{
			name:     "mixed case exact match",
			input:    "Hello",
			word:     "HELLO",
			expected: false,
		},
		{
			name:     "unicode characters",
			input:    "世界Hello",
			word:     "世界",
			expected: true,
		},
		{
			name:     "numbers and letters",
			input:    "Test123",
			word:     "test1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := autocompleteMatch(tt.input, tt.word)
			if result != tt.expected {
				t.Errorf("autocompleteMatch(%q, %q) = %v, want %v", tt.input, tt.word, result, tt.expected)
			}
		})
	}
}

func TestFindWordBounds(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		cursorPos     int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "cursor in middle of word",
			input:         "hello world",
			cursorPos:     3,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "cursor at start of word",
			input:         "hello world",
			cursorPos:     0,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "cursor at end of word",
			input:         "hello world",
			cursorPos:     5,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "cursor on space",
			input:         "hello world",
			cursorPos:     5,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "cursor after space",
			input:         "hello world",
			cursorPos:     6,
			expectedStart: 6,
			expectedEnd:   11,
		},
		{
			name:          "cursor at end of string",
			input:         "hello",
			cursorPos:     5,
			expectedStart: 0,
			expectedEnd:   5,
		},
		{
			name:          "empty string",
			input:         "",
			cursorPos:     0,
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "single character",
			input:         "a",
			cursorPos:     0,
			expectedStart: 0,
			expectedEnd:   1,
		},
		{
			name:          "word with underscores",
			input:         "hello_world_test",
			cursorPos:     7,
			expectedStart: 0,
			expectedEnd:   16,
		},
		{
			name:          "word with numbers",
			input:         "test123abc",
			cursorPos:     5,
			expectedStart: 0,
			expectedEnd:   10,
		},
		{
			name:          "multiple spaces",
			input:         "hello   world",
			cursorPos:     7,
			expectedStart: 7,
			expectedEnd:   7,
		},
		{
			name:          "special characters",
			input:         "hello-world@test",
			cursorPos:     3,
			expectedStart: 0,
			expectedEnd:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := findWordBounds(tt.input, tt.cursorPos)
			if start != tt.expectedStart || end != tt.expectedEnd {
				t.Errorf("findWordBounds(%q, %d) = (%d, %d), want (%d, %d)",
					tt.input, tt.cursorPos, start, end, tt.expectedStart, tt.expectedEnd)
			}
		})
	}
}
