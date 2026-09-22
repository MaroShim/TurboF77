package compiler

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	SourceFiles   []string // List of source files compiled
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

// DebuggerInfo contains path and kind of detected debugger tool
type DebuggerInfo struct {
	Path string
	Kind string // "lldb", "gdb", "wd"
}

// FindDebuggerTool locates the appropriate debugger for the specified compiler
func FindDebuggerTool(compilerKind string) (DebuggerInfo, bool) {
	// For GNU / LLVM compilers (gfortran, flang, ifx, f77):
	if compilerKind == "gfortran" || compilerKind == "flang" || compilerKind == "flang-new" ||
		compilerKind == "ifx" || compilerKind == "ifort" || compilerKind == "f77" || compilerKind == "" {
		if runtime.GOOS == "darwin" {
			// On macOS, prefer native Apple LLDB
			lldbCandidates := []string{"/usr/bin/lldb", "lldb"}
			for _, c := range lldbCandidates {
				if path, err := exec.LookPath(c); err == nil {
					return DebuggerInfo{Path: path, Kind: "lldb"}, true
				}
			}
			if path, err := exec.LookPath("gdb"); err == nil {
				return DebuggerInfo{Path: path, Kind: "gdb"}, true
			}
		} else {
			// On Linux/Windows, prefer GDB
			if path, err := exec.LookPath("gdb"); err == nil {
				return DebuggerInfo{Path: path, Kind: "gdb"}, true
			}
			if path, err := exec.LookPath("lldb"); err == nil {
				return DebuggerInfo{Path: path, Kind: "lldb"}, true
			}
		}
	}

	// For Open Watcom compilers:
	if strings.Contains(compilerKind, "wf") || compilerKind == "owcc" {
		if watcomDir := os.Getenv("WATCOM"); watcomDir != "" {
			candidates := []string{
				filepath.Join(watcomDir, "binnt64", "wd.exe"),
				filepath.Join(watcomDir, "binnt", "wd.exe"),
				filepath.Join(watcomDir, "binl64", "wd"),
				filepath.Join(watcomDir, "binl", "wd"),
				filepath.Join(watcomDir, "binarm64", "wd"),
				filepath.Join(watcomDir, "binw", "wd.exe"),
			}
			for _, c := range candidates {
				if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
					return DebuggerInfo{Path: c, Kind: "wd"}, true
				}
			}
		}
		if path, err := exec.LookPath("wd"); err == nil {
			return DebuggerInfo{Path: path, Kind: "wd"}, true
		}
	}

	return DebuggerInfo{}, false
}

// FindFortranCompiler locates Open Watcom 2.0 FORTRAN 77 or compatible compiler
func FindFortranCompiler() (CompilerInfo, bool) {
	// 1. If user explicitly set WATCOM environment variable, honor it first
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

	// 2. Practical Priority #1: Modern standard compilers (gfortran, flang, ifx)
	modernBins := []string{"gfortran", "gfortran.exe", "flang-new", "flang", "ifx", "ifort"}
	for _, bin := range modernBins {
		if path, err := exec.LookPath(bin); err == nil {
			kind := bin
			if strings.HasPrefix(bin, "gfortran") {
				kind = "gfortran"
			}
			return CompilerInfo{Path: path, Kind: kind}, true
		}
	}

	// 3. Check system PATH for Open Watcom tools
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

	// 4. Check common standard Watcom installation locations
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

	// 5. Classic Unix f77
	if path, err := exec.LookPath("f77"); err == nil {
		return CompilerInfo{Path: path, Kind: "f77"}, true
	}

	return CompilerInfo{}, false
}

// IsFreeFormFortran checks if a filename represents Free-Form Fortran (F90, F95, F2003, F2008, F2018)
func IsFreeFormFortran(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".f90" || ext == ".f95" || ext == ".f03" || ext == ".f08" || ext == ".f18"
}

// IsFixedFormFortran checks if a filename represents Fixed-Form Fortran (F77 and earlier)
func IsFixedFormFortran(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".for" || ext == ".f" || ext == ".f77" || ext == ".for77" || ext == ".ftn" || ext == ".inc"
}

// IsFortranSource checks if a filename is any recognized Fortran source file
func IsFortranSource(path string) bool {
	return IsFixedFormFortran(path) || IsFreeFormFortran(path)
}

// HasProgramStatement checks if a Fortran source file contains a PROGRAM statement
func HasProgramStatement(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	isFree := IsFreeFormFortran(path)
	lines := strings.Split(string(content), "\n")
	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		if len(line) == 0 {
			continue
		}
		if !isFree {
			// In F77, comment in column 1 (0-indexed)
			first := line[0]
			if first == 'C' || first == 'c' || first == '*' || first == '!' {
				continue
			}
		}
		// Strip inline comment
		if idx := strings.Index(line, "!"); idx >= 0 {
			line = line[:idx]
		}
		trimmed := strings.TrimSpace(line)
		// Strip optional statement label
		for len(trimmed) > 0 && unicode.IsDigit(rune(trimmed[0])) {
			trimmed = strings.TrimSpace(trimmed[1:])
		}
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "PROGRAM") {
			rest := upper[len("PROGRAM"):]
			if len(rest) == 0 || rest[0] == ' ' || rest[0] == '\t' {
				return true
			}
		}
	}
	return false
}

