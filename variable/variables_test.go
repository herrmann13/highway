package variable

import (
	"strings"
	"testing"

	"highway/storage"
)

func TestExpandRequestSnapshot(t *testing.T) {
	snapshot := storage.RequestSnapshot{
		URL:         "{{base_url}}/users/{{user_id}}",
		RawBody:     `{"token":"{{token}}"}`,
		QueryParams: [][2]string{{"page", "{{page}}"}},
		Headers:     [][2]string{{"Authorization", "Bearer {{token}}"}},
		Form:        [][2]string{{"user", "{{user_id}}"}},
		Multipart:   [][2]string{{"file", "{{file_name}}"}},
		Auth: storage.AuthConfig{
			BasicUser: "{{user_id}}",
			BasicPass: "{{token}}",
		},
	}
	variables := [][2]string{
		{"base_url", "https://api.example.com"},
		{"user_id", "42"},
		{"token", "secret"},
		{"page", "2"},
		{"file_name", "report.pdf"},
	}

	got, err := ExpandRequestSnapshot(snapshot, variables)
	if err != nil {
		t.Fatal(err)
	}
	if got.URL != "https://api.example.com/users/42" || got.RawBody != `{"token":"secret"}` {
		t.Fatalf("URL ou body não expandidos: %#v", got)
	}
	if got.QueryParams[0][1] != "2" || got.Headers[0][1] != "Bearer secret" {
		t.Fatalf("query ou header não expandidos: %#v", got)
	}
	if got.Form[0][1] != "42" || got.Multipart[0][1] != "report.pdf" {
		t.Fatalf("formulários não expandidos: %#v", got)
	}
	if got.Auth.BasicUser != "42" || got.Auth.BasicPass != "secret" {
		t.Fatalf("auth não expandida: %#v", got.Auth)
	}
	if snapshot.URL != "{{base_url}}/users/{{user_id}}" {
		t.Fatal("o snapshot original foi alterado")
	}
}

func TestExpandRequestSnapshotErrors(t *testing.T) {
	_, err := ExpandRequestSnapshot(storage.RequestSnapshot{URL: "{{missing}}"}, nil)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("erro de variável ausente inesperado: %v", err)
	}

	_, err = ExpandRequestSnapshot(storage.RequestSnapshot{}, [][2]string{{"token", "one"}, {"token", "two"}})
	if err == nil || !strings.Contains(err.Error(), "mais de uma vez") {
		t.Fatalf("erro de variável duplicada inesperado: %v", err)
	}
}
