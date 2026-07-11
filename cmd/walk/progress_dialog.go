//go:build windows

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"

	"garq/internal/db"
	"garq/internal/worker"
)

// ---------------------------------------------------------------------------
// Win32 — controle de cor/estado da ProgressBar
// ---------------------------------------------------------------------------

const (
	pbmSetState  = 0x0410
	pbmSetMarquee = 0x040B
	pbstNormal   = 0x0001 // verde
	pbstError    = 0x0002 // vermelho
	pbstPaused   = 0x0003 // cinza/amarelo

	pbsMarquee = 0x08 // estilo marquee (indeterminado)
)

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	procSendMsg = modUser32.NewProc("SendMessageW")
)

func pbSetState(pb *walk.ProgressBar, state uintptr) {
	procSendMsg.Call(uintptr(pb.Handle()), pbmSetState, state, 0)
}

// ---------------------------------------------------------------------------
// JobSnapshot — leitura do banco para um job específico
// ---------------------------------------------------------------------------

type JobSnapshot struct {
	ID       int64
	Type     string
	Status   string
	Progress float64
	ErrMsg   string
	Payload  string // JSON raw
}

func fetchJob(dbConn *sql.DB, id int64) (JobSnapshot, error) {
	row := dbConn.QueryRow(
		`SELECT id, type, status, progress, COALESCE(error,''), COALESCE(payload,'') FROM jobs WHERE id=?`, id,
	)
	var j JobSnapshot
	err := row.Scan(&j.ID, &j.Type, &j.Status, &j.Progress, &j.ErrMsg, &j.Payload)
	return j, err
}

// ---------------------------------------------------------------------------
// ProgressDialog
// ---------------------------------------------------------------------------

type ProgressDialog struct {
	dialog      *walk.Dialog
	mw          *GarqMainWindow
	dbConn      *sql.DB
	jobID       int64

	// Widgets
	labelTitle  *walk.Label
	labelFile   *walk.Label
	progressBar *walk.ProgressBar
	labelStats  *walk.Label
	labelError  *walk.Label
	btnPause    *walk.PushButton
	btnCancel   *walk.PushButton
	btnDetails  *walk.PushButton
	detailsPane *walk.Composite
	labelFrom   *walk.Label
	labelTo     *walk.Label
	labelBytes  *walk.Label

	// Estado
	expanded    bool
	paused      bool
	startTime   time.Time
	lastPct     int
	doneAt      time.Time
}

