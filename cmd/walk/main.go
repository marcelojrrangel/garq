package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"garq/compress"
	"garq/internal/api"
	copyimpl "garq/internal/copy"
	"garq/internal/db"
	"garq/internal/service"
	"garq/internal/worker"
)

func init() {
	logPath := filepath.Join(os.TempDir(), "garq.log")
	f, err := os.Create(logPath)
	if err == nil {
		log.SetOutput(f)
	}
	log.Println("=== Garq iniciado ===")
}

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
	tabPage    *walk.TabPage
	fileList   *walk.TableView
	pathEdit   *walk.LineEdit
	searchEdit *walk.LineEdit
	previewImage   *walk.ImageView
	previewText    *walk.TextEdit
	previewLabel   *walk.Label
	previewVisible  bool
	previewComposite *walk.Composite
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


// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Shell icon helper
// ---------------------------------------------------------------------------

func getShellIcon(path string) *walk.Icon {
	var shfi win.SHFILEINFO
	ret := win.SHGetFileInfo(
		syscall.StringToUTF16Ptr(path),
		0,
		&shfi,
		uint32(unsafe.Sizeof(shfi)),
		win.SHGFI_ICON|win.SHGFI_SMALLICON,
	)
	if ret == 0 {
		log.Printf("SHGetFileInfo falhou para %s", path)
		return nil
	}
	icon, err := walk.NewIconFromHICON(shfi.HIcon)
	if err != nil {
		log.Printf("NewIconFromHICON falhou: %v", err)
		win.DestroyIcon(shfi.HIcon)
		return nil
	}
	return icon
}

