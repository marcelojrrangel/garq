---
description: "Run an AIOX development workflow (development-cycle, qa-loop, spec-pipeline, etc.)"
agent: aiox-orchestrator
---

Execute an AIOX workflow. Usage:

`/aiox-workflow <workflow-name> [--story <path>] [--epic <name>]`

## Available Workflows
- `development-cycle` — Full dev cycle: PO → Executor → Self-Healing → Quality Gate → DevOps → Checkpoint
- `qa-loop` — Quality assurance loop
- `spec-pipeline` — Specification pipeline
- `greenfield-fullstack` — Greenfield fullstack project
- `greenfield-service` — Greenfield service
- `greenfield-ui` — Greenfield UI
- `brownfield-fullstack` — Brownfield fullstack
- `brownfield-discovery` — Brownfield discovery
- `auto-worktree` — Automated worktree management
- `story-development-cycle` — Single story development cycle

## Execution Process
1. Load workflow from `.aiox-core/development/workflows/<name>.yaml`
2. Parse phases, agents, inputs/outputs, error handlers
3. Validate story file if `--story` provided
4. Execute phases sequentially:
   - Phase 1: Validation (agent from workflow definition)
   - Phase 2: Development (dynamic executor from story)
   - Phase 3: Self-Healing (conditional)
   - Phase 4: Quality Gate (different agent from executor)
   - Phase 5: Push & PR (devops only)
   - Phase 6: Checkpoint (human decision: GO/PAUSE/REVIEW/ABORT)
5. Handle errors using workflow error_handlers
6. Persist state to `.aiox/workflow-state/`

## Example
`/aiox-workflow development-cycle --story docs/stories/STORY-42.md`
