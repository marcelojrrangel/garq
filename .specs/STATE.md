# State

## Decisions

| ID | Decision | Rationale | Date |
|----|----------|-----------|------|
| AD-01 | `internal/service` wraps existing `api.API`, not replaces it | `cmd/app` HTTP server keeps working; `api.API` becomes the REST adapter's data layer; service is the logic layer | 2026-07-11 |
| AD-02 | Lightweight ops (rename, mkdir) run synchronously via service, not via worker queue | These are instant `os.*` calls; queue adds overhead with no UX benefit; the service abstracts the call so it can be swapped to async later if needed | 2026-07-11 |
| AD-03 | Delete-permanent routes through `api.AddDeleteJob` (worker queue) | Unifies with worker model; gives progress dialog; no UI freeze; existing `runDeleteJob` handler already works | 2026-07-11 |
| AD-04 | Progress dialog keeps DB polling but through `service.JobSnapshot(id)` instead of raw SQL | Push channel (`ObserveJob(id) <-chan Snapshot`) is a larger effort; polling-through-service still removes raw SQL from UI and is the 80/20 win | 2026-07-11 |
| AD-05 | `main.go` split is a pure mechanical move (no behavior changes) | Minimizes risk; each file gets `package main`; the split happens after service migration so file ops are already thin | 2026-07-11 |
| AD-06 | Clipboard protocol becomes a typed struct in service, not a string format | `"CUT:\n" + paths` is error-prone; `ClipboardOp{Op: "cut", Paths: []string}` is typed and validated | 2026-07-11 |
| AD-07 | UI-only changes (sort indicators, empty state, disabled states, focus) gate on build+vet only | Go GUI code is untestable without a running app instance; the service layer carries the unit-testable logic | 2026-07-11 |
| AD-08 | `internal/shell` package extracted from `cmd/walk/recycle_windows.go` | Avoids import cycle: `internal/service` cannot import `cmd/walk` (package main); makes recycle logic reusable by `cmd/app` | 2026-07-11 |
| AD-09 | `service.InitDB`, `service.StartWorkerPool`, `service.NewFromDB` helpers added | Allows `cmd/walk` to bootstrap DB and workers without importing `db`, `worker`, `compress`, or `copy` directly | 2026-07-11 |

## Handoff

**Feature**: front-end-melhorias
**Branch**: `feature/front-end-melhorias`
**Status**: Phase 2 complete — service layer implemented; ready for Phase 3
**Phase / Task**: Phase 3, T11 (next)
**Completed**: T1–T10 (Phase 1 + Phase 2)
**In progress**: None
**Next step**: Dispatch Worker 3 for Phase 3 (T11–T17: UI split + UX polish)
**Blockers**: None
**Uncommitted files**: `.specs/STATE.md` update
**Commit range**: `9ff4875..11b2674`