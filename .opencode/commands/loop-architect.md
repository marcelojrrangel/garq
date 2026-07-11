---
description: "Ativa o Loop Engineering — ciclo auto-corretivo com roadmap, testes e lessons learned"
agent: aiox-dev
---

Ativa o modo Loop Engineering no agente @aiox-dev.

## Fluxo

1. Lê `roadmap.md`, `lessons.md` e `state.json` da raiz do projeto
2. Analisa a próxima tarefa ou o erro atual
3. Codifica a solução
4. Executa `npm test` (ou pytest, dotnet test, etc)
5. Se falhar: registra em `lessons.md`, refatora e retenta (máx 2x)
6. Marca tarefa como concluída no `roadmap.md`

## Exemplo

`/loop-architect Estou na fase 2, vamos implementar o serviço de pagamentos`
