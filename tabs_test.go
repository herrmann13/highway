package main

import "testing"

func TestRequestTabMatchesCollectionAndRequest(t *testing.T) {
	rt := &requestTab{collectionName: "Admin", name: "Listar"}
	if !requestTabMatches(rt, "Admin", "Listar") {
		t.Fatal("não identificou a aba correta")
	}
	if requestTabMatches(rt, "API", "Listar") {
		t.Fatal("confundiu requests de collections diferentes")
	}
	if requestTabMatches(nil, "Admin", "Listar") {
		t.Fatal("identificou uma aba nil")
	}
}
