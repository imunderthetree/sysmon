package gui

import (
	"fmt"
	"sync"
	"sysmon/internal"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// HistoryPoint represents a single data point for charting
type HistoryPoint struct {
	Timestamp time.Time
	Value     float64
}

// ThemeMode represents light or dark theme
type ThemeMode int

const (
	ThemeLight ThemeMode = iota
	ThemeDark
)

// AppState manages the application state and data collection
type AppState struct {
	fyneApp    fyne.App
	mainWindow fyne.Window

	// Refresh control
	ticker      *time.Ticker
	stopChan    chan bool
	refreshRate time.Duration
	paused      bool

	// Data storage for charts/display (keep last 60 points)
	cpuHistory     []*HistoryPoint
	memoryHistory  []*HistoryPoint
	networkHistory []*HistoryPoint
	networkUpHistory []*HistoryPoint

	// Current stats
	systemStats   *internal.SystemStats
	processStats  *internal.ProcessStats
	networkStats  *internal.NetworkStats

	// UI components
	tabs          *container.AppTabs
	statusLabel   *widget.Label
	pauseButton   *widget.Button
	themeToggle   *widget.Button
	refreshSlider *widget.Slider

	// UI state
	currentTheme ThemeMode
	mutex        sync.RWMutex
}

// NewApp creates and initializes a new GUI application
func NewApp() *AppState {
	fyneApp := app.NewWithID("sysmon")
	mainWindow := fyneApp.NewWindow()
	mainWindow.Resize(fyne.NewSize(1200, 700))

	state := &AppState{
		fyneApp:     fyneApp,
		mainWindow:  mainWindow,
		refreshRate: 3 * time.Second,
		paused:      false,
		stopChan:    make(chan bool),
		currentTheme: ThemeLight,

		cpuHistory:       make([]*HistoryPoint, 0, 60),
		memoryHistory:    make([]*HistoryPoint, 0, 60),
		networkHistory:   make([]*HistoryPoint, 0, 60),
		networkUpHistory: make([]*HistoryPoint, 0, 60),
	}

	// Apply theme
	state.applyTheme()

	// Create UI
	state.createUI()

	// Start data collection loop
	go state.dataCollectionLoop()

	// Handle window close
	mainWindow.SetOnClosed(func() {
		state.stopChan <- true
	})

	return state
}

// createUI builds the main UI layout
func (s *AppState) createUI() {
	// Create tabs
	s.tabs = container.NewAppTabs()
	s.tabs.OnChanged = func(tab *container.TabItem) {
		s.updateDisplay()
	}

	// Create tab items
	s.tabs.Append(container.NewTabItem("Overview", s.newOverviewTab()))
	s.tabs.Append(container.NewTabItem("Processes", s.newProcessesTab()))
	s.tabs.Append(container.NewTabItem("Network", s.newNetworkTab()))
	s.tabs.Append(container.NewTabItem("Disks", s.newDisksTab()))
	s.tabs.Append(container.NewTabItem("System", s.newSystemTab()))

	// Create status bar with controls
	s.statusLabel = widget.NewLabel("Ready")
	s.pauseButton = widget.NewButton("Pause", s.togglePause)
	s.themeToggle = widget.NewButton("🌙 Dark", s.toggleTheme)

	s.refreshSlider = widget.NewSlider(1, 10)
	s.refreshSlider.Value = 3
	s.refreshSlider.OnChanged = s.changeRefreshRate
	s.refreshSlider.Step = 1

	refreshLabel := widget.NewLabel("Refresh Rate:")
	refreshContainer := container.NewBorder(refreshLabel, nil, nil, nil, s.refreshSlider)

	// Control bar
	controlBar := container.NewBorder(nil, nil, s.pauseButton, s.themeToggle, container.NewVBox(
		s.statusLabel,
		refreshContainer,
	))

	// Main layout
	mainContent := container.NewBorder(nil, controlBar, nil, nil, s.tabs)
	s.mainWindow.SetContent(mainContent)
}


// dataCollectionLoop runs the main data collection and refresh loop
func (s *AppState) dataCollectionLoop() {
	s.ticker = time.NewTicker(s.refreshRate)
	defer s.ticker.Stop()

	// Initial data fetch
	s.fetchData()

	for {
		select {
		case <-s.ticker.C:
			if !s.paused {
				s.fetchData()
				s.updateDisplay()
			}
		case <-s.stopChan:
			return
		}
	}
}

// fetchData collects current system statistics
func (s *AppState) fetchData() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get system stats
	if stats, err := internal.GetSystemStats(); err == nil {
		s.systemStats = stats

		// Add to history (keep last 60 points)
		s.cpuHistory = append(s.cpuHistory, &HistoryPoint{
			Timestamp: time.Now(),
			Value:     stats.CPU.Usage,
		})
		if len(s.cpuHistory) > 60 {
			s.cpuHistory = s.cpuHistory[1:]
		}

		s.memoryHistory = append(s.memoryHistory, &HistoryPoint{
			Timestamp: time.Now(),
			Value:     stats.Memory.UsedPercent,
		})
		if len(s.memoryHistory) > 60 {
			s.memoryHistory = s.memoryHistory[1:]
		}
	}

	// Get process stats
	if pstats, err := internal.GetProcessStats(); err == nil {
		s.processStats = pstats
	}

	// Get network stats
	if nstats, err := internal.GetNetworkStats(); err == nil {
		s.networkStats = nstats
	}

	// Get network speeds
	if speeds, err := internal.GetNetworkSpeeds(); err == nil && len(speeds) > 0 {
		totalDown := 0.0
		for _, speed := range speeds {
			totalDown += speed.DownloadKBps
		}
		s.networkHistory = append(s.networkHistory, &HistoryPoint{
			Timestamp: time.Now(),
			Value:     totalDown,
		})
		if len(s.networkHistory) > 60 {
			s.networkHistory = s.networkHistory[1:]
		}
	}
}

