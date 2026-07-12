# Front-end Melhorias — Architecture Design

## Decision Log

| ID | Decision | Rationale | Date |
|----|----------|-----------|------|
| AD-01 | `internal/service` wraps existing `api.API`, not replaces it | `cmd/app` HTTP server keeps working; `api.API` becomes the REST adapter's data layer; service is the logic layer | 2026-07-11 |
| AD-02 | Lightweight ops (rename, mkdir) run synchronously via service, not via worker queue | These are instant `os.*` calls; queue adds overhead with no UX benefit; the service abstracts the call so it can be swapped to async later if needed | 2026-07-11 |
| AD-03 | Delete-permanent routes through `api.AddDeleteJob` (worker queue) | Unifies with worker model; gives progress dialog; no UI freeze; existing `runDeleteJob` handler already works | 2026-07-11 |
| AD-04 | Progress dialog keeps DB polling but through `service.JobSnapshot(id)` instead of raw SQL | Push channel (`ObserveJob(id) <-chan Snapshot`) is a larger effort; polling-through-service still removes raw SQL from UI and is the 80/20 win | 2026-07-11 |
| AD-05 | `main.go` split is a pure mechanical move (no behavior changes) | Minimizes risk; each file gets `package main`; the split happens after service migration so file ops are already thin | 2026-07-11 |
| AD-06 | Clipboard protocol becomes a typed struct in service, not a string format | `"CUT:\n" + paths` is error-prone; `ClipboardOp{Op: "cut", Paths: []string}` is typed and validated | 2026-07-11 |
| AD-07 | UI-only changes (sort indicators, empty state, disabled states, focus) gate on build+vet only | Go GUI code is untestable without a running app instance; the service layer carries the unit-testable logic | 2026-07-11 |

---

## Components

### 1. `internal/service` — the single UI/backend boundary

```
internal/service/
├── service.go       # Service struct + constructor; holds *api.API, *worker deps
├── files.go         # ListDirectory, ListRoots, CreateFolder, Rename, OpenFile
├── jobs.go          # Copy, Move, Paste, Compress, Extract, Delete, DeletePermanent
├── jobcontrol.go    # PauseJob, ResumeJob, CancelJob, JobSnapshot, ActiveJobCount
├── clipboard.go     # ClipboardOp struct, SetClipboard, GetClipboard, Paste
└── types.go         # Entry, JobSnapshot, ClipboardOp types
```

#### `service.go` — core

```go
package service

type Service struct {
    api *api.API
}

func New(a *api.API) *Service {
    return &Service{api: a}
}
```

The `Service` struct wraps `api.API` and exposes every operation the UI needs. The UI
constructs `service.New(apiInstance)` once in `main()` and passes `*service.Service`
around instead of `*api.API`.

#### `jobcontrol.go` — replaces raw SQL + worker globals

```go
func (s *Service) JobSnapshot(id int64) (JobSnapshot, error) {
    // wraps the raw SQL query that was in progress_dialog.go:fetchJob
    // query: SELECT id, type, status, progress, COALESCE(error,''), COALESCE(payload,'')
}

func (s *Service) ActiveJobCount() (int, error) {
    // wraps the raw SQL query that was in main.go:updateActiveJobsStatus
    // query: SELECT COUNT(*) FROM jobs WHERE status IN ('pending','running')
}

func (s *Service) PauseJob(id int64)  { worker.PauseJob(id) }
func (s *Service) ResumeJob(id int64) { worker.ResumeJob(id) }
func (s *Service) CancelJob(id int64) { worker.CancelJob(id) }
```

#### `clipboard.go` — typed clipboard protocol

```go
type ClipboardOp struct {
    Op    string   // "cut" or "copy"
    Paths []string
}

func (s *Service) SetClipboard(op string, paths []string)
func (s *Service) GetClipboard() (ClipboardOp, error)
func (s *Service) Paste(dest string) (jobID int64, jobType string, err error)
```

`Paste` encapsulates the clipboard parse + enqueue logic that was inline in
`pasteClipboard` at `main.go:1192-1246`. The double-enqueue bug becomes structurally
impossible because `Paste` takes the operation from the clipboard struct, not from a
branch that calls both `AddCopyJob` and `AddMoveJob`.

#### `files.go` — directory operations

```go
func (s *Service) CreateFolder(parent, name string) error
func (s *Service) Rename(oldPath, newName string) error
func (s *Service) OpenFile(path string) error
```

These wrap the inline `os.MkdirAll`, `os.Rename`, `exec.Command` calls that were in
`main.go:1145-1274,800-815`.

