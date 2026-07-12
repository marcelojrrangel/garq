# Front-end Melhorias Tasks

## Execution Protocol (MANDATORY -- do not skip)

Implement these tasks with the `tlc-spec-driven` skill: **activate it by name and follow its Execute flow and Critical Rules.** Do not search for skill files by filesystem path. The skill is the source of truth for the full flow (per-task cycle, sub-agent delegation, adequacy review, Verifier, discrimination sensor).

**If the skill cannot be activated, STOP and tell the user — do not proceed without it.**

---

**Design**: `.specs/features/front-end-melhorias/design.md`
**Status**: Draft

---

## Test Coverage Matrix

> Generated from codebase — existing tests: `compress/compress_test.go`, `internal/db/db_test.go`, `internal/copy/copy_test.go`, `internal/api/filesystem_test.go`. Guidelines found: `AGENTS.md` (build + test commands). Strong defaults applied for service layer (no existing service tests).

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| ---------- | ------------------ | -------------------- | ---------------- | ----------- |
| `internal/service` (business logic) | Unit | All branches; 1:1 to spec ACs; all listed edge cases | `internal/service/service_test.go` | `go test -v ./internal/service/...` |
| `cmd/walk` (UI) | None (build gate) | — (build + vet only; GUI untestable without app instance) | — | `go build ./cmd/walk/ && go vet ./cmd/walk/` |
| `cmd/app` (HTTP) | Build gate | Existing tests pass (regression) | — | `go build ./cmd/app/` |
| `internal/api`, `internal/db`, `internal/copy`, `internal/worker` | Build gate | Existing tests pass (regression) | — | `go test ./compress/... ./internal/...` |

## Gate Check Commands

> Generated from codebase (`AGENTS.md` specifies CGO env vars for build).

| Gate Level | When to Use | Command |
| ---------- | ----------- | ------- |
| Quick | After service unit tests | `$env:GOROOT="L:\sdk\go1.24.0"; $env:Path="L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"; $env:CC="L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"; $env:CGO_ENABLED='1'; go test -v ./internal/service/...` |
| Full | After service tests + regression | `$env:GOROOT="L:\sdk\go1.24.0"; $env:Path="L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"; $env:CC="L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"; $env:CGO_ENABLED='1'; go test -v ./internal/service/... ./compress/... ./internal/...` |
| Build | After UI changes or phase completion | `$env:GOROOT="L:\sdk\go1.24.0"; $env:Path="L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"; $env:CC="L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"; $env:CGO_ENABLED='1'; go build ./cmd/walk/ ./cmd/app/ && go vet ./cmd/walk/ && go test ./compress/... ./internal/...` |

---

## Execution Plan

Phases are ordered and run sequentially — each phase completes before the next begins, and tasks within a phase execute in order.

### Phase 1: P0 Quick Fixes (correctness + safety)

Independent bug fixes; no architectural changes needed. Every handler in this phase will be再造 created later to call the service layer, but the bugs are fixed here first for immediate safety.

```
T1 → T2 → T3 → T4
```

### Phase 2: Service Layer + DB Encapsulation

Architectural refactor: create the service boundary, encapsulate `*sql.DB`, migrate all UI code to call service instead of raw API/DB/worker.

```
T5 → T6 → T7 → T8 → T9 → T10
```

### Phase 3: UI Split + UX Polish

Split `main.go` into focused files, then add UX improvements (sort indicators, empty states, disabled toolbar, ProgressDialog/tab fixes).

```
T11 → T12 → T13 → T14 → T15 → T16 → T17
```

---

## Task Breakdown

### T1: Fix pasteClipboard double-enqueue bug

**What**: Fix the cut-paste handler to enqueue only one job (move for cut, copy for copy-only), removing the orphan copy job bug
**Where**: `cmd/walk/main.go` (function `pasteClipboard` at lines 1192-1246)
**Depends on**: None
**Reuses**: Existing `api.API.AddCopyJob`/`AddMoveJob` methods
**Requirement**: FX-01

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `pasteClipboard` uses if/else (not sequential assignment) for copy vs move job
- [ ] Cut-paste enqueues exactly 1 `move` job (0 copy jobs)
- [ ] Copy-paste enqueues exactly 1 `copy` job
- [ ] Empty/invalid clipboard shows status message and enqueues nothing
- [ ] Clipboard is cleared after cut-paste
- [ ] Build gate passes: `go build ./cmd/walk/`

