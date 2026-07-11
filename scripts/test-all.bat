@echo off
set GOROOT=L:\sdk\go1.24.0
set PATH=L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;%PATH%
set CC=L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe
set CGO_ENABLED=1
cd /d D:\projetos\aiox-core\garq

echo === Rodando todos os testes ===
go test -v .\compress\... .\internal\...
pause
