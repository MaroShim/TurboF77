package syntax

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

var (
	ColorSyntaxKeyword  = tcell.ColorYellow
	ColorSyntaxType     = tcell.ColorLightCyan
	ColorSyntaxString   = tcell.ColorLightCyan
	ColorSyntaxNumber   = tcell.ColorLightGreen
	ColorSyntaxComment  = tcell.NewHexColor(0x808080) // Gray
	ColorSyntaxBuiltin  = tcell.ColorGreen
	ColorSyntaxOperator = tcell.NewHexColor(0xFF55FF) // Light Magenta
	ColorSyntaxLabel    = tcell.ColorDarkCyan
	ColorSyntaxNormal   = tcell.ColorWhite
)

var fortranKeywords = map[string]bool{
	// Fortran 77
	"program": true, "subroutine": true, "function": true, "block": true, "data": true,
	"entry": true, "common": true, "dimension": true, "equivalence": true, "parameter": true,
	"implicit": true, "save": true, "intrinsic": true, "external": true,
	"do": true, "continue": true, "if": true, "then": true, "else": true, "elseif": true, "endif": true,
	"goto": true, "go": true, "to": true, "assign": true,
	"call": true, "return": true, "stop": true, "pause": true, "end": true,
	"read": true, "write": true, "print": true, "open": true, "close": true,
	"inquire": true, "backspace": true, "endfile": true, "rewind": true, "format": true,

	// Modern Fortran (F90 / F95 / F2003 / F2008)
	"module": true, "use": true, "contains": true, "only": true,
	"select": true, "case": true, "default": true, "cycle": true, "exit": true,
	"where": true, "elsewhere": true, "endwhere": true, "forall": true, "endforall": true,
	"type": true, "endtype": true, "intent": true, "in": true, "out": true, "inout": true,
	"allocatable": true, "allocate": true, "deallocate": true,
	"pointer": true, "target": true, "nullify": true,
	"optional": true, "public": true, "private": true,
	"interface": true, "endinterface": true, "procedure": true, "generic": true,
	"operator": true, "assignment": true,
	"recursive": true, "pure": true, "elemental": true, "result": true,
	"namelist": true, "sequence": true, "abstract": true, "extends": true, "class": true,
	"asynchronous": true, "bind": true, "import": true, "associate": true,
	"critical": true, "error": true,
}

var fortranTypes = map[string]bool{
	"integer": true, "real": true, "double": true, "precision": true, "doubleprecision": true,
	"complex": true, "logical": true, "character": true,
	"kind": true, "len": true, "byte": true,
}

var fortranIntrinsics = map[string]bool{
	// F77 Math & General
	"abs": true, "acos": true, "aimag": true, "aint": true, "alog": true, "alog10": true,
	"amax0": true, "amax1": true, "amin0": true, "amin1": true, "amod": true, "anint": true,
	"asin": true, "atan": true, "atan2": true, "cabs": true, "ccos": true, "char": true,
	"clog": true, "cmplx": true, "conjg": true, "cos": true, "cosh": true, "csin": true,
	"csqrt": true, "dabs": true, "dacos": true, "dasin": true, "datan": true, "datan2": true,
	"dcos": true, "dcosh": true, "ddim": true, "dexp": true, "dim": true, "dint": true,
	"dlog": true, "dlog10": true, "dmax1": true, "dmin1": true, "dmod": true, "dnint": true,
	"dprod": true, "dsign": true, "dsin": true, "dsinh": true, "dsqrt": true, "dtan": true,
	"dtanh": true, "exp": true, "float": true, "iabs": true, "ichar": true, "idim": true,
	"idint": true, "idnint": true, "ifix": true, "index": true, "int": true, "isign": true,
	"len": true, "lge": true, "lgt": true, "lle": true, "llt": true, "max": true,
	"max0": true, "max1": true, "min": true, "min0": true, "min1": true, "mod": true,
	"nint": true, "sign": true, "sin": true, "sinh": true, "sngl": true, "sqrt": true,
	"tan": true, "tanh": true,

	// F90+ Array, String, System intrinsics
	"sum": true, "product": true, "matmul": true, "dot_product": true,
	"transpose": true, "reshape": true, "size": true, "shape": true,
	"lbound": true, "ubound": true, "pack": true, "unpack": true, "merge": true,
	"all": true, "any": true, "count": true, "maxval": true, "minval": true,
	"maxloc": true, "minloc": true, "cshift": true, "eoshift": true, "spread": true,
	"trim": true, "adjustl": true, "adjustr": true, "associated": true,
	"present": true, "allocated": true, "null": true,
	"iand": true, "ior": true, "ieor": true, "not": true, "ishft": true,
	"btest": true, "ibset": true, "ibclr": true, "scan": true, "verify": true, "repeat": true,
	"selected_real_kind": true, "selected_int_kind": true,
	"huge": true, "tiny": true, "epsilon": true,
	"cpu_time": true, "system_clock": true, "date_and_time": true,
	"random_number": true, "random_seed": true,
}

