package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"garq/internal/service"
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

func main() {
	dbPath := "garq.db"
	if envPath := os.Getenv("GARQ_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	dbConn, err := service.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Erro ao inicializar banco: %v", err)
	}
	defer dbConn.Close()

	service.StartWorkerPool(dbConn)
	svc := service.NewFromDB(dbConn)
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
