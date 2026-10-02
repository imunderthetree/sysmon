//go:build !tui
// +build !tui

package main

import (
	"flag"
)

func main() {
	// Parse command line flags
	guiMode := flag.Bool("gui", false, "Run in Web GUI mode (default)")
	tuiMode := flag.Bool("tui", false, "Run in Terminal UI mode")
	port := flag.Int("port", 8080, "Port for Web GUI server")
	noBrowser := flag.Bool("no-browser", false, "Do not open browser automatically")
	flag.Parse()

	// Determine which mode to run
	if *tuiMode && !*guiMode {
		initTUI()
		return
	}

	// Default to Web GUI mode
	initGUI(*port, !*noBrowser)
}
