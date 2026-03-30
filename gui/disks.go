package gui

import (
	"fmt"
	"image/color"
	"sysmon/internal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// DisksTab represents the disks view
func (s *AppState) newDisksTab() fyne.CanvasObject {
	// Title
	title := canvas.NewText("Disk Usage", color.White)
	title.TextSize = 24

	// Table
	table := widget.NewTable(
		func() (int, int) { return 11, 5 }, // 10 disks + 1 header, 5 columns
		func() fyne.CanvasObject {
			return container.NewVBox(widget.NewLabel("Cell"))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*fyne.Container).Objects[0].(*widget.Label)

			if id.Row == 0 {
				headers := []string{"Device", "Mount", "Usage %", "Used", "Total"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				s.mutex.RLock()
				defer s.mutex.RUnlock()

				if s.systemStats != nil && id.Row-1 < len(s.systemStats.Disk) {
					disk := s.systemStats.Disk[id.Row-1]
					switch id.Col {
					case 0:
						label.SetText(disk.Device)
					case 1:
						label.SetText(disk.Mountpoint)
					case 2:
						label.SetText(fmt.Sprintf("%.1f%%", disk.UsedPercent))
					case 3:
						label.SetText(internal.FormatBytes(disk.Used))
					case 4:
						label.SetText(internal.FormatBytes(disk.Total))
					}
				}
			}
		},
	)

	table.SetColumnWidth(0, 100)
	table.SetColumnWidth(1, 150)
	table.SetColumnWidth(2, 100)
	table.SetColumnWidth(3, 120)
	table.SetColumnWidth(4, 120)

	mainContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		table,
	)

	scroll := container.NewScroll(mainContent)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	return scroll
}