// updateDisplay refreshes the UI with current data
func (s *AppState) updateDisplay() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Update status label
	if s.systemStats != nil {
		status := fmt.Sprintf("CPU: %.1f%% | Memory: %.1f%% | %s",
			s.systemStats.CPU.Usage,
			s.systemStats.Memory.UsedPercent,
			time.Now().Format("15:04:05"))
		s.statusLabel.SetText(status)
	}

	// Note: Detailed tab updates will be implemented in separate tab files
}

// togglePause pauses/resumes data collection
func (s *AppState) togglePause() {
	s.paused = !s.paused
	if s.paused {
		s.pauseButton.SetText("Resume")
	} else {
		s.pauseButton.SetText("Pause")
	}
}

// toggleTheme switches between light and dark theme
func (s *AppState) toggleTheme() {
	s.mutex.Lock()
	if s.currentTheme == ThemeLight {
		s.currentTheme = ThemeDark
	} else {
		s.currentTheme = ThemeLight
	}
	s.mutex.Unlock()

	s.applyTheme()
}

// applyTheme applies the current theme to the app
func (s *AppState) applyTheme() {
	// For now, Fyne uses the system theme. Additional customization can be added here.
	if s.currentTheme == ThemeDark {
		if s.themeToggle != nil {
			s.themeToggle.SetText("☀️ Light")
		}
	} else {
		if s.themeToggle != nil {
			s.themeToggle.SetText("🌙 Dark")
		}
	}
}

// changeRefreshRate updates the refresh rate based on slider
func (s *AppState) changeRefreshRate(value float64) {
	newRate := time.Duration(value) * time.Second
	if newRate != s.refreshRate {
		s.refreshRate = newRate
		// Stop and restart ticker with new rate
		if s.ticker != nil {
			s.ticker.Stop()
		}
		s.ticker = time.NewTicker(s.refreshRate)
	}
}

// Run starts the GUI application
func (s *AppState) Run() {
	s.mainWindow.ShowAndRun()
}
