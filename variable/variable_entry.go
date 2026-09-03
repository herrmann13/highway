package variable

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type VariableEntry struct {
	widget.Entry
	onAddVariable func(*VariableEntry)
}

func NewVariableEntry(multiline, password bool, onAddVariable func(*VariableEntry)) *VariableEntry {
	entry := &VariableEntry{
		Entry: widget.Entry{
			MultiLine: multiline,
			Password:  password,
			Wrapping:  fyne.TextWrap(fyne.TextTruncateClip),
		},
		onAddVariable: onAddVariable,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *VariableEntry) CreateRenderer() fyne.WidgetRenderer {
	renderer := e.Entry.CreateRenderer()
	e.ExtendBaseWidget(e)
	return renderer
}

func (e *VariableEntry) TappedSecondary(event *fyne.PointEvent) {
	if e.onAddVariable == nil {
		e.Entry.TappedSecondary(event)
		return
	}

	canvas := fyne.CurrentApp().Driver().CanvasForObject(e)
	if canvas == nil {
		return
	}
	item := fyne.NewMenuItem("Adicionar variável", func() { e.onAddVariable(e) })
	pop := widget.NewPopUpMenu(fyne.NewMenu("", item), canvas)
	position := fyne.CurrentApp().Driver().AbsolutePositionForObject(e).Add(event.Position)
	pop.ShowAtPosition(position)
}

func ShowVariablePicker(w fyne.Window, entry *VariableEntry, variables [][2]string) {
	names := make([]string, 0, len(variables))
	seen := make(map[string]bool, len(variables))
	for _, variable := range variables {
		name := strings.TrimSpace(variable[0])
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		dialog.ShowInformation("Variáveis", "Esta collection não possui variáveis configuradas.", w)
		return
	}
	sort.Strings(names)

	selector := widget.NewSelect(names, nil)
	selector.SetSelected(names[0])
	d := dialog.NewForm(
		"Adicionar variável",
		"Inserir",
		"Cancelar",
		[]*widget.FormItem{widget.NewFormItem("Variável", selector)},
		func(ok bool) {
			if ok {
				entry.InsertVariable(selector.Selected)
			}
		},
		w,
	)
	d.Resize(fyne.NewSize(440, 180))
	d.Show()
}

func (e *VariableEntry) InsertVariable(name string) {
	text, offset := InsertVariablePlaceholder(e.Text, e.CursorTextOffset(), name)
	e.SetText(text)
	e.CursorRow, e.CursorColumn = CursorPositionAtOffset(text, offset)
	e.Refresh()
}

func InsertVariablePlaceholder(text string, offset int, name string) (string, int) {
	runes := []rune(text)
	if offset < 0 {
		offset = 0
	}
	if offset > len(runes) {
		offset = len(runes)
	}
	placeholder := "{{" + name + "}}"
	result := string(runes[:offset]) + placeholder + string(runes[offset:])
	return result, offset + len([]rune(placeholder))
}

func CursorPositionAtOffset(text string, offset int) (int, int) {
	row, column := 0, 0
	for i, r := range []rune(text) {
		if i == offset {
			break
		}
		if r == '\n' {
			row++
			column = 0
		} else {
			column++
		}
	}
	return row, column
}
