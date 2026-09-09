package debugger

import (
	"testing"
)

func TestF77Engine(t *testing.T) {
	code := []string{
		"C     Test Fortran Program",
		"      PROGRAM TEST",
		"      INTEGER I, SUM",
		"      SUM = 0",
		"      DO 10 I = 1, 5",
		"          SUM = SUM + I",
		"   10 CONTINUE",
		"      PRINT *, 'SUM IS:', SUM",
		"      STOP",
		"      END",
	}

	eng := NewF77Engine(code, map[int]bool{6: true}) // Breakpoint on line 6 (SUM = SUM + I)
	if eng.CurLine != 4 { // Line 4 is 'SUM = 0'
		t.Fatalf("expected CurLine 4, got %d", eng.CurLine)
	}

	// Step line 4 (SUM = 0)
	eng.Step()
	if v := eng.Variables["SUM"].Value; v != "0" {
		t.Errorf("expected SUM=0, got %s", v)
	}

	// Step line 5 (DO 10 I = 1, 5)
	eng.Step()
	if v := eng.Variables["I"].Value; v != "1" {
		t.Errorf("expected I=1, got %s", v)
	}

	// Step line 6 (SUM = SUM + I -> SUM = 1)
	eng.Step()
	if v := eng.Variables["SUM"].Value; v != "1" {
		t.Errorf("expected SUM=1, got %s", v)
	}
}
