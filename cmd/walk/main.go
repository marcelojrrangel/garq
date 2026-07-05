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

type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
}

type GarqMainWindow struct {
	*walk.MainWindow
	api           *api.API
	navTree       *walk.TreeView
	fileList      *walk.TableView
	pathEdit      *walk.LineEdit
	searchEdit    *walk.LineEdit
	statusLabel   *walk.Label
	previewImage  *walk.ImageView
	previewText   *walk.TextEdit
	previewLabel  *walk.Label
	navModel      *NavTreeModel
	fileModel     *FileTableModel
	allEntries    []FileEntry
	history       []string
	historyIdx    int
	sortBy        int
	sortDirAsc    bool
}

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

type FileTableModel struct {
	walk.TableModelBase
	entries []FileEntry
}

func (m *FileTableModel) RowCount() int { return len(m.entries) }

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
	case size >= 1 << 30:
		return fmt.Sprintf("%.2f GB", float64(size)/(1<<30))
	case size >= 1 << 20:
		return fmt.Sprintf("%.2f MB", float64(size)/(1<<20))
	case size >= 1 << 10:
		return fmt.Sprintf("%.2f KB", float64(size)/(1<<10))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
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

	worker.StartWorkerPool(4, dbConn)
	apiInstance := &api.API{DB: dbConn, Ctx: context.Background()}

	navModel := &NavTreeModel{}
	buildNavTree(navModel)

	mw := &GarqMainWindow{
		api:       apiInstance,
		navModel:  navModel,
		fileModel: &FileTableModel{},
		history:   make([]string, 0),
	}

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "Garq - Gerenciador de Arquivos",
		MinSize:  Size{Width: 1000, Height: 600},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					TreeView{
						AssignTo:             &mw.navTree,
						Model:                mw.navModel,
						OnCurrentItemChanged: mw.onNavItemSelected,
						OnItemActivated:      mw.onNavItemActivated,
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{Text: "<", MinSize: Size{Width: 30}, OnClicked: func() { mw.goBack() }},
									PushButton{Text: ">", MinSize: Size{Width: 30}, OnClicked: func() { mw.goForward() }},
									PushButton{Text: "^", MinSize: Size{Width: 30}, OnClicked: func() { mw.goUp() }},
									LineEdit{
										AssignTo: &mw.pathEdit,
										OnKeyDown: func(key walk.Key) {
											if key == walk.KeyReturn {
												path := mw.pathEdit.Text()
												if path != "" {
													mw.navigateTo(path)
												}
											}
										},
									},
								},
							},
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
									Label{Text: "Buscar:"},
									LineEdit{
										AssignTo:    &mw.searchEdit,
										CueBanner:  "Filtrar arquivos...",
										OnTextChanged: func() { mw.filterBySearch() },
									},
								},
							},
							HSplitter{
								Children: []Widget{
									TableView{
										AssignTo:         &mw.fileList,
										Model:            mw.fileModel,
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
										mw.navigateTo(mw.pathEdit.Text())
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
									Action{Text: "Propriedades\tAlt+Enter", OnTriggered: func() { mw.showProperties() }},
								},
										Columns: []TableViewColumn{
											{Title: "Nome", Width: 350},
											{Title: "Modificado", Width: 150},
											{Title: "Tipo", Width: 100},
											{Title: "Tamanho", Width: 100, Alignment: AlignFar},
										},
										OnItemActivated:     func() { mw.activateSelected() },
										OnCurrentIndexChanged: func() {
											mw.updateStatusBar()
											mw.updatePreview()
										},
									},
									Composite{
										Layout: VBox{},
										Children: []Widget{
											Label{AssignTo: &mw.previewLabel, Text: "Pré-visualização"},
											ImageView{
												AssignTo: &mw.previewImage,
												Mode:     ImageViewModeShrink,
											},
											TextEdit{
												AssignTo: &mw.previewText,
												ReadOnly: true,
												VScroll:  true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Label{AssignTo: &mw.statusLabel, Text: "Pronto"},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	mw.DropFiles().Attach(func(files []string) {
		dest := mw.pathEdit.Text()
		if dest == "" {
			return
		}
		successCount := 0
		for _, src := range files {
			baseName := filepath.Base(src)
			dst := filepath.Join(dest, baseName)
			info, err := os.Stat(src)
			if err != nil {
				continue
			}
			if info.IsDir() {
				if err := copyPath(src, dst); err == nil {
					successCount++
				}
			} else {
				if src == dst {
					dst = getCopyPath(dest, baseName)
				}
				if err := copyPath(src, dst); err == nil {
					successCount++
				}
			}
		}
		mw.statusLabel.SetText(fmt.Sprintf("Importado(s) %d item(s)", successCount))
		mw.navigateTo(dest)
	})

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

	go func() {
		<-time.After(100 * time.Millisecond)
		mw.Synchronize(func() {
			b := mw.Bounds()
			mw.SetBounds(walk.Rectangle{X: b.X, Y: b.Y, Width: b.Width + 1, Height: b.Height})
			mw.SetBounds(b)
		})
	}()

	mw.Run()

	mw.Run()
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

func (mw *GarqMainWindow) onNavItemSelected() {
	item := mw.navTree.CurrentItem()
	if item == nil {
		return
	}
	navItem, ok := item.(*NavItem)
	if !ok {
		return
	}
	if navItem.path != "" {
		mw.navigateTo(navItem.path)
	}
}

func (mw *GarqMainWindow) onNavItemActivated() { mw.onNavItemSelected() }

func (mw *GarqMainWindow) activateSelected() {
	idx := mw.fileList.CurrentIndex()
	if idx >= 0 && idx < len(mw.fileModel.entries) {
		e := mw.fileModel.entries[idx]
		if e.IsDir {
			mw.navigateTo(e.Path)
		} else {
			cmd := exec.Command("cmd", "/c", "start", "", e.Path)
			cmd.Start()
		}
	}
}

func (mw *GarqMainWindow) navigateTo(path string) {
	mw.Synchronize(func() {
		mw.pathEdit.SetText(path)
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

	if len(mw.history) == 0 || mw.history[len(mw.history)-1] != path {
		mw.history = append(mw.history, path)
		mw.historyIdx = len(mw.history) - 1
	}

	mw.Synchronize(func() {
		mw.fileModel.entries = fileEntries
		mw.allEntries = fileEntries
		mw.fileList.SetModel(mw.fileModel)
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", len(entries)))
		if mw.searchEdit != nil {
			mw.searchEdit.SetText("")
		}
	})
}

func (mw *GarqMainWindow) goBack() {
	if mw.historyIdx > 0 {
		mw.historyIdx--
		mw.navigateTo(mw.history[mw.historyIdx])
	}
}

func (mw *GarqMainWindow) goForward() {
	if mw.historyIdx < len(mw.history)-1 {
		mw.historyIdx++
		mw.navigateTo(mw.history[mw.historyIdx])
	}
}

func (mw *GarqMainWindow) goUp() {
	current := mw.pathEdit.Text()
	if current == "" {
		return
	}
	parent := filepath.Dir(current)
	if parent != current {
		mw.navigateTo(parent)
	}
}

func (mw *GarqMainWindow) updateStatusBar() {
	total := len(mw.fileModel.entries)
	sel := len(mw.fileList.SelectedIndexes())
	if sel > 0 {
		mw.statusLabel.SetText(fmt.Sprintf("%d itens, %d selecionados", total, sel))
	} else {
		mw.statusLabel.SetText(fmt.Sprintf("%d itens", total))
	}
}

func (mw *GarqMainWindow) filterBySearch() {
	query := mw.searchEdit.Text()
	if query == "" {
		mw.fileModel.entries = mw.allEntries
	} else {
		var filtered []FileEntry
		for _, e := range mw.allEntries {
			if containsIgnoreCase(e.Name, query) {
				filtered = append(filtered, e)
			}
		}
		mw.fileModel.entries = filtered
	}
	mw.fileList.SetModel(mw.fileModel)
	mw.updateStatusBar()
}

func containsIgnoreCase(s, sub string) bool {
	s = strings.ToLower(s)
	sub = strings.ToLower(sub)
	return strings.Contains(s, sub)
}

func (mw *GarqMainWindow) updatePreview() {
	idx := mw.fileList.CurrentIndex()
	if idx < 0 || idx >= len(mw.fileModel.entries) {
		mw.previewLabel.SetText("Pré-visualização")
		mw.previewImage.SetImage(nil)
		mw.previewText.SetText("")
		return
	}
	entry := mw.fileModel.entries[idx]

	if entry.IsDir {
		mw.previewLabel.SetText(fmt.Sprintf("Pasta: %s", entry.Name))
		mw.previewImage.SetImage(nil)
		mw.previewText.SetText("")
		return
	}

	ext := strings.ToLower(filepath.Ext(entry.Name))
	mw.previewLabel.SetText(fmt.Sprintf("%s — %s", entry.Name, formatSize(entry.Size)))

	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true, ".gif": true, ".ico": true, ".tiff": true}
	textExts := map[string]bool{".txt": true, ".md": true, ".go": true, ".js": true, ".ts": true, ".py": true, ".json": true, ".xml": true, ".html": true, ".css": true, ".csv": true, ".log": true, ".ini": true, ".cfg": true, ".yaml": true, ".yml": true, ".toml": true, ".sh": true, ".bat": true, ".cmd": true}

	if imageExts[ext] {
		img, err := walk.NewImageFromFile(entry.Path)
		if err == nil {
			mw.previewImage.SetImage(img)
			mw.previewText.SetText("")
		} else {
			mw.previewImage.SetImage(nil)
			mw.previewText.SetText(fmt.Sprintf("Erro ao carregar imagem: %v", err))
		}
		return
	}

	mw.previewImage.SetImage(nil)

	if textExts[ext] && entry.Size < 1<<20 {
		data, err := os.ReadFile(entry.Path)
		if err == nil {
			mw.previewText.SetText(string(data))
		} else {
			mw.previewText.SetText(fmt.Sprintf("Erro ao ler arquivo: %v", err))
		}
		return
	}

	mw.previewText.SetText(fmt.Sprintf("Arquivo: %s\nTamanho: %s\nModificado: %s\nTipo: %s",
		entry.Name, formatSize(entry.Size), entry.ModTime, ext))
}

func (mw *GarqMainWindow) getSelectedPaths() []string {
	sel := mw.fileList.SelectedIndexes()
	var paths []string
	for _, idx := range sel {
		if idx >= 0 && idx < len(mw.fileModel.entries) {
			paths = append(paths, mw.fileModel.entries[idx].Path)
		}
	}
	if len(paths) == 0 {
		idx := mw.fileList.CurrentIndex()
		if idx >= 0 && idx < len(mw.fileModel.entries) {
			paths = append(paths, mw.fileModel.entries[idx].Path)
		}
	}
	return paths
}

func (mw *GarqMainWindow) createNewFolder() {
	currentPath := mw.pathEdit.Text()
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
	fullPath := filepath.Join(currentPath, name)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
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
	data := "CUT:\n" + strings.Join(paths, "\n")
	walk.Clipboard().SetText(data)
	mw.statusLabel.SetText(fmt.Sprintf("Recortado(s) %d item(s) - pronto para colar", len(paths)))
}

func (mw *GarqMainWindow) copySelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado para copiar")
		return
	}
	data := "COPY:\n" + strings.Join(paths, "\n")
	walk.Clipboard().SetText(data)
	mw.statusLabel.SetText(fmt.Sprintf("Copiado(s) %d item(s) - pronto para colar", len(paths)))
}

func (mw *GarqMainWindow) pasteClipboard() {
	text, err := walk.Clipboard().Text()
	if err != nil || text == "" {
		mw.statusLabel.SetText("Nada para colar - use Copiar ou Recortar primeiro")
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
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 {
		mw.statusLabel.SetText("Nada para colar")
		return
	}

	dest := mw.pathEdit.Text()
	if dest == "" {
		mw.statusLabel.SetText("Nenhuma pasta de destino")
		return
	}
	successCount := 0
	for _, src := range paths {
		baseName := filepath.Base(src)
		srcDir := filepath.Dir(src)
		dst := filepath.Join(dest, baseName)
		if isCut && srcDir == dest {
			successCount++
			continue
		}
		if isCut && srcDir != dest {
			if err := os.Rename(src, dst); err == nil {
				successCount++
			}
		} else {
			if src == dst {
				dst = getCopyPath(dest, baseName)
			}
			if err := copyPath(src, dst); err == nil {
				successCount++
			}
		}
	}
	if isCut {
		walk.Clipboard().Clear()
	}
	mw.statusLabel.SetText(fmt.Sprintf("Colado(s) %d item(s)", successCount))
	mw.navigateTo(dest)
}

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
	cancelBtn.Clicked().Attach(func() {
		dlg.Close(walk.DlgCmdCancel)
	})

	dlg.SetCancelButton(cancelBtn)
	dlg.SetDefaultButton(okBtn)

	dlg.SetMinMaxSize(walk.Size{Width: w, Height: 0}, walk.Size{Width: w, Height: 0})
	dlg.RequestLayout()

	mwBounds := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{
		X:      mwBounds.X + (mwBounds.Width-w)/2,
		Y:      mwBounds.Y + (mwBounds.Height-150)/2,
		Width:  w,
		Height: 150,
	})

	dlg.Run()

	return result, closedByOK
}

func (mw *GarqMainWindow) renameSelected() {
	sel := mw.fileList.SelectedIndexes()
	if len(sel) != 1 {
		mw.statusLabel.SetText("Selecione exatamente um item para renomear")
		return
	}
	idx := sel[0]
	if idx < 0 || idx >= len(mw.fileModel.entries) {
		mw.statusLabel.SetText("Seleção inválida")
		return
	}
	entry := mw.fileModel.entries[idx]
	mw.statusLabel.SetText(fmt.Sprintf("Renomeando: %s", entry.Name))

	newName, ok := mw.showInputDialog("Renomear", "Novo nome:", entry.Name)
	if !ok {
		mw.statusLabel.SetText("Renomeação cancelada")
		return
	}
	if newName == "" {
		mw.statusLabel.SetText("O nome não pode estar vazio")
		return
	}
	if newName == entry.Name {
		mw.statusLabel.SetText("Nome não alterado")
		return
	}
	newPath := filepath.Join(filepath.Dir(entry.Path), newName)
	if err := os.Rename(entry.Path, newPath); err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao renomear: %v", err))
	} else {
		mw.statusLabel.SetText(fmt.Sprintf("Renomeado para: %s", newName))
		mw.navigateTo(mw.pathEdit.Text())
	}
}

func (mw *GarqMainWindow) deleteSelected() {
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Nenhum item selecionado para excluir")
		return
	}
	count := 0
	for _, p := range paths {
		if err := os.RemoveAll(p); err == nil {
			count++
		}
	}
	mw.statusLabel.SetText(fmt.Sprintf("Excluído(s) %d item(s)", count))
	mw.navigateTo(mw.pathEdit.Text())
}

