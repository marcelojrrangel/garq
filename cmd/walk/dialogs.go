package main

import (
	"github.com/lxn/walk"
)

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
