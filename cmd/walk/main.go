package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"garq/internal/api"
	"garq/internal/db"
	"garq/internal/worker"
)

// ---------------------------------------------------------------------------
// Tipos de dados
// ---------------------------------------------------------------------------

type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
}

// TabPane encapsula todo o estado independente de uma aba.
type TabPane struct {
	tabPage    *walk.TabPage
	fileList   *walk.TableView
	pathEdit   *walk.LineEdit
	searchEdit *walk.LineEdit
	previewImage *walk.ImageView
	previewText  *walk.TextEdit
	previewLabel *walk.Label
	fileModel  *FileTableModel
	allEntries []FileEntry
	history    []string
	historyIdx int
	sortBy     int
	sortDirAsc bool
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
	api         *api.API
	navTree     *walk.TreeView
	tabWidget   *walk.TabWidget
	statusLabel *walk.Label
	navModel    *NavTreeModel
	tabs        []*TabPane
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
	text     string
	path     string
	parent   *NavItem
	children []*NavItem
}

func (item *NavItem) Text() string { return item.text }
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
		return e.ModTime
	case 2:
		if e.IsDir {
			return "Pasta"
		}
		ext := filepath.Ext(e.Name)
		if ext != "" {
			return ext
		}
		return "Arquivo"
	case 3:
		if e.IsDir {
			return ""
		}
		return formatSize(e.Size)
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


// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	dbPath := "garq.db"
	if envPath := os.Getenv("GARQ_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	dbConn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco: %v", err)
	}
	defer dbConn.Close()

	worker.StartWorkerPool(4, dbConn)
	apiInstance := &api.API{DB: dbConn, Ctx: context.Background()}

	navModel := &NavTreeModel{}
	buildNavTree(navModel)

	mw := &GarqMainWindow{
		api:      apiInstance,
		navModel: navModel,
	}

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Garq - Gerenciador de Arquivos",
		MinSize:  Size{Width: 1100, Height: 650},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			// ---- Toolbar global ----
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{Text: "<", MinSize: Size{Width: 30}, OnClicked: func() { mw.goBack() }},
					PushButton{Text: ">", MinSize: Size{Width: 30}, OnClicked: func() { mw.goForward() }},
					PushButton{Text: "^", MinSize: Size{Width: 30}, OnClicked: func() { mw.goUp() }},
					PushButton{Text: "+ Aba", MinSize: Size{Width: 55}, OnClicked: func() { mw.newTab("") }},
					PushButton{Text: "x Aba", MinSize: Size{Width: 55}, OnClicked: func() { mw.closeCurrentTab() }},
				},
			},
			// ---- Conteúdo: navTree + tabWidget ----
			HSplitter{
				Children: []Widget{
					TreeView{
						AssignTo:             &mw.navTree,
						Model:                mw.navModel,
						OnCurrentItemChanged: mw.onNavItemSelected,
						OnItemActivated:      mw.onNavItemActivated,
					},
					TabWidget{
						AssignTo: &mw.tabWidget,
						OnCurrentIndexChanged: func() {
							mw.onTabChanged()
						},
					},
				},
			},
			Label{AssignTo: &mw.statusLabel, Text: "Pronto"},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// Atalhos globais de teclado na janela principal
	mw.KeyDown().Attach(func(key walk.Key) {
		mods := walk.ModifiersDown()
		switch {
		case key == walk.KeyT && mods&walk.ModControl != 0:
			mw.newTab("")
		case key == walk.KeyW && mods&walk.ModControl != 0:
			mw.closeCurrentTab()
		case key == walk.KeyTab && mods&walk.ModControl != 0:
			mw.nextTab()
		}
	})

	// Primeira aba
	mw.newTab("")

	// Drag & drop
	mw.DropFiles().Attach(func(files []string) {
		tp := mw.activeTab()
		if tp == nil {
			return
		}
		dest := tp.currentPath()
		if dest == "" {
			return
		}
		count := 0
		for _, src := range files {
			baseName := filepath.Base(src)
			dst := filepath.Join(dest, baseName)
			info, err := os.Stat(src)
			if err != nil {
				continue
			}
			if info.IsDir() {
				if err := copyPath(src, dst); err == nil {
					count++
				}
			} else {
				if src == dst {
					dst = getCopyPath(dest, baseName)
				}
				if err := copyPath(src, dst); err == nil {
					count++
				}
			}
		}
		mw.statusLabel.SetText(fmt.Sprintf("Importado(s) %d item(s)", count))
		mw.navigateTo(dest)
	})

	// Navega para o primeiro drive ao iniciar
	go func() {
		roots, err := mw.api.ListRoots()
		if err != nil {
			log.Printf("Erro ao listar drives: %v", err)
			return
		}
		if len(roots) > 0 {
			mw.Synchronize(func() { mw.navigateTo(roots[0]) })
		}
	}()

	// Workaround lxn/walk: força relayout inicial
	go func() {
		<-time.After(100 * time.Millisecond)
		mw.Synchronize(func() {
			b := mw.Bounds()
			mw.SetBounds(walk.Rectangle{X: b.X, Y: b.Y, Width: b.Width + 1, Height: b.Height})
			mw.SetBounds(b)
		})
	}()

	mw.Run()
}

