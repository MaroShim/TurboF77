package compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// CompileError holds parsed compiler diagnostic information
type CompileError struct {
	File    string
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Message string
}

// BuildResult holds outcome of the compile and link process
type BuildResult struct {
	Success       bool
	CompilerFound bool
	CompilerName  string
	LinesCompiled int
	Duration      time.Duration
	Errors        []CompileError
	ErrorCount    int
	WarningCount  int
	BinaryPath    string
	RawOutput     string
}

// RunResult holds outcome of program execution
type RunResult struct {
	Output    string
	ExitCode  int
	Duration  time.Duration
	Completed bool
}

// Regex patterns for Open Watcom & standard Fortran diagnostics
// 1. Watcom standard: file.for(line): Error! E1234: message OR file.for(line): Warning! W1234: message
var watcomErrRegex1 = regexp.MustCompile(`(?m)^((?:[a-zA-Z]:)?[^:\n\r()]+)\((\d+)\):\s*(Error!|Warning!|\*ERR\*|\*WRN\*)\s*(?:([EW]\d+):\s*)?(.+)$`)

// 2. Watcom alternate column format: file.for(line,col): Error! ...
var watcomErrRegex2 = regexp.MustCompile(`(?m)^((?:[a-zA-Z]:)?[^:\n\r()]+)\((\d+),(\d+)\):\s*(Error!|Warning!|\*ERR\*|\*WRN\*)\s*(?:([EW]\d+):\s*)?(.+)$`)

// 3. Standard Unix / gfortran format: file.f:line:col: Error: message
var stdErrRegex = regexp.MustCompile(`(?m)^((?:[a-zA-Z]:)?[^:\n\r]+):(\d+):(?:(\d+):)?\s*(error|warning):\s*(.+)$`)

// CompilerInfo contains path and kind of detected compiler
type CompilerInfo struct {
	Path string
	Kind string // "wfl386", "wfc386", "wfl", "gfortran", etc.
}

// FindFortranCompiler locates Open Watcom 2.0 FORTRAN 77 or compatible compiler
func FindFortranCompiler() (CompilerInfo, bool) {
	// 1. Check WATCOM environment variable
	if watcomDir := os.Getenv("WATCOM"); watcomDir != "" {
		candidates := []string{
			filepath.Join(watcomDir, "binnt64", "wfl386.exe"),
			filepath.Join(watcomDir, "binnt", "wfl386.exe"),
			filepath.Join(watcomDir, "binl64", "wfl386"),
			filepath.Join(watcomDir, "binl", "wfl386"),
			filepath.Join(watcomDir, "binarm64", "wfl386"),
			filepath.Join(watcomDir, "binw", "wfl.exe"),
			filepath.Join(watcomDir, "binnt", "wfc386.exe"),
			filepath.Join(watcomDir, "binl64", "wfc386"),
			filepath.Join(watcomDir, "binl", "wfc386"),
		}
		for _, c := range candidates {
			if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
				kind := "wfl386"
				if strings.Contains(filepath.Base(c), "wfc") {
					kind = "wfc386"
				} else if strings.HasPrefix(filepath.Base(c), "wfl.") {
					kind = "wfl"
				}
				return CompilerInfo{Path: c, Kind: kind}, true
			}
		}
	}

	// 2. Check system PATH for Open Watcom tools
	watcomBins := []string{"wfl386", "wfl386.exe", "wfc386", "wfc386.exe", "wfl", "wfl.exe", "owcc", "owcc.exe"}
	for _, bin := range watcomBins {
		if path, err := exec.LookPath(bin); err == nil {
			kind := "wfl386"
			if strings.Contains(bin, "wfc") {
				kind = "wfc386"
			} else if strings.HasPrefix(bin, "wfl.") || bin == "wfl" {
				kind = "wfl"
			} else if strings.Contains(bin, "owcc") {
				kind = "owcc"
			}
			return CompilerInfo{Path: path, Kind: kind}, true
		}
	}

	// 3. Check common standard installation locations
	standardPaths := []string{
		"/opt/watcom/binl64/wfl386",
		"/opt/watcom/binl/wfl386",
		"/usr/local/watcom/binl64/wfl386",
		"/usr/watcom/binl/wfl386",
		"C:\\WATCOM\\BINNT\\wfl386.exe",
		"C:\\WATCOM\\BINW\\wfl.exe",
	}
	for _, p := range standardPaths {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return CompilerInfo{Path: p, Kind: "wfl386"}, true
		}
	}

	// 4. Check for fallback compilers (gfortran / f77)
	fallbackBins := []string{"gfortran", "gfortran.exe", "f77"}
	for _, bin := range fallbackBins {
		if path, err := exec.LookPath(bin); err == nil {
			return CompilerInfo{Path: path, Kind: bin}, true
		}
	}

	return CompilerInfo{}, false
}

