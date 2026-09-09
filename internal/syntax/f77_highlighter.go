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

var f77Keywords = map[string]bool{
	"program": true, "subroutine": true, "function": true, "block": true, "data": true,
	"entry": true, "common": true, "dimension": true, "equivalence": true, "parameter": true,
	"implicit": true, "save": true, "intrinsic": true, "external": true,
	"do": true, "continue": true, "if": true, "then": true, "else": true, "elseif": true, "endif": true,
	"goto": true, "go": true, "to": true, "assign": true,
	"call": true, "return": true, "stop": true, "pause": true, "end": true,
	"read": true, "write": true, "print": true, "open": true, "close": true,
	"inquire": true, "backspace": true, "endfile": true, "rewind": true, "format": true,
}

var f77Types = map[string]bool{
	"integer": true, "real": true, "double": true, "precision": true,
	"complex": true, "logical": true, "character": true,
}

var f77Intrinsics = map[string]bool{
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
}

var f77Operators = map[string]bool{
	".true.": true, ".false.": true,
	".and.": true, ".or.": true, ".not.": true, ".eqv.": true, ".neqv.": true,
	".eq.": true, ".ne.": true, ".lt.": true, ".le.": true, ".gt.": true, ".ge.": true,
}

// Token represents a styled span of runes on a line
type Token struct {
	Style tcell.Style
	Char  rune
}

// HighlightLine parses a line of FORTRAN 77 code and returns styled tokens
func HighlightLine(line string, baseStyle tcell.Style, inBlockComment *bool) []Token {
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

	// 1. FORTRAN 77 Fixed-format comment: Column 1 is 'C', 'c', '*', or '!'
	firstChar := runes[0]
	if firstChar == 'C' || firstChar == 'c' || firstChar == '*' || firstChar == '!' {
		for i := 0; i < n; i++ {
			tokens[i] = Token{Style: commentStyle, Char: runes[i]}
		}
		return tokens
	}

	i := 0

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

	// 4. Main statement and columns >= 7
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

		// Double quoted string (common compiler extension)
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
				if f77Operators[dotWord] {
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
			if f77Keywords[word] {
				style = keywordStyle
			} else if f77Types[word] {
				style = typeStyle
			} else if f77Intrinsics[word] {
				style = builtinStyle
			} else {
				style = baseStyle
			}

			for j := start; j < i; j++ {
				tokens[j] = Token{Style: style, Char: runes[j]}
			}
			continue
		}

		// Default punctuation / operators
		tokens[i] = Token{Style: baseStyle, Char: r}
		i++
	}

	return tokens
}
