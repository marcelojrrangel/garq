# Documento de Arquitetura e Design Técnico — CodeMin

**Produto:** CodeMin — Assistente de Codificação LLM Local (CPU-only)
**Versão do Documento:** 1.0.0
**Autor:** 🏛️ Aria (AIOX System Architect)
**Data:** 2026-07-09
**Base:** PRD v1.0.0 (`docs/codemin-prd.md`) + Histórias de Usuário v1.1.0 (`docs/codemin-stories.md`)
**Status:** Aprovado

---

## 1. Visão Geral da Arquitetura

### 1.1 C4 Level 1 — Diagrama de Contexto do Sistema

O CodeMin é uma interface de linha de comando (CLI) que orquestra modelos de linguagem locais (LLMs) para ferramentas de desenvolvimento como OpenCode e Continue.dev. Ele abstrai a complexidade de setup do Ollama e llama.cpp, oferecendo instalação, gerenciamento de modelos e uma bridge compatível com a API da OpenAI.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                         SISTEMA DE CÓDIGO (Software System)                  │
│                                                                              │
│  ┌──────────┐    ┌──────────────────┐    ┌────────────────────────────┐     │
│  │          │    │                  │    │                            │     │
│  │ Developer│────┼─► CodeMin CLI   │────┼─► OpenCode / Continue.dev  │     │
│  │          │    │                  │    │                            │     │
│  └──────────┘    └─────┬─┬─────────┘    └────────────────────────────┘     │
│                        │ │                                                  │
│                        │ │                                                  │
│                        │ ▼                                                  │
│                        │ ┌──────────────────────┐                          │
│                        │ │    Ollama Runtime    │                          │
│                        │ │   (0.3.x ou superior)│                          │
│                        │ └──────────┬───────────┘                          │
│                        │            │                                      │
│                        │            ▼                                      │
│                        │ ┌──────────────────────┐                          │
│                        │ │    llama.cpp         │                          │
│                        │ │  (binário nativo)    │                          │
│                        │ └──────────┬───────────┘                          │
│                        │            │                                      │
│                        │            ▼                                      │
│                        │ ┌──────────────────────┐                          │
│                        │ │  HuggingFace Hub     │                          │
│                        │ │ (download de modelos)│                          │
│                        │ └──────────────────────┘                          │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

  [Person]       [Software System]     [External System]
  ┌────────┐     ┌──────────────┐      ┌───────────────┐
  │Developer│────►│  CodeMin CLI │─────►│    Ollama     │
  └────────┘     └──────────────┘      └───────┬───────┘
                                               │
                                               ▼
                                        ┌───────────────┐
                                        │   llama.cpp   │
                                        └───────┬───────┘
                                                │
                                                ▼
                                         ┌──────────────┐
                                         │ HuggingFace   │
                                         │   Hub        │
                                         └──────────────┘
```

### 1.2 C4 Level 2 — Diagrama de Container

O CodeMin é estruturado em 4 camadas verticais bem definidas:

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         CodeMin CLI (Node.js 18+)                        │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                   COMMANDS LAYER (Commander.js)                  │   │
│  │                                                                  │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │   │
│  │  │ install  │ │  chat    │ │  status  │ │  doctor  │           │   │
│  │  └─────┬────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘           │   │
│  │        │           │            │            │                 │   │
│  │  ┌─────▼───────────▼────────────▼────────────▼──────┐          │   │
│  │  │              SERVICES LAYER                       │          │   │
│  │  │  ┌─────────────┐ ┌──────────────┐                │          │   │
│  │  │  │InstallEngine│ │ModelManager  │                │          │   │
│  │  │  └──────┬──────┘ └──────┬───────┘                │          │   │
│  │  │  ┌──────▼──────┐ ┌──────▼───────┐                │          │   │
│  │  │  │OllamaManager│ │ConfigGen     │                │          │   │
│  │  │  └──────┬──────┘ └──────┬───────┘                │          │   │
│  │  │  ┌──────▼──────┐ ┌──────▼───────┐                │          │   │
│  │  │  │ChatEngine   │ │HealthChecker │                │          │   │
│  │  │  └──────┬──────┘ └──────┬───────┘                │          │   │
│  │  │  ┌──────▼──────┐                               │          │   │
│  │  │  │BenchmarkEng │                               │          │   │
│  │  │  └─────────────┘                               │          │   │
│  │  └─────────────────────────────────────────────────┘          │   │
│  │                                                               │   │
│  │  ┌──────────────────────────────────────────────────────────┐ │   │
│  │  │                    UTILS LAYER                            │ │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │ │   │
│  │  │  │OllamaCli │ │OpenAI    │ │Downloader│ │Checksum  │   │ │   │
│  │  │  │-ent      │ │-Compat   │ │          │ │Verifier  │   │ │   │
│  │  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │ │   │
│  │  │  ┌──────────┐ ┌──────────┐                             │ │   │
│  │  │  │Platform  │ │Types     │                             │ │   │
│  │  │  └──────────┘ └──────────┘                             │ │   │
│  │  └──────────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    OLLAMA RUNTIME (externo)                       │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │   │
│  │  │  Ollama CLI  │  │  Ollama      │  │  Model Storage        │  │   │
│  │  │  (ollama)    │  │  Server (API)│  │  ~/.ollama/models/    │  │   │
│  │  └──────────────┘  └──────────────┘  └───────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                    FILE SYSTEM                                    │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │   │
│  │  │ ~/.codemin/  │  │  project/    │  │  System Temp          │  │   │
│  │  │  (config)    │  │  .codemin/   │  │  (downloads parciais) │  │   │
│  │  └──────────────┘  └──────────────┘  └───────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────────┘
```

### 1.3 Padrão Arquitetural

**CLI + Runtime Bridge + Service Orchestration**

O CodeMin adota o padrão **CLI + Runtime Bridge + Service Orchestration** com 4 camadas:

| Camada | Responsabilidade | Tecnologia |
|--------|-----------------|------------|
| **Commands Layer** | Interface com usuário (CLI), parsing de argumentos, help | Commander.js |
| **Services Layer** | Orquestração de lógica de negócio, coordenação entre utils | TypeScript Classes |
| **Utils Layer** | Comunicação com runtime externo (Ollama), I/O, utils | TypeScript + undici |
| **Runtime Bridge** | Abstração sobre o Ollama Server (API REST nativa) | OpenAI-compat wrapper |

**Princípios arquiteturais:**

1. **Separação de responsabilidades** — cada camada conhece apenas a camada imediatamente abaixo
2. **Injeção de dependência manual** — sem DI framework, construtores recebem dependências explicitamente
3. **Fail-fast** — validação de pré-requisitos no início de cada comando
4. **Zero telemetry** — nenhum dado sai da máquina do usuário
5. **Runtime Bridge sobre proxy** — comunicação direta com Ollama via REST, sem servidor intermediário

**Fluxo de chamada típico:**

```
User Input
    │
    ▼
┌─────────────────┐     ┌──────────────────┐     ┌───────────────┐
│ Command (CLI)   │────►│ Service          │────►│ Util          │
│ ex.: install.ts │     │ InstallEngine.ts │     │ OllamaClient  │
└─────────────────┘     └──────────────────┘     └───────┬───────┘
                                                         │
                                                         ▼
                                                  ┌───────────────┐
                                                  │ Ollama Server │
                                                  │ :11434/api    │
                                                  └───────────────┘
```

---

## 2. Stack Tecnológica Detalhada

### 2.1 Tabela de Tecnologias

| Tecnologia | Versão | Propósito | Justificativa |
|-----------|--------|-----------|---------------|
| **TypeScript** | 5.3+ | Linguagem principal | Tipagem estática para segurança em código assíncrono complexo; interfaces auto-documentadas |
| **Node.js** | 18+ (LTS) | Runtime | APIs nativas `fetch`, `fs/promises`, `child_process` sem polyfills; disponível em todos SOs |
| **Commander.js** | 11.x | CLI framework | Parser de argumentos com subcomandos, auto-help, sem dependências externas |
| **undici** | 6.x | HTTP client | Implementação nativa performática (fetch), suporte a streaming SSE |
| **@inquirer/prompts** | 5.x | Prompts interativos | Alternativa madura ao enquirer; suporte a autocomplete, checkbox, confirmação |
| **chalk** | 5.x | Terminal styling | Saída colorida cross-platform, sem dependências nativas |
| **js-yaml** | 4.x | YAML parser | Leitura/escrita de `config.yaml` com preservação de formatação |
| **Vitest** | 2.x | Test runner | Compatível com ecossistema Node, execução paralela, cobertura embutida |
| **nock** | 14.x | HTTP mocking | Interceptação de requisições HTTP em testes sem servidor real |
| **Biome** | 1.x | Linter + Formatter | Unificado (substituto ESLint + Prettier), 10x mais rápido em WASM |
| **Ollama** | 0.3+ | Runtime de LLM | Interface padronizada para llama.cpp; server REST nativo + CLI |
| **llama.cpp** | b41+ (via Ollama) | Backend de inferência | Quantização Q4_K_M, suporte a AVX2/AVX512, CPU-only sem GPU |
| **Qwen2.5-Coder** | 7B Q4_K_M | Modelo padrão | Melhor custo-benefício para autocomplete de código em CPU |

### 2.2 Estrutura do package.json

```json
{
  "name": "@codemin/cli",
  "version": "1.0.0",
  "description": "CodeMin - Local LLM orchestrator for development tools",
  "type": "module",
  "bin": {
    "codemin": "./dist/cli/index.js"
  },
  "engines": {
    "node": ">=18.0.0"
  },
  "exports": {
    ".": "./dist/index.js",
    "./bridge": "./dist/bridge/index.js"
  },
  "scripts": {
    "build": "tsc -p tsconfig.json",
    "dev": "tsx watch src/cli/index.ts",
    "start": "node dist/cli/index.js",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "lint": "biome check src/",
    "lint:fix": "biome check --write src/",
    "format": "biome format --write src/",
    "typecheck": "tsc --noEmit",
    "prepare": "npm run build"
  },
  "dependencies": {
    "commander": "^11.1.0",
    "undici": "^6.18.0",
    "@inquirer/prompts": "^5.0.0",
    "chalk": "^5.3.0",
    "js-yaml": "^4.1.0"
  },
  "devDependencies": {
    "typescript": "^5.3.3",
    "@types/node": "^20.11.0",
    "vitest": "^2.0.0",
    "nock": "^14.0.0",
    "@biomejs/biome": "^1.8.0",
    "tsx": "^4.7.0",
    "c8": "^9.1.0"
  }
}
```

### 2.3 Justificativa das Escolhas

| Escolha | Opções Rejeitadas | Motivo |
|---------|------------------|--------|
| **Commander.js** | `yargs` (complexidade), `oclif` (overhead), raw `process.argv` (manual) | Commander.js é o mais leve sem sacrificar funcionalidade |
| **undici** | `axios` (pesado, sem SSE nativo), `got` (mais lento), `node:http` (baixo nível) | undici é o HTTP client mais performático para Node.js |
| **@inquirer/prompts** | `enquirer` (manutenção reduzida), `inquirer` (monorepo grande) | @inquirer é o sucessor moderno do inquirer, modular |
| **Vitest** | `jest` (lento, config verbosa), `mocha` (sem cobertura nativa) | Vitest é 20x mais rápido que Jest em modo watch |
| **Biome** | ESLint + Prettier (2 ferramentas, lentas) | Biome unifica lint + format em uma ferramenta WASM rápida |
| **Ollama** | Chamada direta a llama.cpp (requer build C++) | Ollama abstrai build, download e servidor do llama.cpp |
| **js-yaml** | `yaml` (menos popular), `@std/yaml` (Deno) | js-yaml é o padrão de fato no ecossistema Node.js |

---

## 3. Estrutura de Diretórios do Projeto

### 3.1 Árvore do Projeto

