# Product Requirements Document — CodeMin

**Produto:** CodeMin — Assistente de Codificação LLM Local (CPU-only)
**Versão do Documento:** 1.0.0
**Autor:** Morgan (AIOX Product Manager)
**Data:** 2026-07-09
**Status:** Draft
**Template:** AIOX PRD v2.0

---

## Sumário Executivo

O CodeMin é um sistema de LLM local (CPU-only), gratuito e open-source para assistência de codificação. Ele integra-se ao OpenCode e ao VS Code via Continue.dev, usando o modelo Qwen2.5-Coder 7B quantizado via Ollama/llama.cpp. Seu objetivo é democratizar o acesso a assistentes de IA para programação, eliminando a dependência de GPU e assinaturas pagas. Com 2 comandos para instalar e latência de autocomplete inferior a 3 segundos, o CodeMin entrega uma experiência comparável a soluções comerciais (GitHub Copilot, Cursor) — mas 100% local, offline e com privacidade total.

---

## 1. Problema e Oportunidade

### 1.1 O Problema

Desenvolvedores individuais, entusiastas open-source e profissionais em mercados emergentes enfrentam uma barreira significativa para adotar assistentes de IA na codificação:

| Barreira | Impacto |
|----------|---------|
| **Custo** | GitHub Copilot custa US$ 10-39/mês; Cursor US$ 20/mês; ChatGPT Pro US$ 20/mês — inviável para muitos devs |
| **GPU Obrigatória** | Modelos locais exigem GPU (pelo menos 6-8 GB VRAM) para rodar com desempenho aceitável |
| **Privacidade** | Soluções em nuvem enviam código para servidores externos — violação de políticas em empresas e projetos sensíveis |
| **Dependência de Internet** | Chat e autocomplete param de funcionar sem conexão |
| **Complexidade** | Soluções self-hosted existentes exigem configuração complexa (Docker, CUDA, Python venvs) |

### 1.2 A Oportunidade

Existem **mais de 30 milhões de desenvolvedores no mundo** sem acesso a GPU. Destes:

- ~70% usam notebooks sem GPU dedicada
- ~25% estão em regiões onde assinaturas em dólar são proibitivas
- ~15% trabalham com código proprietário que não pode sair da máquina

Nenhuma solução atual atende **simultaneamente** aos critérios: gratuito + CPU-only + offline + instalação simples + qualidade aceitável.

**CodeMin preenche esse gap.**

---

## 2. Público-Alvo e Personas

### 2.1 Persona 1: Joana — A Dev Caseira

| Atributo | Descrição |
|----------|-----------|
| **Idade** | 24 anos |
| **Ocupação** | Estudante de Ciência da Computação / Desenvolvedora Júnior |
| **Setup** | Notebook Lenovo IdeaPad 3, Ryzen 5 5500U, 16 GB RAM, sem GPU |
| **Dor** | Quer usar assistente de código mas não pode pagar assinatura. Tentou rodar modelos locais e travou o notebook. |
| **Comportamento** | Usa VS Code, programa em Python/JS, faz projetos da faculdade e open-source |
| **Frustração** | "Gastei 2 dias tentando fazer o CodeLlama rodar. Desisti." |
| **Valor do CodeMin** | Instalação em 2 comandos, autocomplete funcionando, zero configuração |

### 2.2 Persona 2: Carlos — O Profissional Sem Verba

| Atributo | Descrição |
|----------|-----------|
| **Idade** | 35 anos |
| **Ocupação** | Desenvolvedor Full Stack em startup brasileira |
| **Setup** | Desktop i7-10700, 32 GB RAM, sem GPU |
| **Dor** | Empresa não libera verba para Copilot. Código do cliente não pode sair da máquina por contrato. |
| **Comportamento** | Trabalha com Java + Spring no backend, React no frontend |
| **Frustração** | "Preciso de ajuda com código legado mas não posso colar no ChatGPT." |
| **Valor do CodeMin** | Privacidade total, refatoração de código legado, sem custo para a empresa |

### 2.3 Persona 3: Ana — A Entusiasta Open-Source

| Atributo | Descrição |
|----------|-----------|
| **Idade** | 29 anos |
| **Ocupação** | Contribuidora de projetos open-source / DevOps |
| **Setup** | ThinkPad X220 (2012!), 8 GB RAM, Core i5 |
| **Dor** | Máquina antiga, quer ferramentas modernas mas hardware limita. Acredita em software livre. |
| **Comportamento** | Trabalha com Go, Rust, Python. Prefere tudo open-source. |
| **Frustração** | "Se não é open-source e não roda no meu ThinkPad, não serve pra mim." |
| **Valor do CodeMin** | Open-source (MIT), CPU-only, roda em hardware velho, filosofia alinhada |

---

## 3. Visão do Produto

### 3.1 O CodeMin é:

- ✅ Um **assistente de codificação local** que roda inteiramente na máquina do usuário
- ✅ **CPU-only** (GPU acelera, mas não é necessária)
- ✅ **Gratuito** — sem assinatura, sem limites, sem contas
- ✅ **Open-source** (licença MIT) — auditável, extensível, comunitário
- ✅ **Offline-first** — funciona sem internet após o download inicial dos modelos
- ✅ **Plug-and-play** — instalação em 2 comandos

### 3.2 O CodeMin NÃO é:

- ❌ **Um modelo novo** — usamos modelos existentes (Qwen2.5-Coder, CodeLlama, DeepSeek-Coder)
- ❌ **Um IDE** — integra-se a IDEs existentes (VS Code, JetBrains) via Continue.dev
- ❌ **Uma plataforma de deploy** — não hospeda modelos, não gerencia infraestrutura
- ❌ **Um competidor do Copilot** — é uma alternativa gratuita e local com escopo mais modesto
- ❌ **Uma ferramenta para todos os casos** — focado em assistência de código, não conversa geral