**Tests**: none (UI handler; service test covers paste logic in T7)
**Gate**: build

---

### T2: Add Shift+Del confirmation dialog

**What**: Add a Yes/No confirmation `walk.MsgBox` before permanent deletion; route through `api.AddDeleteJob` instead of inline `os.RemoveAll`
**Where**: `cmd/walk/main.go` (function `deleteSelectedPermanently` at lines 1293-1310)
**Depends on**: None
**Reuses**: `api.API.AddDeleteJob` (wails.go:77), `walk.MsgBox` patterns
**Requirement**: FX-02

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `deleteSelectedPermanently` shows `walk.MsgBox` with Yes/No, default No, warning icon
- [ ] "No" / Esc → nothing deleted
- [ ] "Yes" → enqueues `api.AddDeleteJob(paths)` and opens progress dialog
- [ ] No selection → status message, no dialog
- [ ] Build gate passes: `go build ./cmd/walk/`

**Tests**: none (UI handler)
**Gate**: build

---

### T3: Async countDirContents in properties dialog

**What**: Move `countDirContents` to a goroutine with "Contando..." placeholder and `Synchronize` result update; guard against disposed dialog
**Where**: `cmd/walk/main.go` (function `showProperties` at lines 1341-1405)
**Depends on**: None
**Reuses**: `countDirContents` (main.go:1471-1484), `mw.Synchronize` pattern
**Requirement**: FX-03

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] Properties dialog for a directory appears immediately with "Conteúdo: Contando..." placeholder
- [ ] Counting runs in a goroutine; result replaces placeholder via `Synchronize`
- [ ] Dialog closed before count completes → goroutine exits without panic (guard with `pd.dialog == nil` or `mw.Synchronize` which is safe)
- [ ] Properties for a file → shows size immediately, no "Contando..."
- [ ] Build gate passes: `go build ./cmd/walk/`

**Tests**: none (UI dialog)
**Gate**: build

---

### T4: Replace drag-drop inline copy with worker job

**What**: Replace inline `copyPath` in the drop handler with `api.AddCopyJob` + progress dialog, so UI stays responsive
**Where**: `cmd/walk/main.go` (drop handler at lines 361-393)
**Depends on**: None
**Reuses**: `api.API.AddCopyJob`, `openProgressDialog` (main.go:1023)
**Requirement**: FX-04

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] Drop handler calls `mw.api.AddCopyJob(files, dest, "replace")` and opens progress dialog
- [ ] UI stays responsive during large drop operations (navigation, tabs functional)
- [ ] Empty/invalid drop → status message, no job enqueued
- [ ] Source == destination → "Origem e destino iguais" status, no job
- [ ] Active tab refreshes after copy completes (existing ProgressDialog auto-refresh)
- [ ] Build gate passes: `go build ./cmd/walk/`

**Tests**: none (UI handler)
**Gate**: build

---

### T5: Create internal/service package with core types and constructor

**What**: Create `internal/service/` package with `Service` struct, `New()` constructor, `types.go` (Entry, JobSnapshot, ClipboardOp types), and `jobcontrol.go` (JobSnapshot, ActiveJobCount, PauseJob, ResumeJob, CancelJob methods)
**Where**: `internal/service/service.go`, `internal/service/types.go`, `internal/service/jobcontrol.go`
**Depends on**: None
**Reuses**: `api.API` struct, `worker.PauseJob/ResumeJob/CancelJob` functions, `db.EnqueueJob` query patterns
**Requirement**: FX-05

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `Service` struct defined with `api *api.API` field
- [ ] `New(a *api.API) *Service` constructor exported
- [ ] `JobSnapshot` type defined (id, type, status, progress, errMsg, payload fields)
- [ ] `ClipboardOp` type defined (op string, paths []string)
- [ ] `Entry` type defined (name, path, isDir, size, modTime, mode)
- [ ] `JobSnapshot(id int64) (JobSnapshot, error)` method wraps the SQL query from `progress_dialog.go:56-63`
- [ ] `ActiveJobCount() (int, error)` method wraps the SQL query from `main.go:987`
- [ ] `PauseJob/ResumeJob/CancelJob` methods delegate to `worker.PauseJob/ResumeJob/CancelJob`
- [ ] Unit tests: `service_test.go` tests `JobSnapshot` and `ActiveJobCount` with temp SQLite DB
- [ ] Quick gate passes: `go test -v ./internal/service/...`

