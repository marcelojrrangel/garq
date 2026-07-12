//go:build windows

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func (mw *GarqMainWindow) compressSelected() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	paths := mw.getSelectedPaths()
	if len(paths) == 0 {
		mw.statusLabel.SetText("Selecione ao menos um item para comprimir")
		return
	}
	mw.compressDialog(paths)
}

func (mw *GarqMainWindow) compressDialog(sources []string) {
	curPath := ""
	if tp := mw.activeTab(); tp != nil {
		curPath = tp.currentPath()
	}

	archiveName := suggestArchiveName(sources, curPath)

	var nameEdit *walk.LineEdit
	var destEdit *walk.LineEdit
	var rbReplace, rbRename, rbSkip *walk.RadioButton
	var acceptBtn, cancelBtn *walk.PushButton

	conflict := "replace"

	dlg, err := walk.NewDialog(mw.MainWindow)
	if err != nil {
		log.Printf("compressDialog: %v", err)
		return
	}
	dlg.SetTitle("Comprimir arquivos")
	dlg.SetLayout(walk.NewVBoxLayout())
	dlg.SetIcon(mw.MainWindow.Icon())

	builder := NewBuilder(dlg)

	if err := (Composite{
		Layout: VBox{Margins: Margins{Left: 14, Top: 10, Right: 14, Bottom: 10}},
		Children: []Widget{
			Label{
				Text: "Comprimir selecionado(s)",
				Font: Font{Bold: true, PointSize: 10},
			},
			Composite{
				Layout: Grid{Columns: 3},
				Children: []Widget{
					Label{Text: "Arquivo:"},
					LineEdit{
						AssignTo: &nameEdit,
						Text:     archiveName,
					},
					PushButton{
						Text: "...",
						MinSize: Size{Width: 30},
						OnClicked: func() {
							fd := new(walk.FileDialog)
							fd.FilePath = nameEdit.Text()
							fd.Filter = "Arquivos 7z (*.7z)|*.7z"
							if ok, _ := fd.ShowSave(mw.MainWindow); ok {
								p := fd.FilePath
								if !strings.HasSuffix(strings.ToLower(p), ".7z") {
									p += ".7z"
								}
								nameEdit.SetText(p)
							}
						},
					},
					Label{Text: "Pasta:"},
					LineEdit{
						AssignTo: &destEdit,
						Text:     curPath,
					},
					PushButton{
						Text: "...",
						MinSize: Size{Width: 30},
						OnClicked: func() {
							p := showFolderDialog(mw, "Selecionar pasta de destino", destEdit.Text())
							if p != "" {
								destEdit.SetText(p)
							}
						},
					},
				},
			},
			GroupBox{
				Title: "Se o arquivo existir",
				Layout: HBox{},
				Children: []Widget{
					RadioButton{
						AssignTo: &rbReplace,
						Text:     "Substituir",
						Value:    "replace",
						OnClicked: func() { conflict = "replace" },
					},
					RadioButton{
						AssignTo: &rbRename,
						Text:     "Renomear automaticamente",
						Value:    "rename",
						OnClicked: func() { conflict = "rename" },
					},
					RadioButton{
						AssignTo: &rbSkip,
						Text:     "Pular",
						Value:    "skip",
						OnClicked: func() { conflict = "skip" },
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &acceptBtn,
						Text:     "OK",
						MinSize:  Size{Width: 80},
						OnClicked: func() {
							name := nameEdit.Text()
							dest := destEdit.Text()
							if name == "" {
								walk.MsgBox(dlg, "Erro", "O nome do arquivo não pode estar vazio", walk.MsgBoxIconError)
								return
							}
							if dest == "" {
								walk.MsgBox(dlg, "Erro", "A pasta de destino não pode estar vazia", walk.MsgBoxIconError)
								return
							}
							if !strings.HasSuffix(strings.ToLower(name), ".7z") {
								name += ".7z"
							}
							fullDest := filepath.Join(dest, name)
							jobID, err := mw.service.Compress(sources, fullDest, conflict)
							if err != nil {
								walk.MsgBox(dlg, "Erro", fmt.Sprintf("Erro ao enfileirar: %v", err), walk.MsgBoxIconError)
								return
							}
							dlg.Close(walk.DlgCmdOK)
							mw.statusLabel.SetText(fmt.Sprintf("⏳ Comprimindo %d item(s)...", len(sources)))
							go func() {
								mw.Synchronize(func() {
									mw.openProgressDialog(jobID, "compress")
								})
							}()
						},
					},
					PushButton{
						AssignTo: &cancelBtn,
						Text:     "Cancelar",
						MinSize:  Size{Width: 80},
						OnClicked: func() { dlg.Close(walk.DlgCmdCancel) },
					},
				},
			},
		},
	}).Create(builder); err != nil {
		log.Printf("compressDialog: create error: %v", err)
		return
	}

	dlg.SetCancelButton(cancelBtn)
	dlg.SetDefaultButton(acceptBtn)
	rbReplace.SetChecked(true)

	w, h := 500, 320
	b := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{
		X: b.X + (b.Width-w)/2, Y: b.Y + (b.Height-h)/2,
		Width: w, Height: h,
	})
	dlg.Run()
}

func suggestArchiveName(sources []string, curPath string) string {
	if len(sources) == 1 {
		base := filepath.Base(sources[0])
		ext := filepath.Ext(base)
		return strings.TrimSuffix(base, ext) + ".7z"
	}
	if curPath != "" {
		return filepath.Base(curPath) + ".7z"
	}
	return "arquivo.7z"
}
