package ui

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"tf77/internal/debugger"
)

// WatchWindow displays inspected variables and debugger status at the bottom of the IDE
type WatchWindow struct {
	Visible      bool
	Variables    []debugger.Variable
	ScrollY      int
	StatusText   string
	WindowNumber int
}

func NewWatchWindow(winNum int) *WatchWindow {
	return &WatchWindow{
		Visible:      false,
		WindowNumber: winNum,
		StatusText:   "No active debug session.",
	}
}

func (w *WatchWindow) SetState(st debugger.DebugState) {
	w.Variables = st.LocalVars
	if st.Exited {
		w.StatusText = fmt.Sprintf("Process exited with code %d. [Debug finished]", st.ExitCode)
	} else if st.CurrentFile != "" {
		fname := filepath.Base(st.CurrentFile)
		fn := st.CurrentFunc
		if fn == "" {
			fn = "MAIN"
		}
		w.StatusText = fmt.Sprintf("Paused at %s:%d in %s", fname, st.CurrentLine, fn)
	} else if st.Active {
		w.StatusText = "Debugging session active (running)..."
	} else {
		w.StatusText = "No active debug session."
	}
}

func (w *WatchWindow) ScrollUp() {
	if w.ScrollY > 0 {
		w.ScrollY--
	}
}

func (w *WatchWindow) ScrollDown(height int) {
	if w.ScrollY+height < len(w.Variables) {
		w.ScrollY++
	}
}

// Draw renders the Watches window inside the specified rectangle
func (w *WatchWindow) Draw(screen tcell.Screen, x, y, width, height int, active bool) {
	if !w.Visible || height < 3 {
		return
	}

	frameStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorBorder)
	if !active {
		frameStyle = tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorDarkGray)
	}
	shadowStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)

	// Drop shadow
	for r := y + 1; r < y+height+1; r++ {
		screen.SetContent(x+width, r, ' ', nil, shadowStyle)
		screen.SetContent(x+width+1, r, ' ', nil, shadowStyle)
	}
	for c := x + 2; c < x+width+2; c++ {
		screen.SetContent(c, y+height, ' ', nil, shadowStyle)
	}

	// Corners
	screen.SetContent(x, y, RuneDoubleTopLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y, RuneDoubleTopRight, nil, frameStyle)
	screen.SetContent(x, y+height-1, RuneDoubleBottomLeft, nil, frameStyle)
	screen.SetContent(x+width-1, y+height-1, RuneDoubleBottomRight, nil, frameStyle)

	// Horizontal borders
	for c := x + 1; c < x+width-1; c++ {
		screen.SetContent(c, y, RuneDoubleHorizontal, nil, frameStyle)
		screen.SetContent(c, y+height-1, RuneDoubleHorizontal, nil, frameStyle)
	}

	// Vertical borders
	for r := y + 1; r < y+height-1; r++ {
		screen.SetContent(x, r, RuneDoubleVertical, nil, frameStyle)
		screen.SetContent(x+width-1, r, RuneDoubleVertical, nil, frameStyle)
	}

	// Title: ` 2 Watches `
	title := fmt.Sprintf(" %d Watches ", w.WindowNumber)
	tx := x + (width-len(title))/2
	titleStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorTitle).Bold(true)
	for i, r := range title {
		screen.SetContent(tx+i, y, r, nil, titleStyle)
	}

	// Interior clear
	contentStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(ColorEditorFg)
	for r := y + 1; r < y+height-1; r++ {
		for c := x + 1; c < x+width-1; c++ {
			screen.SetContent(c, r, ' ', nil, contentStyle)
		}
	}

	// Header row: Status bar inside watch window
	headerStyle := tcell.StyleDefault.Background(ColorEditorLineNumBg).Foreground(ColorEditorLineNumFg).Bold(true)
	statusRunes := []rune(w.StatusText)
	for c := x + 1; c < x+width-1; c++ {
		idx := c - (x + 1)
		r := ' '
		if idx < len(statusRunes) {
			r = statusRunes[idx]
		}
		screen.SetContent(c, y+1, r, nil, headerStyle)
	}

	// Draw variables
	varRowStart := y + 2
	maxRows := height - 3

	varNameStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorYellow).Bold(true)
	varTypeStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorLightCyan)
	varValStyle := tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorWhite)

	if len(w.Variables) == 0 {
		emptyMsg := "No watch expressions or local variables."
		for i, r := range emptyMsg {
			if x+2+i < x+width-1 {
				screen.SetContent(x+2+i, varRowStart, r, nil, tcell.StyleDefault.Background(ColorEditorBg).Foreground(tcell.ColorDarkGray))
			}
		}
		return
	}

	for i := 0; i < maxRows; i++ {
		varIdx := w.ScrollY + i
		if varIdx >= len(w.Variables) {
			break
		}
		v := w.Variables[varIdx]
		curY := varRowStart + i

		// Format: `  NAME: TYPE = VALUE`
		curX := x + 2

		// Name
		for _, r := range v.Name {
			if curX < x+width-1 {
				screen.SetContent(curX, curY, r, nil, varNameStyle)
				curX += runewidth.RuneWidth(r)
			}
		}

		if curX < x+width-1 {
			screen.SetContent(curX, curY, ':', nil, contentStyle)
			curX++
			screen.SetContent(curX, curY, ' ', nil, contentStyle)
			curX++
		}

		// Type
		if v.Type != "" {
			for _, r := range v.Type {
				if curX < x+width-1 {
					screen.SetContent(curX, curY, r, nil, varTypeStyle)
					curX += runewidth.RuneWidth(r)
				}
			}
			if curX < x+width-1 {
				screen.SetContent(curX, curY, ' ', nil, contentStyle)
				curX++
			}
		}

		if curX < x+width-1 {
			screen.SetContent(curX, curY, '=', nil, contentStyle)
			curX++
			screen.SetContent(curX, curY, ' ', nil, contentStyle)
			curX++
		}

		// Value
		for _, r := range v.Value {
			if curX < x+width-1 {
				screen.SetContent(curX, curY, r, nil, varValStyle)
				curX += runewidth.RuneWidth(r)
			}
		}
	}
}
