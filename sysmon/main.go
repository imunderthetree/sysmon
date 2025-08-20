// main.go - Enhanced System Monitor v1.0
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"sysmon/internal"
	"time"
)

// ViewType represents different monitoring views
type ViewType int

const (
	ViewOverview ViewType = iota
	ViewProcesses
	ViewNetwork
	ViewDisks
	ViewSystem
)

// Color constants for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// Application state
type App struct {
	currentView   ViewType
	refreshRate   time.Duration
	paused        bool
	logToFile     bool
	logFile       *os.File
	showHelp      bool
	compactMode   bool
	colorEnabled  bool
	exitRequested bool
}

func main() {
	app := &App{
		currentView:  ViewOverview,
		refreshRate:  3 * time.Second,
		paused:       false,
		logToFile:    false,
		showHelp:     false,
		compactMode:  false,
		colorEnabled: true,
	}

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start keyboard input handler
	inputChan := make(chan rune)
	go handleKeyboardInput(inputChan)

	// Initial display
	app.clearScreen()
	app.displayInterface()

	// Main loop
	ticker := time.NewTicker(app.refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-signalChan:
			app.cleanup()
			return
		case key := <-inputChan:
			if app.handleKeyPress(key) {
				app.cleanup()
				return
			}
		case <-ticker.C:
			if !app.paused && !app.showHelp {
				app.displayInterface()
			}
		}
	}
}

func (app *App) handleKeyPress(key rune) bool {
	switch key {
	case 'q', 'Q':
		return true // Exit
	case 'h', 'H', '?':
		app.showHelp = !app.showHelp
		app.displayInterface()
	case '1':
		app.currentView = ViewOverview
		app.displayInterface()
	case '2':
		app.currentView = ViewProcesses
		app.displayInterface()
	case '3':
		app.currentView = ViewNetwork
		app.displayInterface()
	case '4':
		app.currentView = ViewDisks
		app.displayInterface()
	case '5':
		app.currentView = ViewSystem
		app.displayInterface()
	case 'p', 'P':
		app.paused = !app.paused
		app.displayInterface()
	case 'c', 'C':
		app.compactMode = !app.compactMode
		app.displayInterface()
	case 'l', 'L':
		app.toggleLogging()
	case 'e', 'E':
		app.exportStats()
	case 'r', 'R':
		app.displayInterface() // Refresh
	case '+':
		if app.refreshRate > time.Second {
			app.refreshRate -= time.Second
			ticker := time.NewTicker(app.refreshRate)
			defer ticker.Stop()
		}
	case '-':
		if app.refreshRate < 10*time.Second {
			app.refreshRate += time.Second
			ticker := time.NewTicker(app.refreshRate)
			defer ticker.Stop()
		}
	}
	return false
}

func (app *App) displayInterface() {
	app.clearScreen()

	if app.showHelp {
		app.displayHelp()
		return
	}

	app.displayHeader()

	switch app.currentView {
	case ViewOverview:
		app.displayOverviewView()
	case ViewProcesses:
		app.displayProcessesView()
	case ViewNetwork:
		app.displayNetworkView()
	case ViewDisks:
		app.displayDisksView()
	case ViewSystem:
		app.displaySystemView()
	}

	app.displayFooter()
}

