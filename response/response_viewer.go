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
