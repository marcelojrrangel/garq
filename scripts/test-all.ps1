$env:GOROOT = "L:\sdk\go1.24.0"
$env:Path = "L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;$env:Path"
$env:CC = "L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe"
$env:CGO_ENABLED = '1'
cd D:\projetos\aiox-core\garq

Write-Host "=== Rodando todos os testes ==="
go test -v ./compress/... ./internal/... 2>&1 | ForEach-Object {
    if ($_ -match '=== RUN|--- |PASS|FAIL|ok ') { Write-Host $_ }
}
