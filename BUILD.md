Build and dev instructions

1) Build Go backend (Windows example)
   - Open PowerShell
   - Set CGO for sqlite + cgo features: $env:CGO_ENABLED = '1'
   - go build ./...
   - For Wails bootstrap source only: go build -tags wails ./cmd/wails

2) 7z wrapper (temporary)
   - Default build uses 7z CLI via exec.Command (compress/7z_cli.go).
   - For zero-install on user machine, bundle 7z binary with the app:
     - <app-dir>\7z\7z.exe (Windows) OR <app-dir>\7z\7zz
   - The app searches in this order: GARQ_7Z_PATH -> bundled binary -> PATH -> standard Windows install paths.
   - SDK binding is optional behind build tag sevenzip_sdk + cgo (compress/7z.go).
   - For SDK mode, build native lib (7z_wrapper) and compile with: go build -tags sevenzip_sdk ./...

3) Using Wails
   - In this repo, a minimal embedded frontend is in cmd/wails/frontend/index.html.
   - Build bootstrap: go build -tags wails ./cmd/wails
   - Run desktop app (dev): go run -tags "wails,dev" ./cmd/wails
   - Build desktop app (prod): go build -tags "wails,production" ./cmd/wails

4) Optimizations
   - Implement OS-specific fast-copy syscalls in internal/copy/*.go (copy_file_range, sendfile, CopyFile2, clonefile).
   - Profile with large file transfers and many small files; tune worker pool and buffer sizes.

5) HTTP test endpoints (dev mode, cmd/app)
   - POST /jobs/copy
     body: {"sources":["C:\\input\\a.txt","D:\\other\\b.txt"],"dest":"E:\\target"}
   - POST /jobs/compress
     body: {"sources":["E:\\target\\a.txt","E:\\target\\b.txt"],"dest":"E:\\archives\\files.7z"}
   - POST /jobs/extract
     body: {"archive":"E:\\archives\\folder.7z","dest":"E:\\extracted"}
   - GET /jobs

