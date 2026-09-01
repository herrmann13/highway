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

func TestParseCurlBearerAuth(t *testing.T) {
	rd, err := parseCurl(`curl https://api.example.com/users -H 'accept: application/json' -H 'authorization: Bearer eyJhbGciOiJIUzI1NiJ9'`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Auth.AuthType != "Bearer Token" || rd.Auth.Token != "eyJhbGciOiJIUzI1NiJ9" {
		t.Fatalf("Bearer não foi extraído: %#v", rd.Auth)
	}
	if len(rd.Headers) != 1 || rd.Headers[0][0] != "accept" {
		t.Fatalf("Authorization não foi removido dos headers: %#v", rd.Headers)
	}
}

func TestParseCurlBasicHeaderAuth(t *testing.T) {
	rd, err := parseCurl(`curl https://api.example.com/users -H 'Authorization: Basic YW5hOnNlY3JldG86c2VuaGE='`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Auth.AuthType != "Basic Auth" || rd.Auth.BasicUser != "ana" || rd.Auth.BasicPass != "secreto:senha" {
		t.Fatalf("Basic Auth não foi extraído: %#v", rd.Auth)
	}
	if len(rd.Headers) != 0 {
		t.Fatalf("Authorization não foi removido dos headers: %#v", rd.Headers)
	}
}

func TestParseCurlInvalidBasicHeaderIsPreserved(t *testing.T) {
	rd, err := parseCurl(`curl https://api.example.com/users -H 'Authorization: Basic inválido'`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Auth.AuthType == "Basic Auth" || len(rd.Headers) != 1 {
		t.Fatalf("header Basic inválido foi alterado: %#v %#v", rd.Auth, rd.Headers)
	}
}

func TestParseCurlOAuth2Bearer(t *testing.T) {
	rd, err := parseCurl(`curl --oauth2-bearer '{{token}}' https://api.example.com/users`)
	if err != nil {
		t.Fatal(err)
	}
	if rd.Auth.AuthType != "Bearer Token" || rd.Auth.Token != "{{token}}" {
		t.Fatalf("oauth2 bearer não foi extraído: %#v", rd.Auth)
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
