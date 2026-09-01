package main

import (
	"strings"
	"testing"
)

func TestExpandRequestSnapshot(t *testing.T) {
	snapshot := requestSnapshot{
		url:         "{{base_url}}/users/{{user_id}}",
		rawBody:     `{"token":"{{token}}"}`,
		queryParams: [][2]string{{"page", "{{page}}"}},
		headers:     [][2]string{{"Authorization", "Bearer {{token}}"}},
		form:        [][2]string{{"user", "{{user_id}}"}},
		multipart:   [][2]string{{"file", "{{file_name}}"}},
		auth: authConfig{
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

	got, err := expandRequestSnapshot(snapshot, variables)
	if err != nil {
		t.Fatal(err)
	}
	if got.url != "https://api.example.com/users/42" || got.rawBody != `{"token":"secret"}` {
		t.Fatalf("URL ou body não expandidos: %#v", got)
	}
	if got.queryParams[0][1] != "2" || got.headers[0][1] != "Bearer secret" {
		t.Fatalf("query ou header não expandidos: %#v", got)
	}
	if got.form[0][1] != "42" || got.multipart[0][1] != "report.pdf" {
		t.Fatalf("formulários não expandidos: %#v", got)
	}
	if got.auth.BasicUser != "42" || got.auth.BasicPass != "secret" {
		t.Fatalf("auth não expandida: %#v", got.auth)
	}
	if snapshot.url != "{{base_url}}/users/{{user_id}}" {
		t.Fatal("o snapshot original foi alterado")
	}
}

func TestExpandRequestSnapshotErrors(t *testing.T) {
	_, err := expandRequestSnapshot(requestSnapshot{url: "{{missing}}"}, nil)
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("erro de variável ausente inesperado: %v", err)
	}

	_, err = expandRequestSnapshot(requestSnapshot{}, [][2]string{{"token", "one"}, {"token", "two"}})
	if err == nil || !strings.Contains(err.Error(), "mais de uma vez") {
		t.Fatalf("erro de variável duplicada inesperado: %v", err)
	}
}
