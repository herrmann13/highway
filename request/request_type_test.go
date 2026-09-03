package request

import "testing"

func TestNormalizedRequestType(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"", HTTP},
		{"unknown", HTTP},
		{HTTP, HTTP},
		{GraphQL, GraphQL},
		{WebSocket, WebSocket},
		{GRPC, GRPC},
		{SSE, SSE},
	}

	for _, test := range tests {
		if got := NormalizedRequestType(test.value); got != test.want {
			t.Errorf("NormalizedRequestType(%q) = %q, want %q", test.value, got, test.want)
		}
		if icon := RequestTypeIcon(test.value); icon == nil {
			t.Errorf("RequestTypeIcon(%q) retornou nil", test.value)
		}
	}
}
