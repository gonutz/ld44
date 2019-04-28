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

bin2go -var=menuBackground < menu-background.png > menuBackground.go
if errorlevel 1 (pause & exit)
bin2go -var=pcImage < pc.png > pc.go
if errorlevel 1 (pause & exit)
bin2go -var=pcHot < pc-highlighted.png > pcHot.go
if errorlevel 1 (pause & exit)
bin2go -var=nutshellBack < nutshell.png > nutshellBack.go
if errorlevel 1 (pause & exit)
bin2go -var=nutshellFront < nutshell-front.png > nutshellFront.go
if errorlevel 1 (pause & exit)
