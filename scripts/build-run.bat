@echo off
set GOROOT=L:\sdk\go1.24.0
set PATH=L:\sdk\mingw64\bin;L:\sdk\go1.24.0\bin;%PATH%
set CC=L:\sdk\mingw64\bin\x86_64-w64-mingw32-gcc.exe
set CGO_ENABLED=1
cd /d D:\projetos\aiox-core\garq

echo === Gerando icones ===
go run .\tools\genicons\

echo === Compilando recursos (.rc -^> .syso) ===
"L:\sdk\mingw64\bin\windres.exe" -i cmd/walk/resources.rc -o cmd/walk/resources.syso
if %ERRORLEVEL% NEQ 0 (echo windres falhou! & pause & exit /b 1)

echo === Compilando walk ===
go build -o L:\sdk\garq-new.exe .\cmd\walk\
if %ERRORLEVEL% NEQ 0 (echo Build falhou! & pause & exit /b 1)

echo === Iniciando Garq ===
L:\sdk\garq-new.exe
