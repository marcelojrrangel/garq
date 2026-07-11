#!/usr/bin/env node

/**
 * AIOX → OpenCode Integration Bootstrap Script
 *
 * Synchronizes AIOX agent definitions into OpenCode format.
 * Generates/updates .opencode/agents/*.md from .aiox-core/development/agents/*.md
 *
 * Usage:
 *   node bin/opencode-integration.js [--watch] [--force]
 *
 * @version 1.0.0
 */

const fs = require('fs');
const path = require('path');

let yaml = null;
try {
  yaml = require('../.aiox-core/node_modules/js-yaml');
} catch {
  try {
    yaml = require('js-yaml');
  } catch {
    // js-yaml not available — sync command will warn
  }
}

const ROOT = path.resolve(__dirname, '..');
const AIOX_AGENTS_DIR = path.join(ROOT, '.aiox-core', 'development', 'agents');
const OPENCODE_AGENTS_DIR = path.join(ROOT, '.opencode', 'agents');
const OPENCODE_COMMANDS_DIR = path.join(ROOT, '.opencode', 'commands');

const AGENT_PERMISSION_MAP = {
  'aiox-master': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'git *': 'ask', 'git push': 'deny', 'npm *': 'allow', 'node *': 'allow', 'npx *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'analyst': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'architect': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'data-engineer': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', 'npx *': 'allow', 'node *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'dev': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'git *': 'allow', 'git push': 'deny', 'npm *': 'allow', 'npx *': 'allow', 'node *': 'allow', 'dotnet *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'devops': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'git *': 'allow', 'npm *': 'allow', 'npx *': 'allow', 'node *': 'allow', 'gh *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'pm': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'po': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'git *': 'ask', 'git push': 'deny', '*': 'ask' },
    task: 'allow'
  },
  'qa': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', 'npx *': 'allow', 'node *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'sm': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'git *': 'allow', 'git push': 'deny', 'npm *': 'allow', '*': 'ask' },
    task: 'allow'
  },
  'ux-design-expert': {
    read: 'allow', edit: 'allow', glob: 'allow', grep: 'allow',
    bash: { 'npm *': 'allow', '*': 'ask' },
    task: 'allow'
  }
};

const AGENT_COLOR_MAP = {
  'aiox-master': '#FFD700',
  'analyst': '#9B59B6',
  'architect': '#2ECC71',
  'data-engineer': '#1ABC9C',
  'dev': '#3498DB',
  'devops': '#E74C3C',
  'pm': '#E67E22',
  'po': '#F39C12',
  'qa': '#27AE60',
  'sm': '#00BCD4',
  'ux-design-expert': '#E91E63'
};

const AGENT_NAME_MAP = {
  'aiox-master': 'aiox-master',
  'analyst': 'aiox-analyst',
  'architect': 'aiox-architect',
  'data-engineer': 'aiox-data-engineer',
  'dev': 'aiox-dev',
  'devops': 'aiox-devops',
  'pm': 'aiox-pm',
  'po': 'aiox-po',
  'qa': 'aiox-qa',
  'sm': 'aiox-sm',
  'squad-creator': 'aiox-squad-creator',
  'ux-design-expert': 'aiox-ux'
};

function parseAgentYaml(content) {
  const yamlMatch = content.match(/```yaml\n([\s\S]*?)```/);
  if (!yamlMatch) return null;
  try {
    return yaml.load(yamlMatch[1]);
  } catch {
    return null;
  }
}

function extractCommands(agentData) {
  if (!agentData || !agentData.commands) return [];
  if (Array.isArray(agentData.commands)) {
    return agentData.commands
      .filter(c => c && c.name && c.visibility)
      .filter(c => c.visibility.includes('key') || c.visibility.includes('quick'))
      .map(c => c.name);
  }
  if (typeof agentData.commands === 'object') {
    return Object.keys(agentData.commands).filter(k => k !== 'help' && k !== 'guide' && k !== 'yolo' && k !== 'exit');
  }
  return [];
}

