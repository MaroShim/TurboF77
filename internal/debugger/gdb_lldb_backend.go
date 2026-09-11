package debugger

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"tf77/internal/compiler"
)

const (
	cmdDelimiter = "@@TF77_CMD_END@@"
)

// GdbLldbBackend implements DebuggerBackend using external lldb or gdb subprocesses
type GdbLldbBackend struct {
	mu          sync.Mutex
	tool        compiler.DebuggerInfo
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	lineChan    chan string
	doneChan    chan struct{}
	srcFile     string
	binPath     string
	state       DebugState
	isLldb      bool
	initialized bool
}

// NewGdbLldbBackend creates a backend instance targeting LLDB or GDB
func NewGdbLldbBackend(tool compiler.DebuggerInfo) *GdbLldbBackend {
	isLldb := tool.Kind == "lldb" || strings.Contains(strings.ToLower(tool.Path), "lldb")
	return &GdbLldbBackend{
		tool:   tool,
		isLldb: isLldb,
	}
}

// Start launches the binary under the debugger and stops at the entry or first breakpoint
func (b *GdbLldbBackend) Start(srcFile string, binPath string, bps map[int]bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.srcFile = srcFile
	b.binPath = binPath
	b.state = DebugState{
		Active:      true,
		Running:     false,
		CurrentFile: srcFile,
		CurrentLine: 1,
	}

	var cmd *exec.Cmd
	if b.isLldb {
		cmd = exec.Command(b.tool.Path, "--no-lldbinit", binPath)
	} else {
		cmd = exec.Command(b.tool.Path, "-q", binPath)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open debugger stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return fmt.Errorf("failed to open debugger stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // combine stderr

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return fmt.Errorf("failed to start debugger process: %w", err)
	}

	b.cmd = cmd
	b.stdin = stdin
	b.stdout = stdout
	b.lineChan = make(chan string, 500)
	b.doneChan = make(chan struct{})

	go b.readOutputLoop()

	// Initial configuration
	if b.isLldb {
		_, _ = b.executeCommand("settings set auto-confirm true")
		_, _ = b.executeCommand("settings set stop-line-count-before 0")
		_, _ = b.executeCommand("settings set stop-line-count-after 0")
		// Enable synchronous mode so commands wait for process state change before returning
		_, _ = b.executeCommand("script lldb.debugger.SetAsync(False)")
		// Break at entry routines in Fortran (gfortran uses MAIN__)
		_, _ = b.executeCommand("breakpoint set -n MAIN__")
		_, _ = b.executeCommand("breakpoint set -n main")
	} else {
		_, _ = b.executeCommand("set pagination off")
		_, _ = b.executeCommand("set confirm off")
		_, _ = b.executeCommand("break MAIN__")
		_, _ = b.executeCommand("break main")
	}

	// Set initial user breakpoints
	srcBase := filepath.Base(srcFile)
	for l, set := range bps {
		if set {
			if b.isLldb {
				_, _ = b.executeCommand(fmt.Sprintf("breakpoint set -f %s -l %d", srcBase, l))
			} else {
				_, _ = b.executeCommand(fmt.Sprintf("break %s:%d", srcBase, l))
			}
		}
	}

	// Launch inferior process
	var launchOut string
	if b.isLldb {
		launchOut, err = b.executeCommand("process launch")
	} else {
		launchOut, err = b.executeCommand("run")
	}

	// Check if launch immediately errored (e.g. sandbox attach failure or missing library)
	if err != nil || strings.Contains(launchOut, "attach failed") || strings.Contains(launchOut, "error: process exited with status -1") {
		b.cleanup()
		return fmt.Errorf("failed to launch target under %s: %s", b.tool.Kind, launchOut)
	}

	b.parseCurrentLocation(launchOut)
	if !b.state.Exited {
		b.queryVariables()
	}
	b.initialized = true

	return nil
}

// Continue executes until the next breakpoint or completion
func (b *GdbLldbBackend) Continue() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil || b.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	var cmdStr string
	if b.isLldb {
		cmdStr = "process continue"
	} else {
		cmdStr = "continue"
	}

	out, err := b.executeCommand(cmdStr)
	if err != nil {
		return err
	}

	b.parseCurrentLocation(out)
	if !b.state.Exited {
		b.queryVariables()
	}
	return nil
}

