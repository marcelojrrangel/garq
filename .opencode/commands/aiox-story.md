---
description: "Create, manage, or validate AIOX user stories"
agent: aiox-sm
---

Manage AIOX user stories. Usage:

## Create a new story
`/aiox-story create --from <prd|epic> --name "Story title"`

1. Read the PRD from `docs/prd/` or `docs/prd.md`
2. Read the architecture from `docs/architecture/` or `docs/architecture.md`
3. Draft story using `.aiox-core/development/tasks/sm-create-next-story.md`
4. Story goes in `docs/stories/` directory

## Validate a story
`/aiox-story validate <story-file>`

1. Read the story file
2. Run validation checklist from `.aiox-core/development/checklists/`
3. Check: title, description, acceptance criteria, tasks, status, executor, quality_gate
4. Ensure executor and quality_gate are DIFFERENT agents

## List stories
`/aiox-story list`

1. Scan `docs/stories/` for all story files
2. Show each with status (draft, approved, in-progress, review, done)

## Default story location
Check `devStoryLocation` in `.aiox-core/core-config.yaml`
