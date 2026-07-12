# Front-end Melhorias Specification

## Problem Statement

The `cmd/walk` GUI layer is a leaky abstraction sink: it imports 5 backend packages
directly, runs raw SQL queries, duplicates business logic (3 copy implementations,
inline `os.RemoveAll`/`os.Rename`/`os.MkdirAll`), and contains a latent double-enqueue
bug in `pasteClipboard`. `main.go` is 1518 LOC doing everything from layout to file ops.
Several P0 safety/correctness issues (no `Shift+Del` confirmation, drag-drop freezes UI,
properties dialog hangs on large directories) coexist with UX gaps (no sort indicators,
no empty states, toolbar buttons never disable, status bar clobbers messages).

## Goals

- [ ] Eliminate all P0 correctness and safety issues (double-enqueue, missing
  confirmation, UI thread freezes)
- [ ] Create `internal/service` as the single boundary between UI and backend —
  `cmd/walk` imports shrink from 5 backend packages to 1 (`internal/service`)
- [ ] Remove `database/sql` and raw SQL from `cmd/walk` entirely
- [ ] Split `main.go` (1518 LOC) into focused files (~180 LOC each)
- [ ] Add sort indicators, empty states, context-sensitive disabled toolbar buttons,
  and focus management on tab create/switch

## Out of Scope

| Feature | Reason |
| ----------- | -------------- |
| Breadcrumb composite | Larger effort; separate feature after this foundation lands |
| `IFileDialog` replacing `SHBrowseForFolder` | COM interop scaffolding; separate feature |
| High-DPI icon fetching (`SHGetImageList`) | Visual polish; separate feature |
| In-place rename via `LineEdit` overlay | Complex walk hack; separate feature |
| Transfers queue panel | Requires job observation push channel (P2 of this spec); defer panel UI |
| Undo framework | Separate feature; needs job history schema changes |
| Tab right-click context menu | walk hit-test complexity; separate feature |
| Preferences persistence (DB-backed) | Separate feature; needs schema for user prefs |
| Compress dialog format/level/password | Depends on 7z CLI capabilities; separate feature |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --------------------- | -------------- | --------- | ---------- |
| `internal/api` stays as REST adapter; `internal/service` wraps it | service wraps api.API | `cmd/app` HTTP entrypoint still works; no double-wiring | n |
| Lightweight ops (rename, mkdir) run synchronously, not via worker queue | sync through service | These are instant `os.*` calls; worker queue adds overhead with no UX benefit | n |
| Delete-permanent uses existing `api.AddDeleteJob` (worker queue) | via service.DeletePermanent | Unifies with worker model; progress dialog; no UI freeze | n |
| `ObserveJob(id)` push channel deferred to a follow-up feature | keep polling, but through service.JobSnapshot | Push channel is a larger effort; polling-through-service still removes raw SQL | n |
| `closeCurrentTab` on last tab: disable button, not close app | disable btnCloseTab | Matches Explorer behavior (always 1 tab); user explicitly said "feels broken" | n |
| ProgressDialog 10s auto-close, not "never" | 10s timeout | User who tabs away for <10s still sees result; >10s means they're done waiting | n |
| Tests target: behavioral unit tests for service layer; build+vet gate for UI-only | Go table tests | GUI code untestable without app instance; service is pure logic | n |

**Open questions:** none — all resolved or logged above.

---

## User Stories

### P1: Fix pasteClipboard double-enqueue bug ⭐ MVP

**User Story**: As a user, I want cut-paste to enqueue exactly one job (move, not
copy+move) so that files are moved correctly and no orphan jobs pollute the queue.

**Why P1**: Data correctness bug. `main.go:1229-1232` enqueues `AddCopyJob` THEN
overwrites with `AddMoveJob` when `isCut=true`, leaving an orphaned copy job in the
DB that a worker will execute.

**Acceptance Criteria**:

1. WHEN user cuts (`Ctrl+X`) then pastes (`Ctrl+V`) THEN system SHALL enqueue exactly
   one `move` job and zero `copy` jobs
2. WHEN user copies (`Ctrl+C`) then pastes THEN system SHALL enqueue exactly one
   `copy` job
3. WHEN paste is invoked with empty clipboard or invalid format THEN system SHALL show
   status message and enqueue nothing
4. WHEN cut-paste completes THEN clipboard SHALL be cleared

**Independent Test**: Cut a file, paste in another folder, verify DB has exactly one
`move` job (no `copy` job) via `SELECT count(*) FROM jobs WHERE type='copy'` returning
0 for that operation.

---

### P1: Add Shift+Del confirmation dialog ⭐ MVP

**User Story**: As a user, I want a confirmation dialog before permanent deletion
(`Shift+Del`) so that I don't irreversibly destroy data with a single keystroke.

