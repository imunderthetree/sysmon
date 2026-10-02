package gui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"sysmon/internal"
	"time"
)

//go:embed web/*
var webFS embed.FS

// HistoryData holds time series points for charts
type HistoryData struct {
	CPU     []float64               `json:"cpu"`
	Memory  []float64               `json:"memory"`
	Network []internal.NetworkSpeed `json:"network"`
}

// AppState manages the application data collection, web server, and SSE broadcasts
type AppState struct {
	mu            sync.RWMutex
	systemStats   *internal.SystemStats
	processStats  *internal.ProcessStats
	networkStats  *internal.NetworkStats
	networkSpeeds []internal.NetworkSpeed

	cpuHistory    []float64
	memHistory    []float64
	netHistory    []internal.NetworkSpeed

	refreshRate   time.Duration
	paused        bool
	clients       map[chan []byte]bool
	clientsMu     sync.Mutex
	stopChan      chan struct{}
	server        *http.Server
	Port          int
	OpenBrowserUI bool
}

// NewApp creates a new GUI/Web application state
func NewApp() *AppState {
	return &AppState{
		refreshRate:   1 * time.Second,
		paused:        false,
		cpuHistory:    make([]float64, 0, 60),
		memHistory:    make([]float64, 0, 60),
		netHistory:    make([]internal.NetworkSpeed, 0, 60),
		clients:       make(map[chan []byte]bool),
		stopChan:      make(chan struct{}),
		Port:          8080,
		OpenBrowserUI: true,
	}
}

