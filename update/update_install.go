package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func InstallUpdate(path string) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return installLinuxUpdate(path)
	case "darwin":
		return installMacUpdate(path)
	default:
		return "", fmt.Errorf("atualização não suportada em %s", runtime.GOOS)
	}
}

func installLinuxUpdate(path string) (string, error) {
	if filepath.Ext(path) != ".deb" {
		return "", fmt.Errorf("o instalador Linux deve ser um arquivo .deb")
	}
	if _, err := exec.LookPath("pkexec"); err != nil {
		return "", fmt.Errorf("pkexec não está disponível para autorizar a instalação")
	}
	aptPath, err := exec.LookPath("apt")
	if err != nil {
		return "", fmt.Errorf("apt não está disponível para instalar a atualização")
	}
	if output, err := exec.Command("pkexec", aptPath, "install", "-y", path).CombinedOutput(); err != nil {
		return "", fmt.Errorf("não foi possível instalar a atualização: %w: %s", err, strings.TrimSpace(string(output)))
	}
	_ = os.Remove(path)
	return "Atualização instalada. Feche e abra o Highway novamente para usar a nova versão.", nil
}

func installMacUpdate(dmgPath string) (string, error) {
	if filepath.Ext(dmgPath) != ".dmg" {
		return "", fmt.Errorf("o instalador macOS deve ser um arquivo .dmg")
	}
	appPath, err := currentMacAppPath()
	if err != nil {
		return "", err
	}
	script, err := os.CreateTemp("", "highway-update-*.sh")
	if err != nil {
		return "", err
	}
	scriptPath := script.Name()
	if _, err := script.WriteString(`#!/bin/sh
while kill -0 "$1" 2>/dev/null; do sleep 1; done
tmp_dir=$(mktemp -d)
mount_dir="$tmp_dir/mount"
mkdir "$mount_dir"
/usr/bin/hdiutil attach -readonly -nobrowse -mountpoint "$mount_dir" "$2"
rm -rf "$3"
cp -R "$mount_dir/Highway.app" "$3"
/usr/bin/hdiutil detach "$mount_dir"
console_user=$(/usr/bin/stat -f %Su /dev/console)
if [ -n "$console_user" ] && [ "$console_user" != "root" ]; then
  /usr/bin/sudo -u "$console_user" /usr/bin/open "$3"
fi
rm -rf "$tmp_dir" "$2" "$0"
`); err != nil {
		script.Close()
		_ = os.Remove(scriptPath)
		return "", err
	}
	if err := script.Close(); err != nil {
		_ = os.Remove(scriptPath)
		return "", err
	}
	if err := os.Chmod(scriptPath, 0o700); err != nil {
		_ = os.Remove(scriptPath)
		return "", err
	}

	command := "/bin/sh " + shellQuote(scriptPath) + " " + strconv.Itoa(os.Getpid()) + " " + shellQuote(dmgPath) + " " + shellQuote(appPath) + " >/dev/null 2>&1 &"
	if err := exec.Command("/usr/bin/osascript", "-e", "do shell script "+strconv.Quote(command)+" with administrator privileges").Run(); err != nil {
		_ = os.Remove(scriptPath)
		return "", fmt.Errorf("não foi possível iniciar o atualizador: %w", err)
	}
	return "A atualização será instalada após o Highway ser fechado.", nil
}

func currentMacAppPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	marker := ".app/Contents/MacOS/"
	index := strings.Index(executable, marker)
	if index < 0 {
		return "", fmt.Errorf("instale o Highway em Aplicativos antes de usar a atualização automática")
	}
	return executable[:index+len(".app")], nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