#### `jobs.go` — async operations (via worker queue)

```go
func (s *Service) Copy(sources []string, dest, conflict string) (int64, error)
func (s *Service) Move(sources []string, dest, conflict string) (int64, error)
func (s *Service) Compress(sources []string, dest, conflict string) (int64, error)
func (s *Service) Extract(archive, dest, conflict string) (int64, error)
func (s *Service) Delete(sources []string) (int64, error)         // permanent
func (s *Service) DeleteToRecycle(paths []string) error          // sync, Win32
```

These delegate to `api.API.Add*Job` methods. `DeleteToRecycle` wraps `RecycleItems`
(Win32 SHFileOperation) synchronously — recycle is fast and undoable via the shell.

### 2. `api.API` — encapsulate DB

```go
// Before (wails.go:16)
type API struct {
    DB         *sql.DB       // PUBLIC — leaks to UI
    Ctx        context.Context
    HTTPServer *http.Server
}

// After
type API struct {
    db         *sql.DB       // PRIVATE — service accesses via methods
    Ctx        context.Context
    HTTPServer *http.Server
}
```

`New(db *sql.DB) *API` constructor replaces struct-literal construction. The UI calls
`api.New(dbConn)`, then `service.New(apiInstance)`. No `*sql.DB` ever crosses into
`cmd/walk`.

### 3. `cmd/walk` file split

| New file | Content moved from `main.go` | ~LOC |
|----------|-------------------------------|------|
| `types.go` | `FileEntry`, `TabPane`, `GarqMainWindow`, `NavItem`, `NavTreeModel`, `FileTableModel`, `formatSize` (lines 40-211) | ~200 |
| `main.go` (slimmed) | `init()`, `main()` composition root + window layout (lines 27-35, 253-418) | ~180 |
| `tabs.go` | `newTab`, `closeCurrentTab`, `nextTab`, `onTabChanged`, `tabTitle` (lines 424-678, 244-251) | ~280 |
| `navtree.go` | `buildSectionedNavTree`, `makeNavItem`, `makeSection`, `onNavItemSelected`, `onNavItemActivated` (lines 734-798) | ~90 |
| `navigation.go` | `navigateTo`, `navigateTabTo`, `goBack`, `goForward`, `goUp`, `updateNavButtons`, `updateStatusBar`, `updateActiveJobsStatus`, `refreshActiveTab`, `filterBySearch` (lines 817-1049) | ~250 |
| `preview.go` | `updatePreview` (lines 1057-1118) | ~65 |
| `fileops.go` | `getSelectedPaths`, `createNewFolder`, `cutSelected`, `copySelected`, `pasteClipboard`, `renameSelected`, `deleteSelected`, `deleteSelectedPermanently`, `activateSelected`, `showProperties` (lines 1124-1405) — all thinned to service calls | ~285 |
| `dialogs.go` | `showInputDialog`, `calcDialogWidth`, `openProgressDialog` (lines 1341-1465) | ~130 |
| `shell_icon.go` | `getShellIcon` (lines 222-242) | ~25 |
| `utils.go` | `countDirContents` (stays; used by properties dialog) | ~20 |

`copyPath` and `getCopyPath` are **deleted** — migrated to `service.Copy`. Drag-drop
calls `service.Copy` directly.

### 4. Dependency graph (target)

```
cmd/walk/main.go
    └── internal/service          ← ONLY backend import
            ├── internal/api
            │     └── internal/db
            ├── internal/worker
            └── compress/, internal/copy  (via worker interfaces)
```

### 5. UI improvement details

#### Sort indicators

In `sortTab(tp *TabPane)` at `main.go:1312-1337`, after sort, update the clicked
column's `Title` field:

```go
// Pseudocode — iterate over tp.columns, set Title to base name + ▲/▼
for i := range tp.columns {
    base := tp.columns[i].BaseName // "Nome", "Tamanho", etc.
    if i == tp.sortBy {
        if tp.sortDirAsc { tp.columns[i].Title = base + " ▲" }
        else             { tp.columns[i].Title = base + " ▼" }
    } else {
        tp.columns[i].Title = base
    }
}
tp.fileList.Columns().Clear()
// re-add columns with new titles
```

#### Empty state

In `updateStatusBar()` at `main.go:970-1007`, after setting model:
- If `len(entries) == 0` and `searchText == ""` → show "Pasta vazia"
- If `len(entries) == 0` and `searchText != ""` → show
  `"Nenhum item corresponde a '<filter>'"`