**Tests**: unit
**Gate**: quick

---

### T6: Encapsulate api.API.DB to private; add New() constructor

**What**: Make `api.API.DB` private (`db`), add `api.New(db *sql.DB) *API` constructor, update all callers (`cmd/walk/main.go`, `cmd/app/main.go`)
**Where**: `internal/api/wails.go`, `cmd/walk/main.go` (construction), `cmd/app/main.go`
**Depends on**: T5
**Reuses**: Existing `api.API` methods (they already use `a.DB` internally — change to `a.db`)
**Requirement**: FX-05

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `API.DB` field renamed to `db` (unexported)
- [ ] `api.New(db *sql.DB) *API` constructor exported
- [ ] All `a.DB` references in `wails.go` changed to `a.db`
- [ ] `cmd/walk/main.go` constructs API via `api.New(dbConn)` (not struct literal)
- [ ] `cmd/app/main.go` constructs API via `api.New(dbConn)` (or verified it already does)
- [ ] `go build ./cmd/walk/ ./cmd/app/` succeeds
- [ ] Existing tests pass: `go test ./compress/... ./internal/...`

**Tests**: none (refactor; existing tests are regression gate)
**Gate**: full

---

### T7: Add clipboard, file, and job methods to service layer

**What**: Implement `clipboard.go` (SetClipboard, GetClipboard, Paste), `files.go` (CreateFolder, Rename, OpenFile), `jobs.go` (Copy, Move, Compress, Extract, Delete, DeleteToRecycle) in `internal/service/`
**Where**: `internal/service/clipboard.go`, `internal/service/files.go`, `internal/service/jobs.go`
**Depends on**: T5
**Reuses**: `api.API.Add*Job` methods, `walk.Clipboard()` (in clipboard.go), `os.MkdirAll`/`os.Rename`/`exec.Command` (in files.go)
**Requirement**: FX-06

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `SetClipboard(op string, paths []string)` writes `ClipboardOp` JSON to `walk.Clipboard()`
- [ ] `GetClipboard() (ClipboardOp, error)` reads and parses from `walk.Clipboard()`
- [ ] `Paste(dest string) (jobID int64, jobType string, err error)` reads clipboard, routes cut→move, copy→copy
- [ ] `CreateFolder(parent, name string) error` wraps `os.MkdirAll`
- [ ] `Rename(oldPath, newName string) error` wraps `os.Rename`
- [ ] `OpenFile(path string) error` wraps `exec.Command("cmd","/c","start","",path)`
- [ ] `Copy/Move/Compress/Extract/Delete` delegate to `api.API.Add*Job`
- [ ] `DeleteToRecycle(paths []string) error` calls `RecycleItems` (platform-specific, stub for non-Windows)
- [ ] Unit tests: `Paste` routing (cut→move, copy→copy, empty→error), `CreateFolder`, `Rename` with temp dir
- [ ] Quick gate passes: `go test -v ./internal/service/...`

**Tests**: unit
**Gate**: quick

---

### T8: Migrate cmd/walk to use service layer (all handlers)

**What**: Replace all `mw.api.AddCopyJob/AddMoveJob/...`, `walk.Clipboard().SetText()`, inline `os.*` calls in `cmd/walk/` with `mw.service.Paste()`, `mw.service.CreateFolder()`, etc.
**Where**: `cmd/walk/main.go` (pasteClipboard, cutSelected, copySelected, createNewFolder, renameSelected, deleteSelected, deleteSelectedPermanently, activateSelected, drop handler), `cmd/walk/progress_dialog.go` (fetchJob→service.JobSnapshot, db.UpdateJobStatus→service.CancelJob, worker.PauseJob/ResumeJob→service.PauseJob/ResumeJob)
**Depends on**: T6, T7
**Reuses**: Service methods created in T5/T7
**Requirement**: FX-06

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `cmd/walk/main.go` adds `service *service.Service` field to `GarqMainWindow`
- [ ] `pasteClipboard` calls `mw.service.Paste(dest)` instead of parsing clipboard + enqueuing
- [ ] `cutSelected`/`copySelected` call `mw.service.SetClipboard(op, paths)`
- [ ] `createNewFolder` calls `mw.service.CreateFolder(parent, name)`
- [ ] `renameSelected` calls `mw.service.Rename(oldPath, newName)`
- [ ] `deleteSelectedPermanently` calls `mw.service.Delete(paths)` (enqueues job)
- [ ] `activateSelected` calls `mw.service.OpenFile(path)`
- [ ] `progress_dialog.go` `fetchJob` → `pd.mw.service.JobSnapshot(id)`
- [ ] `progress_dialog.go` `worker.PauseJob/ResumeJob/CancelJob` → `pd.mw.service.PauseJob/ResumeJob/CancelJob`
- [ ] `progress_dialog.go` `db.UpdateJobStatus` → `pd.mw.service.CancelJob(id)` + `db.UpdateJobStatus` removed
- [ ] `progress_dialog.go` no longer imports `database/sql` or `garq/internal/worker`
- [ ] `main.go` no longer imports `garq/internal/db` or `garq/internal/worker`
- [ ] `copyPath` and `getCopyPath` functions deleted from `main.go`
- [ ] Drop handler calls `mw.service.Copy(files, dest, "replace")`
- [ ] Build gate passes: `go build ./cmd/walk/ ./cmd/app/ && go vet ./cmd/walk/ && go test ./compress/... ./internal/...`
- [ ] `grep -r "garq/internal/db" cmd/walk/` returns zero results
- [ ] `grep -r "database/sql" cmd/walk/` returns zero results
- [ ] `grep -r "garq/internal/worker" cmd/walk/` returns zero results

