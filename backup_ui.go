package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func collectionNames(collections []*collection) []string {
	names := make([]string, 0, len(collections))
	for _, c := range collections {
		if c != nil {
			names = append(names, c.Name)
		}
	}
	return names
}

func selectedCollections(collections []*collection, selected []string) []*collection {
	selectedSet := make(map[string]bool, len(selected))
	for _, name := range selected {
		selectedSet[name] = true
	}
	result := make([]*collection, 0, len(selected))
	for _, c := range collections {
		if c != nil && selectedSet[c.Name] {
			result = append(result, c)
		}
	}
	return result
}

func collectionSelectionContent(names []string) *widget.CheckGroup {
	group := widget.NewCheckGroup(names, nil)
	group.SetSelected(names)
	return group
}

func showExportCollectionsDialog(w fyne.Window, collections []*collection) {
	if len(collections) == 0 {
		dialog.ShowInformation("Exportar coleções", "Não há coleções para exportar.", w)
		return
	}
	group := collectionSelectionContent(collectionNames(collections))
	d := dialog.NewCustomConfirm("Exportar coleções", "Continuar", "Cancelar", container.NewVScroll(group), func(ok bool) {
		if !ok {
			return
		}
		selected := selectedCollections(collections, group.Selected)
		if len(selected) == 0 {
			dialog.ShowInformation("Exportar coleções", "Selecione pelo menos uma coleção.", w)
			return
		}
		fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			if err := writeExportBundle(writer, selected); err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("Exportar coleções", "Exportação concluída.", w)
		}, w)
		fileDialog.SetFileName("highway-collections.json")
		fileDialog.Show()
	}, w)
	d.Resize(fyne.NewSize(520, 420))
	d.Show()
}

func showImportCollectionsDialog(w fyne.Window, existing []*collection, onImported func([]*collection)) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		bundle, err := readBundleReader(reader)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		showImportSelectionDialog(w, existing, bundle.Collections, onImported)
	}, w)
	fileDialog.Show()
}

func showImportSelectionDialog(w fyne.Window, existing, incoming []*collection, onImported func([]*collection)) {
	names := collectionNames(incoming)
	group := collectionSelectionContent(names)
	d := dialog.NewCustomConfirm("Importar coleções", "Importar", "Cancelar", container.NewVScroll(group), func(ok bool) {
		if !ok {
			return
		}
		selected := selectedCollections(incoming, group.Selected)
		if len(selected) == 0 {
			dialog.ShowInformation("Importar coleções", "Selecione pelo menos uma coleção.", w)
			return
		}
		imported := make([]*collection, 0, len(selected))
		for _, c := range selected {
			copy := *c
			copy.Name = uniqueCollectionName(existing, c.Name)
			existing = append(existing, &copy)
			imported = append(imported, &copy)
		}
		if onImported != nil {
			onImported(imported)
		}
	}, w)
	d.Resize(fyne.NewSize(520, 420))
	d.Show()
}

func saveImportedCollections(collections []*collection) error {
	for _, c := range collections {
		if c == nil || strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("coleção importada sem nome")
		}
		if err := saveCollection(c); err != nil {
			return err
		}
	}
	return nil
}