// ---------------------------------------------------------------------------
// Gerenciamento de abas
// ---------------------------------------------------------------------------

func (mw *GarqMainWindow) newTab(initialPath string) {
	tabIdx := len(mw.tabs)
	title := fmt.Sprintf("Aba %d", tabIdx+1)

	tp := &TabPane{
		fileModel:  &FileTableModel{},
		history:    make([]string, 0),
		historyIdx: -1,
	}

	// Cria o TabPage e adiciona ao TabWidget
	tabPage, err := walk.NewTabPage()
	if err != nil {
		log.Printf("Erro ao criar TabPage: %v", err)
		return
	}
	tabPage.SetTitle(title)
	tabPage.SetLayout(walk.NewVBoxLayout())

	if err := mw.tabWidget.Pages().Add(tabPage); err != nil {
		log.Printf("Erro ao adicionar TabPage: %v", err)
		return
	}

	// Constrói os widgets dentro do tabPage usando declarative
	var composite *walk.Composite
	builder := NewBuilder(tabPage)
	if err := (Composite{
		AssignTo: &composite,
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			// Barra de endereço + busca
			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{
						AssignTo: &tp.pathEdit,
						OnKeyDown: func(key walk.Key) {
							if key == walk.KeyReturn {
								if path := tp.pathEdit.Text(); path != "" {
									mw.navigateTo(path)
								}
							}
						},
					},
					Label{Text: "Buscar:"},
					LineEdit{
						AssignTo:      &tp.searchEdit,
						CueBanner:     "Filtrar...",
						OnTextChanged: func() { mw.filterBySearch() },
					},
				},
			},
			// Toolbar de operações
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: "Nova Pasta", OnClicked: func() { mw.createNewFolder() }},
					VSeparator{},
					PushButton{Text: "Recortar", OnClicked: func() { mw.cutSelected() }},
					PushButton{Text: "Copiar", OnClicked: func() { mw.copySelected() }},
					PushButton{Text: "Colar", OnClicked: func() { mw.pasteClipboard() }},
					VSeparator{},
					PushButton{Text: "Renomear", OnClicked: func() { mw.renameSelected() }},
					PushButton{Text: "Excluir", OnClicked: func() { mw.deleteSelected() }},
					VSeparator{},
					PushButton{Text: "Ordenar", OnClicked: func() { mw.cycleSortMode() }},
					HSpacer{},
				},
			},
			// Lista de arquivos + painel de preview
			HSplitter{
				Children: []Widget{
					TableView{
						AssignTo:         &tp.fileList,
						Model:            tp.fileModel,
						AlternatingRowBG: true,
						MultiSelection:   true,
						OnKeyDown: func(key walk.Key) {
							mods := walk.ModifiersDown()
							switch {
							case key == walk.KeyF2:
								mw.renameSelected()
							case key == walk.KeyDelete:
								mw.deleteSelected()
							case key == walk.KeyC && mods&walk.ModControl != 0:
								mw.copySelected()
							case key == walk.KeyX && mods&walk.ModControl != 0:
								mw.cutSelected()
							case key == walk.KeyV && mods&walk.ModControl != 0:
								mw.pasteClipboard()
							case key == walk.KeyF5:
								mw.navigateTo(tp.currentPath())
							case key == walk.KeyUp && mods&walk.ModAlt != 0:
								mw.goUp()
							case key == walk.KeyLeft && mods&walk.ModAlt != 0:
								mw.goBack()
							case key == walk.KeyRight && mods&walk.ModAlt != 0:
								mw.goForward()
							case key == walk.KeyReturn && mods&walk.ModAlt != 0:
								mw.showProperties()
							case key == walk.KeyReturn:
								mw.activateSelected()
							case key == walk.KeyT && mods&walk.ModControl != 0:
								mw.newTab("")
							case key == walk.KeyW && mods&walk.ModControl != 0:
								mw.closeCurrentTab()
							case key == walk.KeyTab && mods&walk.ModControl != 0:
								mw.nextTab()
							case key == walk.KeyA && mods&walk.ModControl != 0:
								if tp := mw.activeTab(); tp != nil {
									count := len(tp.fileModel.entries)
									if count > 0 {
										indexes := make([]int, count)
										for i := range indexes {
											indexes[i] = i
										}
										tp.fileList.SetSelectedIndexes(indexes)
									}
								}
							}
						},
						ContextMenuItems: []MenuItem{
							Action{Text: "Abrir", OnTriggered: func() { mw.activateSelected() }},
							Separator{},
							Action{Text: "Renomear\tF2", OnTriggered: func() { mw.renameSelected() }},
							Action{Text: "Excluir\tDel", OnTriggered: func() { mw.deleteSelected() }},
							Separator{},
							Action{Text: "Copiar\tCtrl+C", OnTriggered: func() { mw.copySelected() }},
							Action{Text: "Recortar\tCtrl+X", OnTriggered: func() { mw.cutSelected() }},
							Action{Text: "Colar\tCtrl+V", OnTriggered: func() { mw.pasteClipboard() }},
							Separator{},
							Action{Text: "Nova Pasta", OnTriggered: func() { mw.createNewFolder() }},
							Action{Text: "Nova Aba\tCtrl+T", OnTriggered: func() { mw.newTab("") }},
							Action{Text: "Propriedades\tAlt+Enter", OnTriggered: func() { mw.showProperties() }},
						},
						Columns: []TableViewColumn{
							{Title: "Nome", Width: 350},
							{Title: "Modificado", Width: 150},
							{Title: "Tipo", Width: 100},
							{Title: "Tamanho", Width: 100, Alignment: AlignFar},
						},
						OnItemActivated:       func() { mw.activateSelected() },
						OnCurrentIndexChanged: func() {
							mw.updateStatusBar()
							mw.updatePreview()
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{AssignTo: &tp.previewLabel, Text: "Pré-visualização"},
							ImageView{
								AssignTo: &tp.previewImage,
								Mode:     ImageViewModeShrink,
							},
							TextEdit{
								AssignTo: &tp.previewText,
								ReadOnly: true,
								VScroll:  true,
							},
						},
					},
				},
			},
		},
	}).Create(builder); err != nil {
		log.Printf("Erro ao criar conteúdo da aba: %v", err)
		return
	}
	_ = composite

	tp.tabPage = tabPage
	mw.tabs = append(mw.tabs, tp)
	mw.tabWidget.SetCurrentIndex(tabIdx)

	if initialPath != "" {
		mw.navigateTabTo(tp, initialPath)
	}
}

