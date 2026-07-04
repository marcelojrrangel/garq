Garq - Esqueleto de gerenciador de arquivos (Go + Wails)

Este repositório contém um esqueleto mínimo para um backend de gerenciador de arquivos multiplataforma em Go com:
- Pool de workers para lidar com cópias de arquivos
- Fila de trabalho persistente usando SQLite
- Endpoints HTTP mínimos (a serem substituídos por bindings Wails)
- Exemplo de wrapper cgo que executa o 7z via shell (placeholder para integração do SDK 7z via bindings)

Notas de construção:
- Requer Go >= 1.20 com CGO habilitado para sqlite3
- Para compilar: defina CGO_ENABLED=1 e execute `go build ./...`
- A integração do frontend Wails não está incluída neste esqueleto; use o Wails para vincular a camada HTTP/API a uma aplicação desktop.

Próximos passos:
- Substituir os handlers HTTP por bindings Wails (runtime wails)
- Implementar arquivos de syscall de cópia rápida específicos para cada SO com build tags
- Substituir o wrapper shell do 7z por um binding compilado do SDK quando estiver pronto
