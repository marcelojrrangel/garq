$env:GOROOT = "L:\sdk\go1.24.0"
$env:Path = "L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"
$env:CC = "L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"
$env:CGO_ENABLED = '1'
cd D:\projetos\aiox-core\garq

Write-Host "=== Gerando ícones ==="
go run ./tools/genicons/

Write-Host "=== Compilando recursos (.rc -> .syso) ==="
& "L:\sdk\mingw64\bin\windres.exe" -i cmd/walk/resources.rc -o cmd/walk/resources.syso
if ($LASTEXITCODE -ne 0) { Write-Host "windres falhou!"; exit 1 }

Write-Host "=== Compilando walk ==="
go build -o L:\sdk\garq-new.exe ./cmd/walk/
if ($LASTEXITCODE -ne 0) { Write-Host "Build falhou!"; exit 1 }

Write-Host "=== Iniciando Garq ==="
L:\sdk\garq-new.exe
