package debugger

import (
	"fmt"
	"os"
	"sync"

	"tf77/internal/compiler"
)

// Variable represents a variable displayed in Watch window
type Variable struct {
	Name  string
	Type  string
	Value string
}

// DebugState holds the current runtime state of the debugger
type DebugState struct {
	Active       bool
	Running      bool
	Exited       bool
	ExitCode     int
	CurrentFile  string
	CurrentLine  int
	CurrentFunc  string
	LocalVars    []Variable
	ErrorMessage string
	Output       string
}

// Debuger manages a debug session for FORTRAN 77 code
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	srcFile     string
	backend     DebuggerBackend
	backendType BackendType
	prefEngine  BackendType // preferred engine: BackendInternal (default) or BackendGdbLldb
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
		backendType: BackendInternal,
		prefEngine:  BackendInternal, // Default to built-in interpreter for 100% accurate variable inspection
	}
}

func (d *Debugger) IsActive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.Active && !d.state.Exited
}

func (d *Debugger) GetState() DebugState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

func (d *Debugger) GetBackendType() BackendType {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.backendType
}

func (d *Debugger) GetPreferredEngine() BackendType {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.prefEngine
}

func (d *Debugger) SetPreferredEngine(eng BackendType) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prefEngine = eng
}

func (d *Debugger) ToggleEngine() BackendType {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.prefEngine == BackendInternal {
		d.prefEngine = BackendGdbLldb
	} else {
		d.prefEngine = BackendInternal
	}
	return d.prefEngine
}

func (d *Debugger) ClearBreakpoints() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.breakpoints = make(map[string]map[int]bool)
}

func (d *Debugger) SetBreakpoint(file string, line int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.breakpoints[file] == nil {
		d.breakpoints[file] = make(map[int]bool)
	}
	d.breakpoints[file][line] = true
	if d.backend != nil {
		_ = d.backend.SetBreakpoint(file, line, true)
	}
}

func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.breakpoints[file] == nil {
		d.breakpoints[file] = make(map[int]bool)
	}

	enabled := false
	if d.breakpoints[file][line] {
		delete(d.breakpoints[file], line)
		enabled = false
	} else {
		d.breakpoints[file][line] = true
		enabled = true
	}

	if d.backend != nil {
		_ = d.backend.SetBreakpoint(file, line, enabled)
	}
	return enabled
}

// StartSession initiates an interactive debug session.
// Optional extra arguments:
//
//	extra[0]: binPath (compiled binary executable path)
//	extra[1]: compilerKind ("gfortran", "wfl386", etc.)
func (d *Debugger) StartSession(srcFile string, extra ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var binPath string
	var compilerKind string
	if len(extra) > 0 {
		binPath = extra[0]
	}
	if len(extra) > 1 {
		compilerKind = extra[1]
	}

	d.srcFile = srcFile

	// Helper to start native backend
	startNative := func() error {
		if binPath == "" {
			return fmt.Errorf("no binary compiled for native debug")
		}
		if fi, err := os.Stat(binPath); err != nil || fi.IsDir() {
			return fmt.Errorf("binary executable does not exist: %s", binPath)
		}
		dbgTool, hasDbgTool := compiler.FindDebuggerTool(compilerKind)
		if !hasDbgTool || (dbgTool.Kind != "lldb" && dbgTool.Kind != "gdb") {
			return fmt.Errorf("no native debugger tool found")
		}
		nativeBackend := NewGdbLldbBackend(dbgTool)
		if err := nativeBackend.Start(srcFile, binPath, d.breakpoints); err != nil {
			return err
		}
		d.backend = nativeBackend
		d.backendType = BackendGdbLldb
		d.state = nativeBackend.GetState()
		return nil
	}

	// Helper to start internal interpreter backend
	startInternal := func() error {
		internalBackend := NewInternalBackend()
		if err := internalBackend.Start(srcFile, binPath, d.breakpoints); err != nil {
			return err
		}
		d.backend = internalBackend
		d.backendType = BackendInternal
		d.state = internalBackend.GetState()
		return nil
	}

	// 1. If user explicitly selected Native engine (LLDB/GDB), try native first
	if d.prefEngine == BackendGdbLldb {
		if err := startNative(); err == nil {
			return nil
		}
		// If native failed, fall back to internal
		if err := startInternal(); err == nil {
			return nil
		}
	} else {
		// 2. Default: prefEngine == BackendInternal
		// Safe, instant, and provides 100% accurate variable inspection in Watches window!
		if err := startInternal(); err == nil {
			return nil
		}
		// If internal failed, fall back to native if binary is available
		if err := startNative(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to start debug engine (both internal and native failed)")
}

// Continue executes until the next breakpoint or program completion
func (d *Debugger) Continue() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.backend == nil || d.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	err := d.backend.Continue()
	d.state = d.backend.GetState()
	return err
}

// Step (Trace Into) advances one statement
func (d *Debugger) Step() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.backend == nil || d.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	err := d.backend.StepInto()
	d.state = d.backend.GetState()
	return err
}

// Next (Step Over) advances one statement or steps over function/subroutine
func (d *Debugger) Next() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.backend == nil || d.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	err := d.backend.StepOver()
	d.state = d.backend.GetState()
	return err
}

// Stop terminates the active debug session
func (d *Debugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.backend != nil {
		_ = d.backend.Stop()
	}
	d.state = DebugState{
		Active: false,
		Exited: true,
	}
	d.backend = nil
}
