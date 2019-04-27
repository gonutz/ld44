package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

const (
	gameTitle            = "Computers in a Nutshell"
	decryptFlag          = "--decrypt-log"
	decryptedLogFileName = "computers_in_a_nutshell.enclog"
	logPassword          = "••••••••••••••••••••"
	uninstallFlag        = "--uninstall"
)

func main() {
	// TODO remove this debug code
	//os.Args = append(os.Args, decryptFlag)
	//os.Args = append(os.Args, uninstallFlag)
	//showDecryptionProgress(nil)
	//return
	// TODO remove the above debug code

	if len(os.Args) >= 2 && os.Args[1] == uninstallFlag {
		uninstall()
	} else if len(os.Args) >= 2 && os.Args[1] == decryptFlag {
		decryptor()
	} else {
		createDesktopLog()
	}
}

func uninstall() {
	os.Remove(encryptedLogPath())
	wui.MessageBoxInfo("Uninstall", "All files created by the game have now been deleted")
}

func createDesktopLog() {
	logPath := encryptedLogPath()
	n := strings.Repeat
	logText := n("\r\n", 1000) + n(" ", 1000) + logPassword + n(" ", 1000) + n("\r\n", 1000)
	ioutil.WriteFile(logPath, []byte(logText), 0666)
	wui.MessageBoxError("Error", "Unable to start \""+gameTitle+"\".\r\n"+
		"Please see the encrypted log file on your Desktop for more information.\r\n\r\n"+
		"    "+logPath+"    \r\n\r\n"+
		"To decrypt the file please use this application with flag "+decryptFlag+"\r\n\r\n"+
		"    \""+filepath.Base(os.Args[0])+"\" "+decryptFlag+"    ")
}

func encryptedLogPath() string {
	return filepath.Join(desktopPath(), decryptedLogFileName)
}

func desktopPath() string {
	desktop, ok := w32.SHGetSpecialFolderPath(0, w32.CSIDL_DESKTOP, false)
	if !ok {
		desktop = filepath.Join("C:\\Users", os.Getenv("USERNAME"), "Desktop")
	}
	return desktop
}

func decryptor() {
	window := wui.NewDialogWindow()
	window.SetClientSize(700, 220)
	window.SetTitle(`"` + gameTitle + `"` + " Log File Decryptor")
	window.SetIconFromMem(decryptIcon)

	tahoma, err := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	if err == nil {
		window.SetFont(tahoma)
	}

	logCaption := wui.NewLabel()
	logCaption.SetCenterAlign()
	logCaption.SetBounds(0, 20, window.ClientWidth(), 20)
	logCaption.SetText("Select the Log File to Decrypt")
	window.Add(logCaption)

	selectLog := wui.NewButton()
	selectLog.SetText("Select...")
	selectLog.SetBounds(10, 45, 80, 25)
	window.Add(selectLog)

	logPath := wui.NewEditLine()
	logPath.SetX(selectLog.X() + selectLog.Width() + 5)
	logPath.SetY(selectLog.Y())
	logPath.SetWidth(window.ClientWidth() - 10 - logPath.X())
	logPath.SetHeight(selectLog.Height())
	window.Add(logPath)

	selectLog.SetOnClick(func() {
		open := wui.NewFileOpenDialog()
		ext := filepath.Ext(decryptedLogFileName)
		open.AddFilter(gameTitle+" Log File", ext)
		open.SetTitle("Select Encrypted Log File")
		open.SetInitialPath(desktopPath())
		if ok, path := open.ExecuteSingleSelection(window); ok {
			logPath.SetText(path)
		}
	})

	pwCaption := wui.NewLabel()
	pwCaption.SetText("Please Enter Your Log File Password Here:")
	pwCaption.SetCenterAlign()
	pwCaption.SetBounds(0, 90, window.ClientWidth(), 20)
	window.Add(pwCaption)

	pw := wui.NewEditLine()
	pw.SetPassword(true)
	pw.SetBounds(10, 115, window.ClientWidth()-20, 25)
	window.Add(pw)

	ok := wui.NewButton()
	ok.SetText("OK")
	ok.SetSize(80, 25)
	ok.SetPos(
		(window.ClientWidth()-ok.Width())/2,
		window.ClientHeight()-10-ok.Height(),
	)
	ok.SetOnClick(func() {
		haveLogPath := strings.ToLower(path.Clean(filepath.ToSlash(logPath.Text())))
		wantLogPath := strings.ToLower(filepath.ToSlash(encryptedLogPath()))
		if haveLogPath != wantLogPath {
			wui.MessageBoxError("Error", "Invalid log file. "+
				"Please make sure the encrypted log path is correct and that the file is not corrupted.")
			return
		}

		showDecryptionProgress(window)

		if pw.Text() != logPassword {
			wui.MessageBoxError("Error", "Wrong password.")
			return
		}

		// success
		wui.MessageBoxError("TODO", "Implement more game here")
		window.Close()
	})
	window.Add(ok)

	window.SetOnShow(func() {
		pw.Focus()
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_RETURN}, func() {
		if pw.HasFocus() {
			ok.OnClick()()
		}
	})

	window.Show()
}

func showDecryptionProgress(parent *wui.Window) {
	const maxProgress = 250
	dlg := wui.NewDialogWindow()
	dlg.SetClientSize(maxProgress, 25)
	dlg.SetTitle("Decrypting...")

	progress := 0

	p := wui.NewPaintbox()
	p.SetBounds(dlg.ClientBounds())
	p.SetOnPaint(func(c *wui.Canvas) {
		c.FillRect(0, 0, c.Width(), c.Height(), wui.RGB(0, 192, 0))
		c.FillRect(progress, 0, c.Width(), c.Height(), wui.RGB(240, 240, 240))
	})
	dlg.Add(p)

	start := make(chan bool, 1)
	go func() {
		<-start
		time.Sleep(250 * time.Millisecond)
		for i := 0; i <= maxProgress; i++ {
			progress = i
			p.Paint()
			time.Sleep(8 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
		dlg.Close()
	}()

	dlg.SetOnCanClose(func() bool { return progress >= maxProgress })
	dlg.SetOnShow(func() {
		if parent != nil {
			x, y, w, h := parent.Bounds()
			dlgW, dlgH := dlg.Size()
			dlg.SetPos(x+(w-dlgW)/2, y+(h-dlgH)/2)
		}
		start <- true
	})
	dlg.ShowModal()
}
