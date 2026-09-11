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
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
		backendType: BackendInternal,
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

	if d.backend != nil && file == d.srcFile {
		_ = d.backend.SetBreakpoint(line, enabled)
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

	bps := make(map[int]bool)
	if fileBps, ok := d.breakpoints[srcFile]; ok {
		for l, set := range fileBps {
			if set {
				bps[l] = true
			}
		}
	}

	d.srcFile = srcFile

	// Priority 1: If binary exists and a native debugger tool (LLDB / GDB) is available, try GdbLldbBackend
	if binPath != "" {
		if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
			dbgTool, hasDbgTool := compiler.FindDebuggerTool(compilerKind)
			if hasDbgTool && (dbgTool.Kind == "lldb" || dbgTool.Kind == "gdb") {
				nativeBackend := NewGdbLldbBackend(dbgTool)
				if err := nativeBackend.Start(srcFile, binPath, bps); err == nil {
					d.backend = nativeBackend
					d.backendType = BackendGdbLldb
					d.state = nativeBackend.GetState()
					return nil
				}
				// If native launch fails (e.g. sandbox restriction or attach issue), fall through to internal interpreter
			}
		}
	}

	// Priority 2 / Fallback: Safe, zero-dependency built-in interpreter
	internalBackend := NewInternalBackend()
	if err := internalBackend.Start(srcFile, binPath, bps); err != nil {
		return fmt.Errorf("failed to start internal debug engine: %w", err)
	}

	d.backend = internalBackend
	d.backendType = BackendInternal
	d.state = internalBackend.GetState()
	return nil
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
