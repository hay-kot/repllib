package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hay-kot/repllib"
)

func main() {
	// Create a simple echo REPL
	repl := repllib.New(func(input string) string {
		trimmed := strings.TrimSpace(input)

		// Handle empty input
		if trimmed == "" {
			return ""
		}

		// Handle error demonstration
		if strings.HasPrefix(strings.ToLower(trimmed), "error") {
			return repllib.Error(fmt.Sprintf("Simulated error: %s", trimmed[5:]))
		}

		// Handle special commands
		switch strings.ToLower(trimmed) {
		case "hello":
			return "Hello, World!"
		case "time":
			return fmt.Sprintf("Current time: %s", time.Now().Format("15:04:05"))
		case "help":
			return `Available commands:
  hello - Say hello
  time  - Show current time
  error <msg> - Show error message
  quit  - Exit (or press Ctrl+C)`
		case "quit":
			return "Goodbye!"
		}

		// Default: echo the input
		return fmt.Sprintf("Echo: %s", trimmed)
	}).Build()

	// Start the REPL
	fmt.Println("Welcome to the Echo REPL!")
	fmt.Println("Type 'help' for available commands, or Ctrl+C to quit.")
	fmt.Println()

	if err := repl.Run(); err != nil {
		log.Fatal(err)
	}
}
