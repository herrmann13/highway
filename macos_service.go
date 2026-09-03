package main

import (
	"bytes"
	"embed"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	macOSServiceWorkflowName = "Abrir no Highway.workflow"
	macOSServiceTemplateDir  = "packaging/macos/Abrir no Highway.workflow"
	macOSServiceBinaryToken  = "__HIGHWAY_BIN__"
)

//go:embed "packaging/macos/Abrir no Highway.workflow"
var macOSServiceTemplate embed.FS

func ensureMacOSService() {
	if runtime.GOOS != "darwin" {
		return
	}

	binaryPath, err := os.Executable()
	if err != nil {
		return
	}
	appPath := filepath.Dir(filepath.Dir(filepath.Dir(binaryPath)))
	if appPath != "/Applications/Highway.app" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	destination := filepath.Join(home, "Library", "Services", macOSServiceWorkflowName)
	changed, err := installMacOSService(destination, binaryPath)
	if err != nil || !changed {
		return
	}
	_ = exec.Command("killall", "pbs").Run()
}

func installMacOSService(destination, binaryPath string) (bool, error) {
	workflow, err := macOSServiceTemplate.ReadFile(filepath.Join(macOSServiceTemplateDir, "Contents", "document.wflow"))
	if err != nil {
		return false, err
	}
	if !bytes.Contains(workflow, []byte(macOSServiceBinaryToken)) {
		return false, errors.New("template do serviço macOS sem caminho do Highway")
	}
	workflow = bytes.ReplaceAll(workflow, []byte(macOSServiceBinaryToken), []byte(binaryPath))

	info, err := macOSServiceTemplate.ReadFile(filepath.Join(macOSServiceTemplateDir, "Contents", "Info.plist"))
	if err != nil {
		return false, err
	}

	files := map[string][]byte{
		"Contents/document.wflow": workflow,
		"Contents/Info.plist":     info,
	}
	changed := false
	for relativePath, expected := range files {
		current, err := os.ReadFile(filepath.Join(destination, relativePath))
		if err != nil && !os.IsNotExist(err) {
			return false, err
		}
		if !bytes.Equal(current, expected) {
			changed = true
		}
	}
	if !changed {
		return false, nil
	}

	for relativePath, content := range files {
		path := filepath.Join(destination, relativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return false, err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return false, err
		}
	}
	return true, nil
}