func (app *App) displayHeader() {
	viewNames := []string{"Overview", "Processes", "Network", "Disks", "System"}
	statusColor := ColorGreen
	if app.paused {
		statusColor = ColorYellow
	}

	// Top border
	fmt.Print(app.colorize("┌", ColorCyan))
	fmt.Print(app.colorize(strings.Repeat("─", 78), ColorCyan))
	fmt.Print(app.colorize("┐", ColorCyan))
	fmt.Println()

	// Title and status
	title := fmt.Sprintf("System Monitor v1.0 - %s View", viewNames[app.currentView])
	status := "RUNNING"
	if app.paused {
		status = "PAUSED"
	}

	fmt.Printf("│ %s%s%s%s │\n",
		app.colorize(title, ColorBold+ColorWhite),
		strings.Repeat(" ", 78-len(title)-len(status)-3),
		app.colorize(status, ColorBold+statusColor),
		app.colorize("", ColorReset))

	// Time and refresh info
	timeStr := time.Now().Format("15:04:05")
	refreshStr := fmt.Sprintf("Refresh: %v", app.refreshRate)
	fmt.Printf("│ %s%s%s │\n",
		app.colorize(timeStr, ColorCyan),
		strings.Repeat(" ", 78-len(timeStr)-len(refreshStr)),
		app.colorize(refreshStr, ColorDim))

	// Navigation tabs
	fmt.Print(app.colorize("├", ColorCyan))
	fmt.Print(app.colorize(strings.Repeat("─", 78), ColorCyan))
	fmt.Print(app.colorize("┤", ColorCyan))
	fmt.Println()

	tabStr := ""
	for i, name := range viewNames {
		prefix := fmt.Sprintf("[%d]", i+1)
		if ViewType(i) == app.currentView {
			tabStr += app.colorize(fmt.Sprintf("%s%s ", prefix, name), ColorBold+ColorYellow)
		} else {
			tabStr += app.colorize(fmt.Sprintf("%s%s ", prefix, name), ColorDim)
		}
	}

	fmt.Printf("│ %s%s │\n", tabStr, strings.Repeat(" ", 78-len(stripColors(tabStr))))

	// Bottom border of header
	fmt.Print(app.colorize("└", ColorCyan))
	fmt.Print(app.colorize(strings.Repeat("─", 78), ColorCyan))
	fmt.Print(app.colorize("┘", ColorCyan))
	fmt.Println()
	fmt.Println()
}

func (app *App) displayOverviewView() {
	stats, err := internal.GetSystemStats()
	if err != nil {
		fmt.Printf(app.colorize("Error getting system stats: %v\n", ColorRed), err)
		return
	}

	procStats, _ := internal.GetProcessStats()
	netStats, _ := internal.GetNetworkStats()

	app.displaySystemOverview(stats)

	if procStats != nil {
		app.displayProcessSummary(procStats)
	}

	if netStats != nil {
		app.displayNetworkSummary(netStats)
	}

	// Log stats if enabled
	if app.logToFile {
		app.logStats(stats, procStats, netStats)
	}
}

