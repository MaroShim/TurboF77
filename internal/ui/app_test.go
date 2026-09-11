package ui

import (
	"path/filepath"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAppMultiFileDebugging(t *testing.T) {
	modularMain, err := filepath.Abs("../../examples/modular/main.for")
	if err != nil {
		t.Fatalf("failed to resolve modular main path: %v", err)
	}
	modularMath := filepath.Join(filepath.Dir(modularMain), "math_sub.for")
	modularIO := filepath.Join(filepath.Dir(modularMain), "io_sub.for")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, modularMain)
	defer app.StopDebugging()

	// 1. Set a breakpoint in math_sub.for:6 (CALCSUM = X + Y)
	if err := app.editor.LoadFile(modularMath); err != nil {
		t.Fatalf("failed to load math_sub.for: %v", err)
	}
	app.ToggleBreakpoint(6)

	// 2. Set a breakpoint in io_sub.for:7 (PRINT *, 'Modular F77...')
	if err := app.editor.LoadFile(modularIO); err != nil {
		t.Fatalf("failed to load io_sub.for: %v", err)
	}
	app.ToggleBreakpoint(7)

	// 3. Switch back to main.for and set breakpoint at line 9 (CALL PRINTSUM)
	if err := app.editor.LoadFile(modularMain); err != nil {
		t.Fatalf("failed to load main.for: %v", err)
	}
	app.ToggleBreakpoint(9)

	// Verify FileBreakpoints recorded all files
	if len(app.editor.FileBreakpoints[filepath.Clean(modularMath)]) == 0 {
		t.Errorf("expected math_sub.for breakpoints to be stored in FileBreakpoints")
	}
	if len(app.editor.FileBreakpoints[filepath.Clean(modularIO)]) == 0 {
		t.Errorf("expected io_sub.for breakpoints to be stored in FileBreakpoints")
	}

	// 4. Start debugging from main.for
	if err := app.StartDebugging(); err != nil {
		t.Fatalf("StartDebugging failed: %v", err)
	}

	// 5. CALCSUM is called on line 8 of main.for, so it should stop first at math_sub.for:6
	cleanMath := filepath.Clean(modularMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math_sub.for, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 6 {
		t.Fatalf("expected CurrentIP 6 in math_sub.for, got %d", app.editor.CurrentIP)
	}

	// 6. Continue: should hit main.for:9 (CALL PRINTSUM)
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue failed: %v", err)
	}
	cleanMain := filepath.Clean(modularMain)
	if filepath.Clean(app.editor.FilePath) != cleanMain {
		t.Fatalf("expected editor to switch back to main.for, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 9 {
		t.Fatalf("expected CurrentIP 9 in main.for, got %d", app.editor.CurrentIP)
	}

	// 7. Continue: should hit io_sub.for:7
	if err := app.DebugContinue(); err != nil {
		t.Fatalf("DebugContinue failed: %v", err)
	}
	cleanIO := filepath.Clean(modularIO)
	if filepath.Clean(app.editor.FilePath) != cleanIO {
		t.Fatalf("expected editor to switch to io_sub.for, but is at %s", app.editor.FilePath)
	}
	if app.editor.CurrentIP != 7 {
		t.Fatalf("expected CurrentIP 7 in io_sub.for, got %d", app.editor.CurrentIP)
	}
}

func TestAppF7StepIntoSubroutines(t *testing.T) {
	modularMain, err := filepath.Abs("../../examples/modular/main.for")
	if err != nil {
		t.Fatalf("failed to resolve modular main path: %v", err)
	}
	modularMath := filepath.Join(filepath.Dir(modularMain), "math_sub.for")
	modularIO := filepath.Join(filepath.Dir(modularMain), "io_sub.for")

	simScreen := tcell.NewSimulationScreen("")
	if err := simScreen.Init(); err != nil {
		t.Fatalf("failed to init sim screen: %v", err)
	}
	simScreen.SetSize(80, 25)

	app := NewAppWithScreen(simScreen, modularMain)
	defer app.StopDebugging()

	// Set a breakpoint on line 8 of main.for (SUM = CALCSUM(A, B))
	app.ToggleBreakpoint(8)

	if err := app.StartDebugging(); err != nil {
		t.Fatalf("StartDebugging failed: %v", err)
	}

	st := app.debugger.GetState()
	t.Logf("Backend=%s, State=%+v, EditorFile=%s, EditorIP=%d", app.debugger.GetBackendType(), st, app.editor.FilePath, app.editor.CurrentIP)

	if app.editor.CurrentIP != 8 {
		t.Fatalf("expected CurrentIP 8 in main.for, got %d", app.editor.CurrentIP)
	}

	// F7 (Step Into) on line 8 -> enters CALCSUM in math_sub.for
	if err := app.DebugStepInto(); err != nil {
		t.Fatalf("DebugStepInto failed: %v", err)
	}

	cleanMath := filepath.Clean(modularMath)
	if filepath.Clean(app.editor.FilePath) != cleanMath {
		t.Fatalf("expected editor to switch to math_sub.for on F7, but is at %s", app.editor.FilePath)
	}

	// Step through CALCSUM until returning to main.for
	for i := 0; i < 5; i++ {
		st := app.debugger.GetState()
		if filepath.Clean(app.editor.FilePath) == filepath.Clean(modularMain) {
			break
		}
		if err := app.DebugStepInto(); err != nil {
			t.Fatalf("step %d in CALCSUM failed: %v", i, err)
		}
		t.Logf("CALCSUM step %d: file=%s, line=%d", i, filepath.Base(app.editor.FilePath), st.CurrentLine)
	}

	cleanMain := filepath.Clean(modularMain)
	if filepath.Clean(app.editor.FilePath) != cleanMain {
		t.Fatalf("expected editor to return to main.for after CALCSUM, but is at %s", app.editor.FilePath)
	}

	// Advance to line 9 (CALL PRINTSUM) if still on line 8
	if app.editor.CurrentIP == 8 {
		if err := app.DebugStepInto(); err != nil {
			t.Fatalf("step to line 9 failed: %v", err)
		}
	}

	// Now on line 9 (CALL PRINTSUM): F7 enters PRINTSUM in io_sub.for
	if err := app.DebugStepInto(); err != nil {
		t.Fatalf("DebugStepInto on line 9 failed: %v", err)
	}

	cleanIO := filepath.Clean(modularIO)
	if filepath.Clean(app.editor.FilePath) != cleanIO {
		t.Fatalf("expected editor to switch to io_sub.for on F7, but is at %s", app.editor.FilePath)
	}

	// Stepping in io_sub.for over PRINT * should stay in io_sub.for (no libgfortran leak)
	for i := 0; i < 3; i++ {
		if err := app.DebugStepInto(); err != nil {
			break
		}
		if filepath.Clean(app.editor.FilePath) != cleanIO {
			t.Fatalf("step %d in io_sub.for leaked to %s (expected io_sub.for)", i, app.editor.FilePath)
		}
	}
}
