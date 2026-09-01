package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestImportCommandFromArgs(t *testing.T) {
	command, handled, err := importCommandFromArgs([]string{"import", "--curl-stdin"}, strings.NewReader("curl https://api.example.com/users"))
	if err != nil || !handled || command != "curl https://api.example.com/users" {
		t.Fatalf("comando inválido: %q, %t, %v", command, handled, err)
	}
	if _, handled, err := importCommandFromArgs([]string{"import", "--curl-stdin"}, strings.NewReader("texto")); !handled || err == nil {
		t.Fatal("texto não-cURL foi aceito")
	}
	if _, handled, err := importCommandFromArgs(nil, strings.NewReader("")); handled || err != nil {
		t.Fatal("modo GUI foi tratado como comando CLI")
	}
}

func TestImportFileFromArgsReadsAndRemovesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pending.curl")
	if err := os.WriteFile(path, []byte("curl https://api.example.com/users"), 0o600); err != nil {
		t.Fatal(err)
	}
	command, handled, err := importFileFromArgs([]string{"--import-file", path})
	if err != nil || !handled || command != "curl https://api.example.com/users" {
		t.Fatalf("arquivo de importação inválido: %q, %t, %v", command, handled, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("arquivo de importação não foi removido")
	}
}

func TestImportSocketDeliversCommand(t *testing.T) {
	path := testImportSocketPath(t)
	received := make(chan string, 1)
	closeServer, err := startImportServerAt(path, func(command string) { received <- command })
	if err != nil {
		t.Fatal(err)
	}
	defer closeServer()

	delivered, err := sendImportCommandTo(path, "curl https://api.example.com/users")
	if err != nil || !delivered {
		t.Fatalf("importação não entregue: %t, %v", delivered, err)
	}
	select {
	case command := <-received:
		if command != "curl https://api.example.com/users" {
			t.Fatalf("comando inesperado: %q", command)
		}
	default:
		t.Fatal("servidor não recebeu o comando")
	}
}

func TestImportSocketRejectsInvalidCommand(t *testing.T) {
	path := testImportSocketPath(t)
	closeServer, err := startImportServerAt(path, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	defer closeServer()

	if _, err := sendImportCommandTo(path, "texto comum"); err == nil {
		t.Fatal("servidor aceitou texto não-cURL")
	}
}

func testImportSocketPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(os.TempDir(), fmt.Sprintf("highway-%d.sock", time.Now().UnixNano()))
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}
