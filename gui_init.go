//go:build !tui
// +build !tui

package main

import (
	"sysmon/gui"
)

func initGUI() {
	guiApp := gui.NewApp()
	guiApp.Run()
}
