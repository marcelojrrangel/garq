# Verifier Report — Front-end Melhorias

**Branch:** `feature/front-end-melhorias`  
**Commit range:** `9ff4875..b2440fc` (HEAD `b2440fc`)  
**Verifier:** independent validation against `spec.md`, `design.md`, `tasks.md`

---

## Verdict: PASS (with minor gaps)

All 11 requirement areas have implementation evidence. The full build/vet/test gate passes. Two low-risk spec-precision gaps were found and are documented below; neither blocks the feature.

---

## Per-AC Evidence

| Req | Status | Evidence |
|-----|--------|----------|
| **FX-01** Fix paste double-enqueue | PASS | `internal/service/clipboard.go:Paste` uses a mutually exclusive `switch` on the typed `ClipboardOp`: `cut` -> `s.api.AddMoveJob`, `copy` -> `s.api.AddCopyJob`. `service_test.go:TestPasteRoutesCutToMove` and `TestPasteRoutesCopyToCopy` assert the returned `jobType` and stored `snap.Type`. Empty/invalid clipboard returns an error and `fileops.go:pasteClipboard` shows it in the status bar. |
| **FX-02** Shift+Del confirmation | PASS | `cmd/walk/fileops.go:deleteSelectedPermanently` calls `walk.MsgBox(..., walk.MsgBoxYesNo\|walk.MsgBoxIconWarning\|walk.MsgBoxDefButton2)`. If the result is not `DlgCmdYes` it returns without deleting. On "Yes" it calls `mw.service.Delete(paths)` and opens a progress dialog. No selection -> status message, no dialog. |
| **FX-03** Async properties count | PASS | `cmd/walk/fileops.go:showProperties` builds the dialog immediately, sets `" Conteúdo: Contando..."` for directories, then launches a goroutine that calls `countDirContents` and updates the label via `mw.Synchronize`. A `closed` flag guards against writes after the dialog is closed. File targets show size immediately with no placeholder. |
| **FX-04** Drag-drop via worker | PASS | `cmd/walk/main.go:DropFiles` validates the drop, checks `filepath.Dir(src) == dest` -> `"Origem e destino iguais"`, then calls `mw.service.Copy(files, dest, "replace")` and `mw.openProgressDialog`. Empty/invalid drops show status and enqueue nothing. The progress dialog refreshes the active tab on `done`. |
| **FX-05** `internal/service` boundary | PASS | `cmd/walk` imports only `garq/internal/service` from the project (confirmed by import scan). `progress_dialog.go` polls via `pd.mw.service.JobSnapshot` and calls `service.PauseJob/ResumeJob/CancelJob`. `main.go` uses `service.ActiveJobCount()`. No `*sql.DB` is passed to UI structs. `cmd/app` builds and existing tests pass. |
| **FX-06** Migrate inline file ops | PASS | `createNewFolder` -> `service.CreateFolder`, `renameSelected` -> `service.Rename`, drop -> `service.Copy`, `cutSelected/copySelected` -> `service.SetClipboard`, `pasteClipboard` -> `service.Paste`. `copyPath`/`getCopyPath` are gone from `cmd/walk`. `grep` found no `os.MkdirAll`/`os.Rename` in `cmd/walk`. **Note:** `cmd/walk/recycle_other.go` (build tag `!windows`) still contains `os.RemoveAll`; this is a non-Windows fallback stub and is not compiled on Windows. |
| **FX-07** Split `main.go` | PASS | `main.go` is 186 LOC (≤ 200). New focused files exist: `types.go`, `tabs.go`, `navtree.go`, `navigation.go`, `fileops.go`, `dialogs.go`, `preview.go`, `shell_icon.go`, `utils.go`. All are < 400 LOC. `go build ./cmd/walk/` and `go vet ./cmd/walk/` pass. |
| **FX-08** Sort indicators | PASS | `cmd/walk/navigation.go:sortTab` updates each column title using `tp.columnTitles` plus `" ▲"` or `" ▼"` for the active `tp.sortBy`. Default `sortBy=0` / `sortDirAsc=true` in `tabs.go:newTab`, so "Nome ▲" appears at startup. `ColumnClicked` toggles direction and switches columns. |
| **FX-09** Empty state | PASS | `cmd/walk/navigation.go:updateStatusBar` shows `"Pasta vazia"` when the model has 0 items and the search box is empty; `"Nenhum item corresponde a '<filter>'"` when a filter returns 0 results; otherwise shows the normal item/selection count. |
| **FX-10** Disabled toolbar | PASS | `cmd/walk/navigation.go:updateToolbarState` enables `btnPaste` only when the service clipboard has paths; `btnRename` only for exactly 1 selection; `btnCut/btnCopy/btnDelete/btnCompress` for any selection; `btnExtract` only for 1 archive selection. It is attached to `SelectedIndexesChanged` and called from `onTabChanged`/`navigateTabDirect`. `btnCloseTab` is disabled when `len(mw.tabs) <= 1`. |
| **FX-11** ProgressDialog + tab focus | PASS | `cmd/walk/progress_dialog.go:update` sleeps `10 * time.Second` before auto-close in the `done` branch and does **not** schedule auto-close in the `failed` branch. `tabs.go:newTab` calls `tp.fileList.SetFocus()` after switching; `onTabChanged` does the same. `closeCurrentTab` returns early when only one tab remains, so `Ctrl+W` is a no-op. |