// FindCompanionFiles scans the directory of targetPath for other Fortran source files
// that do NOT contain a PROGRAM statement (i.e. subroutines, functions, block data).
func FindCompanionFiles(targetPath string) []string {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	baseTarget := filepath.Base(absTarget)

	// Avoid scanning companions if the file is an unsaved temporary buffer in temp directory
	if strings.HasPrefix(baseTarget, "tf77_temp_") || strings.HasPrefix(baseTarget, "tf_temp_") ||
		strings.HasPrefix(baseTarget, "tf77_dbg_") || strings.HasPrefix(baseTarget, "tf_dbg_") {
		return nil
	}

	dir := filepath.Dir(absTarget)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var companions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == baseTarget {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		// Only compile source files, exclude .inc headers
		if !IsFortranSource(name) || ext == ".inc" {
			continue
		}
		fullPath := filepath.Join(dir, name)
		if !HasProgramStatement(fullPath) {
			companions = append(companions, fullPath)
		}
	}
	return companions
}

// FindMainProgramFile finds a file in dir that defines a PROGRAM statement, excluding excludePath
func FindMainProgramFile(dir string, excludePath string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	excludeBase := filepath.Base(excludePath)
	var mainFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == excludeBase {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if !IsFortranSource(name) || ext == ".inc" {
			continue
		}
		fullPath := filepath.Join(dir, name)
		if HasProgramStatement(fullPath) {
			mainFiles = append(mainFiles, fullPath)
		}
	}
	if len(mainFiles) == 1 {
		return mainFiles[0]
	}
	return ""
}

// CountFileLines counts non-empty lines in a single file
func CountFileLines(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	trimmed := bytes.TrimRight(content, "\r\n")
	if len(trimmed) == 0 {
		return 0
	}
	return bytes.Count(trimmed, []byte("\n")) + 1
}

// CountTotalLines counts total lines of code across all given files
func CountTotalLines(files []string) int {
	total := 0
	for _, f := range files {
		total += CountFileLines(f)
	}
	return total
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
				total += CountFileLines(path)
			}
			return nil
		})
		return total
	}

	return CountFileLines(targetPath)
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

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	dir := filepath.Dir(absTarget)
	if dir == "" {
		dir = "."
	}

	// Smart Multi-file source resolution:
	// If targetPath has PROGRAM, target is main and all companion subroutine files are included.
	// If targetPath is a subroutine file, link it with main program file in same directory.
	var allSources []string
	if HasProgramStatement(absTarget) {
		allSources = append(allSources, absTarget)
		companions := FindCompanionFiles(absTarget)
		allSources = append(allSources, companions...)
	} else {
		mainProg := FindMainProgramFile(dir, absTarget)
		if mainProg != "" {
			allSources = append(allSources, mainProg, absTarget)
			companions := FindCompanionFiles(mainProg)
			for _, c := range companions {
				if c != absTarget {
					allSources = append(allSources, c)
				}
			}
		} else {
			allSources = append(allSources, absTarget)
			companions := FindCompanionFiles(absTarget)
			allSources = append(allSources, companions...)
		}
	}

	res.SourceFiles = allSources
	res.LinesCompiled = CountTotalLines(allSources)

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

	// Deterministic binary name per target to enable incremental build (Rule 72)
	hash := sha256.Sum256([]byte(absTarget))
	binName := fmt.Sprintf("tf77_bin_%x%s", hash[:8], ext)
	tmpBin := filepath.Join(os.TempDir(), binName)
	res.BinaryPath = tmpBin

	// Check if existing binary is newer than all source files
	if binFi, err := os.Stat(tmpBin); err == nil && !binFi.IsDir() {
		binMtime := binFi.ModTime()
		upToDate := true
		for _, sf := range allSources {
			if sFi, err := os.Stat(sf); err != nil || sFi.ModTime().After(binMtime) {
				upToDate = false
				break
			}
		}
		if upToDate {
			res.Success = true
			res.Duration = time.Since(start)
			return res
		}
	}

	// Relative filenames for compiler arguments (since cmd.Dir = dir)
	var sourceArgs []string
	for _, f := range allSources {
		sourceArgs = append(sourceArgs, filepath.Base(f))
	}

	var cmd *exec.Cmd
	switch compilerInfo.Kind {
	case "wfl386", "wfl":
		// Open Watcom 32-bit/16-bit compile & link utility
		// /d1 : line number debugging info
		// /fe= : output executable name
		feArg := fmt.Sprintf("/fe=%s", tmpBin)
		args := append([]string{"/d1", feArg}, sourceArgs...)
		cmd = exec.Command(compilerInfo.Path, args...)
		cmd.Dir = dir
	case "wfc386":
		// Compile only, then link
		args := append([]string{"/d1"}, sourceArgs...)
		cmd = exec.Command(compilerInfo.Path, args...)
		cmd.Dir = dir
	case "gfortran", "f77", "flang", "flang-new", "ifx", "ifort":
		// GNU Fortran & Modern Fortran compilers
		// Isolate module (.mod) artifacts in temporary directory to prevent pollution (Rules 73, 76)
		tmpModDir := filepath.Join(os.TempDir(), fmt.Sprintf("tf_mods_%d", time.Now().UnixNano()))
		_ = os.MkdirAll(tmpModDir, 0755)
		defer os.RemoveAll(tmpModDir)

		args := append([]string{"-g", "-J", tmpModDir, "-I", tmpModDir, "-o", tmpBin}, sourceArgs...)
		cmd = exec.Command(compilerInfo.Path, args...)
		cmd.Dir = dir
	default:
		cmd = exec.Command(compilerInfo.Path, sourceArgs...)
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

