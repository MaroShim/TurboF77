package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

// StatusItem represents a shortcut key action on the bottom bar
type StatusItem struct {
	KeyName string
	Desc    string
	Action  string
}

// StatusBar renders the classic Borland bottom hotkey strip
type StatusBar struct {
	Items   []StatusItem
	Message string
	MsgTime time.Time
	Col     int
}

func NewStatusBar() *StatusBar {
	return &StatusBar{
		Items: []StatusItem{
			{KeyName: "F1", Desc: "Help", Action: "help_about"},
			{KeyName: "F2", Desc: "Save", Action: "file_save"},
			{KeyName: "F3", Desc: "Open", Action: "file_open"},
			{KeyName: "Alt+F9", Desc: "Compile", Action: "compile_compile"},
			{KeyName: "F9", Desc: "Make", Action: "compile_make"},
			{KeyName: "Ctrl+F9", Desc: "Run", Action: "run_run"},
			{KeyName: "Alt+F5", Desc: "User", Action: "run_userscreen"},
			{KeyName: "F10", Desc: "Menu", Action: "menu_toggle"},
		},
	}
}

func (sb *StatusBar) SetMessage(msg string) {
	sb.Message = msg
	sb.MsgTime = time.Now()
}

func (sb *StatusBar) SetCursorCol(col int) {
	sb.Col = col
}

// GetF77Field returns description of FORTRAN 77 fixed format field for column (1-indexed)
func GetF77Field(col int) string {
	if col <= 0 {
		return "Label"
	}
	if col <= 5 {
		return "Label"
	}
	if col == 6 {
		return "Cont"
	}
	if col <= 72 {
		return "Stmt"
	}
	return "Ident"
}

// Draw renders the status bar on the bottom row
func (sb *StatusBar) Draw(screen tcell.Screen, y, width int) {
	bgStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(ColorStatusBarFg)
	keyStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(ColorStatusBarHotKey).Bold(true)

	// Clear row
	for x := 0; x < width; x++ {
		screen.SetContent(x, y, ' ', nil, bgStyle)
	}

	// 1. Right side: FORTRAN 77 Column guide (e.g. `C:7 [Stmt]`)
	f77Info := fmt.Sprintf(" C:%-2d [%s] ", sb.Col, GetF77Field(sb.Col))
	infoStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorYellow).Bold(true)
	infoLen := len([]rune(f77Info))
	infoStartX := width - infoLen - 1
	if infoStartX > 30 {
		for i, r := range f77Info {
			screen.SetContent(infoStartX+i, y, r, nil, infoStyle)
		}
	} else {
		infoStartX = width
	}

	// 2. If there is an active status message within 3 seconds, display it prominently
	if sb.Message != "" && time.Since(sb.MsgTime) < 3*time.Second {
		tagStyle := tcell.StyleDefault.Background(tcell.ColorNavy).Foreground(tcell.ColorWhite).Bold(true)
		msgStyle := tcell.StyleDefault.Background(ColorStatusBarBg).Foreground(tcell.ColorYellow).Bold(true)
		tag := " [Turbo F77] "
		xPos := 1
		for _, r := range tag {
			screen.SetContent(xPos, y, r, nil, tagStyle)
			xPos++
		}
		xPos++
		for _, r := range sb.Message {
			if xPos < infoStartX-1 {
				screen.SetContent(xPos, y, r, nil, msgStyle)
				xPos++
			}
		}
		return
	}

	// 3. Hotkey Items
	xPos := 1
	for _, item := range sb.Items {
		if xPos >= infoStartX-len([]rune(item.KeyName))-len([]rune(item.Desc))-3 {
			break
		}

		// Draw KeyName (e.g. F1, Alt+F9)
		for _, r := range item.KeyName {
			if xPos < width {
				screen.SetContent(xPos, y, r, nil, keyStyle)
				xPos++
			}
		}

		// Draw Desc (e.g. Help, Save)
		screen.SetContent(xPos, y, ' ', nil, bgStyle)
		xPos++

		for _, r := range item.Desc {
			if xPos < width {
				screen.SetContent(xPos, y, r, nil, bgStyle)
				xPos++
			}
		}

		// Spacer
		screen.SetContent(xPos, y, ' ', nil, bgStyle)
		xPos++
	}
}
