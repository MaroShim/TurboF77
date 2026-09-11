package debugger

// BackendType represents the type of debugger engine in use
type BackendType string

const (
	BackendInternal BackendType = "internal"
	BackendGdbLldb  BackendType = "gdb_lldb"
	BackendWatcom   BackendType = "watcom"
)

// DebuggerBackend defines the common interface for all Fortran debugger implementations
type DebuggerBackend interface {
	// Start initializes and begins execution up to the entry point or first breakpoint
	Start(srcFile string, binPath string, bps map[int]bool) error
	// Continue runs until the next breakpoint, exit, or error
	Continue() error
	// StepOver executes the current line, stepping over subroutine/function calls
	StepOver() error
	// StepInto executes the next statement, stepping into subroutine/function calls
	StepInto() error
	// Stop terminates the active debug session
	Stop() error
	// GetState returns the current execution state, variables, and output
	GetState() DebugState
	// SetBreakpoint adds or removes a breakpoint at the given line
	SetBreakpoint(line int, enabled bool) error
}
