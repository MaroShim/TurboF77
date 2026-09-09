package syntax

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHighlightLine(t *testing.T) {
	baseStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorNavy)

	// 1. Fixed comment line starting with 'C'
	line1 := "C     THIS IS A FORTRAN COMMENT"
	tokens1 := HighlightLine(line1, baseStyle, nil)
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
	tokens2 := HighlightLine(line2, baseStyle, nil)
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
	tokens3 := HighlightLine(line3, baseStyle, nil)
	if len(tokens3) != len(line3) {
		t.Fatalf("expected %d tokens, got %d", len(line3), len(tokens3))
	}
}
