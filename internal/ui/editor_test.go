package ui

import (
	"testing"
)

func TestEditorOperations(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{""}
	ed.CursorX = 0
	ed.CursorY = 0

	// 1. Insert characters
	for _, r := range "PROGRAM TEST" {
		ed.InsertRune(r)
	}
	if ed.Lines[0] != "PROGRAM TEST" {
		t.Errorf("expected 'PROGRAM TEST', got %q", ed.Lines[0])
	}
	if ed.CursorX != 12 {
		t.Errorf("expected cursorX 12, got %d", ed.CursorX)
	}

	// 2. Insert new line without indentation
	ed.InsertNewLine()
	if len(ed.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(ed.Lines))
	}
	if ed.CursorY != 1 {
		t.Errorf("expected cursorY 1, got %d", ed.CursorY)
	}

	// 3. Insert tab (4 spaces to reach tab stop 4)
	ed.InsertTab()
	if ed.CursorX != 4 || ed.Lines[1] != "    " {
		t.Errorf("expected 4 spaces, got %q (cursorX=%d)", ed.Lines[1], ed.CursorX)
	}

	// 4. Smart Backspace: deletes 4 spaces indentation at once
	ed.Backspace()
	if ed.CursorX != 0 || ed.Lines[1] != "" {
		t.Errorf("expected 0 spaces after smart backspace, got %q (cursorX=%d)", ed.Lines[1], ed.CursorX)
	}

	// Test ExpandTabs
	tabbed := "\tPRINT *, 'OK'"
	expanded := ExpandTabs(tabbed, 4)
	if expanded != "    PRINT *, 'OK'" {
		t.Errorf("expected expanded tab to 4 spaces, got %q", expanded)
	}

	// 5. Breakpoint toggling
	isBp := ed.ToggleBreakpoint(1)
	if !isBp || !ed.Breakpoints[1] {
		t.Errorf("expected breakpoint at line 1 to be set")
	}
	isBp = ed.ToggleBreakpoint(1)
	if isBp || ed.Breakpoints[1] {
		t.Errorf("expected breakpoint at line 1 to be cleared")
	}
}
