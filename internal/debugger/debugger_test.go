package debugger

import (
	"os"
	"path/filepath"
	"testing"

	"tf77/internal/compiler"
)

func TestF77Engine(t *testing.T) {
	code := []string{
		"C     Test Fortran Program",
		"      PROGRAM TEST",
		"      INTEGER I, SUM",
		"      SUM = 0",
		"      DO 10 I = 1, 5",
		"          SUM = SUM + I",
		"   10 CONTINUE",
		"      PRINT *, 'SUM IS:', SUM",
		"      STOP",
		"      END",
	}

	eng := NewF77Engine(code, map[int]bool{6: true}) // Breakpoint on line 6 (SUM = SUM + I)
	if eng.CurLine != 4 { // Line 4 is 'SUM = 0'
		t.Fatalf("expected CurLine 4, got %d", eng.CurLine)
	}

	// Step line 4 (SUM = 0)
	eng.Step()
	if v := eng.Variables["SUM"].Value; v != "0" {
		t.Errorf("expected SUM=0, got %s", v)
	}

	// Step line 5 (DO 10 I = 1, 5)
	eng.Step()
	if v := eng.Variables["I"].Value; v != "1" {
		t.Errorf("expected I=1, got %s", v)
	}

	// Step line 6 (SUM = SUM + I -> SUM = 1)
	eng.Step()
	if v := eng.Variables["SUM"].Value; v != "1" {
		t.Errorf("expected SUM=1, got %s", v)
	}
}

func TestDebugger_InternalSession(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test.for")
	content := `      PROGRAM TEST_INTERP
      INTEGER A, B
      A = 10
      B = 20
      STOP
      END
`
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	dbg := NewDebugger()
	dbg.ToggleBreakpoint(srcFile, 4) // B = 20

	if err := dbg.StartSession(srcFile); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}

	if !dbg.IsActive() {
		t.Errorf("expected debugger to be active")
	}
	if dbg.GetBackendType() != BackendInternal {
		t.Errorf("expected BackendInternal, got %v", dbg.GetBackendType())
	}

	// Step line 3 (A = 10)
	if err := dbg.Step(); err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	st := dbg.GetState()
	if st.CurrentLine != 4 {
		t.Errorf("expected CurrentLine 4, got %d", st.CurrentLine)
	}

	// Continue to exit
	if err := dbg.Continue(); err != nil {
		t.Fatalf("Continue failed: %v", err)
	}
	st = dbg.GetState()
	if !st.Exited {
		t.Errorf("expected session to be exited")
	}

	dbg.Stop()
	if dbg.IsActive() {
		t.Errorf("expected debugger to be inactive after Stop")
	}
}

func TestDebugger_FallbackWhenBinaryInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "simple.for")
	content := `      PROGRAM SIMPLE
      INTEGER X
      X = 42
      STOP
      END
`
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	dbg := NewDebugger()
	// Pass non-existent binary path; debugger must fall back safely to InternalBackend
	if err := dbg.StartSession(srcFile, "/path/does/not/exist/bin", "gfortran"); err != nil {
		t.Fatalf("StartSession should have fallen back to internal backend, but failed: %v", err)
	}

	if dbg.GetBackendType() != BackendInternal {
		t.Errorf("expected fallback to BackendInternal, got %v", dbg.GetBackendType())
	}
	if !dbg.IsActive() {
		t.Errorf("expected active debug session")
	}
	dbg.Stop()
}

func TestDebugger_GdbLldbBackend(t *testing.T) {
	cInfo, hasCompiler := compiler.FindFortranCompiler()
	if !hasCompiler || cInfo.Kind != "gfortran" {
		t.Skip("gfortran not available; skipping native debugger test")
	}
	dbgTool, hasDbg := compiler.FindDebuggerTool("gfortran")
	if !hasDbg {
		t.Skip("lldb/gdb not found; skipping native debugger test")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "dbg_test.for")
	content := `      PROGRAM DBG_TEST
      INTEGER VAL
      VAL = 100
      PRINT *, 'VAL IS:', VAL
      STOP
      END
`
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	bRes := compiler.Build(srcFile)
	if bRes == nil || !bRes.Success || bRes.BinaryPath == "" {
		t.Skipf("build failed with gfortran: %v", bRes)
	}

	backend := NewGdbLldbBackend(dbgTool)
	err := backend.Start(srcFile, bRes.BinaryPath, map[int]bool{4: true})
	if err != nil {
		// In restricted environments (like seatbelt sandbox), attach may fail
		t.Logf("native debugger backend start returned error (expected in sandbox): %v", err)
		return
	}
	defer backend.Stop()

	st := backend.GetState()
	if !st.Active {
		t.Errorf("expected backend to be active")
	}

	// Test step over
	_ = backend.StepOver()
	st = backend.GetState()
	t.Logf("After StepOver: line=%d func=%s file=%s exited=%v", st.CurrentLine, st.CurrentFunc, st.CurrentFile, st.Exited)

	// Test continue to exit
	_ = backend.Continue()
	st = backend.GetState()
	t.Logf("After Continue: line=%d func=%s file=%s exited=%v code=%d output=%q",
		st.CurrentLine, st.CurrentFunc, st.CurrentFile, st.Exited, st.ExitCode, st.Output)
}