**Tests**: none (UI migration; service tests already cover logic)
**Gate**: build

---

### T9: Remove direct api.API usage from cmd/walk (where service supersedes)

**What**: Replace `mw.api.ListDirectory` / `mw.api.ListRoots` calls in `cmd/walk/` with `mw.service.ListDirectory` / `mw.service.ListRoots` (service delegates to api). Add these methods to `service/files.go` if not already present.
**Where**: `internal/service/files.go` (add ListDirectory, ListRoots), `cmd/walk/main.go` (navigation.go content)
**Depends on**: T8
**Reuses**: `api.API.ListDirectory`, `api.API.ListRoots`
**Requirement**: FX-05

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `service.ListDirectory(path) ([]Entry, error)` method added (delegates to `api.API.ListDirectory`, converts `DirEntryDTO` → `Entry`)
- [ ] `service.ListRoots() ([]string, error)` method added (delegates to `api.API.ListRoots`)
- [ ] `cmd/walk/main.go` navigation calls `mw.service.ListDirectory` instead of `mw.api.ListDirectory`
- [ ] `cmd/walk/main.go` no longer accesses `mw.api.DB` field (it's private now from T6)
- [ ] `cmd/walk/main.go` does not import `garq/internal/api` directly (service is the only import)
- [ ] Build gate passes: `go build ./cmd/walk/ ./cmd/app/ && go vet ./cmd/walk/ && go test ./compress/... ./internal/...`
- [ ] `grep -r "garq/internal/api" cmd/walk/` returns zero results

**Tests**: none (thin delegation; service ListDirectory wrap verified in build)
**Gate**: build

---

### T10: Delete copyPath/getCopyPath and clean up unused imports

**What**: Remove dead code (`copyPath`, `getCopyPath`) left after migration; remove any now-unused imports from `main.go`
**Where**: `cmd/walk/main.go`
**Depends on**: T8
**Reuses**: NONE
**Requirement**: FX-06

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `copyPath` function deleted from `main.go`
- [ ] `getCopyPath` function deleted from `main.go`
- [ ] No unused imports in `main.go`
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`
- [ ] `grep -r "copyPath\|getCopyPath" cmd/walk/` returns zero results

**Tests**: none (dead code removal)
**Gate**: build

---

### T11: Split main.go — extract types.go

**What**: Move `FileEntry`, `TabPane`, `GarqMainWindow`, `NavItem`, `NavTreeModel`, `NavSection`, `FileTableModel`, `formatSize`, `formatDuration` (if in main.go) to `cmd/walk/types.go`
**Where**: `cmd/walk/types.go` (new), `cmd/walk/main.go` (remove moved code)
**Depends on**: T9
**Reuses**: NONE
**Requirement**: FX-07

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `types.go` contains all struct/type declarations moved from `main.go`
- [ ] `main.go` no longer contains those declarations
- [ ] Both files are `package main`
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`

**Tests**: none (mechanical move)
**Gate**: build

---

### T12: Split main.go — extract tabs.go, navtree.go, navigation.go

**What**: Move tab management to `tabs.go`, nav tree to `navtree.go`, navigation to `navigation.go`
**Where**: `cmd/walk/tabs.go`, `cmd/walk/navtree.go`, `cmd/walk/navigation.go` (new), `cmd/walk/main.go` (remove moved code)
**Depends on**: T11
**Reuses**: NONE
**Requirement**: FX-07

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `tabs.go` contains `newTab`, `closeCurrentTab`, `nextTab`, `onTabChanged`, `togglePreview`, `tabTitle`
- [ ] `navtree.go` contains `buildSectionedNavTree`, `makeNavItem`, `makeSection`, `onNavItemSelected`, `onNavItemActivated`
- [ ] `navigation.go` contains `navigateTo`, `navigateTabTo`, `goBack`, `goForward`, `goUp`, `updateNavButtons`, `updateStatusBar`, `updateActiveJobsStatus`, `refreshActiveTab`, `filterBySearch`, `containsIgnoreCase`
- [ ] `main.go` no longer contains these functions
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`

**Tests**: none (mechanical move)
**Gate**: build

---

### T13: Split main.go — extract fileops.go, dialogs.go, preview.go, shell_icon.go, utils.go

**What**: Move file operations to `fileops.go`, dialogs to `dialogs.go`, preview to `preview.go`, shell icon to `shell_icon.go`, utilities to `utils.go`
**Where**: `cmd/walk/fileops.go`, `cmd/walk/dialogs.go`, `cmd/walk/preview.go`, `cmd/walk/shell_icon.go`, `cmd/walk/utils.go` (new), `cmd/walk/main.go` (remove moved code)
**Depends on**: T12
**Reuses**: NONE
**Requirement**: FX-07

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `fileops.go` contains `getSelectedPaths`, `createNewFolder`, `cutSelected`, `copySelected`, `pasteClipboard`, `renameSelected`, `deleteSelected`, `deleteSelectedPermanently`, `activateSelected`, `showProperties`
- [ ] `dialogs.go` contains `showInputDialog`, `calcDialogWidth`, `openProgressDialog`
- [ ] `preview.go` contains `updatePreview` (and `imageExts`, `textExts` if present)
- [ ] `shell_icon.go` contains `getShellIcon`
- [ ] `utils.go` contains `countDirContents`
- [ ] `main.go` is ≤ 200 LOC (init + main + window layout only)
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`
- [ ] `wc -l cmd/walk/main.go` shows ≤ 200 lines

**Tests**: none (mechanical move)
**Gate**: build

---

### T14: Add sort indicators to column headers

**What**: In `sortTab`, update column header titles to include ▲/▼ suffix for the active sort column
**Where**: `cmd/walk/navigation.go` (post-split; the `sortTab` function)
**Depends on**: T13
**Reuses**: `walk.TableView` `Columns()` API
**Requirement**: FX-08

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] Clicking a column updates its title to `"Name ▲"` (ascending) or `"Name ▼"` (descending)
- [ ] Previous sort column loses its indicator
- [ ] Default (app start) shows "Nome ▲"
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`

**Tests**: none (UI-only)
**Gate**: build

---

### T15: Add empty state messages to status bar

**What**: In `updateStatusBar`, show "Pasta vazia" when directory has 0 items, or "Nenhum item corresponde a '<filter>'" when filter returns 0 results
**Where**: `cmd/walk/navigation.go` (post-split; the `updateStatusBar` function)
**Depends on**: T13
**Reuses**: `statusLabel.SetText`, `tp.searchText`
**Requirement**: FX-09

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] Empty directory → status bar shows "Pasta vazia"
- [ ] Filter with 0 results → status bar shows "Nenhum item corresponde a '<filter>'"
- [ ] Non-empty directory → status bar shows normal item count
- [ ] Filter cleared → status bar reverts to item count
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`