```
D:\projetos\IA\llm-local\
│
├── .github/
│   └── workflows/
│       ├── ci.yml                    # CI: lint, typecheck, test, coverage
│       ├── release.yml               # Release automatizado via semantic-release
│       └── benchmark.yml             # Benchmarks comparativos noturnos
│
├── src/
│   ├── cli/
│   │   ├── index.ts                  # Entry point: configura Commander.js
│   │   │
│   │   ├── commands/
│   │   │   ├── install.ts            # codemin install [modelo]
│   │   │   ├── chat.ts               # codemin chat [--model]
│   │   │   ├── status.ts             # codemin status
│   │   │   ├── doctor.ts             # codemin doctor (diagnóstico)
│   │   │   ├── list.ts               # codemin list
│   │   │   ├── remove.ts             # codemin remove <modelo>
│   │   │   ├── bridge.ts             # codemin bridge (start/stop)
│   │   │   ├── config.ts             # codemin config (get/set/edit)
│   │   │   └── benchmark.ts          # codemin benchmark
│   │   │
│   │   ├── services/
│   │   │   ├── install-engine.ts     # Orquestrador de instalação
│   │   │   ├── model-manager.ts      # Gerenciamento de modelos
│   │   │   ├── ollama-manager.ts     # Gerenciamento do runtime Ollama
│   │   │   ├── config-generator.ts   # Geração de templates de config
│   │   │   ├── chat-engine.ts        # Motor de chat interativo
│   │   │   ├── health-checker.ts     # Verificação de saúde do sistema
│   │   │   └── benchmark-engine.ts   # Motor de benchmarks
│   │   │
│   │   └── bridge/
│   │       └── index.ts              # Bridge OpenAI-compatible server
│   │
│   ├── utils/
│   │   ├── ollama-client.ts          # HTTP client para API REST do Ollama
│   │   ├── openai-compat.ts          # Mapeamento OpenAI → Ollama
│   │   ├── downloader.ts             # Download com retry e progresso
│   │   ├── checksum.ts               # Verificação SHA256
│   │   ├── platform.ts               # Detecção de SO, arquitetura, CPU
│   │   └── logger.ts                 # Logger estruturado (chalk + levels)
│   │
│   ├── types/
│   │   ├── index.ts                  # Re-exports
│   │   ├── models.ts                 # ModelConfig, ModelCatalog, etc.
│   │   ├── config.ts                 # Config interfaces
│   │   ├── ollama.ts                 # Ollama API types
│   │   └── openai.ts                 # OpenAI API types
│   │
│   └── config/
│       ├── models-catalog.json       # Catálogo oficial de modelos
│       └── defaults.yaml             # Configurações padrão
│
├── tests/
│   ├── unit/
│   │   ├── services/
│   │   │   ├── install-engine.spec.ts
│   │   │   ├── model-manager.spec.ts
│   │   │   ├── ollama-manager.spec.ts
│   │   │   └── chat-engine.spec.ts
│   │   ├── utils/
│   │   │   ├── ollama-client.spec.ts
│   │   │   ├── downloader.spec.ts
│   │   │   ├── checksum.spec.ts
│   │   │   └── platform.spec.ts
│   │   └── bridge/
│   │       └── index.spec.ts
│   │
│   ├── integration/
│   │   ├── install-flow.spec.ts      # Fluxo completo de instalação
│   │   ├── bridge-api.spec.ts        # Testes da bridge com Ollama mock
│   │   └── config-gen.spec.ts        # Geração de configurações
│   │
│   ├── fixtures/
│   │   ├── models-catalog.test.json  # Catálogo reduzido para testes
│   │   ├── config.test.yaml          # Config de teste
│   │   ├── ollama-responses/         # Respostas mockadas do Ollama
│   │   │   ├── list-models.json
│   │   │   ├── generate.json
│   │   │   └── chat.json
│   │   └── binaries/                 # Binários dummy para testes
│   │       └── ollama-dummy.exe
│   │
│   └── e2e/
│       ├── install.e2e.ts            # Teste end-to-end (requer Ollama real)
│       └── chat.e2e.ts               # Teste de chat real (opcional --live)
│
├── scripts/
│   ├── build.ps1                     # Script de build Windows
│   ├── release.ps1                   # Script de release
│   └── dev-setup.ps1                 # Setup do ambiente de desenvolvimento
│
├── docs/
│   ├── codemin-arch.md               # ← Este documento
│   ├── README.md
│   └── CONTRIBUTING.md
│
├── .opencode/
│   └── AGENTS.md                     # Configuração de agentes OpenCode
│
├── .aiox-core/                       # Core AIOX framework
│
├── package.json
├── tsconfig.json
├── biome.json
├── vitest.config.ts
└── README.md
```

### 3.2 Diretório do Usuário (`~/.codemin/`)

```
~/.codemin/
│
├── config.yaml                   # Configuração principal do usuário
├── opencode.json                 # Template para OpenCode
├── continue.config.json          # Template para Continue.dev
│
├── models/
│   ├── catalog.json              # Catálogo local de modelos (cópia + custom)
│   └── aliases.json              # Apelidos personalizados para modelos
│
├── ollama/
│   ├── install.log               # Log da instalação do Ollama
│   └── version.json              # Versão instalada do Ollama
│
├── logs/
│   ├── codemin.log               # Log geral
│   ├── install.log               # Log de instalações de modelo
│   └── bridge.log                # Log da bridge API
│
└── cache/
    ├── downloads/                # Downloads parciais (retomáveis)
    └── checksums/                # Cache de checksums verificados
```

---

## 4. Descrição dos Componentes

### 4.1 install.ts (Command)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/commands/install.ts` |
| **Responsabilidade** | Comando `codemin install [modelo]` — orquestrar instalação de modelo local |
| **Entrada** | `modelo` (string opcional), `--from` (caminho local\|URL), `--force` |
| **Saída** | Modelo instalado e configurado. Prints de progresso no terminal |
| **Dependências** | `InstallEngine`, `ModelManager`, `ConfigGenerator` |
| **Comunicação** | Chama `InstallEngine.execute(modelo, opcoes)` |

```typescript
// src/cli/commands/install.ts
import { Command } from 'commander';
import { InstallEngine } from '../services/install-engine.js';
import { ConfigGenerator } from '../services/config-generator.js';
import { logger } from '../../utils/logger.js';

export function registerInstallCommand(program: Command): void {
  program
    .command('install [modelo]')
    .description('Instalar um modelo de linguagem local')
    .option('-f, --force', 'Forçar reinstalação')
    .option('--from <source>', 'Instalar de caminho local ou URL')
    .action(async (modelo?: string, options?: { force?: boolean; from?: string }) => {
      try {
        const engine = new InstallEngine();
        await engine.execute(modelo ?? 'qwen2.5-coder:7b', {
          force: options?.force ?? false,
          source: options?.from,
        });

        const gen = new ConfigGenerator();
        await gen.ensureConfigs();
      } catch (error) {
        logger.error('Falha na instalação', error as Error);
        process.exit(1);
      }
    });
}
```

### 4.2 chat.ts (Command)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/commands/chat.ts` |
| **Responsabilidade** | Comando `codemin chat [--model]` — sessão interativa de chat |
| **Entrada** | `--model` (string), `--system` (system prompt), `--stream` (boolean) |
| **Saída** | Sessão REPL com streaming de tokens no terminal |
| **Dependências** | `ChatEngine` |
| **Comunicação** | Chama `ChatEngine.startSession(modelo, opcoes)` |

### 4.3 status.ts (Command)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/commands/status.ts` |
| **Responsabilidade** | Comando `codemin status` — exibir estado atual do sistema |
| **Entrada** | Nenhuma |
| **Saída** | Tabela no terminal com: Ollama (running/stopped), modelos instalados, recursos |
| **Dependências** | `OllamaManager`, `ModelManager`, `HealthChecker` |
| **Comunicação** | `OllamaManager.status()`, `ModelManager.listInstalled()`, `HealthChecker.run()` |

### 4.4 doctor.ts (Command)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/commands/doctor.ts` |
| **Responsabilidade** | Comando `codemin doctor` — diagnóstico completo do sistema |
| **Entrada** | Nenhuma |
| **Saída** | Relatório de diagnóstico com ✓/✗ para cada verificação |
| **Dependências** | `HealthChecker` |
| **Comunicação** | Chama `HealthChecker.diagnose()` que executa 12 verificações |

### 4.5 install-engine.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/install-engine.ts` |
| **Responsabilidade** | Orquestrador do pipeline completo de instalação de modelo |
| **Entrada** | `modelName: string`, `options: InstallOptions` |
| **Saída** | `InstallResult` (sucesso/erro + metadados) |
| **Dependências** | `OllamaManager`, `ModelManager`, `ConfigGenerator`, `Downloader`, `Checksum` |
| **Comunicação** | Coordena chamadas entre Utils Layer e File System |

```typescript
// src/cli/services/install-engine.ts
import { OllamaManager } from './ollama-manager.js';
import { ModelManager } from './model-manager.js';
import { ConfigGenerator } from './config-generator.js';
import { Downloader } from '../../utils/downloader.js';
import { Checksum } from '../../utils/checksum.js';
import { logger } from '../../utils/logger.js';

export interface InstallOptions {
  force: boolean;
  source?: string;
}

export interface InstallResult {
  success: boolean;
  modelName: string;
  path?: string;
  sizeBytes?: number;
  error?: string;
}

export class InstallEngine {
  private ollama = new OllamaManager();
  private models = new ModelManager();
  private config = new ConfigGenerator();
  private downloader = new Downloader();
  private checksum = new Checksum();

  async execute(modelName: string, options: InstallOptions): Promise<InstallResult> {
    logger.info(`Iniciando instalação de ${modelName}`);

    // Fase 1: Verificar pré-requisitos
    await this.verifyPrerequisites();

    // Fase 2: Resolver modelo (catálogo ou URL direta)
    const modelInfo = await this.models.resolve(modelName, options.source);

    // Fase 3: Download (se necessário)
    if (!await this.models.isCached(modelInfo)) {
      const downloadPath = await this.downloader.download(
        modelInfo.url,
        { sha256: modelInfo.sha256 }
      );
      await this.checksum.verify(downloadPath, modelInfo.sha256);
    }

    // Fase 4: Importar para Ollama
    await this.ollama.importModel(modelInfo);

    // Fase 5: Verificar instalação
    const verified = await this.ollama.verifyModel(modelName);

    // Fase 6: Gerar configs
    if (verified) {
      await this.config.ensureConfigs();
    }

    return { success: verified, modelName, sizeBytes: modelInfo.size };
  }

  private async verifyPrerequisites(): Promise<void> {
    const ollamaOk = await this.ollama.isInstalled();
    if (!ollamaOk) {
      throw new Error('Ollama não encontrado. Execute "codemin doctor" para diagnóstico.');
    }
  }
}
```

### 4.6 model-manager.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/model-manager.ts` |
| **Responsabilidade** | Gerenciamento de catálogo, resolução e cache de modelos |
| **Entrada** | `catalogPath?`, `cachePath?` |
| **Saída** | Métodos para resolver, listar, verificar cache de modelos |
| **Dependências** | `Platform` (para detectar CPU/OS), FS (`node:fs/promises`) |
| **Comunicação** | Lê `models-catalog.json` do projeto e `catalog.json` do usuário |

### 4.7 ollama-manager.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/ollama-manager.ts` |
| **Responsabilidade** | Gerenciar o runtime Ollama: instalar, iniciar, parar, verificar |
| **Entrada** | Comandos via `child_process.execFile` |
| **Saída** | Status, versão, logs do processo |
| **Dependências** | `Platform`, `OllamaClient` |
| **Comunicação** | Executa binário `ollama` via shell ou usa `OllamaClient` para REST API |

```typescript
// src/cli/services/ollama-manager.ts
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { platform } from '../../utils/platform.js';
import { OllamaClient } from '../../utils/ollama-client.js';
import { logger } from '../../utils/logger.js';

const execFileAsync = promisify(execFile);

export class OllamaManager {
  private client = new OllamaClient();

  get binaryPath(): string {
    return platform.isWindows
      ? `${process.env.LOCALAPPDATA}\\Ollama\\ollama.exe`
      : '/usr/bin/ollama';
  }

  async isInstalled(): Promise<boolean> {
    try {
      const { stdout } = await execFileAsync(this.binaryPath, ['--version']);
      return stdout.includes('ollama');
    } catch {
      return false;
    }
  }

  async getVersion(): Promise<string> {
    const { stdout } = await execFileAsync(this.binaryPath, ['--version']);
    return stdout.trim();
  }

  async isRunning(): Promise<boolean> {
    return this.client.ping();
  }

  async start(): Promise<void> {
    if (await this.isRunning()) return;
    logger.info('Iniciando Ollama server...');
    const { spawn } = await import('node:child_process');
    const child = spawn(this.binaryPath, ['serve'], {
      detached: true,
      stdio: 'ignore',
    });
    child.unref();
    await this.waitForReady(30_000);
  }

  async importModel(modelInfo: { name: string; path: string }): Promise<void> {
    logger.info(`Importando modelo ${modelInfo.name}...`);
    const { stdout } = await execFileAsync(this.binaryPath, [
      'create', modelInfo.name,
      '-f', modelInfo.path,
    ]);
    logger.debug(stdout);
  }

  async verifyModel(name: string): Promise<boolean> {
    try {
      const models = await this.client.listModels();
      return models.some(m => m.name === name);
    } catch {
      return false;
    }
  }

  private async waitForReady(timeout: number): Promise<void> {
    const start = Date.now();
    while (Date.now() - start < timeout) {
      if (await this.isRunning()) return;
      await new Promise(r => setTimeout(r, 500));
    }
    throw new Error('Timeout aguardando Ollama server iniciar');
  }
}
```

### 4.8 config-generator.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/config-generator.ts` |
| **Responsabilidade** | Gerar templates de configuração para OpenCode, Continue.dev e config.yaml |
| **Entrada** | `modelName`, `templateType` |
| **Saída** | Arquivos de configuração escritos em `~/.codemin/` |
| **Dependências** | `ModelManager`, `js-yaml`, FS |
| **Comunicação** | Leitura de templates embutidos, escrita no FS do usuário |

