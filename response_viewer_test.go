package main

import (
	"strings"
	"testing"
)

func TestResponsePagesPreservesBody(t *testing.T) {
	lines := make([]string, responsePageMaxLines+10)
	for i := range lines {
		lines[i] = "line"
	}
	body := strings.Join(lines, "\n")
	pages := responsePages(body)
	if len(pages) != 2 {
		t.Fatalf("quantidade de páginas = %d, want 2", len(pages))
	}

	var rebuilt strings.Builder
	for _, page := range pages {
		rebuilt.WriteString(body[page.start:page.end])
	}
	if rebuilt.String() != body {
		t.Fatal("as páginas não preservaram o conteúdo da resposta")
	}
	if pages[0].finishLine-pages[0].startLine+1 != responsePageMaxLines {
		t.Fatalf("primeira página tem %d linhas, want %d", pages[0].finishLine-pages[0].startLine+1, responsePageMaxLines)
	}
}

func TestResponsePagesSplitsLongLine(t *testing.T) {
	body := strings.Repeat("a", responsePageMaxBytes+100)
	pages := responsePages(body)
	if len(pages) != 2 {
		t.Fatalf("quantidade de páginas = %d, want 2", len(pages))
	}
	if pages[0].end-pages[0].start > responsePageMaxBytes {
		t.Fatal("a página excedeu o limite de bytes")
	}
}