**Why P1**: Safety hole. `deleteSelectedPermanently` at `main.go:1293-1310` calls
`os.RemoveAll` inline with zero confirmation.

**Acceptance Criteria**:

1. WHEN user presses `Shift+Del` with files selected THEN system SHALL show a
   `walk.MsgBox` confirmation with Yes/No buttons, default = No
2. WHEN user clicks "No" or presses Esc THEN system SHALL delete nothing
3. WHEN user clicks "Yes" THEN system SHALL enqueue a `delete` job via
   `api.AddDeleteJob` (not inline `os.RemoveAll`) and show progress
4. WHEN no files are selected THEN system SHALL show status message and not show dialog

**Independent Test**: Select a file, press `Shift+Del`, click "No" — verify file still
exists on disk. Repeat, click "Yes" — verify file is deleted and a `delete` job appears
in DB.

---

### P1: Async countDirContents in properties dialog ⭐ MVP

**User Story**: As a user, I want the properties dialog to open immediately and count
directory contents in the background so that the UI doesn't freeze on large folders.

**Why P1**: `main.go:1363-1365` calls `countDirContents` synchronously in
`showProperties`, blocking the UI thread for potentially minutes on a 50k-file
directory.

**Acceptance Criteria**:

1. WHEN user opens properties for a directory THEN dialog SHALL appear immediately
   with "Contando..." placeholder where the content count would be
2. WHEN counting completes THEN system SHALL replace "Contando..." with
   `"Conteúdo: %d arquivo(s), %d pasta(s)"` via `Synchronize`
3. WHEN properties dialog is closed before count completes THEN goroutine SHALL
   exit without panicking or writing to a disposed widget
4. WHEN properties target is a file (not dir) THEN system SHALL show size immediately,
   no "Contando..." placeholder

**Independent Test**: Right-click a large directory (1000+ files), select Properties —
dialog appears instantly, "Contando..." text visible, then replaces with counts within
seconds.

---

### P1: Drag-drop ingest via worker job ⭐ MVP

**User Story**: As a user, I want drag-dropped files to be copied via the worker queue
with a progress dialog so that large drops don't freeze the UI.

**Why P1**: `main.go:361-393` copies dropped files inline with `copyPath` (synchronous
`filepath.Walk` + `os.ReadFile/WriteFile`), blocking the UI thread with no progress.

**Acceptance Criteria**:

1. WHEN user drops files onto the file table THEN system SHALL enqueue a `copy` job
   via `api.AddCopyJob` and open a progress dialog
2. WHEN drop involves a large directory THEN UI SHALL remain responsive (navigation,
   toolbar, other tabs all functional during copy)
3. WHEN drop is empty or invalid THEN system SHALL show status message and enqueue
   nothing
4. WHEN copy job completes THEN active tab SHALL refresh automatically

**Independent Test**: Drag a 100MB+ folder onto the app — progress dialog appears, UI
stays responsive (can switch tabs during copy), files appear in destination after
completion.

---

### P2: Create internal/service layer as single UI/backend boundary

**User Story**: As a developer, I want a single `internal/service` package that the
UI imports so that the UI has no direct dependency on `db`, `worker`, `compress`, or
`copy`, and all business logic lives behind a clean API.

**Why P2**: Architectural foundation for all future work. Currently `cmd/walk` imports
5 backend packages, reaches raw `*sql.DB`, and duplicates worker logic.

**Acceptance Criteria**:

1. WHEN `cmd/walk` is built THEN it SHALL import only `garq/internal/service` from the
   backend (zero imports of `db`, `worker`, `compress`, `copy`)
2. WHEN `cmd/walk/progress_dialog.go` is compiled THEN it SHALL NOT import
   `database/sql`
3. WHEN the UI needs active job count THEN it SHALL call `service.ActiveJobCount()`
   (no raw `SELECT COUNT(*)` SQL)
4. WHEN the UI needs a job snapshot THEN it SHALL call `service.JobSnapshot(id)`
   (no raw `SELECT ... FROM jobs` SQL)
5. WHEN the UI pauses/resumes/cancels a job THEN it SHALL call `service.PauseJob(id)`,
   `service.ResumeJob(id)`, `service.CancelJob(id)` (not `worker.PauseJob` etc.)
6. WHEN `internal/service` is imported by `cmd/walk` THEN no `*sql.DB` shall be
   passed to any UI struct; the service owns the DB handle
7. WHEN `cmd/app` (HTTP server) is built THEN it SHALL still compile and all existing
   tests pass

**Independent Test**: `grep -r "garq/internal/db" cmd/walk/` returns zero results.
`grep -r "database/sql" cmd/walk/` returns zero results.
`go build ./cmd/walk/` and `go build ./cmd/app/` both succeed.

---

### P2: Migrate inline file ops to service layer

