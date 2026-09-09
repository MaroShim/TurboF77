package debugger

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// F77LoopFrame tracks active DO loops in Fortran
type F77LoopFrame struct {
	VarName   string
	CurVal    int64
	EndVal    int64
	Step      int64
	Label     string
	StartLine int // 1-based line of loop body start
}

// F77Engine executes and steps through FORTRAN 77 code interactively
type F77Engine struct {
	Lines       []string
	Breakpoints map[int]bool
	CurLine     int // 1-based current line
	Variables   map[string]Variable
	VarOrder    []string
	Loops       []*F77LoopFrame
	OutputBuf   strings.Builder
	Active      bool
	Exited      bool
	ExitCode    int
	ProgName    string
}

func NewF77Engine(lines []string, breakpoints map[int]bool) *F77Engine {
	eng := &F77Engine{
		Lines:       lines,
		Breakpoints: make(map[int]bool),
		CurLine:     1,
		Variables:   make(map[string]Variable),
		Active:      true,
		Exited:      false,
		ExitCode:    0,
		ProgName:    "MAIN",
	}

	for l, set := range breakpoints {
		if set {
			eng.Breakpoints[l] = true
		}
	}

	// Fast-forward to the first executable line
	eng.advanceToExecutable()
	return eng
}

func (e *F77Engine) isExecutableLine(lineIdx int) bool {
	if lineIdx < 0 || lineIdx >= len(e.Lines) {
		return false
	}
	raw := e.Lines[lineIdx]
	if len(raw) == 0 {
		return false
	}
	// Column 1 comment
	c1 := raw[0]
	if c1 == 'C' || c1 == 'c' || c1 == '*' || c1 == '!' {
		return false
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.HasPrefix(trimmed, "!") {
		return false
	}

	// Strip statement label (if in columns 1-5)
	code := e.getStatementPart(raw)
	upper := strings.ToUpper(strings.TrimSpace(code))

	// Declarations that don't need step stops
	if strings.HasPrefix(upper, "PROGRAM") ||
		strings.HasPrefix(upper, "INTEGER") ||
		strings.HasPrefix(upper, "REAL") ||
		strings.HasPrefix(upper, "DOUBLE") ||
		strings.HasPrefix(upper, "COMPLEX") ||
		strings.HasPrefix(upper, "LOGICAL") ||
		strings.HasPrefix(upper, "CHARACTER") ||
		strings.HasPrefix(upper, "DIMENSION") ||
		strings.HasPrefix(upper, "COMMON") ||
		strings.HasPrefix(upper, "PARAMETER") ||
		strings.HasPrefix(upper, "DATA") ||
		strings.HasPrefix(upper, "IMPLICIT") {
		return false
	}

	return true
}

func (e *F77Engine) getStatementPart(line string) string {
	runes := []rune(line)
	if len(runes) <= 6 {
		return ""
	}
	return string(runes[6:])
}

func (e *F77Engine) advanceToExecutable() {
	for e.CurLine <= len(e.Lines) && !e.isExecutableLine(e.CurLine-1) {
		// Process declarations on non-executable lines (e.g. DATA, PARAMETER)
		e.processDeclarationLine(e.Lines[e.CurLine-1])
		e.CurLine++
	}
	if e.CurLine > len(e.Lines) {
		e.Exited = true
		e.Active = false
	}
}

func (e *F77Engine) processDeclarationLine(line string) {
	upper := strings.ToUpper(strings.TrimSpace(line))
	if strings.HasPrefix(upper, "PROGRAM ") {
		parts := strings.Fields(upper)
		if len(parts) >= 2 {
			e.ProgName = parts[1]
		}
	}
	// PARAMETER (NAME = VALUE)
	if strings.Contains(upper, "PARAMETER") {
		re := regexp.MustCompile(`(?i)PARAMETER\s*\(([^)]+)\)`)
		if m := re.FindStringSubmatch(line); len(m) >= 2 {
			pairs := strings.Split(m[1], ",")
			for _, p := range pairs {
				eq := strings.Split(p, "=")
				if len(eq) == 2 {
					name := strings.ToUpper(strings.TrimSpace(eq[0]))
					val := strings.TrimSpace(eq[1])
					e.setVariable(name, "INTEGER", val)
				}
			}
		}
	}
	// Explicit types declaration
	types := []string{"INTEGER", "REAL", "DOUBLE PRECISION", "LOGICAL", "CHARACTER"}
	for _, tp := range types {
		if strings.HasPrefix(upper, tp+" ") || strings.HasPrefix(upper, tp+"*") {
			after := strings.TrimSpace(upper[len(tp):])
			after = strings.TrimPrefix(after, "*")
			if idx := strings.Index(after, " "); idx != -1 && unicode.IsDigit(rune(after[0])) {
				after = strings.TrimSpace(after[idx:])
			}
			vars := strings.Split(after, ",")
			for _, v := range vars {
				v = strings.TrimSpace(v)
				v = strings.Split(v, "(")[0] // remove array dimension
				if v != "" && !e.hasVariable(v) {
					defaultVal := "0"
					if tp == "REAL" || strings.Contains(tp, "DOUBLE") {
						defaultVal = "0.0"
					} else if tp == "CHARACTER" {
						defaultVal = "''"
					}
					e.setVariable(v, tp, defaultVal)
				}
			}
		}
	}
}

func (e *F77Engine) hasVariable(name string) bool {
	_, ok := e.Variables[strings.ToUpper(name)]
	return ok
}

func (e *F77Engine) setVariable(name, varType, value string) {
	name = strings.ToUpper(name)
	if _, exists := e.Variables[name]; !exists {
		e.VarOrder = append(e.VarOrder, name)
	}
	if varType == "" {
		if existing, ok := e.Variables[name]; ok {
			varType = existing.Type
		} else {
			// Fortran implicit typing: I-N is INTEGER, others are REAL
			first := name[0]
			if first >= 'I' && first <= 'N' {
				varType = "INTEGER"
			} else {
				varType = "REAL"
			}
		}
	}
	e.Variables[name] = Variable{Name: name, Type: varType, Value: value}
}

// Step advances execution by one line
func (e *F77Engine) Step() bool {
	if e.Exited || e.CurLine > len(e.Lines) {
		e.Exited = true
		e.Active = false
		return false
	}

	curIdx := e.CurLine - 1
	line := e.Lines[curIdx]

	// Execute statement on current line
	stmt := strings.TrimSpace(e.getStatementPart(line))
	if stmt == "" {
		stmt = strings.TrimSpace(line)
	}
	upperStmt := strings.ToUpper(stmt)

	// Check if this line is STOP or END
	if upperStmt == "STOP" || strings.HasPrefix(upperStmt, "STOP ") || upperStmt == "END" {
		e.Exited = true
		e.Active = false
		return false
	}

	// 1. DO label var = start, end [, step]
	doRe := regexp.MustCompile(`(?i)^DO\s+(\d+)\s+([A-Za-z]\w*)\s*=\s*([^,]+),\s*([^,]+)(?:,\s*([^\s]+))?`)
	if m := doRe.FindStringSubmatch(stmt); len(m) >= 5 {
		lbl := m[1]
		varName := strings.ToUpper(m[2])
		startVal := e.evalInt(m[3])
		endVal := e.evalInt(m[4])
		stepVal := int64(1)
		if len(m) > 5 && m[5] != "" {
			stepVal = e.evalInt(m[5])
		}

		e.setVariable(varName, "INTEGER", strconv.FormatInt(startVal, 10))
		e.Loops = append(e.Loops, &F77LoopFrame{
			VarName:   varName,
			CurVal:    startVal,
			EndVal:    endVal,
			Step:      stepVal,
			Label:     lbl,
			StartLine: e.CurLine + 1,
		})
		e.CurLine++
		e.advanceToExecutable()
		return true
	}

	// 2. Loop end check: label CONTINUE or matching label
	if len(e.Loops) > 0 {
		topLoop := e.Loops[len(e.Loops)-1]
		// Check if current line has the loop label in columns 1-5
		lineLabel := strings.TrimSpace(line[:min(5, len(line))])
		if lineLabel == topLoop.Label {
			// Increment loop variable
			topLoop.CurVal += topLoop.Step
			e.setVariable(topLoop.VarName, "INTEGER", strconv.FormatInt(topLoop.CurVal, 10))

			if (topLoop.Step > 0 && topLoop.CurVal <= topLoop.EndVal) ||
				(topLoop.Step < 0 && topLoop.CurVal >= topLoop.EndVal) {
				// Loop back to start
				e.CurLine = topLoop.StartLine
				e.advanceToExecutable()
				return true
			} else {
				// Loop finished, pop frame
				e.Loops = e.Loops[:len(e.Loops)-1]
			}
		}
	}

	// 3. PRINT *, ... or WRITE(*, *) ...
	if strings.HasPrefix(upperStmt, "PRINT *") || strings.HasPrefix(upperStmt, "WRITE(") {
		idx := strings.Index(stmt, ",")
		if idx != -1 {
			args := stmt[idx+1:]
			out := e.formatPrint(args)
			e.OutputBuf.WriteString(out + "\n")
		} else {
			e.OutputBuf.WriteString("\n")
		}
	}

	// 4. Assignment: VAR = EXPR
	if eqIdx := strings.Index(stmt, "="); eqIdx != -1 && !strings.HasPrefix(upperStmt, "IF") {
		lhs := strings.TrimSpace(stmt[:eqIdx])
		rhs := strings.TrimSpace(stmt[eqIdx+1:])
		if isSimpleIdentifier(lhs) {
			val := e.evalExpression(rhs)
			e.setVariable(lhs, "", val)
		}
	}

	e.CurLine++
	e.advanceToExecutable()
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isSimpleIdentifier(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func (e *F77Engine) evalInt(expr string) int64 {
	expr = strings.TrimSpace(expr)
	if v, ok := e.Variables[strings.ToUpper(expr)]; ok {
		if n, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
			return n
		}
	}
	if n, err := strconv.ParseInt(expr, 10, 64); err == nil {
		return n
	}
	return 1
}

func (e *F77Engine) evalExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// String literal
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") && len(expr) >= 2 {
		return expr[1 : len(expr)-1]
	}

	// Simple variable lookup
	if v, ok := e.Variables[strings.ToUpper(expr)]; ok {
		return v.Value
	}

	// Number literal
	if _, err := strconv.ParseFloat(expr, 64); err == nil {
		return expr
	}

	// Simple Binary Math: A + B, A - B, A * B, A / B
	for _, op := range []string{"+", "-", "*", "/"} {
		parts := strings.Split(expr, op)
		if len(parts) == 2 {
			v1 := e.evalNumeric(parts[0])
			v2 := e.evalNumeric(parts[1])
			var res float64
			switch op {
			case "+":
				res = v1 + v2
			case "-":
				res = v1 - v2
			case "*":
				res = v1 * v2
			case "/":
				if v2 != 0 {
					res = v1 / v2
				}
			}
			// If both were integers, return integer
			if math.Trunc(res) == res {
				return strconv.FormatInt(int64(res), 10)
			}
			return fmt.Sprintf("%.4f", res)
		}
	}

	return expr
}

func (e *F77Engine) evalNumeric(expr string) float64 {
	expr = strings.TrimSpace(expr)
	if v, ok := e.Variables[strings.ToUpper(expr)]; ok {
		if f, err := strconv.ParseFloat(v.Value, 64); err == nil {
			return f
		}
	}
	if f, err := strconv.ParseFloat(expr, 64); err == nil {
		return f
	}
	return 0.0
}

func (e *F77Engine) formatPrint(args string) string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for _, r := range args {
		if r == '\'' {
			inQuote = !inQuote
			current.WriteRune(r)
		} else if r == ',' && !inQuote {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	var out strings.Builder
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'") && len(p) >= 2 {
			out.WriteString(p[1 : len(p)-1])
		} else {
			// Variable or number
			val := e.evalExpression(p)
			out.WriteString(val)
		}
		if i < len(parts)-1 {
			out.WriteString(" ")
		}
	}
	return out.String()
}

// GetLocalVariables returns list of all tracked variables in declaration order
func (e *F77Engine) GetLocalVariables() []Variable {
	var list []Variable
	for _, name := range e.VarOrder {
		if v, ok := e.Variables[name]; ok {
			list = append(list, v)
		}
	}
	return list
}
