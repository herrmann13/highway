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
	if entries[0].Tooltip != "curl https://api.example.com/users POST" {
		t.Fatalf("tooltip = %q", entries[0].Tooltip)
	}
}

func TestCurlHistorySummaryIgnoresTrailingOptions(t *testing.T) {
	history := NewCurlHistory(1)
	history.Add("curl --request POST https://api.example.com/users --header 'Accept: application/json' --data '{\"name\":\"Ada\"}'")

	entry := history.Entries()[0]
	if entry.Label != "POST https://api.example.com/users" {
		t.Fatalf("label = %q", entry.Label)
	}
	if entry.Tooltip != "curl https://api.example.com/users POST" {
		t.Fatalf("tooltip = %q", entry.Tooltip)
	}
}

func TestCurlHistorySummaryInfersPost(t *testing.T) {
	history := NewCurlHistory(1)
	history.Add("curl https://api.example.com/users --unsupported-option value --data-raw name=ada")

	if got := history.Entries()[0].Tooltip; got != "curl https://api.example.com/users POST" {
		t.Fatalf("tooltip = %q", got)
	}
}
