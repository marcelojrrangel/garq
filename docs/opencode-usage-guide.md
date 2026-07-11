# Guia de Uso: AIOX no OpenCode CLI

Guia completo para usar o framework AIOX nativamente dentro do OpenCode CLI com agentes especializados, comandos custom e workflows orquestrados.

---

## Índice

1. [Visão Geral](#1-visão-geral)
2. [Pré-requisitos](#2-pré-requisitos)
3. [Quick Start](#3-quick-start)
4. [Agentes AIOX (@-mentions)](#4-agentes-aiox--mentions)
5. [Comandos Custom (/aiox-*)](#5-comandos-custom-aiox-)
6. [Fluxos de Desenvolvimento Real](#6-fluxos-de-desenvolvimento-real)
7. [Workflows Orchestrated](#7-workflows-orchestrated)
8. [Dicas e Boas Práticas](#8-dicas-e-boas-práticas)
9. [Referência Rápida](#9-referência-rápida)

---

## 1. Visão Geral

O AIOX fornece 12 agentes especializados de IA que podem ser invocados diretamente no OpenCode via `@`-mention, mais **5 comandos custom** via `/` e **11 skills especializadas** do Tech Leads Club. Cada agente tem uma persona, tom e conjunto de comandos específicos — extraídos das definições em `.aiox-core/development/agents/`.

### O que você pode fazer

- **`@aiox-dev`** — implementar funcionalidades, debugar, refatorar, rodar builds autônomos
- **`@aiox-qa`** — revisar código, executar portões de qualidade, desenhar testes
- **`@aiox-architect`** — desenhar arquitetura, escolher pilha tecnológica, criar planos
- **`@aiox-pm`** — criar PRDs, gerenciar épicos, priorizar funcionalidades
- **`@aiox-sm`** — criar histórias de usuário, validar histórias, gerenciar sprints
- **`@aiox-po`** — gerenciar backlog, validar rascunhos, priorizar
- **`@aiox-analyst`** — pesquisar mercado, analisar concorrência, fazer brainstorming
- **`@aiox-devops`** — CI/CD, push, GitHub, MCP, worktrees (único autorizado a dar push)
- **`@aiox-data-engineer`** — banco de dados, esquemas, migrações, RLS
- **`@aiox-ux`** — wireframes, design system, tokens, acessibilidade
- **`@aiox-master`** — orquestração geral, criar componentes do framework
- **`@aiox-orchestrator`** — executar workflows multi-agente

---

## 2. Pré-requisitos

- OpenCode CLI instalado (`npm install -g opencode-ai` ou `curl -fsSL https://opencode.ai/install | bash`)
- Git instalado e configurado
- Acesso ao repositório do fork AIOX + OpenCode

### Obter o projeto

```bash
# Clone o fork com a integração OpenCode já incluída
git clone https://github.com/marcelojrrangel/aiox-core.git meu-projeto
cd meu-projeto

# Instale as dependências do AIOX
npm install
```

> **Nota:** Este fork já inclui toda a camada de integração (`.opencode/`, `opencode.json`, `bin/`). Diferente do `npx aiox-core init` que baixa o AIOX original sem as modificações.

### Verificar instalação

```bash
# No diretório do projeto, rode opencode
opencode

# Dentro do opencode, verifique a integração
/aiox-init
```

Se o `/aiox-init` mostrar todos os componentes verdes, a integração está pronta.

---

## 3. Quick Start

### 3.1 Obter o projeto e iniciar uma sessão

```bash
# Clone o fork (com AIOX + integração OpenCode)
git clone https://github.com/marcelojrrangel/aiox-core.git meu-projeto
cd meu-projeto

# Instale as dependências
npm install

# Inicie o OpenCode
opencode
```

### 3.2 Verificar o ecossistema AIOX

```
/aiox-help
```

Isso lista todos os 12 agentes, comandos disponíveis e localização dos recursos.

### 3.3 Chamar seu primeiro agente

```
@aiox-dev qual é o status atual do projeto?
```

O agente Dex (Full Stack Developer) será ativado, fará a saudação e aguardará suas instruções.

### 3.4 Executar um comando do agente

Após ativar o `@aiox-dev`:

```
*help
```

Mostra todos os comandos disponíveis para o agente dev (develop, run-tests, build, etc.).

### 3.5 Sair do modo agente

```
*exit
```

Retorna ao modo normal do OpenCode.

---

## 4. Agentes AIOX (@-mentions)

### 4.1 @aiox-dev — Dex, o Desenvolvedor Full Stack 💻

**Quando usar:** Implementação de código, debugging, refatoração, desenvolvimento de histórias.

```
@aiox-dev implemente o fluxo de autenticação de usuário da HISTORIA-42

@aiox-dev debugue a página de login — está retornando 500 ao enviar

@aiox-dev refatore o serviço de pagamento para usar o padrão strategy

@aiox-dev execute os testes

@aiox-dev build HISTORIA-45 --mode yolo

@aiox-dev *gotchas liste os padrões recentes
```

### 4.2 @aiox-qa — Quinn, o Arquiteto de Testes ✅

**Quando usar:** Revisão de código, portões de qualidade, design de testes.

```
@aiox-qa revise a implementação atual da HISTORIA-42

@aiox-qa gate execute o portão de qualidade no módulo de pagamentos

@aiox-qa crie uma estratégia de testes para o fluxo de checkout

@aiox-qa crie-pedido-correcao --story HISTORIA-42

@aiox-qa verificacao-seguranca audite o código de autenticação

@aiox-qa rastreie mapeie os requisitos da HISTORIA-42 para os testes
```

### 4.3 @aiox-architect — Aria, a Arquiteta 🏛️

**Quando usar:** Design de arquitetura, pilha tecnológica, design de API, planos de implementação.

```
@aiox-architect crie-arquitetura-fullstack para um sistema de cobrança SaaS

@aiox-architect avalie-complexidade para o épico de integração de pagamentos

@aiox-architect crie-plano para implementação da HISTORIA-42

@aiox-architect mapeie-codebase documente a estrutura atual do projeto

@aiox-architect pesquise compare Next.js vs Remix para este projeto

@aiox-architect valide-preset-tecnologico --stack typescript-react-node
```

### 4.4 @aiox-pm — Morgan, o Gerente de Produto 📋

**Quando usar:** Criação de PRD, épicos, estratégia de produto.

```
@aiox-pm crie-prd para uma plataforma SaaS multi-inquilino

@aiox-pm crie-epico --name "Integração de Gateway de Pagamento"

@aiox-pm levante-requisitos para o módulo de relatórios

@aiox-pm escreva-spec para o serviço de notificações

@aiox-pm crie-prd-brownfield para a migração do CRM legado
```

### 4.5 @aiox-sm — River, o Scrum Master 🌊

**Quando usar:** Criação de histórias, validação, planejamento de sprint.

```
@aiox-sm crie história a partir do PRD para a funcionalidade de login

@aiox-sm validacao-historia valide HISTORIA-42

@aiox-sm crie a próxima história para o fluxo de redefinição de senha
```

### 4.6 @aiox-po — Pax, o Dono do Produto 🎯

**Quando usar:** Gerenciamento de backlog, priorização, validação de histórias.

```
@aiox-po backlog-adicionar implementar login social OAuth2 --prioridade alta

@aiox-po backlog-priorizar --metodo rice

@aiox-po validar-rascunho-historia HISTORIA-42

@aiox-po fechar-historia HISTORIA-42

@aiox-po indice-historias mostre todas as histórias agrupadas por status
```

### 4.7 @aiox-analyst — Atlas, o Analista 🔍

**Quando usar:** Pesquisa, análise, brainstorming, discovery.

```
@aiox-analyst realize-pesquisa-mercado para assistentes de codificação com IA

@aiox-analyst crie-analise-concorrencia para nossa categoria de produto

@aiox-analyst sessa brainstorming para funcionalidades do Q2

@aiox-analyst pesquise-deps avalie auth0 vs clerk vs supabase auth

@aiox-analyst extraia-padroes da base de código atual
```

### 4.8 @aiox-devops — Gage, o DevOps ⚡

**Quando usar:** Operações git, CI/CD, push, MCP, worktrees. **Único autorizado a dar push.**

```
@aiox-devops push --branch feature/HISTORIA-42

@aiox-devops criar-pr --story HISTORIA-42

@aiox-devops configurar-ci configure GitHub Actions

@aiox-devops criar-worktree HISTORIA-45

@aiox-devops verificacao-saude

@aiox-devops adicionar-mcp --server playwright --url http://localhost:3000
```

### 4.9 @aiox-data-engineer — Dara, a Engenheira de Dados 📊

**Quando usar:** Banco de dados, esquemas, migrações, consultas.

```
@aiox-data-engineer crie-esquema para o sistema de cobrança multi-inquilino

@aiox-data-engineer projete-indices para a tabela de pedidos

@aiox-data-engineer aplique-migracao adicionar-coluna-metodo-pagamento

@aiox-data-engineer auditoria-seguranca audite políticas RLS

@aiox-data-engineer executar-sql SELECT * FROM usuarios WHERE ultimo_login < NOW() - INTERVAL '90 days'

@aiox-data-engineer semear --tabelas usuarios,pedidos,produtos
```

### 4.10 @aiox-ux — Uma, a Designer UX/UI 🎨

**Quando usar:** Wireframes, design system, tokens, acessibilidade.

```
@aiox-ux wireframe --fidelidade alta para a página de checkout

@aiox-ux pesquise necessidades dos usuários para o fluxo de onboarding

@aiox-ux tokenize extraia tokens de design da exportação do Figma

@aiox-ux verificacao-a11y audite a UI atual para acessibilidade

@aiox-ux configure design system com Tailwind e shadcn/ui

@aiox-ux construa componente Button com variantes
```

### 4.11 @aiox-master — Orion, o Orquestrador 👑

**Quando usar:** Tarefas que cruzam múltiplos domínios, criação de componentes do framework.

```
@aiox-master status mostre o status do sistema

@aiox-master crie novo agente para auditoria de segurança

@aiox-master valide-workflow .aiox-core/development/workflows/development-cycle.yaml

@aiox-master execute-workflow development-cycle --story docs/stories/HISTORIA-42.md

@aiox-master depreciar-componente --name auth-legado --reason "Substituído por OAuth2"
```

### 4.12 @aiox-orchestrator — Orquestrador de Workflows 🔄

**Quando usar:** Execução de workflows multi-agente com handoffs automatizados.

```
@aiox-orchestrator orquestre --workflow development-cycle --story HISTORIA-42

@aiox-orchestrator status-orquestracao

@aiox-orchestrator retomar-orquestracao
```

---

## 5. Comandos Custom (/aiox-*)

### 5.1 /aiox-help

Mostra uma visão geral completa do ecossistema AIOX.

```
/aiox-help
```

Exibe:
- Lista de todos os 12 agentes com ícones e funções
- Como usar @-mentions
- Comandos custom disponíveis
- Onde encontrar tasks, workflows e config

### 5.2 /aiox-init

Inicializa ou verifica a instalação do AIOX no projeto.

```
/aiox-init
```

Exibe:
- Status da instalação (`core-config.yaml`)
- Versão do framework
- Diretórios necessários (.aiox-core, .agent, .aiox)
- Health check (Node.js, npm)
- Recomendações se algo estiver faltando

### 5.3 /aiox-story

Gerencia histórias de usuário.

```
# Criar nova história a partir de um PRD
/aiox-story create --from prd --name "Login Social com OAuth2"

# Validar uma história existente
/aiox-story validate docs/stories/HISTORIA-42.md

# Listar todas as histórias
/aiox-story list
```

### 5.4 /aiox-workflow

Executa um workflow de desenvolvimento completo.

```
# Ciclo completo de desenvolvimento
/aiox-workflow development-cycle --story docs/stories/HISTORIA-42.md

# Pipeline de especificação
/aiox-workflow spec-pipeline --epic "Gateway de Pagamento"

# Loop de QA
/aiox-workflow qa-loop --story HISTORIA-42

# Projeto greenfield fullstack
/aiox-workflow greenfield-fullstack

# Discovery de projeto brownfield
/aiox-workflow brownfield-discovery --path ./app-legado
```

### 5.5 /loop-architect

Ativa o modo **Loop Engineering** — ciclo auto-corretivo com roadmap, testes e lessons learned.

```
/loop-architect Estou na fase 2, vamos implementar o módulo X
```

O agente `@aiox-dev` segue o fluxo:
1. Lê `roadmap.md`, `lessons.md` e `state.json` (cria se não existirem)
2. Analisa a próxima tarefa ou o erro atual
3. Codifica a solução
4. Executa os testes automaticamente
5. Se falhar: registra em `lessons.md`, refatora e retenta (máx 2x)
6. Atualiza o roadmap e o state

---

## 5a. Skills Especializadas

O projeto inclui **11 skills** do [Tech Leads Club](https://agent-skills.techleads.club/skills/) em `.opencode/skills/`. Carregue-as com o comando `skill` para estender as capacidades dos agentes:

| Skill | Agente | Quando usar |
|-------|--------|-------------|
| `tlc-spec-driven` | master/orchestrator | Planejamento em 4 fases (especificar → desenhar → tasks → implementar) |
| `security-best-practices` | qa | Revisão de segurança OWASP/CWE por linguagem |
| `playwright-skill` | qa | Automação de testes E2E no navegador |
| `tactical-ddd` | architect | DDD tático (aggregates, repositories, value objects) |
| `figma` | ux | Design-to-code via integração Figma |
| `web-quality-audit` | qa | Auditoria completa de qualidade web |
| `aws-advisor` | devops | Arquitetura AWS (custo, segurança, performance) |
| `skill-architect` | master | Criação de novas skills para o framework |
| `codenavi` | dev | Navegação inteligente em codebases |
| `sentry` | devops | Monitoramento e diagnóstico de erros |
| `loop-engineering` | dev | Ciclo auto-corretivo (roadmap → código → teste → correção) |

Exemplo de uso:
```
@aiox-qa *load-skill security-best-practices
@aiox-qa revise o código de autenticação com foco em OWASP
```

---

## 6. Fluxos de Desenvolvimento Real

### 6.1 Planejamento Completo de Funcionalidade

```bash
# 1. Analista pesquisa o mercado e concorrência
@aiox-analyst realize-pesquisa-mercado para revisão de código com IA

# 2. Gerente de Produto cria o PRD
@aiox-pm crie-prd --name "Funcionalidade de Revisão de Código com IA"
@aiox-pm levante-requisitos

# 3. Arquiteto desenha a solução
@aiox-architect avalie-complexidade para o épico de revisão de código com IA
@aiox-architect crie-arquitetura-fullstack
```

### 6.2 Ciclo de Desenvolvimento de uma História

```bash
# 1. Scrum Master cria a história
@aiox-sm crie história a partir do PRD para "Implementar endpoint da API de Revisão IA"

# 2. Dono do Produto valida
@aiox-po validar-rascunho-historia docs/stories/HISTORIA-42.md

# 3. Dev implementa
@aiox-dev develop HISTORIA-42 --mode interactive

# 4. Dev roda os testes
@aiox-dev execute-os-testes

# 5. QA revisa
@aiox-qa revise HISTORIA-42

# 6. DevOps faz deploy
@aiox-devops push --branch feature/HISTORIA-42
@aiox-devops criar-pr --story HISTORIA-42
```

### 6.3 Portão de Qualidade Completo

```bash
# QA executa o portão de qualidade completo
@aiox-qa gate

# Análise de risco
@aiox-qa perfil-risco

# Design de testes
@aiox-qa design-testes

# Verificação de segurança
@aiox-qa verificacao-seguranca

# Rastreabilidade
@aiox-qa rastreie HISTORIA-42
```

### 6.4 Operações de Banco de Dados

```bash
# Modelagem
@aiox-data-engineer modele-dominio para o domínio de cobrança por assinatura

# Esquema + RLS
@aiox-data-engineer crie-esquema para SaaS multi-inquilino
@aiox-data-engineer crie-politicas-rls para isolamento de inquilinos

# Migração
@aiox-data-engineer execucao-seca adicionar-coluna-ciclo-cobranca
@aiox-data-engineer aplique-migracao adicionar-coluna-ciclo-cobranca
@aiox-data-engineer teste-fumaca

# Performance
@aiox-data-engineer analise-performance para a consulta de pedidos
@aiox-data-engineer projete-indices para a tabela de usuários
```

### 6.5 DevOps & Infraestrutura

```bash
# Setup inicial
@aiox-devops bootstrap-ambiente
@aiox-devops configurar-github

# CI/CD
@aiox-devops configurar-ci

# Isolamento com worktree
@aiox-devops criar-worktree HISTORIA-45
@aiox-devops listar-worktrees
@aiox-devops mesclar-worktree HISTORIA-45

# Ferramentas MCP
@aiox-devops buscar-mcp database
@aiox-devops adicionar-mcp --server supabase --url http://localhost:54321

# Verificação de saúde
@aiox-devops verificacao-saude
```

### 6.6 Pipeline de Especificação (ADE)

```bash
# 1. PM levanta requisitos
@aiox-pm levante-requisitos

# 2. Arquiteto avalia complexidade
@aiox-architect avalie-complexidade

# 3. Analista pesquisa dependências
@aiox-analyst pesquise-deps

# 4. PM escreve a especificação
@aiox-pm escreva-spec

# 5. QA critica a especificação
@aiox-qa critique-spec
```

### 6.7 Build Autônomo (ADE)

```bash
# Build completo com isolamento de worktree
@aiox-dev build HISTORIA-42

# Build autônomo modo loop
@aiox-dev build-autonomo HISTORIA-42

# Verificar status
@aiox-dev status-build

# Retomar de checkpoint
@aiox-dev retomar-build
```

### 6.8 Loop Engineering

O **Loop Engineering** é um ciclo auto-corretivo que segue: roadmap → código → testes → correção.

```bash
# Ativar o modo loop via comando custom
/loop-architect Estou na fase 3, vamos implementar o serviço X
```

Fluxo completo do loop:

```
┌─────────────────┐
│   roadmap.md    │  ← tarefas pendentes
│   lessons.md    │  ← histórico de erros
│   state.json    │  ← estado atual
└────────┬────────┘
         ▼
┌─────────────────┐
│   Analisar      │  ← próxima tarefa ou erro
└────────┬────────┘
         ▼
┌─────────────────┐
│   Codificar     │  ← implementar solução
└────────┬────────┘
         ▼
┌─────────────────┐
│   Testar        │  ← npm test / pytest / dotnet test
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
 Passou    Falhou
    │         │
    │    ┌────▼────────┐
    │    │ lessons.md  │ ← registrar erro
    │    │ Refatorar   │ ← tentar correção (máx 2x)
    │    └────┬────────┘
    │         ▼
    │    ┌─────────┐
    │    │ Perguntar│ ← se 2x falhar
    │    └─────────┘
    ▼
┌─────────────────┐
│ Atualizar       │  ← roadmap.md, state.json
└─────────────────┘
```

---

## 7. Workflows Orchestrated

### 7.1 Ciclo de Desenvolvimento Completo

O workflow `development-cycle` orquestra 6 fases automaticamente:

```
┌──────────┐    ┌──────────────┐    ┌──────────────┐
│   PO     │───▶│   Executor   │───▶│ Auto-Correção│
│ Valida   │    │   (Dinâmico) │    │ (se ativo)   │
└──────────┘    └──────────────┘    └──────────────┘
     │                                      │
     │ ┌────────────────────────────────────┘
     │ │
     │ ▼
┌──────────────┐    ┌──────────┐    ┌──────────────┐
│ Portão Qual. │───▶│  DevOps  │───▶│  Checkpoint  │
│  (≠ Executor)│    │   Push   │    │   (Humano)   │
└──────────────┘    └──────────┘    └──────────────┘
                                           │
                          ┌────────────────┼────────────────┐
                          │                │                │
                          ▼                ▼                ▼
                       [ GO ]          [ PAUSA ]       [ ABORTAR ]
```

```bash
/aiox-workflow development-cycle --story docs/stories/HISTORIA-42.md
```

### 7.2 Loop de Garantia de Qualidade

```bash
/aiox-workflow qa-loop --story HISTORIA-42
```

### 7.3 Pipeline de Especificação

```bash
/aiox-workflow spec-pipeline --epic "Integração de Gateway de Pagamento"
```

### 7.4 Projeto Greenfield

```bash
/aiox-workflow greenfield-fullstack
/aiox-workflow greenfield-service --name servico-notificacoes
/aiox-workflow greenfield-ui
```

### 7.5 Discovery Brownfield

```bash
/aiox-workflow brownfield-discovery --path ./monolito-legado
/aiox-workflow brownfield-fullstack
```

---

## 8. Dicas e Boas Práticas

### 8.1 Encadeamento de Agentes

Você pode usar múltiplos agentes em sequência para um fluxo completo:

```
@aiox-architect desenhe a arquitetura do módulo de pagamentos
# ... revisa o output ...
@aiox-pm transforme isso em histórias de usuário
# ... revisa as histórias ...
@aiox-sm crie a primeira história para implementação
# ... revisa ...
@aiox-dev implemente a história HISTORIA-42
# ... depois de implementar ...
@aiox-qa revise o código implementado
```

### 8.2 Execução de Tarefas Específicas

Cada agente pode executar tarefas específicas. Use `*help` para ver a lista completa dentro do agente:

```
@aiox-dev
*help    # mostra todos os comandos disponíveis
```

### 8.3 Qualidade e Segurança

- O `@aiox-devops` é o **único** autorizado a fazer `git push` — isso é forçado nas permissões
- `@aiox-qa` deve ser sempre um agente **diferente** do executor da história
- Use `@aiox-devops verificacao-saude` para verificar a saúde do sistema

### 8.4 Camada de Memória (Gotchas)

O agente `@aiox-dev` mantém uma memória de padrões e problemas recorrentes (gotchas):

```
@aiox-dev *gotchas                           # listar todos
@aiox-dev *gotcha "Evitar promises aninhadas" # adicionar novo
@aiox-dev *gotchas --category database       # filtrar por categoria
```

### 8.5 Sincronização

Se os agentes AIOX forem atualizados (nova versão do framework), sincronize as definições:

```bash
node bin/opencode-integration.js sync
```

### 8.6 Combinar com Recursos Nativos do OpenCode

A integração é transparente — você pode usar recursos do OpenCode junto com AIOX:

```bash
# Usar /init do OpenCode para atualizar AGENTS.md
/init

# Usar Modo Plano (Tab) para revisar antes de executar
# (aperte Tab para alternar para Modo Plano, depois Tab novamente para Build)

# Usar /undo se o agente fizer algo errado
/undo

# Compartilhar sessão com o time
/share
```

---

## 9. Referência Rápida

### Todos os Agentes

| @-mention | Nome | Ícone | Função | Comandos Principais |
|-----------|------|-------|--------|-------------------|
| `@aiox-master` | Orion | 👑 | Orquestrador | `*status`, `*executar-workflow`, `*criar` |
| `@aiox-orchestrator` | — | 🔄 | Workflows | `*orquestrar`, `*status-orquestracao` |
| `@aiox-analyst` | Atlas | 🔍 | Análise | `*pesquisar-deps`, `*brainstorm` |
| `@aiox-pm` | Morgan | 📋 | Produto | `*criar-prd`, `*levantar-requisitos` |
| `@aiox-architect` | Aria | 🏛️ | Arquitetura | `*avaliar-complexidade`, `*criar-plano` |
| `@aiox-ux` | Uma | 🎨 | Design | `*wireframe`, `*tokenizar`, `*verif-acessibilidade` |
| `@aiox-sm` | River | 🌊 | Scrum | `*criar`, `*validacao-historia` |
| `@aiox-dev` | Dex | 💻 | Desenvolvimento | `*develop`, `*build`, `*executar-testes` |
| `@aiox-qa` | Quinn | ✅ | Qualidade | `*gate`, `*revisar`, `*verificacao-seguranca` |
| `@aiox-po` | Pax | 🎯 | Dono do Produto | `*validar-rascunho-historia`, `*backlog-adicionar` |
| `@aiox-data-engineer` | Dara | 📊 | Dados | `*criar-esquema`, `*aplicar-migracao` |
| `@aiox-devops` | Gage | ⚡ | DevOps | `*push`, `*criar-pr`, `*configurar-ci` |

### Comandos Custom

| Comando | Descrição |
|---------|-----------|
| `/aiox-help` | Ajuda geral do ecossistema |
| `/aiox-init` | Verificar/instalar AIOX |
| `/aiox-story` | Gerenciar histórias |
| `/aiox-workflow` | Executar workflows |
| `/loop-architect` | Loop Engineering (roadmap → código → teste → fix) |

### Skills Disponíveis

| Skill | Agente | Categoria |
|-------|--------|-----------|
| `tlc-spec-driven` | master/orchestrator | Planejamento |
| `security-best-practices` | qa | Segurança |
| `playwright-skill` | qa | Automação |
| `tactical-ddd` | architect | Arquitetura |
| `figma` | ux | Design |
| `web-quality-audit` | qa | Qualidade |
| `aws-advisor` | devops | Cloud |
| `skill-architect` | master | Criação |
| `codenavi` | dev | Navegação |
| `sentry` | devops | Monitoramento |
| `loop-engineering` | dev | Ciclo auto-corretivo |

### Workflows Disponíveis

| Workflow | Descrição |
|----------|-----------|
| `development-cycle` | Ciclo completo PO → Dev → QA → DevOps |
| `qa-loop` | Loop de qualidade |
| `spec-pipeline` | Pipeline de especificação |
| `greenfield-fullstack` | Projeto novo fullstack |
| `greenfield-service` | Novo serviço |
| `greenfield-ui` | Novo frontend |
| `brownfield-fullstack` | Projeto existente fullstack |
| `brownfield-discovery` | Discovery de legado |
| `auto-worktree` | Automação de worktree |
| `story-development-cycle` | Ciclo de história única |

### Localização dos Recursos

| Recurso | Caminho |
|---------|---------|
| Definições dos agentes | `.aiox-core/development/agents/` |
| Tarefas executáveis | `.aiox-core/development/tasks/` (215+) |
| Workflows | `.aiox-core/development/workflows/` |
| Checklists | `.aiox-core/development/checklists/` |
| Templates | `.aiox-core/development/templates/` |
| Config | `.aiox-core/core-config.yaml` |
| Skills OpenCode | `.opencode/skills/aiox-core/SKILL.md` |
| Skills Tech Leads Club | `.opencode/skills/<nome>/SKILL.md` (11 skills) |
| Loop Engineering skill | `.opencode/skills/loop-engineering/SKILL.md` |
| Script de sincronização | `bin/opencode-integration.js` |
| Sessões | `.aiox/sessions/` |
| Logs | `.aiox/logs/` |

---

---

## 10. Exemplo Completo: Criar um Sistema de CRM do Zero

Este exemplo mostra o fluxo completo desde um diretório vazio até um sistema de CRM funcional, usando todos os agentes AIOX no OpenCode.

### 10.1 Obter o Projeto

```bash
# Clone o fork com AIOX + integração OpenCode
git clone https://github.com/marcelojrrangel/aiox-core.git crm-ia
cd crm-ia

# Instale as dependências
npm install

# Abra o OpenCode
opencode
```

### 10.2 Verificar o Ambiente

```
# Dentro do OpenCode, verifique se está tudo pronto
/aiox-init

# Veja a ajuda do ecossistema
/aiox-help
```

### 10.3 Levantar Requisitos com o Analista

```
@aiox-analyst realize uma sessão de discovery para um sistema de CRM
que precisa de: gestão de contatos, pipeline de vendas, emissão de
notas fiscais, relatórios e dashboard. Quero entender o mercado e
os requisitos iniciais.
```

O Atlas vai pesquisar, fazer perguntas e gerar um briefing de projeto.

### 10.4 Criar o PRD com o Product Manager

```
@aiox-pm crie-prd --name "Sistema de CRM Inteligente"
```

O Morgan vai interagir com você para definir:
- Visão do produto
- Público-alvo
- Funcionalidades prioritárias (MoSCoW)
- Métricas de sucesso
- Roadmap inicial

### 10.5 Desenhar a Arquitetura com a Arquiteta

```
@aiox-architect avalie-complexidade para o épico de CRM
```

Após a avaliação:

```
@aiox-architect crie-arquitetura-fullstack
```

A Aria vai definir:
- Stack tecnológica (React + Node + PostgreSQL sugerido)
- Modelo de dados (entidades: Contato, Negócio, Usuário, NotaFiscal)
- API RESTful
- Estrutura de diretórios
- Estratégia de deploy

### 10.6 Criar o Design System com a UX

```
@aiox-ux configure design system com Tailwind e shadcn/ui
@aiox-ux wireframe --fidelidade alta para o dashboard do CRM
@aiox-ux pesquise os padrões de UI para sistemas de gestão empresarial
```

### 10.7 Modelar o Banco de Dados

```
@aiox-data-engineer crie-esquema para o CRM com as tabelas:
contatos (id, nome, email, telefone, empresa_id, criado_em),
empresas (id, nome, cnpj, segmento),
negocios (id, contato_id, valor, etapa, fechado_em),
usuarios (id, nome, email, cargo),
notas_fiscais (id, negocio_id, valor, emissao, status)

@aiox-data-engineer crie-politicas-rls para isolamento por empresa

@aiox-data-engineer semear --tabelas usuarios,empresas
```

### 10.8 Criar o Épico e as Histórias com o PM e Scrum Master

```
@aiox-pm crie-epico --name "Gestão de Contatos"
```

Depois:

```
@aiox-sm crie história para "CRUD de contatos com busca e filtros"

@aiox-sm crie história para "Importação de contatos via CSV"

@aiox-sm crie história para "Vinculação de contatos a empresas"
```

### 10.9 Desenvolver a Primeira História

```
@aiox-po validar-rascunho-historia docs/stories/HISTORIA-1.md
```

Após aprovação:

```
@aiox-dev implemente a história HISTORIA-1 com testes
```

O Dex vai:
1. Criar o modelo Contato no banco
2. Implementar a API REST (GET, POST, PUT, DELETE)
3. Criar a página de listagem com busca e filtros
4. Escrever testes unitários e de integração
5. Executar os testes

```
@aiox-dev execute-os-testes
```

### 10.10 Revisar a Qualidade

```
@aiox-qa revise a HISTORIA-1

@aiox-qa gate

@aiox-qa verificacao-seguranca
```

Se houver problemas, o Quinn gera um relatório com correções:

```
@aiox-dev aplique as correções solicitadas pelo QA
@aiox-dev execute-os-testes
```

### 10.11 Avançar para as Próximas Histórias

```bash
# Pipeline de especificação para a segunda funcionalidade
/aiox-workflow spec-pipeline --epic "Pipeline de Vendas"
```

```
@aiox-sm crie a história para "Kanban de pipeline de vendas com arrastar e soltar"

@aiox-po validar-rascunho-historia docs/stories/HISTORIA-2.md

@aiox-dev implemente a história HISTORIA-2

@aiox-qa revise a HISTORIA-2
```

### 10.12 Entregar com DevOps

Após cada história aprovada:

```
@aiox-devops push --branch feature/HISTORIA-1
@aiox-devops criar-pr --story HISTORIA-1

# Quando o PR for aprovado no GitHub:
@aiox-devops push --branch feature/HISTORIA-2
@aiox-devops criar-pr --story HISTORIA-2
```

### 10.13 Fluxo Contínuo (Sprint)

```bash
# A cada ciclo de desenvolvimento:
/aiox-workflow development-cycle --story docs/stories/HISTORIA-3.md
```

```
@aiox-sm crie a história para "Dashboard com gráficos de métricas de vendas"
@aiox-sm crie a história para "Emissão de notas fiscais integrada"
@aiox-sm crie a história para "Relatórios exportáveis (PDF/Excel)"
```

Cada história segue o mesmo ciclo:
PM/SM → PO (valida) → Dev (implementa) → QA (revisa) → DevOps (entrega)

---

### Resumo do Fluxo Completo

```
Discovery                    →  @aiox-analyst
PRD e Estratégia             →  @aiox-pm
Arquitetura                  →  @aiox-architect
Design System + UX           →  @aiox-ux
Modelagem de Dados           →  @aiox-data-engineer
Épicos e Histórias           →  @aiox-pm → @aiox-sm
Validação das Histórias      →  @aiox-po
Implementação                →  @aiox-dev
Revisão e Qualidade          →  @aiox-qa
Entrega e Deploy             →  @aiox-devops
```

Em poucas horas você sai de um diretório vazio para um sistema de CRM funcional com testes, documentação e pipeline de entrega configurados.

> **Dica final:** Sempre que entrar em um agente, digite `*help` para ver os comandos específicos daquele agente. Cada um tem dezenas de comandos especializados.
