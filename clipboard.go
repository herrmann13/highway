package main

import "strings"

type curlClipboardDetector struct {
	lastContent string
}

func (d *curlClipboardDetector) detect(content string) (string, bool) {
	content = strings.TrimSpace(content)
	if content == d.lastContent {
		return "", false
	}
	d.lastContent = content
	args, err := splitShellArgs(content)
	if err != nil || len(args) == 0 || args[0] != "curl" {
		return "", false
	}
	return content, true
}
