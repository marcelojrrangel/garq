package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

func (mw *GarqMainWindow) navigateTo(path string) {
	tp := mw.activeTab()
	if tp == nil {
		log.Println("navigateTo: aba ativa é nil")
		return
	}
	log.Printf("navigateTo: %s", path)
	mw.navigateTabTo(tp, path)
}

func (mw *GarqMainWindow) navigateTabTo(tp *TabPane, path string) {
	log.Printf("navigateTabTo: %s", path)
	mw.Synchronize(func() {
		tp.pathEdit.SetText(path)
		mw.statusLabel.SetText("Carregando...")
	})

	entries, err := mw.service.ListDirectory(path)
	if err != nil {
		log.Printf("navigateTabTo erro ListDirectory: %v", err)
		mw.Synchronize(func() { mw.statusLabel.SetText(fmt.Sprintf("Erro: %v", err)) })
		return
	}

	fileEntries := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		fileEntries = append(fileEntries, FileEntry{Name: e.Name, Path: e.Path, IsDir: e.IsDir, Size: e.Size, ModTime: e.ModTime})
	}

	if len(tp.history) == 0 || tp.history[len(tp.history)-1] != path {
		tp.history = append(tp.history[:tp.historyIdx+1], path)
		tp.historyIdx = len(tp.history) - 1
	}

	title := tabTitle(path)

	tp.fileModel.entries = fileEntries
	tp.allEntries = fileEntries
	mw.sortTab(tp)

	mw.Synchronize(func() {
		tp.tabPage.SetTitle(title)
		if icon := getShellIcon(path); icon != nil {
			tp.tabPage.SetImage(icon)
		}
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", len(fileEntries)))
		if tp.searchEdit != nil {
			tp.searchEdit.SetText("")
		}
	})
	mw.updateNavButtons()
}

// navigateTabDirect carrega o diretório sem alterar o histórico (usado por goBack/goForward).
func (mw *GarqMainWindow) navigateTabDirect(tp *TabPane, path string) {
	mw.Synchronize(func() {
		tp.pathEdit.SetText(path)
		mw.statusLabel.SetText("Carregando...")
	})

	entries, err := mw.service.ListDirectory(path)
	if err != nil {
		mw.Synchronize(func() { mw.statusLabel.SetText(fmt.Sprintf("Erro: %v", err)) })
		return
	}

	fileEntries := make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		fileEntries = append(fileEntries, FileEntry{Name: e.Name, Path: e.Path, IsDir: e.IsDir, Size: e.Size, ModTime: e.ModTime})
	}

	title := tabTitle(path)

	tp.fileModel.entries = fileEntries
	tp.allEntries = fileEntries
	mw.sortTab(tp)

	mw.Synchronize(func() {
		tp.tabPage.SetTitle(title)
		if icon := getShellIcon(path); icon != nil {
			tp.tabPage.SetImage(icon)
		}
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", len(fileEntries)))
		if tp.searchEdit != nil {
			tp.searchEdit.SetText("")
		}
	})
	mw.updateNavButtons()
}

func (mw *GarqMainWindow) goBack() {
	tp := mw.activeTab()
	if tp == nil {
		log.Println("goBack: aba ativa é nil")
		return
	}
	if tp.historyIdx <= 0 {
		log.Printf("goBack: sem histórico (idx=%d)", tp.historyIdx)
		return
	}
	tp.historyIdx--
	path := tp.history[tp.historyIdx]
	log.Printf("goBack: %s (idx=%d)", path, tp.historyIdx)
	mw.navigateTabDirect(tp, path)
	mw.updateNavButtons()
}

func (mw *GarqMainWindow) goForward() {
	tp := mw.activeTab()
	if tp == nil {
		log.Println("goForward: aba ativa é nil")
		return
	}
	if tp.historyIdx >= len(tp.history)-1 {
		log.Printf("goForward: sem histórico futuro (idx=%d, len=%d)", tp.historyIdx, len(tp.history))
		return
	}
	tp.historyIdx++
	path := tp.history[tp.historyIdx]
	log.Printf("goForward: %s (idx=%d)", path, tp.historyIdx)
	mw.navigateTabDirect(tp, path)
	mw.updateNavButtons()
}