func (mw *GarqMainWindow) closeCurrentTab() {
	if len(mw.tabs) <= 1 {
		// Não fechar a última aba de navegação
		return
	}
	idx := mw.tabWidget.CurrentIndex()
	if idx < 0 || idx >= len(mw.tabs) {
		return
	}
	tp := mw.tabs[idx]
	// Remove da lista
	mw.tabs = append(mw.tabs[:idx], mw.tabs[idx+1:]...)
	// Remove o tabPage do widget
	tp.tabPage.Dispose()
	// Ajusta índice
	newIdx := idx
	if newIdx >= len(mw.tabs) {
		newIdx = len(mw.tabs) - 1
	}
	if newIdx >= 0 {
		mw.tabWidget.SetCurrentIndex(newIdx)
	}
}

func (mw *GarqMainWindow) nextTab() {
	if len(mw.tabs) < 2 {
		return
	}
	cur := mw.tabWidget.CurrentIndex()
	next := (cur + 1) % len(mw.tabs)
	mw.tabWidget.SetCurrentIndex(next)
}

func (mw *GarqMainWindow) onTabChanged() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	mw.updateStatusBar()
}

func buildNavTree(model *NavTreeModel) {
	thisPC := &NavItem{text: "Este PC", path: ""}
	userFolders := []struct{ name, path string }{
		{"Área de Trabalho", filepath.Join(os.Getenv("USERPROFILE"), "Desktop")},
		{"Documentos", filepath.Join(os.Getenv("USERPROFILE"), "Documents")},
		{"Downloads", filepath.Join(os.Getenv("USERPROFILE"), "Downloads")},
		{"Imagens", filepath.Join(os.Getenv("USERPROFILE"), "Pictures")},
		{"Músicas", filepath.Join(os.Getenv("USERPROFILE"), "Music")},
		{"Vídeos", filepath.Join(os.Getenv("USERPROFILE"), "Videos")},
	}
	for _, f := range userFolders {
		if _, err := os.Stat(f.path); err == nil {
			thisPC.children = append(thisPC.children, &NavItem{text: f.name, path: f.path, parent: thisPC})
		}
	}
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			thisPC.children = append(thisPC.children, &NavItem{text: drive, path: drive, parent: thisPC})
		}
	}
	model.roots = []*NavItem{thisPC}
}