var fortranOperators = map[string]bool{
	".true.": true, ".false.": true,
	".and.": true, ".or.": true, ".not.": true, ".eqv.": true, ".neqv.": true,
	".eq.": true, ".ne.": true, ".lt.": true, ".le.": true, ".gt.": true, ".ge.": true,
}

// Token represents a styled span of runes on a line
type Token struct {
	Style tcell.Style
	Char  rune
}

// HighlightLine parses a line of Fortran code and returns styled tokens.
// Optional isFreeForm flag specifies whether to use Free-Form (F90+) or Fixed-Form (F77) syntax rules.
func HighlightLine(line string, baseStyle tcell.Style, inBlockComment *bool, isFreeForm ...bool) []Token {
	freeForm := len(isFreeForm) > 0 && isFreeForm[0]
	runes := []rune(line)
	n := len(runes)
	tokens := make([]Token, n)

	keywordStyle := baseStyle.Foreground(ColorSyntaxKeyword).Bold(true)
	typeStyle := baseStyle.Foreground(ColorSyntaxType).Bold(true)
	stringStyle := baseStyle.Foreground(ColorSyntaxString)
	numberStyle := baseStyle.Foreground(ColorSyntaxNumber)
	commentStyle := baseStyle.Foreground(ColorSyntaxComment)
	builtinStyle := baseStyle.Foreground(ColorSyntaxBuiltin)
	opStyle := baseStyle.Foreground(ColorSyntaxOperator).Bold(true)
	labelStyle := baseStyle.Foreground(ColorSyntaxLabel).Bold(true)
	contStyle := baseStyle.Foreground(tcell.ColorMaroon).Bold(true)

	if n == 0 {
		return tokens
	}

	i := 0

	if !freeForm {
		// 1. FORTRAN 77 Fixed-format comment: Column 1 is 'C', 'c', '*', or '!'
		firstChar := runes[0]
		if firstChar == 'C' || firstChar == 'c' || firstChar == '*' || firstChar == '!' {
			for idx := 0; idx < n; idx++ {
				tokens[idx] = Token{Style: commentStyle, Char: runes[idx]}
			}
			return tokens
		}

		// 2. Statement Label (Columns 1-5, 0-indexed 0 to 4)
		if i < 5 && i < n {
			for i < 5 && i < n {
				if unicode.IsDigit(runes[i]) {
					tokens[i] = Token{Style: labelStyle, Char: runes[i]}
					i++
				} else if runes[i] == '!' {
					for i < n {
						tokens[i] = Token{Style: commentStyle, Char: runes[i]}
						i++
					}
					return tokens
				} else {
					tokens[i] = Token{Style: baseStyle, Char: runes[i]}
					i++
				}
			}
		}

		// 3. Continuation character (Column 6, 0-indexed 5)
		if i == 5 && i < n {
			if runes[i] != ' ' && runes[i] != '0' && runes[i] != '\t' {
				tokens[i] = Token{Style: contStyle, Char: runes[i]}
			} else {
				tokens[i] = Token{Style: baseStyle, Char: runes[i]}
			}
			i++
		}
	}

	// Main statement scanning loop (Free-form from col 0, Fixed-form from col 6/7+)
	for i < n {
		r := runes[i]

		// In-line comment (!)
		if r == '!' {
			for i < n {
				tokens[i] = Token{Style: commentStyle, Char: runes[i]}
				i++
			}
			break
		}

		// String literal: '...' (with '' escape)
		if r == '\'' {
			tokens[i] = Token{Style: stringStyle, Char: r}
			i++
			for i < n {
				if runes[i] == '\'' {
					tokens[i] = Token{Style: stringStyle, Char: runes[i]}
					i++
					if i < n && runes[i] == '\'' {
						tokens[i] = Token{Style: stringStyle, Char: runes[i]}
						i++
						continue
					}
					break
				}
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				i++
			}
			continue
		}

		// Double quoted string
		if r == '"' {
			tokens[i] = Token{Style: stringStyle, Char: r}
			i++
			for i < n {
				if runes[i] == '"' {
					tokens[i] = Token{Style: stringStyle, Char: runes[i]}
					i++
					break
				}
				tokens[i] = Token{Style: stringStyle, Char: runes[i]}
				i++
			}
			continue
		}

		// Modern relational / pointer / type member operators (==, /=, <=, >=, =>, %, ::)
		if i+1 < n {
			twoChars := string(runes[i : i+2])
			if twoChars == "==" || twoChars == "/=" || twoChars == "<=" || twoChars == ">=" || twoChars == "=>" || twoChars == "::" {
				tokens[i] = Token{Style: opStyle, Char: runes[i]}
				tokens[i+1] = Token{Style: opStyle, Char: runes[i+1]}
				i += 2
				continue
			}
		}

		// Dotted operators and logical constants: .TRUE., .EQ., etc.
		if r == '.' && i+1 < n && unicode.IsLetter(runes[i+1]) {
			start := i
			i++
			for i < n && unicode.IsLetter(runes[i]) {
				i++
			}
			if i < n && runes[i] == '.' {
				i++
				dotWord := strings.ToLower(string(runes[start:i]))
				if fortranOperators[dotWord] {
					for j := start; j < i; j++ {
						tokens[j] = Token{Style: opStyle, Char: runes[j]}
					}
					continue
				}
			}
			// Not a recognized dotted operator, rewind
			i = start
		}

		// Numbers (integers, floating point, exponents with E/D)
		if unicode.IsDigit(r) || (r == '.' && i+1 < n && unicode.IsDigit(runes[i+1])) {
			start := i
			hasExp := false
			for i < n {
				ch := runes[i]
				if unicode.IsDigit(ch) || ch == '.' {
					i++
				} else if (ch == 'E' || ch == 'e' || ch == 'D' || ch == 'd') && !hasExp {
					hasExp = true
					i++
					if i < n && (runes[i] == '+' || runes[i] == '-') {
						i++
					}
				} else {
					break
				}
			}
			for j := start; j < i; j++ {
				tokens[j] = Token{Style: numberStyle, Char: runes[j]}
			}
			continue
		}

		// Identifiers, Keywords, Types, Intrinsics
		if unicode.IsLetter(r) || r == '_' || r == '$' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_' || runes[i] == '$') {
				i++
			}
			word := strings.ToLower(string(runes[start:i]))

			var style tcell.Style
			if fortranKeywords[word] {
				style = keywordStyle
			} else if fortranTypes[word] {
				style = typeStyle
			} else if fortranIntrinsics[word] {
				style = builtinStyle
			} else {
				style = baseStyle
			}

			for j := start; j < i; j++ {
				tokens[j] = Token{Style: style, Char: runes[j]}
			}
			continue
		}

		// Single character operators
		if r == '%' || r == '&' || r == '<' || r == '>' || r == '=' || r == '+' || r == '-' || r == '*' || r == '/' {
			tokens[i] = Token{Style: opStyle, Char: r}
			i++
			continue
		}

		// Default punctuation
		tokens[i] = Token{Style: baseStyle, Char: r}
		i++
	}

	return tokens
}
