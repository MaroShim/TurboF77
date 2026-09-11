package debugger

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// InternalBackend implements DebuggerBackend using the built-in F77Engine interpreter
type InternalBackend struct {
	mu      sync.Mutex
	engine  *F77Engine
	srcFile string
	state   DebugState
}

// NewInternalBackend creates a new instance of the internal interpreter backend
func NewInternalBackend() *InternalBackend {
	return &InternalBackend{}
}

func (b *InternalBackend) Start(srcFile string, binPath string, bps map[int]bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	rawLines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	engineBps := make(map[int]bool)
	for l, set := range bps {
		if set {
			engineBps[l] = true
		}
	}

	b.engine = NewF77Engine(rawLines, engineBps)
	b.srcFile = srcFile

	b.state = DebugState{
		Active:      true,
		Running:     false,
		Exited:      b.engine.Exited,
		ExitCode:    b.engine.ExitCode,
		CurrentFile: srcFile,
		CurrentLine: b.engine.CurLine,
		CurrentFunc: b.engine.ProgName,
		LocalVars:   b.engine.GetLocalVariables(),
		Output:      b.engine.OutputBuf.String(),
	}

	return nil
}

func (b *InternalBackend) Continue() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.engine == nil || b.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	// Step at least once to advance from current position
	if !b.engine.Step() {
		b.updateState()
		return nil
	}

	maxSteps := 100000
	for count := 0; count < maxSteps; count++ {
		if b.engine.Exited {
			break
		}
		if b.engine.Breakpoints[b.engine.CurLine] {
			break
		}
		if !b.engine.Step() {
			break
		}
	}

	b.updateState()
	return nil
}

func (b *InternalBackend) StepOver() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.engine == nil || b.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	b.engine.Step()
	b.updateState()
	return nil
}

func (b *InternalBackend) StepInto() error {
	// For internal interpreter, StepInto behaves similarly to Step
	return b.StepOver()
}

func (b *InternalBackend) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = DebugState{
		Active: false,
		Exited: true,
	}
	b.engine = nil
	return nil
}

func (b *InternalBackend) GetState() DebugState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *InternalBackend) SetBreakpoint(line int, enabled bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.engine != nil {
		if enabled {
			b.engine.Breakpoints[line] = true
		} else {
			delete(b.engine.Breakpoints, line)
		}
	}
	return nil
}

func (b *InternalBackend) updateState() {
	if b.engine == nil {
		return
	}
	b.state.Active = b.engine.Active
	b.state.Exited = b.engine.Exited
	b.state.ExitCode = b.engine.ExitCode
	b.state.CurrentLine = b.engine.CurLine
	b.state.CurrentFunc = b.engine.ProgName
	b.state.LocalVars = b.engine.GetLocalVariables()
	b.state.Output = b.engine.OutputBuf.String()
}