func (mw *GarqMainWindow) showProperties() {
	idx := mw.fileList.CurrentIndex()
	if idx < 0 || idx >= len(mw.fileModel.entries) {
		mw.statusLabel.SetText("Nenhum item selecionado")
		return
	}
	entry := mw.fileModel.entries[idx]
	info, err := os.Stat(entry.Path)
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao ler propriedades: %v", err))
		return
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Nome: %s", info.Name()))
	lines = append(lines, fmt.Sprintf("Caminho: %s", entry.Path))

	if info.IsDir() {
		lines = append(lines, "Tipo: Pasta")
		fileCount, folderCount := countDirContents(entry.Path)
		lines = append(lines, fmt.Sprintf("Conteúdo: %d arquivo(s), %d pasta(s)", fileCount, folderCount))
	} else {
		lines = append(lines, fmt.Sprintf("Tipo: %s", filepath.Ext(entry.Name)))
		lines = append(lines, fmt.Sprintf("Tamanho: %s", formatSize(info.Size())))
	}

	lines = append(lines, fmt.Sprintf("Criado: %s", info.ModTime().Format("02/01/2006 15:04:05")))
	lines = append(lines, fmt.Sprintf("Modificado: %s", info.ModTime().Format("02/01/2006 15:04:05")))

	attr := ""
	if info.Mode()&0200 != 0 {
		attr += "Somente Leitura "
	}
	if info.Mode()&0100 != 0 {
		attr += "Executável "
	}
	if attr == "" {
		attr = "Normal"
	}
	lines = append(lines, fmt.Sprintf("Atributos: %s", attr))

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
	h := 300
	dlg.SetMinMaxSize(walk.Size{Width: w, Height: 0}, walk.Size{Width: w, Height: 0})
	dlg.RequestLayout()
	mwBounds := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{
		X:      mwBounds.X + (mwBounds.Width-w)/2,
		Y:      mwBounds.Y + (mwBounds.Height-h)/2,
		Width:  w,
		Height: h,
	})

	dlg.Run()
}

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

func (mw *GarqMainWindow) cycleSortMode() {
	mw.sortBy = (mw.sortBy + 1) % 4
	mw.sortDirAsc = true
	entries := mw.fileModel.entries
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		var less bool
		switch mw.sortBy {
		case 0:
			less = entries[i].Name < entries[j].Name
		case 1:
			less = entries[i].Size < entries[j].Size
		case 2:
			less = filepath.Ext(entries[i].Name) < filepath.Ext(entries[j].Name)
		case 3:
			less = entries[i].ModTime < entries[j].ModTime
		default:
			less = entries[i].Name < entries[j].Name
		}
		return less
	})
	mw.fileModel.entries = entries
	mw.Synchronize(func() { mw.fileList.SetModel(mw.fileModel) })
	names := []string{"Nome", "Tamanho", "Tipo", "Data"}
	mw.statusLabel.SetText(fmt.Sprintf("Ordenado por %s", names[mw.sortBy]))
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
