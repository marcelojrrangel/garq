# Garq - Pontos de Atenção e Melhorias Futuras

Data: 04/07/2026
Status: Backlog

---

## 1. Funcionalidades Faltantes

- **[X] Operações de move/delete** — Implementadas via Wails API + workers
- **[X] Cancelamento de jobs** — `context.Context` no worker loop + `CancelJob` API
- **[X] Exibição de permissões de arquivo** — Coluna "Permissões" no listing (mode do fs)
- **[X] Configuração de conflito (replace/skip/rename)** — Modal + localStorage + workers

## 2. Bugs Conhecidos

- **[X] `CompressMany` tem bug** — Corrigido: caminho CGO agora usa fallback CLI
- **[X] Bindings Wails desatualizados** — ListRoots e ListDirectory adicionados
- **[X] Ctrl+C/X/V não funciona** — Corrigido: escopo global + blur de inputs

## 3. Qualidade de Código

- **[X] Testes unitários Go** — `_test.go` para db, copy, compress (17 testes)
- **[X] `cmd/wails/main.go` é duplicado** — Removido
- **[X] `internal/api/api.go` é código morto** — Deletado
- **[X] `schema.sql` sincronizado** — Espelha `mustReadSchema()` exatamente

## 4. UX/UI

- **[X] UI Explorer 11** — Toolbar, breadcrumb, nav pane, content list, details pane, status bar
- **[X] Configurações de conflito** — Modal com select por operação
- **[X] Permissões de arquivo** — Coluna "Permissões" no content list
- **[ ] Progresso de compactação/extração é grosseiro** — 0→0.5→1.0

## 5. Infraestrutura

- **[ ] DB limpa jobs no startup** — Histórico não persiste (intencional para dev)
- **[X] `go.mod` pede Go 1.25** — Downgraded para Go 1.24
- **[X] `garq.exe` e `garq.db` no .gitignore** — Adicionado

## 6. Placeholder/SDK (Removidos)

- **[X] `compress/7z_wrapper.c`** — Deletado (placeholder inútil)
- **[X] `compress/7z.go`** — Deletado (CGO SDK inexistente)
- **[X] `scripts/build_7zwrapper.ps1`** — Deletado (só imprimia instruções)
