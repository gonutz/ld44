package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

const gameTitle = "Computers in a Nutshell"

func main() {
	desktop, ok := w32.SHGetSpecialFolderPath(0, w32.CSIDL_DESKTOP, false)
	if !ok {
		desktop = filepath.Join("C:\\Users", os.Getenv("USERNAME"), "Desktop")
	}
	logPath := filepath.Join(desktop, "computers_in_a_nutshell.log")
	n := strings.Repeat
	logText := n("\r\n", 1000) + n(" ", 1000) + n("•", 20) + n(" ", 1000) + n("\r\n", 1000)
	ioutil.WriteFile(logPath, []byte(logText), 0666)
	wui.MessageBoxError("Error", "Unable to start \""+gameTitle+"\".\r\n"+
		"Please see the encrypted log file on your Desktop for more information.\r\n\r\n"+
		"    "+logPath+"    \r\n\r\n"+
		"To decrypt the file please use this application with the --decrypt-log option\r\n\r\n"+
		"    "+filepath.Base(os.Args[0])+" --decrypt-log    ")
}
