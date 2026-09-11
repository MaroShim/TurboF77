package syntax

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHighlightLine(t *testing.T) {
	baseStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorNavy)

	// 1. Fixed comment line starting with 'C'
	line1 := "C     THIS IS A FORTRAN COMMENT"
	tokens1 := HighlightLine(line1, baseStyle, nil, false)
	if len(tokens1) != len(line1) {
		t.Fatalf("expected %d tokens, got %d", len(line1), len(tokens1))
	}
	for i, tok := range tokens1 {
		fg, _, _ := tok.Style.Decompose()
		if fg != ColorSyntaxComment {
			t.Errorf("token %d expected comment color, got %v", i, fg)
		}
	}

	// 2. Fixed statement with label 100 and PRINT
	line2 := "  100 PRINT *, 'HELLO, FORTRAN 77'"
	tokens2 := HighlightLine(line2, baseStyle, nil, false)
	if len(tokens2) != len(line2) {
		t.Fatalf("expected %d tokens, got %d", len(line2), len(tokens2))
	}
	// '100' should be styled as label
	fg, _, _ := tokens2[2].Style.Decompose()
	if fg != ColorSyntaxLabel {
		t.Errorf("token 2 ('1') expected label color, got %v", fg)
	}

	// 3. Logical operator .AND.
	line3 := "      IF (A .GT. 0.0 .AND. B .LT. 10.0) THEN"
	tokens3 := HighlightLine(line3, baseStyle, nil, false)
	if len(tokens3) != len(line3) {
		t.Fatalf("expected %d tokens, got %d", len(line3), len(tokens3))
	}

	// 4. Free-Form Modern Fortran (MODULE, USE, ! inline comment)
	line4 := "module math_utils ! modern module definition"
	tokens4 := HighlightLine(line4, baseStyle, nil, true)
	if len(tokens4) != len(line4) {
		t.Fatalf("expected %d tokens, got %d", len(line4), len(tokens4))
	}
	// 'module' keyword
	fgMod, _, _ := tokens4[0].Style.Decompose()
	if fgMod != ColorSyntaxKeyword {
		t.Errorf("expected module to have keyword color, got %v", fgMod)
	}
	// '!' starts comment
	bangIdx := 18
	fgBang, _, _ := tokens4[bangIdx].Style.Decompose()
	if fgBang != ColorSyntaxComment {
		t.Errorf("expected '!' to have comment color, got %v", fgBang)
	}

	// 5. Free-Form Modern Fortran operators (==, ::, ALLOCATABLE)
	line5 := "integer, allocatable :: arr(:)"
	tokens5 := HighlightLine(line5, baseStyle, nil, true)
	fgAlloc, _, _ := tokens5[9].Style.Decompose() // 'allocatable'
	if fgAlloc != ColorSyntaxKeyword {
		t.Errorf("expected allocatable to have keyword color, got %v", fgAlloc)
	}
	fgColon, _, _ := tokens5[21].Style.Decompose() // '::'
	if fgColon != ColorSyntaxOperator {
		t.Errorf("expected '::' to have operator color, got %v", fgColon)
	}
}
