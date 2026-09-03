package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"highway/storage"
)

const exportBundleVersion = 1

type exportBundle struct {
	Version     int                   `json:"version"`
	Collections []*storage.Collection `json:"collections"`
}

func ExportCollections(path string, collections []*storage.Collection) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("não foi possível salvar a exportação: %w", err)
	}
	defer file.Close()
	if err := WriteExportBundle(file, collections); err != nil {
		return err
	}
	return nil
}

func WriteExportBundle(writer io.Writer, collections []*storage.Collection) error {
	bundle := exportBundle{Version: exportBundleVersion, Collections: collections}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("não foi possível preparar a exportação: %w", err)
	}
	if _, err := writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("não foi possível salvar a exportação: %w", err)
	}
	return nil
}

func ReadBundle(path string) (*exportBundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler o arquivo: %w", err)
	}
	return validateBundle(data)
}

func ReadBundleReader(reader io.Reader) (*exportBundle, error) {
	data, err := io.ReadAll(io.LimitReader(reader, 50*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler o arquivo: %w", err)
	}
	return validateBundle(data)
}

func validateBundle(data []byte) (*exportBundle, error) {
	var bundle exportBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("arquivo de exportação inválido: %w", err)
	}
	if bundle.Version != exportBundleVersion {
		return nil, fmt.Errorf("versão de exportação não suportada: %d", bundle.Version)
	}
	if len(bundle.Collections) == 0 {
		return nil, fmt.Errorf("o arquivo não contém coleções")
	}
	for _, c := range bundle.Collections {
		if c == nil || strings.TrimSpace(c.Name) == "" {
			return nil, fmt.Errorf("o arquivo contém uma coleção sem nome")
		}
	}
	return &bundle, nil
}

func UniqueCollectionName(collections []*storage.Collection, base string) string {
	name := strings.TrimSpace(base)
	if !collectionNameExists(collections, name) {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s (%d)", name, i)
		if !collectionNameExists(collections, candidate) {
			return candidate
		}
	}
}

func collectionNameExists(collections []*storage.Collection, name string) bool {
	for _, c := range collections {
		if c != nil && c.Name == name {
			return true
		}
	}
	return false
}
