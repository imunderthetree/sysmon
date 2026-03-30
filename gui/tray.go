package gui

import (
	"fmt"

	"github.com/getlantern/systray"
)

// InitSystemTray initializes the system tray functionality
func (s *AppState) InitSystemTray() {
	go func() {
		systray.Run(s.onReady, s.onExit)
	}()
}

// onReady is called when the system tray is ready
func (s *AppState) onReady() {
	// Create menu items
	showItem := systray.AddMenuItem("Show", "Show the System Monitor window")
	hideItem := systray.AddMenuItem("Hide", "Hide the System Monitor window")
	systray.AddSeparator()
	statsItem := systray.AddMenuItem("Stats", "View current system stats")
	systray.AddSeparator()
	exitItem := systray.AddMenuItem("Exit", "Quit System Monitor")

	// Handle menu clicks
	for {
		select {
		case <-showItem.ClickedCh:
			s.mainWindow.Show()
			s.mainWindow.RequestFocus()
		case <-hideItem.ClickedCh:
			s.mainWindow.Hide()
		case <-statsItem.ClickedCh:
			s.displayQuickStats()
		case <-exitItem.ClickedCh:
			s.stopChan <- true
			return
		}
	}
}

// onExit is called when the system tray is closed
func (s *AppState) onExit() {
	s.stopChan <- true
}

// displayQuickStats displays quick stats (placeholder)
func (s *AppState) displayQuickStats() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.systemStats != nil {
		title := fmt.Sprintf("CPU: %.1f%% | Memory: %.1f%%", s.systemStats.CPU.Usage, s.systemStats.Memory.UsedPercent)
		systray.SetTitle(title)
	}
}
