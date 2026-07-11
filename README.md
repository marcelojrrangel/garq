# AIOX Core — Fork com Integração Nativa para OpenCode CLI

Fork do [SynkraAI/aiox-core](https://github.com/SynkraAI/aiox-core) com camada de integração nativa para o [OpenCode CLI](https://opencode.ai).

**AIOX** (Artificial Intelligence Orchestration eXperience) é um framework de orquestração de IA para desenvolvimento full-stack com **12 agentes especializados**, **215+ tarefas executáveis** e **workflows orquestrados**.

## Início Rápido

```bash
git clone https://github.com/marcelojrrangel/aiox-core.git meu-projeto
cd meu-projeto

# Instalar dependência do plugin OpenCode
cd .opencode && npm install
cd ..

# Abrir OpenCode
opencode

# Verificar integração
/aiox-init
```

## Estrutura

| Caminho | Papel |
|---------|-------|
| `.aiox-core/` | Framework core (orquestração, 215+ tasks, workflows) |
| `.opencode/` | Integração OpenCode (13 subagentes, 5 comandos, 11 skills) |
| `.agent/workflows/` | Instruções de ativação dos agentes |
| `opencode.json` | Config global — permissões, subagentes, comandos |
| `docs/` | Guia de uso completo |
| `bin/opencode-integration.js` | Script de sincronização de agentes |

## Comandos

| Comando | Descrição |
|---------|-----------|
| `/aiox-help` | Ajuda geral do ecossistema |
| `/aiox-init` | Verificar instalação |
| `/aiox-story` | Gerenciar histórias |
| `/aiox-workflow` | Executar workflow |
| `/loop-architect` | Loop Engineering (código → teste → correção) |

## Skills Instaladas

11 skills em `.opencode/skills/` — carregue com `skill` tool. Ver `AGENTS.md` para lista completa.

## Pré-requisitos

- Node.js >= 18, npm >= 9
- OpenCode CLI (`npm install -g opencode-ai`)
- Git

## Licença

MIT
