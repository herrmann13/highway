package clipboard

import (
	"fmt"
	"strings"
	"sync"

	"highway/curl"
)

const DefaultCurlHistoryLimit = 20

type CurlHistoryEntry struct {
	Command string
	Label   string
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
	h.entries = append([]CurlHistoryEntry{{Command: command, Label: curlHistoryLabel(command)}}, h.entries...)
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
	request, err := curl.ParseCurl(command)
	if err == nil && request.URL != "" {
		return fmt.Sprintf("%s %s", request.Method, request.URL)
	}
	return strings.ReplaceAll(command, "\n", " ")
}
