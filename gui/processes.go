package gui

import (
	"fmt"
	"image/color"
	"os"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ProcessesTab represents the processes view
type ProcessesTab struct {
	appState *AppState
	table    *widget.Table
	window   fyne.Window
}

// NewProcessesTab creates a new processes tab
func (s *AppState) newProcessesTab() fyne.CanvasObject {
	tab := &ProcessesTab{appState: s, window: s.mainWindow}

	// Title
	title := canvas.NewText("Top Processes", color.White)
	title.TextSize = 24

	// Create table
	tab.table = widget.NewTable(
		func() (int, int) { return 11, 5 }, // 10 rows + 1 header, 5 columns
		func() fyne.CanvasObject {
			return container.NewVBox(widget.NewLabel("Cell"))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*fyne.Container).Objects[0].(*widget.Label)

			if id.Row == 0 {
				// Header row
				headers := []string{"PID", "Name", "User", "CPU %", "Memory MB"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				// Data rows
				s.mutex.RLock()
				defer s.mutex.RUnlock()

				if s.processStats != nil && id.Row-1 < len(s.processStats.TopCPU) {
					proc := s.processStats.TopCPU[id.Row-1]
					switch id.Col {
					case 0:
						label.SetText(fmt.Sprintf("%d", proc.PID))
					case 1:
						label.SetText(proc.Name)
					case 2:
						label.SetText(proc.Username)
					case 3:
						label.SetText(fmt.Sprintf("%.1f", proc.CPUPercent))
					case 4:
						label.SetText(fmt.Sprintf("%d", proc.MemoryMB))
					}
				}
			}
		},
	)

	// Set column widths
	tab.table.SetColumnWidth(0, 70)  // PID
	tab.table.SetColumnWidth(1, 150) // Name
	tab.table.SetColumnWidth(2, 100) // User
	tab.table.SetColumnWidth(3, 80)  // CPU %
	tab.table.SetColumnWidth(4, 120) // Memory MB

	// Main content
	mainContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		tab.table,
	)

	// Wrap in scroll
	scroll := container.NewScroll(mainContent)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	return scroll
}

// killProcessWithConfirm shows a confirmation dialog and kills the process
func (tab *ProcessesTab) killProcessWithConfirm(pid int32, name string) {
	message := fmt.Sprintf("Are you sure you want to terminate process '%s' (PID: %d)?", name, pid)
	confirm := dialog.NewConfirm("Terminate Process", message, func(confirmed bool) {
		if confirmed {
			// Try graceful termination first
			if proc, err := os.FindProcess(int(pid)); err == nil {
				proc.Signal(syscall.SIGTERM)
			}
		}
	}, tab.window)
	confirm.Show()
}
