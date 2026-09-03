package backup

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"highway/storage"
)

func CollectionNames(collections []*storage.Collection) []string {
	names := make([]string, 0, len(collections))
	for _, c := range collections {
		if c != nil {
			names = append(names, c.Name)
		}
	}
	return names
}

func SelectedCollections(collections []*storage.Collection, selected []string) []*storage.Collection {
	selectedSet := make(map[string]bool, len(selected))
	for _, name := range selected {
		selectedSet[name] = true
	}
	result := make([]*storage.Collection, 0, len(selected))
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

func ShowExportCollectionsDialog(w fyne.Window, collections []*storage.Collection) {
	if len(collections) == 0 {
		dialog.ShowInformation("Exportar coleções", "Não há coleções para exportar.", w)
		return
	}
	group := collectionSelectionContent(CollectionNames(collections))
	d := dialog.NewCustomConfirm("Exportar coleções", "Continuar", "Cancelar", container.NewVScroll(group), func(ok bool) {
		if !ok {
			return
		}
		selected := SelectedCollections(collections, group.Selected)
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
			if err := WriteExportBundle(writer, selected); err != nil {
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

func ShowImportCollectionsDialog(w fyne.Window, existing []*storage.Collection, onImported func([]*storage.Collection)) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		bundle, err := ReadBundleReader(reader)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		showImportSelectionDialog(w, existing, bundle.Collections, onImported)
	}, w)
	fileDialog.Show()
}

func showImportSelectionDialog(w fyne.Window, existing, incoming []*storage.Collection, onImported func([]*storage.Collection)) {
	names := CollectionNames(incoming)
	group := collectionSelectionContent(names)
	d := dialog.NewCustomConfirm("Importar coleções", "Importar", "Cancelar", container.NewVScroll(group), func(ok bool) {
		if !ok {
			return
		}
		selected := SelectedCollections(incoming, group.Selected)
		if len(selected) == 0 {
			dialog.ShowInformation("Importar coleções", "Selecione pelo menos uma coleção.", w)
			return
		}
		imported := make([]*storage.Collection, 0, len(selected))
		for _, c := range selected {
			copy := *c
			copy.Name = UniqueCollectionName(existing, c.Name)
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

func SaveImportedCollections(collections []*storage.Collection) error {
	for _, c := range collections {
		if c == nil || strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("coleção importada sem nome")
		}
		if err := storage.SaveCollection(c); err != nil {
			return err
		}
	}
	return nil
}
