---
description: "Initialize or verify AIOX framework installation in the project"
agent: aiox-master
---

Initialize AIOX in the current project:

1. Read `.aiox-core/core-config.yaml` to check if already initialized
2. If not initialized, guide the user to run `npx aiox-core install`
3. If initialized, show current config:
   - Project type: `{project.type}`
   - Installed at: `{project.installedAt}`
   - Version: `{project.version}`
   - User profile: `{user_profile}`
   - IDE configs enabled: list of active IDEs
4. Check `.aiox-core/` directory structure is complete (core/, cli/, development/, data/)
5. Check `.agent/workflows/` has all 12 agent activation files
6. Check `.aiox/logs/` and `.aiox/sessions/` exist
7. Run health check: verify Node.js >= 18, npm >= 9
8. Show summary of findings and any recommendations