// ---------------------------------------------------------------------------
// Navegação
// ---------------------------------------------------------------------------

func (mw *GarqMainWindow) onNavItemSelected() {
	item := mw.navTree.CurrentItem()
	if item == nil {
		return
	}
	navItem, ok := item.(*NavItem)
	if !ok || navItem.path == "" {
		return
	}
	mw.navigateTo(navItem.path)
}

func (mw *GarqMainWindow) onNavItemActivated() { mw.onNavItemSelected() }

func (mw *GarqMainWindow) activateSelected() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	idx := tp.fileList.CurrentIndex()
	if idx >= 0 && idx < len(tp.fileModel.entries) {
		e := tp.fileModel.entries[idx]
		if e.IsDir {
			mw.navigateTo(e.Path)
		} else {
			cmd := exec.Command("cmd", "/c", "start", "", e.Path)
			cmd.Start()
		}
	}
}

// navigateTo navega a aba ativa para o path dado.
func (mw *GarqMainWindow) navigateTo(path string) {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	mw.navigateTabTo(tp, path)
}

// navigateTabTo navega uma aba específica para o path dado.
func (mw *GarqMainWindow) navigateTabTo(tp *TabPane, path string) {
	mw.Synchronize(func() {
		tp.pathEdit.SetText(path)
		mw.statusLabel.SetText("Carregando...")
	})

	entries, err := mw.api.ListDirectory(path)
	if err != nil {
		mw.Synchronize(func() { mw.statusLabel.SetText(fmt.Sprintf("Erro: %v", err)) })
		return
	}

	var fileEntries []FileEntry
	for _, e := range entries {
		name, _ := e["name"].(string)
		p, _ := e["path"].(string)
		isDir, _ := e["is_dir"].(bool)
		size, _ := e["size"].(int64)
		modTime, _ := e["mod_time"].(string)
		fileEntries = append(fileEntries, FileEntry{Name: name, Path: p, IsDir: isDir, Size: size, ModTime: modTime})
	}

	sort.Slice(fileEntries, func(i, j int) bool {
		if fileEntries[i].IsDir != fileEntries[j].IsDir {
			return fileEntries[i].IsDir
		}
		return fileEntries[i].Name < fileEntries[j].Name
	})

	if len(tp.history) == 0 || tp.history[len(tp.history)-1] != path {
		// Trunca forward history ao navegar
		tp.history = append(tp.history[:tp.historyIdx+1], path)
		tp.historyIdx = len(tp.history) - 1
	}

	// Atualiza título da aba
	tabTitle := filepath.Base(path)
	if tabTitle == "" || tabTitle == "." {
		tabTitle = path
	}

	mw.Synchronize(func() {
		tp.fileModel.entries = fileEntries
		tp.allEntries = fileEntries
		tp.fileList.SetModel(tp.fileModel)
		tp.tabPage.SetTitle(tabTitle)
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", len(fileEntries)))
		if tp.searchEdit != nil {
			tp.searchEdit.SetText("")
		}
	})
}

func (mw *GarqMainWindow) goBack() {
	tp := mw.activeTab()
	if tp == nil || tp.historyIdx <= 0 {
		return
	}
	tp.historyIdx--
	path := tp.history[tp.historyIdx]
	// Navega sem adicionar ao histórico
	mw.navigateTabDirect(tp, path)
}

