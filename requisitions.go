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
	"sort"
	"strconv"
	"strings"
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
	key   *widget.Entry
	value *widget.Entry
	row   *fyne.Container
}

type responseResult struct {
	statusCode int
	body       string
	headers    http.Header
	duration   time.Duration
}

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
	a := app.NewWithID("com.herrmann.requisitions")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Requisição HTTP")

	collections, err := loadCollections()
	if err != nil {
		collections = []*collection{}
	}

	tabs := container.NewDocTabs()
	tabIndex := map[*container.TabItem]*requestTab{}

	var tree *widget.Tree

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

	addTab := func(rd *requestData, collectionName string) *container.TabItem {
		rt := newRequestTab(w, rd, collectionName, sync, func(rt *requestTab) {
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
					rd := requestData{Name: name}
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

	showCollectionMenu := func(btn *widget.Button, colName string) {
		c := fyne.CurrentApp().Driver().CanvasForObject(btn)
		if c == nil {
			return
		}
		newItem := fyne.NewMenuItem("Nova requisição", func() {
			createRequestInCollection(colName)
		})
		deleteItem := fyne.NewMenuItem("Excluir coleção", func() {
			deleteCollectionFlow(colName)
		})
		pop := widget.NewPopUpMenu(fyne.NewMenu("", newItem, deleteItem), c)
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
			return newRenameLabel()
		},
		func(uid string, branch bool, node fyne.CanvasObject) {
			if branch {
				box, ok := node.(*fyne.Container)
				if !ok || len(box.Objects) < 3 {
					return
				}
				label := box.Objects[0].(*renameLabel)
				btn := box.Objects[2].(*widget.Button)
				colName := strings.TrimPrefix(uid, "col:")
				label.SetText(colName)
				label.onTapped = func() { tree.Select(uid) }
				label.onDoubleTapped = func() { renameCollectionFlow(colName) }
				btn.OnTapped = func() { showCollectionMenu(btn, colName) }
				return
			}
			label, ok := node.(*renameLabel)
			if !ok {
				return
			}
			rest := strings.TrimPrefix(uid, "req:")
			if idx := strings.LastIndex(rest, "/"); idx >= 0 {
				label.SetText(rest[idx+1:])
			} else {
				label.SetText(rest)
			}
			label.onTapped = func() { tree.Select(uid) }
			label.onDoubleTapped = func() {
				idx := strings.LastIndex(rest, "/")
				if idx < 0 {
					return
				}
				colName := rest[:idx]
				reqName := rest[idx+1:]
				for _, rt := range tabIndex {
					if rt.collectionName == colName && rt.name == reqName {
						renameRequest(rt)
						return
					}
				}
				for _, c := range collections {
					if c.Name != colName {
						continue
					}
					for _, r := range c.Requests {
						if r.Name != reqName {
							continue
						}
						rd := r
						openTab(&rd, colName)
						renameRequest(tabIndex[tabs.Selected()])
						return
					}
				}
			}
		},
	)
	tree.OpenAllBranches()

	tree.OnSelected = func(uid string) {
		if !strings.HasPrefix(uid, "req:") {
			return
		}
		rest := strings.TrimPrefix(uid, "req:")
		idx := strings.LastIndex(rest, "/")
		if idx < 0 {
			return
		}
		colName := rest[:idx]
		reqName := rest[idx+1:]
		for _, c := range collections {
			if c.Name != colName {
				continue
			}
			for _, r := range c.Requests {
				if r.Name == reqName {
					rd := r
					openTab(&rd, colName)
				}
			}
		}
	}

	tabs.CreateTab = func() *container.TabItem { return addTab(nil, "") }
	tabs.OnClosed = func(item *container.TabItem) {
		delete(tabIndex, item)
	}

	newCollectionButton := widget.NewButton("Nova Coleção", func() {
		d := dialog.NewEntryDialog("Nova coleção", "Nome da coleção:", func(name string) {
			if strings.TrimSpace(name) == "" {
				return
			}
			c := &collection{Name: name}
			if err := saveCollection(c); err != nil {
				dialog.ShowInformation("Erro", err.Error(), w)
				return
			}
			collections = append(collections, c)
			tree.Refresh()
		}, w)
		d.Show()
	})

	actionBar := container.NewHBox(newCollectionButton)
	sidebar := container.NewBorder(actionBar, nil, nil, nil, container.NewScroll(tree))

	split := container.NewHSplit(sidebar, tabs)
	split.SetOffset(0.22)

	openTab(nil, "")

	w.SetContent(split)
	w.Resize(fyne.NewSize(1200, 750))
	w.ShowAndRun()
}

func treeChildUIDs(uid string, collections []*collection) []string {
	if uid == "" {
		ids := make([]string, 0, len(collections))
		for _, c := range collections {
			ids = append(ids, "col:"+c.Name)
		}
		return ids
	}
	if strings.HasPrefix(uid, "col:") {
		name := strings.TrimPrefix(uid, "col:")
		for _, c := range collections {
			if c.Name != name {
				continue
			}
			ids := make([]string, 0, len(c.Requests))
			for _, r := range c.Requests {
				ids = append(ids, "req:"+name+"/"+r.Name)
			}
			return ids
		}
	}
	return nil
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

func showError(statusText *canvas.Text, responseBody, responseHeaders *widget.Entry, err error) {
	statusText.Text = "Erro: " + err.Error()
	statusText.Color = errorColor()
	statusText.Refresh()
	responseBody.SetText("")
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

func keyValueSection(addLabel, keyPlaceholder, valuePlaceholder string, defaults [][2]string, onChange func()) (*fyne.Container, *[]kvPair) {
	pairs := []kvPair{}
	list := container.NewVBox()

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder(keyPlaceholder)
	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder(valuePlaceholder)

	addPair := func(k, v string) {
		pair := kvPair{
			key:   widget.NewEntry(),
			value: widget.NewEntry(),
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	result.statusCode = resp.StatusCode
	result.duration = time.Since(start)
	result.headers = resp.Header
	result.body = string(respBody)

	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, respBody, "", "  "); err == nil {
			result.body = pretty.String()
		}
	}

	return result, nil
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
