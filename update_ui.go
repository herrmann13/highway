package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func checkForUpdates(a fyne.App, w fyne.Window, button *widget.Button) {
	button.Disable()
	button.SetText("Verificando atualizações...")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
		defer cancel()
		release, err := fetchLatestRelease(ctx)
		current, currentErr := parseSemanticVersion(appVersion)
		latest, latestErr := parseSemanticVersion(release.TagName)
		fyne.Do(func() {
			button.Enable()
			button.SetText("Verificar atualizações")
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if currentErr != nil || latestErr != nil {
				dialog.ShowError(fmt.Errorf("não foi possível comparar as versões"), w)
				return
			}
			if !latest.newerThan(current) {
				dialog.ShowInformation("Atualizações", "Você já está usando a versão mais recente (v"+appVersion+").", w)
				return
			}
			asset, err := releaseAssetForPlatform(release, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			message := "A versão " + release.TagName + " está disponível.\n\nDeseja baixar e instalar agora?"
			if notes := strings.TrimSpace(release.Body); notes != "" {
				message += "\n\nNotas da release:\n" + truncateUpdateNotes(notes)
			}
			dialog.ShowConfirm("Atualização disponível", message, func(ok bool) {
				if ok {
					downloadAndInstallUpdate(a, w, button, release, asset)
				}
			}, w)
		})
	}()
}

func downloadAndInstallUpdate(a fyne.App, w fyne.Window, button *widget.Button, release githubRelease, asset releaseAsset) {
	button.Disable()
	button.SetText("Baixando atualização...")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*updateTimeout)
		defer cancel()
		path, err := downloadAsset(ctx, asset)
		if err == nil {
			err = verifyAssetChecksum(ctx, release, asset, path)
		}
		if err == nil {
			_, err = installUpdate(path)
		}
		if err != nil {
			_ = os.Remove(path)
		}
		fyne.Do(func() {
			button.Enable()
			button.SetText("Verificar atualizações")
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if runtime.GOOS == "darwin" {
				a.Quit()
				return
			}
			dialog.ShowInformation("Atualização instalada", "A atualização foi instalada. Feche e abra o Highway novamente para usar a nova versão.", w)
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