func tabTitle(path string) string {
	trimmed := strings.TrimRight(path, "\\")
	base := filepath.Base(trimmed)
	if base == "" || base == "." {
		return path
	}
	return base
}

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

	store := &db.DBStore{DB: dbConn}
	worker.StartWorkerPool(4, store, compress.CLIAdapter{}, copyimpl.CopierAdapter{})
	apiInstance := api.New(dbConn)
	svc := service.New(apiInstance)
	log.Printf("Banco inicializado: %s", dbPath)

	navModel := &NavTreeModel{}
	buildSectionedNavTree(navModel)

	mw := &GarqMainWindow{
		service:  svc,
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
					PushButton{AssignTo: &mw.btnBack, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.goBack() }},
					PushButton{AssignTo: &mw.btnForward, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.goForward() }},
					PushButton{AssignTo: &mw.btnUp, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.goUp() }},
					VSeparator{},
					PushButton{AssignTo: &mw.btnNewTab, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.newTab("") }},
					PushButton{AssignTo: &mw.btnCloseTab, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.closeCurrentTab() }},
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

	// Aplica ícones nos botões da toolbar global
	navBtn := func(b *walk.PushButton, id int, tip string) {
		if b == nil {
			return
		}
		ic, _ := walk.NewIconFromResourceId(id)
		if ic == nil {
			b.SetText(tip)
			return
		}
		hwnd := b.Handle()
		if hwnd != 0 {
			if style := win.GetWindowLong(hwnd, win.GWL_STYLE); style&0x40 == 0 {
				win.SetWindowLong(hwnd, win.GWL_STYLE, style|0x40)
			}
		}
		b.SetImage(ic)
		b.SetToolTipText(tip)
	}
	navBtn(mw.btnBack, 112, "Voltar (Alt+←)")
	navBtn(mw.btnForward, 113, "Avançar (Alt+→)")
	navBtn(mw.btnUp, 114, "Subir (Alt+↑)")
	navBtn(mw.btnNewTab, 115, "Nova aba (Ctrl+T)")
	navBtn(mw.btnCloseTab, 116, "Fechar aba (Ctrl+W)")

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
		if len(files) == 0 {
			mw.statusLabel.SetText("Nenhum arquivo solto")
			return
		}
		tp := mw.activeTab()
		if tp == nil {
			mw.statusLabel.SetText("Nenhuma aba ativa")
			return
		}
		dest := tp.currentPath()
		if dest == "" {
			mw.statusLabel.SetText("Destino inválido")
			return
		}
		for _, src := range files {
			if filepath.Dir(src) == dest {
				mw.statusLabel.SetText("Origem e destino iguais")
				return
			}
		}
		jobID, err := mw.service.Copy(files, dest, "replace")
		if err != nil {
			mw.statusLabel.SetText(fmt.Sprintf("Erro ao enfileirar cópia: %v", err))
			return
		}
		mw.statusLabel.SetText(fmt.Sprintf("⏳ Iniciando cópia de %d item(s)...", len(files)))
		mw.openProgressDialog(jobID, "copy")
	})

	// Navega para o primeiro drive ao iniciar
	go func() {
		roots, err := mw.service.ListRoots()
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
		fileModel:      &FileTableModel{},
		history:        make([]string, 0),
		historyIdx:     -1,
		previewVisible: false,
		sortBy:         0,
		sortDirAsc:     true,
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
	var btnNewFolder, btnCut, btnCopy, btnPaste, btnRename, btnDelete, btnCompress, btnExtract, btnPreview *walk.PushButton

	// Carrega ícones de recursos embutidos no .exe (resources.syso)
	ico := func(id int) *walk.Icon {
		ic, err := walk.NewIconFromResourceId(id)
		if err != nil {
			return nil
		}
		return ic
	}
	icFolder   := ico(101)
	icCut      := ico(103)
	icCopy     := ico(104)
	icPaste    := ico(105)
	icRename   := ico(106)
	icDelete   := ico(107)
	icCompress := ico(108)
	icExtract  := ico(109)
	icPreview  := ico(111)

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
			// Toolbar de operações (ícones do shell + tooltips)
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{AssignTo: &btnNewFolder, Image: icFolder, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.createNewFolder() }},
					VSeparator{},
					PushButton{AssignTo: &btnCut, Image: icCut, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.cutSelected() }},
					PushButton{AssignTo: &btnCopy, Image: icCopy, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.copySelected() }},
					PushButton{AssignTo: &btnPaste, Image: icPaste, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.pasteClipboard() }},
					VSeparator{},
					PushButton{AssignTo: &btnRename, Image: icRename, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.renameSelected() }},
					PushButton{AssignTo: &btnDelete, Image: icDelete, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.deleteSelected() }},
					VSeparator{},
					PushButton{AssignTo: &btnCompress, Image: icCompress, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.compressSelected() }},
					PushButton{AssignTo: &btnExtract, Image: icExtract, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.extractSelected() }},
					VSeparator{},
					PushButton{AssignTo: &btnPreview, Image: icPreview, Text: "", MinSize: Size{Width: 28, Height: 28}, MaxSize: Size{Width: 28, Height: 28}, OnClicked: func() { mw.togglePreview() }},
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
								if mods&walk.ModShift != 0 {
									mw.deleteSelectedPermanently()
								} else {
									mw.deleteSelected()
								}
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
							Action{Text: "Mover para Lixeira\tDel", OnTriggered: func() { mw.deleteSelected() }},
							Action{Text: "Excluir permanentemente\tShift+Del", OnTriggered: func() { mw.deleteSelectedPermanently() }},
							Separator{},
							Action{Text: "Copiar\tCtrl+C", OnTriggered: func() { mw.copySelected() }},
							Action{Text: "Recortar\tCtrl+X", OnTriggered: func() { mw.cutSelected() }},
							Action{Text: "Colar\tCtrl+V", OnTriggered: func() { mw.pasteClipboard() }},
							Separator{},
							Action{Text: "Nova Pasta", OnTriggered: func() { mw.createNewFolder() }},
							Action{Text: "Nova Aba\tCtrl+T", OnTriggered: func() { mw.newTab("") }},
							Action{Text: "Propriedades\tAlt+Enter", OnTriggered: func() { mw.showProperties() }},
							Separator{},
							Action{Text: "Comprimir...", OnTriggered: func() { mw.compressSelected() }},
							Action{Text: "Extrair aqui", OnTriggered: func() { mw.extractHere() }},
							Action{Text: "Extrair para...", OnTriggered: func() { mw.extractSelected() }},
						},
						Columns: []TableViewColumn{
							{Title: "Nome", Width: 250},
							{Title: "Tipo", Width: 90},
							{Title: "Tamanho", Width: 100, Alignment: AlignFar},
							{Title: "Modificado", Width: 150},
						},
						OnItemActivated:       func() { mw.activateSelected() },
						OnCurrentIndexChanged: func() {
							mw.updateStatusBar()
							mw.updatePreview()
						},
					},
					Composite{
						AssignTo: &tp.previewComposite,
						Layout:   VBox{},
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

	// Força BS_ICON em cada botão da toolbar (declarative.Image não aplica o estilo)
		setBtn := func(b *walk.PushButton, ic *walk.Icon, tip string) {
		if b == nil || ic == nil {
			return
		}
		hwnd := b.Handle()
		if hwnd == 0 {
			return
		}
		if style := win.GetWindowLong(hwnd, win.GWL_STYLE); style&0x40 == 0 {
			win.SetWindowLong(hwnd, win.GWL_STYLE, style|0x40) // BS_ICON
		}
		b.SetImage(ic)
		b.SetToolTipText(tip)
	}
	setBtn(btnNewFolder, icFolder, "Nova pasta")
	setBtn(btnCut, icCut, "Recortar (Ctrl+X)")
	setBtn(btnCopy, icCopy, "Copiar (Ctrl+C)")
	setBtn(btnPaste, icPaste, "Colar (Ctrl+V)")
	setBtn(btnRename, icRename, "Renomear (F2)")
	setBtn(btnDelete, icDelete, "Excluir (Del)")
	setBtn(btnCompress, icCompress, "Comprimir")
	setBtn(btnExtract, icExtract, "Extrair")
	setBtn(btnPreview, icPreview, "Pré-visualização")

	tp.fileList.ColumnClicked().Attach(func(col int) {
		if tp.sortBy == col {
			tp.sortDirAsc = !tp.sortDirAsc
		} else {
			tp.sortBy = col
			tp.sortDirAsc = true
		}
		mw.sortTab(tp)
	})

	tp.tabPage = tabPage
	if tp.previewComposite != nil {
		tp.previewComposite.SetVisible(false)
	}
	mw.tabs = append(mw.tabs, tp)
	mw.tabWidget.SetCurrentIndex(tabIdx)

	log.Printf("Nova aba criada [%d]: %s", tabIdx, title)
	if initialPath != "" {
		mw.navigateTabTo(tp, initialPath)
	}
}

