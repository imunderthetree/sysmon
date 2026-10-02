//go:build !tui
// +build !tui

package main

import (
	"sysmon/gui"
)

func initGUI(port int, openBrowser bool) {
	guiApp := gui.NewApp()
	if port > 0 {
		guiApp.Port = port
	}
	guiApp.OpenBrowserUI = openBrowser
	guiApp.Run()
}
