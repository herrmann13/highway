package importer

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"highway/curl"
	"highway/storage"
)

const maxCurlImportBytes = 2 * 1024 * 1024

type importMessage struct {
	Command string `json:"command"`
}

type importResponse struct {
	Error string `json:"error,omitempty"`
}

func ImportCommandFromArgs(args []string, input io.Reader) (string, bool, error) {
	if len(args) == 0 || args[0] != "import" {
		return "", false, nil
	}
	if len(args) != 2 || args[1] != "--curl-stdin" {
		return "", true, fmt.Errorf("uso: highway import --curl-stdin")
	}
	data, err := io.ReadAll(io.LimitReader(input, maxCurlImportBytes+1))
	if err != nil {
		return "", true, fmt.Errorf("erro ao ler cURL: %w", err)
	}
	if len(data) > maxCurlImportBytes {
		return "", true, fmt.Errorf("cURL excede o limite de %d MB", maxCurlImportBytes/(1024*1024))
	}
	command := strings.TrimSpace(string(data))
	if !isCurlCommand(command) {
		return "", true, fmt.Errorf("o texto selecionado não é um comando cURL válido")
	}
	return command, true, nil
}

func ImportFileFromArgs(args []string) (string, bool, error) {
	if len(args) == 0 || args[0] != "--import-file" {
		return "", false, nil
	}
	if len(args) != 2 || args[1] == "" {
		return "", true, fmt.Errorf("uso interno: highway --import-file <arquivo>")
	}
	data, err := os.ReadFile(args[1])
	_ = os.Remove(args[1])
	if err != nil {
		return "", true, fmt.Errorf("erro ao ler importação pendente: %w", err)
	}
	if len(data) > maxCurlImportBytes {
		return "", true, fmt.Errorf("cURL excede o limite de %d MB", maxCurlImportBytes/(1024*1024))
	}
	command := strings.TrimSpace(string(data))
	if !isCurlCommand(command) {
		return "", true, fmt.Errorf("o texto selecionado não é um comando cURL válido")
	}
	return command, true, nil
}

func LaunchHighwayWithImport(command string) error {
	configDir, err := storage.ConfigDir()
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(configDir, ".import-*.curl")
	if err != nil {
		return err
	}
	path := file.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if err = file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	if _, err = file.WriteString(command); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	appPath := "/Applications/Highway.app"
	if _, err = os.Stat(appPath); err != nil {
		return fmt.Errorf("Highway.app não encontrado em /Applications")
	}
	return exec.Command("/usr/bin/open", "-a", appPath, "--args", "--import-file", path).Start()
}

func isCurlCommand(command string) bool {
	args, err := curl.SplitShellArgs(command)
	return err == nil && len(args) > 0 && args[0] == "curl"
}

func importSocketPath() (string, error) {
	configDir, err := storage.ConfigDir()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(configDir))
	return filepath.Join(os.TempDir(), fmt.Sprintf("highway-%x.sock", digest[:8])), nil
}

func SendImportCommand(command string) (bool, error) {
	path, err := importSocketPath()
	if err != nil {
		return false, err
	}
	return sendImportCommandTo(path, command)
}

func sendImportCommandTo(path, command string) (bool, error) {
	conn, err := net.DialTimeout("unix", path, time.Second)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.ECONNREFUSED) {
			return false, nil
		}
		return false, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return false, err
	}
	if err := json.NewEncoder(conn).Encode(importMessage{Command: command}); err != nil {
		return false, err
	}
	var response importResponse
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return false, err
	}
	if response.Error != "" {
		return false, errors.New(response.Error)
	}
	return true, nil
}

func StartImportServer(handle func(string)) (func(), error) {
	path, err := importSocketPath()
	if err != nil {
		return nil, err
	}
	return startImportServerAt(path, handle)
}

func startImportServerAt(path string, handle func(string)) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		if conn, err := net.DialTimeout("unix", path, 250*time.Millisecond); err == nil {
			conn.Close()
			return nil, fmt.Errorf("outra instância do Highway já está em execução")
		}
		if err := os.Remove(path); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		listener.Close()
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleImportConnection(conn, handle)
		}
	}()

	return func() {
		listener.Close()
		_ = os.Remove(path)
	}, nil
}

func handleImportConnection(conn net.Conn, handle func(string)) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	var message importMessage
	err := json.NewDecoder(io.LimitReader(conn, maxCurlImportBytes+1)).Decode(&message)
	if err == nil {
		message.Command = strings.TrimSpace(message.Command)
		if len(message.Command) > maxCurlImportBytes {
			err = fmt.Errorf("cURL excede o limite de %d MB", maxCurlImportBytes/(1024*1024))
		} else if !isCurlCommand(message.Command) {
			err = errors.New("o texto selecionado não é um comando cURL válido")
		}
	}
	if err == nil {
		handle(message.Command)
	}
	response := importResponse{}
	if err != nil {
		response.Error = err.Error()
	}
	_ = json.NewEncoder(conn).Encode(response)
}