// Run starts the background monitor, HTTP server, and opens the user's browser
func (app *AppState) Run() {
	// Start collector
	go app.startCollector()

	// Setup HTTP handler
	handler := app.setupRoutes()

	// Find available port
	listener, port, err := findAvailablePort(app.Port)
	if err != nil {
		log.Fatalf("Failed to bind port: %v", err)
	}
	app.Port = port
	app.server = &http.Server{Handler: handler}

	url := fmt.Sprintf("http://localhost:%d", app.Port)

	fmt.Println("\n========================================================")
	fmt.Printf(" 🖥️  SysMon Web GUI running at: %s\n", url)
	fmt.Println("    Press Ctrl+C to stop the monitor")
	fmt.Println("========================================================\n")

	// Open browser
	if app.OpenBrowserUI {
		go func() {
			time.Sleep(200 * time.Millisecond)
			openBrowser(url)
		}()
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down SysMon GUI...")
		close(app.stopChan)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		app.server.Shutdown(ctx)
	}()

	if err := app.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// startCollector periodically fetches metrics and pushes to SSE clients
func (app *AppState) startCollector() {
	// First immediate collection
	app.collectMetrics()

	app.mu.RLock()
	currentRate := app.refreshRate
	app.mu.RUnlock()

	ticker := time.NewTicker(currentRate)
	defer ticker.Stop()

	for {
		select {
		case <-app.stopChan:
			return
		case <-ticker.C:
			app.mu.RLock()
			paused := app.paused
			rate := app.refreshRate
			app.mu.RUnlock()

			if rate != currentRate {
				currentRate = rate
				ticker.Reset(rate)
			}

			if !paused {
				app.collectMetrics()
			}
		}
	}
}

func (app *AppState) collectMetrics() {
	sysStats, _ := internal.GetSystemStats()
	procStats, _ := internal.GetProcessStats()
	netStats, _ := internal.GetNetworkStats()
	netSpeeds, _ := internal.GetNetworkSpeeds()

	app.mu.Lock()
	if sysStats != nil {
		app.systemStats = sysStats
		app.cpuHistory = append(app.cpuHistory, sysStats.CPU.Usage)
		if len(app.cpuHistory) > 60 {
			app.cpuHistory = app.cpuHistory[len(app.cpuHistory)-60:]
		}
		app.memHistory = append(app.memHistory, sysStats.Memory.UsedPercent)
		if len(app.memHistory) > 60 {
			app.memHistory = app.memHistory[len(app.memHistory)-60:]
		}
	}
	if procStats != nil {
		app.processStats = procStats
	}
	if netStats != nil {
		app.networkStats = netStats
	}
	if netSpeeds != nil {
		app.networkSpeeds = netSpeeds
		var aggregateSpeed internal.NetworkSpeed
		for _, s := range netSpeeds {
			aggregateSpeed.DownloadKBps += s.DownloadKBps
			aggregateSpeed.UploadKBps += s.UploadKBps
		}
		app.netHistory = append(app.netHistory, aggregateSpeed)
		if len(app.netHistory) > 60 {
			app.netHistory = app.netHistory[len(app.netHistory)-60:]
		}
	}
	payload := app.buildPayloadLocked()
	app.mu.Unlock()

	// Broadcast to SSE clients
	data, err := json.Marshal(payload)
	if err == nil {
		app.broadcast(data)
	}
}

func (app *AppState) buildPayloadLocked() map[string]interface{} {
	return map[string]interface{}{
		"system":          app.systemStats,
		"processes":       app.processStats,
		"network":         app.networkStats,
		"speeds":          app.networkSpeeds,
		"paused":          app.paused,
		"refresh_rate_ms": app.refreshRate.Milliseconds(),
		"history": map[string]interface{}{
			"cpu":     app.cpuHistory,
			"memory":  app.memHistory,
			"network": app.netHistory,
		},
	}
}

func (app *AppState) broadcast(data []byte) {
	app.clientsMu.Lock()
	defer app.clientsMu.Unlock()

	msg := append([]byte("data: "), data...)
	msg = append(msg, []byte("\n\n")...)

	for ch := range app.clients {
		select {
		case ch <- msg:
		default:
			// Client blocked or slow, skip
		}
	}
}

func (app *AppState) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Static web files from embed.FS
	staticContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to initialize web filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticContent))

	mux.Handle("/", fileServer)

	// API Stats
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		app.mu.RLock()
		payload := app.buildPayloadLocked()
		app.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	})

	// Server-Sent Events (SSE)
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		msgChan := make(chan []byte, 10)

		app.clientsMu.Lock()
		app.clients[msgChan] = true
		app.clientsMu.Unlock()

		defer func() {
			app.clientsMu.Lock()
			delete(app.clients, msgChan)
			app.clientsMu.Unlock()
		}()

		// Send initial snapshot immediately
		app.mu.RLock()
		initialPayload := app.buildPayloadLocked()
		app.mu.RUnlock()

		if initialData, err := json.Marshal(initialPayload); err == nil {
			fmt.Fprintf(w, "data: %s\n\n", initialData)
			flusher.Flush()
		}

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgChan:
				w.Write(msg)
				flusher.Flush()
			}
		}
	})

	// Settings update
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Paused        *bool `json:"paused"`
			RefreshRateMs *int  `json:"refresh_rate_ms"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		app.mu.Lock()
		if req.Paused != nil {
			app.paused = *req.Paused
		}
		if req.RefreshRateMs != nil && *req.RefreshRateMs >= 250 {
			app.refreshRate = time.Duration(*req.RefreshRateMs) * time.Millisecond
		}
		payload := app.buildPayloadLocked()
		app.mu.Unlock()

		// Broadcast state update
		if data, err := json.Marshal(payload); err == nil {
			app.broadcast(data)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// Process Kill
	mux.HandleFunc("/api/process/kill", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PID int `json:"pid"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		proc, err := os.FindProcess(req.PID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		if err := proc.Kill(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// Export stats
	mux.HandleFunc("/api/export", func(w http.ResponseWriter, r *http.Request) {
		app.mu.RLock()
		payload := app.buildPayloadLocked()
		app.mu.RUnlock()

		filename := fmt.Sprintf("sysmon_export_%s.json", time.Now().Format("20060102_150405"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(payload)
	})

	return mux
}

// findAvailablePort finds an open port starting from startPort
func findAvailablePort(startPort int) (net.Listener, int, error) {
	for port := startPort; port < startPort+100; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return listener, port, nil
		}
	}
	return nil, 0, fmt.Errorf("no available port found starting from %d", startPort)
}

// openBrowser launches the URL in the default web browser
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open browser: %v. Please open %s manually.", err, url)
	}
}