// newProgressDialog cria e exibe o dialog não-modal para um job.
func newProgressDialog(mw *GarqMainWindow, dbConn *sql.DB, jobID int64, jobType string) *ProgressDialog {
	pd := &ProgressDialog{
		mw:        mw,
		dbConn:    dbConn,
		jobID:     jobID,
		startTime: time.Now(),
	}

	title := opTitle(jobType)

	dlg, err := walk.NewDialog(mw.MainWindow)
	if err != nil {
		log.Printf("ProgressDialog: %v", err)
		return nil
	}
	dlg.SetTitle(title)
	dlg.SetLayout(walk.NewVBoxLayout())
	pd.dialog = dlg

	builder := NewBuilder(dlg)

	var detailsPane *walk.Composite
	if err := (Composite{
		Layout: VBox{Margins: Margins{Left: 10, Top: 8, Right: 10, Bottom: 8}},
		Children: []Widget{
			// Título da operação
			Label{
				AssignTo: &pd.labelTitle,
				Text:     title,
				Font:     Font{Bold: true, PointSize: 10},
			},
			// Arquivo atual
			Label{
				AssignTo: &pd.labelFile,
				Text:     "Preparando...",
				MinSize:  Size{Width: 420},
			},
			// Barra de progresso
			ProgressBar{
				AssignTo: &pd.progressBar,
				MinSize:  Size{Width: 420, Height: 18},
				MaxValue: 100,
				Value:    0,
			},
			// Linha de stats: velocidade · tempo restante · N de N
			Label{
				AssignTo: &pd.labelStats,
				Text:     "",
			},
			// Erro (oculto até precisar)
			Label{
				AssignTo: &pd.labelError,
				Text:     "",
			},
			// Botões Pausar + Cancelar
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo: &pd.btnPause,
						Text:     "⏸  Pausar",
						MinSize:  Size{Width: 90},
						OnClicked: func() { pd.togglePause() },
					},
					PushButton{
						AssignTo: &pd.btnCancel,
						Text:     "✕  Cancelar",
						MinSize:  Size{Width: 100},
						OnClicked: func() { pd.cancel() },
					},
				},
			},
			// Botão expandir detalhes
			PushButton{
				AssignTo:  &pd.btnDetails,
				Text:      "∨  Detalhes",
				OnClicked: func() { pd.toggleDetails() },
			},
			// Painel de detalhes (oculto por padrão)
			Composite{
				AssignTo: &detailsPane,
				Layout:   VBox{},
				Visible:  false,
				Children: []Widget{
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{Text: "De:", MinSize: Size{Width: 35}},
							Label{AssignTo: &pd.labelFrom, Text: ""},
						},
					},
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{Text: "Para:", MinSize: Size{Width: 35}},
							Label{AssignTo: &pd.labelTo, Text: ""},
						},
					},
					Label{AssignTo: &pd.labelBytes, Text: ""},
				},
			},
		},
	}).Create(builder); err != nil {
		log.Printf("ProgressDialog: create error: %v", err)
		return nil
	}

	pd.detailsPane = detailsPane

	// Preenche De/Para a partir do payload do job
	j, err := fetchJob(dbConn, jobID)
	if err == nil {
		from, to := pathsFromPayload(j.Payload, j.Type)
		if from != "" {
			pd.labelFrom.SetText(truncPath(from, 55))
		}
		if to != "" {
			pd.labelTo.SetText(truncPath(to, 55))
		}
	}

	// Dimensões iniciais e posição centrada
	w, h := 480, 240
	mwB := mw.Bounds()
	dlg.SetBounds(walk.Rectangle{
		X:      mwB.X + (mwB.Width-w)/2,
		Y:      mwB.Y + (mwB.Height-h)/2,
		Width:  w,
		Height: h,
	})

	// Abre não-modal (Show, não Run)
	dlg.Show()

	// Inicia polling de atualização
	pd.startPolling()

	return pd
}

// ---------------------------------------------------------------------------
// Polling e atualização de estado
// ---------------------------------------------------------------------------

func (pd *ProgressDialog) startPolling() {
	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if pd.dialog == nil {
				return
			}
			j, err := fetchJob(pd.dbConn, pd.jobID)
			if err != nil {
				continue
			}
			pd.mw.Synchronize(func() {
				pd.update(j)
			})
			// Para de pollar após conclusão ou falha
			if j.Status == "done" || j.Status == "failed" {
				return
			}
		}
	}()
}

