package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type responseViewer struct {
	list     *widget.List
	fullBody string
	lines    []string
}

func newResponseViewer() *responseViewer {
	viewer := &responseViewer{}
	viewer.list = widget.NewList(
		func() int { return len(viewer.lines) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Truncation = fyne.TextTruncateClip
			return label
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(fmt.Sprintf("%6d  %s", id+1, viewer.lines[id]))
		},
	)
	return viewer
}

func (v *responseViewer) setResponse(body string, lines []string) {
	v.fullBody = body
	v.lines = lines
	v.list.Refresh()
	v.list.ScrollToTop()
}

func responseLines(body string) []string {
	return strings.Split(body, "\n")
}

func (v *responseViewer) clear() {
	v.fullBody = ""
	v.lines = nil
	v.list.Refresh()
}
