# Revisão de Código — Backend Garq

**Data:** 2026-07-11  
**Escopo:** `cmd/app/`, `internal/`, `compress/` (excluindo `cmd/walk/`)  
**Analistas:** @aiox-architect (🏛️ Aria) + @aiox-dev (💻 Dex)  
**Coordenador:** @aiox-orchestrator (🔄)

---

## Visão Geral

29 achados no total — 4 críticos, 7 altos, 14 médios, 4 baixos.

---

## 🔴 Crítico

### C1. Command Injection via 7z `@listfile` e injeção de argumentos

**Arquivo:** `compress/7z_cli.go:35`  
**Justificativa:** `exec.CommandContext` protege contra shell injection, mas o 7-Zip aceita sintaxe `@arquivo` (lê lista de paths de um arquivo) e argumentos iniciados com `--`. `sources` e `dest` vêm do payload JSON do usuário via HTTP API. Um atacante pode enviar `"sources": ["@C:\\Users\\victim\\passwords.txt"]` para exfiltrar arquivos ou `"-o../../etc"` para controlar o destino.

**Sugestão:**
```go
for _, s := range sources {
    if strings.HasPrefix(s, "@") || strings.HasPrefix(s, "-") {
        return fmt.Errorf("invalid source path: %s", s)
    }
}
```
Validar também `dest` e `archive` contra caracteres suspeitos.

---

### C2. Path Traversal sem restrição — Filesystem API

**Arquivo:** `internal/api/filesystem.go:41`, `internal/api/http.go:61`  
**Justificativa:** `ListDirectory(path)` recebe `path` diretamente da query string HTTP (`?path=..\..\Windows\System32`) e chama `os.ReadDir(path)` sem qualquer sanitização. Qualquer diretório do sistema pode ser listado. A API HTTP está em `127.0.0.1:19876`, mas ainda assim é um vetor de ataque.

**Sugestão:**
```go
func safePath(root, userPath string) (string, error) {
    abs, err := filepath.Abs(filepath.Join(root, userPath))
    if err != nil { return "", err }
    if !strings.HasPrefix(abs, filepath.Clean(root)) {
        return "", errors.New("path outside allowed root")
    }
    return abs, nil
}
```

---

### C3. Path Traversal em Jobs — Delete, Copy, Move, Compress

**Arquivo:** `internal/worker/worker.go:183-196, 377-400`  
**Justificativa:** Todos os `run*Job` recebem `sources`, `dest`, `archive` do `payload map[string]any` que veio do banco (enfileirado via HTTP API). `runDeleteJob` pode chamar `os.RemoveAll("C:\\Windows\\System32")`. `runCopyJob` pode copiar qualquer arquivo para qualquer lugar. Sem validação de diretório permitido.

**Sugestão:**
- Adicionar `validatePath(root, path)` no início de cada `run*Job`.
- Definir uma "raiz permitida" configurável (`GARQ_ALLOWED_ROOT`).
- Rejeitar operações em paths que contenham `..`, ou resolvê-los para absoluto e verificar prefixo.

---

### C4. InitDB destrói todos os jobs no startup

**Arquivo:** `internal/db/db.go:18-23`  
**Justificativa:** Em toda inicialização, `DELETE FROM jobs` remove todos os jobs pendentes/running. Jobs de longa duração (ex.: compressão de 50GB) são perdidos se o servidor reiniciar. O reset do `sqlite_sequence` é silenciosamente ignorado (`_, _ = db.Exec`).

**Sugestão:**
- Remover o `DELETE` automático.
- Se estado limpo for desejado, usar flag `--clean` ou `GARQ_CLEAN_START=true`.
- No mínimo, logar quantos jobs foram removidos.

---

## 🟡 Alto

### A1. Duas portas HTTP concorrentes e conflitantes

**Arquivo:** `cmd/app/main.go:111` + `internal/api/http.go:11,32-42`  
**Justificativa:** Duas interfaces HTTP: porta 8080 (handlers manuais em `main.go`) e porta 19876 (`StartHTTPServer` em `http.go`). A segunda fica em goroutine solta sem graceful shutdown, sem `*http.Server`, sem tratamento de erro. Duas superfícies de ataque, duas fontes de confusão.

