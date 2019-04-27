package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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
	gammaValues            = "offset 1.72, exponent -0.247, ramp-up 7"
	gammaFlag              = "--gamma=1.72,-0.246,7"
)

func main() {
	// TODO remove this debug code
	//os.Args = append(os.Args, decryptFlag)
	//os.Args = append(os.Args, uninstallFlag)
	//showDecryptionProgress(nil)
	//return
	//os.Args = append(os.Args, fixGraphicsFlag)
	//os.Args = append(os.Args, gammaFlag)
	// TODO remove the above debug code

	rand.Seed(time.Now().UnixNano())

	if len(os.Args) == 1 {
		createDesktopLog()
	} else if len(os.Args) > 2 {
		wui.MessageBoxError("Error", "Too many arguments.")
	} else if os.Args[1] == decryptFlag {
		decrypt()
	} else if os.Args[1] == fixGraphicsFlag {
		fixGraphics()
	} else if os.Args[1] == gammaFlag {
		selectPassword()
	} else if os.Args[1] == uninstallFlag {
		uninstall()
	} else {
		wui.MessageBoxError("Invalid Argument", "Unknown flag '"+os.Args[1]+"'.")
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

	window := wui.NewDialogWindow()
	window.SetTitle(gameTitle)
	window.SetIconFromMem(mainIcon)
	window.SetClientSize(640, 480)
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
	window.SetClientSize(700, 250)
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

	autoExtract := wui.NewButton()
	autoExtract.SetText("Auto-Extract Password")
	autoExtract.SetBounds(window.ClientWidth()/2-80, 150, 160, 25)
	autoExtract.SetVisible(false)
	window.Add(autoExtract)
	autoExtract.SetOnClick(func() {
		showProgress("Extracting...", window)
		wui.MessageBoxError("Error", "Unable to automatically extract password from\r\n\r\n"+
			"    \""+logPath.Text()+"\"    \r\n\r\n"+
			"Please open the file in a text editor and extract the password manually.")
	})

	logPath.SetOnTextChange(func() {
		autoExtract.SetVisible(strings.TrimSpace(logPath.Text()) != "")
	})

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
							"loading image 'nutshell_%011d.png'... OK\r\n",
							nutshellImageIndex,
						)
					}
					computerImageIndex++
					return fmt.Sprintf(
						"loading image 'computer_%011d.png'... OK\r\n",
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

		showProgress("Decrypting...", window)

		if !correctPassword {
			wui.MessageBoxError("Error", "Wrong password.")
			return
		}

		wui.MessageBoxInfo("Log File Decrypted", "The log file was decrypted successfully.\r\n\r\n"+
			"For easier human consumption the clear text log was split into "+
			"multiple files to make sure no file is too large for viewing in a"+
			" text editor.\r\n\r\n"+
			"You can find all log files in your Documents folder:\r\n\r\n"+
			"    "+documentsPath()+"    ")
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
func showProgress(title string, parent *wui.Window) {
	const maxProgress = 250
	dlg := wui.NewDialogWindow()
	dlg.SetClientSize(maxProgress, 25)
	dlg.SetTitle(title)

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
	go removeClearTextLogs() // log stage clear - delete the many log files

	const (
		tileSize   = 40
		tileCountX = 15
		tileCountY = 10
		yOffset    = 60
	)
	lightColors := []wui.Color{
		wui.RGB(50, 50, 50),
		wui.RGB(30, 30, 30),
		wui.RGB(10, 10, 10),
		wui.RGB(0, 1, 0),
	}

	window := wui.NewDialogWindow()
	window.SetTitle(gameTitle + " - Diagnostics")
	window.SetClientSize(tileCountX*tileSize+1, yOffset+tileCountY*tileSize+1)
	window.SetIconFromMem(fixGraphicsIcon)

	tahoma, err := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	if err == nil {
		window.SetFont(tahoma)
	}

	line := func(text string, y int) {
		l := wui.NewLabel()
		l.SetBounds(0, y, window.ClientWidth(), 20)
		l.SetCenterAlign()
		l.SetText(text)
		window.Add(l)
	}
	line("Gamma Correction Calibration", 10)
	line("Please select the square that appears brightest to you", 30)

	hotX, hotY := -1, -1 // tile under mouse
	randTile := func() (x, y int) {
		return 1 + rand.Intn(tileCountX-2), 1 + rand.Intn(tileCountY-2)
	}
	lightX, lightY := randTile()
	p := wui.NewPaintbox()
	p.SetBounds(0, yOffset, tileCountX*tileSize+1, tileCountY*tileSize+1)
	p.SetOnPaint(func(c *wui.Canvas) {
		c.FillRect(0, 0, c.Width(), c.Height(), wui.RGB(0, 0, 0))
		border := wui.RGB(192, 192, 192)
		for x := 0; x < tileCountX+1; x++ {
			borderX := x * tileSize
			c.Line(borderX, 0, borderX, c.Height(), border)
		}
		for y := 0; y < tileCountY+1; y++ {
			borderY := y * tileSize
			c.Line(0, borderY, c.Width(), borderY, border)
		}
		c.FillRect(lightX*tileSize+1, lightY*tileSize+1, tileSize-1, tileSize-1, lightColors[0])
		if hotX >= 0 && hotX < tileCountX && hotY >= 0 && hotY < tileCountY {
			c.DrawRect(hotX*tileSize, hotY*tileSize, tileSize+1, tileSize+1, wui.RGB(0, 192, 0))
		}
	})
	window.Add(p)
	window.SetOnMouseMove(func(x, y int) {
		y -= yOffset
		if y < 0 {
			hotX, hotY = -1, -1
		} else {
			hotX, hotY = x/tileSize, y/tileSize
		}
		p.Paint()
	})
	window.SetOnMouseDown(func(b wui.MouseButton, x, y int) {
		if hotX < 0 || hotX >= tileCountX || hotY < 0 || hotY >= tileCountY {
			return
		}
		correct := hotX == lightX && hotY == lightY
		hotX, hotY = -1, -1
		p.Paint()
		showProgress("Calibrating...", window)
		lightX, lightY = randTile()
		p.Paint()
		if correct {
			if len(lightColors) == 1 {
				wui.MessageBoxInfo("Success", "The following gamma settings were detected:\r\n\r\n"+
					"    "+gammaValues+"    \r\n\r\n"+
					"Please restart the game with these parameters:\r\n\r\n"+
					"    \""+filepath.Base(os.Args[0])+"\" "+gammaFlag+"    ")
				window.Close()
				return
			}
			lightColors = lightColors[1:]
		} else {
			wui.MessageBoxError("Error", "Inconsistent gamma settings detected, please repeat the last step.")
		}
	})

	window.Show()
}

func selectPassword() {
	window := wui.NewDialogWindow()
	window.SetTitle(gameTitle)
	window.SetClientSize(640, 480)
	window.SetIconFromMem(mainIcon)

	tahoma, err := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	if err == nil {
		window.SetFont(tahoma)
	}

	img, _ := png.Decode(bytes.NewReader(menuBackground))
	composed := image.NewRGBA(img.Bounds())
	draw.Draw(
		composed,
		img.Bounds(),
		image.NewUniform(color.RGBA{240, 240, 240, 255}),
		image.ZP,
		draw.Src,
	)
	draw.Draw(composed, img.Bounds(), img, img.Bounds().Min, draw.Over)
	background := wui.NewImage(composed)

	back := wui.NewPaintbox()
	back.SetBounds(window.ClientBounds())
	back.SetOnPaint(func(c *wui.Canvas) {
		c.DrawImage(background, background.Bounds(), 0, 0)
	})
	window.Add(back)

	newGame := wui.NewButton()
	newGame.SetText("New Game")
	newGame.SetSize(100, 25)
	newGame.SetPos(
		(window.ClientWidth()-newGame.Width())/2,
		window.ClientHeight()/2-newGame.Height()-10,
	)
	newGame.SetOnClick(func() {
		choosePassword(window)
	})
	window.Add(newGame)

	exit := wui.NewButton()
	exit.SetText("Exit")
	exit.SetBounds(newGame.Bounds())
	exit.SetY(newGame.Y() + newGame.Height() + 10)
	exit.SetOnClick(window.Close)
	window.Add(exit)

	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.SetOnShow(func() {
		newGame.Focus()
	})

	window.Show()
}

func choosePassword(parent *wui.Window) {
	window := wui.NewDialogWindow()
	window.SetTitle("New Game")
	window.SetClientSize(520, 250)
	window.SetIconFromMem(mainIcon)

	tahoma, err := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	if err == nil {
		window.SetFont(tahoma)
	}

	line := func(text string, y int) {
		l := wui.NewLabel()
		l.SetBounds(0, y, window.ClientWidth(), 20)
		l.SetCenterAlign()
		l.SetText(text)
		window.Add(l)
	}
	line("Please select a password for your saved game.", 10)
	line("Progress will automatically be saved to our Trusted Servers ®.", 30)
	line("To ensure data security, your password must meet our password standards.", 50)

	pwCaption := wui.NewLabel()
	pwCaption.SetText("Password:")
	pwCaption.SetRightAlign()
	pwCaption.SetBounds(0, 110, 140, 25)
	window.Add(pwCaption)

	pw := wui.NewEditLine()
	pw.SetPassword(true)
	pw.SetBounds(pwCaption.X()+pwCaption.Width()+20, pwCaption.Y(), 200, pwCaption.Height())
	window.Add(pw)

	strength := wui.NewLabel()
	strength.SetBounds(pw.X()+pw.Width()+20, pw.Y(), 130, pw.Height())
	window.Add(strength)

	repeatCaption := wui.NewLabel()
	repeatCaption.SetText("Repeat:")
	repeatCaption.SetRightAlign()
	repeatCaption.SetBounds(pwCaption.Bounds())
	repeatCaption.SetY(repeatCaption.Y() + 40)
	window.Add(repeatCaption)

	repeat := wui.NewEditLine()
	repeat.SetPassword(true)
	repeat.SetBounds(pw.Bounds())
	repeat.SetY(repeatCaption.Y())
	window.Add(repeat)

	match := wui.NewLabel()
	match.SetBounds(strength.Bounds())
	match.SetY(repeat.Y())
	window.Add(match)

	ok := wui.NewButton()
	ok.SetText("OK")
	ok.SetSize(80, 25)
	ok.SetPos((window.ClientWidth()-ok.Width())/2, window.ClientHeight()-40)
	ok.SetEnabled(false)
	window.Add(ok)

	updateText := func() {
		passwordStrength := computePasswordStrength(pw.Text())
		bothMatch := pw.Text() == repeat.Text()

		strength.SetText(fmt.Sprintf("(%s)", passwordStrength))
		if bothMatch {
			match.SetText("")
		} else {
			match.SetText("(mismatch)")
		}

		if passwordStrength == veryStrong && bothMatch {
			ok.SetEnabled(true)
		}
	}
	pw.SetOnTextChange(updateText)
	repeat.SetOnTextChange(updateText)

	ok.SetOnClick(func() {
		s := computePasswordStrength(pw.Text())
		if s != medium {
			wui.MessageBoxError("Illegal Password", "Studies have shown that "+
				"forcing people to use very complicated passwords results in "+
				"these people writing their passwords down on paper. This "+
				"allows spies to read this security-critical information by "+
				"breaking in people's homes or stealing their wallets.\r\n\r\n"+
				"Thus we encourage you to rather use a medium strength password"+
				" instead and keep it in your head rather than on paper.")
			return
		}

		d := editDistance(pw.Text(), repeat.Text())
		if d == 0 {
			wui.MessageBoxError("Illegal Password", "The password you have entered "+
				"in the second edit box is already in use by another edit box "+
				"in the system.\r\n\r\n"+
				"Please change the password in the second edit box.")
			return
		}
		if d >= 2 {
			wui.MessageBoxError(
				"Illegal Password",
				"The two passwords do not match enough.",
			)
			return
		}

		// at this point we accept the password
		wui.MessageBoxError("TODO", "Implement more game here")
		window.Close()
	})

	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.SetOnShow(func() {
		if parent != nil {
			x, y, w, h := parent.Bounds()
			windowW, windowH := window.Size()
			window.SetPos(x+(w-windowW)/2, y+(h-windowH)/2)
		}
		pw.Focus()
	})

	window.ShowModal()
}

type passwordStrength int

const (
	tooShort passwordStrength = -1 + iota
	veryWeak
	weak
	medium
	strong
	veryStrong
)

func (s passwordStrength) String() string {
	switch s {
	case tooShort:
		return "too short"
	case veryWeak:
		return "very weak"
	case weak:
		return "weak"
	case medium:
		return "medium"
	case strong:
		return "strong"
	case veryStrong:
		return "veryStrong"
	default:
		return ""
	}
}

func computePasswordStrength(pw string) passwordStrength {
	if utf8.RuneCountInString(pw) < 8 {
		return tooShort
	}
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range pw {
		hasLower = hasLower || unicode.IsLower(r)
		hasUpper = hasUpper || unicode.IsUpper(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
		hasSpecial = hasSpecial || !(unicode.IsLetter(r) || unicode.IsDigit(r))
	}
	score := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	s := score(hasLower) + score(hasUpper) + score(hasDigit) + score(hasSpecial)
	return passwordStrength(s)
}

// editDistance returns the Levenshtein distance between s1 and s2.
// https://en.wikipedia.org/wiki/Wagner%E2%80%93Fischer_algorithm
// TODO find out why the tests fail
func editDistance(s1, s2 string) int {
	a := []rune(s1)
	b := []rune(s2)
	m := len(a)
	n := len(b)
	d := make([]int, (m+1)*(n+1))

	get := func(i, j int) int {
		return d[i+j*m]
	}
	set := func(i, j, to int) {
		d[i+j*m] = to
	}

	for i := 0; i <= m; i++ {
		set(i, 0, i)
	}
	for j := 0; j < n; j++ {
		set(0, j, j)
	}
	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			if a[i-1] == b[j-1] {
				set(i, j, get(i-1, j-1))
			} else {
				set(i, j, min(
					get(i-1, j)+1,
					get(i, j-1)+1,
					get(i-1, j-1)+1,
				))
			}
		}
	}

	return get(m, n)
}

func min(a int, b ...int) int {
	if len(b) == 0 {
		return a
	}
	if a < b[0] {
		return min(a, b[1:]...)
	} else {
		return min(b[0], b[1:]...)
	}
}
