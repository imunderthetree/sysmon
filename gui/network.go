package gui

import (
	"image/color"
	"sysmon/internal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NetworkTab represents the network view
func (s *AppState) newNetworkTab() fyne.CanvasObject {
	// Title
	title := canvas.NewText("Network Statistics", color.White)
	title.TextSize = 24

	// Labels
	statsLabel := widget.NewLabel("Loading network stats...")
	interfacesLabel := widget.NewLabel("Network Interfaces:")

	// Summary box
	summaryBox := container.NewVBox(
		canvas.NewText("Summary", color.White),
		statsLabel,
	)

	// Interfaces table
	table := widget.NewTable(
		func() (int, int) { return 9, 4 }, // 8 interfaces + 1 header, 4 columns
		func() fyne.CanvasObject {
			return container.NewVBox(widget.NewLabel("Cell"))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*fyne.Container).Objects[0].(*widget.Label)

			if id.Row == 0 {
				headers := []string{"Interface", "Sent", "Received", "Status"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
				}
			} else {
				s.mutex.RLock()
				defer s.mutex.RUnlock()

				if s.networkStats != nil && id.Row-1 < len(s.networkStats.Interfaces) {
					iface := s.networkStats.Interfaces[id.Row-1]
					switch id.Col {
					case 0:
						label.SetText(iface.Name)
					case 1:
						label.SetText(internal.FormatNetworkBytes(iface.BytesSent))
					case 2:
						label.SetText(internal.FormatNetworkBytes(iface.BytesRecv))
					case 3:
						status := "Down"
						if iface.IsUp {
							status = "Up"
						}
						label.SetText(status)
					}
				}
			}
		},
	)

	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 120)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 80)

	mainContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		summaryBox,
		widget.NewSeparator(),
		interfacesLabel,
		table,
	)

	scroll := container.NewScroll(mainContent)
	scroll.SetMinSize(fyne.NewSize(600, 400))

	return scroll
}