func (pd *ProgressDialog) update(j JobSnapshot) {
	if pd.dialog == nil {
		return
	}

	pct := int(j.Progress * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	pd.progressBar.SetValue(pct)
	pd.lastPct = pct

	elapsed := time.Since(pd.startTime)

	switch j.Status {
	case "pending":
		pd.labelFile.SetText("Aguardando na fila...")
		pd.labelStats.SetText("")
		pbSetState(pd.progressBar, pbstNormal)

	case "running":
		pbSetState(pd.progressBar, pbstNormal)
		pd.labelFile.SetText(opTitle(j.Type) + "...")

		// Estima tempo restante
		remaining := ""
		if pct > 2 && elapsed.Seconds() > 0 {
			totalEst := elapsed.Seconds() / (float64(pct) / 100.0)
			rem := time.Duration((totalEst - elapsed.Seconds()) * float64(time.Second))
			if rem > 0 {
				remaining = fmt.Sprintf("· ~%s restante", fmtDuration(rem))
			}
		}
		pd.labelStats.SetText(fmt.Sprintf("%d%%  %s", pct, remaining))
		pd.labelError.SetText("")

	case "done":
		pbSetState(pd.progressBar, pbstNormal)
		pd.progressBar.SetValue(100)
		pd.labelTitle.SetText("✓  Concluído")
		pd.labelFile.SetText(fmt.Sprintf("Operação concluída em %s", fmtDuration(elapsed)))
		pd.labelStats.SetText("")
		pd.labelError.SetText("")
		pd.btnPause.SetEnabled(false)
		pd.btnCancel.SetText("Fechar")
		// Auto-fecha após 3 segundos e atualiza a aba ativa
		if pd.doneAt.IsZero() {
			pd.doneAt = time.Now()
			// Atualiza a listagem imediatamente
			pd.mw.refreshActiveTab()
			go func() {
				time.Sleep(3 * time.Second)
				pd.mw.Synchronize(func() {
					if pd.dialog != nil {
						pd.dialog.Close(walk.DlgCmdOK)
					}
				})
			}()
		}

	case "failed":
		pbSetState(pd.progressBar, pbstError)
		pd.labelTitle.SetText("✗  Falhou")
		pd.labelFile.SetText("A operação foi interrompida")
		pd.labelStats.SetText("")
		if j.ErrMsg != "" {
			pd.labelError.SetText("⚠  " + j.ErrMsg)
		}
		pd.btnPause.SetEnabled(false)
		pd.btnCancel.SetText("Fechar")
	}

	// Atualiza statusbar da janela principal com contagem de jobs ativos
	pd.mw.updateActiveJobsStatus()
}

// ---------------------------------------------------------------------------
// Ações dos botões
// ---------------------------------------------------------------------------

func (pd *ProgressDialog) togglePause() {
	if pd.paused {
		worker.ResumeJob(pd.jobID)
		pd.paused = false
		pd.btnPause.SetText("⏸  Pausar")
		pbSetState(pd.progressBar, pbstNormal)
		pd.labelStats.SetText("Retomando...")
	} else {
		worker.PauseJob(pd.jobID)
		pd.paused = true
		pd.btnPause.SetText("▶  Retomar")
		pbSetState(pd.progressBar, pbstPaused)
		pd.labelStats.SetText("⏸  Pausado")
	}
}

func (pd *ProgressDialog) cancel() {
	// Se já está em estado final, fecha apenas
	if pd.btnCancel.Text() == "Fechar" {
		pd.dialog.Close(walk.DlgCmdCancel)
		return
	}
	worker.CancelJob(pd.jobID)
	db.UpdateJobStatus(pd.dbConn, pd.jobID, "failed", float64(pd.lastPct)/100.0, "cancelado pelo usuário")
	pd.dialog.Close(walk.DlgCmdCancel)
}

func (pd *ProgressDialog) toggleDetails() {
	pd.expanded = !pd.expanded
	pd.detailsPane.SetVisible(pd.expanded)

	b := pd.dialog.Bounds()
	if pd.expanded {
		pd.btnDetails.SetText("∧  Detalhes")
		b.Height += 90
	} else {
		pd.btnDetails.SetText("∨  Detalhes")
		b.Height -= 90
	}
	pd.dialog.SetBounds(b)
	pd.dialog.RequestLayout()
}

// SetPaths preenche os labels De/Para no painel de detalhes.
func (pd *ProgressDialog) SetPaths(from, to string) {
	if pd.labelFrom != nil {
		pd.labelFrom.SetText(truncPath(from, 55))
	}
	if pd.labelTo != nil {
		pd.labelTo.SetText(truncPath(to, 55))
	}
}

// ---------------------------------------------------------------------------
// Utilitários
// ---------------------------------------------------------------------------

// pathsFromPayload extrai os caminhos De/Para do JSON do payload do job.
func pathsFromPayload(payload string, jobType string) (from, to string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return "", ""
	}
	switch jobType {
	case "copy", "move", "compress":
		// sources é []string
		if sources, ok := m["sources"].([]any); ok && len(sources) > 0 {
			if s, ok := sources[0].(string); ok {
				from = s
				if len(sources) > 1 {
					from = fmt.Sprintf("%s (+%d)", s, len(sources)-1)
				}
			}
		}
		if dest, ok := m["dest"].(string); ok {
			to = dest
		}
	case "extract":
		if archive, ok := m["archive"].(string); ok {
			from = archive
		}
		if dest, ok := m["dest"].(string); ok {
			to = dest
		}
	case "delete":
		if sources, ok := m["sources"].([]any); ok && len(sources) > 0 {
			if s, ok := sources[0].(string); ok {
				from = s
			}
		}
		to = "🗑 Excluindo permanentemente"
	}
	return
}

func opTitle(jobType string) string {
	switch jobType {
	case "copy":
		return "Copiando arquivos"
	case "move":
		return "Movendo arquivos"
	case "delete":
		return "Excluindo arquivos"
	case "compress":
		return "Comprimindo arquivos"
	case "extract":
		return "Extraindo arquivos"
	default:
		return "Processando"
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func truncPath(p string, max int) string {
	if len(p) <= max {
		return p
	}
	return "..." + p[len(p)-(max-3):]
}
