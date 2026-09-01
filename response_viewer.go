package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	responsePageMaxBytes = 250 * 1024
	responsePageMaxLines = 5000
)

type responsePage struct {
	start, end            int
	startLine, finishLine int
}

type responseViewer struct {
	entry       *widget.Entry
	fullBody    string
	pages       []responsePage
	currentPage int
	pageLabel   *widget.Label
	previous    *widget.Button
	next        *widget.Button
	controls    *fyne.Container
}

func newResponseViewer() *responseViewer {
	viewer := &responseViewer{
		entry:     widget.NewMultiLineEntry(),
		pageLabel: widget.NewLabel(""),
	}
	viewer.entry.SetPlaceHolder("Resposta")
	viewer.entry.Disable()
	viewer.previous = widget.NewButton("Anterior", func() {
		if viewer.currentPage > 0 {
			viewer.currentPage--
			viewer.renderPage()
		}
	})
	viewer.next = widget.NewButton("Próxima", func() {
		if viewer.currentPage+1 < len(viewer.pages) {
			viewer.currentPage++
			viewer.renderPage()
		}
	})
	viewer.controls = container.NewHBox(viewer.previous, viewer.pageLabel, layout.NewSpacer(), viewer.next)
	viewer.clear()
	return viewer
}

func (v *responseViewer) setResponse(body string) {
	v.fullBody = body
	v.pages = responsePages(body)
	v.currentPage = 0
	v.renderPage()
}

func (v *responseViewer) clear() {
	v.fullBody = ""
	v.pages = nil
	v.currentPage = 0
	v.entry.SetText("")
	v.pageLabel.SetText("")
	v.controls.Hide()
}

func (v *responseViewer) renderPage() {
	if len(v.pages) == 0 {
		v.clear()
		return
	}
	page := v.pages[v.currentPage]
	v.entry.SetText(v.fullBody[page.start:page.end])
	if v.currentPage > 0 {
		v.previous.Enable()
	} else {
		v.previous.Disable()
	}
	if v.currentPage+1 < len(v.pages) {
		v.next.Enable()
	} else {
		v.next.Disable()
	}
	if len(v.pages) == 1 {
		v.pageLabel.SetText("")
		v.controls.Hide()
		return
	}
	v.pageLabel.SetText(fmt.Sprintf("Linhas %d-%d de %d", page.startLine, page.finishLine, lineCount(v.fullBody)))
	v.controls.Show()
}

func responsePages(body string) []responsePage {
	if body == "" {
		return nil
	}

	pages := make([]responsePage, 0, len(body)/responsePageMaxBytes+1)
	start, startLine := 0, 1
	for start < len(body) {
		limit := start + responsePageMaxBytes
		if limit > len(body) {
			limit = len(body)
		}
		end := limit
		if limit < len(body) {
			if newline := strings.LastIndex(body[start:limit], "\n"); newline >= 0 {
				end = start + newline + 1
			} else {
				for end > start && body[end]&0xc0 == 0x80 {
					end--
				}
			}
		}

		lines := strings.Count(body[start:end], "\n")
		if lines >= responsePageMaxLines {
			lineEnd, count := start, 0
			for lineEnd < end && count < responsePageMaxLines {
				next := strings.IndexByte(body[lineEnd:end], '\n')
				if next < 0 {
					break
				}
				lineEnd += next + 1
				count++
			}
			if lineEnd > start {
				end = lineEnd
				lines = count
			}
		}

		finishLine := startLine + lines
		if end == len(body) && body[end-1] != '\n' {
			finishLine++
		}
		pages = append(pages, responsePage{start: start, end: end, startLine: startLine, finishLine: finishLine - 1})
		start = end
		startLine = finishLine
	}
	return pages
}

func lineCount(body string) int {
	if body == "" {
		return 0
	}
	return strings.Count(body, "\n") + 1
}
