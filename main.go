package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io/ioutil"
	"math"
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
	updatedVersion         = "0.0.12.43785634"
	versionFlag            = "--version=" + updatedVersion
)

func main() {
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
	} else if os.Args[1] == versionFlag {
		playGame()
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
			"To protect your privacy the log file has been encrypted.\r\n"+
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

		wui.MessageBoxInfo("Log File Decrypted",
			"The log file was decrypted successfully.\r\n\r\n"+
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
	defer dlg.Destroy()
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
				wui.MessageBoxInfo("Success",
					"The following gamma settings were detected:\r\n\r\n"+
						"    "+gammaValues+"    \r\n\r\n"+
						"Please restart the game with these parameters:\r\n\r\n"+
						"    \""+filepath.Base(os.Args[0])+"\" "+gammaFlag+"    ")
				window.Close()
				return
			}
			lightColors = lightColors[1:]
		} else {
			wui.MessageBoxError(
				"Error",
				"Inconsistent gamma settings detected, please repeat the last step.",
			)
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

	background := makeImage(menuBackground)
	back := wui.NewPaintbox()
	back.SetBounds(window.ClientBounds())
	back.SetOnPaint(func(c *wui.Canvas) {
		c.FillRect(0, 0, c.Width(), c.Height(), wui.RGB(240, 240, 240))
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
		if choosePassword(window) {
			updateGame(window)
		}
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

func choosePassword(parent *wui.Window) (success bool) {
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
		if s == tooShort {
			wui.MessageBoxError("Illegal Password", "The password is too short.")
			return
		}
		if s == veryWeak || s == weak {
			wui.MessageBoxError("Illegal Password", "The password is too weak.")
			return
		}
		if s != medium {
			wui.MessageBoxError("Illegal Password", "Studies have shown that "+
				"complicated passwords are often written down on paper. This "+
				"presents a large security vulnerability. We thus encourage "+
				"you to rather use a medium strength password instead and keep "+
				"it in your head rather than on paper.")
			return
		}

		d := editDistance(pw.Text(), repeat.Text())
		if d == 0 {
			wui.MessageBoxError("Illegal Password", "The password you have "+
				"entered in the bottom edit box is already in use by another "+
				"edit box in the system.")
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
		success = true
		window.Close()
	})

	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_RETURN}, func() {
		if ok.Enabled() && (pw.HasFocus() || repeat.HasFocus()) {
			ok.OnClick()()
		}
	})
	window.SetOnShow(func() {
		if parent != nil {
			x, y, w, h := parent.Bounds()
			windowW, windowH := window.Size()
			window.SetPos(x+(w-windowW)/2, y+(h-windowH)/2)
		}
		pw.Focus()
	})

	window.ShowModal()

	return success
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
		return "very strong"
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
// http://www.golangprograms.com/golang-program-for-implementation-of-levenshtein-distance.html
func editDistance(s1, s2 string) int {
	a := []rune(s1)
	b := []rune(s2)
	m := len(a)
	n := len(b)

	column := make([]int, m+1)
	for y := range column {
		column[y] = y
	}

	for x := 1; x <= n; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= m; y++ {
			oldkey := column[y]
			var inc int
			if a[y-1] != b[x-1] {
				inc = 1
			}
			column[y] = min(column[y]+1, column[y-1]+1, lastkey+inc)
			lastkey = oldkey
		}
	}

	return column[m]
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

func updateGame(parent *wui.Window) {
	showProgress("Starting Game...", parent)
	for !wui.MessageBoxYesNo(
		"Important Update",
		"You are using an outdated version of \""+gameTitle+
			"\". A newer version is available for download.\r\n\r\n"+
			"Do you want to update your game now?",
	) {
	}
	showProgress("Downloading...", parent)
	showProgress("Installing...", parent)
	wui.MessageBoxInfo(
		"Restart",
		"Please restart the game with flag\r\n\r\n"+
			"    "+versionFlag+"    \r\n\r\n"+
			"Starting without this flag allows you to still use the previous "+
			"version. Keeping all versions at all times provides you with the "+
			"most control over your gaming experience.\r\n\r\n"+
			"Thank you for playing \""+gameTitle+"\".",
	)
}

func playGame() {
	window := wui.NewDialogWindow()
	window.SetTitle(gameTitle + " - v" + updatedVersion)
	window.SetClientSize(640, 480)
	window.SetIconFromMem(mainIcon)

	tahoma, err := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	if err == nil {
		window.SetFont(tahoma)
	}

	state := "menu"
	var enterState func(s string)
	pcX, pcY := 10, 10
	overallPCMovement := 0.0
	nextAutoSave := 0.0
	pcIsHot := false
	pcIsMoving := false
	var pcMoveX, pcMoveY int
	mouseX := 0

	const starY = 200

	background := makeImage(menuBackground)
	pc := makeImage(pcImage)
	pcHot := makeImage(pcHot)
	nutshell := makeImage(nutshellBack)
	nutshellFront := makeImage(nutshellFront)
	starEmpty := makeImage(emptyStar)
	starHalf := makeImage(halfStar)
	starFull := makeImage(fullStar)

	autoSave := func() {
		pcIsMoving = false
		pcIsHot = false
		showProgress("Auto-Save...", window)
		wui.MessageBoxError(
			"Error",
			"Unable to synchronize your save game with the server. Please try again later.",
		)
	}

	back := wui.NewPaintbox()
	back.SetBounds(window.ClientBounds())
	largeFont, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -40})
	back.SetOnPaint(func(c *wui.Canvas) {
		c.FillRect(0, 0, c.Width(), c.Height(), wui.RGB(240, 240, 240))
		if state == "menu" {
			c.DrawImage(background, background.Bounds(), 0, 0)
		} else if state == "instructions" {
			c.SetFont(largeFont)
			c.TextRectFormat(
				0, 0, c.Width(), c.Height()/3*2,
				"Put the Computer in the Nutshell",
				wui.FormatCenter, wui.RGB(0, 0, 0),
			)
			c.SetFont(tahoma)
			c.TextRectFormat(
				0, c.Height()/2, c.Width(), c.Height()/2,
				"Click to continue",
				wui.FormatCenter, wui.RGB(92, 92, 92),
			)
		} else if state == "playing" {
			pc := pc
			if pcIsHot && !pcIsMoving {
				pc = pcHot
			}
			nutshellX := c.Width() - nutshell.Width() - 10
			nutshellY := c.Height() - nutshell.Height() - 10
			c.DrawImage(nutshell, nutshell.Bounds(), nutshellX, nutshellY)
			c.DrawImage(pc, pc.Bounds(), pcX, pcY)
			c.DrawImage(nutshellFront, nutshellFront.Bounds(), nutshellX, nutshellY)

			pcCenterX := pcX + pc.Width()/2
			pcCenterY := pcY + pc.Height()/2
			nutCenterX := nutshellX + nutshell.Width()/2
			nutCenterY := nutshellY + nutshell.Height()/3
			dx, dy := pcCenterX-nutCenterX, pcCenterY-nutCenterY
			if dx*dx+dy*dy < 10*10 {
				enterState("won")
			}
		} else if state == "won" {
			c.SetFont(largeFont)
			c.TextRectFormat(
				0, 0, c.Width(), c.Height()/3*2,
				"Well done!",
				wui.FormatCenter, wui.RGB(0, 0, 0),
			)
			c.SetFont(tahoma)
			c.TextRectFormat(
				0, c.Height()/2, c.Width(), c.Height()/2,
				"Click to continue",
				wui.FormatCenter, wui.RGB(92, 92, 92),
			)
			nutshellX := c.Width() - nutshell.Width() - 10
			nutshellY := c.Height() - nutshell.Height() - 10
			c.DrawImage(nutshell, nutshell.Bounds(), nutshellX, nutshellY)
			c.DrawImage(pc, pc.Bounds(), pcX, pcY)
			c.DrawImage(nutshellFront, nutshellFront.Bounds(), nutshellX, nutshellY)
		} else if state == "rating" {
			c.SetFont(largeFont)
			c.TextRectFormat(
				0, 0, c.Width(), c.Height()/2,
				"Please rate the game",
				wui.FormatCenter, wui.RGB(0, 0, 0),
			)
			for i := 0; i < 5; i++ {
				x := 80 + i*100
				star := starFull
				if mouseX < x {
					star = starEmpty
				} else if mouseX < x+star.Width()/2 {
					star = starHalf
				}
				c.DrawImage(star, star.Bounds(), x, starY)
			}
			c.SetFont(tahoma)
			c.TextRectFormat(
				0, c.Height()/2, c.Width(), c.Height()/2,
				"Click to rate",
				wui.FormatCenter, wui.RGB(92, 92, 92),
			)
		} else if state == "rated" {
			c.SetFont(largeFont)
			c.TextRectFormat(
				0, 0, c.Width(), c.Height()/2,
				"Thanks for the thumbs up!",
				wui.FormatCenter, wui.RGB(0, 0, 0),
			)
			for i := 0; i < 5; i++ {
				c.DrawImage(starFull, starFull.Bounds(), 80+i*100, starY)
			}
			c.SetFont(tahoma)
			c.TextRectFormat(
				0, c.Height()/2, c.Width(), c.Height()/2,
				"Click to upload",
				wui.FormatCenter, wui.RGB(92, 92, 92),
			)
		}
	})
	window.Add(back)

	newGame := wui.NewButton()
	newGame.SetText("Play")
	newGame.SetSize(100, 25)
	newGame.SetPos(
		(window.ClientWidth()-newGame.Width())/2,
		window.ClientHeight()/2-newGame.Height()-10,
	)
	window.Add(newGame)

	exit := wui.NewButton()
	exit.SetText("Exit")
	exit.SetBounds(newGame.Bounds())
	exit.SetY(newGame.Y() + newGame.Height() + 10)
	exit.SetOnClick(window.Close)
	window.Add(exit)

	enterState = func(s string) {
		state = s
		if s == "menu" {
			newGame.SetVisible(true)
			exit.SetVisible(true)
		}
		if s == "instructions" {
			newGame.SetVisible(false)
			exit.SetVisible(false)
		}
		back.Paint()
	}

	newGame.SetOnClick(func() {
		enterState("instructions")
	})

	window.SetOnMouseMove(func(x, y int) {
		if state == "playing" {
			if pcIsMoving {
				dx, dy := x-pcMoveX, y-pcMoveY
				pcX += dx
				pcY += dy
				overallPCMovement += math.Sqrt(float64(dx*dx + dy*dy))
				pcMoveX, pcMoveY = x, y
				if overallPCMovement >= nextAutoSave {
					back.Paint()
					autoSave()
					nextAutoSave = overallPCMovement + 20
				}
			}
			pcIsHot = x >= pcX && x < pcX+pc.Width() && y >= pcY && y < pcY+pc.Height()
			back.Paint()
		}
		if state == "rating" {
			mouseX = x
			back.Paint()
		}
	})
	window.SetOnMouseDown(func(button wui.MouseButton, x, y int) {
		if state == "instructions" {
			enterState("playing")
		} else if state == "won" {
			enterState("rating")
		} else if state == "playing" {
			if button == wui.MouseButtonLeft && pcIsHot {
				pcMoveX, pcMoveY = x, y
				pcIsMoving = true
				back.Paint()
			}
		} else if state == "rating" {
			enterState("rated")
		} else if state == "rated" {
			showProgress("Uploading...", window)
			wui.MessageBoxError("Error", "Failed to upload your rating.\r\n\r\n"+
				"I hope you enjoyed the game. Judging from your rating "+
				"you are eager to give this game a great review with a detailed "+
				"comment on its Ludum Dare site\r\n\r\n"+
				"    https://ldjam.com/events/ludum-dare/44/$141527    \r\n\r\n"+
				"Remember that though the game might have been rather short, "+
				"much emphasis was put on compatibility, security and "+
				"reliability so you got to enjoy a high quality game.\r\n\r\n"+
				"Thanks for playing!\r\n\r\n"+
				"- gonutz\r\n\r\n\r\n\r\n\r\n"+
				"P.S. did you know that you can hit Ctrl+C in a "+
				"Windows message box and it will save the text to clipboard? "+
				"Go ahead, try it.")
			window.Close()
		}
	})
	window.SetOnMouseUp(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			pcIsMoving = false
			back.Paint()
		}
	})
	window.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, window.Close)
	window.SetOnShow(func() {
		newGame.Focus()
	})

	window.Show()
}

func makeImage(pngData []byte) *wui.Image {
	img, _ := png.Decode(bytes.NewReader(pngData))
	return wui.NewImage(img)
}