- Otherwise → show normal count

#### Disabled toolbar buttons

In `navigation.go` (post-split), add `updateToolbarState(tp *TabPane)`:
- Called on `OnSelectedIndexesChanged`, `onTabChanged`, and `navigateTabTo`
- Checks selection count and clipboard state
- Sets `btnPaste.SetEnabled(hasClipboard)`, `btnRename.SetEnabled(len(sel)==1)`,
  `btnCut/SetEnabled(len(sel)>0)`, etc.

#### ProgressDialog auto-close

In `progress_dialog.go:320-327`, change `3 * time.Second` to `10 * time.Second`.
In the `"failed"` case (line 331-340), do NOT set the auto-close goroutine — the
dialog stays until user clicks "Fechar".

#### Tab focus management

In `tabs.go` (post-split):
- `newTab()`: after `tabWidget.SetCurrentIndex(idx)`, call
  `tp.fileList.SetFocus()`
- `onTabChanged()`: after switching, call `activeTab().fileList.SetFocus()`
- In `closeCurrentTab()`: `if len(mw.tabs) <= 1 { return }` guard at top

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| `api.API.DB` → private breaks `cmd/app` | `cmd/app` uses `api.API` via method calls, not direct field access; constructor `api.New(db)` replaces literal; verify with `go build ./cmd/app/` |
| Service layer adds a layer of indirection for simple calls | Thin wrappers (1-line delegation); the indirection cost is a function call; the benefit is a clean test boundary |
| `main.go` split could introduce import cycles | All files are `package main`; no new packages; Go compiler catches cycles at build time |
| `countDirContents` goroutine writes to disposed dialog | Guard with `if pd.dialog == nil { return }` in the `Synchronize` callback; or use a `context.Context` with cancel |
| Worker global state still accessed by service | Phase 2 replaces `worker.PauseJob` etc. calls in UI with `service.PauseJob`; service delegates to worker; UI no longer imports worker |

---

## Test Strategy

| Layer | Test Type | What to test | How |
|-------|-----------|--------------|-----|
| `internal/service` | Unit (table tests) | Clipboard parse/set/get, Paste routing logic (cut→move, copy→copy), CreateFolder, Rename, JobSnapshot query, ActiveJobCount query | `internal/service/service_test.go` — uses temp SQLite DB, real filesystem temp dir |
| `cmd/walk` (UI) | Build + vet gate | Compiles, no vet warnings, no raw SQL, no backend imports | `go build ./cmd/walk/ && go vet ./cmd/walk/` |
| `cmd/app` (HTTP) | Existing tests pass | `internal/api`, `internal/db`, `internal/copy`, `compress` tests still pass | `go test ./compress/... ./internal/...` |

Service tests are the critical path: they verify the business logic that was
previously inline in the UI. The service is pure Go (no walk dependency), so it's
fully unit-testable.

---

## Package Structure (Target)

```
garq/
├── cmd/
│   ├── walk/
│   │   ├── main.go            # ~180 LOC: init, main, window layout
│   │   ├── types.go           # structs, models, formatSize
│   │   ├── tabs.go            # tab lifecycle + focus
│   │   ├── navtree.go         # nav tree build + handlers
│   │   ├── navigation.go      # navigate, refresh, search, toolbar state
│   │   ├── preview.go         # preview pane
│   │   ├── fileops.go         # thin handlers → service calls
│   │   ├── dialogs.go         # input + properties dialogs
│   │   ├── progress_dialog.go # polling via service.JobSnapshot
│   │   ├── compress_dialog.go # compress options
│   │   ├── extract_dialog.go  # extract options
│   │   ├── folder_dialog.go   # SHBrowseForFolder (stays)
│   │   ├── recycle_windows.go # SHFileOperation (stays; called via service)
│   │   ├── shell_icon.go      # getShellIcon
│   │   └── utils.go           # countDirContents
│   └── app/
│       └── main.go            # HTTP server (uses api.New)
│
├── internal/
│   ├── service/               # NEW
│   │   ├── service.go
│   │   ├── files.go
│   │   ├── jobs.go
│   │   ├── jobcontrol.go
│   │   ├── clipboard.go
│   │   ├── types.go
│   │   └── service_test.go
│   ├── api/                   # DB becomes private
│   │   ├── wails.go
│   │   ├── http.go
│   │   ├── filesystem.go
│   │   └── types.go
│   ├── worker/                # unchanged
│   ├── db/                    # unchanged
│   └── copy/                  # unchanged
│
└── compress/                  # unchanged
```