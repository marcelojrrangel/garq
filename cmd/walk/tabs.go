package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func tabTitle(path string) string {
	trimmed := strings.TrimRight(path, "\\")
	base := filepath.Base(trimmed)
	if base == "" || base == "." {
		return path
	}
	return base
}

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
	icFolder := ico(101)
	icCut := ico(103)
	icCopy := ico(104)
	icPaste := ico(105)
	icRename := ico(106)
	icDelete := ico(107)
	icCompress := ico(108)
	icExtract := ico(109)
	icPreview := ico(111)

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
