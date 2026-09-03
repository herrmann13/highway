package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallMacOSService(t *testing.T) {
	destination := filepath.Join(t.TempDir(), macOSServiceWorkflowName)
	binaryPath := "/Applications/Highway.app/Contents/MacOS/Highway"

	changed, err := installMacOSService(destination, binaryPath)
	if err != nil || !changed {
		t.Fatalf("primeira instalação = (%t, %v), quer (true, nil)", changed, err)
	}

	workflowPath := filepath.Join(destination, "Contents", "document.wflow")
	workflow, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(workflow, []byte(binaryPath)) || bytes.Contains(workflow, []byte(macOSServiceBinaryToken)) {
		t.Fatalf("caminho do binário não foi aplicado: %s", workflow)
	}
	infoPath := filepath.Join(destination, "Contents", "Info.plist")
	info, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Info.plist não foi instalado: %v", err)
	}
	expectedInfo, err := macOSServiceTemplate.ReadFile(filepath.Join(macOSServiceTemplateDir, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(info, expectedInfo) {
		t.Fatal("Info.plist instalado não corresponde ao template")
	}

	changed, err = installMacOSService(destination, binaryPath)
	if err != nil || changed {
		t.Fatalf("instalação idempotente = (%t, %v), quer (false, nil)", changed, err)
	}

	updatedBinaryPath := "/Applications/Highway.app/Contents/MacOS/Highway-next"
	changed, err = installMacOSService(destination, updatedBinaryPath)
	if err != nil || !changed {
		t.Fatalf("atualização do caminho = (%t, %v), quer (true, nil)", changed, err)
	}
	workflow, err = os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(workflow, []byte(updatedBinaryPath)) {
		t.Fatalf("workflow não foi atualizado: %s", workflow)
	}
}
