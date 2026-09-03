package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateCollectionsCopiesLegacyFilesWithoutOverwrite(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "carteiro", "collections")
	destination := filepath.Join(root, "highway", "collections")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "legacy.json"), []byte(`{"name":"Legacy"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "current.json"), []byte(`{"name":"Old"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "current.json"), []byte(`{"name":"New"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := MigrateCollections(source, destination); err != nil {
		t.Fatal(err)
	}
	legacy, err := os.ReadFile(filepath.Join(destination, "legacy.json"))
	if err != nil || string(legacy) != `{"name":"Legacy"}` {
		t.Fatalf("collection antiga não foi migrada: %q, %v", legacy, err)
	}
	current, err := os.ReadFile(filepath.Join(destination, "current.json"))
	if err != nil || string(current) != `{"name":"New"}` {
		t.Fatalf("collection existente foi sobrescrita: %q, %v", current, err)
	}
}