---

## Discrimination Sensor (Mutation Testing)

A temporary copy of the repo was mutated; the real repo was not changed.

| Fault | Mutation | Result |
|-------|----------|--------|
| 1 — Paste routing | Swapped `cut` -> `AddCopyJob` and `copy` -> `AddMoveJob` in `internal/service/clipboard.go` | `TestPasteRoutesCutToMove` and `TestPasteRoutesCopyToCopy` **failed** — killed. |
| 2 — ActiveJobCount | Changed query to `WHERE status IN ('running')`, excluding `pending` | `TestActiveJobCount` **failed** (`count = 0, want 1`) — killed. |
| 3 — Shift+Del confirmation | Removed the `walk.MsgBox` confirmation in `deleteSelectedPermanently` | Inspected only (no UI test harness). The original confirmation dialog is present and correct; removing it would be a visible spec violation. |

**Sensor result:** 2 behavior mutants killed / 0 survived. UI-only fault inspected and confirmed present in source.

The temporary copy was deleted after testing.

---

## Gate Verification

Run with the project-required CGO environment:

```powershell
$env:GOROOT="L:\sdk\go1.24.0"
$env:Path="L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"
$env:CC="L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"
$env:CGO_ENABLED='1'
go build ./cmd/walk/ ./cmd/app/
go vet ./cmd/walk/
go test ./compress/... ./internal/...
```

| Gate | Result |
|------|--------|
| `go build ./cmd/walk/ ./cmd/app/` | PASS |
| `go vet ./cmd/walk/` | PASS (no warnings) |
| `go test ./compress/... ./internal/...` | PASS — 27 test cases passed across `compress`, `internal/api`, `internal/copy`, `internal/db`, `internal/service` |

---

## Architecture Checks

| Check | Expected | Actual |
|-------|----------|--------|
| `grep -r '"database/sql"' cmd/walk/` | nothing | nothing — PASS |
| `grep -r 'garq/internal/(db\|worker\|api\|copy\|compress)' cmd/walk/` | nothing | nothing — PASS |
| `wc -l cmd/walk/main.go` | ≤ 200 | 186 — PASS |
| `grep -r 'copyPath\|getCopyPath' cmd/walk/` | nothing | nothing — PASS |
| `cmd/walk` backend imports | only `garq/internal/service` | confirmed — PASS |

---

## Gaps (Ranked)

1. **Non-Windows recycle stub still in `cmd/walk` with `os.RemoveAll`** (low risk)  
   `cmd/walk/recycle_other.go` (`//go:build !windows`) contains `os.RemoveAll`. It is not compiled on Windows and is not invoked by the migrated UI, but the FX-06 independent test (`grep -r "os.RemoveAll\|os.MkdirAll\|os.Rename" cmd/walk/`) does not return zero because of this file. Recommended cleanup: delete the now-dead `cmd/walk/recycle_other.go` and rely on `internal/shell/recycle_stub.go`.

2. **Clipboard cleared at enqueue time, not when the move job completes** (low risk)  
   FX-01 AC 4 says "WHEN cut-paste completes THEN clipboard SHALL be cleared". `service.Paste` clears the clipboard immediately after enqueuing the move job. This prevents accidental double-paste and matches the design intent, but it does not wait for the worker job to finish.

3. **Properties placeholder label has a leading space and inconsistent prefix** (cosmetic)  
   `showProperties` initializes the directory label as `" Conteúdo: Contando..."` (leading space) and later replaces it with `"Conteúdo: %d arquivo(s), %d pasta(s)"` (no leading space). A minor UI polish issue.

---

## Diff Range

`git diff --stat 9ff4875..b2440fc` — 28 files changed, ~3,287 insertions, ~1,409 deletions. New files: `internal/service/*`, `cmd/walk/{types,tabs,navtree,navigation,fileops,dialogs,preview,shell_icon,utils}.go`, `internal/shell/recycle_stub.go`. Removed/migrated: inline copy helpers and the old monolithic `main.go` body.