**User Story**: As a developer, I want all file operations (rename, mkdir, delete,
copy-dedup) to go through the service layer so there are no inline `os.*` calls in
the UI and no duplicated logic.

**Why P2**: Eliminates 3 duplicated copy implementations, removes inline
`os.RemoveAll`/`os.Rename`/`os.MkdirAll`, and unifies clipboard protocol.

**Acceptance Criteria**:

1. WHEN user creates a new folder THEN UI SHALL call `service.CreateFolder(parent,
   name)` (not inline `os.MkdirAll`)
2. WHEN user renames a file THEN UI SHALL call `service.Rename(oldPath, newName)`
   (not inline `os.Rename`)
3. WHEN user drops files THEN UI SHALL call `service.Copy(sources, dest, conflict)`
   (not inline `copyPath`/`getCopyPath`)
4. WHEN cut/copy is invoked THEN UI SHALL call `service.SetClipboard(op, paths)`
   (not raw `"CUT:\n" + strings.Join`)
5. WHEN paste is invoked THEN UI SHALL call `service.Paste(dest)` which returns
   `(jobID, jobType, error)` (not parse clipboard text inline)
6. WHEN `copyPath` or `getCopyPath` functions are searched in `cmd/walk/` THEN zero
   results shall be found (migrated to service)

**Independent Test**: `grep -r "os.RemoveAll\|os.MkdirAll\|os.Rename" cmd/walk/`
returns zero results (excluding `recycle_windows.go` which uses Win32 API, not `os.*`).
`grep -r "copyPath\|getCopyPath" cmd/walk/` returns zero results.

---

### P2: Split main.go into focused files

**User Story**: As a developer, I want `main.go` split into thematic files so that
each file is ~200 LOC and has a single responsibility, making the codebase
navigable and merge-conflict-free.

**Why P2**: `main.go` at 1518 LOC is unmaintainable — types, layout, navigation, file
ops, dialogs, and utilities are all in one file.

**Acceptance Criteria**:

1. WHEN the split is complete THEN `main.go` SHALL be ≤ 200 LOC (composition root +
   window layout only)
2. WHEN a developer searches for navigation logic THEN it SHALL be in `navigation.go`
3. WHEN a developer searches for file operations THEN they SHALL be in `fileops.go`
4. WHEN a developer searches for type declarations THEN they SHALL be in `types.go`
5. WHEN `go build ./cmd/walk/` is run THEN it SHALL succeed with zero errors
6. WHEN `go vet ./cmd/walk/` is run THEN it SHALL report zero warnings
7. WHEN existing tests are run THEN all SHALL pass (no behavior changes)

**Independent Test**: `go build ./cmd/walk/` succeeds. `wc -l cmd/walk/main.go` shows
≤ 200 lines. Each new file < 400 lines.

---

### P3: Sort indicators on column headers

**User Story**: As a user, I want to see which column is sorted and in which direction
so that I can understand the current sort state of the file list.

