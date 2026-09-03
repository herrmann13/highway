package clipboard

import (
	"strings"

	"highway/curl"
)

type CurlClipboardDetector struct {
	lastContent string
}

func (d *CurlClipboardDetector) Detect(content string) (string, bool) {
	content = strings.TrimSpace(content)
	if content == d.lastContent {
		return "", false
	}
	d.lastContent = content
	args, err := curl.SplitShellArgs(content)
	if err != nil || len(args) == 0 || args[0] != "curl" {
		return "", false
	}
	return content, true
}
