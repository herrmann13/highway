package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

func parseCurl(command string) (requestData, error) {
	args, err := splitShellArgs(command)
	if err != nil {
		return requestData{}, err
	}
	if len(args) == 0 || args[0] != "curl" {
		return requestData{}, fmt.Errorf("o comando deve começar com curl")
	}

	rd := requestData{Type: requestTypeHTTP, Method: "GET", BodyType: "raw"}
	var rawData []string
	var encodedData []string
	forceGet := false
	var bearerToken string

	nextValue := func(index *int, option string) (string, error) {
		*index = *index + 1
		if *index >= len(args) {
			return "", fmt.Errorf("a opção %s exige um valor", option)
		}
		return args[*index], nil
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-X", "--request":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			rd.Method = strings.ToUpper(value)
		case "-H", "--header":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			parts := strings.SplitN(value, ":", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return requestData{}, fmt.Errorf("header inválido: %q", value)
			}
			rd.Headers = append(rd.Headers, [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			rawData = append(rawData, value)
		case "--data-urlencode":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			encodedData = append(encodedData, value)
		case "-F", "--form":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			parts := strings.SplitN(value, "=", 2)
			if len(parts) != 2 || parts[0] == "" {
				return requestData{}, fmt.Errorf("campo multipart inválido: %q", value)
			}
			rd.Multipart = append(rd.Multipart, [2]string{parts[0], parts[1]})
			rd.BodyType = "multipart/form-data"
		case "-G", "--get":
			forceGet = true
		case "-u", "--user":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			credentials := strings.SplitN(value, ":", 2)
			rd.Auth.AuthType = "Basic Auth"
			rd.Auth.BasicUser = credentials[0]
			if len(credentials) == 2 {
				rd.Auth.BasicPass = credentials[1]
			}
		case "--oauth2-bearer":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			bearerToken = value
		case "--url":
			value, err := nextValue(&i, arg)
			if err != nil {
				return requestData{}, err
			}
			rd.URL = value
		case "-L", "--location", "-s", "--silent", "-k", "--insecure", "-i", "--include", "-v", "--verbose", "--compressed", "--globoff", "--fail", "--fail-with-body":
			// These options change cURL behavior but not the HTTP request definition.
		default:
			if strings.HasPrefix(arg, "-") {
				return requestData{}, fmt.Errorf("opção cURL não suportada: %s", arg)
			}
			if rd.URL != "" {
				return requestData{}, fmt.Errorf("mais de uma URL foi informada")
			}
			rd.URL = arg
		}
	}
	if bearerToken != "" {
		rd.Auth.AuthType = "Bearer Token"
		rd.Auth.Token = bearerToken
	}
	var basicUser, basicPass string
	var basicFound bool
	rd.Headers, basicUser, basicPass, basicFound = extractBasicAuth(rd.Headers)
	if basicFound {
		rd.Auth.AuthType = "Basic Auth"
		rd.Auth.BasicUser = basicUser
		rd.Auth.BasicPass = basicPass
	}
	rd.Headers, bearerToken, _ = extractBearerAuth(rd.Headers)
	if bearerToken != "" {
		rd.Auth.AuthType = "Bearer Token"
		rd.Auth.Token = bearerToken
	}

	if rd.URL == "" {
		return requestData{}, fmt.Errorf("URL não encontrada no comando cURL")
	}

	u, err := url.Parse(rd.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return requestData{}, fmt.Errorf("URL inválida: %q", rd.URL)
	}
	for key, values := range u.Query() {
		for _, value := range values {
			rd.Params = append(rd.Params, [2]string{key, value})
		}
	}
	u.RawQuery = ""
	rd.URL = u.String()

	if forceGet {
		rd.Method = "GET"
	}
	if len(rawData) > 0 || len(encodedData) > 0 || len(rd.Multipart) > 0 {
		if rd.Method == "GET" && !forceGet {
			rd.Method = "POST"
		}
	}

	if forceGet {
		for _, data := range rawData {
			values, err := url.ParseQuery(data)
			if err != nil {
				return requestData{}, fmt.Errorf("dados de query inválidos: %w", err)
			}
			for key, entries := range values {
				for _, value := range entries {
					rd.Params = append(rd.Params, [2]string{key, value})
				}
			}
		}
		for _, data := range encodedData {
			parts := strings.SplitN(data, "=", 2)
			key := parts[0]
			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}
			rd.Params = append(rd.Params, [2]string{key, value})
		}
	}

	if !forceGet && len(rawData) > 0 {
		rd.RawBody = strings.Join(rawData, "&")
		if !hasHeader(rd.Headers, "Content-Type") {
			rd.Headers = append(rd.Headers, [2]string{"Content-Type", "application/x-www-form-urlencoded"})
		}
	}
	if !forceGet && len(encodedData) > 0 {
		rd.BodyType = "x-www-form-urlencoded"
		for _, data := range encodedData {
			parts := strings.SplitN(data, "=", 2)
			key := parts[0]
			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}
			rd.Form = append(rd.Form, [2]string{key, value})
		}
	}

	return rd, nil
}

func extractBearerAuth(headers [][2]string) ([][2]string, string, bool) {
	filtered := make([][2]string, 0, len(headers))
	var token string
	found := false
	for _, header := range headers {
		if strings.EqualFold(strings.TrimSpace(header[0]), "Authorization") {
			value := strings.TrimSpace(header[1])
			if len(value) > len("Bearer") && strings.EqualFold(value[:len("Bearer")], "Bearer") && (value[len("Bearer")] == ' ' || value[len("Bearer")] == '\t') {
				candidate := strings.TrimSpace(value[len("Bearer"):])
				if candidate != "" {
					token = candidate
					found = true
					continue
				}
			}
		}
		filtered = append(filtered, header)
	}
	return filtered, token, found
}

func extractBasicAuth(headers [][2]string) ([][2]string, string, string, bool) {
	filtered := make([][2]string, 0, len(headers))
	var user, password string
	found := false
	for _, header := range headers {
		if strings.EqualFold(strings.TrimSpace(header[0]), "Authorization") {
			value := strings.TrimSpace(header[1])
			if len(value) > len("Basic") && strings.EqualFold(value[:len("Basic")], "Basic") && (value[len("Basic")] == ' ' || value[len("Basic")] == '\t') {
				encoded := strings.TrimSpace(value[len("Basic"):])
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				if err == nil {
					credentials := string(decoded)
					if separator := strings.IndexByte(credentials, ':'); separator >= 0 {
						user = credentials[:separator]
						password = credentials[separator+1:]
						found = true
						continue
					}
				}
			}
		}
		filtered = append(filtered, header)
	}
	return filtered, user, password, found
}

func hasHeader(headers [][2]string, name string) bool {
	for _, header := range headers {
		if strings.EqualFold(header[0], name) {
			return true
		}
	}
	return false
}

func splitShellArgs(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	quote := rune(0)
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for _, r := range command {
		if escaped {
			if r != '\n' {
				current.WriteRune(r)
			}
			escaped = false
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		return nil, fmt.Errorf("escape incompleto no comando cURL")
	}
	if quote != 0 {
		return nil, fmt.Errorf("aspas não fechadas no comando cURL")
	}
	flush()
	return args, nil
}
