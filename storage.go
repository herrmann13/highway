package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type requestData struct {
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
	Auth      authConfig  `json:"auth"`
}

type collection struct {
	Name      string        `json:"name"`
	Variables [][2]string   `json:"variables"`
	Requests  []requestData `json:"requests"`
}

var invalidFileChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func collectionsDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "highway", "collections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := migrateCollections(filepath.Join(base, "carteiro", "collections"), dir); err != nil {
		return "", err
	}
	return dir, nil
}

func migrateCollections(sourceDir, destinationDir string) error {
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

func collectionFilePath(dir, name string) string {
	safe := invalidFileChars.ReplaceAllString(name, "_")
	return filepath.Join(dir, safe+".json")
}

func loadCollections() ([]*collection, error) {
	dir, err := collectionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []*collection
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var c collection
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		result = append(result, &c)
	}
	return result, nil
}

func saveCollection(c *collection) error {
	dir, err := collectionsDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(collectionFilePath(dir, c.Name), data, 0o644)
}

func renameCollection(c *collection, name string) error {
	if name == c.Name {
		return nil
	}

	dir, err := collectionsDir()
	if err != nil {
		return err
	}

	renamed := *c
	renamed.Name = name
	data, err := json.MarshalIndent(&renamed, "", "  ")
	if err != nil {
		return err
	}

	oldPath := collectionFilePath(dir, c.Name)
	newPath := collectionFilePath(dir, name)
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

func deleteCollection(name string) error {
	dir, err := collectionsDir()
	if err != nil {
		return err
	}
	return os.Remove(collectionFilePath(dir, name))
}

func upsertRequest(collections []*collection, colName, oldName string, rd requestData) error {
	for _, c := range collections {
		if c.Name != colName {
			continue
		}
		if oldName != "" {
			for i := range c.Requests {
				if c.Requests[i].Name == oldName {
					if rd.Name != oldName && requestNameExists(c, rd.Name, i) {
						return fmt.Errorf("já existe uma requisição com o nome %q", rd.Name)
					}
					c.Requests[i] = rd
					return saveCollection(c)
				}
			}
		}
		if requestNameExists(c, rd.Name, -1) {
			return fmt.Errorf("já existe uma requisição com o nome %q", rd.Name)
		}
		c.Requests = append(c.Requests, rd)
		return saveCollection(c)
	}
	return fmt.Errorf("coleção não encontrada: %s", colName)
}

func requestNameExists(c *collection, name string, except int) bool {
	for i, r := range c.Requests {
		if i != except && r.Name == name {
			return true
		}
	}
	return false
}

func uniqueRequestName(c *collection, base string) string {
	name := base
	for i := 2; ; i++ {
		if !requestNameExists(c, name, -1) {
			return name
		}
		name = fmt.Sprintf("%s %d", base, i)
	}
}
