---
title: "CodeMin — Histórias de Usuário"
version: "1.1.0"
author: "River (AIOX Scrum Master) + Pax (AIOX Product Owner)"
date: "2026-07-09"
source: "docs/codemin-prd.md"
status: "PO-Approved"
total_stories: 28
phases:
  mvp: "Sprint 1-2 — 10 histórias (1 enabler)"
  v2: "Sprint 3-6 — 9 histórias"
  v3: "Sprint 7-12 — 9 histórias (3 placeholders)"
  note: "Placeholders = histórias Could não detalhadas; adicionar durante refinement"
personas:
  joana: "Estudante / Dev Júnior — notebook sem GPU, sem verba, VS Code + Python/JS"
  carlos: "Dev Full Stack — startup BR, Java+React, código sensível, sem verba corporativa"
  ana: "Contribuidora Open-Source — hardware antigo (8 GB RAM), Go/Rust/Python, filosofia MIT"
---

# CodeMin — Histórias de Usuário

> **Produto:** CodeMin — Assistente de Codificação LLM Local (CPU-only)
> **Base:** PRD v1.0.0 (`docs/codemin-prd.md`)
> **Criado por:** 🌊 River (AIOX Scrum Master)

---

## Sumário

- [Fase MVP (Sprint 1-2)](#fase-mvp-sprint-1-2)
- [Fase V2 (Sprint 3-6)](#fase-v2-sprint-3-6)
- [Fase V3 (Sprint 7-12)](#fase-v3-sprint-7-12)
- [Resumo de Estimativas](#resumo-de-estimativas)

---

## Fase MVP (Sprint 1-2)

---

### FR-MVP-01 — CLI de Instalação

**Fase:** MVP
**Sprint:** 1
**Prioridade:** Must
**Story Points:** 8

#### Descrição

Como **Joana**, que quer usar um assistente de código mas tem medo de configurações complexas, eu quero executar um único comando `codemin install` para que todo o ambiente seja instalado automaticamente — Ollama, modelo e configurações — sem precisar de conhecimento técnico avançado.

#### Critérios de Aceitação

- [ ] `codemin install` completa com sucesso em < 10 minutos em conexão de banda larga
- [ ] Cria o diretório `~/.codemin/` com estrutura de subdiretórios (`models/`, `configs/`, `logs/`)
- [ ] Instala ou detecta Ollama (Windows, macOS, Linux) automaticamente
- [ ] Exibe barra de progresso com logs visíveis para cada etapa (1/4 Verificando, 2/4 Instalando Ollama, 3/4 Baixando modelo, 4/4 Configurando)
- [ ] Em caso de falha, exibe mensagem clara de erro e instruções de resolução
- [ ] Funciona nas 3 plataformas: Windows 10+, macOS 12+, Linux (Ubuntu 20.04+ / Fedora 36+)
- [ ] Instalação não requer `sudo` nem permissões administrativas (exceto no Windows se necessário para PATH)

#### Notas Técnicas

- **Dependências:** Nenhuma (é a porta de entrada)
- **Componentes envolvidos:** CLI (Node.js/TypeScript), shell scripts de instalação, npm
- **Complexidade:** Alta — orquestra múltiplos subsistemas (Ollama, download, configs) com fallbacks por plataforma
- **Observação:** O comando deve ser idempotente — executar novamente deve atualizar componentes existentes sem quebrar

---

### FR-MVP-02 — Download de Modelo

**Fase:** MVP
**Sprint:** 1
**Prioridade:** Must
**Story Points:** 5

#### Descrição

Como **Joana**, que tem uma conexão de internet limitada, eu quero que o `codemin install` baixe o modelo Qwen2.5-Coder 7B Q4_K_M automaticamente com barra de progresso e suporte a resume para que eu não precise reiniciar o download se a conexão cair.

#### Critérios de Aceitação

- [ ] Download do modelo GGUF Q4_K_M (~4.7 GB) do HuggingFace com barra de progresso em tempo real
- [ ] Resume automático de download interrompido (não recomeça do zero)
- [ ] Validação de checksum SHA256 após download completo
- [ ] Modelo armazenado em `~/.codemin/models/qwen2.5-coder-7b-q4_k_m.gguf`
- [ ] Falha com mensagem clara se espaço em disco for insuficiente (< 10 GB livres)
- [ ] Tempo de download exibido ao final

#### Notas Técnicas

- **Dependências:** FR-MVP-01 (estrutura de diretórios)
- **Componentes envolvidos:** CLI (download via HTTPS), HuggingFace API, disco local
- **Complexidade:** Média — download resiliente com resume e validação
- **URL do modelo:** `https://huggingface.co/bartowski/Qwen2.5-Coder-7B-Instruct-GGUF/resolve/main/Qwen2.5-Coder-7B-Instruct-Q4_K_M.gguf`

---

### FR-MVP-03 — Chat Contextual (OpenCode)

**Fase:** MVP
**Sprint:** 2
**Prioridade:** Must
**Story Points:** 8

#### Descrição

Como **Ana**, que trabalha com código em projetos open-source no VS Code, eu quero selecionar um trecho de código e fazer perguntas sobre ele via OpenCode (Ctrl+I) para receber explicações e sugestões sem sair do editor.

#### Critérios de Aceitação

- [ ] Chat funcional via OpenCode com modelo local Qwen2.5-Coder 7B
- [ ] Resposta contextualmente relevante ao código selecionado em < 15s (P95)
- [ ] Respostas em streaming (token a token) para feedback imediato
- [ ] Suporte a perguntas em português e inglês
- [ ] Histórico de conversa mantido durante a sessão do OpenCode
- [ ] System prompt otimizado para assistência de codificação incluído na configuração

#### Notas Técnicas

- **Dependências:** FR-MVP-01, FR-MVP-02, FR-MVP-04 (opencode.json)
- **Componentes envolvidos:** OpenCode, Ollama, CLI (geração de config), modelo Qwen2.5-Coder
- **Complexidade:** Alta — integração entre OpenCode e Ollama via API compatível OpenAI; otimização de contexto e streaming
- **Parâmetros do modelo:** temperature 0.2, maxTokens 2048, contextLength 8192

---

### FR-MVP-04 — Config Generator → opencode.json

**Fase:** MVP
**Sprint:** 1
**Prioridade:** Must
**Story Points:** 3

#### Descrição

Como **Carlos**, que precisa de privacidade no código do cliente, eu quero que o `codemin install` gere automaticamente o arquivo `opencode.json` configurado para usar o modelo local para que eu possa começar a usar o assistente imediatamente sem configuração manual.

#### Critérios de Aceitação

- [ ] Arquivo `opencode.json` gerado em `~/.codemin/configs/opencode.json`
- [ ] Configuração contém: provider `openai`, model `qwen2.5-coder:7b`, apiBase `http://localhost:11434/v1`
- [ ] Parâmetros otimizados: stream `true`, maxTokens 2048, temperature 0.2, contextLength 8192
- [ ] CLI exibe instruções claras de onde colocar o arquivo (raiz do projeto)
- [ ] Arquivo gerado é válido (JSON parseável)
- [ ] Comando `codemin config` mostra o caminho do arquivo gerado

#### Notas Técnicas

- **Dependências:** FR-MVP-01 (estrutura de diretórios)
- **Componentes envolvidos:** CLI (geração de JSON), sistema de arquivos
- **Complexidade:** Baixa — geração de arquivo JSON estático com template
- **Template de saída:** Ver PRD seção 8.1

---

### FR-MVP-05 — Config Generator → continue.config.json

**Fase:** MVP
**Sprint:** 1
**Prioridade:** Must
**Story Points:** 3

#### Descrição

Como **Joana**, que usa VS Code e quer autocomplete e chat no editor, eu quero que o `codemin install` gere automaticamente o `config.json` do Continue.dev para que a extensão reconheça o modelo local sem configuração manual.

#### Critérios de Aceitação

- [ ] Arquivo `config.json` gerado em `~/.codemin/configs/continue.config.json`
- [ ] Configuração contém `models` com provider `ollama` apontando para `qwen2.5-coder:7b`, contextLength 8192
- [ ] Configuração contém `tabAutocompleteModel` configurado para autocomplete (preparado para V2)
- [ ] System message personalizado do CodeMin incluído
- [ ] CLI exibe instruções de onde colocar o arquivo (`~/.continue/config.json`)
- [ ] Arquivo gerado é JSON válido
- [ ] Compatível com Continue.dev v0.8.x ou superior (verificar documentação na data de implementação)

#### Notas Técnicas

- **Dependências:** FR-MVP-01 (estrutura de diretórios)
- **Componentes envolvidos:** CLI (geração de JSON), Continue.dev, sistema de arquivos
- **Complexidade:** Baixa — geração de arquivo JSON estático com template
- **Template de saída:** Ver PRD seção 8.2

---

### FR-MVP-06 — Status Check (`codemin status`)

**Fase:** MVP
**Sprint:** 1
**Prioridade:** Must
**Story Points:** 2

#### Descrição

Como **Ana**, que gosta de ter controle sobre suas ferramentas, eu quero executar `codemin status` para ver rapidamente se tudo está funcionando — Ollama rodando, modelo baixado, configurações presentes — com indicadores visuais claros.

#### Critérios de Aceitação

- [ ] Comando `codemin status` executa e retorna em < 2 segundos
- [ ] Verifica e exibe: ✅/❌ Ollama está rodando (localhost:11434)
- [ ] Verifica e exibe: ✅/❌ Modelo Qwen2.5-Coder 7B está baixado
- [ ] Verifica e exibe: ✅/❌ Configurações do OpenCode e Continue.dev existem
- [ ] Verifica e exibe: ✅/❌ Espaço em disco suficiente (> 5 GB livres)
- [ ] Saída formatada com cores e ícones para leitura rápida
- [ ] Exit code 0 se tudo OK, exit code 1 se algum componente falhar

#### Notas Técnicas

- **Dependências:** FR-MVP-01, FR-MVP-02, FR-MVP-04, FR-MVP-05
- **Componentes envolvidos:** CLI (health check), Ollama API, sistema de arquivos
- **Complexidade:** Baixa — verificações isoladas e independentes
- **Comando relacionado:** `codemin doctor` (diagnóstico mais detalhado — V2)

---

### FR-MVP-07 — Chat Nativo no Terminal

**Fase:** MVP
**Sprint:** 2
**Prioridade:** Should
**Story Points:** 5

#### Descrição

Como **Ana**, que prefere o terminal ao invés de IDEs pesadas, eu quero um comando `codemin chat` que abra uma interface de chat interativa no terminal para conversar com o modelo sem precisar de VS Code ou OpenCode.

#### Critérios de Aceitação

- [ ] `codemin chat` inicia sessão interativa no terminal com prompt estilizado
- [ ] Respostas em streaming (token a token) com tempo de primeira resposta < 5s
- [ ] Histórico da sessão mantido (entender contexto de perguntas anteriores)
- [ ] Comandos especiais: `/exit` (sair), `/clear` (limpar histórico), `/help` (comandos)
- [ ] Suporte a multiline (digitar perguntas com Enter+Shift)
- [ ] Tempo total de resposta < 15s (P95) para prompts de até 500 tokens
- [ ] Destaque de sintaxe básico para blocos de código na resposta
- [ ] Ctrl+C interrompe geração e retorna ao prompt, não fecha o chat

#### Notas Técnicas

- **Dependências:** FR-MVP-01, FR-MVP-02 (modelo baixado)
- **Componentes envolvidos:** CLI (Node.js readline/ink), Ollama API, streaming
- **Complexidade:** Média — interface interativa no terminal com streaming e histórico
- **Bibliotecas sugeridas:** `ink` (React para terminal) ou `readline` nativo; `chalk` para cores

---

### FR-MVP-08 — Geração de Código via Linguagem Natural

**Fase:** MVP
**Sprint:** 2
**Prioridade:** Should
**Story Points:** 5

#### Descrição

Como **Joana**, que está aprendendo a programar e precisa de ajuda com tarefas comuns, eu quero descrever em linguagem natural o que preciso ("Crie uma função que valida CPF em Python") para que o CodeMin gere o código correspondente pronto para uso.

#### Critérios de Aceitação

- [ ] Comando `codemin generate "descrição"` ou via chat no OpenCode gera código funcional
- [ ] Código gerado é sintaticamente válido (passa em parser/linter da linguagem alvo)
- [ ] Respeita a linguagem do arquivo atual (se usado via OpenCode com arquivo aberto)
- [ ] Inclui comentários explicativos no código gerado
- [ ] Geração completa em < 20s para funções de até 50 linhas
- [ ] Suporta: Python, JavaScript, TypeScript (MVP); Java (preparação)

#### Notas Técnicas

- **Dependências:** FR-MVP-03 (chat contextual), FR-MVP-07 (chat terminal)
- **Componentes envolvidos:** OpenCode / CLI, Ollama, modelo Qwen2.5-Coder
- **Complexidade:** Média — engenharia de prompt para geração de código; validação sintática
- **System prompt adicional:** "Você é um assistente de geração de código. Gere APENAS o código solicitado, sem explicações extras, a menos que solicitado."

---

### FR-MVP-09 — Explicação de Código

**Fase:** MVP
**Sprint:** 2
**Prioridade:** Could
**Story Points:** 3

#### Descrição

Como **Joana**, que está aprendendo a programar e encontra códigos que não entende, eu quero selecionar um trecho de código e pedir uma explicação em linguagem natural para compreender o que ele faz, quais padrões usa e qual sua complexidade.

#### Critérios de Aceitação

- [ ] Explicação gerada em linguagem natural (português ou inglês) para o código selecionado
- [ ] Identifica e explica: propósito geral, algoritmos usados, estruturas de dados
- [ ] Menção de complexidade de tempo/espaço quando relevante (Big O)
- [ ] Explicação em < 15s para trechos de até 100 linhas
- [ ] Acessível via OpenCode (selecionar código → "Explique") no MVP
- [ ] Bônus (se tempo permitir): comando `codemin explain <arquivo>:<linhas>` — validar se entra no escopo do MVP
- [ ] Explicação é concisa (máximo 3 parágrafos) mas completa

#### Notas Técnicas

- **Dependências:** FR-MVP-03 (chat contextual)
- **Componentes envolvidos:** OpenCode, Ollama, CLI
- **Complexidade:** Média — engenharia de prompt para análise de código
- **Prompt sugestão:** "Explique o seguinte código de forma didática: identifique o propósito, algoritmos, complexidade e padrões usados."

---

### FR-MVP-DOC-01 — Documentação Inicial / README

**Fase:** MVP
**Tipo:** Enabler (Documentação)
**Sprint:** 2
**Prioridade:** Could
**Story Points:** 2

#### Descrição

Como **Ana**, que acredita que software livre deve ser bem documentado, eu quero um README completo e documentação inicial do CodeMin para que eu e outros usuários possamos instalar, configurar e usar a ferramenta sem depender de terceiros.

#### Critérios de Aceitação

- [ ] README.md na raiz do repositório com: visão geral, instalação (2 comandos), pré-requisitos
- [ ] Seção de configuração com exemplos dos arquivos gerados (opencode.json, continue.config.json)
- [ ] Seção de uso com exemplos dos comandos principais (install, status, chat, generate)
- [ ] Badges de status (CI, licença, versão)
- [ ] Seção de troubleshooting com problemas comuns e soluções
- [ ] Documentação em português (README primário) com link para versão em inglês
- [ ] Seção de contribuição e licença (MIT)

#### Notas Técnicas

- **Dependências:** FR-MVP-01 a FR-MVP-09 (documentar o que foi implementado)
- **Componentes envolvidos:** Repositório GitHub, Markdown
- **Complexidade:** Baixa — documentação estática
- **Observação:** README deve ser atualizado a cada release com novas funcionalidades

---

## Fase V2 (Sprint 3-6)

---

### FR-V2-01 — Autocomplete FIM via Continue.dev

**Fase:** V2
**Sprint:** 3
**Prioridade:** Must
**Story Points:** 8

#### Descrição

Como **Joana**, que quer produtividade enquanto digita código no VS Code, eu quero sugestões de autocomplete inline (Fill-in-the-Middle) enquanto digito, via Continue.dev, para que eu escreva código mais rápido sem precisar parar para consultar o chat.

#### Critérios de Aceitação

- [ ] Sugestões de autocomplete aparecem enquanto digita no VS Code (via Continue.dev)
- [ ] Latência P95 < 3s para exibir primeira sugestão
- [ ] Sugestões são contextualmente relevantes (baseadas no código ao redor)
- [ ] Aceitação com `Tab` insere o código sugerido
- [ ] Rejeição com `Esc` descarta a sugestão
- [ ] Funciona para Python, JavaScript, TypeScript (MVP + V2 languages)
- [ ] Não atrapalha a digitação (sugestões não aparecem em momentos inapropriados)
- [ ] Pode ser desabilitado temporariamente por comando

#### Notas Técnicas

- **Dependências:** FR-MVP-05 (continue.config.json com tabAutocompleteModel)
- **Componentes envolvidos:** Continue.dev, VS Code, Ollama, modelo Qwen2.5-Coder (suporte FIM)
- **Complexidade:** Alta — FIM requer formato de prompt específico; otimização de latência crítica
- **Observação:** Qwen2.5-Coder tem suporte nativo a FIM. Template FIM: `<|fim_prefix|>{prefix}<|fim_suffix|>{suffix}<|fim_middle|>`
- **Parâmetros:** `maxTokens` para autocomplete deve ser baixo (64-128 tokens) para latência mínima

---

### FR-V2-02 — Refatoração de Código

**Fase:** V2
**Sprint:** 4
**Prioridade:** Must
**Story Points:** 5

#### Descrição

Como **Carlos**, que trabalha com código legado em Java que não pode ser enviado para serviços em nuvem, eu quero selecionar um trecho de código e dar instruções de refatoração ("Extraia esse trecho para uma função separada") para que o CodeMin gere o código refatorado localmente sem comprometer a privacidade.

#### Critérios de Aceitação

- [ ] Comando de refatoração acessível via OpenCode e `codemin refactor "instrução"`
- [ ] Código refatorado é sintaticamente válido na linguagem alvo
- [ ] Suporta refatorações comuns: extrair função, renomear variável, simplificar condicional, converter loop para comprehension/list.map
- [ ] Preserva a semântica do código original (comportamento idêntico)
- [ ] Mudanças destacadas no diff mostrado ao usuário antes de aplicar
- [ ] Opções: [Aplicar] [Copiar] [Descartar] para o código gerado
- [ ] Suporta Python, JavaScript, TypeScript, Java

#### Notas Técnicas

- **Dependências:** FR-MVP-03 (chat contextual), FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** OpenCode, Ollama, CLI
- **Complexidade:** Média — engenharia de prompt para refatoração; diff preview
- **Prompt:** "Refatore o código abaixo seguindo a instrução: {instrução}. Mantenha a semântica idêntica. Retorne APENAS o código refatorado."

---

### FR-V2-03 — Suporte Multi-linguagem (Python, JS, TS, Java)

**Fase:** V2
**Sprint:** 3
**Prioridade:** Must
**Story Points:** 8

#### Descrição

Como **Carlos**, que trabalha com Java no backend e React (JS/TS) no frontend, eu quero que o CodeMin funcione bem em todas as 4 linguagens principais (Python, JavaScript, TypeScript, Java) para que eu possa usar a mesma ferramenta em todo o meu stack.

#### Critérios de Aceitação

- [ ] Geração de código funcional nas 4 linguagens
- [ ] Chat contextual entende e responde corretamente código nas 4 linguagens
- [ ] Acurácia sintática > 90% (código gerado passa em parser/linter específico da linguagem)
- [ ] System prompt adaptado por linguagem quando detectada
- [ ] Autocomplete FIM (FR-V2-01) funciona nas 4 linguagens
- [ ] Testes automatizados de geração para cada linguagem com fixtures representativas

#### Notas Técnicas

- **Dependências:** FR-MVP-08 (geração de código), FR-V2-01 (autocomplete), FR-V2-02 (refatoração)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder (multi-linguagem nativo), linters/parsers
- **Complexidade:** Alta — requer validação sintática por linguagem; prompts específicos
- **Observação:** Qwen2.5-Coder foi treinado com múltiplas linguagens, mas precisamos testar e ajustar prompts por linguagem

---

### FR-V2-04 — Detecção de Bugs

**Fase:** V2
**Sprint:** 4
**Prioridade:** Must
**Story Points:** 5

#### Descrição

Como **Carlos**, que revisa código legado e precisa encontrar problemas rapidamente, eu quero selecionar um trecho de código e pedir uma análise de bugs para que o CodeMin aponte possíveis erros como null pointer, out-of-bounds e variáveis não utilizadas.

#### Critérios de Aceitação

- [ ] Detecta e reporta com linha específica: null pointer / NoneType errors
- [ ] Detecta e reporta: acesso fora dos limites da lista/array (out-of-bounds)
- [ ] Detecta e reporta: variáveis declaradas mas não utilizadas
- [ ] Detecta e reporta: tipos incompatíveis / type mismatch
- [ ] Reporta severidade (Alta/Média/Baixa) para cada bug encontrado
- [ ] Sugere correção para cada bug detectado
- [ ] Cobertura de detecção > 70% em dataset de bugs sintéticos
- [ ] Funciona para Python, JavaScript, TypeScript, Java

#### Notas Técnicas

- **Dependências:** FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder, CLI
- **Complexidade:** Média — engenharia de prompt para análise estática; classificação por severidade
- **Prompt:** "Analise o código abaixo e identifique bugs. Para cada bug: linha, tipo, severidade (Alta/Média/Baixa), e sugestão de correção. Formato: | Linha | Tipo | Severidade | Correção |"

---

### FR-V2-05 — Code Review

**Fase:** V2
**Sprint:** 5
**Prioridade:** Should
**Story Points:** 5

#### Descrição

Como **Carlos**, que não pode expor código do cliente em serviços externos, eu quero uma análise de segurança e boas práticas do meu código (localmente) para identificar vulnerabilidades como SQL injection, XSS, hardcoded secrets antes do deploy.

#### Critérios de Aceitação

- [ ] Identifica e reporta: possíveis SQL injections
- [ ] Identifica e reporta: possíveis XSS (Cross-Site Scripting)
- [ ] Identifica e reporta: hardcoded secrets (senhas, API keys, tokens)
- [ ] Identifica e reporta: falta de validação de entrada
- [ ] Identifica e reporta: práticas inseguras (eval, exec, shell injection)
- [ ] Classifica cada achado por severidade (Crítico/Alto/Médio/Baixo/Info)
- [ ] Gera relatório formatado no terminal e opção de exportar JSON
- [ ] Funciona para Python, JavaScript, TypeScript, Java

#### Notas Técnicas

- **Dependências:** FR-V2-04 (detecção de bugs), FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder, CLI
- **Complexidade:** Média — engenharia de prompt para segurança; categorização OWASP Top 10
- **Observação:** Não substitui ferramentas especializadas (Semgrep, SonarQube) — é uma camada adicional de segurança

---

### FR-V2-06 — Geração de Testes

**Fase:** V2
**Sprint:** 5
**Prioridade:** Should
**Story Points:** 5

#### Descrição

Como **Carlos**, que precisa escrever testes mas tem pouco tempo, eu quero selecionar uma função/método e pedir ao CodeMin que gere testes unitários automaticamente no framework apropriado para garantir a qualidade do código sem esforço manual.

#### Critérios de Aceitação

- [ ] Gera testes no framework da linguagem: pytest (Python), Jest (JS/TS), JUnit (Java)
- [ ] Cobertura > 80% da função alvo (incluindo casos de borda)
- [ ] Testes gerados são executáveis (passam em ambiente com as dependências corretas)
- [ ] Inclui: caso feliz (happy path), casos de erro, casos de borda (edge cases)
- [ ] Nome dos testes segue convenção da linguagem (test_*, should*, @Test)
- [ ] Usuário pode escolher: [Copiar] [Salvar em arquivo] [Descartar]
- [ ] Funciona para Python, JavaScript, TypeScript, Java

#### Notas Técnicas

- **Dependências:** FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder, CLI
- **Complexidade:** Média — engenharia de prompt para geração de testes; validação de sintaxe
- **Prompt:** "Gere testes unitários para a função abaixo usando {framework}. Cubra: caso feliz, casos de erro, edge cases. Siga as convenções de nomenclatura da linguagem."

---

### FR-V2-07 — Documentação Automática

**Fase:** V2
**Sprint:** 5
**Prioridade:** Should
**Story Points:** 3

#### Descrição

Como **Ana**, que contribui com projetos open-source e precisa manter documentação de qualidade, eu quero selecionar uma função e gerar automaticamente sua docstring/comentários JSDoc/JavaDoc para manter o código bem documentado sem trabalho manual.

#### Critérios de Aceitação

- [ ] Gera documentação seguindo a convenção da linguagem: PEP 257 (Python), JSDoc (JS/TS), JavaDoc (Java)
- [ ] Inclui: descrição da função, parâmetros (tipo + descrição), retorno, exceções
- [ ] Exemplo de uso incluído quando relevante
- [ ] Documentação gerada é coerente com a implementação atual
- [ ] Preserva documentação existente (não sobrescreve, apenas adiciona se ausente)
- [ ] Pode processar arquivo inteiro ou seleção
- [ ] Funciona para Python, JavaScript, TypeScript, Java

#### Notas Técnicas

- **Dependências:** FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder, CLI
- **Complexidade:** Média — parsing de assinatura de função; geração de documentação estruturada
- **Ferramentas auxiliares:** Parser AST para extrair assinaturas de função sem enviar corpo inteiro

---

### FR-V2-08 — Comando de Atualização (`codemin update`)

**Fase:** V2
**Sprint:** 6
**Prioridade:** Could
**Story Points:** 3

#### Descrição

Como **Ana**, que quer manter o CodeMin sempre na versão mais recente, eu quero executar `codemin update` para atualizar tanto a ferramenta quanto os modelos disponíveis para garantir que estou usando as versões mais estáveis e recentes.

#### Critérios de Aceitação

- [ ] `codemin update --check` verifica se há nova versão do CodeMin (npm) sem atualizar
- [ ] `codemin update` atualiza o pacote npm do CodeMin para a última versão estável
- [ ] `codemin update --models` verifica se há novas versões dos modelos instalados
- [ ] Em caso de falha na atualização, rollback para versão anterior
- [ ] Mensagem clara de changelog/resumo do que foi atualizado

#### Notas Técnicas

- **Dependências:** FR-MVP-01 (CLI instalado)
- **Componentes envolvidos:** npm, CLI, Ollama API
- **Complexidade:** Baixa — wrapper sobre npm update + ollama pull
- **Observação:** Pode ser implementado como script simples; não requer estado complexo

---

### FR-V2-09 — Fallback Automático para Modelo Menor

**Fase:** V2
**Sprint:** 6
**Prioridade:** Could
**Story Points:** 5

#### Descrição

Como **Ana**, cujo ThinkPad tem apenas 8 GB de RAM, eu quero que o CodeMin detecte automaticamente quando a memória é insuficiente para o modelo 7B e ofereça usar um modelo menor (Qwen2.5-Coder 1.5B) para que eu possa usar o assistente mesmo em hardware limitado sem travamentos.

#### Critérios de Aceitação

- [ ] Detecção automática de RAM total < 12 GB na instalação → sugere fallback para modelo 1.5B
- [ ] `codemin use --small` alterna para Qwen2.5-Coder 1.5B
- [ ] `codemin use --full` alterna de volta para 7B
- [ ] Modelo 1.5B usa < 3 GB de RAM total (sistema + Ollama)
- [ ] Comando `codemin status` mostra qual modelo está ativo e consumo de RAM
- [ ] Troca entre modelos não requer reinstalação completa
- [ ] Download do modelo 1.5B também tem barra de progresso e resume

#### Notas Técnicas

- **Dependências:** FR-MVP-02 (download de modelo), FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Ollama, CLI (gerenciamento de modelos), sistema (detecção de RAM)
- **Modelo fallback:** Qwen2.5-Coder 1.5B Q4_K_M (~1 GB)
- **Detecção:** `os.totalmem()` no Node.js ou leitura de `/proc/meminfo` / `wmic`
- **Complexidade:** Média — detecção de hardware; gerenciamento de múltiplos modelos

---

## Fase V3 (Sprint 7-12)

---

### FR-V3-01 — Suporte a 8+ Linguagens

**Fase:** V3
**Sprint:** 7
**Prioridade:** Must
**Story Points:** 8

#### Descrição

Como **Ana**, que programa em Go e Rust além de Python, eu quero que o CodeMin suporte 8 linguagens de programação (adicionando Go, Rust, C++, C#, Ruby) para que eu possa usar o assistente em todos os meus projetos sem limitação de linguagem.

#### Critérios de Aceitação

- [ ] Geração de código funcional nas 8 linguagens: Python, JS, TS, Java, Go, Rust, C++, C#, Ruby
- [ ] Chat contextual entende e responde corretamente código nas 8 linguagens
- [ ] Acurácia sintática > 85% em todas as 8 linguagens (código gerado passa em parser/linter)
- [ ] Autocomplete FIM funciona nas 8 linguagens
- [ ] Refatoração, detecção de bugs, geração de testes e code review funcionam nas 8 linguagens
- [ ] Testes automatizados de geração para todas as 8 linguagens
- [ ] Documentação atualizada com exemplos por linguagem

#### Notas Técnicas

- **Dependências:** FR-V2-03 (4 linguagens base)
- **Componentes envolvidos:** Ollama, modelo Qwen2.5-Coder, linters/parsers (gofmt, rustfmt, etc.)
- **Complexidade:** Alta — requer validação sintática; prompts e system prompts específicos por linguagem
- **Observação:** Qwen2.5-Coder suporta naturalmente muitas linguagens, mas qualidade varia — requer tuning de prompt por linguagem

---

### FR-V3-02 — Múltiplos Modelos Selecionáveis

**Fase:** V3
**Sprint:** 8
**Prioridade:** Must
**Story Points:** 5

#### Descrição

Como **Ana**, que prefere software livre e quer escolher a ferramenta ideal para cada tarefa, eu quero poder selecionar entre diferentes modelos (Qwen, CodeLlama, DeepSeek-Coder) para escolher o que funciona melhor no meu hardware e para minha tarefa específica.

#### Critérios de Aceitação

- [ ] `codemin list-models` mostra modelos disponíveis (online e instalados)
- [ ] `codemin use qwen2.5-coder:7b` alterna para o modelo especificado
- [ ] `codemin model info <modelo>` mostra detalhes: tamanho, RAM estimada, linguagens suportadas
- [ ] Download de novo modelo com `codemin model download <nome>`
- [ ] Troca entre modelos sem reinstalar o CodeMin
- [ ] Benchmark comparativo integrado (`codemin bench --compare`) entre modelos instalados
- [ ] Mínimo de 3 modelos disponíveis para download: Qwen2.5-Coder 7B, CodeLlama 7B, DeepSeek-Coder 6.7B

#### Notas Técnicas

- **Dependências:** FR-V3-03 (gerenciamento de modelos), FR-MVP-02 (download)
- **Componentes envolvidos:** CLI, Ollama, HuggingFace / Ollama registry
- **Complexidade:** Média — abstração sobre Ollama para gerenciar múltiplos modelos
- **Observação:** Modelos devem ser compatíveis com a mesma API (ollama pull <modelo>)

---

### FR-V3-03 — Gerenciamento de Modelos (CLI)

**Fase:** V3
**Sprint:** 8
**Prioridade:** Must
**Story Points:** 5

#### Descrição

Como **Ana**, que gerencia múltiplos modelos localmente, eu quero comandos completos de gerenciamento (`codemin model list`, `model download`, `model remove`, `model switch`) para controlar quais modelos estão instalados sem precisar usar o Ollama diretamente.

#### Critérios de Aceitação

- [ ] `codemin model list` lista todos os modelos instalados com tamanho e versão
- [ ] `codemin model download <modelo>` baixa modelo específico com progresso
- [ ] `codemin model remove <modelo>` remove modelo com confirmação
- [ ] `codemin model switch <modelo>` alterna o modelo ativo
- [ ] `codemin model search <termo>` busca modelos disponíveis no registro
- [ ] Confirmação antes de remover modelos (> 1 GB)
- [ ] Comandos têm semântica consistente e mensagens de erro claras
- [ ] `codemin model prune` remove modelos não utilizados (com confirmação)

#### Notas Técnicas

- **Dependências:** FR-V3-02 (seleção de modelos), FR-MVP-02 (download)
- **Componentes envolvidos:** CLI, Ollama API, sistema de arquivos (model storage)
- **Complexidade:** Média — CRUD de modelos; integração com Ollama registry
- **Armazenamento:** Modelos em `~/.codemin/models/` ou via Ollama (`~/.ollama/models/`)

---

### FR-V3-04 — Autocomplete Multi-arquivo (Context-Aware)

**Fase:** V3
**Sprint:** 9
**Prioridade:** Should
**Story Points:** 8

#### Descrição

Como **Carlos**, que trabalha em projetos grandes com múltiplos arquivos, eu quero que o autocomplete do CodeMin considere imports, símbolos e tipos de outros arquivos do projeto para que as sugestões sejam contextualmente precisas e relevantes.

#### Critérios de Aceitação

- [ ] Sugestões de autocomplete baseadas em imports de outros arquivos do projeto
- [ ] Reconhecimento de símbolos e tipos definidos em arquivos relacionados
- [ ] Precisão de sugestões > 60% (proporção de sugestões aceitas pelo usuário)
- [ ] Latência P95 < 2s mesmo com contexto multi-arquivo
- [ ] Cache inteligente de contexto do projeto (não re-processa arquivos a cada keystroke)
- [ ] Funciona para projetos Python, JS/TS, Java (com build systems: pip, npm, maven/gradle)
- [ ] Configurável: profundidade máxima de arquivos analisados

#### Notas Técnicas

- **Dependências:** FR-V2-01 (autocomplete FIM), FR-V2-03 (multi-linguagem)
- **Componentes envolvidos:** Continue.dev, Ollama, CLI (análise de projeto), sistema de arquivos
- **Complexidade:** Alta — construção de contexto multi-arquivo; RAG leve (retrieval de símbolos)
- **Implementação:** Indexação de símbolos do projeto (AST parsing leve); injeção de contexto relevante no prompt FIM

---

### FR-V3-05 — Plugin VS Code Oficial

**Fase:** V3
**Sprint:** 10
**Prioridade:** Should
**Story Points:** 8

#### Descrição

Como **Joana**, que quer uma experiência integrada sem configurar extensões separadas, eu quero instalar uma extensão oficial do CodeMin no VS Code (1 clique) que abstraia o Continue.dev e ofereça comandos dedicados (CodeMin: Chat, CodeMin: Explicar, CodeMin: Revisar) para usar tudo em um só lugar.

#### Critérios de Aceitação

- [ ] Extensão publicada no VS Code Marketplace
- [ ] Comandos disponíveis na paleta (Ctrl+Shift+P): `CodeMin: Chat`, `CodeMin: Explicar`, `CodeMin: Revisar`, `CodeMin: Gerar Testes`, `CodeMin: Status`
- [ ] Painel lateral com chat integrado e status do CodeMin
- [ ] Botão de status na barra inferior (mostra modelo ativo, status da conexão)
- [ ] Instalação 1 clique (não requer config manual do Continue.dev)
- [ ] Detection automática do CodeMin CLI (se não instalado, sugere instalação)
- [ ] Suporte a temas claro/escuro
- [ ] Extensão publicada no VS Code Marketplace (acessível publicamente)

#### Notas Técnicas

- **Dependências:** FR-V3-04 (multi-arquivo), FR-MVP-03 a FR-MVP-09 (funcionalidades core)
- **Componentes envolvidos:** VS Code Extension API, TypeScript, Webview (painel lateral), Ollama API
- **Complexidade:** Alta — extensão VS Code completa com Webview, comandos, integração com terminal
- **Template:** Yo Code generator; VS Code Extension API v1.85+
- **Observação:** O plugin abstrai o Continue.dev — o usuário não precisa instalar extensão separada

---

### FR-V3-06 — Suporte JetBrains

**Fase:** V3
**Sprint:** 11
**Prioridade:** Should
**Story Points:** 5

#### Descrição

Como **Carlos**, que usa IntelliJ IDEA para desenvolvimento Java na startup, eu quero usar o CodeMin via Continue.dev nas IDEs JetBrains (IntelliJ, PyCharm, GoLand) para ter assistência de código sem sair do meu ambiente de trabalho principal.

#### Critérios de Aceitação

- [ ] Autocomplete FIM funcional no IntelliJ IDEA via Continue.dev
- [ ] Chat contextual funcional no IntelliJ IDEA via Continue.dev
- [ ] Suporte ao PyCharm (Python) e GoLand (Go) via Continue.dev
- [ ] Configuração gerada pelo `codemin install` compatível com JetBrains + Continue.dev
- [ ] `codemin doctor` valida configuração para JetBrains
- [ ] Documentação específica para instalação/configuração em JetBrains
- [ ] Latência < 3s autocomplete, < 15s chat nas IDEs JetBrains

#### Notas Técnicas

- **Dependências:** FR-V2-01 (autocomplete FIM), FR-V2-03 (multi-linguagem), FR-V3-01 (8+ linguagens)
- **Componentes envolvidos:** JetBrains IDEs, Continue.dev (plugin JetBrains), Ollama
- **Complexidade:** Média — depende da maturidade do Continue.dev para JetBrains; testes específicos por IDE
- **Observação:** Continue.dev para JetBrains está em desenvolvimento ativo — verificar compatibilidade na data de implementação

---

### FR-V3-07 — Interface Web Local (GUI)

**Fase:** V3
**Sprint:** 11
**Prioridade:** Could
**Story Points:** 8

#### Descrição

Como **Joana**, que prefere interfaces gráficas ao terminal, eu quero acessar uma interface web local (`http://localhost:8080`) com chat, configurações e gerenciamento de modelos para interagir com o CodeMin visualmente.

#### Critérios de Aceitação

- [ ] Interface web responsiva acessível em `http://localhost:8080`
- [ ] Chat com histórico de conversa
- [ ] Gerenciamento visual de modelos (listar, baixar, remover, alternar)
- [ ] Exibição de status (modelo ativo, RAM, CPU)
- [ ] Tema claro/escuro

#### Notas Técnicas

- **Dependências:** FR-V3-03 (gerenciamento de modelos)
- **Componentes envolvidos:** Servidor web embutido (Express/Fastify), React/Vue para frontend, Ollama API
- **Complexidade:** Alta — frontend completo com comunicação em tempo real
- **Status:** Placeholder — detalhar durante refinement

---

### FR-V3-08 — Fine-tuning para Codebase Específica

**Fase:** V3
**Sprint:** 12
**Prioridade:** Could
**Story Points:** 13

#### Descrição

Como **Carlos**, que trabalha com código proprietário e quer um modelo especializado na base da empresa, eu quero executar `codemin finetune <repo>` para fazer fine-tuning do modelo base na minha codebase, gerando adapters LoRA que melhoram a precisão das sugestões.

#### Critérios de Aceitação

- [ ] `codemin finetune <dir>` preprocessa repositório git local
- [ ] Gera adapters LoRA treinados na codebase do usuário
- [ ] Funde adapters ao modelo base para uso
- [ ] Fine-tuning completo em < 2 horas em hardware-alvo

#### Notas Técnicas

- **Dependências:** FR-V3-02 (múltiplos modelos)
- **Complexidade:** Muito Alta — requer integração com llama.cpp fine-tuning; consome muitos recursos
- **Status:** Placeholder — requer spike técnico antes de detalhar
- **Risco:** Fine-tuning em CPU pode ser extremamente lento; validar viabilidade

---

### FR-V3-09 — Benchmark de Desempenho (`codemin bench`)

**Fase:** V3
**Sprint:** 11
**Prioridade:** Could
**Story Points:** 3

#### Descrição

Como **Ana**, que gosta de métricas e otimização, eu quero executar `codemin bench` para medir tokens/s, latência e uso de RAM/CPU do modelo ativo para comparar desempenho entre modelos e configurations.

#### Critérios de Aceitação

- [ ] `codemin bench` executa benchmark do modelo ativo
- [ ] Mede: tokens/s (geração), latência (primeiro token), pico de RAM, uso de CPU
- [ ] `codemin bench --compare` compara entre modelos instalados
- [ ] Gera relatório formatado no terminal e opcionalmente em JSON
- [ ] Benchmark completo em < 5 minutos

#### Notas Técnicas

- **Dependências:** FR-V3-02 (múltiplos modelos)
- **Componentes envolvidos:** CLI, Ollama API
- **Complexidade:** Baixa — chamadas sequenciais à API com medição de tempo
- **Status:** Placeholder — detalhar durante refinement

---

## Resumo de Estimativas

### Por Fase

| Fase | Histórias | Total SP | Média SP |
|------|-----------|----------|----------|
| MVP  | 10        | 44       | 4.4      |
| V2   | 9         | 47       | 5.2      |
| V3   | 9         | 58       | 6.4      |
| **Total** | **28** | **149** | **5.3** |

### Por Prioridade

| Prioridade | Histórias | Total SP |
|------------|-----------|----------|
| Must       | 13        | 73       |
| Should     | 8         | 43       |
| Could      | 7         | 33       |

### Por Sprint

| Sprint | Histórias | SP  |
|--------|-----------|-----|
| 1      | 5         | 21  |
| 2      | 5         | 23  |
| 3      | 2         | 16  |
| 4      | 2         | 10  |
| 5      | 3         | 13  |
| 6      | 2         | 8   |
| 7      | 1         | 8   |
| 8      | 2         | 10  |
| 9      | 1         | 8   |
| 10     | 1         | 8   |
| 11     | 2         | 11  |
| 12     | 1         | 13  |

---

_Arquivo gerado por 🌊 River (AIOX Scrum Master) em 2026-07-09_
_Fonte: PRD v1.0.0 — Morgan (AIOX Product Manager)_

---

## Revisão do PO

**Revisor:** 🎯 Pax (AIOX Product Owner)
**Data da Revisão:** 2026-07-09
**Versão Revisada:** 1.1.0

### Status das Histórias

| Status | Quantidade | IDs |
|--------|-----------|-----|
| ✅ Approved | 22 | FR-MVP-01, FR-MVP-02, FR-MVP-03, FR-MVP-04, FR-MVP-06, FR-MVP-07, FR-MVP-08, FR-MVP-DOC-01, FR-V2-01, FR-V2-02, FR-V2-03, FR-V2-04, FR-V2-05, FR-V2-06, FR-V2-07, FR-V2-09, FR-V3-01, FR-V3-02, FR-V3-03, FR-V3-04, FR-V3-06, FR-V3-09 |
| 🔄 Needs Refinement | 4 | FR-MVP-05 (AC vago sobre versão do Continue.dev), FR-MVP-09 (comando `codemin explain` não previsto no PRD MVP), FR-V3-05 (AC de "1.000 instalações" movido para nota — verificar), FR-V3-07 (placeholder — detalhar) |
| ⚠️ Needs Discussion | 2 | FR-V3-08 (Fine-tuning: viabilidade técnica em CPU; requer spike), FR-V2-08 (Update: confirmar se entra como Could no V2 ou adia para V3) |
| 📌 Placeholder | 3 | FR-V3-07, FR-V3-08, FR-V3-09 (adicionados como esboço; detalhar durante refinement do V3) |

### Principais Correções Realizadas

| # | Correção | Tipo | Detalhe |
|---|----------|------|---------|
| 1 | **FR-MVP-10 → FR-MVP-DOC-01** | ID Conflict | PRD reserva FR-MVP-10 para "Autocomplete FIM (Won't Have)". A história de documentação agora tem ID próprio. |
| 2 | **FR-V2-08 → FR-V2-09** | ID Mismatch | PRD mapeia FR-V2-08 = Update (Could) e FR-V2-09 = Fallback (Could). Corrigido para alinhar com PRD. |
| 3 | **FR-V2-08 adicionada** | História faltante | Update (`codemin update`) não constava nas histórias originais. Adicionada como Could, 3 SP. |
| 4 | **FR-V3-07, FR-V3-08, FR-V3-09 adicionadas** | Histórias faltantes | GUI Web, Fine-tuning e Benchmark estavam no PRD mas não nas histórias. Adicionadas como placeholders. |
| 5 | **FR-MVP-05 AC corrigido** | AC vago | "Compatível com Continue.dev versão atual" → "Compatível com Continue.dev v0.8.x ou superior" |
| 6 | **FR-MVP-09 AC corrigido** | AC impreciso | Comando `codemin explain` não está no CLI do PRD MVP. Movido para "bônus se tempo permitir". |
| 7 | **FR-V3-05 AC corrigido** | Métrica de negócio | "1.000+ instalações" não é AC de desenvolvimento. Substituído por "Extensão publicada no Marketplace". |

### Recomendações para o Sprint Planning

1. **Sprint 1** (21 SP, 5 histórias): Carga OK. Foco total na fundação — instalação, download, configs e health check. A ordem está correta: tudo que o Sprint 2 precisa deve estar pronto.

2. **Sprint 2** (23 SP, 5 histórias): **É a sprint mais crítica.** Chat contextual (8 SP) e geração de código (5 SP) são o coração do MVP. Se houver risco de estouro, considerar mover FR-MVP-09 (Explicação — Could) ou FR-MVP-DOC-01 (Documentação — Could) para o backlog.

3. **Sprint 3** (16 SP, 2 histórias): Ambas são Must e tecnicamente complexas (FIM + Multi-linguagem). 16 SP para 2 histórias é aceitável, mas monitorar de perto — se FR-V2-01 atrasar, impacta FR-V2-02 e FR-V2-04.

4. **Sprint 6** (8 SP, 2 histórias): Leve. Ideal para absorver débitos técnicos ou histórias Carry-over de sprints anteriores.

5. **Sprints 11-12** (24 SP, 3 histórias): Pesado. FR-V3-08 (Fine-tuning) com 13 SP é a maior história do backlog. Sugiro:
   - Mover FR-V3-08 para backlog de longo prazo até o spike técnico ser concluído
   - Avaliar se FR-V3-07 (GUI) pode ser despriorizada se o Plugin VS Code (FR-V3-05) entregar interface suficiente

### Riscos Identificados

| ID | Risco | Impacto | Ação |
|----|-------|---------|------|
| R01 | **Dependência do Continue.dev** para FIM (V2) | Se Continue.dev mudar API ou tiver bugs, bloqueia FR-V2-01 a FR-V2-07 | Testar integração cedo; manter fallback para chat via OpenCode |
| R02 | **Latência de autocomplete < 3s** | Pode ser difícil de atingir em CPUs fracas com modelo 7B | Validar em hardware-alvo (Ryzen 5, 16 GB) antes do Go/No-Go V2 |
| R03 | **FR-V3-08 Fine-tuning em CPU** | Pode ser inviável (dias de treinamento em CPU) | **Spike técnico necessário** antes de incluir no planejamento |
| R04 | **Download de 4.7 GB do modelo** | Usuários com conexão lenta podem abandonar | Garantir resume funcionando; considerar opção de download via torrent |
| R05 | **Sobrecarga de SP no Sprint 2** (23 SP) | Time pode não entregar tudo | Preparar plano de contingência: realocar FR-MVP-09 e FR-MVP-DOC-01 |

### Dívida Técnica Aceita

| Item | Justificativa | Plano |
|------|---------------|-------|
| **FR-MVP-10 omitido (FIM Autocomplete)** | Explicitamente Won't Have no MVP por latência e complexidade | Implementar como FR-V2-01 no V2 |
| **Testes de integração end-to-end** | Não detalhados em nenhuma história; CI/CD está implícito | Adicionar história de infraestrutura de testes no Sprint 1 ou 2 |
| **FR-V3-08 Fine-tuning como placeholder** | Requer validação técnica antes de detalhar | Spike técnico agendado para Sprint 10 |
| **4 histórias Could não priorizadas** (FR-V2-08, V3-07, V3-08, V3-09) | Baixo valor relativo vs. esforço | Manter no backlog; reavaliar no V3 planning |

### Notas Finais

- **Personas:** Distribuição equilibrada. Joana (8), Carlos (8), Ana (9) — todas bem representadas.
- **Estimativas:** Média de 5.3 SP por história — saudável. Nenhuma história individual > 13 SP (FR-V3-08 é outlier e precisa de validação).
- **Dependências:** Mapeadas corretamente nas notas técnicas de cada história. Recomendo criar um **gráfico de dependências** visual no board do sprint.
- **Próximo passo:** Levar as 4 histórias "Needs Refinement" para a próxima sessão de refinement. As 2 "Needs Discussion" devem ser discutidas com o time e o PM antes do Sprint Planning.

---

_Revisão realizada por 🎯 Pax (AIOX Product Owner) em 2026-07-09_
