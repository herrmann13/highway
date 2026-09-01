package main

import "fyne.io/fyne/v2"

const (
	requestTypeHTTP      = "http"
	requestTypeGraphQL   = "graphql"
	requestTypeWebSocket = "websocket"
	requestTypeGRPC      = "grpc"
	requestTypeSSE       = "sse"
)

var requestTypeIcons = map[string]fyne.Resource{
	requestTypeHTTP:      fyne.NewStaticResource("request-http.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#40a9ff" d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm6.9 6h-3.1a15.8 15.8 0 0 0-1.5-3.4A8.1 8.1 0 0 1 18.9 8ZM12 4c.8 1.1 1.5 2.5 1.9 4h-3.8c.4-1.5 1.1-2.9 1.9-4ZM4.3 14a8.3 8.3 0 0 1 0-4h3.4a16.5 16.5 0 0 0 0 4H4.3Zm.8 2h3.1a15.8 15.8 0 0 0 1.5 3.4A8.1 8.1 0 0 1 5.1 16Zm3.1-8H5.1a8.1 8.1 0 0 1 4.6-3.4A15.8 15.8 0 0 0 8.2 8Zm3.8 12c-.8-1.1-1.5-2.5-1.9-4h3.8c-.4 1.5-1.1 2.9-1.9 4Zm2.4-6h-4.8a14.5 14.5 0 0 1 0-4h4.8a14.5 14.5 0 0 1 0 4Zm-.1 5.4a15.8 15.8 0 0 0 1.5-3.4h3.1a8.1 8.1 0 0 1-4.6 3.4Zm1.9-5.4a16.5 16.5 0 0 0 0-4h3.4a8.3 8.3 0 0 1 0 4h-3.4Z"/></svg>`)),
	requestTypeGraphQL:   fyne.NewStaticResource("request-graphql.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#e535ab" d="m12 2 8.7 5v10L12 22l-8.7-5V7L12 2Zm0 2.3L5.3 8.1v7.8l6.7 3.8 6.7-3.8V8.1L12 4.3Zm-1 3.2h2v3.5l3 1.7-1 1.7-3-1.8-3 1.8-1-1.7 3-1.7V7.5Z"/></svg>`)),
	requestTypeWebSocket: fyne.NewStaticResource("request-websocket.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#8b5cf6" d="M7 4h10v2H7V4Zm0 14h10v2H7v-2Zm7.7-10.3L18 11l-3.3 3.3-1.4-1.4 1.9-1.9-1.9-1.9 1.4-1.4ZM9.3 7.7 6 11l3.3 3.3 1.4-1.4L8.8 11l1.9-1.9-1.4-1.4Z"/></svg>`)),
	requestTypeGRPC:      fyne.NewStaticResource("request-grpc.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#22c55e" d="M7 4a3 3 0 1 1 0 6 3 3 0 0 1 0-6Zm10 10a3 3 0 1 1 0 6 3 3 0 0 1 0-6ZM7 12h2v2H7v-2Zm4 0h2v2h-2v-2Zm4-2h2v2h-2v-2Zm-6.6-1.5 5.2 5.2-1.4 1.4L7 9.9l1.4-1.4Z"/></svg>`)),
	requestTypeSSE:       fyne.NewStaticResource("request-sse.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#f59e0b" d="M4 6h2v12H4V6Zm4-2h2v16H8V4Zm4 4h2v8h-2V8Zm4-3h2v14h-2V5Z"/></svg>`)),
}

func normalizedRequestType(value string) string {
	switch value {
	case requestTypeGraphQL, requestTypeWebSocket, requestTypeGRPC, requestTypeSSE:
		return value
	default:
		return requestTypeHTTP
	}
}

func requestTypeIcon(value string) fyne.Resource {
	return requestTypeIcons[normalizedRequestType(value)]
}
