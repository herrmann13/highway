package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type kvPair struct {
	key   *variableEntry
	value *variableEntry
	row   *fyne.Container
}

type responseResult struct {
	statusCode int
	body       string
	headers    http.Header
	duration   time.Duration
}

const maxResponseBytes = 50 * 1024 * 1024

type authConfig struct {
	AuthType              string `json:"authType"`
	Token                 string `json:"token"`
	BasicUser             string `json:"basicUser"`
	BasicPass             string `json:"basicPass"`
	APIKeyName            string `json:"apiKeyName"`
	APIKeyValue           string `json:"apiKeyValue"`
	APIKeyLocation        string `json:"apiKeyLocation"`
	OAuth1ConsumerKey     string `json:"oauth1ConsumerKey"`
	OAuth1ConsumerSecret  string `json:"oauth1ConsumerSecret"`
	OAuth1AccessToken     string `json:"oauth1AccessToken"`
	OAuth1TokenSecret     string `json:"oauth1TokenSecret"`
	OAuth1SignatureMethod string `json:"oauth1SignatureMethod"`
	GrantType             string `json:"grantType"`
	TokenURL              string `json:"tokenURL"`
	ClientID              string `json:"clientID"`
	ClientSecret          string `json:"clientSecret"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	Scope                 string `json:"scope"`
}

type renameLabel struct {
	widget.Label
	onTapped       func()
	onDoubleTapped func()
}

func newRenameLabel() *renameLabel {
	label := &renameLabel{}
	label.ExtendBaseWidget(label)
	return label
}

func (l *renameLabel) CreateRenderer() fyne.WidgetRenderer {
	renderer := l.Label.CreateRenderer()
	l.ExtendBaseWidget(l)
	return renderer
}

func (l *renameLabel) Tapped(*fyne.PointEvent) {
	if l.onTapped != nil {
		l.onTapped()
	}
}

func (l *renameLabel) DoubleTapped(*fyne.PointEvent) {
	if l.onDoubleTapped != nil {
		l.onDoubleTapped()
	}
}

func showRenameDialog(w fyne.Window, title, current string, validate func(string) error, onSave func(string)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(current)
	nameEntry.Validator = func(name string) error {
		return validate(strings.TrimSpace(name))
	}

	d := dialog.NewForm(
		title,
		"Salvar",
		"Cancelar",
		[]*widget.FormItem{widget.NewFormItem("Nome", nameEntry)},
		func(ok bool) {
			if ok {
				onSave(strings.TrimSpace(nameEntry.Text))
			}
		},
		w,
	)
	d.Resize(fyne.NewSize(520, 180))
	d.Show()
}

func main() {
	pendingImport, handled, err := importFileFromArgs(os.Args[1:])
	if handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		runHighway(pendingImport)
		return
	}
	pendingImport, handled, err = importCommandFromArgs(os.Args[1:], os.Stdin)
	if handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		delivered, err := sendImportCommand(pendingImport)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro ao enviar importação para o Highway:", err)
			return
		}
		if delivered {
			return
		}
		if err := launchHighwayWithImport(pendingImport); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	runHighway(pendingImport)
}

func runHighway(pendingImport string) {
	a := app.NewWithID("com.herrmann.highway")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Highway")

	collections, err := loadCollections()
	if err != nil {
		collections = []*collection{}
	}

	tabs := container.NewDocTabs()
	tabIndex := map[*container.TabItem]*requestTab{}

	var tree *widget.Tree
	var selectedCollection string

	sync := func(rt *requestTab) {
		newName := rt.name
		if rt.collectionName == "" {
			rt.name = newName
			if rt.item != nil {
				rt.item.Text = newName
			}
			return
		}
		if err := upsertRequest(collections, rt.collectionName, rt.name, rt.editor.toData(newName)); err != nil {
			dialog.ShowInformation("Erro", err.Error(), w)
			return
		}
		rt.name = newName
		if rt.item != nil {
			rt.item.Text = newName
		}
		tree.Refresh()
	}

	var renameRequest func(*requestTab)
	variablesForCollection := func(colName string) [][2]string {
		for _, c := range collections {
			if c.Name == colName {
				return c.Variables
			}
		}
		return nil
	}

	addTab := func(rd *requestData, collectionName string) *container.TabItem {
		rt := newRequestTab(w, rd, collectionName, sync, variablesForCollection, func(rt *requestTab) {
			renameRequest(rt)
		})
		item := container.NewTabItem(rt.name, rt.content)
		rt.item = item
		tabIndex[item] = rt
		return item
	}

	openTab := func(rd *requestData, collectionName string) {
		item := addTab(rd, collectionName)
		tabs.Append(item)
		tabs.Select(item)
	}

	createRequestInCollection := func(colName string) {
		for _, c := range collections {
			if c.Name != colName {
				continue
			}

			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("Ex.: Listar usuários")
			nameEntry.Validator = func(name string) error {
				name = strings.TrimSpace(name)
				if name == "" {
					return fmt.Errorf("informe o nome da requisição")
				}
				if requestNameExists(c, name, -1) {
					return fmt.Errorf("já existe uma requisição com o nome %q", name)
				}
				return nil
			}

			d := dialog.NewForm(
				"Nova requisição",
				"Criar",
				"Cancelar",
				[]*widget.FormItem{widget.NewFormItem("Nome da requisição", nameEntry)},
				func(ok bool) {
					if !ok {
						return
					}
					name := strings.TrimSpace(nameEntry.Text)
					rd := requestData{Name: name, Type: requestTypeHTTP}
					c.Requests = append(c.Requests, rd)
					if err := saveCollection(c); err != nil {
						dialog.ShowInformation("Erro", err.Error(), w)
						return
					}
					openTab(&rd, colName)
					tree.Refresh()
				},
				w,
			)
			d.Resize(fyne.NewSize(520, 180))
			d.Show()
			return
		}
	}

	renameRequest = func(rt *requestTab) {
		oldName := rt.name
		showRenameDialog(w, "Renomear requisição", oldName, func(name string) error {
			if name == "" {
				return fmt.Errorf("informe o nome da requisição")
			}
			if rt.collectionName == "" {
				return nil
			}
			for _, c := range collections {
				if c.Name == rt.collectionName && requestNameExists(c, name, -1) && name != oldName {
					return fmt.Errorf("já existe uma requisição com o nome %q", name)
				}
			}
			return nil
		}, func(name string) {
			if name == oldName {
				return
			}
			if rt.collectionName != "" {
				for _, c := range collections {
					if c.Name != rt.collectionName {
						continue
					}
					for i := range c.Requests {
						if c.Requests[i].Name != oldName {
							continue
						}
						c.Requests[i].Name = name
						if err := saveCollection(c); err != nil {
							c.Requests[i].Name = oldName
							dialog.ShowInformation("Erro", err.Error(), w)
							return
						}
						break
					}
					break
				}
			}
			for _, openTab := range tabIndex {
				if openTab.collectionName != rt.collectionName || openTab.name != oldName {
					continue
				}
				openTab.name = name
				openTab.nameLabel.SetText(name)
				openTab.item.Text = name
			}
			tabs.Refresh()
			tree.Refresh()
		})
	}

	renameCollectionFlow := func(oldName string) {
		var target *collection
		for _, c := range collections {
			if c.Name == oldName {
				target = c
				break
			}
		}
		if target == nil {
			return
		}

		showRenameDialog(w, "Renomear coleção", oldName, func(name string) error {
			if name == "" {
				return fmt.Errorf("informe o nome da coleção")
			}
			for _, c := range collections {
				if c != target && c.Name == name {
					return fmt.Errorf("já existe uma coleção com o nome %q", name)
				}
			}
			return nil
		}, func(name string) {
			if err := renameCollection(target, name); err != nil {
				dialog.ShowInformation("Erro", err.Error(), w)
				return
			}
			for _, rt := range tabIndex {
				if rt.collectionName == oldName {
					rt.collectionName = name
				}
			}
			tree.Refresh()
		})
	}

	deleteRequestFlow := func(colName, reqName string) {
		dialog.ShowConfirm("Excluir requisição", "Excluir a requisição \""+reqName+"\"?", func(ok bool) {
			if !ok {
				return
			}
			for _, c := range collections {
				if c.Name != colName {
					continue
				}
				for i, request := range c.Requests {
					if request.Name != reqName {
						continue
					}
					c.Requests = append(c.Requests[:i], c.Requests[i+1:]...)
					if err := saveCollection(c); err != nil {
						c.Requests = append(c.Requests, request)
						copy(c.Requests[i+1:], c.Requests[i:len(c.Requests)-1])
						c.Requests[i] = request
						dialog.ShowInformation("Erro", err.Error(), w)
						return
					}
					for item, rt := range tabIndex {
						if rt.collectionName == colName && rt.name == reqName {
							tabs.Remove(item)
							delete(tabIndex, item)
						}
					}
					tree.Refresh()
					return
				}
			}
		}, w)
	}

	deleteCollectionFlow := func(colName string) {
		dialog.ShowConfirm("Excluir", "Excluir a coleção \""+colName+"\"?", func(ok bool) {
			if !ok {
				return
			}
			if err := deleteCollection(colName); err != nil {
				dialog.ShowInformation("Erro", err.Error(), w)
				return
			}
			for i, c := range collections {
				if c.Name == colName {
					collections = append(collections[:i], collections[i+1:]...)
					break
				}
			}
			for _, rt := range tabIndex {
				if rt.collectionName == colName {
					rt.collectionName = ""
				}
			}
			tree.Refresh()
		}, w)
	}

	showCollectionVariables := func(colName string) {
		var target *collection
		for _, c := range collections {
			if c.Name == colName {
				target = c
				break
			}
		}
		if target == nil {
			return
		}

		var pairs *[]kvPair
		section, pairs := keyValueSection("Adicionar variável", "base_url", "https://api.exemplo.com", target.Variables, func() {
			variables := snapshotPairs(*pairs)
			if _, err := variableValues(variables); err != nil {
				dialog.ShowInformation("Variáveis", err.Error(), w)
				return
			}
			previous := target.Variables
			target.Variables = variables
			if err := saveCollection(target); err != nil {
				target.Variables = previous
				dialog.ShowInformation("Erro", err.Error(), w)
			}
		}, func() *variableEntry { return newVariableEntry(false, false, nil) })

		d := dialog.NewCustom("Variáveis: "+target.Name, "Fechar", section, w)
		d.Resize(fyne.NewSize(680, 420))
		d.Show()
	}

	showCollectionMenu := func(btn *widget.Button, colName string) {
		c := fyne.CurrentApp().Driver().CanvasForObject(btn)
		if c == nil {
			return
		}
		newItem := fyne.NewMenuItem("Nova requisição", func() {
			createRequestInCollection(colName)
		})
		variablesItem := fyne.NewMenuItem("Variáveis", func() {
			showCollectionVariables(colName)
		})
		deleteItem := fyne.NewMenuItem("Excluir coleção", func() {
			deleteCollectionFlow(colName)
		})
		pop := widget.NewPopUpMenu(fyne.NewMenu("", newItem, variablesItem, deleteItem), c)
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(btn)
		pop.ShowAtPosition(pos.Add(fyne.NewPos(0, btn.Size().Height)))
	}

	tree = widget.NewTree(
		func(uid string) []string { return treeChildUIDs(uid, collections) },
		func(uid string) bool { return uid == "" || strings.HasPrefix(uid, "col:") },
		func(branch bool) fyne.CanvasObject {
			if branch {
				return container.NewHBox(
					newRenameLabel(),
					layout.NewSpacer(),
					widget.NewButton("⋯", nil),
				)
			}
			return container.NewHBox(
				widget.NewIcon(requestTypeIcon(requestTypeHTTP)),
				newRenameLabel(),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.CancelIcon(), nil),
			)
		},
		func(uid string, branch bool, node fyne.CanvasObject) {
			if branch {
				box, ok := node.(*fyne.Container)
				if !ok || len(box.Objects) < 3 {
					return
				}
				label := box.Objects[0].(*renameLabel)
				btn := box.Objects[2].(*widget.Button)
				c, _, ok := collectionForUID(collections, uid)
				if !ok {
					return
				}
				label.SetText(c.Name)
				label.onTapped = func() { tree.Select(uid) }
				label.onDoubleTapped = func() { renameCollectionFlow(c.Name) }
				btn.OnTapped = func() { showCollectionMenu(btn, c.Name) }
				return
			}
			box, ok := node.(*fyne.Container)
			if !ok || len(box.Objects) < 4 {
				return
			}
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*renameLabel)
			deleteButton := box.Objects[3].(*widget.Button)
			c, request, ok := requestForUID(collections, uid)
			if !ok {
				return
			}
			label.SetText(request.Name)
			icon.SetResource(requestTypeIcon(request.Type))
			label.onTapped = func() { tree.Select(uid) }
			label.onDoubleTapped = func() {
				for _, rt := range tabIndex {
					if rt.collectionName == c.Name && rt.name == request.Name {
						renameRequest(rt)
						return
					}
				}
				rd := *request
				openTab(&rd, c.Name)
				renameRequest(tabIndex[tabs.Selected()])
			}
			deleteButton.OnTapped = func() {
				deleteRequestFlow(c.Name, request.Name)
			}
		},
	)
	tree.OpenAllBranches()

	tree.OnSelected = func(uid string) {
		if c, _, ok := collectionForUID(collections, uid); ok {
			selectedCollection = c.Name
			return
		}
		c, request, ok := requestForUID(collections, uid)
		if !ok {
			return
		}
		selectedCollection = c.Name
		rd := *request
		openTab(&rd, c.Name)
	}

	tabs.CreateTab = func() *container.TabItem { return addTab(nil, "") }
	tabs.OnClosed = func(item *container.TabItem) {
		delete(tabIndex, item)
	}

	newCollectionButton := widget.NewButton("Nova Coleção", func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Ex.: API de usuários")
		nameEntry.Validator = func(name string) error {
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("informe o nome da coleção")
			}
			for _, c := range collections {
				if c.Name == name {
					return fmt.Errorf("já existe uma coleção com o nome %q", name)
				}
			}
			return nil
		}

		d := dialog.NewForm(
			"Nova coleção",
			"Criar",
			"Cancelar",
			[]*widget.FormItem{widget.NewFormItem("Nome da coleção", nameEntry)},
			func(ok bool) {
				if !ok {
					return
				}
				c := &collection{Name: strings.TrimSpace(nameEntry.Text)}
				if err := saveCollection(c); err != nil {
					dialog.ShowInformation("Erro", err.Error(), w)
					return
				}
				collections = append(collections, c)
				tree.Refresh()
			},
			w,
		)
		d.Resize(fyne.NewSize(520, 180))
		d.Show()
	})

	var showCurlImportDialog func(string)
	showCurlImportDialog = func(command string) {
		if len(collections) == 0 {
			dialog.ShowInformation("Importação", "Crie uma coleção antes de importar uma requisição.", w)
			return
		}

		collectionNames := make([]string, 0, len(collections))
		for _, c := range collections {
			collectionNames = append(collectionNames, c.Name)
		}
		destination := widget.NewSelect(collectionNames, nil)
		if selectedCollection != "" {
			destination.SetSelected(selectedCollection)
		} else {
			destination.SetSelected(collectionNames[0])
		}

		curlEntry := widget.NewMultiLineEntry()
		curlEntry.SetMinRowsVisible(10)
		curlEntry.SetPlaceHolder("curl https://api.exemplo.com/usuarios -H 'Accept: application/json'")
		if command == "" {
			command = w.Clipboard().Content()
		}
		curlEntry.SetText(command)

		d := dialog.NewForm(
			"Importar cURL",
			"Importar",
			"Cancelar",
			[]*widget.FormItem{
				widget.NewFormItem("Collection", destination),
				widget.NewFormItem("Comando cURL", curlEntry),
			},
			func(ok bool) {
				if !ok {
					return
				}
				rd, err := parseCurl(curlEntry.Text)
				if err != nil {
					dialog.ShowInformation("cURL inválido", err.Error(), w)
					return
				}
				for _, c := range collections {
					if c.Name != destination.Selected {
						continue
					}
					rd.Name = uniqueRequestName(c, "Importação cURL")
					c.Requests = append(c.Requests, rd)
					if err := saveCollection(c); err != nil {
						c.Requests = c.Requests[:len(c.Requests)-1]
						dialog.ShowInformation("Erro", err.Error(), w)
						return
					}
					selectedCollection = c.Name
					openTab(&rd, c.Name)
					tree.Refresh()
					return
				}
				dialog.ShowInformation("Importação", "Selecione uma collection válida para importar.", w)
			},
			w,
		)
		d.Resize(fyne.NewSize(780, 460))
		d.Show()
	}

	var importButton *widget.Button
	importButton = widget.NewButton("Importação", func() {
		c := fyne.CurrentApp().Driver().CanvasForObject(importButton)
		if c == nil {
			return
		}
		item := fyne.NewMenuItem("cURL", func() { showCurlImportDialog("") })
		pop := widget.NewPopUpMenu(fyne.NewMenu("", item), c)
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(importButton)
		pop.ShowAtPosition(pos.Add(fyne.NewPos(0, importButton.Size().Height)))
	})

	var monitorCurl atomic.Bool
	monitorCurl.Store(a.Preferences().BoolWithFallback("detect-curl", false))
	monitorCheck := widget.NewCheck("Detectar cURLs", func(enabled bool) {
		monitorCurl.Store(enabled)
		a.Preferences().SetBool("detect-curl", enabled)
	})
	monitorCheck.SetChecked(monitorCurl.Load())

	var curlPromptOpen atomic.Bool
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		detector := &curlClipboardDetector{}
		for range ticker.C {
			if !monitorCurl.Load() {
				continue
			}
			command, detected := detector.detect(w.Clipboard().Content())
			if !detected || !curlPromptOpen.CompareAndSwap(false, true) {
				continue
			}
			fyne.Do(func() {
				if !monitorCurl.Load() {
					curlPromptOpen.Store(false)
					return
				}
				dialog.ShowConfirm("cURL detectado", "Abrir esta requisição no Highway?", func(open bool) {
					curlPromptOpen.Store(false)
					if open {
						showCurlImportDialog(command)
					}
				}, w)
			})
		}
	}()

	actionBar := container.NewVBox(container.NewHBox(newCollectionButton, importButton), monitorCheck)
	sidebar := container.NewBorder(actionBar, nil, nil, nil, container.NewScroll(tree))

	split := container.NewHSplit(sidebar, tabs)
	split.SetOffset(0.22)

	openTab(nil, "")

	w.SetContent(split)
	w.Resize(fyne.NewSize(1200, 750))
	closeImportServer, err := startImportServer(func(command string) {
		fyne.Do(func() {
			w.Show()
			w.RequestFocus()
			showCurlImportDialog(command)
		})
	})
	if err == nil {
		defer closeImportServer()
	}
	if pendingImport != "" {
		fyne.Do(func() { showCurlImportDialog(pendingImport) })
	}
	w.ShowAndRun()
}

func treeChildUIDs(uid string, collections []*collection) []string {
	if uid == "" {
		ids := make([]string, 0, len(collections))
		for i := range collections {
			ids = append(ids, collectionUID(i))
		}
		return ids
	}
	if c, collectionIndex, ok := collectionForUID(collections, uid); ok {
		ids := make([]string, 0, len(c.Requests))
		for i := range c.Requests {
			ids = append(ids, requestUID(collectionIndex, i))
		}
		return ids
	}
	return nil
}

func collectionUID(index int) string {
	return "col:" + strconv.Itoa(index)
}

func requestUID(collectionIndex, requestIndex int) string {
	return "req:" + strconv.Itoa(collectionIndex) + ":" + strconv.Itoa(requestIndex)
}

func collectionForUID(collections []*collection, uid string) (*collection, int, bool) {
	if !strings.HasPrefix(uid, "col:") {
		return nil, 0, false
	}
	index, err := strconv.Atoi(strings.TrimPrefix(uid, "col:"))
	if err != nil || index < 0 || index >= len(collections) {
		return nil, 0, false
	}
	return collections[index], index, true
}

func requestForUID(collections []*collection, uid string) (*collection, *requestData, bool) {
	if !strings.HasPrefix(uid, "req:") {
		return nil, nil, false
	}
	parts := strings.Split(strings.TrimPrefix(uid, "req:"), ":")
	if len(parts) != 2 {
		return nil, nil, false
	}
	collectionIndex, err := strconv.Atoi(parts[0])
	if err != nil || collectionIndex < 0 || collectionIndex >= len(collections) {
		return nil, nil, false
	}
	requestIndex, err := strconv.Atoi(parts[1])
	if err != nil || requestIndex < 0 || requestIndex >= len(collections[collectionIndex].Requests) {
		return nil, nil, false
	}
	return collections[collectionIndex], &collections[collectionIndex].Requests[requestIndex], true
}

func sectionPanel(title string, accent, bg color.Color, content fyne.CanvasObject) *fyne.Container {
	titleText := canvas.NewText(title, accent)
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.TextSize = theme.TextSize()

	header := container.NewVBox(
		container.NewPadded(titleText),
		widget.NewSeparator(),
	)

	return container.NewStack(
		canvas.NewRectangle(bg),
		container.NewBorder(header, nil, nil, nil, container.NewPadded(content)),
	)
}

func showError(statusText *canvas.Text, responseBody *responseViewer, responseHeaders *widget.Entry, err error) {
	statusText.Text = "Erro: " + err.Error()
	statusText.Color = errorColor()
	statusText.Refresh()
	responseBody.clear()
	responseHeaders.SetText("")
}

func statusColor(code int) color.Color {
	switch {
	case code >= 200 && code < 300:
		return color.RGBA{R: 0x2e, G: 0x7d, B: 0x32, A: 0xff}
	case code >= 300 && code < 400:
		return color.RGBA{R: 0xf9, G: 0xa8, B: 0x25, A: 0xff}
	case code >= 400:
		return color.RGBA{R: 0xc6, G: 0x28, B: 0x28, A: 0xff}
	default:
		return theme.ForegroundColor()
	}
}

func errorColor() color.Color {
	return color.RGBA{R: 0xc6, G: 0x28, B: 0x28, A: 0xff}
}

func snapshotPairs(pairs []kvPair) [][2]string {
	out := make([][2]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, [2]string{p.key.Text, p.value.Text})
	}
	return out
}

func keyValueSection(addLabel, keyPlaceholder, valuePlaceholder string, defaults [][2]string, onChange func(), newEntry func() *variableEntry) (*fyne.Container, *[]kvPair) {
	pairs := []kvPair{}
	list := container.NewVBox()

	keyEntry := newEntry()
	keyEntry.SetPlaceHolder(keyPlaceholder)
	valueEntry := newEntry()
	valueEntry.SetPlaceHolder(valuePlaceholder)

	addPair := func(k, v string) {
		pair := kvPair{
			key:   newEntry(),
			value: newEntry(),
		}
		pair.key.SetText(k)
		pair.value.SetText(v)
		pair.key.SetPlaceHolder(keyPlaceholder)
		pair.value.SetPlaceHolder(valuePlaceholder)

		if onChange != nil {
			pair.key.OnChanged = func(string) { onChange() }
			pair.value.OnChanged = func(string) { onChange() }
		}

		removeButton := widget.NewButton("Remover", func() {
			list.Remove(pair.row)
			for i := range pairs {
				if pairs[i].key == pair.key {
					pairs = append(pairs[:i], pairs[i+1:]...)
					break
				}
			}
			list.Refresh()
			if onChange != nil {
				onChange()
			}
		})

		pair.row = container.NewGridWithColumns(3, pair.key, pair.value, removeButton)
		list.Add(pair.row)
		pairs = append(pairs, pair)
		list.Refresh()
	}

	addButton := widget.NewButton(addLabel, func() {
		k := strings.TrimSpace(keyEntry.Text)
		if k == "" {
			return
		}
		addPair(k, valueEntry.Text)
		keyEntry.SetText("")
		valueEntry.SetText("")
		if onChange != nil {
			onChange()
		}
	})

	headerRow := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Chave", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Valor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
	)
	addRow := container.NewGridWithColumns(3, keyEntry, valueEntry, addButton)

	for _, d := range defaults {
		addPair(d[0], d[1])
	}

	return container.NewBorder(headerRow, addRow, nil, nil, container.NewScroll(list)), &pairs
}

func buildBody(bodyType, raw string, form, mp [][2]string) (io.Reader, string, error) {
	switch bodyType {
	case "x-www-form-urlencoded":
		v := url.Values{}
		for _, p := range form {
			if p[0] == "" {
				continue
			}
			v.Add(p[0], p[1])
		}
		return strings.NewReader(v.Encode()), "application/x-www-form-urlencoded", nil
	case "multipart/form-data":
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		for _, p := range mp {
			if p[0] == "" {
				continue
			}
			fw, err := mw.CreateFormField(p[0])
			if err != nil {
				return nil, "", fmt.Errorf("erro no campo multipart %s: %w", p[0], err)
			}
			if _, err := fw.Write([]byte(p[1])); err != nil {
				return nil, "", err
			}
		}
		if err := mw.Close(); err != nil {
			return nil, "", err
		}
		return &buf, mw.FormDataContentType(), nil
	default:
		return strings.NewReader(raw), "", nil
	}
}

func sendRequest(method, rawURL string, params, headers [][2]string, reqBody io.Reader, contentType string, cfg authConfig) (responseResult, error) {
	var result responseResult

	u, err := url.Parse(rawURL)
	if err != nil {
		return result, fmt.Errorf("url inválida: %w", err)
	}
	if u.Scheme == "" {
		return result, fmt.Errorf("url deve conter esquema (http:// ou https://)")
	}

	q := u.Query()
	for _, p := range params {
		if p[0] == "" {
			continue
		}
		q.Add(p[0], p[1])
	}
	u.RawQuery = q.Encode()

	var bodyBytes []byte
	if reqBody != nil {
		bodyBytes, err = io.ReadAll(reqBody)
		if err != nil {
			return result, fmt.Errorf("erro ao ler body: %w", err)
		}
	}

	newReq := func() (*http.Request, error) {
		var reader io.Reader
		if bodyBytes != nil {
			reader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequest(method, u.String(), reader)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar request: %w", err)
		}
		for _, h := range headers {
			if h[0] == "" {
				continue
			}
			req.Header.Set(h[0], h[1])
		}
		if contentType != "" && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
		return req, nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()

	var resp *http.Response
	if cfg.AuthType == "Digest Auth" {
		resp, err = doDigest(client, newReq, method, u, cfg)
	} else {
		req, reqErr := newReq()
		if reqErr != nil {
			return result, reqErr
		}
		if reqErr := applyAuth(req, cfg); reqErr != nil {
			return result, reqErr
		}
		resp, err = client.Do(req)
	}
	if err != nil {
		return result, fmt.Errorf("erro na request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return result, fmt.Errorf("erro ao ler resposta: %w", err)
	}
	if len(respBody) > maxResponseBytes {
		return result, fmt.Errorf("resposta excede o limite de %d MB", maxResponseBytes/(1024*1024))
	}

	result.statusCode = resp.StatusCode
	result.duration = time.Since(start)
	result.headers = resp.Header
	result.body = formatResponseBody(respBody)

	return result, nil
}

func formatResponseBody(body []byte) string {
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		return pretty.String()
	}
	return string(body)
}

func applyAuth(req *http.Request, cfg authConfig) error {
	switch cfg.AuthType {
	case "No Auth", "":
		return nil
	case "Bearer Token":
		if strings.TrimSpace(cfg.Token) == "" {
			return fmt.Errorf("token bearer vazio")
		}
		req.Header.Set("Authorization", "Bearer "+cfg.Token)
	case "Basic Auth":
		if strings.TrimSpace(cfg.BasicUser) == "" {
			return fmt.Errorf("usuário vazio")
		}
		token := base64.StdEncoding.EncodeToString([]byte(cfg.BasicUser + ":" + cfg.BasicPass))
		req.Header.Set("Authorization", "Basic "+token)
	case "API Key":
		if strings.TrimSpace(cfg.APIKeyName) == "" {
			return fmt.Errorf("nome da API Key vazio")
		}
		if strings.TrimSpace(cfg.APIKeyValue) == "" {
			return fmt.Errorf("valor da API Key vazio")
		}
		if cfg.APIKeyLocation == "query" {
			query := req.URL.Query()
			query.Set(cfg.APIKeyName, cfg.APIKeyValue)
			req.URL.RawQuery = query.Encode()
		} else {
			req.Header.Set(cfg.APIKeyName, cfg.APIKeyValue)
		}
	case "OAuth 1.0":
		header, err := buildOAuth1Header(req, cfg)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", header)
	case "OAuth 2.0":
		token, err := fetchToken(cfg)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	default:
		return fmt.Errorf("tipo de autenticação desconhecido: %s", cfg.AuthType)
	}
	return nil
}

func fetchToken(cfg authConfig) (string, error) {
	if strings.TrimSpace(cfg.TokenURL) == "" {
		return "", fmt.Errorf("token URL vazia")
	}

	form := url.Values{}
	form.Set("grant_type", cfg.GrantType)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	if cfg.GrantType == "password" {
		form.Set("username", cfg.Username)
		form.Set("password", cfg.Password)
	}
	if strings.TrimSpace(cfg.Scope) != "" {
		form.Set("scope", cfg.Scope)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(cfg.TokenURL, form)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta do token: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token endpoint retornou %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar token: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access_token não encontrado na resposta")
	}
	return tokenResp.AccessToken, nil
}

func formatHeaders(h http.Header) string {
	var sb strings.Builder
	for k, vals := range h {
		for _, v := range vals {
			sb.WriteString(k + ": " + v + "\n")
		}
	}
	return sb.String()
}

func doDigest(client *http.Client, newReq func() (*http.Request, error), method string, u *url.URL, cfg authConfig) (*http.Response, error) {
	req, err := newReq()
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}
	challenge := resp.Header.Get("WWW-Authenticate")
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(challenge)), "digest") {
		return resp, nil
	}
	resp.Body.Close()

	header, err := buildDigestHeader(method, u, challenge, cfg.BasicUser, cfg.BasicPass)
	if err != nil {
		return nil, err
	}

	req, err = newReq()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", header)
	return client.Do(req)
}

func buildDigestHeader(method string, u *url.URL, challenge, username, password string) (string, error) {
	params := parseDigestChallenge(challenge)
	realm := params["realm"]
	nonce := params["nonce"]
	qop := params["qop"]
	opaque := params["opaque"]

	if nonce == "" {
		return "", fmt.Errorf("digest: nonce ausente no challenge")
	}

	uri := u.RequestURI()
	ha1 := md5Hex(username + ":" + realm + ":" + password)
	ha2 := md5Hex(method + ":" + uri)

	fields := fmt.Sprintf(`username="%s", realm="%s", nonce="%s", uri="%s"`, username, realm, nonce, uri)

	var response string
	if strings.Contains(qop, "auth") {
		cnonce := randomHex(8)
		nc := "00000001"
		response = md5Hex(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":auth:" + ha2)
		fields += fmt.Sprintf(`, qop=auth, nc=%s, cnonce="%s"`, nc, cnonce)
	} else {
		response = md5Hex(ha1 + ":" + nonce + ":" + ha2)
	}

	if opaque != "" {
		fields += fmt.Sprintf(`, opaque="%s"`, opaque)
	}
	fields += fmt.Sprintf(`, response="%s"`, response)

	return "Digest " + fields, nil
}

func parseDigestChallenge(challenge string) map[string]string {
	out := map[string]string{}
	challenge = strings.TrimSpace(challenge)
	if i := strings.IndexByte(challenge, ' '); i >= 0 {
		challenge = challenge[i+1:]
	}
	for _, part := range splitComma(challenge) {
		part = strings.TrimSpace(part)
		eq := strings.IndexByte(part, '=')
		if eq < 0 {
			continue
		}
		k := strings.TrimSpace(part[:eq])
		v := strings.TrimSpace(part[eq+1:])
		v = strings.Trim(v, `"`)
		out[strings.ToLower(k)] = v
	}
	return out
}