### 4.9 chat-engine.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/chat-engine.ts` |
| **Responsabilidade** | Sessão interativa de chat com streaming de tokens |
| **Entrada** | `modelName`, `systemPrompt`, `stream: boolean` |
| **Saída** | Texto gerado pelo LLM, token a token via SSE |
| **Dependências** | `OllamaClient` |
| **Comunicação** | `POST /api/chat` do Ollama com SSE streaming |

### 4.10 health-checker.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/health-checker.ts` |
| **Responsabilidade** | Diagnóstico completo de 12 verificações |
| **Entrada** | Nenhuma |
| **Saída** | `CheckResult[]` com status, mensagem e sugestão |
| **Dependências** | `OllamaManager`, `ModelManager`, `Platform`, `FS` |
| **Comunicação** | Leitura de sistema, execução de binários, chamadas HTTP |

### 4.11 benchmark-engine.ts (Service)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/cli/services/benchmark-engine.ts` |
| **Responsabilidade** | Executar benchmarks de performance (latência, throughput, P50/P95/P99) |
| **Entrada** | `modelName`, `prompts: string[]`, `warmupRounds`, `testRounds` |
| **Saída** | `BenchmarkReport` com estatísticas |
| **Dependências** | `OllamaClient` |
| **Comunicação** | Chamadas repetidas ao Ollama, medição com `performance.now()` |

### 4.12 ollama-client.ts (Util)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/utils/ollama-client.ts` |
| **Responsabilidade** | HTTP client tipado para API REST do Ollama |
| **Entrada** | Métodos: `ping()`, `listModels()`, `generate()`, `chat()`, `pullModel()`, `deleteModel()` |
| **Saída** | Respostas tipadas (TypeScript interfaces) |
| **Dependências** | `undici` (fetch + SSE streaming) |
| **Comunicação** | `http://localhost:11434/api/*` |

### 4.13 openai-compat.ts (Util)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/utils/openai-compat.ts` |
| **Responsabilidade** | Mapeamento de parâmetros OpenAI → Ollama (e vice-versa) |
| **Entrada** | `OpenAIRequest` |
| **Saída** | `OllamaRequest` |
| **Dependências** | Nenhuma (função pura) |
| **Comunicação** | Chamada síncrona de transformação |

### 4.14 downloader.ts (Util)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/utils/downloader.ts` |
| **Responsabilidade** | Download de arquivos com retry, progresso e resume |
| **Entrada** | `url: string`, `destination?: string`, `options: { sha256?, retry?, timeout? }` |
| **Saída** | `downloadedPath: string` |
| **Dependências** | `undici` (fetch), FS |
| **Comunicação** | HTTP(S) para HuggingFace Hub ou URLs arbitrárias |

### 4.15 checksum.ts (Util)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/utils/checksum.ts` |
| **Responsabilidade** | Verificação de integridade via SHA256 |
| **Entrada** | `filePath: string`, `expectedHash: string` |
| **Saída** | `boolean` (válido/inválido) |
| **Dependências** | `node:crypto` |
| **Comunicação** | Stream de leitura do arquivo + hash |

### 4.16 platform.ts (Util)

| Atributo | Detalhe |
|----------|---------|
| **Arquivo** | `src/utils/platform.ts` |
| **Responsabilidade** | Detecção de plataforma: SO, arquitetura, CPU features, RAM |
| **Entrada** | Nenhuma (lê do sistema) |
| **Saída** | `PlatformInfo` com todas as informações |
| **Dependências** | `node:os`, `node:process`, `node:child_process` |
| **Comunicação** | `os.cpus()`, `os.totalmem()`, `process.arch` |

---

## 5. Fluxos Detalhados

### 5.1 Fluxo: `codemin install`

```
Usuário                    CodeMin CLI                  Sistema                    Ollama                   HuggingFace
  │                           │                           │                         │                          │
  │ codemin install           │                           │                         │                          │
  │ qwen2.5-coder:7b          │                           │                         │                          │
  │──────────────────────────►│                           │                         │                          │
  │                           │                           │                         │                          │
  │                           │  1. Verificar pré-req     │                         │                          │
  │                           │──────────────────────────►│                         │                          │
  │                           │                           │  ollama --version        │                          │
  │                           │                           │─────────────────────────►│                          │
  │                           │                           │◄─────────────────────────│                          │
  │                           │                           │ ollama 0.3.0             │                          │
  │                           │◄──────────────────────────│                         │                          │
  │                           │                           │                         │                          │
  │                           │  2. Resolver modelo       │                         │                          │
  │                           │  ┌───────────────────┐    │                         │                          │
  │                           │  │ models-catalog    │    │                         │                          │
  │                           │  │ qwen.2.5-coder:7b │    │                         │                          │
  │                           │  │ url: hf.co/...    │    │                         │                          │
  │                           │  │ sha256: a1b2...   │    │                         │                          │
  │                           │  └───────────────────┘    │                         │                          │
  │                           │                           │                         │                          │
  │                           │  3. Baixar modelo         │                         │                          │
  │                           │──────────────────────────►│                         │                          │
  │                           │                           │  GET /qwen/7b-Q4_K_M.gguf│                          │
  │                           │                           │──────────────────────────────────────────────────────►│
  │                           │                           │                         │                          │
  │   ████████░░ 72%          │                           │◄───── Stream ──────────│◄───── Stream ─────────────│
  │◄──────────────────────────│                           │  128 MB/s               │                          │
  │                           │                           │                         │                          │
  │                           │  4. Verificar SHA256      │                         │                          │
  │                           │──────────────────────────►│                         │                          │
  │                           │                           │ crypto:createHash()     │                          │
  │                           │                           │  a1b2c3... === match    │                          │
  │                           │◄──────────────────────────│ OK                      │                          │
  │                           │                           │                         │                          │
  │                           │  5. Importar Ollama       │                         │                          │
  │                           │──────────────────────────►│                         │                          │
  │                           │                           │ ollama create qwen2.5:  │                         │
  │                           │                           │────────────────────────►│                          │
  │                           │                           │                         │                          │
  │                           │  6. Verificar modelo      │                         │                          │
  │                           │──────────────────────────►│                         │                          │
  │                           │                           │ ollama list             │                          │
  │                           │                           │────────────────────────►│                          │
  │                           │                           │◄────────────────────────│                          │
  │                           │                           │ qwen2.5-coder:7b exists │                          │
  │                           │◄──────────────────────────│                         │                          │
  │                           │                           │                         │                          │
  │                           │  7. Gerar configurações   │                         │                          │
  │                           │  ┌───────────────────┐    │                         │                          │
  │                           │  │ config.yaml       │    │                         │                          │
  │                           │  │ opencode.json     │    │                         │                          │
  │                           │  │ continue.config   │    │                         │                          │
  │                           │  └───────────────────┘    │                         │                          │
  │                           │                           │                         │                          │
  │  ✓ Modelo instalado!      │                           │                         │                          │
  │◄──────────────────────────│                           │                         │                          │
  │                           │                           │                         │                          │
```

### 5.2 Fluxo: OpenCode → CodeMin → Ollama → llama.cpp (Chat)

```
┌──────────┐    ┌──────────────┐    ┌────────────┐    ┌────────────┐    ┌─────────────┐
│ OpenCode  │    │ CodeMin      │    │ Ollama     │    │ llama.cpp  │    │ Prompt      │
│ (Agent)   │    │ Bridge       │    │ Server     │    │ (backend)  │    | (CPU)      │
└─────┬─────┘    └──────┬───────┘    └─────┬──────┘    └─────┬──────┘    └──────┬──────┘
      │                  │                  │                │                  │
      │ POST /v1/chat/   │                  │                │                  │
      │ completions      │                  │                │                  │
      │ {model: "qwen",  │                  │                │                  │
      │  messages: [...],│                  │                │                  │
      │  stream: true}   │                  │                │                  │
      │─────────────────►│                  │                │                  │
      │                  │                  │                │                  │
      │    1. OpenAI → Ollama mapping      │                │                  │
      │    ┌─────────────────────────┐     │                │                  │
      │    │ max_tokens → num_predict│     │                │                  │
      │    │ temperature → temp      │     │                │                  │
      │    │ top_p → top_p           │     │                │                  │
      │    │ stream → stream         │     │                │                  │
      │    │ messages → messages     │     │                │                  │
      │    │ (system→role:system)    │     │                │                  │
      │    └─────────────────────────┘     │                │                  │
      │                  │                  │                │                  │
      │                  │ POST /api/chat  │                │                  │
      │                  │ {model:"qwen2.5:│                │                  │
      │                  │  7b", messages: │                │                  │
      │                  │  [...], stream: │                │                  │
      │                  │  true, options: │                │                  │
      │                  │  {num_predict:  │                │                  │
      │                  │   2048, temp:   │                │                  │
      │                  │   0.7}}         │                │                  │
      │                  │────────────────►│                │                  │
      │                  │                  │                │                  │
      │                  │                  │ ── llama_eval()─────►          │
      │                  │                  │◄── token 1 ────────│◄────────────│
      │                  │                  │◄── token 2 ────────│◄────────────│
      │                  │                  │◄── token 3 ────────│◄────────────│
      │                  │                  │     ...            │              │
      │                  │                  │                │                  │
      │                  │◄── SSE stream ──│◄── tokens ─────│                  │
      │                  │ data: {"choices"│                │                  │
      │◄── SSE stream ───│ [{"delta":      │                │                  │
      │ data: {"choices" │ {"content":"A"}}│                │                  │
      │ [{...}]}         │ ]}              │                │                  │
      │                  │ data: [DONE]    │                │                  │
      │                  │                 │                │                  │
```

### 5.3 Fluxo: VS Code → Continue.dev → CodeMin → Ollama → llama.cpp (FIM)

```
┌──────────┐    ┌──────────────┐    ┌────────────┐    ┌────────────┐    ┌──────────────┐
│ VS Code   │    │ Continue.dev │    │ CodeMin    │    │ Ollama     │    │ llama.cpp    │
│ (Editor)  │    │ (Extension)  │    │ Bridge     │    │ Server     │    │ (FIM mode)   │
└─────┬─────┘    └──────┬───────┘    └─────┬──────┘    └─────┬──────┘    └──────┬───────┘
      │                  │                  │                │                  │
      │[user digita]      │                  │                │                  │
      │ function calc() { │                  │                │                  │
      │   return          │                  │                │                  │
      │                   │                  │                │                  │
      │ Continue.dev      │                  │                │                  │
      │ detecta trigger   │                  │                │                  │
      │ (300ms idle)      │                  │                │                  │
      │                  │                  │                │                  │
      │ POST /v1/         │                  │                │                  │
      │ completions       │                  │                │                  │
      │ {model: "qwen2.5" │                  │                │                  │
      │  coder:7b",       │                  │                │                  │
      │  prompt: "function│                  │                │                  │
      │  calc() {\\n  ",   │                  │                │                  │
      │  suffix: "(a,b)", │                  │                │                  │
      │  max_tokens: 64,  │                  │                │                  │
      │  temperature: 0.2,│                  │                │                  │
      │  stream: false}   │                  │                │                  │
      │─────────────────►│                  │                │                  │
      │                  │                  │                │                  │
      │    1. Detectar FIM (Fill-in-Middle)                 │                  │
      │    prompt + suffix → template FIM                   │                  │
      │    ┌────────────────────────────────────┐           │                  │
      │    │ <fim_prefix>function calc() {       │           │                  │
      │    │ return<fim_suffix>(a,b)<fim_middle> │           │                  │
      │    └────────────────────────────────────┘           │                  │
      │                  │                  │                │                  │
      │                  │ POST /api/       │                │                  │
      │                  │ generate         │                │                  │
      │                  │ {model:"qwen2.5  │                │                  │
      │                  │  coder:7b",      │                │                  │
      │                  │  prompt:"<fim_   │                │                  │
      │                  │  prefix>...<fim_ │                │                  │
      │                  │  suffix>...<fim_ │                │                  │
      │                  │  middle>",       │                │                  │
      │                  │  options:{       │                │                  │
      │                  │   num_predict:64,│                │                  │
      │                  │   temperature:0.2│                │                  │
      │                  │  }, stream:false}│                │                  │
      │                  │────────────────►│                │                  │
      │                  │                  │                │                  │
      │                  │                  │ FIM special    │                  │
      │                  │                  │ tokens detect  │                  │
      │                  │                  │ ── llama_eval()───►              │
      │                  │                  │◄── "return"────│◄────────────────│
      │                  │                  │◄── " a + b"────│◄────────────────│
      │                  │                  │◄── ";"─────────│◄────────────────│
      │                  │                  │                │                  │
      │                  │◄── {"response":  │                │                  │
      │                  │     "return a +  │                │                  │
      │                  │     b;"}         │                │                  │
      │◄── "return a + b"│                  │                │                  │
      │   ;"             │                  │                │                  │
      │                  │                  │                │                  │
      │ [código inserido] │                  │                │                  │
      │ function calc() { │                  │                │                  │
      │   return a + b;   │                  │                │                  │
      │ }                 │                  │                │                  │
```

