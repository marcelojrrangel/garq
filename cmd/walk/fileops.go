package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
)

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
		contentLabel = addLabel("Conteúdo: Contando...")
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
