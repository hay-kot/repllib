# repllib

A Bubble Tea library for quickly building REPL-style interactive CLIs in Go. Create powerful command-line interfaces with history, tab completion, and custom evaluation logic.

[![Go Reference](https://pkg.go.dev/badge/github.com/hay-kot/repllib.svg)](https://pkg.go.dev/github.com/hay-kot/repllib)

## Install

```bash
go get github.com/hay-kot/repllib
```

```go
import "github.com/hay-kot/repllib"
```

## Features

- **Interactive REPL Interface** - Built on Bubble Tea for smooth terminal interactions
- **Command History** - Navigate through previous commands with ↑/↓ arrow keys
- **Tab Completion** - Extensible tab completion system
- **Custom Evaluation** - Plug in your own command evaluation logic
- **Error Handling** - Built-in error styling and display
- **Flexible Builder Pattern** - Customize prompts, completion, and history storage
- **Memory History** - Built-in in-memory history with option to provide custom storage

## Quick Start

Here's a simple echo server that demonstrates the basic functionality:

```go
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
	repl := repllib.NewRepl(func(input string) string {
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

	if err := repl.Loop(); err != nil {
		log.Fatal(err)
	}
}
```
