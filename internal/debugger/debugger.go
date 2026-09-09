package debugger

import (
	"fmt"
	"os"
	"strings"
	"sync"
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

// Debugger manages a debug session for FORTRAN 77 code
type Debugger struct {
	mu          sync.Mutex
	breakpoints map[string]map[int]bool // file -> lines
	state       DebugState
	srcFile     string
	engine      *F77Engine
}

func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]map[int]bool),
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

func (d *Debugger) ToggleBreakpoint(file string, line int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.breakpoints[file] == nil {
		d.breakpoints[file] = make(map[int]bool)
	}
	if d.breakpoints[file][line] {
		delete(d.breakpoints[file], line)
		if d.engine != nil {
			delete(d.engine.Breakpoints, line)
		}
		return false
	} else {
		d.breakpoints[file][line] = true
		if d.engine != nil {
			d.engine.Breakpoints[line] = true
		}
		return true
	}
}

// StartSession initiates an interactive debug session for the given Fortran file
func (d *Debugger) StartSession(srcFile string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	rawLines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	bps := make(map[int]bool)
	if fileBps, ok := d.breakpoints[srcFile]; ok {
		for l, set := range fileBps {
			if set {
				bps[l] = true
			}
		}
	}

	d.engine = NewF77Engine(rawLines, bps)
	d.srcFile = srcFile

	d.state = DebugState{
		Active:      true,
		Running:     false,
		Exited:      d.engine.Exited,
		ExitCode:    d.engine.ExitCode,
		CurrentFile: srcFile,
		CurrentLine: d.engine.CurLine,
		CurrentFunc: d.engine.ProgName,
		LocalVars:   d.engine.GetLocalVariables(),
		Output:      d.engine.OutputBuf.String(),
	}

	return nil
}

// Continue executes until the next breakpoint or program completion
func (d *Debugger) Continue() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.engine == nil || d.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	// Step at least once to move off current line
	if !d.engine.Step() {
		d.updateStateFromEngine()
		return nil
	}

	// Continue until breakpoint or exit
	maxSteps := 100000
	for count := 0; count < maxSteps; count++ {
		if d.engine.Exited {
			break
		}
		if d.engine.Breakpoints[d.engine.CurLine] {
			break
		}
		if !d.engine.Step() {
			break
		}
	}

	d.updateStateFromEngine()
	return nil
}

// Step (Trace Into) advances one statement
func (d *Debugger) Step() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.engine == nil || d.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	d.engine.Step()
	d.updateStateFromEngine()
	return nil
}

// Next (Step Over) advances one statement or steps through loop body
func (d *Debugger) Next() error {
	// For Fortran single-file routines, Next behaves similarly to Step
	return d.Step()
}

// Stop terminates the active debug session
func (d *Debugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.state = DebugState{
		Active: false,
		Exited: true,
	}
	d.engine = nil
}

func (d *Debugger) updateStateFromEngine() {
	if d.engine == nil {
		return
	}
	d.state.Active = d.engine.Active
	d.state.Exited = d.engine.Exited
	d.state.ExitCode = d.engine.ExitCode
	d.state.CurrentLine = d.engine.CurLine
	d.state.CurrentFunc = d.engine.ProgName
	d.state.LocalVars = d.engine.GetLocalVariables()
	d.state.Output = d.engine.OutputBuf.String()
}
