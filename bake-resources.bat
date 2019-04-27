go get github.com/gonutz/ico
if errorlevel 1 (pause & exit)
go get github.com/gonutz/bin2go/v2/bin2go
if errorlevel 1 (pause & exit)
go get github.com/gonutz/rsrc
if errorlevel 1 (pause & exit)

ico main-icon.png main-icon.ico
if errorlevel 1 (pause & exit)
rsrc -ico main-icon.ico -arch=386 -o=rsrc_386.syso
if errorlevel 1 (pause & exit)
bin2go -var=mainIcon < main-icon.ico > mainIcon.go
if errorlevel 1 (pause & exit)

ico decrypt-icon.png decrypt-icon.ico
if errorlevel 1 (pause & exit)
bin2go -var=decryptIcon < decrypt-icon.ico > decryptIcon.go
if errorlevel 1 (pause & exit)

ico fix-graphics-icon.png fix-graphics-icon.ico
if errorlevel 1 (pause & exit)
bin2go -var=fixGraphicsIcon < fix-graphics-icon.ico > fixGraphicsIcon.go
if errorlevel 1 (pause & exit)
