package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SimpleLineChart renders a simple line chart
type SimpleLineChart struct {
	widget.BaseWidget
	data       []*HistoryPoint
	title      string
	height     float32
	maxValue   float64
	yAxisLabel string
}

// NewSimpleLineChart creates a new line chart
func NewSimpleLineChart(title string, height float32) *SimpleLineChart {
	chart := &SimpleLineChart{
		title:  title,
		height: height,
		data:   make([]*HistoryPoint, 0, 60),
	}
	chart.ExtendBaseWidget(chart)
	return chart
}

// SetData updates the chart data
func (c *SimpleLineChart) SetData(data []*HistoryPoint, max float64) {
	c.data = data
	c.maxValue = max
	c.Refresh()
}

// CreateRenderer implements the widget interface
func (c *SimpleLineChart) CreateRenderer() fyne.WidgetRenderer {
	return &lineChartRenderer{chart: c}
}

// lineChartRenderer handles rendering of the chart
type lineChartRenderer struct {
	chart      *SimpleLineChart
	background *canvas.Rectangle
}

// Layout lays out the chart
func (r *lineChartRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
}

// MinSize returns the minimum size
func (r *lineChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, r.chart.height)
}

// Refresh redraws the chart
func (r *lineChartRenderer) Refresh() {
	r.redraw()
}

// Objects returns all objects in this renderer
func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background}
}

// Destroy cleans up resources
func (r *lineChartRenderer) Destroy() {}

// redraw redraws the chart
func (r *lineChartRenderer) redraw() {
	// This is a placeholder. In a real implementation,
	// this would render the actual chart canvas.
	// For now, we'll use a simple label-based approach
}

// SimpleChartDisplay creates a text-based chart display
func CreateChartDisplay(title string, data []*HistoryPoint, maxValue float64, height int) *fyne.Container {
	titleLabel := canvas.NewText(title, color.White)
	titleLabel.TextSize = 16

	// Create simple text representation
	chartText := widget.NewLabel(formatChartData(data, maxValue, height))

	return container.NewVBox(titleLabel, chartText)
}

// formatChartData creates a text representation of chart data
func formatChartData(data []*HistoryPoint, maxValue float64, height int) string {
	if len(data) == 0 {
		return "No data available"
	}

	// Create ASCII chart representation
	chart := ""

	// Determine scale
	if maxValue <= 0 {
		maxValue = 100
	}

	// Create rows for each percentage level
	levels := []float64{100, 75, 50, 25, 0}

	for _, level := range levels {
		row := fmt.Sprintf("%3.0f%% | ", level)

		// Add data points
		for i, point := range data {
			if i%max(1, len(data)/40) == 0 {
				//Map value to displayed character
				if point.Value >= level {
					row += "█"
				} else if point.Value >= level-25 {
					row += "▄"
				} else {
					row += " "
				}
			}
		}

		chart += row + "\n"
	}

	// Add x-axis
	chart += "      " + "└" + string([]rune{'─'}) + "┘\n"
	chart += fmt.Sprintf("      0s                                                     60s\n")
	chart += fmt.Sprintf("Max: %.1f%%", maxValue)

	return chart
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
