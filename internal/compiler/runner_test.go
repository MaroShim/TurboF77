package compiler

import (
	"testing"
)

func TestParseErrors_WatcomStandard(t *testing.T) {
	output := `Open Watcom F77 Optimizing Compiler Version 2.0
hello.for(12): Error! E1024: Symbol 'X' has not been declared
hello.for(15): Warning! W2011: Variable 'Y' defined but not used
`
	errs := ParseErrors(output, "/tmp")
	if len(errs) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(errs))
	}

	if errs[0].Line != 12 || errs[0].Level != "error" {
		t.Errorf("expected line 12 error, got line %d level %s", errs[0].Line, errs[0].Level)
	}
	if errs[0].Message != "[E1024] Symbol 'X' has not been declared" {
		t.Errorf("unexpected message: %s", errs[0].Message)
	}

	if errs[1].Line != 15 || errs[1].Level != "warning" {
		t.Errorf("expected line 15 warning, got line %d level %s", errs[1].Line, errs[1].Level)
	}
	if errs[1].Message != "[W2011] Variable 'Y' defined but not used" {
		t.Errorf("unexpected message: %s", errs[1].Message)
	}
}

func TestParseErrors_WatcomColumn(t *testing.T) {
	output := `CALC.FOR(24,7): Error! E1005: Syntax error near '+'`
	errs := ParseErrors(output, "")
	if len(errs) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(errs))
	}
	if errs[0].Line != 24 || errs[0].Column != 7 || errs[0].Level != "error" {
		t.Errorf("unexpected error info: %+v", errs[0])
	}
}

func TestParseErrors_StandardUnix(t *testing.T) {
	output := `math.f:30:15: error: division by zero`
	errs := ParseErrors(output, "")
	if len(errs) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(errs))
	}
	if errs[0].Line != 30 || errs[0].Column != 15 || errs[0].Level != "error" {
		t.Errorf("unexpected error info: %+v", errs[0])
	}
}

func TestIsFortranSource(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{"hello.for", true},
		{"MAIN.F", true},
		{"TEST.F77", true},
		{"header.inc", true},
		{"main.go", false},
		{"test.rs", false},
	}

	for _, c := range cases {
		got := IsFortranSource(c.path)
		if got != c.expected {
			t.Errorf("IsFortranSource(%s) = %v, expected %v", c.path, got, c.expected)
		}
	}
}
