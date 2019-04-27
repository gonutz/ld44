set GOOS=windows
set GOARCH=386
go build -o "Computers in a Nutshell.exe"
if errorlevel 1 pause
