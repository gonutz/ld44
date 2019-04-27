go get github.com/gonutz/ico
if errorlevel 1 (pause & exit)
go get github.com/gonutz/bin2go/v2/bin2go
if errorlevel 1 (pause & exit)

ico decrypt-icon.png decrypt-icon.ico
if errorlevel 1 (pause & exit)
bin2go -var=decryptIcon < decrypt-icon.ico > decryptIcon.go
if errorlevel 1 (pause & exit)
