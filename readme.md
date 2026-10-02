# 🖥️ SysMon v2.0 — Real-Time System Observability Cockpit

A powerful, ultra-responsive real-time system monitoring tool built in Go that provides comprehensive insights into your machine's performance with both a modern **Web GUI Dashboard** and a high-performance **Terminal UI (TUI)**.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)
![Build](https://img.shields.io/badge/Build-Pure%20Go%20(Zero%20CGO)-success?style=flat-square)

---

## ✨ Features

### 🖼️ Dual Interface Modes
- **Modern Web GUI** *(Default)*: Embedded single-page telemetry dashboard with live retina canvas charts, process management, dark/light themes, and real-time Server-Sent Events (SSE). **Zero external C compilers or CGO required**—compiles out of the box on any system.
- **Terminal TUI Mode**: Lightweight, interactive terminal interface with ANSI colors and keyboard shortcuts designed for headless servers and SSH sessions.

### 📊 Multiple Monitoring Views
- **📊 Overview**: At-a-glance system health summary with CPU, RAM, Disk, and Network throughput gauges, live mini trend charts, and top workload tables.
- **⚡ CPU & RAM**: High-resolution 60-frame continuous load graphs, detailed memory breakdown (Installed, Active Used, Free, Cache, Buffers), and a **Logical Processor Core Grid** displaying per-core load in real time.
- **💻 Process Explorer**: Live process table with instantaneous search, multi-column sorting (PID, Name, User, CPU%, Mem%), category filter chips, and interactive **Kill Process** controls.
- **🌐 Network**: Real-time duplex throughput meters (Download Rx / Upload Tx) with aggregate transfer counters and detailed per-adapter statistics (packets, errors, drops).
- **💾 Disks**: Visual storage volume cards with mountpoints, file system types, and capacity progress bars.
- **ℹ️ System**: Comprehensive host specifications, OS build, kernel release, and live uptime clock.

### ⚡ Task Manager-Style Telemetry & Continuous Updates
- **Live Taskbar & Browser Tab Title**: Live CPU and RAM stats stream directly into the browser tab title and Windows taskbar button (e.g., `⚡ CPU 14% · RAM 78% — SysMon`), letting you glance at system performance without switching windows.
- **Dynamic Status Favicon**: Tab icon dynamically shifts color (cyan, amber, red) based on CPU load.
- **Logical Processor Cores Grid**: Real-time activity meters for every logical CPU core (e.g. 16 cores), matching the Windows Task Manager Performance view.
- **Background Keep-Alive**: Uses the Page Visibility API and SSE auto-reconnect to stream uninterrupted even when minimized or running in background tabs.
- **High-Frequency Update Speeds**: Configurable polling rates from `0.5s (Real-Time)`, `1.0s (Normal)`, `2.0s`, to `5.0s`.

### 🛠️ Developer & Power User Controls
- **Bespoke Vector Branding**: Crisp SVG telemetry logo and vector icon system—no emoji placeholders.
- **One-Click JSON Data Export**: Instant download of complete point-in-time system snapshots.
- **Dark & Light Mode**: Engineered dark cockpit theme with accessible light mode, persisted via `localStorage`.
- **Pure Go Architecture**: Zero CGO, zero MinGW/GCC toolchains, zero Node.js build steps. Embedded static assets via Go's standard `embed.FS`.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.25 or higher
- Any modern web browser (Edge, Chrome, Firefox, Safari)

### Installation & Build

1. **Clone the repository:**
   ```bash
   git clone https://github.com/imunderthetree/sysmon.git
   cd sysmon
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Build the binary:**
   ```bash
   # Windows
   go build -o sysmon.exe

   # Linux / macOS
   go build -o sysmon
   ```

---

## 🎯 Usage

### Running the Web GUI (Default)
Running `sysmon` without flags starts the local web server and automatically launches your default web browser:
```bash
./sysmon
```
> The dashboard will be available at **`http://localhost:8080`** (or the next available port).

### Command-Line Flags
| Flag | Default | Description |
|------|---------|-------------|
| `-gui` | `true` | Run in Web GUI mode |
| `-tui` | `false` | Run in interactive Terminal UI mode |
| `-port <number>` | `8080` | Port for the Web GUI server |
| `-no-browser` | `false` | Start server without auto-opening the browser (ideal for headless servers) |

#### Examples
```bash
# Run Terminal UI (TUI)
./sysmon -tui

# Run headless Web GUI on custom port 9090
./sysmon -port 9090 -no-browser
```

---

## ⌨️ Terminal UI (TUI) Controls

When running in `-tui` mode, use single-key keyboard commands:

### Navigation
| Key | Action |
|-----|--------|
| `1` - `5` | Switch view (`1`: Overview, `2`: Processes, `3`: Network, `4`: Disks, `5`: System) |
| `H` or `?` | Toggle help overlay |
| `Q` | Graceful shutdown |

### Control & Settings
| Key | Action |
|-----|--------|
| `P` | Pause / resume real-time updates |
| `R` | Force refresh |
| `C` | Toggle compact layout |
| `+` / `-` | Increase / decrease refresh frequency |
| `L` | Toggle file logging to `logs/` |
| `E` | Export current snapshot to JSON in `exports/` |

---

## 🏗️ Project Architecture

```
sysmon/
├── gui/                      # Web GUI Subsystem
│   ├── app.go                # HTTP server, SSE broadcaster, telemetry collector
│   └── web/                  # Embedded frontend assets (zero CDN dependencies)
│       ├── index.html        # Semantic HTML5 single-page application
│       ├── style.css         # Modern dark-tech telemetry design system
│       └── app.js            # Client-side state, SSE receiver, and Canvas charts
├── internal/                 # System Telemetry Engine (Pure Go)
│   ├── stats.go              # CPU, logical cores, memory, disk, host info
│   ├── processes.go          # Process enumeration, sorting, and stats
│   └── network.go            # Network adapters and speed calculations
├── gui_init.go               # Web GUI bootstrap logic
├── main_default.go           # CLI flags and application entry point
├── main_tui.go               # TUI build-tagged entry point
├── main.go                   # Terminal UI rendering engine
├── go.mod                    # Module definitions
└── readme.md                 # Documentation
```

### Core API Endpoints
When running in Web GUI mode, SysMon exposes a lightweight REST & streaming API:
* `GET /`: Serves the embedded web dashboard.
* `GET /api/stats`: Returns current system telemetry JSON.
* `GET /api/events`: Server-Sent Events (SSE) live telemetry stream.
* `POST /api/settings`: Updates update interval or paused state (`{ "paused": bool, "refresh_rate_ms": int }`).
* `POST /api/process/kill`: Terminates a process by PID (`{ "pid": int }`).
* `GET /api/export`: Direct JSON download attachment of current metrics.

---

## 📊 Export Data Schema

Exported JSON contains comprehensive hardware, OS, and process metadata:

```json
{
  "system": {
    "cpu": {
      "usage": 14.8,
      "cores": 16,
      "model_name": "12th Gen Intel(R) Core(TM) i5-12600HX",
      "per_core": [12.5, 18.2, 8.4, 22.1, ...]
    },
    "memory": {
      "total": 12578578432,
      "available": 2665701376,
      "used": 9912877056,
      "used_percent": 78.8
    },
    "disk": [...],
    "host": {
      "hostname": "Workstation",
      "os": "windows",
      "platform": "Microsoft Windows 11 Pro",
      "uptime": 595199
    }
  },
  "processes": {
    "total_processes": 284,
    "running_processes": 12,
    "sleeping_processes": 272,
    "top_cpu": [...],
    "top_memory": [...]
  },
  "network": {
    "interfaces": [...],
    "total_sent": 412048590,
    "total_recv": 8943189210
  }
}
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
1. Fork the project
2. Create your feature branch (`git checkout -b feature/NewFeature`)
3. Commit your changes (`git commit -m 'Add NewFeature'`)
4. Push to the branch (`git push origin feature/NewFeature`)
5. Open a Pull Request

---

## 📝 License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
