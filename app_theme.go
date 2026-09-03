package main

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/fonts/JetBrainsMono-Regular.ttf
var jetBrainsMonoRegular []byte

//go:embed assets/fonts/JetBrainsMono-Bold.ttf
var jetBrainsMonoBold []byte

//go:embed assets/fonts/JetBrainsMono-Italic.ttf
var jetBrainsMonoItalic []byte

//go:embed assets/fonts/JetBrainsMono-BoldItalic.ttf
var jetBrainsMonoBoldItalic []byte

var (
	jetBrainsMonoRegularResource    = fyne.NewStaticResource("JetBrainsMono-Regular.ttf", jetBrainsMonoRegular)
	jetBrainsMonoBoldResource       = fyne.NewStaticResource("JetBrainsMono-Bold.ttf", jetBrainsMonoBold)
	jetBrainsMonoItalicResource     = fyne.NewStaticResource("JetBrainsMono-Italic.ttf", jetBrainsMonoItalic)
	jetBrainsMonoBoldItalicResource = fyne.NewStaticResource("JetBrainsMono-BoldItalic.ttf", jetBrainsMonoBoldItalic)
)

type highwayTheme struct {
	base fyne.Theme
}

func newHighwayTheme() fyne.Theme {
	return &highwayTheme{base: theme.DarkTheme()}
}

func (t *highwayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *highwayTheme) Font(style fyne.TextStyle) fyne.Resource {
	switch {
	case style.Bold && style.Italic:
		return jetBrainsMonoBoldItalicResource
	case style.Bold:
		return jetBrainsMonoBoldResource
	case style.Italic:
		return jetBrainsMonoItalicResource
	default:
		return jetBrainsMonoRegularResource
	}
}

func (t *highwayTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *highwayTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
