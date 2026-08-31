package main

import (
	"fmt"
	"image/color"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type requestSnapshot struct {
	method      string
	url         string
	bodyType    string
	rawBody     string
	queryParams [][2]string
	headers     [][2]string
	form        [][2]string
	multipart   [][2]string
	auth        authConfig
}

type requestEditor struct {
	method           *widget.Select
	urlEntry         *widget.Entry
	bodyType         *widget.Select
	rawBody          *widget.Entry
	form             *[]kvPair
	multipart        *[]kvPair
	headers          *[]kvPair
	params           *[]kvPair
	formSection      *fyne.Container
	multipartSection *fyne.Container

	authorization         *widget.Select
	authorizationContent  *widget.Entry
	basicUserEntry        *widget.Entry
	basicPassEntry        *widget.Entry
	apiKeyNameEntry       *widget.Entry
	apiKeyValueEntry      *widget.Entry
	apiKeyLocation        *widget.Select
	oauth1ConsumerKey     *widget.Entry
	oauth1ConsumerSecret  *widget.Entry
	oauth1AccessToken     *widget.Entry
	oauth1TokenSecret     *widget.Entry
	oauth1SignatureMethod *widget.Select
	oauthGrantType        *widget.Select
	oauthTokenURL         *widget.Entry
	oauthClientID         *widget.Entry
	oauthClientSecret     *widget.Entry
	oauthUsername         *widget.Entry
	oauthPassword         *widget.Entry
	oauthScope            *widget.Entry

	basicForm  *widget.Form
	apiKeyForm *widget.Form
	oauth1Form *widget.Form
	oauthForm  *widget.Form

	tabs *container.AppTabs
}

func newRequestEditor(rd *requestData, onEdit func()) *requestEditor {
	e := &requestEditor{}

	e.urlEntry = widget.NewEntry()
	e.urlEntry.SetPlaceHolder("http://localhost:3000")

	e.rawBody = widget.NewMultiLineEntry()
	e.rawBody.SetText("{}")

	e.method = widget.NewSelect(
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, nil,
	)

	var defHeaders, defParams, defForm, defMultipart [][2]string
	if rd != nil {
		defHeaders = rd.Headers
		defParams = rd.Params
		defForm = rd.Form
		defMultipart = rd.Multipart
	} else {
		defHeaders = [][2]string{
			{"Content-Type", "application/json"},
			{"Accept", "application/json"},
		}
	}

	headersSection, headerPairs := keyValueSection("Adicionar Header", "Content-Type", "application/json", defHeaders, onEdit)
	paramsSection, paramPairs := keyValueSection("Adicionar Param", "chave", "valor", defParams, onEdit)
	formSection, formPairs := keyValueSection("Adicionar Campo", "chave", "valor", defForm, onEdit)
	multipartSection, multipartPairs := keyValueSection("Adicionar Campo", "chave", "valor", defMultipart, onEdit)

	e.headers = headerPairs
	e.params = paramPairs
	e.form = formPairs
	e.multipart = multipartPairs
	e.formSection = formSection
	e.multipartSection = multipartSection

	e.bodyType = widget.NewSelect(
		[]string{"raw", "x-www-form-urlencoded", "multipart/form-data"},
		nil,
	)

	e.authorizationContent = widget.NewMultiLineEntry()
	e.authorizationContent.SetMinRowsVisible(1)
	e.authorizationContent.SetPlaceHolder("Auth Content")

	e.basicUserEntry = widget.NewEntry()
	e.basicUserEntry.SetPlaceHolder("usuário")
	e.basicPassEntry = widget.NewPasswordEntry()
	e.basicPassEntry.SetPlaceHolder("senha")
	e.basicForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Usuário", Widget: e.basicUserEntry},
			{Text: "Senha", Widget: e.basicPassEntry},
		},
	}

	e.apiKeyNameEntry = widget.NewEntry()
	e.apiKeyNameEntry.SetPlaceHolder("X-API-Key")
	e.apiKeyValueEntry = widget.NewEntry()
	e.apiKeyValueEntry.SetPlaceHolder("valor")
	e.apiKeyLocation = widget.NewSelect([]string{"header", "query"}, nil)
	e.apiKeyForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Key", Widget: e.apiKeyNameEntry},
			{Text: "Value", Widget: e.apiKeyValueEntry},
			{Text: "Add to", Widget: e.apiKeyLocation},
		},
	}

	e.oauth1ConsumerKey = widget.NewEntry()
	e.oauth1ConsumerKey.SetPlaceHolder("consumer key")
	e.oauth1ConsumerSecret = widget.NewEntry()
	e.oauth1ConsumerSecret.SetPlaceHolder("consumer secret")
	e.oauth1AccessToken = widget.NewEntry()
	e.oauth1AccessToken.SetPlaceHolder("access token")
	e.oauth1TokenSecret = widget.NewEntry()
	e.oauth1TokenSecret.SetPlaceHolder("token secret")
	e.oauth1SignatureMethod = widget.NewSelect(
		[]string{"HMAC-SHA1", "HMAC-SHA256", "PLAINTEXT"}, nil,
	)
	e.oauth1Form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Consumer Key", Widget: e.oauth1ConsumerKey},
			{Text: "Consumer Secret", Widget: e.oauth1ConsumerSecret},
			{Text: "Access Token", Widget: e.oauth1AccessToken},
			{Text: "Token Secret", Widget: e.oauth1TokenSecret},
			{Text: "Signature Method", Widget: e.oauth1SignatureMethod},
		},
	}

	e.oauthGrantType = widget.NewSelect([]string{"client_credentials", "password"}, nil)
	e.oauthTokenURL = widget.NewEntry()
	e.oauthTokenURL.SetPlaceHolder("https://.../oauth/token")
	e.oauthClientID = widget.NewEntry()
	e.oauthClientID.SetPlaceHolder("client_id")
	e.oauthClientSecret = widget.NewEntry()
	e.oauthClientSecret.SetPlaceHolder("client_secret")
	e.oauthUsername = widget.NewEntry()
	e.oauthUsername.SetPlaceHolder("username")
	e.oauthPassword = widget.NewEntry()
	e.oauthPassword.SetPlaceHolder("password")
	e.oauthScope = widget.NewEntry()
	e.oauthScope.SetPlaceHolder("scope")
	e.oauthForm = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Grant Type", Widget: e.oauthGrantType},
			{Text: "Token URL", Widget: e.oauthTokenURL},
			{Text: "Client ID", Widget: e.oauthClientID},
			{Text: "Client Secret", Widget: e.oauthClientSecret},
			{Text: "Username", Widget: e.oauthUsername},
			{Text: "Password", Widget: e.oauthPassword},
			{Text: "Scope", Widget: e.oauthScope},
		},
	}

	e.authorization = widget.NewSelect(
		[]string{"No Auth", "API Key", "Bearer Token", "Basic Auth", "Digest Auth", "OAuth 1.0", "OAuth 2.0"},
		nil,
	)

	authSection := container.NewBorder(
		nil, nil, e.authorization, nil,
		container.NewMax(e.authorizationContent, e.basicForm, e.apiKeyForm, e.oauth1Form, e.oauthForm),
	)

	bodyTab := container.NewBorder(
		e.bodyType, nil, nil, nil,
		container.NewMax(e.rawBody, e.formSection, e.multipartSection),
	)

	e.tabs = container.NewAppTabs(
		container.NewTabItem("Query Params", paramsSection),
		container.NewTabItem("Headers", headersSection),
		container.NewTabItem("Authorization", authSection),
		container.NewTabItem("Body", bodyTab),
	)

	e.applyDefaults(rd)
	e.wireOnEdit(onEdit)

	return e
}

