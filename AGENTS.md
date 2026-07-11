# Garq — AGENTS.md

## Build (Windows)

- **CGO required** for `lxn/walk` (Win32 GUI). Set before any Go command:
  ```powershell
  $env:GOROOT="L:\sdk\go1.24.0"
  $env:Path="L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"
  $env:CC="L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"
  $env:CGO_ENABLED='1'
  ```
- Two entrypoints:
  - `cmd/walk/` — Win32 desktop GUI (lxn/walk). Build: `go build -o L:\sdk\garq-new.exe ./cmd/walk/`
  - `cmd/app/` — HTTP API server. Build: `go build ./cmd/app/`
- Scripts: `scripts/build-run.ps1` / `.bat`, `scripts/test-all.ps1` / `.bat`
- **Icons regenerated at build time** by `tools/genicons/` → `assets/icons/*.ico`. Run `go run ./tools/genicons/` before build if icons missing.
- **Resources:** `cmd/walk/resources.rc` compiled by `windres` → `cmd/walk/resources.syso`, linked by Go linker. Contains manifest + toolbar icons. Run `windres -i cmd/walk/resources.rc -o cmd/walk/resources.syso` if `resources.syso` is stale.

## Tests

```powershell
go test -v ./compress/... ./internal/...
```

Targets: `compress/`, `internal/api/`, `internal/copy/`, `internal/db/`. No integration dependencies.

## Logging

`init()` in `cmd/walk/main.go` redirects `log` to `%TEMP%\garq.log`. Check there for runtime debugging.

## Architecture

- **`compress/`** — 7z CLI (`exec.Command`) with progress streaming via `-bsp1`, context cancellation
- **`internal/db/`** — SQLite (pure-Go via `modernc.org/sqlite`, no CGO needed for DB)
- **`internal/worker/`** — Worker pool (4 goroutines) pulling jobs from DB; supports pause/cancel via context + channels
- **`internal/api/`** — Two layers: `wails.go` (UI bindings), `filesystem.go` (directory listing), `http.go` (REST handlers)
- **`internal/copy/`** — Platform-specific copy: `CopyFileW` on Windows, buffered fallback
- **`cmd/walk/`** — All UI in one file (`main.go`, ~1465 loc) plus dialog helpers: `compress_dialog.go`, `extract_dialog.go`, `folder_dialog.go`, `progress_dialog.go`

## Walk (GUI) Quirks

- Uses `declarative` package (builder pattern). Widgets are declared struct-style, then created via `Create()`.
- `declarative.PushButton` has **no `ToolTip` field**. Set tooltips after `Create()` via `AssignTo` + `b.SetToolTipText(...)`.
- `MinSize: Size{28, 26}` for toolbar buttons. Button text is short (2-3 chars).
- Tree uses `SHGetFileInfoW` for icons per node.
- Nav tree has section headers (disabled nodes) + real folders.
- Preview pane is a `Composite` toggled via `SetVisible()`.
- `init()` manifest: `cmd/walk/manifest.xml` with DPI awareness `PerMonitorV2`.
- Tab pages use `Pages().Remove()` for removal (not `Dispose()` directly).

## DB Schema

Single table `jobs` (id, type, payload, status, progress, error). WAL journal mode.

## 7z CLI

Searches in order: `GARQ_7Z_PATH` env → bundled `<app>/7z/7z.exe` → `%PATH%` → standard install paths. No CGO SDK binding.