### 5.4 Fluxo: `codemin doctor`

```
Usuário                    CodeMin CLI                  Sistema                    Ollama
  │                           │                           │                         │
  │ codemin doctor            │                           │                         │
  │──────────────────────────►│                           │                         │
  │                           │                           │                         │
  │                           │ ┌─ Running 12 checks ──┐  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 1. Node.js version    │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ node --version = 18     │
  │                           │ │ ✓ Node.js 20.11.0     │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 2. Ollama installed   │  │                         │
  │                           │ │───────────────────────►│  │                         │
  │                           │ │◄────────────────────── │  │                         │
  │                           │ │ ✓ Ollama 0.3.0        │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 3. Ollama server      │  │                         │
  │                           │ │───────────────────────►│  │                         │
  │                           │ │◄────────────────────── │  │                         │
  │                           │ │ ✓ Server running      │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 4. Ollama API         │  │                         │
  │                           │ │───────────────────────►│  │ GET /api/tags           │
  │                           │ │◄────────────────────── │  │                         │
  │                           │ │ ✓ API respondendo      │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 5. Modelo padrão      │  │                         │
  │                           │ │───────────────────────►│  │                         │
  │                           │ │◄────────────────────── │  │                         │
  │                           │ │ ✓ qwen2.5-coder:7b    │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 6. Config files       │  │                         │
  │                           │ │  ~/.codemin/config    │  │                         │
  │                           │ │  ~/.codemin/opencode  │  │                         │
  │                           │ │  ~/.codemin/continue  │  │                         │
  │                           │ │ ✓ Todos presentes     │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 7. Disco disponível   │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ 45GB free               │
  │                           │ │ ✓ 45GB = 20GB mínimo  │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 8. RAM disponível     │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ 32GB total              │
  │                           │ │ ✓ 32GB = 16GB mínimo  │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │ 9. CPU features       │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ AVX2 ✓, SSSE3 ✓         │
  │                           │ │ ✓ AVX2 suportado      │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │10. Permissões         │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ ~/.codemin writable ✓   │
  │                           │ │ ✓ Tudo ok             │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │11. Porta 11434        │  │                         │
  │                           │ │───────────────────────►│  │                         │
  │                           │ │◄────────────────────── │  │                         │
  │                           │ │ ✓ Porta disponível     │  │                         │
  │                           │ │                       │  │                         │
  │                           │ │12. Network (Hugging   │  │                         │
  │                           │ │     Face)             │  │                         │
  │                           │ │─────────────────────► │  │                         │
  │                           │ │◄──────────────────────│  │ ✓ Acessível             │
  │                           │ └───────────────────────┘  │                         │
  │                           │                           │                         │
  │  ┌─────────────────────────────────────┐               │                         │
  │  │ CodeMin Doctor Report               │               │                         │
  │  │─────────────────────────────────────│               │                         │
  │  │ ✓ Node.js    v20.11.0               │               │                         │
  │  │ ✓ Ollama     0.3.0                  │               │                         │
  │  │ ✓ Server     Running :11434         │               │                         │
  │  │ ✓ API        OK                     │               │                         │
  │  │ ✓ Modelo     qwen2.5-coder:7b       │               │                         │
  │  │ ✓ Configs    OK (3/3)               │               │                         │
  │  │ ✓ Disco      45 GB livre            │               │                         │
  │  │ ✓ RAM        32 GB                  │               │                         │
  │  │ ✓ CPU        AVX2 sim              │               │                         │
  │  │ ✓ Permissões OK                     │               │                         │
  │  │ ✓ Porta      11434 livre            │               │                         │
  │  │ ✓ Network    Conexão OK             │               │                         │
  │  │─────────────────────────────────────│               │                         │
  │  │ 12/12 checks passed                │               │                         │
  │  └─────────────────────────────────────┘               │                         │
  │◄──────────────────────────│                           │                         │
```

---

## 6. API Bridge (OpenAI-compatible)

### 6.1 Estratégia

O CodeMin **não implementa um proxy separado**. Em vez disso, ele expõe um endpoint HTTP que traduz chamadas no formato OpenAI API para chamadas nativas ao Ollama via REST. Isso elimina latência extra, consumo de porta e complexidade operacional.

**Decisão arquitetural:** Bridge → chamada direta ao Ollama (sem servidor intermediário)

```
OpenCode/Continue.dev                    CodeMin Bridge                    Ollama Server
       │                                      │                                │
       │  POST /v1/chat/completions           │                                │
       │─────────────────────────────────────►│                                │
       │                                      │                                │
       │     Valida request OpenAI            │                                │
       │     Mapeia parâmetros                │                                │
       │     ┌─────────────────────┐          │                                │
       │     │ OpenAI → Ollama     │          │                                │
       │     └─────────────────────┘          │                                │
       │                                      │  POST /api/chat               │
       │                                      │───────────────────────────────►│
       │                                      │                                │
       │                                      │◄────── SSE stream ────────────│
       │                                      │                                │
       │     Mapeia resposta Ollama → OpenAI  │                                │
       │     ┌─────────────────────┐          │                                │
       │     │ Ollama → OpenAI     │          │                                │
       │     └─────────────────────┘          │                                │
       │                                      │                                │
       │◄──── SSE stream (formato OpenAI) ────│                                │
```

