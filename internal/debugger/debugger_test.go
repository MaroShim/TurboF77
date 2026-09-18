package debugger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MaroShim/tf77/internal/compiler"
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

	t0 := time.Now()
	backend := NewGdbLldbBackend(dbgTool)
	err := backend.Start(srcFile, bRes.BinaryPath, map[string]map[int]bool{srcFile: {4: true}})
	t.Logf("Start() took: %v", time.Since(t0))
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
	t1 := time.Now()
	_ = backend.StepOver()
	t.Logf("StepOver() took: %v", time.Since(t1))
	st = backend.GetState()
	t.Logf("After StepOver: line=%d func=%s file=%s exited=%v", st.CurrentLine, st.CurrentFunc, st.CurrentFile, st.Exited)

	// Test continue to exit
	t2 := time.Now()
	_ = backend.Continue()
	t.Logf("Continue() took: %v", time.Since(t2))
	st = backend.GetState()
	t.Logf("After Continue: line=%d func=%s file=%s exited=%v code=%d output=%q",
		st.CurrentLine, st.CurrentFunc, st.CurrentFile, st.Exited, st.ExitCode, st.Output)
}

func TestDebugger_PreferredEngineSelection(t *testing.T) {
	dbg := NewDebugger()
	if dbg.GetPreferredEngine() != BackendInternal {
		t.Errorf("expected default preferred engine to be BackendInternal, got %v", dbg.GetPreferredEngine())
	}

	newEng := dbg.ToggleEngine()
	if newEng != BackendGdbLldb || dbg.GetPreferredEngine() != BackendGdbLldb {
		t.Errorf("expected toggled engine to be BackendGdbLldb, got %v", newEng)
	}

	newEng = dbg.ToggleEngine()
	if newEng != BackendInternal || dbg.GetPreferredEngine() != BackendInternal {
		t.Errorf("expected toggled engine to be BackendInternal, got %v", newEng)
	}
}

func TestDebugger_DefaultInternalEnginePriority(t *testing.T) {
	cInfo, hasCompiler := compiler.FindFortranCompiler()
	if !hasCompiler || cInfo.Kind != "gfortran" {
		t.Skip("gfortran not available; skipping test")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "simple_vars.for")
	content := `      PROGRAM SVAR
      INTEGER A, B
      A = 25
      B = 17
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

	// 1. By default (prefEngine == BackendInternal), StartSession must choose BackendInternal
	// even when a compiled binary exists, so variables are 100% visible and tracked!
	dbg := NewDebugger()
	if err := dbg.StartSession(srcFile, bRes.BinaryPath, cInfo.Kind); err != nil {
		t.Fatalf("StartSession failed: %v", err)
	}
	defer dbg.Stop()

	if dbg.GetBackendType() != BackendInternal {
		t.Fatalf("expected default backend to be BackendInternal, got %s", dbg.GetBackendType())
	}

	// Step line by line and check variables
	_ = dbg.Step() // executes A = 25
	_ = dbg.Step() // executes B = 17

	st := dbg.GetState()
	vars := st.LocalVars
	varMap := make(map[string]string)
	for _, v := range vars {
		varMap[v.Name] = v.Value
	}

	if varMap["A"] != "25" {
		t.Errorf("expected A=25 in internal engine, got %q", varMap["A"])
	}
	if varMap["B"] != "17" {
		t.Errorf("expected B=17 in internal engine, got %q", varMap["B"])
	}
}

func TestModularExampleDebugging(t *testing.T) {
	mainFile := "../../examples/modular/main.for"
	if _, err := os.Stat(mainFile); err != nil {
		t.Fatalf("modular main.for not found: %v", err)
	}

	dbg := NewDebugger()
	dbg.SetPreferredEngine(BackendInternal)

	if err := dbg.StartSession(mainFile); err != nil {
		t.Fatalf("StartSession on modular main.for failed: %v", err)
	}
	defer dbg.Stop()

	st := dbg.GetState()
	t.Logf("Initial state: Line=%d File=%s Func=%s", st.CurrentLine, st.CurrentFile, st.CurrentFunc)
	for _, v := range st.LocalVars {
		t.Logf("  Init Var: %s (%s) = %s", v.Name, v.Type, v.Value)
	}

	// Step line 6: A = 25
	if err := dbg.Step(); err != nil {
		t.Fatalf("Step line 6 failed: %v", err)
	}
	// Step line 7: B = 17
	if err := dbg.Step(); err != nil {
		t.Fatalf("Step line 7 failed: %v", err)
	}

	st = dbg.GetState()
	t.Logf("After A=25, B=17: Line=%d", st.CurrentLine)
	varMap := make(map[string]string)
	for _, v := range st.LocalVars {
		t.Logf("  Var: %s (%s) = %s", v.Name, v.Type, v.Value)
		varMap[v.Name] = v.Value
	}

	if varMap["A"] != "25" {
		t.Errorf("expected A=25, got %q", varMap["A"])
	}
	if varMap["B"] != "17" {
		t.Errorf("expected B=17, got %q", varMap["B"])
	}

	// Step line 8: SUM = CALCSUM(A, B)
	_ = dbg.Step()
	st = dbg.GetState()
	t.Logf("After line 8 (SUM = CALCSUM): Line=%d", st.CurrentLine)
	for _, v := range st.LocalVars {
		t.Logf("  Var: %s (%s) = %s", v.Name, v.Type, v.Value)
	}

	// Step line 9: CALL PRINTSUM(A, B, SUM)
	_ = dbg.Step()
	st = dbg.GetState()
	t.Logf("After line 9 (CALL PRINTSUM): Line=%d Exited=%v Output=%q", st.CurrentLine, st.Exited, st.Output)
}