func (e *requestEditor) wireOnEdit(onEdit func()) {
	if onEdit == nil {
		return
	}
	changed := func(string) { onEdit() }
	e.urlEntry.OnChanged = changed
	e.rawBody.OnChanged = changed
	e.method.OnChanged = changed
	e.bodyType.OnChanged = func(value string) {
		e.applyBodyType(value)
		onEdit()
	}
	e.authorization.OnChanged = func(value string) {
		e.applyAuthType(value)
		onEdit()
	}
	e.authorizationContent.OnChanged = changed
	e.basicUserEntry.OnChanged = changed
	e.basicPassEntry.OnChanged = changed
	e.apiKeyNameEntry.OnChanged = changed
	e.apiKeyValueEntry.OnChanged = changed
	e.apiKeyLocation.OnChanged = changed
	e.oauth1ConsumerKey.OnChanged = changed
	e.oauth1ConsumerSecret.OnChanged = changed
	e.oauth1AccessToken.OnChanged = changed
	e.oauth1TokenSecret.OnChanged = changed
	e.oauth1SignatureMethod.OnChanged = changed
	e.oauthGrantType.OnChanged = changed
	e.oauthTokenURL.OnChanged = changed
	e.oauthClientID.OnChanged = changed
	e.oauthClientSecret.OnChanged = changed
	e.oauthUsername.OnChanged = changed
	e.oauthPassword.OnChanged = changed
	e.oauthScope.OnChanged = changed
}

