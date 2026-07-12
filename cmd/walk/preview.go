package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
)

func (mw *GarqMainWindow) updatePreview() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	if !tp.previewVisible {
		return
	}
	idx := tp.fileList.CurrentIndex()
	if idx < 0 || idx >= len(tp.fileModel.entries) {
		tp.previewLabel.SetText("Pré-visualização")
		tp.previewImage.SetImage(nil)
		tp.previewText.SetText("")
		return
	}
	entry := tp.fileModel.entries[idx]

	if entry.IsDir {
		tp.previewLabel.SetText(fmt.Sprintf("Pasta: %s", entry.Name))
		tp.previewImage.SetImage(nil)
		tp.previewText.SetText("")
		return
	}

	ext := strings.ToLower(filepath.Ext(entry.Name))
	tp.previewLabel.SetText(fmt.Sprintf("%s — %s", entry.Name, formatSize(entry.Size)))

	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true, ".gif": true, ".ico": true, ".tiff": true}
	textExts := map[string]bool{
		".txt": true, ".md": true, ".go": true, ".js": true, ".ts": true, ".py": true,
		".json": true, ".xml": true, ".html": true, ".css": true, ".csv": true, ".log": true,
		".ini": true, ".cfg": true, ".yaml": true, ".yml": true, ".toml": true,
		".sh": true, ".bat": true, ".cmd": true,
	}

	if imageExts[ext] {
		img, err := walk.NewImageFromFile(entry.Path)
		if err == nil {
			tp.previewImage.SetImage(img)
			tp.previewText.SetText("")
		} else {
			tp.previewImage.SetImage(nil)
			tp.previewText.SetText(fmt.Sprintf("Erro ao carregar imagem: %v", err))
		}
		return
	}

	tp.previewImage.SetImage(nil)

	if textExts[ext] && entry.Size < 1<<20 {
		data, err := os.ReadFile(entry.Path)
		if err == nil {
			tp.previewText.SetText(string(data))
		} else {
			tp.previewText.SetText(fmt.Sprintf("Erro ao ler arquivo: %v", err))
		}
		return
	}

	tp.previewText.SetText(fmt.Sprintf("Arquivo: %s\nTamanho: %s\nModificado: %s\nTipo: %s",
		entry.Name, formatSize(entry.Size), entry.ModTime, ext))
}