func (mw *GarqMainWindow) goUp() {
	tp := mw.activeTab()
	if tp == nil {
		log.Println("goUp: aba ativa é nil")
		return
	}
	current := tp.currentPath()
	if current == "" {
		log.Println("goUp: caminho vazio")
		return
	}
	parent := filepath.Dir(current)
	if parent != current {
		log.Printf("goUp: %s -> %s", current, parent)
		mw.navigateTo(parent)
	}
}

func (mw *GarqMainWindow) updateNavButtons() {
	tp := mw.activeTab()
	if tp == nil || mw.btnBack == nil || mw.btnForward == nil {
		return
	}
	mw.btnBack.SetEnabled(tp.historyIdx > 0)
	mw.btnForward.SetEnabled(tp.historyIdx < len(tp.history)-1)
}

func (mw *GarqMainWindow) updateStatusBar() {
	tp := mw.activeTab()
	if tp == nil {
		mw.statusLabel.SetText("Pronto")
		return
	}
	total := len(tp.fileModel.entries)
	sel := len(tp.fileList.SelectedIndexes())
	if sel > 0 {
		mw.statusLabel.SetText(fmt.Sprintf("%d itens, %d selecionados", total, sel))
	} else {
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", total))
	}
}

// updateActiveJobsStatus mostra na statusbar quantos jobs estão rodando.
func (mw *GarqMainWindow) updateActiveJobsStatus() {
	count, err := mw.service.ActiveJobCount()
	if err != nil {
		return
	}
	tp := mw.activeTab()
	base := "Pronto"
	if tp != nil {
		total := len(tp.fileModel.entries)
		base = fmt.Sprintf("%d itens", total)
	}
	if count > 0 {
		mw.statusLabel.SetText(fmt.Sprintf("%s  ·  ⏳ %d operação(ões) em andamento", base, count))
	} else {
		mw.statusLabel.SetText(base)
	}
}

// refreshActiveTab recarrega o diretório atual da aba ativa (equivalente ao F5).
func (mw *GarqMainWindow) refreshActiveTab() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	path := tp.currentPath()
	if path == "" {
		return
	}
	go mw.navigateTabTo(tp, path)
}

// openProgressDialog abre um dialog de progresso para um job recém-criado.
func (mw *GarqMainWindow) openProgressDialog(jobID int64, jobType string) {
	newProgressDialog(mw, jobID, jobType)
}

func (mw *GarqMainWindow) filterBySearch() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	query := tp.searchEdit.Text()
	if query == "" {
		tp.fileModel.entries = tp.allEntries
	} else {
		var filtered []FileEntry
		for _, e := range tp.allEntries {
			if containsIgnoreCase(e.Name, query) {
				filtered = append(filtered, e)
			}
		}
		tp.fileModel.entries = filtered
	}
	tp.fileList.SetModel(tp.fileModel)
	mw.updateStatusBar()
}

func containsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func (mw *GarqMainWindow) sortTab(tp *TabPane) {
	asc := tp.sortDirAsc
	for _, entries := range [][]FileEntry{tp.fileModel.entries, tp.allEntries} {
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			var less bool
			switch tp.sortBy {
			case 1:
				extI := filepath.Ext(entries[i].Name)
				extJ := filepath.Ext(entries[j].Name)
				less = extI < extJ
			case 2:
				less = entries[i].Size < entries[j].Size
			case 3:
				less = entries[i].ModTime.Before(entries[j].ModTime)
			default:
				less = entries[i].Name < entries[j].Name
			}
			if asc {
				return less
			}
			return !less
		})
	}

	cols := tp.fileList.Columns()
	for i, base := range tp.columnTitles {
		title := base
		if i == tp.sortBy {
			if asc {
				title += " ▲"
			} else {
				title += " ▼"
			}
		}
		if i < cols.Len() {
			cols.At(i).SetTitle(title)
		}
	}

	mw.Synchronize(func() { tp.fileList.SetModel(tp.fileModel) })
}
