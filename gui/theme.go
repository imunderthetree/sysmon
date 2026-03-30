package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme implements fyne.Theme with custom colors
type CustomTheme struct {
	mode ThemeMode
}

// NewCustomTheme creates a new custom theme
func NewCustomTheme(mode ThemeMode) fyne.Theme {
	return &CustomTheme{mode: mode}
}

// Color returns the color for the specified name
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
		}
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	case theme.ColorNameButton:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		}
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNameDisabledButton:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 40, G: 40, B: 40, A: 255}
		}
		return color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	case theme.ColorNameForeground:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 120, G: 120, B: 120, A: 255}
	case theme.ColorNamePlaceHolder:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
		}
		return color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	case theme.ColorNameHover:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
		}
		return color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 100}
	case theme.ColorNameSeparator:
		if t.mode == ThemeDark {
			return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		}
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the default font (delegates to default theme)
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the requested icon (delegates to default theme)
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the requested size (delegates to default theme)
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