**Sugestão:** Unificar em `cmd/app/main.go` com um router (chi, gin, ou `http.ServeMux`), prefixos claros `/api/v1/jobs/`, `/api/v1/filesystem/`. Remover `StartHTTPServer` e `http.go` handlers.

---

### A2. Nenhuma interface — acoplamento concreto total

**Arquivo:** `internal/worker/worker.go:15-17`, `internal/api/wails.go:15-16`  
**Justificativa:** Zero interfaces definidas no projeto. `worker.go` depende de `compress.CompressManyCtx` e `copyimpl.CopyFile` (concretas). `api.API` expõe `*sql.DB` bruto. Impossível testar worker, substituir implementações, ou isolar componentes.

**Sugestão:** Definir interfaces nos consumidores:
```go
type JobStore interface { ... }
type Compressor interface { CompressManyCtx(...) error; ExtractCtx(...) error }
type FileCopier interface { CopyFile(src, destDir string) error }
```

---

### A3. Contexto ignorado — workers não podem ser shutdown graceful

**Arquivo:** `internal/worker/worker.go:127`, `cmd/app/main.go:26`  
**Justificativa:** `registerJob` usa `context.Background()` em vez de receber contexto do caller. `StartWorkerPool` não aceita `ctx`. Quando `main()` recebe SIGTERM, workers continuam rodando. Não há propagação de cancelamento.

**Sugestão:**
```go
func StartWorkerPool(ctx context.Context, n int, conn *sql.DB)
// workerLoop usa ctx como base, não context.Background()
```

---

### A4. `map[string]any` em toda parte — type safety zero

**Arquivo:** `internal/db/db.go:85`, `internal/worker/worker.go:159-315`, `internal/api/wails.go:84`  
**Justificativa:** Payloads trafegam como `map[string]any`. `getString`, `getStringSlice` fazem type assertion em runtime. Erros de tipo só aparecem em produção. O compilador não pega campos renomeados ou tipos incorretos.

**Sugestão:** Structs tipadas por job:
```go
type CopyPayload struct {
    Sources  []string `json:"sources"`
    Dest     string   `json:"dest"`
    Conflict string   `json:"conflict"`
}
```
Usar `json.Unmarshal` no struct correto baseado no `typ`.

---

### A5. Pause/Resume com mecanismo frágil de dois sinais no mesmo canal

**Arquivo:** `internal/worker/worker.go:50-93`  
**Justificativa:** `checkPause` usa um `chan struct{}` com buffer=2 para codificar dois estados (pause = 1º sinal, resume = 2º sinal). Race condition: se `PauseJob` envia o sinal entre unlock e select, ou se `ResumeJob` é chamado antes de `PauseJob`, o worker pode travar. `defer delete(pauseChs, jobID)` no cancel pode tornar o canal órfão.

**Sugestão:** Usar `sync.Mutex` + variável booleana + polling com `time.Sleep` pequeno, ou `sync.Cond`:
```go
type pauseState struct {
    mu     sync.Mutex
    paused bool
    cond   *sync.Cond
}
```

---

### A6. Regex compilado a cada linha de output do 7z

**Arquivo:** `compress/7z_cli.go:52`  
**Justificativa:** `regexp.MustCompile` dentro de `parseProgress` é chamado para cada linha de saída do 7z (potencialmente centenas). Compilação de regex é cara, causando pressão no GC e lentidão.

**Sugestão:** Mover para variável package-level:
```go
var progressRe = regexp.MustCompile(`(\d{1,3})%`)
```

---

### A7. `srv.Close()` em vez de `srv.Shutdown()` — sem graceful drain

**Arquivo:** `cmd/app/main.go:126`  
**Justificativa:** `Close()` fecha listener e aborta conexões ativas imediatamente. Requisições em andamento são interrompidas abruptamente.