// StepOver executes the current line, stepping over function calls
func (b *GdbLldbBackend) StepOver() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil || b.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	var cmdStr string
	if b.isLldb {
		cmdStr = "thread step-over"
	} else {
		cmdStr = "next"
	}

	out, err := b.executeCommand(cmdStr)
	if err != nil {
		return err
	}

	b.parseCurrentLocation(out)
	if !b.state.Exited {
		b.queryVariables()
	}
	return nil
}

// StepInto advances one statement, stepping into function calls
func (b *GdbLldbBackend) StepInto() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil || b.state.Exited {
		return fmt.Errorf("no active debug session")
	}

	var cmdStr string
	if b.isLldb {
		cmdStr = "thread step-in"
	} else {
		cmdStr = "step"
	}

	out, err := b.executeCommand(cmdStr)
	if err != nil {
		return err
	}

	b.parseCurrentLocation(out)
	if !b.state.Exited {
		b.queryVariables()
	}
	return nil
}

// Stop terminates the debugging session
func (b *GdbLldbBackend) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.cleanup()
	return nil
}

// GetState returns the current state of execution
func (b *GdbLldbBackend) GetState() DebugState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// SetBreakpoint adds or removes a breakpoint
func (b *GdbLldbBackend) SetBreakpoint(line int, enabled bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cmd == nil || b.state.Exited {
		return nil
	}

	srcBase := filepath.Base(b.srcFile)
	var cmdStr string
	if b.isLldb {
		if enabled {
			cmdStr = fmt.Sprintf("breakpoint set -f %s -l %d", srcBase, line)
		} else {
			cmdStr = fmt.Sprintf("breakpoint clear -f %s -l %d", srcBase, line)
		}
	} else {
		if enabled {
			cmdStr = fmt.Sprintf("break %s:%d", srcBase, line)
		} else {
			cmdStr = fmt.Sprintf("clear %s:%d", srcBase, line)
		}
	}

	_, err := b.executeCommand(cmdStr)
	return err
}

func (b *GdbLldbBackend) readOutputLoop() {
	defer close(b.lineChan)
	scanner := bufio.NewScanner(b.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case b.lineChan <- line:
		case <-b.doneChan:
			return
		}
	}
}

