package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
