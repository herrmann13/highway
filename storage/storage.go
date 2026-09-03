package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RequestData struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	BodyType  string      `json:"bodyType"`
	RawBody   string      `json:"rawBody"`
	Form      [][2]string `json:"form"`
	Multipart [][2]string `json:"multipart"`
	Headers   [][2]string `json:"headers"`
	Params    [][2]string `json:"params"`
	Auth      AuthConfig  `json:"auth"`
}

type Collection struct {
	Name      string        `json:"name"`
	Variables [][2]string   `json:"variables"`
	Requests  []RequestData `json:"requests"`
}

type AuthConfig struct {
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

type RequestSnapshot struct {
	Method      string
	URL         string
	BodyType    string
	RawBody     string
	QueryParams [][2]string
	Headers     [][2]string
	Form        [][2]string
	Multipart   [][2]string
	Auth        AuthConfig
}

var invalidFileChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "highway")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func CollectionsDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "collections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := filepath.Dir(configDir)
	if err := MigrateCollections(filepath.Join(base, "carteiro", "collections"), dir); err != nil {
		return "", err
	}
	return dir, nil
}

func MigrateCollections(sourceDir, destinationDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		destination := filepath.Join(destinationDir, entry.Name())
		if _, err := os.Stat(destination); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		data, err := os.ReadFile(filepath.Join(sourceDir, entry.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func CollectionFilePath(dir, name string) string {
	safe := invalidFileChars.ReplaceAllString(name, "_")
	return filepath.Join(dir, safe+".json")
}

func LoadCollections() ([]*Collection, error) {
	dir, err := CollectionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []*Collection
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var c Collection
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		result = append(result, &c)
	}
	return result, nil
}

func SaveCollection(c *Collection) error {
	dir, err := CollectionsDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CollectionFilePath(dir, c.Name), data, 0o644)
}

func RenameCollection(c *Collection, name string) error {
	if name == c.Name {
		return nil
	}

	dir, err := CollectionsDir()
	if err != nil {
		return err
	}

	renamed := *c
	renamed.Name = name
	data, err := json.MarshalIndent(&renamed, "", "  ")
	if err != nil {
		return err
	}

	oldPath := CollectionFilePath(dir, c.Name)
	newPath := CollectionFilePath(dir, name)
	if err := os.WriteFile(newPath, data, 0o644); err != nil {
		return err
	}
	if oldPath != newPath {
		if err := os.Remove(oldPath); err != nil {
			_ = os.Remove(newPath)
			return err
		}
	}

	c.Name = name
	return nil
}

func DeleteCollection(name string) error {
	dir, err := CollectionsDir()
	if err != nil {
		return err
	}
	return os.Remove(CollectionFilePath(dir, name))
}

func UpsertRequest(collections []*Collection, colName, oldName string, rd RequestData) error {
	for _, c := range collections {
		if c.Name != colName {
			continue
		}
		if oldName != "" {
			for i := range c.Requests {
				if c.Requests[i].Name == oldName {
					if rd.Name != oldName && RequestNameExists(c, rd.Name, i) {
						return fmt.Errorf("já existe uma requisição com o nome %q", rd.Name)
					}
					c.Requests[i] = rd
					return SaveCollection(c)
				}
			}
		}
		if RequestNameExists(c, rd.Name, -1) {
			return fmt.Errorf("já existe uma requisição com o nome %q", rd.Name)
		}
		c.Requests = append(c.Requests, rd)
		return SaveCollection(c)
	}
	return fmt.Errorf("coleção não encontrada: %s", colName)
}

func RequestNameExists(c *Collection, name string, except int) bool {
	for i, r := range c.Requests {
		if i != except && r.Name == name {
			return true
		}
	}
	return false
}

func UniqueRequestName(c *Collection, base string) string {
	name := base
	for i := 2; ; i++ {
		if !RequestNameExists(c, name, -1) {
			return name
		}
		name = fmt.Sprintf("%s %d", base, i)
	}
}