func splitComma(s string) []string {
	var parts []string
	var sb strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			sb.WriteRune(r)
		case r == ',' && !inQuote:
			parts = append(parts, sb.String())
			sb.Reset()
		default:
			sb.WriteRune(r)
		}
	}
	if sb.Len() > 0 {
		parts = append(parts, sb.String())
	}
	return parts
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func buildOAuth1Header(req *http.Request, cfg authConfig) (string, error) {
	if strings.TrimSpace(cfg.OAuth1ConsumerKey) == "" {
		return "", fmt.Errorf("consumer key vazio")
	}
	sigMethod := cfg.OAuth1SignatureMethod
	if sigMethod == "" {
		sigMethod = "HMAC-SHA1"
	}

	oauth := map[string]string{
		"oauth_consumer_key":     cfg.OAuth1ConsumerKey,
		"oauth_nonce":            randomHex(16),
		"oauth_signature_method": sigMethod,
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_version":          "1.0",
	}
	if cfg.OAuth1AccessToken != "" {
		oauth["oauth_token"] = cfg.OAuth1AccessToken
	}

	params := map[string][]string{}
	for k, vs := range req.URL.Query() {
		params[k] = vs
	}
	for k, v := range oauth {
		params[k] = []string{v}
	}

	baseString := oauthBaseString(req, params)
	signingKey := percentEncode(cfg.OAuth1ConsumerSecret) + "&" + percentEncode(cfg.OAuth1TokenSecret)
	signature, err := oauthSign(sigMethod, signingKey, baseString)
	if err != nil {
		return "", err
	}
	oauth["oauth_signature"] = signature

	keys := make([]string, 0, len(oauth))
	for k := range oauth {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, percentEncode(k)+`="`+percentEncode(oauth[k])+`"`)
	}
	return "OAuth " + strings.Join(parts, ", "), nil
}

func oauthBaseString(req *http.Request, params map[string][]string) string {
	var pairs []string
	for k, vs := range params {
		for _, v := range vs {
			pairs = append(pairs, percentEncode(k)+"="+percentEncode(v))
		}
	}
	sort.Strings(pairs)
	normalized := strings.Join(pairs, "&")

	host := strings.ToLower(req.URL.Host)
	baseURI := req.URL.Scheme + "://" + host + req.URL.Path

	return strings.ToUpper(req.Method) + "&" + percentEncode(baseURI) + "&" + percentEncode(normalized)
}

func oauthSign(method, key, baseString string) (string, error) {
	switch method {
	case "PLAINTEXT":
		return key, nil
	case "HMAC-SHA256":
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(baseString))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
	case "HMAC-SHA1":
		mac := hmac.New(sha1.New, []byte(key))
		mac.Write([]byte(baseString))
		return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
	default:
		return "", fmt.Errorf("signature method desconhecido: %s", method)
	}
}

func percentEncode(s string) string {
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if strings.ContainsRune(unreserved, rune(b)) {
			sb.WriteByte(b)
		} else {
			fmt.Fprintf(&sb, "%%%02X", b)
		}
	}
	return sb.String()
}