// IsFortranSource checks if a filename is a Fortran 77 source file
func IsFortranSource(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".for" || ext == ".f" || ext == ".f77" || ext == ".for77" || ext == ".inc"
}

// CountLines counts total lines of Fortran code in the specified target or directory
func CountLines(targetPath string) int {
	total := 0
	info, err := os.Stat(targetPath)
	if err != nil {
		return 0
	}

	if info.IsDir() {
		_ = filepath.Walk(targetPath, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && IsFortranSource(path) {
				content, err := os.ReadFile(path)
				if err == nil {
					total += bytes.Count(content, []byte("\n")) + 1
				}
			}
			return nil
		})
		return total
	}

	content, err := os.ReadFile(targetPath)
	if err == nil {
		return bytes.Count(content, []byte("\n")) + 1
	}
	return 0
}

// ParseErrors extracts CompileError list from compiler diagnostic text
func ParseErrors(output string, workDir string) []CompileError {
	var errs []CompileError
	lines := strings.Split(output, "\n")

	resolvePath := func(p string) string {
		p = strings.TrimSpace(p)
		if filepath.IsAbs(p) {
			return p
		}
		if workDir != "" {
			return filepath.Join(workDir, p)
		}
		return p
	}

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if line == "" {
			continue
		}

		// 1. Watcom standard: file.for(line): Error! E1234: message
		if m := watcomErrRegex1.FindStringSubmatch(line); len(m) >= 6 {
			file := resolvePath(m[1])
			lNum, _ := strconv.Atoi(m[2])
			lvlStr := strings.ToLower(m[3])
			level := "error"
			if strings.Contains(lvlStr, "warn") || strings.Contains(lvlStr, "wrn") {
				level = "warning"
			}
			msg := strings.TrimSpace(m[5])
			if code := strings.TrimSpace(m[4]); code != "" {
				msg = fmt.Sprintf("[%s] %s", code, msg)
			}
			errs = append(errs, CompileError{
				File:    file,
				Line:    lNum,
				Column:  1,
				Level:   level,
				Message: msg,
			})
			continue
		}

		// 2. Watcom column format: file.for(line,col): Error! ...
		if m := watcomErrRegex2.FindStringSubmatch(line); len(m) >= 7 {
			file := resolvePath(m[1])
			lNum, _ := strconv.Atoi(m[2])
			cNum, _ := strconv.Atoi(m[3])
			lvlStr := strings.ToLower(m[4])
			level := "error"
			if strings.Contains(lvlStr, "warn") || strings.Contains(lvlStr, "wrn") {
				level = "warning"
			}
			msg := strings.TrimSpace(m[6])
			if code := strings.TrimSpace(m[5]); code != "" {
				msg = fmt.Sprintf("[%s] %s", code, msg)
			}
			errs = append(errs, CompileError{
				File:    file,
				Line:    lNum,
				Column:  cNum,
				Level:   level,
				Message: msg,
			})
			continue
		}

		// 3. Standard Unix / gfortran format: file.f:line:col: error: message
		if m := stdErrRegex.FindStringSubmatch(line); len(m) >= 6 {
			file := resolvePath(m[1])
			lNum, _ := strconv.Atoi(m[2])
			cNum := 1
			if m[3] != "" {
				cNum, _ = strconv.Atoi(m[3])
			}
			level := strings.ToLower(m[4])
			msg := strings.TrimSpace(m[5])
			errs = append(errs, CompileError{
				File:    file,
				Line:    lNum,
				Column:  cNum,
				Level:   level,
				Message: msg,
			})
			continue
		}
	}

	return errs
}

