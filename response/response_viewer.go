package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ResponseViewer struct {
	List     *widget.List
	FullBody string
	lines    []string
}

type ResponseHeadersViewer struct {
	Content fyne.CanvasObject
	List    *widget.List
	lines   []string
}

func NewResponseHeadersViewer() *ResponseHeadersViewer {
	viewer := &ResponseHeadersViewer{}
	viewer.List = widget.NewList(
		func() int { return len(viewer.lines) },
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Truncation = fyne.TextTruncateClip
			return label
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(viewer.lines[id])
		},
	)
	viewer.Content = viewer.List
	return viewer
}

func (v *ResponseHeadersViewer) SetHeaders(headers string) {
	headers = strings.TrimSuffix(headers, "\n")
	if headers == "" {
		v.lines = nil
	} else {
		v.lines = strings.Split(headers, "\n")
	}
	v.List.Refresh()
	v.List.ScrollToTop()
}

func (v *ResponseHeadersViewer) Clear() {
	v.SetHeaders("")
}

func NewResponseViewer() *ResponseViewer {
	viewer := &ResponseViewer{}
	viewer.List = widget.NewList(
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

func (v *ResponseViewer) SetResponse(body string, lines []string) {
	v.FullBody = body
	v.lines = lines
	v.List.Refresh()
	v.List.ScrollToTop()
}

func ResponseLines(body string) []string {
	return strings.Split(body, "\n")
}

func (v *ResponseViewer) Clear() {
	v.FullBody = ""
	v.lines = nil
	v.List.Refresh()
}

func FormatResponseBody(body []byte) string {
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		return pretty.String()
	}
	return string(body)
}
