package clipboard

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"highway/curl"
)

const DefaultCurlHistoryLimit = 20

type CurlHistoryEntry struct {
	Command string
	Label   string
	Tooltip string
}

type CurlHistory struct {
	mu      sync.RWMutex
	entries []CurlHistoryEntry
	limit   int
}

func NewCurlHistory(limit int) *CurlHistory {
	if limit < 1 {
		limit = DefaultCurlHistoryLimit
	}
	return &CurlHistory{limit: limit}
}

func (h *CurlHistory) Add(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for i, entry := range h.entries {
		if entry.Command == command {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}
	h.entries = append([]CurlHistoryEntry{{
		Command: command,
		Label:   curlHistoryLabel(command),
		Tooltip: curlHistoryTooltip(command),
	}}, h.entries...)
	if len(h.entries) > h.limit {
		h.entries = h.entries[:h.limit]
	}
}

func (h *CurlHistory) Entries() []CurlHistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]CurlHistoryEntry(nil), h.entries...)
}

func curlHistoryLabel(command string) string {
	if requestURL, method, ok := curlHistoryParts(command); ok {
		return fmt.Sprintf("%s %s", method, requestURL)
	}
	return "cURL inválido"
}

func curlHistoryTooltip(command string) string {
	if requestURL, method, ok := curlHistoryParts(command); ok {
		return fmt.Sprintf("curl %s %s", requestURL, method)
	}
	return "cURL inválido"
}

func curlHistoryParts(command string) (string, string, bool) {
	args, err := curl.SplitShellArgs(command)
	if err != nil || len(args) == 0 || args[0] != "curl" {
		return "", "", false
	}

	method := "GET"
	hasData := false
	forceGet := false
	requestURL := ""

	valueFor := func(index *int) string {
		*index++
		if *index >= len(args) {
			return ""
		}
		return args[*index]
	}
	setURL := func(value string) {
		if requestURL == "" && isHTTPURL(value) {
			requestURL = value
		}
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-X" || arg == "--request":
			if value := valueFor(&i); value != "" {
				method = strings.ToUpper(value)
			}
		case strings.HasPrefix(arg, "--request="):
			method = strings.ToUpper(strings.TrimPrefix(arg, "--request="))
		case arg == "-G" || arg == "--get":
			forceGet = true
		case arg == "-d" || arg == "--data" || arg == "--data-raw" || arg == "--data-binary" || arg == "--data-ascii" || arg == "--data-urlencode" || arg == "-F" || arg == "--form":
			hasData = true
			_ = valueFor(&i)
		case strings.HasPrefix(arg, "--data=") || strings.HasPrefix(arg, "--data-raw=") || strings.HasPrefix(arg, "--data-binary=") || strings.HasPrefix(arg, "--data-ascii=") || strings.HasPrefix(arg, "--data-urlencode=") || strings.HasPrefix(arg, "--form="):
			hasData = true
		case arg == "--url":
			setURL(valueFor(&i))
		case strings.HasPrefix(arg, "--url="):
			setURL(strings.TrimPrefix(arg, "--url="))
		case arg == "-H" || arg == "--header" || arg == "-u" || arg == "--user" || arg == "--oauth2-bearer":
			_ = valueFor(&i)
		case strings.HasPrefix(arg, "--header=") || strings.HasPrefix(arg, "--user=") || strings.HasPrefix(arg, "--oauth2-bearer="):
		case strings.HasPrefix(arg, "-"):
			continue
		default:
			setURL(arg)
		}
	}

	if forceGet {
		method = "GET"
	} else if method == "GET" && hasData {
		method = "POST"
	}
	return requestURL, method, requestURL != ""
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
