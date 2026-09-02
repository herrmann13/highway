package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportAndReadBundle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "highway-export.json")
	original := []*collection{{Name: "API", Variables: [][2]string{{"base_url", "https://example.com"}}, Requests: []requestData{{Name: "Listar", Method: "GET"}}}}
	if err := exportCollections(path, original); err != nil {
		t.Fatal(err)
	}
	bundle, err := readBundle(path)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Version != exportBundleVersion || len(bundle.Collections) != 1 || bundle.Collections[0].Requests[0].Name != "Listar" {
		t.Fatalf("bundle incorreto: %#v", bundle)
	}
}

func TestReadBundleRejectsInvalidFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	for _, content := range []string{`{`, `{"version":2,"collections":[]}`, `{"version":1,"collections":[]}`, `{"version":1,"collections":[{"name":""}]}`} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := readBundle(path); err == nil {
			t.Fatalf("arquivo inválido foi aceito: %s", content)
		}
	}
}

func TestUniqueCollectionName(t *testing.T) {
	collections := []*collection{{Name: "API"}, {Name: "API (2)"}}
	if got := uniqueCollectionName(collections, "Nova"); got != "Nova" {
		t.Fatalf("nome inesperado: %q", got)
	}
	if got := uniqueCollectionName(collections, "API"); got != "API (3)" {
		t.Fatalf("nome de conflito inesperado: %q", got)
	}
}
