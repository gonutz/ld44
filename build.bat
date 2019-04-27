call bake-resources.bat
if errorlevel 1 (pause & exit)

set GOOS=windows
set GOARCH=386
go build -ldflags="-H=windowsgui" -o "Computers in a Nutshell.exe"
if errorlevel 1 pause
