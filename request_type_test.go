package main

import "testing"

func TestNormalizedRequestType(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"", requestTypeHTTP},
		{"unknown", requestTypeHTTP},
		{requestTypeHTTP, requestTypeHTTP},
		{requestTypeGraphQL, requestTypeGraphQL},
		{requestTypeWebSocket, requestTypeWebSocket},
		{requestTypeGRPC, requestTypeGRPC},
		{requestTypeSSE, requestTypeSSE},
	}

	for _, test := range tests {
		if got := normalizedRequestType(test.value); got != test.want {
			t.Errorf("normalizedRequestType(%q) = %q, want %q", test.value, got, test.want)
		}
		if icon := requestTypeIcon(test.value); icon == nil {
			t.Errorf("requestTypeIcon(%q) retornou nil", test.value)
		}
	}
}