func (mw *GarqMainWindow) goForward() {
	tp := mw.activeTab()
	if tp == nil || tp.historyIdx >= len(tp.history)-1 {
		return
	}
	tp.historyIdx++
	path := tp.history[tp.historyIdx]
	mw.navigateTabDirect(tp, path)
}

// navigateTabDirect carrega o diretório sem alterar o histórico (usado por goBack/goForward).
func (mw *GarqMainWindow) navigateTabDirect(tp *TabPane, path string) {
	mw.Synchronize(func() {
		tp.pathEdit.SetText(path)
		mw.statusLabel.SetText("Carregando...")
	})

	entries, err := mw.api.ListDirectory(path)
	if err != nil {
		mw.Synchronize(func() { mw.statusLabel.SetText(fmt.Sprintf("Erro: %v", err)) })
		return
	}

	var fileEntries []FileEntry
	for _, e := range entries {
		name, _ := e["name"].(string)
		p, _ := e["path"].(string)
		isDir, _ := e["is_dir"].(bool)
		size, _ := e["size"].(int64)
		modTime, _ := e["mod_time"].(string)
		fileEntries = append(fileEntries, FileEntry{Name: name, Path: p, IsDir: isDir, Size: size, ModTime: modTime})
	}

	sort.Slice(fileEntries, func(i, j int) bool {
		if fileEntries[i].IsDir != fileEntries[j].IsDir {
			return fileEntries[i].IsDir
		}
		return fileEntries[i].Name < fileEntries[j].Name
	})

	tabTitle := filepath.Base(path)
	if tabTitle == "" || tabTitle == "." {
		tabTitle = path
	}

	mw.Synchronize(func() {
		tp.fileModel.entries = fileEntries
		tp.allEntries = fileEntries
		tp.fileList.SetModel(tp.fileModel)
		tp.tabPage.SetTitle(tabTitle)
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", len(fileEntries)))
		if tp.searchEdit != nil {
			tp.searchEdit.SetText("")
		}
	})
}

func (mw *GarqMainWindow) goUp() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	current := tp.currentPath()
	if current == "" {
		return
	}
	parent := filepath.Dir(current)
	if parent != current {
		mw.navigateTo(parent)
	}
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
	rows, err := mw.api.DB.Query(`SELECT COUNT(*) FROM jobs WHERE status IN ('pending','running')`)
	if err != nil {
		return
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		rows.Scan(&count)
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
	newProgressDialog(mw, mw.api.DB, jobID, jobType)
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


// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

func (mw *GarqMainWindow) updatePreview() {
	tp := mw.activeTab()
	if tp == nil {
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

// ---------------------------------------------------------------------------
// Operações de arquivo (operam sempre na aba ativa)
// ---------------------------------------------------------------------------

func (mw *GarqMainWindow) getSelectedPaths() []string {
	tp := mw.activeTab()
	if tp == nil {
		return nil
	}
	sel := tp.fileList.SelectedIndexes()
	var paths []string
	for _, idx := range sel {
		if idx >= 0 && idx < len(tp.fileModel.entries) {
			paths = append(paths, tp.fileModel.entries[idx].Path)
		}
	}
	if len(paths) == 0 {
		idx := tp.fileList.CurrentIndex()
		if idx >= 0 && idx < len(tp.fileModel.entries) {
			paths = append(paths, tp.fileModel.entries[idx].Path)
		}
	}
	return paths
}

func (mw *GarqMainWindow) createNewFolder() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	currentPath := tp.currentPath()
	if currentPath == "" {
		mw.statusLabel.SetText("Nenhuma pasta aberta")
		return
	}
	name, ok := mw.showInputDialog("Nova Pasta", "Nome da pasta:", "Nova Pasta")
	if !ok {
		mw.statusLabel.SetText("Criação cancelada")
		return
	}
	if name == "" {
		mw.statusLabel.SetText("O nome não pode estar vazio")
		return
	}
	if err := os.MkdirAll(filepath.Join(currentPath, name), 0755); err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao criar pasta: %v", err))
		return
	}
	mw.statusLabel.SetText(fmt.Sprintf("Pasta criada: %s", name))
	mw.navigateTo(currentPath)
}

func (mw *GarqMainWindow) cutSelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado para recortar")
		return
	}
	walk.Clipboard().SetText("CUT:\n" + strings.Join(paths, "\n"))
	mw.statusLabel.SetText(fmt.Sprintf("Recortado(s) %d item(s)", len(paths)))
}