func (e *requestEditor) applyDefaults(rd *requestData) {
	e.method.SetSelected("GET")
	e.bodyType.SetSelected("raw")
	e.apiKeyLocation.SetSelected("header")
	e.oauth1SignatureMethod.SetSelected("HMAC-SHA1")
	e.oauthGrantType.SetSelected("client_credentials")
	e.authorization.SetSelected("No Auth")
	e.applyBodyType("raw")
	e.applyAuthType("No Auth")

	if rd == nil {
		return
	}

	e.urlEntry.SetText(rd.URL)
	e.rawBody.SetText(rd.RawBody)

	if rd.Method != "" {
		e.method.SetSelected(rd.Method)
	}
	if rd.BodyType != "" {
		e.bodyType.SetSelected(rd.BodyType)
		e.applyBodyType(rd.BodyType)
	}

	e.authorizationContent.SetText(rd.Auth.Token)
	e.basicUserEntry.SetText(rd.Auth.BasicUser)
	e.basicPassEntry.SetText(rd.Auth.BasicPass)
	e.apiKeyNameEntry.SetText(rd.Auth.APIKeyName)
	e.apiKeyValueEntry.SetText(rd.Auth.APIKeyValue)
	e.oauth1ConsumerKey.SetText(rd.Auth.OAuth1ConsumerKey)
	e.oauth1ConsumerSecret.SetText(rd.Auth.OAuth1ConsumerSecret)
	e.oauth1AccessToken.SetText(rd.Auth.OAuth1AccessToken)
	e.oauth1TokenSecret.SetText(rd.Auth.OAuth1TokenSecret)
	e.oauthTokenURL.SetText(rd.Auth.TokenURL)
	e.oauthClientID.SetText(rd.Auth.ClientID)
	e.oauthClientSecret.SetText(rd.Auth.ClientSecret)
	e.oauthUsername.SetText(rd.Auth.Username)
	e.oauthPassword.SetText(rd.Auth.Password)
	e.oauthScope.SetText(rd.Auth.Scope)

	if rd.Auth.APIKeyLocation != "" {
		e.apiKeyLocation.SetSelected(rd.Auth.APIKeyLocation)
	}
	if rd.Auth.OAuth1SignatureMethod != "" {
		e.oauth1SignatureMethod.SetSelected(rd.Auth.OAuth1SignatureMethod)
	}
	if rd.Auth.GrantType != "" {
		e.oauthGrantType.SetSelected(rd.Auth.GrantType)
	}
	if rd.Auth.AuthType != "" {
		e.authorization.SetSelected(rd.Auth.AuthType)
		e.applyAuthType(rd.Auth.AuthType)
	}
}

func (e *requestEditor) applyBodyType(value string) {
	e.rawBody.Hide()
	e.formSection.Hide()
	e.multipartSection.Hide()
	switch value {
	case "raw":
		e.rawBody.Show()
	case "x-www-form-urlencoded":
		e.formSection.Show()
	case "multipart/form-data":
		e.multipartSection.Show()
	}
}

func (e *requestEditor) applyAuthType(value string) {
	e.authorizationContent.Hide()
	e.basicForm.Hide()
	e.apiKeyForm.Hide()
	e.oauth1Form.Hide()
	e.oauthForm.Hide()
	switch value {
	case "API Key":
		e.apiKeyForm.Show()
	case "Bearer Token":
		e.authorizationContent.Show()
		e.authorizationContent.Enable()
		e.authorizationContent.SetPlaceHolder("token")
	case "Basic Auth", "Digest Auth":
		e.basicForm.Show()
	case "OAuth 1.0":
		e.oauth1Form.Show()
	case "OAuth 2.0":
		e.oauthForm.Show()
	default:
		e.authorizationContent.Show()
		e.authorizationContent.Disable()
	}
}

func (e *requestEditor) snapshot() requestSnapshot {
	return requestSnapshot{
		method:      e.method.Selected,
		url:         e.urlEntry.Text,
		bodyType:    e.bodyType.Selected,
		rawBody:     e.rawBody.Text,
		queryParams: snapshotPairs(*e.params),
		headers:     snapshotPairs(*e.headers),
		form:        snapshotPairs(*e.form),
		multipart:   snapshotPairs(*e.multipart),
		auth: authConfig{
			AuthType:              e.authorization.Selected,
			Token:                 e.authorizationContent.Text,
			BasicUser:             e.basicUserEntry.Text,
			BasicPass:             e.basicPassEntry.Text,
			APIKeyName:            e.apiKeyNameEntry.Text,
			APIKeyValue:           e.apiKeyValueEntry.Text,
			APIKeyLocation:        e.apiKeyLocation.Selected,
			OAuth1ConsumerKey:     e.oauth1ConsumerKey.Text,
			OAuth1ConsumerSecret:  e.oauth1ConsumerSecret.Text,
			OAuth1AccessToken:     e.oauth1AccessToken.Text,
			OAuth1TokenSecret:     e.oauth1TokenSecret.Text,
			OAuth1SignatureMethod: e.oauth1SignatureMethod.Selected,
			GrantType:             e.oauthGrantType.Selected,
			TokenURL:              e.oauthTokenURL.Text,
			ClientID:              e.oauthClientID.Text,
			ClientSecret:          e.oauthClientSecret.Text,
			Username:              e.oauthUsername.Text,
			Password:              e.oauthPassword.Text,
			Scope:                 e.oauthScope.Text,
		},
	}
}