**Sugestão:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Printf("shutdown error: %v", err)
}
```

---

## 🟢 Médio

### M1. Erros de `UpdateJobStatus` silenciados com `_ =`

**Arquivo:** `internal/worker/worker.go:148,152,155,189,199,238,277,359,372,405`  
**Justificativa:** Múltiplos `_ = db.UpdateJobStatus(...)` descartam erros. Se falha na atualização (DB desconectado, disco cheio), frontend nunca vê progresso nem conclusão.

**Sugestão:** Logar o erro:
```go
if err := db.UpdateJobStatus(...); err != nil {
    log.Printf("failed to update job %d: %v", jobID, err)
}
```

---

### M2. Busy polling no DB — 4 queries/segundo quando ocioso

**Arquivo:** `internal/worker/worker.go:116-118`  
**Justificativa:** 4 workers rodam `BEGIN + SELECT + UPDATE + COMMIT` a cada 1s mesmo sem jobs. Em WAL mode o impacto é menor, mas ainda desperdício de I/O.

**Sugestão:** Usar canal de notificação. `EnqueueJob` envia para `chan struct{}` que workers escutam. `select` com `time.After(3s)` como fallback.

---

### M3. `rows.Scan` falha silenciosamente em `GetJobs`

**Arquivo:** `internal/api/wails.go:97-98`  
**Justificativa:** Se `Scan` falha em uma linha, `continue` a descarta sem log. Usuário recebe resultado incompleto sem saber.

**Sugestão:**
```go
if err := rows.Scan(...); err != nil {
    log.Printf("scan error for row: %v", err)
    continue
}
```

---

### M4. Nenhuma verificação de `rows.Err()` após loop

**Arquivo:** `internal/api/wails.go:91-118`  
**Justificativa:** Após `for rows.Next()`, nunca se chama `rows.Err()`. Erro intermitente do DB (disco cheio, temp table issue) é perdido.

**Sugestão:**
```go
if err := rows.Err(); err != nil {
    return nil, err
}
```

---

### M5. `log.Fatal` em goroutine do servidor HTTP

**Arquivo:** `cmd/app/main.go:116-117`  
**Justificativa:** `log.Fatal` dentro da goroutine de `ListenAndServe` mata o processo inteiro se o servidor falhar depois de iniciado.

**Sugestão:**
```go
log.Printf("HTTP server error: %v", err)
```
E sinalizar shutdown.

---

### M6. `copyFileSimple` — double `out.Close()`

**Arquivo:** `internal/worker/worker.go:421-431`  
**Justificativa:** `defer out.Close()` + `return out.Close()` = duas chamadas. A segunda retorna `os.ErrClosed`, e o erro da primeira (verdadeiro) é perdido.

**Sugestão:** Remover `defer`, usar apenas o `Close()` explícito, capturar erro.

---

### M7. `fmt.Println` de debug em produção

**Arquivo:** `internal/api/filesystem.go:22,24,30,33,36`  
**Justificativa:** 4+ chamadas a `fmt.Println` para debug (ListRoots checa drives A-Z e loga cada uma). Em servidor HTTP, polui stdout desnecessariamente.

**Sugestão:** Remover ou usar `log.Printf` condicional.

---

### M8. Dead field `API.Ctx` nunca usado

**Arquivo:** `internal/api/wails.go:16`  
**Justificativa:** `Ctx context.Context` declarado mas nunca setado nem lido. Enganoso.

**Sugestão:** Remover.

---

### M9. Erro de handler HTTP ignorado em `http.go`

**Arquivo:** `internal/api/http.go:41`  
**Justificativa:** `go http.ListenAndServe(addr, mux)` — retorno de erro descartado. Se porta ocupada, servidor falha silenciosamente.

**Sugestão:**
```go
go func() {
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Printf("HTTP API server error: %v", err)
    }
}()
```

---

### M10. `Content-Type` não definido em handlers JSON

**Arquivo:** `cmd/app/main.go:48,69,91,103`, `internal/api/http.go:47,60`  
**Justificativa:** `json.NewEncoder(w).Encode()` sem `w.Header().Set("Content-Type", "application/json")` antes.

**Sugestão:** Adicionar header antes de cada encode.

---

### M11. `EnqueueJob` sem validação de tipo

**Arquivo:** `internal/db/db.go:44-49`  
**Justificativa:** `typ` não é validado. Qualquer string vai para DB e depois o worker cai em `default: unsupported job type`. Erro só aparece na execução.

**Sugestão:**
```go
var validTypes = map[string]bool{"copy": true, "compress": true, ...}
if !validTypes[typ] { return 0, fmt.Errorf("invalid job type: %s", typ) }
```

---

### M12. Erro interno vaza para o cliente HTTP

**Arquivo:** `internal/api/http.go:50-51,70-71`  
**Justificativa:** `json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})` expõe detalhes internos (paths, permissões).

**Sugestão:** Logar internamente, retornar mensagem genérica.

---

### M13. Lazy DLL carregado a cada `CopyFile` (Windows)

**Arquivo:** `internal/copy/copy_windows.go:55-56`  
**Justificativa:** `NewLazySystemDLL` + `NewProc` em cada chamada de `CopyFile`. Em cópia de muitos arquivos, overhead desnecessário.

**Sugestão:** `sync.Once` para carregar DLL uma vez no pacote.

---

### M14. `InitDB` hardcoded path, sem flags

**Arquivo:** `cmd/app/main.go:18`, `internal/db/db.go:10`  
**Justificativa:** `dbPath := "garq.db"` — hardcoded, sem CLI flag ou env. Em produção, impossível configurar caminho.

**Sugestão:** Usar `os.Getenv("GARQ_DB_PATH")` com fallback `"garq.db"`.

---

## ⚪ Baixo

### B1. Worker pool size hardcoded

**Arquivo:** `cmd/app/main.go:26`  
**Justificativa:** `worker.StartWorkerPool(4, dbConn)` — número fixo de 4 goroutines.

**Sugestão:** `os.Getenv("GARQ_WORKER_COUNT")` com fallback para `runtime.NumCPU()`.

---

### B2. Constante `"replace"` duplicada 8x

**Arquivo:** `internal/api/wails.go` (5×) + `internal/worker/worker.go` (5×)  
**Justificativa:** `if conflict == "" { conflict = "replace" }` aparece em 10 lugares.

**Sugestão:** Constante package-level.

---

### B3. `EnqueueJob` marshala `any` para JSON → string — ineficiente

**Arquivo:** `internal/db/db.go:44-49`  
**Justificativa:** `json.Marshal(payload)` + `string(b)` — alocação extra.

**Sugestão:** Aceitar `[]byte` ou fazer marshaling no caller.

---

### B4. Handlers HTTP repetitivos no `main.go`

**Arquivo:** `cmd/app/main.go:30-104`  
**Justificativa:** 4 closures quase idênticos (~20 linhas cada).

**Sugestão:** Factory function `jsonJobHandler(handlerFn) http.HandlerFunc`.

---

## 🔧 Plano de Ação

| Fase | Itens | Esforço estimado |
|------|-------|------------------|
| **Fase 1 — Correções Imediatas** | C1, C2, C3, C4 | 1-2 dias |
| **Fase 2 — Qualidade** | A1, A4, A5, A6, A7 | 2-3 dias |
| **Fase 3 — Arquitetura** | A2, A3 + interfaces + DIP + OCP | 2-3 dias |
| **Fase 4 — Melhorias** | M1 a M14 | 2-3 dias |

**Total estimado:** 7-11 dias de desenvolvimento.

---

## Métricas do Backend

| Métrica | Valor |
|---------|-------|
| Arquivos analisados | 14 |
| Linhas de código | ~1,350 |
| Interfaces definidas | 0 |
| `map[string]any` usos | 10+ |
| Erros silenciados (`_ =`) | 12 |
| Funções > 50 linhas | 2 (`workerLoop`, `find7zExecutable`) |

---

## Conclusão

O backend do Garq é funcional e faz o que se propõe, mas tem fragilidades sérias de segurança (🔴 Path traversal + command injection) e arquiteturais (🔴 God Object `API`, zero interfaces, DIP violado). Os itens críticos devem ser resolvidos antes de qualquer deploy em produção. As melhorias arquiteturais da Fase 3 são recomendadas para evitar que a dívida técnica impeça evoluções futuras.
