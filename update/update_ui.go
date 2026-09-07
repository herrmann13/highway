package update

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"highway/i18n"
	"highway/version"
)

func CheckForUpdates(a fyne.App, w fyne.Window, button *widget.Button) {
	button.Disable()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
		defer cancel()
		release, err := FetchLatestRelease(ctx)
		current, currentErr := ParseSemanticVersion(version.AppVersion)
		latest, latestErr := ParseSemanticVersion(release.TagName)
		fyne.Do(func() {
			button.Enable()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if currentErr != nil || latestErr != nil {
				dialog.ShowError(fmt.Errorf("%s", i18n.T("update.compare")), w)
				return
			}
			if !latest.NewerThan(current) {
				dialog.ShowInformation(i18n.T("update.title"), i18n.Tf("update.upToDate", version.AppVersion), w)
				return
			}
			asset, err := ReleaseAssetForPlatform(release, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			message := i18n.Tf("update.available", release.TagName)
			if notes := strings.TrimSpace(release.Body); notes != "" {
				message += "\n\n" + i18n.T("update.notes") + truncateUpdateNotes(notes)
			}
			dialog.ShowConfirm(i18n.T("update.availableTitle"), message, func(ok bool) {
				if ok {
					DownloadAndInstallUpdate(a, w, button, release, asset)
				}
			}, w)
		})
	}()
}

func DownloadAndInstallUpdate(a fyne.App, w fyne.Window, button *widget.Button, release GithubRelease, asset ReleaseAsset) {
	button.Disable()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*updateTimeout)
		defer cancel()
		path, err := DownloadAsset(ctx, asset)
		if err == nil {
			err = VerifyAssetChecksum(ctx, release, asset, path)
		}
		if err == nil {
			_, err = InstallUpdate(path)
		}
		if err != nil {
			_ = os.Remove(path)
		}
		fyne.Do(func() {
			button.Enable()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if runtime.GOOS == "darwin" {
				a.Quit()
				return
			}
			dialog.ShowInformation(i18n.T("update.installedTitle"), i18n.T("update.installed"), w)
		})
	}()
}

func truncateUpdateNotes(notes string) string {
	const maxRunes = 1200
	runes := []rune(notes)
	if len(runes) <= maxRunes {
		return notes
	}
	return string(runes[:maxRunes]) + "\n..."
}