func (app *App) displaySystemOverview(stats *internal.SystemStats) {
	// System Info
	fmt.Printf("%s🖥️  System Information%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
	fmt.Printf("   Hostname: %s | OS: %s | Uptime: %s\n\n",
		app.colorize(stats.Host.Hostname, ColorCyan),
		app.colorize(stats.Host.OS, ColorCyan),
		app.colorize(internal.FormatUptime(stats.Host.Uptime), ColorGreen))

	// CPU
	cpuColor := app.getUsageColor(stats.CPU.Usage)
	fmt.Printf("%s🔧 CPU Usage: %.1f%%%s %s\n",
		app.colorize("", ColorBold+ColorBlue),
		stats.CPU.Usage,
		app.colorize("", ColorReset),
		app.getProgressBar(stats.CPU.Usage, 40, cpuColor))

	if !app.compactMode {
		fmt.Printf("   Cores: %d | Model: %s\n\n",
			stats.CPU.Cores,
			app.colorize(app.truncateString(stats.CPU.ModelName, 50), ColorDim))
	}

	// Memory
	memColor := app.getUsageColor(stats.Memory.UsedPercent)
	fmt.Printf("%s💾 Memory: %.1f%%%s %s\n",
		app.colorize("", ColorBold+ColorBlue),
		stats.Memory.UsedPercent,
		app.colorize("", ColorReset),
		app.getProgressBar(stats.Memory.UsedPercent, 40, memColor))

	if !app.compactMode {
		fmt.Printf("   Used: %s / %s | Free: %s\n\n",
			app.colorize(internal.FormatBytes(stats.Memory.Used), ColorYellow),
			app.colorize(internal.FormatBytes(stats.Memory.Total), ColorCyan),
			app.colorize(internal.FormatBytes(stats.Memory.Available), ColorGreen))
	}

	// Disk Usage Summary
	if !app.compactMode {
		fmt.Printf("%s💽 Disk Usage:%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
		for i, disk := range stats.Disk {
			if i >= 3 { // Show max 3 disks in overview
				break
			}
			diskColor := app.getUsageColor(disk.UsedPercent)
			device := app.truncateString(filepath.Base(disk.Device), 15)
			fmt.Printf("   %-15s %6.1f%% %s %s / %s\n",
				app.colorize(device, ColorCyan),
				disk.UsedPercent,
				app.getProgressBar(disk.UsedPercent, 20, diskColor),
				app.colorize(internal.FormatBytes(disk.Used), ColorYellow),
				app.colorize(internal.FormatBytes(disk.Total), ColorDim))
		}
		fmt.Println()
	}
}

func (app *App) displayProcessSummary(stats *internal.ProcessStats) {
	fmt.Printf("%s📄 Process Summary%s\n", app.colorize("", ColorBold+ColorPurple), app.colorize("", ColorReset))
	fmt.Printf("   Total: %s | Running: %s | Sleeping: %s\n\n",
		app.colorize(fmt.Sprintf("%d", stats.TotalProcesses), ColorCyan),
		app.colorize(fmt.Sprintf("%d", stats.RunningProcs), ColorGreen),
		app.colorize(fmt.Sprintf("%d", stats.SleepingProcs), ColorYellow))

	if !app.compactMode {
		fmt.Printf("%s🔥 Top CPU Processes:%s\n", app.colorize("", ColorBold+ColorRed), app.colorize("", ColorReset))
		for i, proc := range stats.TopCPU {
			if i >= 3 || proc.CPUPercent < 0.1 {
				break
			}
			fmt.Printf("   %-20s %6.1f%% %s\n",
				app.colorize(app.truncateString(proc.Name, 20), ColorCyan),
				proc.CPUPercent,
				app.colorize(app.formatMB(proc.MemoryMB), ColorDim))
		}
		fmt.Println()
	}
}

func (app *App) displayNetworkSummary(stats *internal.NetworkStats) {
	fmt.Printf("%s🌐 Network Summary%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("   Active Interfaces: %s | Connections: %s\n",
		app.colorize(fmt.Sprintf("%d", stats.ActiveIfaces), ColorCyan),
		app.colorize(fmt.Sprintf("%d", stats.Connections), ColorCyan))
	fmt.Printf("   Total Traffic: ↑%s ↓%s\n\n",
		app.colorize(internal.FormatNetworkBytes(stats.TotalSent), ColorRed),
		app.colorize(internal.FormatNetworkBytes(stats.TotalRecv), ColorGreen))
}

func (app *App) displayProcessesView() {
	procStats, err := internal.GetProcessStats()
	if err != nil {
		fmt.Printf(app.colorize("Error getting process stats: %v\n", ColorRed), err)
		return
	}

	// Process counts
	fmt.Printf("%s📊 Process Statistics%s\n", app.colorize("", ColorBold+ColorPurple), app.colorize("", ColorReset))
	fmt.Printf("Total: %s | Running: %s | Sleeping: %s\n\n",
		app.colorize(fmt.Sprintf("%d", procStats.TotalProcesses), ColorCyan),
		app.colorize(fmt.Sprintf("%d", procStats.RunningProcs), ColorGreen),
		app.colorize(fmt.Sprintf("%d", procStats.SleepingProcs), ColorYellow))

	// Top CPU processes
	fmt.Printf("%s🔥 Top CPU Usage:%s\n", app.colorize("", ColorBold+ColorRed), app.colorize("", ColorReset))
	fmt.Printf("   %-6s %-25s %-12s %8s %10s\n", "PID", "Name", "User", "CPU%", "Memory")
	fmt.Printf("   %s\n", app.colorize(strings.Repeat("─", 65), ColorDim))

	limit := 10
	if app.compactMode {
		limit = 5
	}

	for i, proc := range procStats.TopCPU {
		if i >= limit || proc.CPUPercent < 0.1 {
			break
		}
		cpuColor := app.getUsageColor(float64(proc.CPUPercent))
		fmt.Printf("   %-6d %-25s %-12s %s%7.1f%%%s %9s\n",
			proc.PID,
			app.colorize(app.truncateString(proc.Name, 25), ColorCyan),
			app.colorize(app.truncateString(proc.Username, 12), ColorDim),
			app.colorize("", cpuColor),
			proc.CPUPercent,
			app.colorize("", ColorReset),
			app.colorize(app.formatMB(proc.MemoryMB), ColorYellow))
	}

	fmt.Println()

	// Top Memory processes
	fmt.Printf("%s💾 Top Memory Usage:%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
	fmt.Printf("   %-6s %-25s %-12s %8s %10s\n", "PID", "Name", "User", "Mem%", "Memory")
	fmt.Printf("   %s\n", app.colorize(strings.Repeat("─", 65), ColorDim))

	for i, proc := range procStats.TopMemory {
		if i >= limit || proc.MemPercent < 0.1 {
			break
		}
		memColor := app.getUsageColor(float64(proc.MemPercent))
		fmt.Printf("   %-6d %-25s %-12s %s%7.1f%%%s %9s\n",
			proc.PID,
			app.colorize(app.truncateString(proc.Name, 25), ColorCyan),
			app.colorize(app.truncateString(proc.Username, 12), ColorDim),
			app.colorize("", memColor),
			proc.MemPercent,
			app.colorize("", ColorReset),
			app.colorize(app.formatMB(proc.MemoryMB), ColorYellow))
	}
}

func (app *App) displayNetworkView() {
	netStats, err := internal.GetNetworkStats()
	if err != nil {
		fmt.Printf(app.colorize("Error getting network stats: %v\n", ColorRed), err)
		return
	}

	netSpeeds, _ := internal.GetNetworkSpeeds()

	// Network summary
	fmt.Printf("%s🌐 Network Overview%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("Active Interfaces: %s | Connections: %s\n",
		app.colorize(fmt.Sprintf("%d", netStats.ActiveIfaces), ColorCyan),
		app.colorize(fmt.Sprintf("%d", netStats.Connections), ColorCyan))
	fmt.Printf("Total Traffic: ↑%s ↓%s\n\n",
		app.colorize(internal.FormatNetworkBytes(netStats.TotalSent), ColorRed),
		app.colorize(internal.FormatNetworkBytes(netStats.TotalRecv), ColorGreen))

	// Current speeds
	if len(netSpeeds) > 0 {
		fmt.Printf("%s📊 Current Network Activity:%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
		fmt.Printf("   %-20s %15s %15s %15s\n", "Interface", "Upload", "Download", "Total")
		fmt.Printf("   %s\n", app.colorize(strings.Repeat("─", 70), ColorDim))

		for i, speed := range netSpeeds {
			if i >= 5 {
				break
			}
			totalSpeed := speed.UploadKBps + speed.DownloadKBps
			fmt.Printf("   %-20s %15s %15s %15s\n",
				app.colorize(app.truncateString(speed.Interface, 20), ColorCyan),
				app.colorize(internal.FormatNetworkSpeed(speed.UploadKBps), ColorRed),
				app.colorize(internal.FormatNetworkSpeed(speed.DownloadKBps), ColorGreen),
				app.colorize(internal.FormatNetworkSpeed(totalSpeed), ColorYellow))
		}
		fmt.Println()
	}

	// Interface statistics
	topInterfaces := internal.GetTopNetworkInterfaces(netStats.Interfaces, 8)
	if len(topInterfaces) > 0 {
		fmt.Printf("%s📈 Network Interfaces (Total Traffic):%s\n", app.colorize("", ColorBold+ColorPurple), app.colorize("", ColorReset))
		fmt.Printf("   %-20s %-15s %-15s %8s\n", "Interface", "Sent", "Received", "Status")
		fmt.Printf("   %s\n", app.colorize(strings.Repeat("─", 65), ColorDim))

		for _, iface := range topInterfaces {
			statusColor := ColorRed
			status := "Down"
			if iface.IsUp {
				status = "Up"
				statusColor = ColorGreen
			}

			fmt.Printf("   %-20s %-15s %-15s %s\n",
				app.colorize(app.truncateString(iface.Name, 20), ColorCyan),
				app.colorize(internal.FormatNetworkBytes(iface.BytesSent), ColorRed),
				app.colorize(internal.FormatNetworkBytes(iface.BytesRecv), ColorGreen),
				app.colorize(status, statusColor))
		}
	}
}

func (app *App) displayDisksView() {
	stats, err := internal.GetSystemStats()
	if err != nil {
		fmt.Printf(app.colorize("Error getting system stats: %v\n", ColorRed), err)
		return
	}

	fmt.Printf("%s💽 Disk Usage Details%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
	fmt.Printf("   %-20s %-10s %-12s %-12s %-12s %s\n", "Device", "Usage", "Used", "Free", "Total", "Mount Point")
	fmt.Printf("   %s\n", app.colorize(strings.Repeat("─", 90), ColorDim))

	for _, disk := range stats.Disk {
		device := app.truncateString(filepath.Base(disk.Device), 20)
		usageColor := app.getUsageColor(disk.UsedPercent)

		fmt.Printf("   %-20s %s%9.1f%%%s %-12s %-12s %-12s %s\n",
			app.colorize(device, ColorCyan),
			app.colorize("", usageColor),
			disk.UsedPercent,
			app.colorize("", ColorReset),
			app.colorize(internal.FormatBytes(disk.Used), ColorYellow),
			app.colorize(internal.FormatBytes(disk.Free), ColorGreen),
			app.colorize(internal.FormatBytes(disk.Total), ColorDim),
			app.colorize(app.truncateString(disk.Mountpoint, 20), ColorPurple))

		// Progress bar for each disk
		if !app.compactMode {
			fmt.Printf("   %20s %s\n", "", app.getProgressBar(disk.UsedPercent, 50, usageColor))
		}
	}
}

func (app *App) displaySystemView() {
	stats, err := internal.GetSystemStats()
	if err != nil {
		fmt.Printf(app.colorize("Error getting system stats: %v\n", ColorRed), err)
		return
	}

	// Detailed system information
	fmt.Printf("%s🖥️  Detailed System Information%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
	fmt.Printf("   Hostname:      %s\n", app.colorize(stats.Host.Hostname, ColorCyan))
	fmt.Printf("   Operating System: %s\n", app.colorize(stats.Host.OS, ColorCyan))
	fmt.Printf("   Platform:      %s\n", app.colorize(stats.Host.Platform, ColorCyan))
	fmt.Printf("   Kernel Version: %s\n", app.colorize(stats.Host.KernelVersion, ColorCyan))
	fmt.Printf("   System Uptime: %s\n\n", app.colorize(internal.FormatUptime(stats.Host.Uptime), ColorGreen))

	// Detailed CPU information
	fmt.Printf("%s🔧 CPU Information%s\n", app.colorize("", ColorBold+ColorRed), app.colorize("", ColorReset))
	fmt.Printf("   Model:         %s\n", app.colorize(stats.CPU.ModelName, ColorCyan))
	fmt.Printf("   Logical Cores: %s\n", app.colorize(fmt.Sprintf("%d", stats.CPU.Cores), ColorYellow))
	fmt.Printf("   Current Usage: %s%.1f%%%s\n\n",
		app.colorize("", app.getUsageColor(stats.CPU.Usage)),
		stats.CPU.Usage,
		app.colorize("", ColorReset))

	// Detailed memory information
	fmt.Printf("%s💾 Memory Information%s\n", app.colorize("", ColorBold+ColorBlue), app.colorize("", ColorReset))
	fmt.Printf("   Total:         %s\n", app.colorize(internal.FormatBytes(stats.Memory.Total), ColorCyan))
	fmt.Printf("   Used:          %s (%.1f%%)\n",
		app.colorize(internal.FormatBytes(stats.Memory.Used), ColorYellow),
		stats.Memory.UsedPercent)
	fmt.Printf("   Available:     %s\n", app.colorize(internal.FormatBytes(stats.Memory.Available), ColorGreen))
	fmt.Printf("   Free:          %s\n", app.colorize(internal.FormatBytes(stats.Memory.Free), ColorGreen))
	fmt.Printf("   Buffers:       %s\n", app.colorize(internal.FormatBytes(stats.Memory.Buffers), ColorDim))
	fmt.Printf("   Cached:        %s\n\n", app.colorize(internal.FormatBytes(stats.Memory.Cached), ColorDim))
}

func (app *App) displayFooter() {
	fmt.Println()
	fmt.Print(app.colorize("┌", ColorCyan))
	fmt.Print(app.colorize(strings.Repeat("─", 78), ColorCyan))
	fmt.Print(app.colorize("┐", ColorCyan))
	fmt.Println()

	controls := ""
	if app.logToFile {
		controls += app.colorize("[L]og:ON ", ColorGreen)
	} else {
		controls += app.colorize("[L]og:OFF ", ColorRed)
	}

	if app.paused {
		controls += app.colorize("[P]ause:ON ", ColorYellow)
	} else {
		controls += app.colorize("[P]ause:OFF ", ColorGreen)
	}

	if app.compactMode {
		controls += app.colorize("[C]ompact:ON ", ColorYellow)
	} else {
		controls += app.colorize("[C]ompact:OFF ", ColorGreen)
	}

	fmt.Printf("│ %s%s │\n", controls, strings.Repeat(" ", 78-len(stripColors(controls))))

	shortcuts := app.colorize("[H]elp [E]xport [R]efresh [+/-]Speed [Q]uit", ColorDim)
	fmt.Printf("│ %s%s │\n", shortcuts, strings.Repeat(" ", 78-len(stripColors(shortcuts))))

	fmt.Print(app.colorize("└", ColorCyan))
	fmt.Print(app.colorize(strings.Repeat("─", 78), ColorCyan))
	fmt.Print(app.colorize("┘", ColorCyan))
	fmt.Println()
}

func (app *App) displayHelp() {
	fmt.Printf("%s📚 System Monitor Help%s\n\n", app.colorize("", ColorBold+ColorYellow), app.colorize("", ColorReset))

	fmt.Printf("%sNavigation:%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("  %s1-5%s    Switch between views (Overview, Processes, Network, Disks, System)\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %sH/?%s    Show/hide this help screen\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %sQ%s      Quit the application\n\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))

	fmt.Printf("%sControl:%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("  %sP%s      Pause/resume updates\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %sR%s      Force refresh\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %sC%s      Toggle compact mode\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %s+/-%s    Increase/decrease refresh rate\n\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))

	fmt.Printf("%sLogging & Export:%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("  %sL%s      Toggle logging to file\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %sE%s      Export current stats to JSON file\n\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))

	fmt.Printf("%sColor Legend:%s\n", app.colorize("", ColorBold+ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("  %s●%s Low usage (< 60%%)\n", app.colorize("", ColorGreen), app.colorize("", ColorReset))
	fmt.Printf("  %s●%s Medium usage (60-80%%)\n", app.colorize("", ColorYellow), app.colorize("", ColorReset))
	fmt.Printf("  %s●%s High usage (> 80%%)\n\n", app.colorize("", ColorRed), app.colorize("", ColorReset))

	fmt.Printf("%sPress any key to return...%s", app.colorize("", ColorDim), app.colorize("", ColorReset))
}

// Helper functions
func (app *App) colorize(text string, color string) string {
	if !app.colorEnabled {
		return text
	}
	return color + text + ColorReset
}

func (app *App) getUsageColor(percent float64) string {
	if percent > 80 {
		return ColorRed
	} else if percent > 60 {
		return ColorYellow
	}
	return ColorGreen
}

func (app *App) getProgressBar(percent float64, width int, color string) string {
	filled := int(percent / 100 * float64(width))
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			if percent > 80 {
				bar += app.colorize("█", ColorRed)
			} else if percent > 60 {
				bar += app.colorize("▓", ColorYellow)
			} else {
				bar += app.colorize("▒", ColorGreen)
			}
		} else {
			bar += app.colorize("░", ColorDim)
		}
	}
	bar += app.colorize("]", ColorReset)
	return bar
}

func (app *App) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (app *App) formatMB(mb uint64) string {
	if mb >= 1024 {
		return fmt.Sprintf("%.1fGB", float64(mb)/1024)
	}
	return fmt.Sprintf("%dMB", mb)
}

func (app *App) clearScreen() {
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
}

func (app *App) toggleLogging() {
	if app.logToFile {
		if app.logFile != nil {
			app.logFile.Close()
			app.logFile = nil
		}
		app.logToFile = false
	} else {
		// Create logs directory if it doesn't exist
		os.MkdirAll("logs", 0755)

		// Create log file with timestamp
		filename := fmt.Sprintf("logs/sysmon_%s.log", time.Now().Format("20060102_150405"))
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Error creating log file: %v", err)
			return
		}
		app.logFile = file
		app.logToFile = true
	}
	app.displayInterface()
}

func (app *App) logStats(stats *internal.SystemStats, procStats *internal.ProcessStats, netStats *internal.NetworkStats) {
	if app.logFile == nil {
		return
	}

	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"system":    stats,
		"processes": procStats,
		"network":   netStats,
	}

	data, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Error marshaling log entry: %v", err)
		return
	}

	_, err = app.logFile.Write(append(data, '\n'))
	if err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

func (app *App) exportStats() {
	// Create exports directory if it doesn't exist
	os.MkdirAll("exports", 0755)

	// Get current stats
	stats, err := internal.GetSystemStats()
	if err != nil {
		log.Printf("Error getting stats for export: %v", err)
		return
	}

	procStats, _ := internal.GetProcessStats()
	netStats, _ := internal.GetNetworkStats()

	exportData := map[string]interface{}{
		"export_timestamp": time.Now().Format(time.RFC3339),
		"system":           stats,
		"processes":        procStats,
		"network":          netStats,
		"view":             app.currentView,
		"refresh_rate":     app.refreshRate.String(),
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("exports/sysmon_export_%s.json", time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating export file: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(exportData); err != nil {
		log.Printf("Error encoding export data: %v", err)
		return
	}

	log.Printf("Stats exported to %s", filename)
}

func (app *App) cleanup() {
	if app.logFile != nil {
		app.logFile.Close()
	}
	app.clearScreen()
	fmt.Println("System Monitor shutdown complete. Goodbye!")
}

func handleKeyboardInput(inputChan chan rune) {
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			close(inputChan)
			return
		}
		inputChan <- char
	}
}

func stripColors(text string) string {
	// Remove ANSI color codes
	re := regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(text, "")
}

