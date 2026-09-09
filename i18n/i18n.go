package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed translations/*.json
var translationsFS embed.FS

var (
	mu       sync.RWMutex
	locale   = "pt"
	messages = map[string]map[string]string{}
)

func init() {
	for _, lang := range []string{"pt", "en"} {
		data, err := translationsFS.ReadFile("translations/" + lang + ".json")
		if err != nil {
			continue
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		messages[lang] = m
	}
}

// SetLocale switches the active language if it is supported.
func SetLocale(code string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := messages[code]; ok {
		locale = code
	}
}

// Locale returns the currently active language code.
func Locale() string {
	mu.RLock()
	defer mu.RUnlock()
	return locale
}

// T returns the translation for key in the active language, falling back to
// Portuguese and finally to the key itself.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if m, ok := messages[locale][key]; ok && m != "" {
		return m
	}
	if m, ok := messages["pt"][key]; ok && m != "" {
		return m
	}
	return key
}

// Tf returns the translation for key formatted with the given arguments.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

// Errorf wraps err with the localized message for key.
func Errorf(key string, err error) error {
	return fmt.Errorf("%s: %w", T(key), err)
}