### 3.3 Posicionamento

> **"CodeMin: Seu copiloto local, gratuito e privado — sem GPU, sem assinatura, sem desculpas."**

---

## 4. Objetivos e Métricas de Sucesso (OKRs)

### 4.1 OKR Trimestre 1 (MVP + V2)

| Objetivo | Key Result | Métrica | Meta |
|----------|-----------|---------|------|
| **O1: Estabelecer base de usuários ativos** | KR1: Downloads do CLI | Contagem de downloads npm/github | 2.000 downloads no primeiro mês |
| | KR2: Usuários ativos semanais | Heartbeat de uso (comandos executados) | 500 usuários ativos |
| | KR3: Issues/PRs da comunidade | Contagem no GitHub | 20+ contribuições |
| **O2: Entregar experiência de qualidade mínima viável** | KR1: Latência de autocomplete | P95 da latência em hardware-alvo | < 3s |
| | KR2: Latência de chat | P95 da latência | < 15s |
| | KR3: Taxa de erro de instalação | Falhas em `codemin install` | < 10% |
| **O3: Garantir satisfação do usuário** | KR1: NPS da ferramenta | Survey pós-uso | > 30 |
| | KR2: Taxa de retenção D7 | Usaram nos dias 0 e 7 | > 40% |
| | KR3: Issues críticas abertas | Contagem no GitHub | < 5 |

### 4.2 OKR Trimestre 2 (V3)

| Objetivo | Key Result | Meta |
|----------|-----------|------|
| **O4: Expandir capacidade do sistema** | Suporte a 8+ linguagens | 8 linguagens com qualidade aceitável |
| | Autocomplete por contexto | Precisão de sugestões > 60% |
| | Tempo de resposta < 1.5s autocomplete | P95 < 1.5s |
| **O5: Construir ecossistema** | Plugin VS Code publicado na marketplace | 1.000 instalações |
| | Documentação completa em PT/EN | 95% de cobertura |
| | 5+ contribuidores externos ativos | Média de 5 PRs/semana |

---

## 5. Funcionalidades por Fase (MoSCoW)

### 5.1 MVP (Sprint 1-2)

| Categoria | Funcionalidade | ID | Prioridade |
|-----------|---------------|----|------------|
| **Must Have** | Instalação via CLI (`codemin install`) | FR-MVP-01 | M |
| **Must Have** | Download automático do modelo Qwen2.5-Coder 7B | FR-MVP-02 | M |
| **Must Have** | Chat contextual sobre código (OpenCode) | FR-MVP-03 | M |
| **Must Have** | Config Generator → `opencode.json` | FR-MVP-04 | M |
| **Must Have** | Config Generator → `continue.config.json` | FR-MVP-05 | M |
| **Must Have** | Comando `codemin status` (verificação de saúde) | FR-MVP-06 | M |
| **Should Have** | Comando `codemin chat` (CLI nativa) | FR-MVP-07 | S |
| **Should Have** | Geração de código a partir de linguagem natural | FR-MVP-08 | S |
| **Could Have** | Explicação de código selecionado | FR-MVP-09 | C |
| **Won't Have** | Autocomplete inline (FIM) via Continue.dev no MVP | FR-MVP-10 | W |

### 5.2 V2 (Sprint 3-6)

| Categoria | Funcionalidade | ID | Prioridade |
|-----------|---------------|----|------------|
| **Must Have** | Autocomplete inline (FIM) via Continue.dev | FR-V2-01 | M |
| **Must Have** | Refatoração de código | FR-V2-02 | M |
| **Must Have** | Suporte a múltiplas linguagens (Python, JS/TS, Java) | FR-V2-03 | M |
| **Must Have** | Detecção de bugs em código selecionado | FR-V2-04 | M |
| **Should Have** | Code review (segurança + boas práticas) | FR-V2-05 | S |
| **Should Have** | Geração de testes unitários | FR-V2-06 | S |
| **Should Have** | Documentação automática de funções/métodos | FR-V2-07 | S |
| **Could Have** | Comando `codemin update` (atualização de modelos) | FR-V2-08 | C |
| **Could Have** | Fallback automático para modelo menor (Qwen2.5-Coder 1.5B) | FR-V2-09 | C |
| **Won't Have** | Suporte a JetBrains | FR-V2-10 | W |

### 5.3 V3 (Sprint 7-12)