func (b *GdbLldbBackend) executeCommand(cmdStr string) (string, error) {
	if b.cmd == nil || b.stdin == nil {
		return "", fmt.Errorf("debugger process is not running")
	}

	var fullCmd string
	if b.isLldb {
		fullCmd = fmt.Sprintf("%s\nscript print('%s')\n", cmdStr, cmdDelimiter)
	} else {
		fullCmd = fmt.Sprintf("%s\necho %s\\n\n", cmdStr, cmdDelimiter)
	}

	if _, err := io.WriteString(b.stdin, fullCmd); err != nil {
		return "", fmt.Errorf("failed to write to debugger: %w", err)
	}

	var outLines []string
	timeout := time.After(4 * time.Second)

	for {
		select {
		case line, ok := <-b.lineChan:
			if !ok {
				return strings.Join(outLines, "\n"), nil
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == cmdDelimiter {
				return strings.Join(outLines, "\n"), nil
			}
			outLines = append(outLines, line)

			// Fast exit detection: if the process reported exit or termination,
			// don't wait for delimiter or timeout
			if reExitLldb.MatchString(trimmed) || reExitGdb.MatchString(trimmed) ||
				strings.Contains(trimmed, "exited normally") {
				return strings.Join(outLines, "\n"), nil
			}
		case <-timeout:
			return strings.Join(outLines, "\n"), nil
		}
	}
}

var (
	reFrameLldb = regexp.MustCompile(`(?i)frame #\d+:[^` + "`" + `]*` + "`" + `?([A-Za-z0-9_]+)?.*? at ([^:]+):(\d+)`)
	reFrameGdb  = regexp.MustCompile(`(?i)(?:#\d+\s+)?(?:[0-9a-fx]+\s+in\s+)?([A-Za-z0-9_]+)\s*(?:\(.*?\))?\s+at\s+([^:]+):(\d+)`)
	reExitLldb  = regexp.MustCompile(`(?i)exited with status\s*=\s*(\d+)`)
	reExitGdb   = regexp.MustCompile(`(?i)exited with code\s*(\d+)`)
)

func (b *GdbLldbBackend) parseCurrentLocation(output string) {
	lines := strings.Split(output, "\n")
	var progOut []string

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}

		// Check process termination
		if m := reExitLldb.FindStringSubmatch(trimmed); len(m) > 1 {
			code, _ := strconv.Atoi(m[1])
			b.state.Exited = true
			b.state.Active = false
			b.state.ExitCode = code
			continue
		}
		if m := reExitGdb.FindStringSubmatch(trimmed); len(m) > 1 {
			code, _ := strconv.Atoi(m[1])
			b.state.Exited = true
			b.state.Active = false
			b.state.ExitCode = code
			continue
		}
		if strings.Contains(trimmed, "exited normally") {
			b.state.Exited = true
			b.state.Active = false
			b.state.ExitCode = 0
			continue
		}

		// Check frame locations
		if m := reFrameLldb.FindStringSubmatch(trimmed); len(m) >= 4 {
			fn := m[1]
			file := m[2]
			lineNum, err := strconv.Atoi(m[3])
			if err == nil && lineNum > 0 {
				b.state.CurrentLine = lineNum
				if fn != "" {
					b.state.CurrentFunc = fn
				}
				if file != "" {
					b.state.CurrentFile = file
				}
			}
			continue
		}

		if m := reFrameGdb.FindStringSubmatch(trimmed); len(m) >= 4 {
			fn := m[1]
			file := m[2]
			lineNum, err := strconv.Atoi(m[3])
			if err == nil && lineNum > 0 {
				b.state.CurrentLine = lineNum
				if fn != "" {
					b.state.CurrentFunc = fn
				}
				if file != "" {
					b.state.CurrentFile = file
				}
			}
			continue
		}

		// Collect user program output by ignoring debugger prompts & metadata
		if !strings.HasPrefix(trimmed, "(lldb)") &&
			!strings.HasPrefix(trimmed, "(gdb)") &&
			!strings.HasPrefix(trimmed, "Process ") &&
			!strings.HasPrefix(trimmed, "Target ") &&
			!strings.HasPrefix(trimmed, "* thread ") &&
			!strings.HasPrefix(trimmed, "-> ") &&
			!strings.HasPrefix(trimmed, "warning:") &&
			!strings.HasPrefix(trimmed, "script print") &&
			!strings.HasPrefix(trimmed, "echo ") &&
			!strings.Contains(trimmed, cmdDelimiter) {
			progOut = append(progOut, l)
		}
	}

	if len(progOut) > 0 {
		if b.state.Output != "" {
			b.state.Output += "\n"
		}
		b.state.Output += strings.Join(progOut, "\n")
	}
}

var reVarLine = regexp.MustCompile(`^(?:\((?P<type>[^)]+)\)\s+)?(?P<name>[A-Za-z0-9_]+)\s*=\s*(?P<val>.*)$`)

func (b *GdbLldbBackend) queryVariables() {
	if b.state.Exited || b.cmd == nil {
		return
	}

	var queryCmd string
	if b.isLldb {
		queryCmd = "frame variable"
	} else {
		queryCmd = "info locals"
	}

	out, err := b.executeCommand(queryCmd)
	if err != nil {
		return
	}

	var vars []Variable
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || strings.HasPrefix(trimmed, "error:") || strings.HasPrefix(trimmed, "warning:") {
			continue
		}
		m := reVarLine.FindStringSubmatch(trimmed)
		if len(m) > 3 {
			vType := m[1]
			if vType == "" {
				vType = "auto"
			}
			vars = append(vars, Variable{
				Name:  m[2],
				Type:  vType,
				Value: m[3],
			})
		}
	}

	// If variables were successfully read, update LocalVars
	if len(vars) > 0 {
		b.state.LocalVars = vars
	}
}

func (b *GdbLldbBackend) cleanup() {
	if b.doneChan != nil {
		select {
		case <-b.doneChan:
		default:
			close(b.doneChan)
		}
	}
	if b.stdin != nil {
		_ = b.stdin.Close()
	}
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
	}
	b.state.Active = false
	b.state.Exited = true
}