func (mw *GarqMainWindow) copySelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado para copiar")
		return
	}
	walk.Clipboard().SetText("COPY:\n" + strings.Join(paths, "\n"))
	mw.statusLabel.SetText(fmt.Sprintf("Copiado(s) %d item(s)", len(paths)))
}

func (mw *GarqMainWindow) pasteClipboard() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	text, err := walk.Clipboard().Text()
	if err != nil || text == "" {
		mw.statusLabel.SetText("Nada para colar")
		return
	}
	lines := strings.SplitN(text, "\n", 2)
	if len(lines) < 2 {
		mw.statusLabel.SetText("Área de transferência inválida")
		return
	}
	isCut := lines[0] == "CUT:"
	var paths []string
	for _, p := range strings.Split(lines[1], "\n") {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nada para colar")
		return
	}
	dest := tp.currentPath()
	if dest == "" {
		mw.statusLabel.SetText("Nenhuma pasta de destino")
		return
	}

	// Enfileira via worker e abre dialog de progresso
	jobType := "copy"
	if isCut {
		jobType = "move"
	}
	jobID, err := mw.api.AddCopyJob(paths, dest, "replace")
	if isCut {
		jobID, err = mw.api.AddMoveJob(paths, dest, "replace")
	}
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao enfileirar: %v", err))
		return
	}
	if isCut {
		walk.Clipboard().Clear()
	}
	mw.statusLabel.SetText(fmt.Sprintf("⏳ Iniciando %s de %d item(s)...", jobType, len(paths)))
	go func() {
		mw.Synchronize(func() {
			mw.openProgressDialog(jobID, jobType)
		})
	}()
}

func (mw *GarqMainWindow) renameSelected() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	sel := tp.fileList.SelectedIndexes()
	if len(sel) != 1 {
		mw.statusLabel.SetText("Selecione exatamente um item para renomear")
		return
	}
	idx := sel[0]
	if idx < 0 || idx >= len(tp.fileModel.entries) {
		return
	}
	entry := tp.fileModel.entries[idx]
	newName, ok := mw.showInputDialog("Renomear", "Novo nome:", entry.Name)
	if !ok || newName == "" || newName == entry.Name {
		return
	}
	newPath := filepath.Join(filepath.Dir(entry.Path), newName)
	if err := os.Rename(entry.Path, newPath); err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao renomear: %v", err))
	} else {
		mw.statusLabel.SetText(fmt.Sprintf("Renomeado para: %s", newName))
		mw.navigateTo(tp.currentPath())
	}
}

func (mw *GarqMainWindow) deleteSelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado")
		return
	}
	count := 0
	for _, p := range paths {
		if err := os.RemoveAll(p); err == nil {
			count++
		}
	}
	mw.statusLabel.SetText(fmt.Sprintf("Excluído(s) %d item(s)", count))
	tp := mw.activeTab()
	if tp != nil {
		mw.navigateTo(tp.currentPath())
	}
}

func (mw *GarqMainWindow) cycleSortMode() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	tp.sortBy = (tp.sortBy + 1) % 4
	entries := tp.fileModel.entries
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		switch tp.sortBy {
		case 1:
			return entries[i].Size < entries[j].Size
		case 2:
			return filepath.Ext(entries[i].Name) < filepath.Ext(entries[j].Name)
		case 3:
			return entries[i].ModTime < entries[j].ModTime
		default:
			return entries[i].Name < entries[j].Name
		}
	})
	tp.fileModel.entries = entries
	mw.Synchronize(func() { tp.fileList.SetModel(tp.fileModel) })
	names := []string{"Nome", "Tamanho", "Tipo", "Data"}
	mw.statusLabel.SetText(fmt.Sprintf("Ordenado por %s", names[tp.sortBy]))
}

