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

func (mw *GarqMainWindow) extractSelected() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	paths := mw.getSelectedPaths()
	if len(paths) != 1 {
		mw.statusLabel.SetText("Selecione exatamente um arquivo .7z")
		return
	}
	archive := paths[0]
	if !isArchiveFile(archive) {
		mw.statusLabel.SetText("Selecione um arquivo .7z para extrair")
		return
	}
	mw.extractDialog(archive)
}

func (mw *GarqMainWindow) extractHere() {
	tp := mw.activeTab()
	if tp == nil {
		return
	}
	paths := mw.getSelectedPaths()
	if len(paths) != 1 {
		mw.statusLabel.SetText("Selecione exatamente um arquivo .7z")
		return
	}
	archive := paths[0]
	if !isArchiveFile(archive) {
		mw.statusLabel.SetText("Selecione um arquivo .7z para extrair")
		return
	}
	curPath := tp.currentPath()
	baseName := strings.TrimSuffix(filepath.Base(archive), filepath.Ext(archive))
	dest := filepath.Join(curPath, baseName)

	jobID, err := mw.api.AddExtractJob(archive, dest, "rename")
	if err != nil {
		mw.statusLabel.SetText(fmt.Sprintf("Erro ao enfileirar: %v", err))
		return
	}
	mw.statusLabel.SetText(fmt.Sprintf("⏳ Extraindo %s...", filepath.Base(archive)))
	go func() {
		mw.Synchronize(func() {
			mw.openProgressDialog(jobID, "extract")
		})
	}()
}

func (mw *GarqMainWindow) extractDialog(archive string) {
	curPath := ""
	if tp := mw.activeTab(); tp != nil {
		curPath = tp.currentPath()
	}

	baseName := strings.TrimSuffix(filepath.Base(archive), filepath.Ext(archive))
	defaultDest := filepath.Join(curPath, baseName)

	var destEdit *walk.LineEdit
	var rbReplace, rbRename *walk.RadioButton
	var acceptBtn, cancelBtn *walk.PushButton

	conflict := "replace"

	dlg, err := walk.NewDialog(mw.MainWindow)
	if err != nil {
		log.Printf("extractDialog: %v", err)
		return
	}
	dlg.SetTitle("Extrair arquivo")
	dlg.SetLayout(walk.NewVBoxLayout())
	dlg.SetIcon(mw.MainWindow.Icon())

	builder := NewBuilder(dlg)

	if err := (Composite{
		Layout: VBox{Margins: Margins{Left: 14, Top: 10, Right: 14, Bottom: 10}},
		Children: []Widget{
			Label{
				Text: fmt.Sprintf("Extrair: %s", filepath.Base(archive)),
				Font: Font{Bold: true, PointSize: 10},
			},
			Composite{
				Layout: Grid{Columns: 3},
				Children: []Widget{
					Label{Text: "Arquivo:"},
					LineEdit{
						Text:     archive,
						ReadOnly: true,
						MinSize:  Size{Width: 370},
					},
					Label{}, // spacer
					Label{Text: "Extrair para:"},
					LineEdit{
						AssignTo: &destEdit,
						Text:     defaultDest,
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
			Label{
				Text: "(a pasta será criada se não existir)",
			},
			GroupBox{
				Title: "Se a pasta de destino existir",
				Layout: HBox{},
				Children: []Widget{
					RadioButton{
						AssignTo: &rbReplace,
						Text:     "Substituir arquivos existentes",
						Value:    "replace",
						OnClicked: func() { conflict = "replace" },
					},
					RadioButton{
						AssignTo: &rbRename,
						Text:     "Renomear pasta extraída",
						Value:    "rename",
						OnClicked: func() { conflict = "rename" },
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
							dest := destEdit.Text()
							if dest == "" {
								walk.MsgBox(dlg, "Erro", "A pasta de destino não pode estar vazia", walk.MsgBoxIconError)
								return
							}
							jobID, err := mw.api.AddExtractJob(archive, dest, conflict)
							if err != nil {
								walk.MsgBox(dlg, "Erro", fmt.Sprintf("Erro ao enfileirar: %v", err), walk.MsgBoxIconError)
								return
							}
							dlg.Close(walk.DlgCmdOK)
							mw.statusLabel.SetText(fmt.Sprintf("⏳ Extraindo %s...", filepath.Base(archive)))
							go func() {
								mw.Synchronize(func() {
									mw.openProgressDialog(jobID, "extract")
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
		log.Printf("extractDialog: create error: %v", err)
		return
	}

	dlg.SetCancelButton(cancelBtn)
	dlg.SetDefaultButton(acceptBtn)
	rbReplace.SetChecked(true)

	w, h := 520, 300
	b := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{
		X: b.X + (b.Width-w)/2, Y: b.Y + (b.Height-h)/2,
		Width: w, Height: h,
	})
	dlg.Run()
}

func isArchiveFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".7z" || ext == ".zip" || ext == ".rar" || ext == ".tar" || ext == ".gz"
}