| Categoria | Funcionalidade | ID | Prioridade |
|-----------|---------------|----|------------|
| **Must Have** | Suporte total a 8 linguagens (+Go, Rust, C++, C#, Ruby) | FR-V3-01 | M |
| **Must Have** | Múltiplos modelos selecionáveis (Qwen, CodeLlama, DeepSeek) | FR-V3-02 | M |
| **Must Have** | Gerenciamento de modelos (lista, download, remover) | FR-V3-03 | M |
| **Should Have** | Autocomplete multi-arquivo (context-aware) | FR-V3-04 | S |
| **Should Have** | Plugin VS Code oficial (abstraindo Continue.dev) | FR-V3-05 | S |
| **Should Have** | Suporte a JetBrains (via Continue.dev) | FR-V3-06 | S |
| **Could Have** | CodeMin GUI (interface web local) | FR-V3-07 | C |
| **Could Have** | Fine-tuning de modelos para codebase específica | FR-V3-08 | C |
| **Could Have** | Comandos de benchmarking (`codemin bench`) | FR-V3-09 | C |
| **Won't Have** | Suporte a GPU CUDA otimizado (aceleração adicional) | FR-V3-10 | W |

> **Legenda MoSCoW:** M = Must Have (essencial para a release), S = Should Have (importante mas não crítico), C = Could Have (desejável), W = Won't Have (explicitamente fora do escopo)

---

## 6. Requisitos Funcionais Detalhados

### 6.1 MVP

| ID | Nome | Descrição | Prioridade | Critério de Aceitação |
|----|------|-----------|------------|----------------------|
| FR-MVP-01 | CLI de Instalação | Comando `codemin install` que instala todas as dependências e faz download do modelo | M | `codemin install` completa em < 5 min em banda larga; cria diretório `~/.codemin/`; logs de progresso visíveis |
| FR-MVP-02 | Download de Modelo | Download automático do Qwen2.5-Coder 7B Q4_K_M (~4.7 GB) do HuggingFace | M | Download com barra de progresso; resume se interrompido; valida checksum SHA256; armazena em `~/.codemin/models/` |
| FR-MVP-03 | Chat Contextual | Interface de chat via OpenCode que entende o contexto do código aberto no editor | M | Enviar pergunta sobre código selecionado → receber resposta em < 15s; resposta relevante ao contexto enviado |
| FR-MVP-04 | Config OpenCode | Gerar `opencode.json` com provider configurado para Ollama (compatível API OpenAI) | M | Arquivo gerado com endpoint `http://localhost:11434/v1`, model `qwen2.5-coder:7b`, parâmetros otimizados; instruções de onde colocar o arquivo |
| FR-MVP-05 | Config Continue.dev | Gerar `config.json` para Continue.dev com modelo local | M | Arquivo gerado com `models` apontando para Ollama; `tabAutocompleteModel` configurado; documentado no README |
| FR-MVP-06 | Status Check | Comando `codemin status` que verifica integridade da instalação | M | Verifica: Ollama rodando, modelo baixado, configurações existentes; saída legível com ✓/✗ |
| FR-MVP-07 | Chat Nativo | Comando `codemin chat` para conversar sem IDE | S | Interface de chat no terminal; mantém histórico por sessão; responde em < 15s |
| FR-MVP-08 | Geração de Código | A partir de descrição em linguagem natural, gerar código no contexto do projeto | S | "Crie uma função que valida CPF em Python" → retorna código Python funcional; respeita linguagem do arquivo atual |
| FR-MVP-09 | Explicação de Código | Selecionar trecho de código e pedir explicação | C | Explicação em linguagem natural do código selecionado; identifica padrões, algoritmos, complexidade |
| FR-MVP-10 | Autocomplete FIM | ~ (adiado para V2) | W | N/A |

### 6.2 V2

| ID | Nome | Descrição | Prioridade | Critério de Aceitação |
|----|------|-----------|------------|----------------------|
| FR-V2-01 | Autocomplete FIM | Autocomplete inline (Fill-in-the-Middle) via Continue.dev no VS Code | M | Sugestões aparecem enquanto digita; latência < 3s (P95); precisa de < 5s para exibir primeira sugestão |
| FR-V2-02 | Refatoração | Selecionar código + instrução de refatoração → código refatorado | M | "Extraia esse trecho para uma função separada" → executa corretamente; código gerado é sintaticamente válido |
| FR-V2-03 | Multi-linguagem | Suporte a Python, JavaScript, TypeScript, Java | M | Geração e chat funcionam nas 4 linguagens; acurácia sintática > 90% |
| FR-V2-04 | Detecção de Bugs | Analisar código selecionado e apontar possíveis bugs | M | Detecta: null pointer, out-of-bounds, variáveis não usadas, tipos incompatíveis; reporta com linha e explicação |
| FR-V2-05 | Code Review | Análise de segurança e boas práticas do código | S | Identifica: SQL injection, XSS, hardcoded secrets, falta de validação; classifica por severidade |
| FR-V2-06 | Geração de Testes | Gerar testes unitários a partir de função/método selecionado | S | Gera testes no framework da linguagem (pytest, Jest, JUnit); 80%+ de cobertura da função alvo |
| FR-V2-07 | Documentação | Gerar docstring/comentários JSDoc para código selecionado | S | Gera documentação seguindo convenção da linguagem (PEP 257, JSDoc, JavaDoc) |
| FR-V2-08 | Atualização | `codemin update` para atualizar CodeMin e modelos | C | Verifica versão online; faz download incremental se possível; rollback em caso de falha |
| FR-V2-09 | Fallback | Se 7B consumir muita RAM, usar automática ou manualmente modelo 1.5B | C | Detecção automática de RAM < 12 GB → sugerir fallback; `codemin use codermodel --small` |

### 6.3 V3

| ID | Nome | Descrição | Prioridade | Critério de Aceitação |
|----|------|-----------|------------|----------------------|
| FR-V3-01 | 8+ Linguagens | Suporte expandido: Go, Rust, C++, C#, Ruby (além das 4 do V2) | M | Geração funcional e testes sintáticos em todas; acurácia > 85% |
| FR-V3-02 | Múltiplos Modelos | Seleção entre Qwen, CodeLlama, DeepSeek-Coder, Mistral | M | `codemin list-models`, `codemin use <modelo>`; troca sem reinstalar; benchmark comparativo integrado |
| FR-V3-03 | Gerenciamento | Listar, baixar, remover, alternar entre modelos | M | `codemin model list`, `model download`, `model remove`; semântica de gerenciamento completa |
| FR-V3-04 | Multi-arquivo | Autocomplete que considera múltiplos arquivos abertos | S | Sugestões baseadas em imports, símbolos e tipos de outros arquivos do projeto |
| FR-V3-05 | Plugin VS Code | Extensão VS Code oficial (interface própria, abstrai Continue.dev) | S | Comandos `CodeMin: Chat`, `CodeMin: Explicar`, `CodeMin: Revisar`; painel lateral; instalação 1 clique |
| FR-V3-06 | JetBrains | Suporte a IDEs JetBrains via Continue.dev | S | Autocomplete e chat funcionando em IntelliJ, PyCharm, GoLand |
| FR-V3-07 | GUI Web | Interface web local (`http://localhost:8080`) com chat e configurações | C | Interface responsiva; gerenciamento visual de modelos; chat com histórico |
| FR-V3-08 | Fine-tuning | Scripts para fine-tuning do modelo base na codebase do usuário | C | `codemin finetune` preprocessa repositório gital; gera LoRA adapters; funde ao modelo base |
| FR-V3-09 | Benchmark | `codemin bench` para medir desempenho | C | Mede: tokens/s, latência, uso de RAM/CPU; compara entre modelos; gera relatório |
| FR-V3-10 | GPU Bonus | Aceleração CUDA/Metal para GPUs disponíveis (opcional) | W | Detecta GPU disponível; usa llama.cpp com aceleração; mantém compatibilidade CPU |

---

## 7. Requisitos Não Funcionais

| ID | Categoria | Requisito | Métrica / Especificação |
|----|-----------|-----------|------------------------|
| NFR-01 | Privacidade | Zero dados saem da máquina do usuário | Inspeção de rede Wireshark: 0 pacotes para internet durante uso; verificação em CI/CD de ausência de telemetria |
| NFR-02 | Hardware | CPU-only (sem GPU necessária) | Funciona em CPU sem GPU presente; GPU é bônus, não requisito |
| NFR-03 | RAM | Mínima: 8 GB; Recomendada: 16 GB+ | Modelo 7B Q4_K_M usa ~4.7 GB RAM; sistema + Ollama < 6 GB total com 7B; ~3 GB com 1.5B |
| NFR-04 | Latência Autocomplete | Resposta do modelo em < 3s (P95) | Medido em hardware-alvo (Ryzen 5, 16 GB DDR4, NVMe SSD); FIM (Fill-in-the-Middle) |
| NFR-05 | Latência Chat | Resposta do modelo em < 15s (P95) | Medido em hardware-alvo; streaming progressivo (token a token) |
| NFR-06 | Throughput | Mínimo de 5 tokens/s em hardware-alvo | Benchmark com prompt de 500 tokens + geração de 200 tokens |
| NFR-07 | Licenciamento | Open-source, permissiva | MIT ou Apache 2.0; sem cláusulas restritivas |
| NFR-08 | Custo | Zero para o usuário final | Sem assinatura, sem microtransações, sem conta, sem limites |
| NFR-09 | Offline | Funciona sem internet | Apenas download inicial requer internet; cache de modelos local; fallback gracioso |
| NFR-10 | Instalação | Simples — 1 a 2 comandos | `npm install -g codemin` ou `curl | bash`; instalação completa < 10 min |
| NFR-11 | Tamanho | Download razoável | Modelo principal: ~4.7 GB (GGUF Q4_K_M); ferramenta: < 50 MB |
| NFR-12 | Armazenamento | Gerenciamento de espaço | `codemin clean` limpa caches; aviso quando disco < 10 GB livre |
| NFR-13 | Sistema Operacional | Cross-platform | Windows (10+), macOS (12+), Linux (Ubuntu 20.04+, Fedora 36+) |
| NFR-14 | Arquitetura | x86_64 (amd64); ARM64 experimental | Suporte primário x86_64; ARM64 (Apple Silicon, Raspberry Pi 5) como experimental |
| NFR-15 | Segurança | Sandbox de execução | Código gerado NÃO é executado automaticamente; CLI não executa comandos sem confirmação explícita |
| NFR-16 | Manutenibilidade | Código modular e testado | Cobertura de testes > 80%; linting; CI/CD automatizado |

---

## 8. Integrações

### 8.1 OpenCode (Provider Compatível com OpenAI)

O OpenCode suporta providers compatíveis com a API da OpenAI. O CodeMin expõe o Ollama como um endpoint que imita essa API.

**Configuração gerada (`opencode.json`):**

```json
{
  "provider": "openai",
  "model": "qwen2.5-coder:7b",
  "apiBase": "http://localhost:11434/v1",
  "apiKey": "codemin-local",
  "stream": true,
  "maxTokens": 2048,
  "temperature": 0.2,
  "contextLength": 8192
}
```

**Funcionalidades integradas:**
- Chat contextual sobre código
- Geração de código com system prompt otimizado para codificação
- Refatoração e explicação via chat

### 8.2 Continue.dev (Autocomplete no VS Code)

O Continue.dev é uma extensão open-source para VS Code e JetBrains que suporta modelos locais.

**Configuração gerada (`~/.continue/config.json`):**

```json
{
  "models": [
    {
      "title": "CodeMin Qwen 7B",
      "provider": "ollama",
      "model": "qwen2.5-coder:7b",
      "contextLength": 8192
    }
  ],
  "tabAutocompleteModel": {
    "title": "CodeMin Autocomplete",
    "provider": "ollama",
    "model": "qwen2.5-coder:7b"
  },
  "systemMessage": "Você é CodeMin, um assistente de codificação especializado..."
}
```

**Funcionalidades integradas:**
- Autocomplete inline (FIM) — V2+
- Chat no editor
- Ações de código (refatorar, explicar, revisar)

### 8.3 Ollama (Gerenciamento de Modelos)

Ollama é o runtime de modelos locais. O CodeMin gerencia o Ollama indiretamente.

**Responsabilidades:**
- Iniciar/parar servidor Ollama (`ollama serve`)
- Baixar modelos via `ollama pull qwen2.5-coder:7b`
- Otimizar parâmetros para CPU (n_ctx, num_thread)
- Health check do servidor

### 8.4 VS Code + JetBrains

| IDE | Funcionalidade | Via | Status |
|-----|---------------|-----|--------|
| VS Code | Chat contextual | Continue.dev | MVP |
| VS Code | Autocomplete inline | Continue.dev | V2 |
| VS Code | Ações de código | Continue.dev + OpenCode | V2 |
| VS Code | Plugin oficial | Extensão própria | V3 |
| JetBrains | Chat + Autocomplete | Continue.dev | V3 |

### 8.5 Diagrama de Integração

```
┌─────────────────────────────────────────────────────┐
│                   MÁQUINA LOCAL                      │
│                                                      │
│  ┌──────────┐    ┌──────────────┐    ┌───────────┐  │
│  │ OpenCode │◄──►│  CodeMin     │◄──►│  Ollama   │  │
│  │  (agente)│    │  CLI Manager │    │ (runtime) │  │
│  └──────────┘    └──────┬───────┘    └─────┬─────┘  │
│                         │                  │         │
│  ┌──────────┐           │         ┌────────▼──────┐  │
│  │ Continue │◄──────────┘         │ llama.cpp     │  │
│  │  .dev    │                    │ (CPU backend) │  │
│  └────┬─────┘                    └───────┬────────┘  │
│       │                                  │           │
│  ┌────▼─────┐                   ┌────────▼────────┐  │
│  │ VS Code  │                   │ Qwen2.5-Coder 7B│  │
│  │ (Editor) │                   │ (GGUF Q4_K_M)   │  │
│  └──────────┘                   └─────────────────┘  │
│                                                      │
│  LEGENDA: ──► Dados          ─ ─ ► Configuração      │
└─────────────────────────────────────────────────────┘
```

---

## 9. Stack Tecnológica Recomendada

| Componente | Tecnologia | Versão | Justificativa |
|-----------|-----------|--------|---------------|
| **Modelo Base** | Qwen2.5-Coder Instruct | 7B (Q4_K_M) | Melhor custo-benefício CPU; estado-da-arte em code LLMs na categoria sub-10B |
| **Runtime LLM** | Ollama + llama.cpp | Ollama 0.5+ | Gerenciamento simplificado; GGUF nativo; aceleração CPU via BLAS |
| **Formato de Modelo** | GGUF | Q4_K_M | Quantização 4-bit com mínima perda de qualidade (~4.7 GB); carregamento eficiente em RAM |
| **CLI** | Node.js (TypeScript) | Node 18+ LTS | Cross-platform nativo; npm para distribuição; ecossistema maduro |
| **Interface OpenCode** | OpenAI-compatible API | - | OpenCode já suporta; endpoint local sem auth |
| **IDE Bridge** | Continue.dev | VS Code extension | Open-source; suporte FIM; multi-IDE |
| **Config Management** | JSON/YAML | - | Formatos universais; sem banco de dados |
| **Instalação** | npm + shell script | - | `npm i -g codemin` + script pós-instalação |
| **Testes** | Vitest (CLI) | - | Rápido, TypeScript nativo |
| **CI/CD** | GitHub Actions | - | Grátis para open-source |

### 9.1 Alternativas Consideradas (e Por Que Foram Rejeitadas)

| Alternativa | Motivo da Rejeição |
|-------------|-------------------|
| **GPT4All** | API menos flexível; sem suporte FIM nativo; integração mais complexa com OpenCode |
| **LocalAI** | Mais pesado; curva de aprendizado maior; overhead Docker |
| **llama.cpp puro** | Sem gerenciamento de modelos; usuário precisa baixar e configurar manualmente |
| **CodeLlama 7B via Ollama** | Desempenho inferior ao Qwen2.5-Coder em benchmarks de código (HumanEval+) |
| **DeepSeek-Coder 6.7B** | Ótimo modelo, mas sem suporte Ollama tão maduro quanto Qwen; consumo de RAM similar |
| **Stable Code 3B** | Muito pequeno para qualidade aceitável em tarefas complexas; útil como fallback |

---

## 10. UX / Experiência do Usuário

### 10.1 Fluxo de Instalação Ideal

```
INSTALAÇÃO (2 comandos, < 5 minutos)
═══════════════════════════════════════

Terminal:
  $ npm install -g codemin
  ✓ CodeMin CLI instalado

  $ codemin install
  ╔═══════════════════════════════════════╗
  ║  🚀 CodeMin - Instalação             ║
  ║                                       ║
  ║  1/4 Verificando sistema...           ║  ✓ 
  ║  2/4 Instalando Ollama...             ║  ✓
  ║  3/4 Baixando modelo (4.7 GB)...      ║  ████████░░ 72%
  ║  4/4 Configurando OpenCode...         ║  ...
  ║                                       ║
  ║  ✅ Instalação completa!              ║
  ║                                       ║
  ║  Próximos passos:                     ║
  ║  • codemin chat    → Chat no terminal ║
  ║  • codemin status  → Verificar saúde  ║
  ║  • codemin doctor  → Diagnosticar     ║
  ║  • codemin config  → Ver configs      ║
  ╚═══════════════════════════════════════╝
```

### 10.2 Jornada do Usuário (3 Cenários)

#### Cenário A: Chat no OpenCode (MVP)

```
1. Usuário executa `codemin install`
2. Usuário coloca `opencode.json` na raiz do projeto (instruído pelo CLI)
3. Abre OpenCode no VS Code
4. Seleciona código → `Ctrl+I` → "Explique este código"
5. CodeMin (via Ollama) processa → resposta aparece em < 15s
6. ✅ Sucesso!
```

#### Cenário B: Autocomplete (V2)

```
1. Usuário abre VS Code
2. Continue.dev detecta modelo local CodeMin
3. Usuário começa a digitar `function valida` 
4. CodeMin sugere continuação em < 3s
5. Usuário aceita com `Tab`
6. ✅ Sucesso! Latência imperceptível
```

#### Cenário C: Refatoração + Testes (V2)

```
1. Usuário seleciona função enorme no VS Code
2. Abre chat do Continue.dev
3. Digita: "Extraia a lógica de validação de e-mail para uma função separada"
4. CodeMin analisa contexto e gera código refatorado
5. Usuário revisa e aplica
6. Seleciona função original → "Gere testes unitários para esta função"
7. CodeMin gera pytest/Jest com casos de borda
8. ✅ Sucesso!
```

### 10.3 Wireframe Conceitual (ASCII)

```
┌──────────────────────────────────────────────────────────┐
│  VS Code  (com CodeMin + Continue.dev)                    │
│ ┌─────────┬────────────────────────────────────────────┐ │
│ │ Explorer│  arquivo.py                                  │ │
│ │         │  ┌────────────────────────────────────────┐ │ │
│ │ src/    │  │ def process_data(items):               │ │ │
│ │  main   │  │     result = []                        │ │ │
│ │   .py   │  │     for i in range(len(items)):        │ │ │
│ │ tests/  │  │         item = items[i]               │ │ │
│ │         │  │         if item['status'] == 'ok':     │ │ │
│ │         │  │             result.append(item)        │ │ │
│ │         │  │     return result                      │ │ │
│ │         │  │                                         │ │ │
│ │         │  │  [Selecionado] → Ctrl+I                 │ │ │
│ │         │  └────────────────────────────────────────┘ │ │
│ │         ├────────────────────────────────────────────┤ │
│ │         │  ◉ CodeMin Chat                             │ │
│ │         │  ─────────────────────                      │ │
│ │         │  > Refatore usando list comprehension        │ │
│ │         │                                             │ │
│ │         │  ├ CodeMin:                           40s │ │
│ │         │  │ def process_data(items):                │ │ │
│ │         │  │     return [item for item in items  │ │ │
│ │         │  │             if item.get('status')==] │ │ │
│ │         │  │              'ok']                     │ │ │
│ │         │  │                                         │ │ │
│ │         │  │ [Aplicar] [Copiar] [Descartar]          │ │ │
│ │         │  └────────────────────────────────────────┘ │ │
│ └─────────┴────────────────────────────────────────────┘ │
│                                                    │
│  [✓] CodeMin ativo  │  ollama: qwen2.5-coder:7b    │
│  Tokens: 5.2/s       │  RAM: 5.1/16 GB │  CPU: 45% │
└──────────────────────────────────────────────────────────┘
```

### 10.4 CLI UX — Comandos Planejados

```
codemin
├── install          → Instala CodeMin (Ollama + modelo + configs)
├── status           → Verifica status da instalação
├── doctor           → Diagnóstico completo
├── chat [args]      → Chat interativo no terminal
├── config           → Mostra caminhos das configs geradas
├── model
│   ├── list         → Lista modelos disponíveis e instalados
│   ├── download     → Baixa modelo específico
│   ├── remove       → Remove modelo
│   └── switch       → Alterna entre modelos
├── update           → Atualiza CodeMin
├── clean            → Remove caches e arquivos temporários
├── bench [model]    → Benchmark de desempenho
├── finetune [repo]  → Fine-tuning do modelo para codebase
├── uninstall        → Remove CodeMin completamente
└── help             → Ajuda
```

---

## 11. Riscos e Mitigações

| ID | Risco | Probabilidade | Impacto | Mitigação |
|----|-------|---------------|---------|-----------|
| R01 | **Modelo 7B muito lento em CPUs fracas** | Alta (i5/Ryzen 5 de entrada) | Alto — experiência ruins | Fallback automático para Qwen2.5-Coder 1.5B ou CodeLlama 3B quando RAM < 12 GB; benchmark de detecção automática |
| R02 | **Usuário com 8 GB RAM tem travamentos** | Média | Médio — abandono | Detectar na instalação; alertar e oferecer modelo menor; guia de otimização (fechar Chrome, etc.) |
| R03 | **Download de 4.7 GB em conexão lenta** | Média | Médio — frustração | Barra de progresso; resume suportado; download via torrent como fallback |
| R04 | **Ollama não suporta Windows nativamente bem** | Média | Alto — perda de usuários Windows | Testar exhaustivamente no Windows; winget/choco como fallback; documentar troubleshooting |
| R05 | **Mudanças na API do OpenCode quebram compatibilidade** | Baixa | Alto — perda de integração principal | Testes de compatibilidade em CI/CD; pin de versões; fallback para chamadas direct ao Ollama |
| R06 | **Modelo Qwen2.5-Coder fica obsoleto** | Média | Médio — qualidade inferior | Suporte a múltiplos modelos; migração fácil entre modelos; comunidade escolhe novos defaults |
| R07 | **Usuário não sabe configurar OpenCode + Continue** | Alta | Alto — abandono na configuração | CLI gera arquivos automaticamente; `codemin doctor` valida; documentação visual com screenshots |
| R08 | **Consumo de CPU impacta produtividade** | Média | Médio — redução de uso | `codemin idle` pausa servidor; integração com sensor de bateria; throttling automático |
| R09 | **Concorrência de soluções locais gratuitas** | Baixa | Baixo — diferenciação necessária | Foco em UX impecável (2 comandos); integração OpenCode nativa; comunidade ativa |
| R10 | **Vazamento de dados via prompt injection** | Média | Alto — privacidade comprometida | Sanitização de inputs; sandbox do modelo; documentação de segurança |

---

## 12. Roadmap Sugerido

### Sprint 1-2: MVP (Dias 1-14)

```
Sprint 1 (Dias 1-7)          Sprint 2 (Dias 8-14)
┌──────────────────┐        ┌──────────────────┐
│ CLI skeleton     │        │ Chat funcional   │
│ codemin install  │        │ codemin chat     │
│ Download modelo  │───────►│ Integração       │
│ Health check     │        │  OpenCode        │
│ Config Generator │        │ Geração código   │
│ Documentação     │        │ Explicação (C)   │
│ MVP básico       │        │ Testes MVP       │
└──────────────────┘        └──────────────────┘
                                        │
                                        ▼
                            🚀 MVP Release
                          (Dia 14-15)

GO/NO-GO MVP:
  □ codemin install funciona em Windows/Mac/Linux
  □ Chat OpenCode funcional com latência < 15s
  □ Configs geradas corretamente
  □ 3 linguagens (Python, JS, TS)
```

### Sprint 3-6: V2 (Semanas 3-6)

```
Sprint 3                    Sprint 4                    Sprint 5-6
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ Autocomplete FIM  │       │ Refatoração     │       │ Code Review      │
│ Setup Continue   │──────►│ Detecção de bugs│──────►│ Testes unitários │
│ Multi-linguagem  │       │ 4 linguagens    │       │ Documentação     │
│ (Python, JS, TS, │       │ Documentação    │       │ Fallback 1.5B    │
│  Java)           │       │ Comandos novos  │       │ Beta público     │
└──────────────────┘       └──────────────────┘       └──────────────────┘
                                                                │
                                                                ▼
                                                    🚀 V2 Release
                                                  (Semanas 6-7)

GO/NO-GO V2:
  □ Autocomplete < 3s P95
  □ Refatoração funcional em 4 linguagens
  □ Detecção de bugs com acurácia > 70%
  □ NPS > 30 no beta
  □ < 10% de crash rate
```

### Sprint 7-12: V3 (Semanas 7-12)

```
Sprint 7-8                  Sprint 9-10                 Sprint 11-12
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ 8+ linguagens    │       │ Multi-modelo     │       │ Plugin VS Code   │
│ Go, Rust, C++    │──────►│ Selecionável     │──────►│ Interface web    │
│ C#, Ruby         │       │ Gerenciamento    │       │ Suporte JetBrains│
│ Benchmark suite  │       │ Multi-arquivo    │       │ Fine-tuning      │
└──────────────────┘       └──────────────────┘       └──────────────────┘
                                                                │
                                                                ▼
                                                    🚀 V3 Release
                                                  (Semana 12-13)

GO/NO-GO V3:
  □ 8 linguagens cobertas
  □ 3+ modelos selecionáveis
  □ Plugin VS Code publicado na marketplace
  □ Performance: autocomplete < 2s, chat < 10s
  □ 100+ usuários ativos semanais
  □ 80%+ cobertura de testes no código do CodeMin
```

---

## 13. Critérios de Go/No-Go

### 13.1 Gate MVP

| Critério | Mínimo | Ideal | Como Medir |
|----------|--------|-------|------------|
| Instalação funcional | 3 plataformas (Win/Mac/Linux) | +ARM64 | Testes em CI/CD |
| Chat OpenCode | Responde em < 15s | < 10s | Benchmark manual |
| Config auto-gerada | OpenCode + Continue | +README | `codemin config` |
| Latência P95 | < 15s chat | < 10s | `codemin bench` |
| RAM total | < 6 GB com 7B | < 5 GB | `htop` / Task Manager |
| **Decisão** | 4/6 critérios mínimos + nenhum crítico falhando | - | - |

### 13.2 Gate V2

| Critério | Mínimo | Ideal | Como Medir |
|----------|--------|-------|------------|
| Autocomplete FIM | < 5s | < 3s | Benchmark |
| Refatoração | 3/4 linguagens | 4/4 | Testes funcionais |
| Bug Detection | > 60% precisão | > 75% | Dataset de bugs sintéticos |
| 4 linguagens | Python, JS, TS, Java | +Go | Testes de geração |
| NPS beta | > 20 | > 40 | Survey |
| Crash rate | < 15% | < 5% | Error tracking local (opt-in) |
| **Decisão** | 5/7 critérios mínimos | - | - |

### 13.3 Gate V3

| Critério | Mínimo | Ideal | Como Medir |
|----------|--------|-------|------------|
| 8 linguagens | 6/8 | 8/8 | Testes automáticos |
| Multi-modelo | 2 modelos | 3+ modelos | Testes de integração |
| Plugin VS Code | Publicado | +50 instalações | VS Code Marketplace |
| Performance | < 2s auto, < 10s chat | < 1.5s, < 7s | Benchmark |
| Usuários ativos | 100+ | 500+ | Analytics opt-in |
| Cobertura de testes | 70% | 80% | Vitest / c8 |
| Open-source health | 3+ contribuidores | 10+ | GitHub insights |
| **Decisão** | 5/7 critérios mínimos | - | - |

---

## 14. Glossário

| Termo | Definição |
|-------|-----------|
| **GGUF** | Formato de arquivo para modelos quantizados do llama.cpp. Otimizado para carregamento rápido e eficiente em RAM. |
| **Q4_K_M** | Esquema de quantização 4-bit com qualidade intermediária (K_M = K-mean Medium). Balanço entre tamanho (~4.7 GB para 7B) e qualidade. |
| **FIM** | Fill-in-the-Middle. Técnica de autocomplete onde o modelo completa código entre um prefixo e um sufixo. |
| **MoSCoW** | Técnica de priorização: Must have, Should have, Could have, Won't have. |
| **Ollama** | Runtime de modelos LLM local. Gerencia download, servidor e inferência. |
| **llama.cpp** | Engine de inferência C++ otimizada para CPU. Base do Ollama. |
| **Quantização** | Técnica de compressão de modelos que reduz precisão numérica (32-bit → 4-bit) para diminuir tamanho e requisitos de RAM. |
| **Context Window** | Quantidade máxima de tokens que o modelo pode processar em uma única chamada. Qwen2.5-Coder suporta 32K tokens. |
| **Token** | Unidade básica de processamento do LLM. ~0.75 palavras em inglês, ~0.4 em português. |
| **Prompt** | Texto de entrada enviado ao modelo para gerar uma resposta. |
| **OpenCode** | (Fictício para este PRD) Editor/framework que interage com modelos via API compatível com OpenAI. |
| **Continue.dev** | Extensão open-source para VS Code/JetBrains que conecta IDEs a LLMs locais ou remotos. |

---

## 15. Referências

### Modelos
- [Qwen2.5-Coder-7B-Instruct (HuggingFace)](https://huggingface.co/Qwen/Qwen2.5-Coder-7B-Instruct)
- [Qwen2.5-Coder GGUF (bartowski)](https://huggingface.co/bartowski/Qwen2.5-Coder-7B-Instruct-GGUF)
- [CodeLlama (Meta)](https://ai.meta.com/blog/code-llama-large-language-model-coding/)
- [DeepSeek-Coder](https://github.com/deepseek-ai/deepseek-coder)

### Ferramentas
- [Ollama](https://ollama.com/)
- [llama.cpp](https://github.com/ggerganov/llama.cpp)
- [Continue.dev](https://docs.continue.dev/)
- [OpenCode — OpenAI API Compatible Providers](https://opencode.ai/docs)

### Benchmark e Referência Técnica
- [HumanEval+ (EvalPlus)](https://github.com/evalplus/evalplus)
- [BigCode Leaderboard](https://huggingface.co/spaces/bigcode/bigcode-models-leaderboard)
- [llama.cpp Performance Guide](https://github.com/ggerganov/llama.cpp/wiki/Speed-and-memory-usage)

### Padrões e Licenças
- [MIT License](https://opensource.org/license/mit/)
- [SemVer 2.0](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)

---

## 16. Questões em Aberto

- [ ] **Suporte ARM64 (Apple Silicon / Raspberry Pi):** Deve ser MVP ou V2? O Qwen2.5-Coder 7B funciona bem em M1?
- [ ] **Modo "turbo" com batches:** Vale a pena implementar processamento em lote para throughput? Impacto na latência?
- [ ] **Modelo de fallback ideal:** Qual o melhor modelo sub-3B para CPU fraca? Qwen2.5-Coder 1.5B vs Stable Code 3B vs DeepSeek-Coder 1.3B?
- [ ] **Analytics opt-in:** Devemos coletar métricas de uso anônimas (opt-in) para melhorar o produto? Trade-off com privacidade.
- [ ] **Estratégia de suporte:** Discord? GitHub Discussions? Issues apenas? Qual canal oficial?
- [ ] **Nome final:** "CodeMin" é definitivo? Alternativas: MiniCoder, LocalCoder, CodeLocal, CPUCoder?
- [ ] **Publicação npm:** Pacote já reservado? Verificar disponibilidade de `codemin` no npm.
- [ ] **Licenciamento do modelo:** Qwen2.5-Coder é Apache 2.0 — OK para uso comercial? Confirmar termos.

---

## 17. Apêndice — Comparação com Alternativas

| Característica | CodeMin | GitHub Copilot | Cursor | CodeLLaMA (manual) |
|----------------|---------|---------------|--------|-------------------|
| **Custo** | Grátis | US$ 10-39/mês | US$ 20/mês | Grátis |
| **GPU necessária** | ❌ Não | N/A (nuvem) | N/A (nuvem) | Sim (6 GB+) |
| **CPU-only** | ✅ Sim | N/A | N/A | Parcial (lento) |
| **Online/Offline** | Offline | Online | Online | Offline |
| **Privacidade** | Total | Código vai pra nuvem | Código vai pra nuvem | Total |
| **Autocomplete** | ≤ 3s (V2) | < 1s | < 1s | > 5s |
| **Qualidade** | Boa (7B Q4) | Excelente (GPT-4) | Excelente | Boa |
| **Instalação** | 2 comandos | Plugin + login | Download + login | 5+ passos manuais |
| **Multi-linguagem** | 8+ (V3) | 20+ | 20+ | Ilimitado |
| **Open-source** | ✅ Sim | ❌ Não | ❌ Não | ✅ Sim |
| **Tamanho download** | ~4.7 GB | N/A | N/A | 12+ GB (fp16) |

---

**Gerado por:** Morgan (AIOX Product Manager)
**Template:** AIOX PRD v2.0
**Data de Criação:** 2026-07-09
**Revisões:** 1.0.0 — Versão inicial

---