func (mw *GarqMainWindow) closeCurrentTab() {
	if len(mw.tabs) <= 1 {
		log.Println("closeCurrentTab: apenas 1 aba, ignorando")
		return
	}
	idx := mw.tabWidget.CurrentIndex()
	if idx < 0 || idx >= len(mw.tabs) {
		log.Printf("closeCurrentTab: índice inválido %d (tabs=%d)", idx, len(mw.tabs))
		return
	}
	tp := mw.tabs[idx]
	path := tp.currentPath()
	log.Printf("Fechando aba [%d]: %s", idx, path)
	mw.tabs = append(mw.tabs[:idx], mw.tabs[idx+1:]...)
	mw.tabWidget.Pages().Remove(tp.tabPage)
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
		log.Println("onTabChanged: aba ativa é nil")
		return
	}
	mw.updateNavButtons()
	mw.updateStatusBar()
}

func (mw *GarqMainWindow) togglePreview() {
	tp := mw.activeTab()
	if tp == nil || tp.previewComposite == nil {
		log.Println("togglePreview: aba ou composite nil")
		return
	}
	tp.previewVisible = !tp.previewVisible
	tp.previewComposite.SetVisible(tp.previewVisible)
	log.Printf("togglePreview: %v", tp.previewVisible)
}

func makeNavItem(text, path string, parent *NavItem) *NavItem {
	return &NavItem{text: text, path: path, parent: parent, children: nil, isSection: false}
}

func makeSection(text string) *NavItem {
	return &NavItem{text: text, path: "", parent: nil, children: nil, isSection: true}
}

func buildSectionedNavTree(model *NavTreeModel) {
	qa := makeSection("Acesso Rápido")
	for _, f := range []struct{ name, path string }{
		{"Área de Trabalho", filepath.Join(os.Getenv("USERPROFILE"), "Desktop")},
		{"Downloads", filepath.Join(os.Getenv("USERPROFILE"), "Downloads")},
		{"Documentos", filepath.Join(os.Getenv("USERPROFILE"), "Documents")},
	} {
		if _, err := os.Stat(f.path); err == nil {
			qa.children = append(qa.children, makeNavItem(f.name, f.path, qa))
		}
	}

	drives := makeNavItem("Unidades", "", nil)
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives.children = append(drives.children, makeNavItem(drive, drive, drives))
		}
	}

	thisPC := makeSection("Este PC")
	for _, f := range []struct{ name, path string }{
		{"Área de Trabalho", filepath.Join(os.Getenv("USERPROFILE"), "Desktop")},
		{"Documentos", filepath.Join(os.Getenv("USERPROFILE"), "Documents")},
		{"Downloads", filepath.Join(os.Getenv("USERPROFILE"), "Downloads")},
		{"Imagens", filepath.Join(os.Getenv("USERPROFILE"), "Pictures")},
		{"Músicas", filepath.Join(os.Getenv("USERPROFILE"), "Music")},
		{"Vídeos", filepath.Join(os.Getenv("USERPROFILE"), "Videos")},
	} {
		if _, err := os.Stat(f.path); err == nil {
			thisPC.children = append(thisPC.children, makeNavItem(f.name, f.path, thisPC))
		}
	}
	thisPC.children = append(thisPC.children, drives)
	drives.parent = thisPC

	model.roots = []*NavItem{qa, thisPC}
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
	if !ok || navItem.path == "" || navItem.isSection {
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
			mw.service.OpenFile(e.Path)
		}
	}
}

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


// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

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
	if err := mw.service.CreateFolder(currentPath, name); err != nil {
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
	mw.service.SetClipboard("cut", paths)
	mw.statusLabel.SetText(fmt.Sprintf("Recortado(s) %d item(s)", len(paths)))
}

func (mw *GarqMainWindow) copySelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado para copiar")
		return
	}
	mw.service.SetClipboard("copy", paths)
	mw.statusLabel.SetText(fmt.Sprintf("Copiado(s) %d item(s)", len(paths)))
}

func (mw *GarqMainWindow) pasteClipboard() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	dest := tp.currentPath()
	if dest == "" {
		mw.statusLabel.SetText("Nenhuma pasta de destino")
		return
	}
	jobID, jobType, err := mw.service.Paste(dest)
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Nada para colar: %v", err))
		return
	}
	mw.statusLabel.SetText(fmt.Sprintf("⏳ Iniciando %s...", jobType))
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
	if err := mw.service.Rename(entry.Path, newName); err != nil {
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
	if err := mw.service.DeleteToRecycle(paths); err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao mover para lixeira: %v", err))
		return
	}
	mw.statusLabel.SetText(fmt.Sprintf("🗑 %d item(s) movido(s) para a Lixeira", len(paths)))
	tp := mw.activeTab()
	if tp != nil {
		mw.navigateTo(tp.currentPath())
	}
}

func (mw *GarqMainWindow) deleteSelectedPermanently() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado")
		return
	}
	result := walk.MsgBox(mw.MainWindow, "Excluir permanentemente",
		fmt.Sprintf("Deseja excluir permanentemente %d item(s)? Esta ação não pode ser desfeita.", len(paths)),
		walk.MsgBoxYesNo|walk.MsgBoxIconWarning|walk.MsgBoxDefButton2)
	if result != walk.DlgCmdYes {
		mw.statusLabel.SetText("Exclusão cancelada")
		return
	}
	jobID, err := mw.service.Delete(paths)
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao enfileirar: %v", err))
		return
	}
	mw.statusLabel.SetText(fmt.Sprintf("⏳ Iniciando exclusão de %d item(s)...", len(paths)))
	go func() {
		mw.Synchronize(func() {
			mw.openProgressDialog(jobID, "delete")
		})
	}()
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
	mw.Synchronize(func() { tp.fileList.SetModel(tp.fileModel) })
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

	isDir := info.IsDir()

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

	addLabel := func(text string) *walk.Label {
		lbl, _ := walk.NewLabel(dlg)
		lbl.SetText(text)
		dlg.Children().Add(lbl)
		return lbl
	}

	addLabel(fmt.Sprintf("Nome: %s", info.Name()))
	addLabel(fmt.Sprintf("Caminho: %s", entry.Path))
	var contentLabel *walk.Label
	if isDir {
		addLabel("Tipo: Pasta")
		contentLabel = addLabel(" Conteúdo: Contando...")
	} else {
		addLabel(fmt.Sprintf("Tipo: %s", filepath.Ext(entry.Name)))
		addLabel(fmt.Sprintf("Tamanho: %s", formatSize(info.Size())))
	}
	addLabel(fmt.Sprintf("Modificado: %s", info.ModTime().Format("02/01/2006 15:04:05")))

	closed := false
	dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		closed = true
	})
	if isDir {
		go func() {
			fc, dc := countDirContents(entry.Path)
			mw.Synchronize(func() {
				if closed {
					return
				}
				contentLabel.SetText(fmt.Sprintf("Conteúdo: %d arquivo(s), %d pasta(s)", fc, dc))
			})
		}()
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

