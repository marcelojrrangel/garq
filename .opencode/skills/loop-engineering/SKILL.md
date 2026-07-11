# Role: Loop Engineering Architect

Ative este skill quando o usuário solicitar "loop engineering", "loop architect", "auto-correção", "ciclo TDD", "desenvolvimento autônomo" ou "modo loop".

## Fluxo Obrigatório

1. **Contexto** — Leia `@roadmap.md`, `@lessons.md` e `@state.json` da raiz do projeto
2. **Analisar** — Identifique a próxima tarefa no roadmap ou o erro atual
3. **Codificar** — Implemente a solução
4. **Testar** — Execute `npm test` (ou pytest, dotnet test, etc conforme o projeto)
5. **Auto-correção** — Se falhar: registre o erro em `@lessons.md`, refatore e reteste. Máx 2 tentativas, depois pergunte ao humano.
6. **Atualizar** — Marque a tarefa como concluída no `@roadmap.md`

## Regras de Ouro

- Se não existir suíte de testes, peça ao humano para criar antes de iniciar o loop
- `@lessons.md` é o log de experiência — evita repetir erros anteriores
- `@state.json` mantém o estado atual do loop (fase, tarefa, tentativas)
- Se 2 correções consecutivas falharem no mesmo código, PARE e pergunte

## Arquivos de Controle (criar na raiz do projeto se não existirem)

- `roadmap.md` — tarefas em ordem, marcadores `- [ ]` / `- [x]`
- `lessons.md` — lições aprendidas, bugs recorrentes, decisões
- `state.json` — `{ "phase": "...", "currentTask": "...", "attempts": 0 }`
