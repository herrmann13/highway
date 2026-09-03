package clipboard

import "testing"

func TestCurlHistoryAddsMostRecentFirst(t *testing.T) {
	history := NewCurlHistory(2)
	history.Add("curl https://example.com/first")
	history.Add("curl -X POST https://example.com/second")
	history.Add("curl https://example.com/third")

	entries := history.Entries()
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Command != "curl https://example.com/third" {
		t.Fatalf("entries[0] = %q", entries[0].Command)
	}
	if entries[1].Command != "curl -X POST https://example.com/second" {
		t.Fatalf("entries[1] = %q", entries[1].Command)
	}
}

func TestCurlHistoryMovesDuplicateToTop(t *testing.T) {
	history := NewCurlHistory(3)
	first := "curl https://example.com/first"
	second := "curl https://example.com/second"
	history.Add(first)
	history.Add(second)
	history.Add(first)

	entries := history.Entries()
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if entries[0].Command != first || entries[1].Command != second {
		t.Fatalf("unexpected entry order: %#v", entries)
	}
}

func TestCurlHistoryLabel(t *testing.T) {
	history := NewCurlHistory(1)
	history.Add("curl -X POST https://api.example.com/users")

	entries := history.Entries()
	if entries[0].Label != "POST https://api.example.com/users" {
		t.Fatalf("label = %q", entries[0].Label)
	}
}