// Build compiles the target Fortran file using Open Watcom 2.0 or compatible compiler
func Build(targetPath string) *BuildResult {
	start := time.Now()
	res := &BuildResult{}

	lines := CountLines(targetPath)
	res.LinesCompiled = lines

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	dir := filepath.Dir(absTarget)
	if dir == "" {
		dir = "."
	}

	compilerInfo, found := FindFortranCompiler()
	res.CompilerFound = found
	if !found {
		res.Success = false
		res.CompilerName = "Open Watcom 2.0 F77 (Not Found)"
		res.Duration = time.Since(start)
		res.ErrorCount = 1
		res.RawOutput = "Error: Open Watcom 2.0 FORTRAN 77 compiler (wfl386) was not found.\n" +
			"Please install Open Watcom 2.0 or set the WATCOM environment variable.\n" +
			"Example: export WATCOM=/opt/watcom\n" +
			"         export PATH=$WATCOM/binl64:$PATH"
		res.Errors = append(res.Errors, CompileError{
			File:    filepath.Base(targetPath),
			Line:    1,
			Column:  1,
			Level:   "error",
			Message: "Open Watcom 2.0 (wfl386) not found. Check WATCOM environment variable or PATH.",
		})
		return res
	}

	res.CompilerName = filepath.Base(compilerInfo.Path)

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	tmpBin := filepath.Join(os.TempDir(), fmt.Sprintf("tf77_bin_%d%s", time.Now().UnixNano(), ext))
	res.BinaryPath = tmpBin

	var cmd *exec.Cmd
	switch compilerInfo.Kind {
	case "wfl386", "wfl":
		// Open Watcom 32-bit/16-bit compile & link utility
		// /d1 : line number debugging info
		// /fe= : output executable name
		// /quiet : suppress logo banner if desired
		feArg := fmt.Sprintf("/fe=%s", tmpBin)
		cmd = exec.Command(compilerInfo.Path, "/d1", feArg, filepath.Base(absTarget))
		cmd.Dir = dir
	case "wfc386":
		// Compile only, then link
		objFile := strings.TrimSuffix(filepath.Base(absTarget), filepath.Ext(absTarget)) + ".obj"
		cmd = exec.Command(compilerInfo.Path, "/d1", filepath.Base(absTarget))
		cmd.Dir = dir
		_ = objFile
	case "gfortran", "f77":
		// Fallback GNU Fortran compiler
		cmd = exec.Command(compilerInfo.Path, "-g", "-o", tmpBin, filepath.Base(absTarget))
		cmd.Dir = dir
	default:
		cmd = exec.Command(compilerInfo.Path, filepath.Base(absTarget))
		cmd.Dir = dir
	}

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	runErr := cmd.Run()
	res.Duration = time.Since(start)
	res.RawOutput = outBuf.String()

	if runErr != nil {
		res.Success = false
		res.Errors = ParseErrors(res.RawOutput, dir)
		for _, e := range res.Errors {
			if e.Level == "error" {
				res.ErrorCount++
			} else {
				res.WarningCount++
			}
		}
		if res.ErrorCount == 0 && len(res.Errors) > 0 {
			res.ErrorCount = len(res.Errors)
		} else if res.ErrorCount == 0 {
			res.ErrorCount = 1
			displayMsg := strings.TrimSpace(res.RawOutput)
			if displayMsg == "" {
				displayMsg = fmt.Sprintf("Build failed with code %v", runErr)
			}
			res.Errors = append(res.Errors, CompileError{
				File:    filepath.Base(targetPath),
				Line:    1,
				Column:  1,
				Level:   "error",
				Message: displayMsg,
			})
		}
	} else {
		res.Success = true
		warnings := ParseErrors(res.RawOutput, dir)
		for _, w := range warnings {
			if w.Level == "warning" {
				res.WarningCount++
			}
		}
	}

	return res
}

// Run executes the compiled binary and returns output
// RunBinary executes the binary in workDir with args
func RunBinary(binPath string, workDir string, args []string) *RunResult {
	start := time.Now()
	res := &RunResult{}

	cmd := exec.Command(binPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.Output = outBuf.String()
	res.Completed = true

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
	} else {
		res.ExitCode = 0
	}

	return res
}

func Run(binPath string, args ...string) *RunResult {
	return RunBinary(binPath, "", args)
}

