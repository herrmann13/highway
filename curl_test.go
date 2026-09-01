package main

import "testing"

func TestParseCurlJSONPost(t *testing.T) {
	rd, err := parseCurl(`curl -X POST 'https://api.example.com/users?active=true' -H 'Content-Type: application/json' -H 'X-Token: abc' -d '{"name":"Ana"}'`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Method != "POST" || rd.URL != "https://api.example.com/users" {
		t.Fatalf("requisição inesperada: %#v", rd)
	}
	if rd.RawBody != `{"name":"Ana"}` || len(rd.Headers) != 2 {
		t.Fatalf("body ou headers incorretos: %#v", rd)
	}
	if len(rd.Params) != 1 || rd.Params[0] != [2]string{"active", "true"} {
		t.Fatalf("params incorretos: %#v", rd.Params)
	}
}

func TestParseCurlGetWithData(t *testing.T) {
	rd, err := parseCurl(`curl -G https://api.example.com/search -d 'term=golang' --data-urlencode 'page=2'`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Method != "GET" || rd.RawBody != "" {
		t.Fatalf("GET incorreto: %#v", rd)
	}
	if len(rd.Params) != 2 {
		t.Fatalf("params incorretos: %#v", rd.Params)
	}
}

func TestParseCurlMultipartAndBasicAuth(t *testing.T) {
	rd, err := parseCurl("curl --user ana:segredo -F 'title=Relatório' -F 'status=draft' https://api.example.com/documents")
	if err != nil {
		t.Fatal(err)
	}
	if rd.Method != "POST" || rd.BodyType != "multipart/form-data" {
		t.Fatalf("multipart incorreto: %#v", rd)
	}
	if rd.Auth.AuthType != "Basic Auth" || rd.Auth.BasicUser != "ana" || rd.Auth.BasicPass != "segredo" {
		t.Fatalf("autenticação incorreta: %#v", rd.Auth)
	}
	if len(rd.Multipart) != 2 {
		t.Fatalf("campos multipart incorretos: %#v", rd.Multipart)
	}
}

func TestParseCurlMultilineAndErrors(t *testing.T) {
	rd, err := parseCurl(`curl \
  --url https://api.example.com/items \
  --data-urlencode 'filter=em aberto'`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Method != "POST" || rd.BodyType != "x-www-form-urlencoded" || len(rd.Form) != 1 {
		t.Fatalf("formulário incorreto: %#v", rd)
	}
	if _, err := parseCurl("curl --proxy localhost https://api.example.com"); err == nil {
		t.Fatal("esperava erro para opção não suportada")
	}
}
