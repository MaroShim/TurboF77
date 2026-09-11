package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newSimScreen(t *testing.T, w, h int) tcell.Screen {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	s.SetSize(w, h)
	return s
}

func TestEditorDraw_FocusedCursorAndGutter(t *testing.T) {
	screen := newSimScreen(t, 80, 25)

	ed := NewEditor("main.for", 1)
	ed.Lines = []string{
		"      PROGRAM MAIN",
		"      WRITE(*,*) 'HELLO'",
		"      STOP",
		"      END",
	}
	ed.ShowLineNums = true
	ed.CursorY = 1 // line 2: "      WRITE(*,*) 'HELLO'"
	ed.CursorX = 6 // on 'W'

	// Draw focused editor
	ed.Draw(screen, 0, 1, 80, 20, true)

	// 1. Check Gutter at row 2 (screenY = 1 + 1 = 2):
	// Must contain the cursor indicator '▸'
	r, _, _, _ := screen.GetContent(0, 2)
	if r != '▸' {
		t.Errorf("expected gutter marker '▸' on current line, got %q (rune %d)", r, r)
	}

	// Other lines (e.g. line 0, screenY = 1) must NOT contain '▸'
	r0, _, _, _ := screen.GetContent(0, 1)
	if r0 == '▸' {
		t.Errorf("expected non-cursor line gutter to NOT have '▸', got %q", r0)
	}

	// 2. Check Hardware Cursor:
	// Screen coordinates for cursor:
	// lineNumWidth: digits + 2 = 3 + 2 = 5
	// Cursor is at col 6, so screenX = 5 + 6 = 11, screenY = 2
	cellRune, _, style, _ := screen.GetContent(11, 2)
	if cellRune != 'W' {
		t.Errorf("expected cursor cell to contain 'W', got %q", cellRune)
	}
	_, bg, _ := style.Decompose()
	if bg != ColorEditorBg {
		t.Errorf("expected cursor cell to have normal background %v, got %v", ColorEditorBg, bg)
	}
}

func TestEditorDraw_UnfocusedState(t *testing.T) {
	screen := newSimScreen(t, 80, 25)

	ed := NewEditor("main.for", 1)
	ed.Lines = []string{
		"      PROGRAM MAIN",
		"      WRITE(*,*) 'HELLO'",
	}
	ed.ShowLineNums = true
	ed.CursorY = 1
	ed.CursorX = 6

	// Draw unfocused editor (e.g., when dialog or menu is active)
	ed.Draw(screen, 0, 1, 80, 20, false)

	// Gutter must NOT have '▸' when unfocused
	r, _, _, _ := screen.GetContent(0, 2)
	if r == '▸' {
		t.Errorf("expected no '▸' gutter marker when unfocused, got %q", r)
	}

	// Cursor cell has normal editor background
	cellRune, _, style, _ := screen.GetContent(11, 2)
	if cellRune != 'W' {
		t.Errorf("expected 'W', got %q", cellRune)
	}
	_, bg, _ := style.Decompose()
	if bg != ColorEditorBg {
		t.Errorf("expected ColorEditorBg when unfocused, got %v", bg)
	}
}

func TestEditorDraw_ColumnGuides(t *testing.T) {
	screen := newSimScreen(t, 90, 25)

	ed := NewEditor("main.for", 1)
	ed.Lines = []string{
		"     C Continuation marker column test",
		"      EMPTY LINE",
	}
	ed.ShowLineNums = false
	ed.ShowColumnGuides = true
	ed.CursorY = 0
	ed.CursorX = 0

	// Draw without line numbers
	ed.Draw(screen, 0, 1, 90, 20, false)

	// With ShowLineNums = false, codeStartX = 0
	// Line 1 ("      EMPTY LINE") is at screenY = 2.
	// Index 5 (Column 6 in F77 1-based convention) must display RuneColumnGuide if space.
	// Since "      " has spaces at 0..5, index 5 is space, rendered as RuneColumnGuide.
	rCol6, _, styleCol6, _ := screen.GetContent(5, 2)
	if rCol6 != RuneColumnGuide {
		t.Errorf("expected RuneColumnGuide at column index 5 (col 6), got %q (rune %d)", rCol6, rCol6)
	}
	fg6, _, _ := styleCol6.Decompose()
	if fg6 != ColorEditorGuideLine {
		t.Errorf("expected guide line foreground color %v, got %v", ColorEditorGuideLine, fg6)
	}

	// Index 71 (Column 72 in F77 1-based convention) in empty line area must display RuneColumnGuide
	rCol72, _, styleCol72, _ := screen.GetContent(71, 2)
	if rCol72 != RuneColumnGuide {
		t.Errorf("expected RuneColumnGuide at column index 71 (col 72), got %q (rune %d)", rCol72, rCol72)
	}
	fg72, _, _ := styleCol72.Decompose()
	if fg72 != ColorEditorGuideLine {
		t.Errorf("expected guide line foreground color %v, got %v", ColorEditorGuideLine, fg72)
	}

	// Below EOF (e.g. screenY = 5)
	rEofGuide, _, _, _ := screen.GetContent(71, 5)
	if rEofGuide != RuneColumnGuide {
		t.Errorf("expected RuneColumnGuide below EOF at column 71, got %q", rEofGuide)
	}
}
