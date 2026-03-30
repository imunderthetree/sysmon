//go:build !tui
// +build !tui

package main

import (
	"flag"
)

func main() {
	// Parse command line flags
	guiMode := flag.Bool("gui", false, "Run in GUI mode (using Fyne)")
	tuiMode := flag.Bool("tui", false, "Run in Terminal UI mode")
	flag.Parse()

	// Determine which mode to run
	// Default to GUI mode if no mode specified
	if *guiMode || (!*tuiMode && !*guiMode) {
		initGUI()
		return
	}

	// Run TUI mode
	initTUI()
}
