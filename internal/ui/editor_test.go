package ui

import (
	"os"
	"path/filepath"
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

func TestNavigationStack(t *testing.T) {
	ed := NewEditor("file1.f", 1)
	ed.Lines = []string{"line1", "line2", "line3"}
	ed.CursorY = 0
	ed.CursorX = 0

	// Push initial position
	ed.PushNavLocation()

	// Move cursor
	ed.CursorY = 2
	ed.CursorX = 4
	ed.PushNavLocation()

	// Verify duplicate push suppression
	ed.PushNavLocation()
	if len(ed.navBackStack) != 2 {
		t.Fatalf("expected 2 locations in back stack, got %d", len(ed.navBackStack))
	}

	// Move cursor to another location
	ed.CursorY = 1
	ed.CursorX = 2

	// Navigate back -> should return to line 3 (index 2), col 4
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack to succeed")
	}
	if ed.CursorY != 2 || ed.CursorX != 4 {
		t.Errorf("expected Cursor (2, 4), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate forward -> should return to line 2 (index 1), col 2
	if !ed.NavigateForward() {
		t.Fatalf("expected NavigateForward to succeed")
	}
	if ed.CursorY != 1 || ed.CursorX != 2 {
		t.Errorf("expected Cursor (1, 2), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate back twice -> to initial (0, 0)
	ed.NavigateBack()
	if !ed.NavigateBack() {
		t.Fatalf("expected second NavigateBack to succeed")
	}
	if ed.CursorY != 0 || ed.CursorX != 0 {
		t.Errorf("expected Cursor (0, 0), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate back again -> empty stack, should return false
	if ed.NavigateBack() {
		t.Errorf("expected NavigateBack to fail on empty stack")
	}
}

func TestNavigationMultiFile(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.f")
	f2 := filepath.Join(tmpDir, "b.f")
	if err := os.WriteFile(f1, []byte("      SUBROUTINE A\n      END\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("      SUBROUTINE B\n      END\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ed := NewEditor("", 1)
	if err := ed.LoadFile(f1); err != nil {
		t.Fatal(err)
	}
	ed.CursorY = 0
	ed.CursorX = 6

	// Save location before jumping to f2
	ed.PushNavLocation()

	if err := ed.LoadFile(f2); err != nil {
		t.Fatal(err)
	}
	ed.CursorY = 0
	ed.CursorX = 6

	// Navigate back -> should switch back to f1 and restore cursor
	if !ed.NavigateBack() {
		t.Fatalf("expected NavigateBack across files to succeed")
	}
	if ed.FilePath != f1 {
		t.Errorf("expected FilePath %s, got %s", f1, ed.FilePath)
	}
	if ed.CursorY != 0 || ed.CursorX != 6 {
		t.Errorf("expected Cursor (0, 6), got (%d, %d)", ed.CursorY, ed.CursorX)
	}

	// Navigate forward -> should switch to f2
	if !ed.NavigateForward() {
		t.Fatalf("expected NavigateForward across files to succeed")
	}
	if ed.FilePath != f2 {
		t.Errorf("expected FilePath %s, got %s", f2, ed.FilePath)
	}
}

func TestUndoRedo(t *testing.T) {
	ed := NewEditor("", 1)
	ed.Lines = []string{""}
	ed.CursorX = 0
	ed.CursorY = 0

	// Initial typing
	for _, r := range "hello" {
		ed.InsertRune(r)
	}
	if ed.Lines[0] != "hello" {
		t.Fatalf("expected 'hello', got %q", ed.Lines[0])
	}

	// Undo typing 'o'
	if !ed.Undo() {
		t.Fatalf("expected Undo to succeed")
	}
	if ed.Lines[0] != "hell" {
		t.Errorf("expected 'hell' after undo, got %q", ed.Lines[0])
	}

	// Redo typing 'o'
	if !ed.Redo() {
		t.Fatalf("expected Redo to succeed")
	}
	if ed.Lines[0] != "hello" {
		t.Errorf("expected 'hello', got %q", ed.Lines[0])
	}

	// Insert new line
	ed.InsertNewLine()
	for _, r := range "world" {
		ed.InsertRune(r)
	}
	if len(ed.Lines) != 2 || ed.Lines[1] != "world" {
		t.Fatalf("expected line 1 to be 'world', got %v", ed.Lines)
	}

	// Undo 'd'
	ed.Undo()
	if ed.Lines[1] != "worl" {
		t.Errorf("expected 'worl', got %q", ed.Lines[1])
	}

	// Undo until line 2 disappears
	for len(ed.Lines) > 1 {
		if !ed.Undo() {
			t.Fatalf("expected Undo to succeed until newline removed")
		}
	}
	if len(ed.Lines) != 1 || ed.Lines[0] != "hello" {
		t.Errorf("expected single line 'hello', got %v", ed.Lines)
	}
}

