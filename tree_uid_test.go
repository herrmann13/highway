package main

import (
	"testing"

	"highway/storage"
)

func TestTreeUIDsSupportLegacyNames(t *testing.T) {
	collections := []*storage.Collection{
		{
			Name: "API / Produção",
			Requests: []storage.RequestData{
				{Name: "Listar usuários / ativos"},
				{Name: "Buscar relatório mensal"},
			},
		},
	}

	if got := treeChildUIDs("", collections); len(got) != 1 || got[0] != "col:0" {
		t.Fatalf("UID da collection incorreto: %#v", got)
	}
	if got := treeChildUIDs("col:0", collections); len(got) != 2 || got[0] != "req:0:0" || got[1] != "req:0:1" {
		t.Fatalf("UIDs das requests incorretos: %#v", got)
	}

	c, request, ok := requestForUID(collections, "req:0:0")
	if !ok || c.Name != "API / Produção" || request.Name != "Listar usuários / ativos" {
		t.Fatalf("request não foi resolvida: %#v %#v", c, request)
	}
}

func TestTreeUIDRejectsInvalidIndexes(t *testing.T) {
	collections := []*storage.Collection{{Name: "API", Requests: []storage.RequestData{{Name: "Listar"}}}}
	if _, _, ok := collectionForUID(collections, "col:1"); ok {
		t.Fatal("aceitou collection fora do intervalo")
	}
	if _, _, ok := requestForUID(collections, "req:0:1"); ok {
		t.Fatal("aceitou request fora do intervalo")
	}
	if _, _, ok := requestForUID(collections, "req:API/Listar"); ok {
		t.Fatal("aceitou UID legado baseado em nomes")
	}
}