**Why P3**: Universal table expectation. Users can't tell sort state
(`main.go:657-665` toggles sort but header text doesn't update).

**Acceptance Criteria**:

1. WHEN user clicks the "Nome" column THEN column header SHALL show "Nome ▲" (ascending)
   or "Nome ▼" (descending)
2. WHEN sort direction is ascending THEN header SHALL display ▲
3. WHEN sort direction is descending THEN header SHALL display ▼
4. WHEN user clicks a different column THEN previous column header SHALL lose its
   indicator and the new column SHALL gain it
5. WHEN application starts THEN default sort column ("Nome", ascending) SHALL show ▲

**Independent Test**: Click "Nome" column header — text changes to "Nome ▲". Click again
— changes to "Nome ▼". Click "Tamanho" — "Nome" loses arrow, "Tamanho" gains it.

---

### P3: Empty state in file table

**User Story**: As a user, I want to see a message when a directory is empty or a
filter returns no results so that I don't think the app is broken.

**Why P3**: `main.go:858-867` shows a blank table with just headers — users assume
crash/error.

**Acceptance Criteria**:

1. WHEN directory listing returns zero items THEN status bar SHALL show "Pasta vazia"
2. WHEN filter search returns zero results THEN status bar SHALL show
   `"Nenhum item corresponde a '<filter>'"`
3. WHEN a non-empty directory is loaded THEN status bar SHALL show the normal item count
4. WHEN filter is cleared and directory has files THEN status bar SHALL revert to
   item count

**Independent Test**: Navigate to an empty folder — status bar shows "Pasta vazia".
Type a non-matching filter — status bar shows "Nenhum item corresponde a '...'".

---

### P3: Context-sensitive disabled toolbar buttons

**User Story**: As a user, I want toolbar buttons to be disabled when their action
can't be performed so that I get immediate visual feedback about what's available.

**Why P3**: Discovery-by-failure UX. Paste is always enabled even with empty clipboard;
Delete/Compress/Extract always enabled even with no selection (`main.go:501-518`).

**Acceptance Criteria**:

1. WHEN clipboard has no Garq content THEN btnPaste SHALL be disabled
2. WHEN no files are selected THEN btnRename SHALL be disabled (requires exactly 1)
3. WHEN no files are selected THEN btnCut, btnCopy, btnDelete SHALL be disabled
4. WHEN no files are selected THEN btnCompress SHALL be disabled
5. WHEN exactly 1 non-archive file is selected THEN btnExtract SHALL be disabled
6. WHEN selection changes THEN all button enabled-states SHALL update within the same
   event handler

**Independent Test**: Open app, select nothing — Cut/Copy/Rename/Delete/Compress/
Extract all grayed out. Select a file — they become enabled. Deselect — grayed out
again.

---

### P3: ProgressDialog and tab management UX fixes

**User Story**: As a user, I want the progress dialog to stay open long enough to
see results, and tab switching to manage focus correctly.

**Why P3**: ProgressDialog auto-closes in 3s (user loses feedback); new tabs don't
focus the file list; closing the last tab silently does nothing.

**Acceptance Criteria**:

1. WHEN a job completes THEN ProgressDialog SHALL auto-close after 10 seconds (not 3)
2. WHEN a job fails THEN ProgressDialog SHALL NOT auto-close (stays until user clicks
   "Fechar")
3. WHEN a new tab is created THEN file list SHALL receive focus
4. WHEN user switches tabs THEN file list of the new active tab SHALL receive focus
5. WHEN only 1 tab remains THEN btnCloseTab SHALL be disabled
6. WHEN user presses Ctrl+W on the last tab THEN nothing SHALL happen (tab stays)

**Independent Test**: Start a long copy job, let it complete, tab away for 5 seconds,
tab back — dialog still visible. Create a new tab with `Ctrl+T` — file list has
keyboard focus (arrow keys navigate immediately). With 1 tab, `Ctrl+W` does nothing.

---

## Edge Cases

- WHEN paste is invoked with a path that no longer exists (source deleted after cut)
  THEN system SHALL show error in status and skip the missing source
- WHEN properties dialog goroutine writes to a closed dialog THEN no panic (guard with
  `if pd.dialog == nil` or channel-based cancellation)
- WHEN drag-drop drops onto the same directory as the source THEN system SHALL detect
  no-op and show status "Origem e destino iguais" without enqueuing a job
- WHEN all toolbar icons fail to load THEN buttons SHALL show text fallback (existing
  pattern at `main.go:326`)
- WHEN service method returns an error THEN UI SHALL show error message in status bar
  (not swallow silently)

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| -------------- | ----- | ----- | ------ |
| FX-01 | P1: Fix pasteClipboard double-enqueue | Phase 1 | Pending |
| FX-02 | P1: Add Shift+Del confirmation | Phase 1 | Pending |
| FX-03 | P1: Async countDirContents | Phase 1 | Pending |
| FX-04 | P1: Drag-drop via worker job | Phase 1 | Pending |
| FX-05 | P2: Create internal/service layer | Phase 2 | Pending |
| FX-06 | P2: Migrate inline file ops to service | Phase 2 | Pending |
| FX-07 | P2: Split main.go into focused files | Phase 2 | Pending |
| FX-08 | P3: Sort indicators on column headers | Phase 3 | Pending |
| FX-09 | P3: Empty state in file table | Phase 3 | Pending |
| FX-10 | P3: Context-sensitive disabled toolbar | Phase 3 | Pending |
| FX-11 | P3: ProgressDialog + tab management UX | Phase 3 | Pending |

**ID format:** `FX-NN` (frontend improvement)

**Status values:** Pending → In Design → In Tasks → Implementing → Verified

**Coverage:** 11 total, 11 mapped to stories, 0 unmapped ✅

---

## Success Criteria

- [ ] Zero `database/sql` imports in `cmd/walk/` (`grep -r "database/sql" cmd/walk/`
  returns nothing)
- [ ] Zero backend package imports in `cmd/walk/` except `internal/service`
- [ ] `main.go` ≤ 200 LOC
- [ ] `go build ./cmd/walk/ && go build ./cmd/app/` both succeed
- [ ] `go vet ./cmd/walk/` reports zero warnings
- [ ] All existing tests pass (`compress/`, `internal/api/`, `internal/copy/`,
  `internal/db/`)
- [ ] `pasteClipboard` with cut enqueues exactly 1 move job (0 copy jobs)
- [ ] `Shift+Del` shows confirmation dialog
- [ ] Properties dialog opens instantly for large directories
- [ ] Sort indicators visible on column headers
- [ ] Empty directories show "Pasta vazia" in status bar
- [ ] Toolbar buttons disabled when no selection