**Tests**: none (UI-only)
**Gate**: build

---

### T16: Add context-sensitive disabled toolbar buttons

**What**: Add `updateToolbarState(tp *TabPane)` that disables/enables toolbar buttons based on selection and clipboard state; call it on selection change and tab switch
**Where**: `cmd/walk/navigation.go` (post-split), `cmd/walk/tabs.go` (call on tab switch)
**Depends on**: T13
**Reuses**: `walk.PushButton.SetEnabled`, `walk.Clipboard().Text()`
**Requirement**: FX-10

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] `btnPaste` disabled when clipboard has no Garq content
- [ ] `btnRename` disabled when selection count != 1
- [ ] `btnCut`, `btnCopy`, `btnDelete` disabled when selection is empty
- [ ] `btnCompress` disabled when selection is empty
- [ ] `btnExtract` disabled when selection != 1 or selection is not an archive
- [ ] `updateToolbarState` called on `OnSelectedIndexesChanged` and `onTabChanged`
- [ ] `btnCloseTab` disabled when `len(mw.tabs) <= 1`
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/`

**Tests**: none (UI-only)
**Gate**: build

---

### T17: Fix ProgressDialog auto-close and tab focus management

**What**: Change ProgressDialog auto-close from 3s to 10s; never auto-close on failure; focus file list on new tab and tab switch
**Where**: `cmd/walk/progress_dialog.go` (auto-close), `cmd/walk/tabs.go` (focus management)
**Depends on**: T16
**Reuses**: `walk.SetFocus`, existing tab creation logic
**Requirement**: FX-11

**Tools**:
- MCP: NONE
- Skill: NONE

**Done when**:
- [ ] ProgressDialog auto-closes after 10 seconds on success (was 3)
- [ ] ProgressDialog does NOT auto-close on failure (stays until user clicks "Fechar")
- [ ] New tab creation → `tp.fileList.SetFocus()` after `SetCurrentIndex`
- [ ] Tab switch → active tab's `fileList.SetFocus()`
- [ ] Last tab → `btnCloseTab` disabled (from T16), `Ctrl+W` is a no-op
- [ ] Build gate passes: `go build ./cmd/walk/ && go vet ./cmd/walk/ && go test ./compress/... ./internal/...`

**Tests**: none (UI-only)
**Gate**: build

---

## Phase Execution Map

```
Phase 1:  T1 ──→ T2 ──→ T3 ──→ T4
Phase 2:  T5 ──→ T6 ──→ T7 ──→ T8 ──→ T9 ──→ T10
Phase 3:  T11 ──→ T12 ──→ T13 ──→ T14 ──→ T15 ──→ T16 ──→ T17
```

Execution is strictly sequential — there is no intra-phase parallelism. A single agent
(or batch worker) works one task at a time, in order.

**Batch packing (3 batches of ~5-7 tasks):**
- Batch 1 (Phase 1 + start of Phase 2): T1–T7 (7 tasks)
- Batch 2 (rest of Phase 2): T8–T10 + start of Phase 3 T11 (4 tasks, small batch)
  - Alternative: Batch 2 = T8–T13 (6 tasks, phase-aligned: finishes Phase 2 + first half of Phase 3)
- Batch 3 (rest of Phase 3): T14–T17 (4 tasks)

**Recommended: 3 workers, phase-aligned:**
- Worker 1: Phase 1 (T1–T4, 4 tasks)
- Worker 2: Phase 2 (T5–T10, 6 tasks)
- Worker 3: Phase 3 (T11–T17, 7 tasks)

---

## Task Granularity Check

| Task | Scope | Status |
|------|-------|--------|
| T1: Fix pasteClipboard double-enqueue | 1 function, 1 file | ✅ Granular |
| T2: Add Shift+Del confirmation | 1 function, 1 file | ✅ Granular |
| T3: Async countDirContents | 1 function, 1 file | ✅ Granular |
| T4: Drag-drop via worker job | 1 handler, 1 file | ✅ Granular |
| T5: Create service package core | 3 files, 1 package | ✅ Granular (cohesive package init) |
| T6: Encapsulate api.API.DB | 3 files (api + 2 cmd), 1 field rename | ✅ Granular |
| T7: Add clipboard/file/job methods to service | 3 files in same package | ✅ Granular (cohesive: implements the service API) |
| T8: Migrate cmd/walk to service | 2 files (main.go + progress_dialog.go), many function edits | ⚠️ OK (all are mechanical replacements of one pattern: `mw.api.X()` → `mw.service.X()`) |
| T9: Remove direct api usage | 2 files | ✅ Granular |
| T10: Delete dead code | 1 file, delete 2 functions | ✅ Granular |
| T11: Extract types.go | 2 files (new + modify main.go) | ✅ Granular |
| T12: Extract tabs/navtree/navigation.go | 4 files (3 new + modify main.go) | ⚠️ OK (cohesive: one split operation, 3 sibling files) |
| T13: Extract fileops/dialogs/preview/shell_icon/utils.go | 6 files (5 new + modify main.go) | ⚠️ OK (cohesive: completing the split in one pass; all are mechanical moves) |
| T14: Sort indicators | 1 function, 1 file | ✅ Granular |
| T15: Empty state | 1 function, 1 file | ✅ Granular |
| T16: Disabled toolbar buttons | 1 new function + 2 call sites | ✅ Granular |
| T17: ProgressDialog + tab focus | 2 files, small edits | ⚠️ OK (two related UX fixes, same phase) |

---

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
|------|----------------------|---------------|--------|
| T1 | None | No incoming arrows | ✅ Match |
| T2 | None | No incoming arrows | ✅ Match |
| T3 | None | No incoming arrows | ✅ Match |
| T4 | None | No incoming arrows | ✅ Match |
| T5 | None | Phase 2 start, no deps within phase 2 | ✅ Match |
| T6 | T5 | T5 → T6 | ✅ Match |
| T7 | T5 | T5 → T7 | ✅ Match |
| T8 | T6, T7 | T6 → T8, T7 → T8 | ✅ Match |
| T9 | T8 | T8 → T9 | ✅ Match |
| T10 | T8 | T8 → T10 | ✅ Match |
| T11 | T9 | T9 → T11 (phase 2 end → phase 3 start) | ✅ Match |
| T12 | T11 | T11 → T12 | ✅ Match |
| T13 | T12 | T12 → T13 | ✅ Match |
| T14 | T13 | T13 → T14 | ✅ Match |
| T15 | T13 | T13 → T15 | ✅ Match |
| T16 | T13 | T13 → T16 | ✅ Match |
| T17 | T16 | T16 → T17 | ✅ Match |

**Note**: The diagram (phase execution map) shows sequential execution within each phase; the cross-check confirms every `Depends on` in the task body has a corresponding arrow in the phase map.

---

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
|------|---------------------------|-----------------|-----------|--------|
| T1 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T2 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T3 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T4 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T5 | internal/service | Unit | unit | ✅ OK |
| T6 | internal/api | None (build gate; refactor) | none | ✅ OK |
| T7 | internal/service | Unit | unit | ✅ OK |
| T8 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T9 | internal/service + cmd/walk | None (build gate; thin delegation) | none | ✅ OK |
| T10 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T11 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T12 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T13 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T14 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T15 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T16 | cmd/walk (UI) | None (build gate) | none | ✅ OK |
| T17 | cmd/walk (UI) | None (build gate) | none | ✅ OK |

**Rules check:**
- T5 and T7 create `internal/service` code → matrix requires unit tests → both have `Tests: unit` ✅
- All UI tasks (T1–T4, T8–T17) modify `cmd/walk` → matrix says "None (build gate)" → all have `Tests: none` ✅
- No test deferral or violation detected.

---

## Tools Per Task

All tasks use:
- **MCP**: NONE (no external API calls needed; all work is local Go code)
- **Skill**: NONE (the `tlc-spec-driven` skill is the execution protocol, not a per-task tool)

**Exception**: The Verifier (feature-level validation after T17) uses `tlc-spec-driven`'s `validate.md` checklist.

---

## Commit Message Format

All commits follow Conventional Commits:
```
<type>(<scope>): <description>
```

| Task | Type | Scope | Description |
|------|------|-------|------------|
| T1 | fix | clipboard | prevent double-enqueue in pasteClipboard |
| T2 | fix | ui | add confirmation dialog for permanent delete |
| T3 | fix | ui | make countDirContents async in properties dialog |
| T4 | fix | ui | route drag-drop through worker queue |
| T5 | feat | service | create internal/service package with jobcontrol |
| T6 | refactor | api | encapsulate DB field and add New constructor |
| T7 | feat | service | add clipboard, file, and job methods |
| T8 | refactor | ui | migrate all handlers to service layer |
| T9 | refactor | ui | remove direct api imports from cmd/walk |
| T10 | refactor | ui | delete dead copyPath and getCopyPath |
| T11 | refactor | ui | extract type declarations to types.go |
| T12 | refactor | ui | extract tabs, navtree, navigation to separate files |
| T13 | refactor | ui | extract fileops, dialogs, preview, utils to separate files |
| T14 | feat | ui | add sort indicators to column headers |
| T15 | feat | ui | add empty state messages to status bar |
| T16 | feat | ui | disable toolbar buttons based on selection state |
| T17 | fix | ui | extend progress auto-close to 10s and fix tab focus |