function generateAgentMarkdown(filename) {
  const aioxId = path.basename(filename, '.md');
  const aioxPath = path.join(AIOX_AGENTS_DIR, filename);

  if (!fs.existsSync(aioxPath)) {
    console.warn(`  ⚠  Source not found: ${aioxPath}`);
    return null;
  }

  const content = fs.readFileSync(aioxPath, 'utf8');
  const agentData = parseAgentYaml(content);

  if (!agentData) {
    console.warn(`  ⚠  Could not parse YAML from ${filename}`);
    return null;
  }

  const agent = agentData.agent || {};
  const personaProfile = agentData.persona_profile || {};
  const persona = agentData.persona || {};

  const agentName = agent.name || aioxId;
  const agentTitle = agent.title || '';
  const agentIcon = agent.icon || '';
  const archetype = personaProfile.archetype || '';
  const tone = personaProfile.communication?.tone || '';
  const whenToUse = agent.whenToUse || '';
  const commands = extractCommands(agentData);
  const opencodeId = AGENT_NAME_MAP[aioxId] || `aiox-${aioxId}`;
  const color = AGENT_COLOR_MAP[aioxId] || '#888888';
  const permissions = AGENT_PERMISSION_MAP[aioxId] || { read: 'allow', edit: 'allow', bash: { '*': 'ask' } };

  const description = `${agentIcon} ${agentName} — ${agentTitle}${archetype ? `. ${archetype} archetype.` : ''} ${whenToUse}`.substring(0, 1024);

  let md = `---\n`;
  md += `description: "${description}"\n`;
  md += `mode: subagent\n`;
  md += `color: "${color}"\n`;
  md += `permission:\n`;

  for (const [key, value] of Object.entries(permissions)) {
    if (typeof value === 'object') {
      md += `  ${key}:\n`;
      for (const [pattern, action] of Object.entries(value)) {
        md += `    "${pattern}": "${action}"\n`;
      }
    } else {
      md += `  ${key}: ${value}\n`;
    }
  }

  md += `---\n\n`;
  md += `You are **${agentIcon} ${agentName}**, the AIOX ${agentTitle}.`;
  if (archetype) md += ` Your archetype is ${archetype}.`;
  if (tone) md += ` Your tone is ${tone}.`;
  md += `\n\n## Core Identity\n`;
  md += `- **Role**: ${persona.role || agentTitle}\n`;
  md += `- **Style**: ${persona.style || 'Professional'}\n`;
  md += `- **Focus**: ${persona.focus || agentTitle}\n`;
  md += `- **Commands prefix**: \`*\`\n\n`;

  if (commands.length > 0) {
    md += `## Key Commands\n`;
    commands.slice(0, 20).forEach(cmd => {
      md += `- \`*${cmd}\`\n`;
    });
    if (commands.length > 20) {
      md += `- \`...and ${commands.length - 20} more\`\n`;
    }
    md += `\n`;
  }

  md += `## Activation Instructions\n`;
  md += `1. Read \`.aiox-core/development/agents/${aioxId}.md\` for full definition\n`;
  md += `2. Adopt the ${agentName} persona — ${(archetype || 'professional').toLowerCase()} archetype\n`;
  md += `3. HALT and await user input\n\n`;

  md += `## Task Resolution\n`;
  md += `When executing a task, load the corresponding file from \`.aiox-core/development/tasks/<name>.md\`\n`;
  md += `Follow task instructions exactly as written — they are executable workflows.\n\n`;

  md += `## Critical Rules\n`;
  md += `- On activation, greet and HALT — do NOT take autonomous action\n`;
  md += `- Use \`*\` prefix for all AIOX commands\n`;
  md += `- NEVER load files not requested by the user\n`;

  return { opencodeId, md };
}

