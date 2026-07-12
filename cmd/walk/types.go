package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lxn/walk"

	"garq/internal/service"
)

// ---------------------------------------------------------------------------
// Tipos de dados
// ---------------------------------------------------------------------------

type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// TabPane encapsula todo o estado independente de uma aba.
type TabPane struct {
	tabPage          *walk.TabPage
	fileList         *walk.TableView
	pathEdit         *walk.LineEdit
	searchEdit       *walk.LineEdit
	previewImage     *walk.ImageView
	previewText      *walk.TextEdit
	previewLabel     *walk.Label
	previewVisible   bool
	previewComposite *walk.Composite
	fileModel        *FileTableModel
	allEntries       []FileEntry
	history          []string
	historyIdx       int
	sortBy           int
	sortDirAsc       bool
	columnTitles     []string
}

func (tp *TabPane) currentPath() string {
	if tp.pathEdit == nil {
		return ""
	}
	return tp.pathEdit.Text()
}

// GarqMainWindow é a janela principal.
type GarqMainWindow struct {
	*walk.MainWindow
	service     *service.Service
	navTree     *walk.TreeView
	tabWidget   *walk.TabWidget
	statusLabel *walk.Label
	navModel    *NavTreeModel
	tabs        []*TabPane
	btnBack     *walk.PushButton
	btnForward  *walk.PushButton
	btnUp       *walk.PushButton
	btnNewTab   *walk.PushButton
	btnCloseTab *walk.PushButton
}

func (mw *GarqMainWindow) activeTab() *TabPane {
	if mw.tabWidget == nil || len(mw.tabs) == 0 {
		return nil
	}
	idx := mw.tabWidget.CurrentIndex()
	if idx < 0 || idx >= len(mw.tabs) {
		return nil
	}
	return mw.tabs[idx]
}

// ---------------------------------------------------------------------------
// NavTree
// ---------------------------------------------------------------------------

type NavItem struct {
	text      string
	path      string
	parent    *NavItem
	children  []*NavItem
	isSection bool
}

func (item *NavItem) Text() string { return item.text }
func (item *NavItem) Image() interface{} {
	if item.path != "" {
		return item.path
	}
	if item.text == "Unidades" {
		return "C:\\"
	}
	return os.Getenv("USERPROFILE")
}
func (item *NavItem) Parent() walk.TreeItem {
	if item.parent == nil {
		return nil
	}
	return item.parent
}
func (item *NavItem) ChildCount() int { return len(item.children) }
func (item *NavItem) ChildAt(index int) walk.TreeItem {
	if index >= 0 && index < len(item.children) {
		return item.children[index]
	}
	return nil
}

type NavTreeModel struct {
	walk.TreeModelBase
	roots []*NavItem
}

func (m *NavTreeModel) RootCount() int { return len(m.roots) }
func (m *NavTreeModel) RootAt(index int) walk.TreeItem {
	if index >= 0 && index < len(m.roots) {
		return m.roots[index]
	}
	return nil
}
func (m *NavTreeModel) LazyPopulation() bool { return true }

// ---------------------------------------------------------------------------
// FileTableModel
// ---------------------------------------------------------------------------

type FileTableModel struct {
	walk.TableModelBase
	entries []FileEntry
}

func (m *FileTableModel) RowCount() int { return len(m.entries) }

func (m *FileTableModel) Image(index int) interface{} {
	if index >= len(m.entries) {
		return ""
	}
	return m.entries[index].Path
}

func (m *FileTableModel) Value(row, col int) interface{} {
	if row >= len(m.entries) {
		return ""
	}
	e := m.entries[row]
	switch col {
	case 0:
		return e.Name
	case 1:
		if e.IsDir {
			return "Pasta"
		}
		ext := filepath.Ext(e.Name)
		if ext != "" {
			return ext
		}
		return "Arquivo"
	case 2:
		if e.IsDir {
			return ""
		}
		return formatSize(e.Size)
	case 3:
		if e.ModTime.IsZero() {
			return ""
		}
		return e.ModTime.Format("02/01/2006 15:04:05")
	}
	return ""
}

func formatSize(size int64) string {
	switch {
	case size >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(size)/(1<<30))
	case size >= 1<<20:
		return fmt.Sprintf("%.2f MB", float64(size)/(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.2f KB", float64(size)/(1<<10))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
