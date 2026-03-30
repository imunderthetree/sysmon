package gui

import (
	"fmt"
	"image/color"
	"sysmon/internal"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// OverviewTab represents the overview/dashboard view
type OverviewTab struct {
	appState *AppState

	// UI components
	cpuLabel       *widget.Label
	memLabel       *widget.Label
	diskLabel      *widget.Label
	networkLabel   *widget.Label
	uptimeLabel    *widget.Label
	processesLabel *widget.Label

	cpuBar       *widget.ProgressBar
	memBar       *widget.ProgressBar
	hostLabel    *widget.Label
}

// NewOverviewTab creates a new overview tab
func (s *AppState) newOverviewTab() fyne.CanvasObject {
	tab := &OverviewTab{appState: s}

	// Initialize labels
	tab.cpuLabel = widget.NewLabel("CPU: ---%")
	tab.memLabel = widget.NewLabel("Memory: ---%")
	tab.diskLabel = widget.NewLabel("Disk: --")
	tab.networkLabel = widget.NewLabel("Network: -- ▲ -- ▼")
	tab.uptimeLabel = widget.NewLabel("Uptime: --")
	tab.processesLabel = widget.NewLabel("Processes: --")
	tab.hostLabel = widget.NewLabel("Host: --")

	// Initialize progress bars
	tab.cpuBar = widget.NewProgressBar()
	tab.memBar = widget.NewProgressBar()

	// Title
	title := canvas.NewText("System Overview", color.White)
	title.TextSize = 24

	// System Info Section
	sysInfoBox := container.NewVBox(
		tab.hostLabel,
		tab.uptimeLabel,
		tab.processesLabel,
	)
	sysInfoBorder := container.NewBorder(nil, nil, nil, nil, sysInfoBox)

	// CPU Section
	cpuSection := container.NewVBox(
		tab.cpuLabel,
		tab.cpuBar,
	)

	// Memory Section
	memSection := container.NewVBox(
		tab.memLabel,
		tab.memBar,
	)

	// Network Section
	networkSection := container.NewVBox(
		tab.networkLabel,
	)

	// Disk Section
	diskSection := container.NewVBox(
		tab.diskLabel,
	)

	// Charts section
	cpuChartLabel := widget.NewLabel("CPU usage (last 60s):")
	cpuChart := widget.NewLabel("Loading...")

	memChartLabel := widget.NewLabel("Memory usage (last 60s):")
	memChart := widget.NewLabel("Loading...")

	chartsSection := container.NewVBox(
		widget.NewSeparator(),
		cpuChartLabel,
		cpuChart,
		widget.NewSeparator(),
		memChartLabel,
		memChart,
	)

	// Main grid layout
	mainContent := container.NewVBox(
		title,
		sysInfoBorder,
		widget.NewSeparator(),
		container.NewHBox(
			container.NewVBox(cpuSection, memSection),
			container.NewVBox(networkSection, diskSection),
		),
		chartsSection,
	)

	// Wrap in scroll for larger content
	scroll := container.NewScroll(mainContent)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	// Start update goroutine
	go tab.updateLoop()

	return scroll
}

// updateLoop periodically updates the overview display
func (tab *OverviewTab) updateLoop() {
	for {
		select {
		case <-tab.appState.stopChan:
			return
		default:
			tab.updateDisplay()
			tab.appState.mutex.RLock()
			refreshRate := tab.appState.refreshRate
			tab.appState.mutex.RUnlock()

			// Use a timer so updates happen between data collections
			<-getTimer(refreshRate / 2).C
		}
	}
}

// updateDisplay updates all labels with current data
func (tab *OverviewTab) updateDisplay() {
	tab.appState.mutex.RLock()
	defer tab.appState.mutex.RUnlock()

	if tab.appState.systemStats == nil {
		return
	}

	stats := tab.appState.systemStats

	// Update CPU
	cpuPercent := stats.CPU.Usage
	tab.cpuLabel.SetText(fmt.Sprintf("CPU: %.1f%% (%d cores)", cpuPercent, stats.CPU.Cores))
	tab.cpuBar.SetValue(cpuPercent / 100)

	// Update Memory
	memPercent := stats.Memory.UsedPercent
	tab.memLabel.SetText(fmt.Sprintf("Memory: %.1f%% (%s / %s)",
		memPercent,
		internal.FormatBytes(stats.Memory.Used),
		internal.FormatBytes(stats.Memory.Total)))
	tab.memBar.SetValue(memPercent / 100)

	// Update Host
	tab.hostLabel.SetText(fmt.Sprintf("Host: %s (%s)", stats.Host.Hostname, stats.Host.OS))

	// Update Uptime
	tab.uptimeLabel.SetText(fmt.Sprintf("Uptime: %s", internal.FormatUptime(stats.Host.Uptime)))

	// Update Processes
	if tab.appState.processStats != nil {
		tab.processesLabel.SetText(fmt.Sprintf("Processes: %d running, %d total",
			tab.appState.processStats.RunningProcs,
			tab.appState.processStats.TotalProcesses))
	}

	// Update Network
	if tab.appState.networkStats != nil {
		tab.networkLabel.SetText(fmt.Sprintf("Network: ▲ %s ▼ %s",
			internal.FormatNetworkBytes(tab.appState.networkStats.TotalSent),
			internal.FormatNetworkBytes(tab.appState.networkStats.TotalRecv)))
	}

	// Update Disk
	if len(stats.Disk) > 0 {
		primaryDisk := stats.Disk[0]
		tab.diskLabel.SetText(fmt.Sprintf("Disk: %.1f%% used (%s / %s)",
			primaryDisk.UsedPercent,
			internal.FormatBytes(primaryDisk.Used),
			internal.FormatBytes(primaryDisk.Total)))
	}
}

// Helper function for timer
func getTimer(d time.Duration) *time.Timer {
	return time.NewTimer(d)
}
