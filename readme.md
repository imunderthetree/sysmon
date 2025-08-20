# 🖥️ System Monitor v2.0

A powerful, real-time system monitoring tool built in Go that provides comprehensive insights into your system's performance with an intuitive terminal-based interface.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=flat-square)

## ✨ Features

### 📊 Multiple Monitoring Views
- **Overview**: Complete system summary with key metrics
- **Processes**: Detailed process monitoring with CPU and memory usage
- **Network**: Real-time network activity and interface statistics
- **Disks**: Comprehensive disk usage information
- **System**: In-depth system information and specifications

### 🎮 Interactive Controls
- **Real-time Updates**: Configurable refresh rates (1-10 seconds)
- **Pause/Resume**: Pause monitoring to examine specific moments
- **Compact Mode**: Space-efficient display for smaller terminals
- **Keyboard Navigation**: Intuitive single-key commands

### 📈 Advanced Features
- **Data Export**: JSON export functionality for analysis
- **Logging**: Optional file logging with timestamps
- **Progress Bars**: Visual representation of resource usage
- **Color-coded Metrics**: Intuitive color scheme for quick assessment
- **Cross-platform**: Works on Linux, macOS, and Windows

## 🚀 Quick Start

### Prerequisites
- Go 1.25 or higher
- Terminal with color support (recommended)

### Installation

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/sysmon.git
cd sysmon
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Build the application:**
```bash
go build -o sysmon
```

4. **Run the system monitor:**
```bash
./sysmon
```

### One-liner Installation
```bash
git clone https://github.com/yourusername/sysmon.git && cd sysmon && go mod download && go build -o sysmon && ./sysmon
```

## 🎯 Usage

### Navigation
| Key | Action |
|-----|--------|
| `1-5` | Switch between views (Overview, Processes, Network, Disks, System) |
| `H` or `?` | Show/hide help screen |
| `Q` | Quit application |

### Control
| Key | Action |
|-----|--------|
| `P` | Pause/resume updates |
| `R` | Force refresh |
| `C` | Toggle compact mode |
| `+/-` | Increase/decrease refresh rate |

### Data Management
| Key | Action |
|-----|--------|
| `L` | Toggle logging to file |
| `E` | Export current stats to JSON |

## 📸 Screenshots

### Overview View
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ System Monitor v2.0 - Overview View                               RUNNING │
│ 14:23:45                                                    Refresh: 3s │
├──────────────────────────────────────────────────────────────────────────────┤
│ [1]Overview [2]Processes [3]Network [4]Disks [5]System                      │
└──────────────────────────────────────────────────────────────────────────────┘

🖥️ System Information
   Hostname: my-computer | OS: linux | Uptime: 2d 14h 23m

🔧 CPU Usage: 15.2% ████████████████░░░░░░░░░░░░░░░░░░░░░░░░
   Cores: 8 | Model: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz

💾 Memory: 45.3% ██████████████████░░░░░░░░░░░░░░░░░░░░░░
   Used: 7.2GB / 16.0GB | Free: 8.8GB
```

## 🏗️ Architecture

The project is organized into clean, modular components:

```
sysmon/
├── main.go              # Main application and UI logic
├── internal/
│   ├── stats.go         # System statistics collection
│   ├── processes.go     # Process monitoring
│   └── network.go       # Network statistics
├── go.mod              # Go module definition
├── logs/               # Generated log files (when logging enabled)
└── exports/            # Generated export files
```

### Key Components

- **Main Application** (`main.go`): Terminal UI, keyboard handling, and view management
- **System Stats** (`internal/stats.go`): CPU, memory, disk, and host information
- **Process Monitor** (`internal/processes.go`): Process enumeration and statistics
- **Network Monitor** (`internal/network.go`): Network interface and traffic monitoring

## 🔧 Configuration

### Environment Variables
Currently, the application uses default settings. Future versions will support:
- `SYSMON_REFRESH_RATE`: Default refresh rate
- `SYSMON_LOG_DIR`: Custom log directory
- `SYSMON_EXPORT_DIR`: Custom export directory

### Customization
The application supports runtime customization through keyboard shortcuts:
- Refresh rate: Adjustable from 1-10 seconds
- Display modes: Normal and compact views
- Color themes: Automatic based on terminal capabilities

## 📊 Data Export Format

Exported JSON includes comprehensive system information:

```json
{
  "export_timestamp": "2024-01-15T14:23:45Z",
  "system": {
    "cpu": { "usage": 15.2, "cores": 8 },
    "memory": { "total": 16777216000, "used": 7516192768 },
    "disk": [...]
  },
  "processes": {
    "total_processes": 245,
    "top_cpu": [...],
    "top_memory": [...]
  },
  "network": {
    "interfaces": [...],
    "total_sent": 1024000,
    "total_recv": 2048000
  }
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
6. Push to the branch (`git push origin feature/AmazingFeature`)
7. Open a Pull Request

### Code Style
- Follow standard Go formatting (`go fmt`)
- Add comments for exported functions
- Keep functions focused and modular
- Use meaningful variable names

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gopsutil](https://github.com/shirou/gopsutil) - Cross-platform library for system and process monitoring
- [Go team](https://golang.org/) - For the excellent Go programming language
- All contributors who help improve this project

## 🐛 Known Issues

- Process CPU usage calculation may take a moment to stabilize on first run
- Some system information may not be available on all platforms
- Network speed calculations require at least two measurement cycles

## 🚧 Roadmap

- [ ] Historical data tracking and graphs
- [ ] Web-based dashboard
- [ ] Alert system for resource thresholds
- [ ] Plugin system for custom monitors
- [ ] Configuration file support
- [ ] Docker containerization

## 📞 Support

If you encounter any issues or have questions:
1. Check the [Issues](https://github.com/yourusername/sysmon/issues) page
2. Create a new issue with detailed information
3. Include your operating system and Go version

---

**Made with ❤️ and Go**
