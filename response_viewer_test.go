package main

import "testing"

func TestFormatResponseBody(t *testing.T) {
	got := formatResponseBody([]byte(`{"users":[{"id":1,"name":"Ana"}]}`))
	want := "{\n  \"users\": [\n    {\n      \"id\": 1,\n      \"name\": \"Ana\"\n    }\n  ]\n}"
	if got != want {
		t.Fatalf("JSON formatado incorretamente:\n%s", got)
	}
}

func TestFormatResponseBodyKeepsNonJSON(t *testing.T) {
	const body = "not json"
	if got := formatResponseBody([]byte(body)); got != body {
		t.Fatalf("body não JSON foi alterado: %q", got)
	}
}