func (e *requestEditor) toData(name string) requestData {
	s := e.snapshot()
	return requestData{
		Name:      name,
		Method:    s.method,
		URL:       s.url,
		BodyType:  s.bodyType,
		RawBody:   s.rawBody,
		Form:      s.form,
		Multipart: s.multipart,
		Headers:   s.headers,
		Params:    s.queryParams,
		Auth:      s.auth,
	}
}

type requestTab struct {
	name           string
	collectionName string
	item           *container.TabItem
	nameEntry      *widget.Entry
	editor         *requestEditor
	status         *canvas.Text
	respBody       *widget.Entry
	respHeaders    *widget.Entry
	content        *fyne.Container
	sync           func(*requestTab)
	saveTimer      *time.Timer
}

func (rt *requestTab) scheduleSync() {
	if rt.sync == nil {
		return
	}
	if rt.saveTimer != nil {
		rt.saveTimer.Stop()
	}
	rt.saveTimer = time.AfterFunc(600*time.Millisecond, func() {
		fyne.Do(func() { rt.sync(rt) })
	})
}

func newRequestTab(w fyne.Window, rd *requestData, collectionName string, sync func(*requestTab)) *requestTab {
	name := "Nova Requisição"
	if rd != nil && rd.Name != "" {
		name = rd.Name
	}

	rt := &requestTab{
		name:           name,
		collectionName: collectionName,
		sync:           sync,
	}

	rt.editor = newRequestEditor(rd, rt.scheduleSync)

	rt.nameEntry = widget.NewEntry()
	rt.nameEntry.SetText(rt.name)
	rt.nameEntry.SetPlaceHolder("Nome da requisição")
	rt.nameEntry.OnChanged = func(string) { rt.scheduleSync() }

	status := canvas.NewText("", theme.ForegroundColor())
	status.TextSize = theme.TextSize()
	rt.status = status

	respBody := widget.NewMultiLineEntry()
	respBody.SetPlaceHolder("Resposta")
	respBody.Disable()
	respHeaders := widget.NewMultiLineEntry()
	respHeaders.SetPlaceHolder("Headers da resposta")
	respHeaders.Disable()
	rt.respBody = respBody
	rt.respHeaders = respHeaders

	copyButton := widget.NewButton("Copiar", func() {
		w.Clipboard().SetContent(respBody.Text)
	})
	statusRow := container.NewHBox(status, layout.NewSpacer(), copyButton)

	responseTabs := container.NewAppTabs(
		container.NewTabItem("Body", respBody),
		container.NewTabItem("Headers", respHeaders),
	)

	sendButton := widget.NewButton("Enviar", func() {
		s := rt.editor.snapshot()
		status.Text = "Enviando..."
		status.Color = theme.ForegroundColor()
		status.Refresh()

		go func() {
			reqBody, contentType, err := buildBody(s.bodyType, s.rawBody, s.form, s.multipart)
			if err != nil {
				fyne.Do(func() { showError(status, respBody, respHeaders, err) })
				return
			}
			result, err := sendRequest(s.method, s.url, s.queryParams, s.headers, reqBody, contentType, s.auth)
			fyne.Do(func() {
				if err != nil {
					showError(status, respBody, respHeaders, err)
					return
				}
				status.Text = fmt.Sprintf(
					"%d %s — %s",
					result.statusCode,
					http.StatusText(result.statusCode),
					result.duration.Round(time.Millisecond),
				)
				status.Color = statusColor(result.statusCode)
				status.Refresh()
				respBody.SetText(result.body)
				respHeaders.SetText(formatHeaders(result.headers))
			})
		}()
	})

	urlRow := container.NewBorder(nil, nil, rt.editor.method, sendButton, rt.editor.urlEntry)

	requestPanel := sectionPanel("Requisição", theme.PrimaryColor(), theme.InputBackgroundColor(), rt.editor.tabs)
	responsePanel := sectionPanel(
		"Resposta",
		color.RGBA{R: 0x81, G: 0xc7, B: 0x84, A: 0xff},
		theme.BackgroundColor(),
		container.NewVBox(statusRow, responseTabs),
	)

	split := container.NewVSplit(requestPanel, responsePanel)
	split.SetOffset(0.4)

	rt.content = container.NewBorder(
		container.NewVBox(rt.nameEntry, urlRow, widget.NewSeparator()),
		nil, nil, nil, split,
	)

	return rt
}
