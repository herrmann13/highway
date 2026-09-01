package main

import "testing"

func TestInsertVariablePlaceholder(t *testing.T) {
	tests := []struct {
		text   string
		offset int
		name   string
		want   string
		cursor int
	}{
		{"users", 0, "base_url", "{{base_url}}users", 12},
		{"abcdef", 3, "id", "abc{{id}}def", 9},
		{"users", 5, "token", "users{{token}}", 14},
		{"ábc", 1, "id", "á{{id}}bc", 7},
	}

	for _, test := range tests {
		got, cursor := insertVariablePlaceholder(test.text, test.offset, test.name)
		if got != test.want || cursor != test.cursor {
			t.Errorf("insertVariablePlaceholder(%q, %d, %q) = (%q, %d), want (%q, %d)", test.text, test.offset, test.name, got, cursor, test.want, test.cursor)
		}
	}
}

func TestCursorPositionAtOffset(t *testing.T) {
	row, column := cursorPositionAtOffset("one\ntwo", 6)
	if row != 1 || column != 2 {
		t.Fatalf("posição = (%d, %d), want (1, 2)", row, column)
	}
}
