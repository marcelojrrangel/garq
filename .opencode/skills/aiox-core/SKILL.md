---
name: aiox-core
description: "Core AIOX framework — orchestration engine, agent system, workflow definitions, and task library. Load when you need to understand the AIOX framework internals."
license: MIT
compatibility: opencode
metadata:
  version: 5.2.9
  source: https://github.com/SynkraAI/aiox-core
---

## What this skill provides

This skill loads the core AIOX framework definition so the agent can work with the full multi-agent system.

## Key files to load on-demand

### Framework Entry Point
- `.aiox-core/index.js` — Core module exports (MetaAgent, TaskManager, ElicitationEngine, TemplateEngine, etc.)
- `.aiox-core/core-config.yaml` — Project configuration (paths, features, agent settings)

### Orchestration Engine
- `.aiox-core/core/orchestration/index.js` — All orchestration exports
- `.aiox-core/core/orchestration/workflow-orchestrator.js` — Multi-phase workflow execution
- `.aiox-core/core/orchestration/master-orchestrator.js` — ADE Master Orchestrator
- `.aiox-core/core/orchestration/workflow-executor.js` — Development cycle execution
- `.aiox-core/core/orchestration/executor-assignment.js` — Dynamic executor assignment
- `.aiox-core/core/orchestration/terminal-spawner.js` — Sub-agent terminal spawning
- `.aiox-core/core/orchestration/session-state.js` — Session persistence and recovery
- `.aiox-core/core/orchestration/subagent-prompt-builder.js` — Sub-agent prompt construction
- `.aiox-core/core/orchestration/context-manager.js` — Workflow context management
- `.aiox-core/core/orchestration/recovery-handler.js` — Self-healing recovery system
- `.aiox-core/core/orchestration/gate-evaluator.js` — Quality gate evaluation

### Agent System
- `.aiox-core/development/agents/` — All 12 agent definitions (YAML + persona)
- `.agent/workflows/` — Agent activation instructions

### Task Library (215+ tasks)
- `.aiox-core/development/tasks/` — Executable task workflows

### Workflow Definitions (15 workflows)
- `.aiox-core/development/workflows/` — YAML workflow definitions

### Quality System
- `.aiox-core/core/quality-gates/` — Quality gate implementation
- `.aiox-core/development/checklists/` — Validation checklists

### Data & Configuration
- `.aiox-core/data/` — Knowledge base, tool registry, workflow chains, patterns

## When to use
- When an agent needs to understand the AIOX framework architecture
- When executing workflows that require orchestration
- When a task references core modules
- When troubleshooting framework issues

## Loading strategy
Do NOT load all files at once. Load specific files based on the user's request:
- For workflows → load the specific workflow YAML
- For agents → load the specific agent definition
- For tasks → load the specific task file
- For orchestration → load the relevant orchestrator module