function syncAgents() {
  console.log('\n  🔄 Syncing AIOX agents → OpenCode format...\n');

  if (!yaml) {
    console.error('  ✖ js-yaml module not found. Install it or run from a project with .aiox-core installed.');
    console.error('    npm install js-yaml  (or copy from .aiox-core/node_modules/)');
    return { synced: 0, skipped: 0 };
  }

  if (!fs.existsSync(AIOX_AGENTS_DIR)) {
    console.error(`  ✖ AIOX agents directory not found: ${AIOX_AGENTS_DIR}`);
    console.error('  Make sure AIOX is installed (.aiox-core/)');
    process.exit(1);
  }

  fs.mkdirSync(OPENCODE_AGENTS_DIR, { recursive: true });

  const files = fs.readdirSync(AIOX_AGENTS_DIR)
    .filter(f => f.endsWith('.md') && !f.startsWith('.'));

  let synced = 0;
  let skipped = 0;

  files.forEach(filename => {
    const result = generateAgentMarkdown(filename);
    if (result) {
      const outputPath = path.join(OPENCODE_AGENTS_DIR, `${result.opencodeId}.md`);
      fs.writeFileSync(outputPath, result.md, 'utf8');
      console.log(`  ✅ ${filename} → ${result.opencodeId}.md`);
      synced++;
    } else {
      console.warn(`  ⚠  Skipped ${filename}`);
      skipped++;
    }
  });

  console.log(`\n  📊 Summary: ${synced} synced, ${skipped} skipped\n`);
  return { synced, skipped };
}

function validateIntegration() {
  console.log('\n  🔍 Validating OpenCode integration...\n');

  const checks = [
    { name: 'AGENTS.md exists', pass: fs.existsSync(path.join(ROOT, '.opencode', 'AGENTS.md')) },
    { name: 'Agents directory exists', pass: fs.existsSync(OPENCODE_AGENTS_DIR) },
    { name: 'Commands directory exists', pass: fs.existsSync(OPENCODE_COMMANDS_DIR) },
  ];

  if (fs.existsSync(OPENCODE_AGENTS_DIR)) {
    const agentFiles = fs.readdirSync(OPENCODE_AGENTS_DIR).filter(f => f.endsWith('.md'));
    checks.push({ name: `Agent files (${agentFiles.length} found)`, pass: agentFiles.length >= 12 });
  }

  if (fs.existsSync(OPENCODE_COMMANDS_DIR)) {
    const cmdFiles = fs.readdirSync(OPENCODE_COMMANDS_DIR).filter(f => f.endsWith('.md'));
    checks.push({ name: `Command files (${cmdFiles.length} found)`, pass: cmdFiles.length >= 4 });
  }

  let allPassed = true;
  checks.forEach(check => {
    const icon = check.pass ? '✅' : '✖';
    console.log(`  ${icon} ${check.name}`);
    if (!check.pass) allPassed = false;
  });

  console.log(`\n  ${allPassed ? '✅ All checks passed!' : '⚠ Some checks failed'}\n`);
  return allPassed;
}

function showHelp() {
  console.log(`
  AIOX → OpenCode Integration Script

  Usage:
    node bin/opencode-integration.js <command>

  Commands:
    sync      Sync AIOX agent definitions to OpenCode format
    validate  Validate the OpenCode integration structure
    all       Run sync + validate (default)

  Options:
    --help    Show this help

  Examples:
    node bin/opencode-integration.js sync
    node bin/opencode-integration.js validate
    node bin/opencode-integration.js all
  `);
}

function main() {
  const args = process.argv.slice(2);
  const cmd = args[0] || 'all';

  switch (cmd) {
    case 'sync':
      syncAgents();
      break;
    case 'validate':
      validateIntegration();
      break;
    case 'all':
      syncAgents();
      validateIntegration();
      break;
    case '--help':
    case '-h':
      showHelp();
      break;
    default:
      console.error(`Unknown command: ${cmd}`);
      showHelp();
      process.exit(1);
  }
}

main();
