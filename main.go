package main

/*
TODOs

if the user enters the wrong password over and over again, give her a hint to
look in the encrypted log file itself.

what if the user misses the dialog that tells her to look through the logs in
her documents folder? restarting the program should remember this
*/

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

const (
	gameTitle              = "Computers in a Nutshell"
	decryptFlag            = "--decrypt-log"
	decryptedLogFileName   = "computers_in_a_nutshell.enclog"
	logPassword            = "••••••••••••••••••••"
	uninstallFlag          = "--uninstall"
	clearLogFileNameFormat = "computers_in_a_nutshell_%05d.log"
	fixGraphicsFlag        = "--fix-graphics"
)

func main() {
	// TODO remove this debug code
	//os.Args = append(os.Args, decryptFlag)
	//os.Args = append(os.Args, uninstallFlag)
	//showDecryptionProgress(nil)
	//return
	//os.Args = append(os.Args, fixGraphicsFlag)
	// TODO remove the above debug code

	if len(os.Args) >= 2 && os.Args[1] == uninstallFlag {
		uninstall()
	} else if len(os.Args) >= 2 && os.Args[1] == fixGraphicsFlag {
		fixGraphics()
	} else if len(os.Args) >= 2 && os.Args[1] == decryptFlag {
		decrypt()
	} else {
		createDesktopLog()
	}
}

func uninstall() {
	os.Remove(encryptedLogPath())
	removeClearTextLogs()
	wui.MessageBoxInfo("Uninstall", "All files created by the game have now been deleted")
}

func removeClearTextLogs() {
	documents := documentsPath()
	for i := 1; i <= 1000; i++ {
		logPath := filepath.Join(documents, fmt.Sprintf(clearLogFileNameFormat, i))
		os.Remove(logPath)
	}
}

func createDesktopLog() {
	logPath := encryptedLogPath()
	n := strings.Repeat
	logText := n("\r\n", 1000) + n(" ", 1000) + logPassword + n(" ", 1000) + n("\r\n", 1000)
	ioutil.WriteFile(logPath, []byte(logText), 0666)

	window := wui.NewWindow()
	window.SetTitle(gameTitle)
	window.SetIconFromMem(mainIcon)
	window.SetSize(640, 480)
	window.SetOnShow(func() {
		wui.MessageBoxError("Error", "Unable to start \""+gameTitle+"\".\r\n"+
			"Please see the log file on your Desktop for more information.\r\n\r\n"+
			"    "+logPath+"    \r\n\r\n"+
			"To protect your privacy the log file has been encrypted.\r\n\r\n"+
			"Decrypt the file using this application with flag "+decryptFlag+"\r\n\r\n"+
			"    \""+filepath.Base(os.Args[0])+"\" "+decryptFlag+"    ")
		window.Close()
	})
	window.Show()
}

func encryptedLogPath() string {
	return filepath.Join(desktopPath(), decryptedLogFileName)
}

func desktopPath() string {
	path, ok := w32.SHGetSpecialFolderPath(0, w32.CSIDL_DESKTOP, false)
	if !ok {
		path = filepath.Join("C:\\Users", os.Getenv("USERNAME"), "Desktop")
	}
	return path
}

func documentsPath() string {
	path, ok := w32.SHGetSpecialFolderPath(0, w32.CSIDL_MYDOCUMENTS, false)
	if !ok {
		path = filepath.Join("C:\\Users", os.Getenv("USERNAME"), "Documents")
	}
	return path
}

func decrypt() {
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

		correctPassword := pw.Text() == logPassword

		if correctPassword {
			// Start creating a lot of log files in the user's Documents folder.
			// These logs contain only one line of useful information.
			// All lines in the log files have the same length, this way the
			// player cannot infer which log file is the right one just by
			// looking at its file size.
			go func() {
				const (
					fileCount    = 1000
					linesPerFile = 100
				)
				nutshellImageIndex := 0
				computerImageIndex := 0
				// Place the actual useful information somewhere in the last
				// half of the log messages.
				realInfoCounter := fileCount*linesPerFile/2 + rand.Intn(fileCount*linesPerFile/4)

				nextLine := func() string {
					realInfoCounter--
					if realInfoCounter == 0 {
						return "InitGraphics... FAIL (try flag --fix-graphics)\r\n"
					}
					if rand.Intn(2) == 0 {
						nutshellImageIndex++
						return fmt.Sprintf(
							"loading 'nutshell_%017d.png'... OK\r\n",
							nutshellImageIndex,
						)
					}
					computerImageIndex++
					return fmt.Sprintf(
						"loading 'computer_%017d.png'... OK\r\n",
						computerImageIndex,
					)
				}

				documents := documentsPath()
				for i := 1; i <= fileCount; i++ {
					logPath := filepath.Join(documents, fmt.Sprintf(clearLogFileNameFormat, i))
					var content string
					for j := 0; j < linesPerFile; j++ {
						content += nextLine()
					}
					ioutil.WriteFile(logPath, []byte(content), 0666)
				}

				os.Remove(encryptedLogPath())
			}()
		}

		showDecryptionProgress(window)

		if !correctPassword {
			wui.MessageBoxError("Error", "Wrong password.")
			return
		}

		wui.MessageBoxInfo("Log File Decrypted", "TODO: include this text: For easier human consumption the clear text log was split into multiple files to make sure no file is too large for your text editor.")
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

// showDecryptionProgress opens a modal window, slowly fills it with a progress
// bar, then closes it.
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

func fixGraphics() {
	go removeClearTextLogs()
	wui.MessageBoxError("TODO", "Implement more game here")
}
