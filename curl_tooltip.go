package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type curlHistoryLabel struct {
	*widget.Label

	mu          sync.Mutex
	tooltipText string
	tooltip     *widget.PopUp
	timer       *time.Timer
	hovered     bool
}

func newCurlHistoryLabel() *curlHistoryLabel {
	label := &curlHistoryLabel{Label: widget.NewLabel("")}
	label.TextStyle = fyne.TextStyle{Monospace: true}
	label.Truncation = fyne.TextTruncateEllipsis
	return label
}

func (l *curlHistoryLabel) SetTooltip(text string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tooltipText = text
}

func (l *curlHistoryLabel) MouseIn(*desktop.MouseEvent) {
	l.mu.Lock()
	l.hovered = true
	if l.timer != nil {
		l.timer.Stop()
	}
	text := l.tooltipText
	l.timer = time.AfterFunc(450*time.Millisecond, func() {
		l.mu.Lock()
		if !l.hovered || text == "" {
			l.mu.Unlock()
			return
		}
		l.mu.Unlock()
		fyne.Do(func() { l.showTooltip(text) })
	})
	l.mu.Unlock()
}

func (l *curlHistoryLabel) MouseMoved(*desktop.MouseEvent) {}

func (l *curlHistoryLabel) MouseOut() {
	l.mu.Lock()
	l.hovered = false
	if l.timer != nil {
		l.timer.Stop()
	}
	popup := l.tooltip
	l.tooltip = nil
	l.mu.Unlock()
	if popup != nil {
		popup.Hide()
	}
}

func (l *curlHistoryLabel) showTooltip(text string) {
	l.mu.Lock()
	if !l.hovered || l.tooltip != nil {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	canvas := fyne.CurrentApp().Driver().CanvasForObject(l)
	if canvas == nil {
		return
	}
	content := widget.NewLabel(text)
	content.TextStyle = fyne.TextStyle{Monospace: true}
	content.Wrapping = fyne.TextWrapOff
	popup := widget.NewPopUp(content, canvas)
	position := fyne.CurrentApp().Driver().AbsolutePositionForObject(l)
	popup.ShowAtPosition(position.Add(fyne.NewPos(0, l.Size().Height+4)))

	l.mu.Lock()
	if l.hovered {
		l.tooltip = popup
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	popup.Hide()
}