func (mw *GarqMainWindow) showProperties() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	idx := tp.fileList.CurrentIndex()
	if idx < 0 || idx >= len(tp.fileModel.entries) {
		mw.statusLabel.SetText("Nenhum item selecionado")
		return
	}
	entry := tp.fileModel.entries[idx]
	info, err := os.Stat(entry.Path)
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro: %v", err))
		return
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Nome: %s", info.Name()))
	lines = append(lines, fmt.Sprintf("Caminho: %s", entry.Path))
	if info.IsDir() {
		lines = append(lines, "Tipo: Pasta")
		fc, dc := countDirContents(entry.Path)
		lines = append(lines, fmt.Sprintf("Conteúdo: %d arquivo(s), %d pasta(s)", fc, dc))
	} else {
		lines = append(lines, fmt.Sprintf("Tipo: %s", filepath.Ext(entry.Name)))
		lines = append(lines, fmt.Sprintf("Tamanho: %s", formatSize(info.Size())))
	}
	lines = append(lines, fmt.Sprintf("Modificado: %s", info.ModTime().Format("02/01/2006 15:04:05")))

	dlg, err := walk.NewDialog(mw.MainWindow)
	if err != nil {
		return
	}
	dlg.SetTitle("Propriedades")
	dlg.SetLayout(walk.NewVBoxLayout())

	titleLabel, _ := walk.NewLabel(dlg)
	titleLabel.SetText("Propriedades de: " + info.Name())
	font, _ := walk.NewFont("Segoe UI", 11, walk.FontBold)
	titleLabel.SetFont(font)
	dlg.Children().Add(titleLabel)

	for _, line := range lines {
		lbl, _ := walk.NewLabel(dlg)
		lbl.SetText(line)
		dlg.Children().Add(lbl)
	}

	okBtn, _ := walk.NewPushButton(dlg)
	okBtn.SetText("OK")
	okBtn.Clicked().Attach(func() { dlg.Close(walk.DlgCmdOK) })
	dlg.Children().Add(okBtn)
	dlg.SetDefaultButton(okBtn)

	w := calcDialogWidth(info.Name())
	if w < 500 {
		w = 500
	}
	dlg.SetMinMaxSize(walk.Size{Width: w, Height: 0}, walk.Size{Width: w, Height: 0})
	dlg.RequestLayout()
	b := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{X: b.X + (b.Width-w)/2, Y: b.Y + (b.Height-300)/2, Width: w, Height: 300})
	dlg.Run()
}

// ---------------------------------------------------------------------------
// Diálogos auxiliares
// ---------------------------------------------------------------------------

func calcDialogWidth(text string) int {
	w := len(text)*8 + 80
	if w < 350 {
		w = 350
	}
	if w > 700 {
		w = 700
	}
	return w
}

func (mw *GarqMainWindow) showInputDialog(title, prompt, defaultValue string) (string, bool) {
	dlg, err := walk.NewDialog(mw.MainWindow)
	if err != nil {
		return "", false
	}
	dlg.SetTitle(title)
	dlg.SetLayout(walk.NewVBoxLayout())

	w := calcDialogWidth(defaultValue)

	hdrLabel, _ := walk.NewLabel(dlg)
	hdrLabel.SetText(prompt)
	nameEdit, _ := walk.NewLineEdit(dlg)
	nameEdit.SetText(defaultValue)
	nameEdit.SetMinMaxSize(walk.Size{Width: w - 40, Height: 24}, walk.Size{})
	okBtn, _ := walk.NewPushButton(dlg)
	okBtn.SetText("OK")
	cancelBtn, _ := walk.NewPushButton(dlg)
	cancelBtn.SetText("Cancelar")

	dlg.Children().Add(hdrLabel)
	dlg.Children().Add(nameEdit)
	dlg.Children().Add(okBtn)
	dlg.Children().Add(cancelBtn)

	var result string
	closedByOK := false

	okBtn.Clicked().Attach(func() {
		result = nameEdit.Text()
		closedByOK = true
		dlg.Close(walk.DlgCmdOK)
	})
	cancelBtn.Clicked().Attach(func() { dlg.Close(walk.DlgCmdCancel) })
	dlg.SetCancelButton(cancelBtn)
	dlg.SetDefaultButton(okBtn)
	dlg.SetMinMaxSize(walk.Size{Width: w, Height: 0}, walk.Size{Width: w, Height: 0})
	dlg.RequestLayout()
	b := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{X: b.X + (b.Width-w)/2, Y: b.Y + (b.Height-150)/2, Width: w, Height: 150})
	dlg.Run()

	return result, closedByOK
}

// ---------------------------------------------------------------------------
// Utilitários
// ---------------------------------------------------------------------------

func countDirContents(path string) (files, folders int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			folders++
		} else {
			files++
		}
	}
	return
}

func copyPath(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func getCopyPath(dir, name string) string {
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	dst := filepath.Join(dir, name)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return dst
	}
	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s - Cópia(%d)%s", base, i, ext)
		dst = filepath.Join(dir, newName)
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			return dst
		}
	}
}
