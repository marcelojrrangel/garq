# AIOX Framework for OpenCode

This project uses the **AIOX (Artificial Intelligence Orchestration eXperience)** framework — a multi-agent system for full-stack development with 12 specialized AI agents, 215+ executable tasks, and orchestrated workflows.

## Project Structure

```
.agent/workflows/          # Agent activation instructions (12 agents)
.aiox-core/                # Core AIOX framework
  ├── core/                # Orchestration engine, quality gates, MCP, etc.
  ├── cli/                 # Commander.js CLI (aiox commands)
  ├── development/
  │   ├── agents/          # Agent definitions (YAML + persona)
  │   ├── tasks/           # 215+ executable task workflows
  │   ├── workflows/       # Orchestrated workflow definitions
  │   ├── templates/       # Document templates (PRD, ADR, stories)
  │   └── checklists/      # Quality checklists
  ├── data/                # Knowledge base, tool registry, patterns
  └── scripts/             # Activation and utility scripts
.aiox/                     # Sessions, logs, state
.opencode/                 # OpenCode integration
  ├── agents/              # AIOX agents as OpenCode subagents (@aiox-*)
  ├── commands/            # Custom /aiox-* commands
  └── skills/aiox-core/    # Core framework skill
```

## Available AIOX Agents

Use `@aiox-<name>` to invoke any agent:

| Agent | ID | Role | Icon |
|-------|-----|------|------|
| @aiox-master | Orion | Master Orchestrator & Framework Developer | 👑 |
| @aiox-orchestrator | — | Workflow Orchestrator & Team Coordinator | 🔄 |
| @aiox-analyst | Atlas | Business Analyst & Research | 🔍 |
| @aiox-pm | Morgan | Product Manager (PRD, epics, strategy) | 📋 |
| @aiox-architect | Aria | System Architect & Tech Design | 🏛️ |
| @aiox-ux | Uma | UX/UI Designer & Design System | 🎨 |
| @aiox-sm | River | Scrum Master (stories, sprints) | 🌊 |
| @aiox-dev | Dex | Full Stack Developer (implementation) | 💻 |
| @aiox-qa | Quinn | Test Architect & Quality Advisor | ✅ |
| @aiox-po | Pax | Product Owner (backlog, prioritization) | 🎯 |
| @aiox-data-engineer | Dara | Database Architect & Ops | 📊 |
| @aiox-devops | Gage | DevOps & Git Ops (exclusive push) | ⚡ |

## Custom Commands

| Command | Description |
|---------|-------------|
| `/aiox-help` | Show AIOX help and agent list |
| `/aiox-init` | Initialize AIOX in the project |
| `/aiox-story` | Create/manage user stories |
| `/aiox-workflow` | Run an AIOX workflow |
| `/loop-architect` | Loop Engineering — auto-correction cycle (roadmap → code → test → fix) |

## Key Conventions

- **Agent commands** use `*` prefix (e.g., `*help`, `*develop`, `*gate`)
- **Task files** are in `.aiox-core/development/tasks/<name>.md`
- **Workflows** are in `.aiox-core/development/workflows/<name>.yaml`
- **Core config** is in `.aiox-core/core-config.yaml`
- Tasks have an `IDE-FILE-RESOLUTION` section: map task names to `.aiox-core/development/tasks/<name>.md`
- When an agent needs to execute a task, load the corresponding `.md` file from the tasks directory
- Workflows define phases, agents, inputs/outputs, and error handlers

## Installed Skills

10 skills from [Tech Leads Club](https://agent-skills.techleads.club/skills/) in `.opencode/skills/` — load with `skill` tool:

| Skill | Usage | Agent |
|-------|-------|-------|
| `tlc-spec-driven` | 4-phase planning (Spec→Design→Tasks→Impl) | master/orchestrator |
| `security-best-practices` | OWASP/CWE per-language review | qa |
| `playwright-skill` | E2E browser automation tests | qa |
| `tactical-ddd` | Tactical DDD (aggregates, repositories) | architect |
| `figma` | Design-to-code via Figma | ux |
| `web-quality-audit` | Full web quality audit | qa |
| `aws-advisor` | AWS architecture (cost, security, perf) | devops |
| `skill-architect` | Create new skills for the framework | master |
| `codenavi` | Intelligent codebase navigation | dev |
| `sentry` | Error monitoring | devops |
| `loop-engineering` | Self-correcting cycle (roadmap→code→test→fix) | dev |

## Workflow Engine

The AIOX workflow engine lives in `.aiox-core/core/orchestration/`:
- `workflow-executor.js` — executes multi-phase workflows
- `master-orchestrator.js` — ADE master orchestration
- `terminal-spawner.js` — spawns sub-agents in clean terminals
- `session-state.js` — persists/resumes session state

Use `/aiox-workflow <name> --story <path>` to run a workflow.

## Quality Gates

AIOX has a defense-in-depth quality system:
1. Pre-commit (local — lint, typecheck)
2. Pre-push (local — story validation)
3. CI/CD (cloud — full test suite, 80% coverage minimum)

QA agent (`@aiox-qa`) is the quality gate authority and must be a DIFFERENT agent from the executor.