### 6.2 Endpoints

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/v1/chat/completions` | POST | Chat completo com histórico de mensagens |
| `/v1/completions` | POST | Geração de texto simples (usado para FIM/autocomplete) |
| `/v1/models` | GET | Lista modelos disponíveis no Ollama |
| `/v1/embeddings` | POST | Geração de embeddings |
| `/health` | GET | Health check da bridge |

### 6.3 Request/Response — Chat Completions

**Request (OpenAI format):**

```json
{
  "model": "qwen2.5-coder:7b",
  "messages": [
    {
      "role": "system",
      "content": "Você é um assistente de código."
    },
    {
      "role": "user",
      "content": "Escreva uma função quickSort em TypeScript"
    }
  ],
  "max_tokens": 1024,
  "temperature": 0.7,
  "top_p": 0.9,
  "stream": true
}
```

**Response (streaming — SSE):**

```
data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{"content":"function"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{"content":" quickSort"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{"content":"(arr"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{"content":")"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1700000000,"model":"qwen2.5-coder:7b","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**Response (não-streaming):**

```json
{
  "id": "chatcmpl-abc123",
  "object": "chat.completion",
  "created": 1700000000,
  "model": "qwen2.5-coder:7b",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "function quickSort(arr: number[]): number[] {\n  if (arr.length <= 1) return arr;\n  const pivot = arr[0];\n  const left = arr.slice(1).filter(x => x < pivot);\n  const right = arr.slice(1).filter(x => x >= pivot);\n  return [...quickSort(left), pivot, ...quickSort(right)];\n}"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 42,
    "completion_tokens": 67,
    "total_tokens": 109
  }
}
```

### 6.4 Request/Response — Completions (FIM / Autocomplete)

**Request (OpenAI format):**

```json
{
  "model": "qwen2.5-coder:7b",
  "prompt": "def calculate_sum(a, b):\n    return ",
  "suffix": "\n\nresult = calculate_sum(3, 5)",
  "max_tokens": 64,
  "temperature": 0.2,
  "stop": ["\n\n"],
  "stream": false
}
```

**Response:**

```json
{
  "id": "cmpl-abc456",
  "object": "text_completion",
  "created": 1700000000,
  "model": "qwen2.5-coder:7b",
  "choices": [
    {
      "text": "a + b",
      "index": 0,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 18,
    "completion_tokens": 3,
    "total_tokens": 21
  }
}
```

### 6.5 OpenAI → Ollama Parameter Mapping

| Parâmetro OpenAI | Parâmetro Ollama | Descrição | Exemplo |
|-----------------|-----------------|-----------|---------|
| `model` | `model` | Nome do modelo | `"qwen2.5-coder:7b"` |
| `messages` | `messages` | Array de mensagens (mesmo formato) | `[{role,content}]` |
| `prompt` | `prompt` | Prompt para completions/texto | `"function calc()"` |
| `suffix` | `suffix` | Sufixo para FIM | `"(a,b)"` |
| `max_tokens` | `options.num_predict` | Máximo de tokens a gerar | `2048` |
| `temperature` | `options.temperature` | Temperatura (0.0 — 2.0) | `0.7` |
| `top_p` | `options.top_p` | Nucleus sampling | `0.9` |
| `top_k` | `options.top_k` | Top-K sampling | `40` |
| `frequency_penalty` | `options.frequency_penalty` | Penalidade de frequência | `0.0` |
| `presence_penalty` | `options.presence_penalty` | Penalidade de presença | `0.0` |
| `stop` | `options.stop` | Sequências de parada | `["\n"]` |
| `stream` | `stream` | Streaming ativado/desativado | `true` |
| `seed` | `options.seed` | Semente para reproducibilidade | `42` |
| `n` | — | Não suportado (sempre 1) | — |
| `logprobs` | — | Não suportado | — |
| `user` | — | Ignorado | — |

### 6.6 Error Handling

| Código | Significado | Causa Provável | Ação |
|--------|-------------|----------------|------|
| `400` | Bad Request | JSON malformado, campo obrigatório ausente | Validar request antes de enviar |
| `404` | Model Not Found | Modelo não existe no Ollama | Executar `codemin install <modelo>` |
| `413` | Payload Too Large | Contexto excede limite do modelo | Reduzir `max_tokens` ou mensagens |
| `429` | Too Many Requests | Ollama sobrecarregado | Aguardar e retentar com backoff |
| `500` | Internal Error | Ollama server crashou | Executar `codemin doctor` |
| `502` | Bad Gateway | Ollama não está rodando | Executar `ollama serve` |
| `503` | Service Unavailable | Modelo ainda carregando | Aguardar alguns segundos |

**Exemplo de resposta de erro:**

```json
{
  "error": {
    "message": "Model 'qwen2.5-coder:7b' not found. Please pull the model first.",
    "type": "not_found",
    "param": "model",
    "code": "model_not_found"
  }
}
```

---

## 7. Gerenciamento de Modelos

### 7.1 Estrutura de Diretórios

**Modelos gerenciados via Ollama:**

```
~/.ollama/
├── models/
│   ├── blobs/                  # Armazenamento de blobs (sharded)
│   │   ├── sha256-a1b2c3...
│   │   ├── sha256-d4e5f6...
│   │   └── ...
│   └── manifests/
│       └── registry.ollama.ai/
│           └── library/
│               ├── qwen2.5-coder/
│               │   ├── 7b/     # Manifest do modelo 7B
│               │   └── 1.5b/   # Manifest do modelo 1.5B
│               └── deepseek-coder/
│                   └── 6.7b/
```

**Modelos em armazenamento direto (alternativa):**

```
~/.codemin/
└── models/
    ├── direct/                 # Modelos GGUF sem Ollama
    │   ├── qwen2.5-coder-7b-q4_k_m.gguf
    │   └── codegemma-2b.gguf
    └── catalog.json            # Catálogo local unificado
```

### 7.2 Catálogo de Modelos (`models-catalog.json`)

```json
{
  "$schema": "https://codemin.dev/schemas/models-catalog.json",
  "version": "1.0.0",
  "updated": "2026-07-09",
  "models": [
    {
      "id": "qwen2.5-coder:7b",
      "name": "Qwen2.5-Coder 7B Instruct",
      "provider": "Alibaba Cloud",
      "description": "Modelo padrão para autocomplete e chat de código. Melhor custo-benefício em CPU.",
      "languages": ["typescript", "javascript", "python", "java", "go", "rust", "c++", "c#"],
      "contextLength": 32768,
      "size": "4.7 GB",
      "sizeBytes": 4700000000,
      "quantization": "Q4_K_M",
      "urls": {
        "ollama": "qwen2.5-coder:7b",
        "gguf": "https://huggingface.co/Qwen/Qwen2.5-Coder-7B-Instruct-GGUF/resolve/main/qwen2.5-coder-7b-instruct-q4_k_m.gguf"
      },
      "sha256": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1",
      "requirements": {
        "minRam": "8 GB",
        "recommendedRam": "16 GB",
        "cpuFeatures": ["avx2"],
        "minDisk": "5 GB"
      },
      "performance": {
        "tokensPerSecond": "15-25 t/s (CPU AVX2, 8 threads)",
        "timeToFirstToken": "300-800ms"
      },
      "recommended": true
    },
    {
      "id": "qwen2.5-coder:1.5b",
      "name": "Qwen2.5-Coder 1.5B Instruct",
      "provider": "Alibaba Cloud",
      "description": "Modelo leve para máquinas com pouca RAM ou CPU fraca.",
      "languages": ["typescript", "javascript", "python", "java"],
      "contextLength": 32768,
      "size": "1.0 GB",
      "sizeBytes": 1000000000,
      "quantization": "Q4_K_M",
      "urls": {
        "ollama": "qwen2.5-coder:1.5b",
        "gguf": "https://huggingface.co/Qwen/Qwen2.5-Coder-1.5B-Instruct-GGUF/resolve/main/qwen2.5-coder-1.5b-instruct-q4_k_m.gguf"
      },
      "sha256": "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c",
      "requirements": {
        "minRam": "2 GB",
        "recommendedRam": "4 GB",
        "cpuFeatures": [],
        "minDisk": "1.5 GB"
      },
      "performance": {
        "tokensPerSecond": "40-60 t/s (CPU AVX2)",
        "timeToFirstToken": "100-300ms"
      },
      "recommended": false
    },
    {
      "id": "codellama:7b-code",
      "name": "CodeLlama 7B Code",
      "provider": "Meta",
      "description": "Modelo especializado em geração de código com suporte a FIM nativo.",
      "languages": ["typescript", "javascript", "python", "java", "c++", "php"],
      "contextLength": 16384,
      "size": "3.8 GB",
      "sizeBytes": 3800000000,
      "quantization": "Q4_K_M",
      "urls": {
        "ollama": "codellama:7b-code",
        "gguf": "https://huggingface.co/TheBloke/CodeLlama-7B-GGUF/resolve/main/codellama-7b.Q4_K_M.gguf"
      },
      "sha256": "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d",
      "requirements": {
        "minRam": "8 GB",
        "recommendedRam": "16 GB",
        "cpuFeatures": ["avx2"],
        "minDisk": "4 GB"
      },
      "performance": {
        "tokensPerSecond": "12-20 t/s (CPU AVX2, 8 threads)",
        "timeToFirstToken": "400-1000ms"
      },
      "recommended": false,
      "tags": ["fim-native"]
    },
    {
      "id": "deepseek-coder:6.7b",
      "name": "DeepSeek Coder 6.7B Instruct",
      "provider": "DeepSeek",
      "description": "Modelo competitivo em benchmarks HumanEval, ótimo para geração de código.",
      "languages": ["python", "typescript", "javascript", "java", "go", "rust"],
      "contextLength": 16384,
      "size": "4.1 GB",
      "sizeBytes": 4100000000,
      "quantization": "Q4_K_M",
      "urls": {
        "ollama": "deepseek-coder:6.7b",
        "gguf": "https://huggingface.co/TheBloke/deepseek-coder-6.7B-instruct-GGUF/resolve/main/deepseek-coder-6.7b-instruct.Q4_K_M.gguf"
      },
      "sha256": "d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e",
      "requirements": {
        "minRam": "8 GB",
        "recommendedRam": "16 GB",
        "cpuFeatures": ["avx2"],
        "minDisk": "5 GB"
      },
      "performance": {
        "tokensPerSecond": "14-22 t/s (CPU AVX2, 8 threads)",
        "timeToFirstToken": "350-900ms"
      },
      "recommended": false
    }
  ]
}
```

### 7.3 Algoritmo de Download

```
                        ALGORITMO DE DOWNLOAD

  1. RESOLVER URL
     ├── Verificar se modelo existe no catálogo local (catalog.json)
     ├── Se existir: obter URL + SHA256 do catálogo
     └── Se não existir:
          ├── Se --from <url>: usar URL fornecida
          └── Se --from <path>: copiar arquivo local

  2. VERIFICAR CACHE
     ├── Procurar em ~/.codemin/cache/downloads/
     ├── Se parcial existir: HEAD request para verificar suporte a Range
     │   ├── Se suportado: resume com header Range
     │   └── Se não: reiniciar download
     └── Se completo: verificar SHA256

  3. DOWNLOAD (com retry)
     ├── fetch(url) com streaming para arquivo temporário
     ├── Atualizar barra de progresso a cada 1MB
     ├── Tentativas: 3 com backoff exponencial (1s → 2s → 4s)
     ├── Timeout total: 30 minutos
     └── Em caso de falha: salvar progresso parcial

  4. VERIFICAR SHA256
     ├── Criar hash stream do arquivo baixado
     ├── Comparar com hash esperado
     ├── Se match: mover para destino final
     └── Se mismatch: deletar arquivo, logar erro, retentar download

  5. REGISTRAR
     ├── Atualizar cache de checksums
     ├── Registrar no catalog.json local (com data de instalação)
     └── Remover arquivo temporário

  6. IMPORTAR (se gerenciado via Ollama)
     ├── ollama create <nome> -f <Modelfile>
     └── Verificar com ollama list
```

### 7.4 Verificação SHA256 (TypeScript)

```typescript
// src/utils/checksum.ts
import { createHash } from 'node:crypto';
import { createReadStream } from 'node:fs';
import { logger } from './logger.js';

export class Checksum {
  /**
   * Verifica integridade de arquivo via SHA256.
   * Usa streaming para não carregar o arquivo inteiro em memória.
   */
  async verify(filePath: string, expectedHash: string): Promise<boolean> {
    const computed = await this.computeHash(filePath);
    const isValid = computed.toLowerCase() === expectedHash.toLowerCase();

    if (isValid) {
      logger.info(`? SHA256 verificado: ${filePath}`);
    } else {
      logger.error(`? SHA256 mismatch em: ${filePath}`);
      logger.debug(`Esperado: ${expectedHash}`);
      logger.debug(`Obtido:   ${computed}`);
    }

    return isValid;
  }

  /**
   * Computa hash SHA256 de um arquivo via stream.
   */
  async computeHash(filePath: string): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const hash = createHash('sha256');
      const stream = createReadStream(filePath);

      stream.on('data', (chunk: Buffer) => {
        hash.update(chunk);
      });

      stream.on('end', () => {
        resolve(hash.digest('hex'));
      });

      stream.on('error', (err) => {
        reject(new Error(`Erro ao ler arquivo para hash: ${err.message}`));
      });
    });
  }

  /**
   * Verificação segura com deleção automática em caso de falha.
   */
  async verifyOrDelete(filePath: string, expectedHash: string): Promise<boolean> {
    const isValid = await this.verify(filePath, expectedHash);

    if (!isValid) {
      const { unlink } = await import('node:fs/promises');
      await unlink(filePath);
      logger.warn(`Arquivo deletado por falha de checksum: ${filePath}`);
    }

    return isValid;
  }
}
```

### 7.5 Estratégia de Fallback (Árvore de Decisão)

```
                        +----------------------------------------------+
                        |  usuario executa                             |
                        |  codemin install <modelo>                    |
                        +-----------------------+----------------------+
                                                |
                                                v
                   +------------------------------------------------------+
                   | Modelo existe no catalogo local?                     |
                   +------------------+------------------+----------------+
                                      | Sim              | Nao
                                      v                  v
                   +----------------------+  +-----------------------------+
                   | Usar URL do          |  | Modelo e um alias           |
                   | catalogo             |  | conhecido?                  |
                   +----------+-----------+  +---------+--------+---------+
                              |                   Sim    |        |  Nao
                              |                         v        v
                              |              +--------------------+  +--------------------+
                              |              | Resolver alias     |  | Tentar pull do     |
                              |              | -> modelo real     |  | Ollama Registry    |
                              |              +--------+-----------+  +---------+----------+
                              |                       |                        |
                              |                       v                        v
                              |              +--------------------+  +--------------------+
                              |              | Usar URL do        |  | Pull falhou?       |
                              |              | catalogo           |  |                    |
                              |              +--------+-----------+  +--+--------+--------+
                              |                       |              Sim |        | Nao
                              |                       |                 |        |
                              |                       |                 v        v
                              |                       |        +--------------------+ +----------+
                              |                       |        | Tentar URL         | | Modelo   |
                              |                       |        | alternativa        | | baixado  |
                              |                       |        | do HuggingFace     | | com      |
                              |                       |        | (diferente quant)  | | sucesso  |
                              |                       |        +--------+-----------+ +----------+
                              |                       |                 |
                              |                       |                 v
                              |                       |        +--------------------+
                              |                       |        | Falhou?            |
                              |                       |        +--+--------------+-+
                              |                       |       Sim |              | Nao
                              |                       |         |              |
                              |                       |         v              v
                              |                       |  +--------------------+ +----------+
                              |                       |  | ERRO: Nao foi      | | Modelo   |
                              |                       |  | possivel baixar    | | baixado  |
                              |                       |  | modelo. Sugerir:   | +----------+
                              |                       |  | 1. Verificar net   |
                              |                       |  | 2. Modelo alt      |
                              |                       |  | 3. Download manual |
                              |                       |  +--------------------+
                              v                       v
                   +--------------------------------------------------------+
                   |                Download bem-sucedido                   |
                   +----------------------------+---------------------------+
                                                |
                                                v
                   +--------------------------------------------------------+
                   |            Verificar SHA256                            |
                   +------------------+------------------+------------------+
                                      | Match             | Mismatch
                                      v                  v
                              +----------------+  +---------------------------+
                              | Importar       |  | Deletar arquivo           |
                              | modelo         |  | Tentar novamente (max 3)  |
                              +-------+--------+  +-------------+-------------+
                                      |                          |
                                      v                          v
                              +----------------+         +----------------+
                              | Modelo         |         | Se 3 falhas:   |
                              | instalado!     |         | ERRO critico   |
                              +----------------+         +----------------+
```

### 7.6 Cache e Versionamento

O CodeMin mantém um cache local para evitar downloads redundantes:

```typescript
interface VersionCache {
  [modelId: string]: {
    version: string;           // Hash SHA256 do modelo
    installedAt: string;       // ISO date
    source: 'ollama' | 'direct';
    sizeBytes: number;
    quant: string;             // ex: "Q4_K_M"
  };
}
```

**Política de cache:**
- Modelos baixados via GGUF direto são cacheados em `~/.codemin/cache/downloads/`
- O checksum SHA256 é salvo em `~/.codemin/cache/checksums/`
- Se o mesmo modelo for solicitado novamente e o checksum bater, o download é pulado
- A flag `--force` ignora o cache e força novo download
- Modelos do Ollama são gerenciados pelo próprio Ollama (`ollama pull` detecta updates)

---

## 8. Configuração e Templates

### 8.1 `~/.codemin/config.yaml` (Completo)

```yaml
# CodeMin Configuration
# v1.0.0
# Gerado automaticamente em: 2026-07-09

# --- Modelo Padrao ------------------------------------------------
default_model: qwen2.5-coder:7b

# --- Modelos ------------------------------------------------------
models:
  catalog: ~/.codemin/models/catalog.json
  aliases: ~/.codemin/models/aliases.json
  download:
    cache_dir: ~/.codemin/cache/downloads
    retry:
      attempts: 3
      backoff: exponential
      initial_delay_ms: 1000
    timeout_ms: 1800000        # 30 minutos
    verify_checksum: true
    resume: true

# --- Ollama --------------------------------------------------------
ollama:
  binary: auto                  # auto | caminho explicito
  host: localhost
  port: 11434
  timeout_ms: 30000
  max_retries: 3
  health_check:
    interval_ms: 5000
    timeout_ms: 30000

# --- Bridge API ----------------------------------------------------
bridge:
  enabled: true
  host: localhost
  port: 11435                  # Porta da bridge (diferente da Ollama :11434)
  cors:
    enabled: true
    origins:
      - "*"                    # Em producao, restrinja a origens especificas
  endpoints:
    chat: /v1/chat/completions
    completions: /v1/completions
    models: /v1/models
    embeddings: /v1/embeddings

# --- Chat ----------------------------------------------------------
chat:
  default_system_prompt: |
    Voce e um assistente especializado em programacao.
    Responda com clareza, fornecendo exemplos de codigo quando relevante.
    Use TypeScript como linguagem padrao a menos que especificado.
  history_file: ~/.codemin/logs/chat-history.json
  max_history: 50

# --- Performance ---------------------------------------------------
performance:
  cpu:
    threads: auto               # auto = numero de cores fisicos
    batch_size: 512
    flash_attention: true
    mmap: true
    mlock: false                # Habilite apenas se tiver RAM suficiente
  context:
    default: 4096
    max: 32768                  # Limitado pelo modelo
    autocomplete: 2048          # Contexto menor para autocomplete (mais rapido)
    chat: 8192
  generation:
    default_temperature: 0.7
    autocomplete_temperature: 0.2
    chat_temperature: 0.7
    top_p: 0.9
    top_k: 40
    frequency_penalty: 0.0
    presence_penalty: 0.0

# --- OpenCode Integration ------------------------------------------
opencode:
  auto_configure: true
  config_path: ~/.codemin/opencode.json

# --- Continue.dev Integration --------------------------------------
continue:
  auto_configure: true
  config_path: ~/.codemin/continue.config.json
  autocomplete:
    enabled: true
    trigger_mode: after_stop    # after_stop | always | manual
    debounce_ms: 300

# --- Logging -------------------------------------------------------
logging:
  level: info                   # debug | info | warn | error
  file: ~/.codemin/logs/codemin.log
  max_size_mb: 10
  max_files: 3
  console: true

# --- Telemetry -----------------------------------------------------
telemetry:
  enabled: false                # Zero telemetry por padrao
  # Nenhum dado e coletado ou enviado
```

### 8.2 Template `opencode.json`

```json
{
  "model": {
    "provider": "openai",
    "apiBase": "http://localhost:11435/v1",
    "model": "qwen2.5-coder:7b",
    "apiKey": "not-needed"
  },
  "completions": {
    "enabled": true,
    "model": {
      "provider": "openai",
      "apiBase": "http://localhost:11435/v1",
      "model": "qwen2.5-coder:7b",
      "apiKey": "not-needed"
    },
    "parameters": {
      "maxTokens": 2048,
      "temperature": 0.2,
      "topP": 0.9,
      "stop": []
    }
  },
  "chat": {
    "enabled": true,
    "parameters": {
      "maxTokens": 4096,
      "temperature": 0.7,
      "topP": 0.9
    }
  },
  "embeddings": {
    "enabled": false
  }
}
```

### 8.3 Template `continue.config.json`

```json
{
  "models": [
    {
      "title": "CodeMin Qwen 7B",
      "provider": "openai",
      "model": "qwen2.5-coder:7b",
      "apiKey": "not-needed",
      "apiBase": "http://localhost:11435/v1",
      "contextLength": 32768,
      "completionOptions": {
        "maxTokens": 4096,
        "temperature": 0.7,
        "topP": 0.9
      }
    },
    {
      "title": "CodeMin Qwen 1.5B (Rapido)",
      "provider": "openai",
      "model": "qwen2.5-coder:1.5b",
      "apiKey": "not-needed",
      "apiBase": "http://localhost:11435/v1",
      "contextLength": 32768,
      "completionOptions": {
        "maxTokens": 2048,
        "temperature": 0.7,
        "topP": 0.9
      }
    }
  ],
  "tabAutocompleteModel": {
    "title": "CodeMin Autocomplete",
    "provider": "openai",
    "model": "qwen2.5-coder:7b",
    "apiKey": "not-needed",
    "apiBase": "http://localhost:11435/v1",
    "completionOptions": {
      "maxTokens": 64,
      "temperature": 0.2,
      "topP": 0.9
    }
  },
  "experimental": {
    "local": {
      "promptTemplates": {
        "qwen2.5-coder": {
          "options": {
            "stop": ["<|im_end|>", "<|im_start|>"]
          }
        }
      }
    }
  }
}
```

### 8.4 Parâmetros Otimizados para CPU

**Parâmetros de Chat:**

| Parâmetro | Chat (Padrão) | Chat (CPU Fraca) | Chat (CPU Forte) |
|-----------|--------------|------------------|------------------|
| Modelo | Qwen 7B Q4_K_M | Qwen 1.5B Q4_K_M | Qwen 7B Q4_K_M |
| `num_thread` | auto (cores físicos) | 2-4 | 12-16 |
| `batch_size` | 512 | 256 | 1024 |
| `num_predict` | 4096 | 2048 | 8192 |
| `temperature` | 0.7 | 0.7 | 0.7 |
| `top_k` | 40 | 40 | 40 |
| `top_p` | 0.9 | 0.9 | 0.95 |
| `repeat_penalty` | 1.1 | 1.1 | 1.1 |
| `mmap` | true | true | true |
| `mlock` | false | false | true |
| RAM necessária | 8-12 GB | 2-4 GB | 16+ GB |
| Tokens/s esperado | 15-25 | 40-60 | 20-35 |

**Parâmetros de Autocomplete (FIM):**

| Parâmetro | Valor | Motivo |
|-----------|-------|--------|
| `num_predict` | 64 | Autocomplete precisa de poucos tokens |
| `temperature` | 0.2 | Baixa temperatura = previsibilidade |
| `top_p` | 0.9 | Nucleus sampling padrão |
| `top_k` | 20 | Mais restritivo que chat |
| `batch_size` | 128 | Batch pequeno = latência menor |
| `num_thread` | auto | Aproveitar todos os cores |
| `mmap` | true | Mapear modelo em memória |
| `mlock` | false | Não travar (mais seguro) |
| `fim_prefix` | `<fim_prefix>` | Token especial FIM do Qwen |
| `fim_suffix` | `<fim_suffix>` | Token especial FIM do Qwen |
| `fim_middle` | `<fim_middle>` | Token especial FIM do Qwen |
| Latência alvo | < 500ms | Primeiro token em até 500ms |

---

## 9. Decisões Arquiteturais (ADRs)

### ADR-001: Ollama vs llama.cpp Puro

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-001 |
| **Título** | Usar Ollama como runtime em vez de chamar llama.cpp diretamente |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
O CodeMin precisa executar modelos de linguagem localmente. Duas opções principais: integrar diretamente com llama.cpp (subprocesso + API nativa) ou usar o Ollama como intermediário que gerencia o llama.cpp.

**Decisão:**
Usar **Ollama** como runtime padrão, com fallback para llama.cpp direto em cenários avançados.

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **Ollama** | Instalação simplificada, CLI madura, API REST nativa, gerenciamento de modelos, GPU/CPU automático | Camada extra, dependência externa, menos controle fino |
| **llama.cpp puro** | Controle total, sem dependências, menor overhead, atualizações imediatas | Build complexo (C++), sem gerenciamento de modelos, sem API pronta, requer compilação nativa |
| **llama.cpp via node-llama.cpp** | Integração direta Node.js, sem HTTP overhead | Bindings nativos frágeis, manutenção pesada, falha em Windows |

**Consequências:**
- Positivas: Setup simplificado para o usuário final; API padronizada (REST); gerenciamento de modelos delegado
- Negativas: Dependência do Ollama estar instalado; versão do Ollama pode atrasar features do llama.cpp; consumo extra de ~50MB RAM do processo Ollama server

---

### ADR-002: Node.js vs Rust/Python

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-002 |
| **Título** | Node.js (TypeScript) como linguagem de implementação |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
Escolha da linguagem de implementação para o CodeMin CLI.

**Decisão:**
**Node.js 18+ com TypeScript 5.3+** como plataforma principal.

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **Node.js + TypeScript** | Mesmo ecossistema das ferramentas alvo (VS Code, OpenCode), async nativo, NPM, tipagem | Performance inferior para CPU-bound, garbage collector, depende de runtime externo |
| **Rust** | Performance máxima, binário único, zero runtime, segurança de memória | Curva de aprendizado alta, ecossistema menor para CLI, build cross-compilation complexo |
| **Python** | Ecossistema ML maduro, simplicidade, vastas bibliotecas | GIL limita paralelismo, distribuição complexa (pip vs Docker), performance inferior |

**Consequências:**
- Positivas: Integração trivial com OpenCode e Continue.dev (ambos TS/JS); acesso ao ecossistema NPM; async/await para chamadas HTTP; TypeScript fornece segurança sem sacrificar produtividade
- Negativas: Necessidade de Node.js 18+ instalado; performance 5-10x inferior a Rust em parsing pesado (não crítico para CLI); distribuição via npm requer Node.js

---

### ADR-003: Bridge API vs Chamada Direta

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-003 |
| **Título** | Bridge API direta ao Ollama sem servidor proxy |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
OpenCode e Continue.dev esperam uma API compatível com OpenAI. O Ollama tem sua própria API (diferente da OpenAI). Como expor interface OpenAI-compatível sem adicionar complexidade?

**Decisão:**
Uma **bridge in-process** que traduz chamadas OpenAI para chamadas Ollama no mesmo processo Node.js, sem servidor HTTP separado. A bridge usa o endpoint HTTP nativo do CodeMin.

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **Bridge in-process (escolhida)** | Zero latência extra, sem processo adicional, deploy simples | Acopla runtime ao CLI, porta extra consumida |
| **Proxy HTTP separado** | Desacoplado, pode ser escalado, reutilizável | Complexidade operacional, latência extra, gerenciamento de processo |
| **Plugin Ollama no OpenCode** | Mais integrado, sem bridge | Não resolve para Continue.dev + VS Code, limitado a uma ferramenta |

**Consequências:**
- Positivas: Latência mínima (apenas transformação JSON); deploy simples (mesmo processo); funciona com qualquer ferramenta que suporte OpenAI API
- Negativas: CLI precisa manter o servidor bridge rodando; consumo extra de ~20-30MB RAM para o servidor HTTP

---

### ADR-004: npm vs curl|bash Distribuição

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-004 |
| **Título** | Distribuição via npm (npmjs.com) |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
Como distribuir o CodeMin para os usuários finais?

**Decisão:**
**Distribuição primária via npm** (`npx @codemin/cli`), com fallback para GitHub Releases (tarball/zip).

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **npm (escolhida)** | `npx` sem instalação, versionamento semântico, familiar para devs, CI/CD integrado | Requer Node.js, não funciona sem npm |
| **curl|bash** | Oneliner universal, sem dependências | Risco de segurança, sem versionamento, difícil debugar |
| **Homebrew** | Simples em macOS, fórmula declarativa | Apenas macOS/Linux, manutenção extra |
| **GitHub Releases** | Universal, sem dependências | Download manual, sem auto-update |

**Consequências:**
- Positivas: `npx @codemin/cli install` funciona sem instalação prévia; atualizações via `npm update -g @codemin/cli`; CI/CD com semantic-release
- Negativas: Requer Node.js 18+ (pré-requisito explícito); Windows pode ter caminho longo; latência do `npx` (download na primeira execução)

---

### ADR-005: JSON vs YAML vs TOML

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-005 |
| **Título** | YAML para configuração principal do usuário |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
O CodeMin precisa de um formato de configuração para `~/.codemin/config.yaml`. Também gera `opencode.json` e `continue.config.json` nos formatos esperados por essas ferramentas.

**Decisão:**
**YAML** para configuração principal do CodeMin. **JSON** para os templates gerados (compatibilidade com ferramentas externas).

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **YAML (escolhido para config)** | Legível, comentários suportados, hierarquia natural, menos verboso que JSON | Sensível a indentação, parsing mais complexo, sem schema oficial |
| **JSON** | Universal, schemas (JSON Schema), parsing nativo JS | Sem comentários, mais verboso, chaves com aspas |
| **TOML** | Legível, tabelas, comentários, sem ambiguidade | Menos conhecido, ecossistema menor, parsing sem suporte nativo |

**Consequências:**
- Positivas: Configuração legível com comentários explicativos; usuários podem editar manualmente; js-yaml é maduro e estável
- Negativas: YAML pode ser ambíguo em casos complexos; validação mais difícil que JSON Schema; templates externos em JSON (dois formatos)

---

### ADR-006: Gerenciamento de Múltiplos Modelos

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-006 |
| **Título** | Catálogo centralizado com extensibilidade local |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
O CodeMin precisa gerenciar múltiplos modelos (diferentes fornecedores, tamanhos, quantizações).

**Decisão:**
**Catálogo centralizado** (`models-catalog.json` no pacote) + **catálogo local** (`~/.codemin/models/catalog.json`) com merge hierárquico. Catálogo local sobrescreve o central.

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **Catálogo + local (escolhida)** | Flexível, atualizável, usuário pode customizar | Complexidade de merge, dois arquivos |
| **Apenas catálogo remoto** | Sempre atualizado, sem arquivos locais | Depende de internet, sem customização |
| **Apenas local** | Simples, offline | Usuário precisa manter, sem atualizações |
| **Banco de dados (SQLite)** | Consultas complexas, relacionamentos | Overkill para catálogo de modelos |

**Consequências:**
- Positivas: Catálogo atualizado via npm; usuário pode adicionar modelos privados; aliases permitem atalhos
- Negativas: Lógica de merge necessária; conflitos entre catálogos requerem resolução

---

### ADR-007: Fallback Automático vs Manual

| Campo | Detalhe |
|-------|---------|
| **ID** | ADR-007 |
| **Título** | Fallback automático com confirmação do usuário |
| **Status** | Aceito |
| **Data** | 2026-07-09 |

**Contexto:**
Quando um download de modelo falha, ou um modelo não está disponível, o CodeMin precisa decidir se faz fallback automaticamente para um modelo alternativo ou se pergunta ao usuário.

**Decisão:**
**Fallback automático com confirmação.** Quando o modelo primário falha, o CodeMin tenta alternativas automaticamente e pergunta ao usuário antes de prosseguir.

**Opções Consideradas:**

| Opção | Prós | Contras |
|-------|------|---------|
| **Automático com confirmação (escolhida)** | Experiência fluida, usuário tem controle | Interrupção para confirmação |
| **Automático silencioso** | Experiência mais fluida, sem atrito | Usuário pode não saber o que foi instalado |
| **Manual sempre** | Controle total | Frustrante, especialmente para non-tech users |
| **Apenas erro** | Simples, sem surpresas | Usuário precisa resolver sozinho |

**Consequências:**
- Positivas: Resiliência a falhas de download; usuário sempre sabe o que está sendo instalado; reduz suporte
- Negativas: Requer implementação de árvore de fallback; usuário precisa estar presente para confirmar

---

## 10. Considerações de Performance

### 10.1 Otimizações para CPU

| Otimização | Recomendação | Impacto | Observação |
|-----------|-------------|---------|------------|
| **AVX2** | Obrigatório para Q4_K_M | 2-3x mais rápido que sem AVX | Detectado via `platform.ts` (CPUID) |
| **AVX512** | Opcional | +20-40% sobre AVX2 | Apenas em CPUs Intel recentes (Xeon, Core 12th+) |
| **SSSE3** | Mínimo necessário | 1.5x sobre SSE2 | Presente em CPUs de 2008+ |
| **num_thread** | `auto` = cores físicos | 70-90% eficiência com 8 threads | Não usar hyperthreading (pode reduzir performance) |
| **batch_size** | 512 (chat), 128 (autocomplete) | +15-25% throughput | Batch muito grande aumenta latência |
| **mmap** | `true` sempre | -30% RAM, suporta swap | Carrega modelo sob demanda |
| **mlock** | `false` (padrão) | Previne swapping do modelo | Habilite se tiver RAM >= 16GB livre |
| **flash_attention** | `true` se suportado | -50% RAM para contexto longo | Disponível no llama.cpp b41+ |
| **GPU offloading** | `false` (CPU-only) | N/A | CodeMin é focado em CPU |

### 10.2 Parâmetros do llama.cpp

| Parâmetro | Valor Padrão | Chat | Autocomplete | Descrição |
|-----------|-------------|------|--------------|-----------|
| `--threads` | auto | 8-16 | 8-16 | Número de threads CPU |
| `--batch-size` | 512 | 512 | 128 | Tamanho do batch para prompt processing |
| `--n-predict` | -1 (unlimited) | 4096 | 64 | Máximo de tokens a gerar |
| `--ctx-size` | 4096 | 8192 | 2048 | Tamanho do contexto |
| `--temp` | 0.8 | 0.7 | 0.2 | Temperatura de sampling |
| `--top-k` | 40 | 40 | 20 | Top-K sampling |
| `--top-p` | 0.95 | 0.9 | 0.9 | Nucleus sampling |
| `--repeat-penalty` | 1.1 | 1.1 | 1.0 | Penalidade de repetição |
| `--mlock` | false | false | false | Lock do modelo em RAM |
| `--no-mmap` | false | false | false | Usar mmap (não desabilitar) |
| `--no-kv-offload` | true (CPU) | true | true | Manter KV cache na CPU |

### 10.3 Gerenciamento de RAM

| Cenário | RAM Total | RAM Disponível para LLM | Config Recomendada |
|---------|-----------|------------------------|-------------------|
| **Mínimo (Qwen 1.5B)** | 4 GB | 2 GB | `ctx-size: 2048`, `batch-size: 256`, Q4_K_M |
| **Recomendado (Qwen 7B)** | 8-16 GB | 6-12 GB | `ctx-size: 4096`, `batch-size: 512`, Q4_K_M |
| **Confortável (Qwen 7B + chat)** | 16-32 GB | 12-24 GB | `ctx-size: 8192`, `batch-size: 512`, Q4_K_M |
| **Máximo (Qwen 7B + contexto longo)** | 32+ GB | 24+ GB | `ctx-size: 32768`, `batch-size: 1024`, Q4_K_M, `mlock: true` |

### 10.4 Context Window Sizing

```
                    CONTEXT WINDOW SIZING

Para Qwen2.5-Coder 7B Q4_K_M (max suportado: 32768 tokens)

            +--------------------------------------------------------------+
            |                                                              |
  32768     |░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░|
            |                                                              |
            |  Contexto Maximo (32768)                                     |
            |  Uso: analise de codigo completo, refatoracao                |
            |  RAM: ~4.8 GB (modelo) + ~50 MB (KV cache)                  |
            |                                                              |
  16384     |████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░|
            |                                                              |
            |  Contexto Grande (16384)                                     |
            |  Uso: chat com historico extenso, revisao de PR              |
            |  RAM: ~4.8 GB (modelo) + ~25 MB (KV cache)                  |
            |                                                              |
   8192     |████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░|
            |                                                              |
            |  Contexto Chat Padrao (8192)                                |
            |  Uso: sessoes de chat interativo                            |
            |  RAM: ~4.8 GB (modelo) + ~12 MB (KV cache)                  |
            |                                                              |
   4096     |████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░|
            |                                                              |
            |  Contexto Padrao (4096)                                     |
            |  Uso: tarefas gerais, perguntas rapidas                      |
            |  RAM: ~4.8 GB (modelo) + ~6 MB (KV cache)                   |
            |                                                              |
   2048     |████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░|
            |                                                              |
            |  Contexto Autocomplete (2048)                                |
            |  Uso: FIM, autocomplete de codigo                           |
            |  RAM: ~4.8 GB (modelo) + ~3 MB (KV cache)                   |
            |  Latencia alvo: < 500ms primeiro token                      |
            |                                                              |
      0     +--------------------------------------------------------------+
```

### 10.5 Streaming vs Batch

| Aspecto | Streaming | Batch |
|---------|-----------|-------|
| **TTFT (Time to First Token)** | 300-800ms (Qwen 7B) | 2-10s (processa tudo antes) |
| **Experiência do usuário** | Melhor (percepção de velocidade) | Pior (espera total) |
| **Uso de RAM** | Menor (tokens descartados) | Maior (mantém logits) |
| **Throughput** | Menor (overhead por token) | Maior (paralelização) |
| **Caso de uso** | Chat interativo | Autocomplete (FIM) |
| **Implementação** | SSE (Server-Sent Events) | HTTP response única |
| **Consumo CPU** | Distribuído no tempo | Pico no final |

---

## 11. Considerações de Segurança

### 11.1 Sandboxing (execFile vs exec)

O CodeMin **sempre** usa `execFile` em vez de `exec` para executar binários externos. Isso previne injeção de comandos via shell:

```typescript
// SEGURO: execFile nao passa por shell
import { execFile } from 'node:child_process';

// Correto: argumentos como array, sem shell intermediario
execFile('/usr/bin/ollama', ['pull', 'qwen2.5-coder:7b'], (err, stdout) => {
  // processa saida
});

// ERRADO: exec passa por shell, vulneravel a injecao
// exec('ollama pull ' + modelName, ...) // NUNCA FAZER
```

### 11.2 Sanitização de Prompts

Toda entrada do usuário em prompts é sanitizada antes de enviar ao modelo:

```typescript
export function sanitizePrompt(input: string): string {
  return input
    .replace(/[\0\n\r]/g, ' ')   // Remove null bytes e quebras de linha
    .trim();
}
```

### 11.3 Zero Telemetry Policy

O CodeMin **nunca** coleta ou envia dados de telemetria:

- `telemetry.enabled` é `false` por padrão e imutável sem consentimento explícito
- Nenhum tracking pixel, analytics, crash reporter ou fingerprinting
- Nenhuma requisição externa além de:
  - Downloads de modelos (autorizados pelo usuário)
  - API do Ollama (localhost apenas)
- Política verificável em `src/config/defaults.yaml`

### 11.4 Checksum Validation

Todo download de modelo é verificado com SHA256 antes do uso. Arquivos com checksum inválido são deletados automaticamente.

### 11.5 Security Checklist

| Item | Status | Implementação |
|------|--------|---------------|
| Injeção de comando | Mitigado | `execFile` sempre, nunca `exec` |
| Path traversal | Mitigado | Validação de caminhos com `path.resolve` + `path.normalize` |
| SSRF | Mitigado | URLs de download validadas contra allowlist |
| DOS (contexto infinito) | Mitigado | `max_tokens` e `ctx-size` com limites máximos |
| Vazamento de dados | Mitigado | Zero telemetry, apenas localhost |
| Dependency supply chain | Mitigado | Dependências mínimas, lockfile, CI/CD audit |
| Prompt injection | Parcial | Sanitização de entrada, limite de tokens |
| Binário não confiável | Mitigado | SHA256 checksum obrigatório |

---

## 12. Testes e Qualidade

### 12.1 Pirâmide de Testes

```
                    /\
                   /  \
                  /    \
                 / E2E  \           <- 5% (tests end-to-end reais)
                /--------\
               /          \
              / Integração \        <- 15% (nock + mocks HTTP)
             /--------------\
            /                \
           /   Unit Tests     \    <- 80% (testes isolados)
          /--------------------\
```

### 12.2 Testes Unitários

| Componente | Arquivo | O que testar | Casos |
|-----------|---------|-------------|-------|
| `Checksum` | `checksum.spec.ts` | SHA256 verify, computeHash, verifyOrDelete | match, mismatch, arquivo inexistente |
| `Downloader` | `downloader.spec.ts` | Download completo, parcial, retry, timeout | sucesso, resume, falha 3x, timeout |
| `OllamaClient` | `ollama-client.spec.ts` | Ping, listModels, generate, chat, error handling | 200, 404, 500, connection refused |
| `OpenAICompat` | `openai-compat.spec.ts` | Mapeamento de parâmetros, FIM detection | chat, completions, FIM, edge cases |
| `Platform` | `platform.spec.ts` | SO detection, CPU features, RAM parsing | Windows, Linux, macOS, ARM, x64 |
| `InstallEngine` | `install-engine.spec.ts` | Fluxo completo, fases individuais | install, verify fail, prerequisites fail |
| `ModelManager` | `model-manager.spec.ts` | Resolve, list, cache, catalog merge | catalogo vazio, modelo não encontrado, alias |
| `OllamaManager` | `ollama-manager.spec.ts` | isInstalled, isRunning, start, stop | installed, not installed, timeout |
| `ChatEngine` | `chat-engine.spec.ts` | Start session, streaming, error recovery | normal flow, stream error, abort |
| `Bridge` | `bridge/index.spec.ts` | Endpoints, parameter mapping, SSE format | chat, completions, models, errors |

**Exemplo de teste unitário (Checksum):**

```typescript
// tests/unit/utils/checksum.spec.ts
import { describe, it, expect } from 'vitest';
import { Checksum } from '../../../src/utils/checksum.js';

describe('Checksum', () => {
  const checksum = new Checksum();

  it('deve verificar SHA256 corretamente', async () => {
    const filePath = './tests/fixtures/test-model.bin';
    const expectedHash = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';
    const result = await checksum.verify(filePath, expectedHash);
    expect(result).toBe(true);
  });

  it('deve rejeitar SHA256 incorreto', async () => {
    const filePath = './tests/fixtures/test-model.bin';
    const wrongHash = '0000000000000000000000000000000000000000000000000000000000000000';
    const result = await checksum.verify(filePath, wrongHash);
    expect(result).toBe(false);
  });

  it('deve lançar erro para arquivo inexistente', async () => {
    await expect(
      checksum.computeHash('./nonexistent.bin')
    ).rejects.toThrow();
  });
});
```

### 12.3 Testes de Integração (com nock)

```typescript
// tests/integration/bridge-api.spec.ts
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import nock from 'nock';

describe('Bridge API Integration', () => {
  beforeAll(() => {
    nock('http://localhost:11434')
      .post('/api/chat')
      .reply(200, (_, body) => {
        const parsed = JSON.parse(body as string);
        return {
          model: parsed.model,
          message: { role: 'assistant', content: 'Hello!' },
          done: true,
        };
      });
  });

  afterAll(() => {
    nock.cleanAll();
  });

  it('deve traduzir OpenAI chat para Ollama chat', async () => {
    const response = await fetch('http://localhost:11435/v1/chat/completions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model: 'qwen2.5-coder:7b',
        messages: [{ role: 'user', content: 'Hi' }],
        max_tokens: 100,
      }),
    });

    const data = await response.json();
    expect(data.choices[0].message.content).toBe('Hello!');
    expect(data.choices[0].finish_reason).toBe('stop');
  });
});
```

### 12.4 Mock do Ollama para Testes

```typescript
// tests/helpers/ollama-mock.ts
import nock from 'nock';

export class OllamaMock {
  private baseUrl = 'http://localhost:11434';

  start(): void {
    nock(this.baseUrl)
      .get('/api/tags')
      .reply(200, {
        models: [
          { name: 'qwen2.5-coder:7b', modified_at: '2026-07-09T00:00:00Z', size: 4700000000 }
        ]
      });

    nock(this.baseUrl)
      .post('/api/chat')
      .reply(200, (_, body) => {
        const parsed = JSON.parse(body as string);
        return {
          model: parsed.model,
          created_at: new Date().toISOString(),
          message: { role: 'assistant', content: 'Token simulado' },
          done: true,
          total_duration: 1000000000,
          load_duration: 50000000,
          prompt_eval_count: 42,
          eval_count: 10,
        };
      });

    nock(this.baseUrl)
      .post('/api/generate')
      .reply(200, (_, body) => {
        const parsed = JSON.parse(body as string);
        return {
          model: parsed.model,
          response: 'return a + b;',
          done: true,
        };
      });
  }

  stop(): void {
    nock.cleanAll();
  }
}
```

### 12.5 Testes de Benchmark

```typescript
// tests/benchmarks/chat-benchmark.ts
import { OllamaClient } from '../../src/utils/ollama-client.js';

interface BenchmarkResult {
  p50: number;
  p95: number;
  p99: number;
  avgLatency: number;
  tokensPerSecond: number;
}

export async function runBenchmark(
  model: string,
  prompts: string[],
  warmupRounds = 3,
  testRounds = 20
): Promise<BenchmarkResult> {
  const client = new OllamaClient();
  const latencies: number[] = [];

  // Warmup
  for (let i = 0; i < warmupRounds; i++) {
    await client.generate({ model, prompt: 'warmup' });
  }

  // Test
  for (let i = 0; i < testRounds; i++) {
    const prompt = prompts[i % prompts.length];
    const start = performance.now();
    const response = await client.generate({ model, prompt });
    const elapsed = performance.now() - start;
    latencies.push(elapsed);
  }

  // Stats
  const sorted = [...latencies].sort((a, b) => a - b);
  return {
    p50: sorted[Math.floor(testRounds * 0.5)],
    p95: sorted[Math.floor(testRounds * 0.95)],
    p99: sorted[Math.floor(testRounds * 0.99)],
    avgLatency: latencies.reduce((a, b) => a + b, 0) / testRounds,
    tokensPerSecond: 1000 / (latencies.reduce((a, b) => a + b, 0) / testRounds),
  };
}
```

### 12.6 Cobertura Mínima

| Tipo | Cobertura Mínima | Meta |
|------|-----------------|------|
| Linhas | 80% | 90% |
| Branches | 75% | 85% |
| Funções | 85% | 95% |
| Statements | 80% | 90% |

Cobertura é verificada no CI via `vitest run --coverage`. Falha abaixo dos mínimos bloqueia o merge.

---

## 13. Roadmap Técnico

### 13.1 MVP (Sprint 1-2)

**Objetivo:** `codemin install` funcional + bridge compatível com OpenAI

- [ ] Setup do projeto TypeScript + Commander.js + Vitest
- [ ] Comando `install` com download de modelo do HuggingFace
- [ ] SHA256 checksum validation
- [ ] OllamaManager: detectar Ollama, iniciar/parar server
- [ ] Bridge API: endpoint `/v1/chat/completions` (stream + non-stream)
- [ ] Bridge API: endpoint `/v1/completions` (com FIM support)
- [ ] Bridge API: endpoint `/v1/models`
- [ ] OpenAI → Ollama parameter mapping
- [ ] Comando `status`: exibir estado do sistema
- [ ] Comando `doctor`: 12 verificações de diagnóstico
- [ ] `config-generator`: gerar `opencode.json` e `continue.config.json`
- [ ] Catálogo com Qwen 7B, Qwen 1.5B, CodeLlama 7B, DeepSeek 6.7B
- [ ] Zero telemetry config
- [ ] CI com lint, typecheck, test, coverage

### 13.2 V2 (Sprint 3-6)

**Objetivo:** Performance, estabilidade, experiência do usuário

- [ ] Download paralelo com retry e resume
- [ ] Download progress bar com ETA
- [ ] Benchmark engine: P50/P95/P99 latência, tokens/s
- [ ] Comando `benchmark`: executar e salvar resultados
- [ ] CPU feature detection (AVX2, SSSE3, etc.)
- [ ] Auto tune: parâmetros otimizados baseados no hardware detectado
- [ ] Comando `config`: get/set/edit configurações
- [ ] Comando `list`: listar modelos instalados
- [ ] Comando `remove`: remover modelo
- [ ] Chat interativo com histórico (sessões)
- [ ] Fallback automático com confirmação
- [ ] Cache de checksums
- [ ] Aliases de modelos
- [ ] Catálogo local extensível
- [ ] Testes de integração com nock
- [ ] Testes E2E (com Ollama real, opcional)

### 13.3 V3 (Sprint 7-12)

**Objetivo:** Maturação, comunidade, ecossistema

- [ ] Suporte a embeddings (`/v1/embeddings`)
- [ ] Modo servidor dedicado (`codemin bridge --daemon`)
- [ ] Plugin para Continue.dev (instalação automatizada)
- [ ] Plugin para OpenCode (instalação automatizada)
- [ ] VS Code extension manager integration
- [ ] Ollama auto-install (download do binário)
- [ ] Modo offline completo
- [ ] Catálogo remoto com auto-update
- [ ] Modelos customizados (Modelfile)
- [ ] LoRA adapters
- [ ] Multi-session chat
- [ ] Export/import de configurações
- [ ] GitHub Actions para release automatizado
- [ ] Documentação completa em português e inglês
- [ ] Site/documentation (GitHub Pages)

---

## Apêndice A — Glossário Técnico

| Termo | Definição |
|-------|-----------|
| **AIOX** | Artificial Intelligence Orchestration eXperience — framework multi-agente |
| **Bridge** | Camada de tradução entre API OpenAI e API Ollama |
| **CLI** | Command-Line Interface — interface de linha de comando |
| **Continue.dev** | Extensão VS Code para autocomplete e chat com LLMs |
| **FIM** | Fill-In-Middle — técnica de preenchimento de código entre prefixo e sufixo |
| **GGUF** | GPT-Generated Unified Format — formato de modelo do llama.cpp |
| **HuggingFace Hub** | Repositório central de modelos de ML |
| **KV Cache** | Key-Value cache — cache de atenção para eficiência em geração sequencial |
| **llama.cpp** | Implementação em C++ de inferência de LLMs otimizada para CPU |
| **LLM** | Large Language Model — modelo de linguagem de grande escala |
| **mmap** | Memory-mapped file — técnica de carregamento parcial de arquivos |
| **mlock** | Memory lock — impede que páginas de memória sejam swapadas |
| **nock** | Biblioteca Node.js para mocking HTTP em testes |
| **Ollama** | Runtime e gerenciador de LLMs locais |
| **OpenCode** | Ferramenta CLI para engenharia de software assistida por IA |
| **OpenAI API** | Interface REST padronizada para LLMs (chat, completions, embeddings) |
| **Q4_K_M** | Quantização de 4 bits com K-quants mistos (balance quality/size) |
| **SHA256** | Algoritmo de hash criptográfico de 256 bits |
| **SSE** | Server-Sent Events — protocolo de streaming HTTP |
| **TTFT** | Time to First Token — latência até o primeiro token gerado |

---

## Apêndice B — Diagrama de Dependências

```
                        +-----------------------+
                        |     Commander.js      |
                        |  (CLI framework)      |
                        +----------+------------+
                                   |
              +--------------------+-------------------+
              |                    |                   |
              v                    v                   v
    +-------------------+  +------------------+  +------------------+
    |  InstallEngine    |  |  ChatEngine      |  |  HealthChecker   |
    |  (install-flow)   |  |  (chat-session)  |  |  (diagnostics)   |
    +--------+----------+  +--------+---------+  +--------+---------+
             |                      |                    |
             v                      v                    v
    +-------------------+  +------------------+  +------------------+
    |  OllamaManager    |  |  ModelManager    |  |  ConfigGenerator |
    |  (runtime mgmt)   |  |  (catalog)       |  |  (templates)     |
    +--------+----------+  +--------+---------+  +--------+---------+
             |                      |                    |
             v                      v                    v
    +-------------------+  +------------------+  +------------------+
    |  OllamaClient     |  |  Downloader      |  |  Checksum        |
    |  (HTTP + SSE)     |  |  (fetch + retry) |  |  (SHA256)        |
    +--------+----------+  +--------+---------+  +--------+---------+
             |                      |                    |
             v                      v                    v
    +-------------------+  +------------------+  +------------------+
    |  undici (fetch)   |  |  HuggingFace Hub |  |  node:crypto     |
    |  (HTTP client)    |  |  (download src)  |  |  (hash stream)   |
    +-------------------+  +------------------+  +------------------+

Dependencias externas:
  +---------------------------+
  |  Dependencia              |  Usado por
  +---------------------------+
  |  commander                |  cli/index.ts
  |  undici                   |  ollama-client, downloader
  |  @inquirer/prompts        |  chat, install prompts
  |  chalk                    |  logger, doctor output
  |  js-yaml                  |  config-generator
  |  node:child_process       |  ollama-manager
  |  node:crypto              |  checksum
  |  node:fs/promises         |  model-manager, downloader
  |  node:os                  |  platform
  |  node:http                |  bridge server
  +---------------------------+

Dependencias de desenvolvimento:
  +---------------------------+
  |  typescript               |  Compilacao
  |  vitest                   |  Test runner
  |  nock                     |  HTTP mocking
  |  @biomejs/biome           |  Lint + format
  |  tsx                      |  Dev server (watch)
  |  @types/node              |  Tipos Node.js
  +---------------------------+
```

---

## Apêndice C — Máquina de Estados da Instalação

```
                         +---------------------------+
                         |     INITIAL               |
                         | Estado: Usuario executa   |
                         | codemin install <modelo>  |
                         +------------+--------------+
                                      |
                                      v
                         +---------------------------+
                         |     CHECK_PREREQUISITES   |
                         | Verificar:                |
                         | • Node.js >= 18           |
                         | • Ollama instalado        |<----+
                         | • Ollama server rodando   |     |
                         | • Disco disponivel        |     |
                         +------------+--------------+     |
                                      |                    |
                    +-----------------+-----------------+  |
                    | Todas ok        | Falha           |  |
                    v                 v                 |  |
          +-------------------+  +-------------------+  |  |
          | RESOLVE_MODEL     |  | ERROR             |  |  |
          | Buscar no         |  | "Pre-requisitos   |  |  |
          | catalogo:         |  |  nao atendidos"   |  |  |
          | • Catalog.json    |  | Mostrar:          |  |  |
          | • Alias?          |  | • codemin doctor  |  |  |
          | • URL direta?     |  +-------------------+  |  |
          +--------+----------+                         |  |
                   |                                    |  |
      +------------+------------+                       |  |
      | OK                      | NOT FOUND             |  |
      v                         v                       |  |
+-------------------+  +--------------------+           |  |
| CHECK_CACHE       |  | ASK_SOURCE         |           |  |
| Verificar cache   |  | "Modelo nao        |           |  |
| local:            |  | encontrado.        |           |  |
| • Checksum bate?  |  | Tentar pull do     |           |  |
| • Arquivo existe? |  | Ollama ou URL?"    |           |  |
+--------+----------+  +---------+----------+           |  |
         |                       |                      |  |
    +----+----+             +----+----+                 |  |
    | Cache   | No Cache    | Ollama  | URL            |  |
    | OK      |             | Pull    |                |  |
    v        v             v        v                  |  |
+--------+ +--------+  +--------+ +----------------+   |  |
| VERIFY | |DOWNLOAD|  |PULL    | |DOWNLOAD_FROM   |   |  |
| CHECK- | |        |  |OLLAMA  | |URL             |   |  |
| SUM    | |        |  |        | |                |   |  |
+----+---+ +---+----+  +---+----+ +--------+-------+   |  |
     |         |            |               |           |  |
     |    +----+            |               |           |  |
     |    | Retry (max 3)   |               |           |  |
     |    v                 |               |           |  |
     | +------+ NO         |               |           |  |
     +-+VERIFY+------------+---------------+-----------+  |
       | SHA  | YES                                      |
       +--+---+                                          |
          |                                              |
    +-----+-----+                                        |
    | Match      | Mismatch                              |
    v           v                                        |
+--------+ +----------+                                  |
|IMPORT   | | DELETE   |                                  |
|MODEL    | | + RETRY  |                                  |
+----+----+ +----+-----+                                  |
     |           |                                        |
     |      +----+----+                                   |
     |      | Retry   |                                   |
     |      | > 3?    |                                   |
     |      +----+----+                                   |
     |           |                                        |
     |      +----+----+                                   |
     |      | YES     | NO (retry)------------------------+
     |      v         v
     | +--------+ +--------+
     | | CRITICAL| | DOWNLOAD|
     | | ERROR   | | AGAIN   |
     | | "Falha  | +--------+
     | |  SHA256"|
     | +--------+
     |
     v
+-------------------+
| VERIFY_INSTALL    |
| • ollama list     |
| • Modelo aparece? |
+--------+----------+
         |
    +----+----+
    | OK      | FALHA
    v         v
+--------+ +--------+
| GEN    | | LOG    |
| CONFIG | | ERRO   |
| • open- | | "Modelo|
|   code  | |  nao   |
| • conti-| |  encon-|
|   nue   | |  trado"|
| • yaml  | +--------+
+----+----+
     |
     v
+--------+
| DONE   |
| "Modelo|
|  insta-|
|  lado   |
|  com    |
|  su-    |
|  cesso!"|
+--------+

Estados de erro possiveis:
  ERROR_PREREQUISITES -> INITIAL (apos correcao)
  ERROR_DOWNLOAD -> DOWNLOAD (retry ate 3x)
  ERROR_CHECKSUM -> DOWNLOAD (retry ate 3x)
  ERROR_IMPORT -> IMPORT (retry)
  ERROR_VERIFY -> RESOLVE_MODEL (alternativa)
```

---
