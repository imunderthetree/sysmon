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

// SystemTab represents the system information view
func (s *AppState) newSystemTab() fyne.CanvasObject {
	// Title
	title := canvas.NewText("System Information", color.White)
	title.TextSize = 24

	// Create labels for all system info
	hostLabel := widget.NewLabel("Host: --")
	osLabel := widget.NewLabel("OS: --")
	platformLabel := widget.NewLabel("Platform: --")
	kernelLabel := widget.NewLabel("Kernel: --")
	uptimeLabel := widget.NewLabel("Uptime: --")

	cpuModelLabel := widget.NewLabel("CPU Model: --")
	cpuCoresLabel := widget.NewLabel("CPU Cores: --")
	cpuUsageLabel := widget.NewLabel("CPU Usage: --")

	memTotalLabel := widget.NewLabel("Total Memory: --")
	memUsedLabel := widget.NewLabel("Used Memory: --")
	memFreeLabel := widget.NewLabel("Free Memory: --")
	memBuffersLabel := widget.NewLabel("Buffers: --")
	memCachedLabel := widget.NewLabel("Cached: --")

	// Start update loop
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			default:
				s.mutex.RLock()
				if s.systemStats != nil {
					stats := s.systemStats

					// Update host info
					hostLabel.SetText(fmt.Sprintf("Host: %s", stats.Host.Hostname))
					osLabel.SetText(fmt.Sprintf("OS: %s", stats.Host.OS))
					platformLabel.SetText(fmt.Sprintf("Platform: %s", stats.Host.Platform))
					kernelLabel.SetText(fmt.Sprintf("Kernel: %s", stats.Host.KernelVersion))
					uptimeLabel.SetText(fmt.Sprintf("Uptime: %s", internal.FormatUptime(stats.Host.Uptime)))

					// Update CPU info
					cpuModelLabel.SetText(fmt.Sprintf("CPU Model: %s", stats.CPU.ModelName))
					cpuCoresLabel.SetText(fmt.Sprintf("CPU Cores: %d", stats.CPU.Cores))
					cpuUsageLabel.SetText(fmt.Sprintf("CPU Usage: %.1f%%", stats.CPU.Usage))

					// Update memory info
					memTotalLabel.SetText(fmt.Sprintf("Total Memory: %s", internal.FormatBytes(stats.Memory.Total)))
					memUsedLabel.SetText(fmt.Sprintf("Used Memory: %s (%.1f%%)", internal.FormatBytes(stats.Memory.Used), stats.Memory.UsedPercent))
					memFreeLabel.SetText(fmt.Sprintf("Free Memory: %s", internal.FormatBytes(stats.Memory.Free)))
					memBuffersLabel.SetText(fmt.Sprintf("Buffers: %s", internal.FormatBytes(stats.Memory.Buffers)))
					memCachedLabel.SetText(fmt.Sprintf("Cached: %s", internal.FormatBytes(stats.Memory.Cached)))
				}
				s.mutex.RUnlock()

				// Use timer for updates
				<-time.NewTimer(s.refreshRate / 2).C
			}
		}
	}()

	// Layout
	hostInfoBox := container.NewVBox(
		canvas.NewText("Host Information", color.White),
		hostLabel,
		osLabel,
		platformLabel,
		kernelLabel,
		uptimeLabel,
	)

	cpuInfoBox := container.NewVBox(
		canvas.NewText("CPU Information", color.White),
		cpuModelLabel,
		cpuCoresLabel,
		cpuUsageLabel,
	)

	memInfoBox := container.NewVBox(
		canvas.NewText("Memory Information", color.White),
		memTotalLabel,
		memUsedLabel,
		memFreeLabel,
		memBuffersLabel,
		memCachedLabel,
	)

	mainContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		hostInfoBox,
		widget.NewSeparator(),
		cpuInfoBox,
		widget.NewSeparator(),
		memInfoBox,
	)

	scroll := container.NewScroll(mainContent)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	return scroll
}
