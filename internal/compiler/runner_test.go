package compiler

import (
	"os"
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
		{"module.f90", true},
		{"modern.F95", true},
		{"calc.f08", true},
		{"main.go", false},
		{"test.rs", false},
	}

	for _, c := range cases {
		got := IsFortranSource(c.path)
		if got != c.expected {
			t.Errorf("IsFortranSource(%s) = %v, expected %v", c.path, got, c.expected)
		}
	}

	if !IsFreeFormFortran("math.f90") || !IsFreeFormFortran("algo.F95") {
		t.Errorf("expected IsFreeFormFortran to be true for f90/f95")
	}
	if IsFreeFormFortran("legacy.for") || IsFreeFormFortran("classic.f") {
		t.Errorf("expected IsFreeFormFortran to be false for .for/.f")
	}
	if !IsFixedFormFortran("legacy.for") || !IsFixedFormFortran("classic.f") {
		t.Errorf("expected IsFixedFormFortran to be true for .for/.f")
	}
}

func TestHasProgramStatement(t *testing.T) {
	tempDir := t.TempDir()

	mainFile := tempDir + "/main.for"
	os.WriteFile(mainFile, []byte("C Comment line\n      PROGRAM MAIN\n      WRITE(*,*) 'HI'\n      END\n"), 0644)

	subFile := tempDir + "/sub.for"
	os.WriteFile(subFile, []byte("C Subroutine\n      SUBROUTINE MYSUB(X)\n      INTEGER X\n      X = X + 1\n      RETURN\n      END\n"), 0644)

	commentProg := tempDir + "/comm.for"
	os.WriteFile(commentProg, []byte("C      PROGRAM FAKE\n!      PROGRAM NOT_REAL\n      SUBROUTINE TEST()\n      END\n"), 0644)

	if !HasProgramStatement(mainFile) {
		t.Errorf("expected HasProgramStatement(main.for) to be true")
	}
	if HasProgramStatement(subFile) {
		t.Errorf("expected HasProgramStatement(sub.for) to be false")
	}
	if HasProgramStatement(commentProg) {
		t.Errorf("expected HasProgramStatement(comm.for) to be false because PROGRAM is commented")
	}
}

func TestMultiFileDiscovery(t *testing.T) {
	tempDir := t.TempDir()

	mainFile := tempDir + "/main.for"
	os.WriteFile(mainFile, []byte("      PROGRAM MAIN\n      CALL MYSUB()\n      END\n"), 0644)

	sub1 := tempDir + "/sub1.for"
	os.WriteFile(sub1, []byte("      SUBROUTINE MYSUB()\n      END\n"), 0644)

	sub2 := tempDir + "/sub2.f"
	os.WriteFile(sub2, []byte("      REAL FUNCTION CALC(A)\n      CALC = A * 2.0\n      RETURN\n      END\n"), 0644)

	otherMain := tempDir + "/other.for"
	os.WriteFile(otherMain, []byte("      PROGRAM OTHER\n      END\n"), 0644)

	// Companions of mainFile should be sub1.for and sub2.f, but NOT other.for
	companions := FindCompanionFiles(mainFile)
	if len(companions) != 2 {
		t.Fatalf("expected 2 companion files, got %d: %v", len(companions), companions)
	}

	totalLines := CountTotalLines(append([]string{mainFile}, companions...))
	if totalLines != 3+2+4 {
		t.Errorf("expected total lines 9, got %d", totalLines)
	}

	// Main program file for sub1 should be mainFile (since otherMain also exists, exactly 1 check should prevent ambiguity or pick main)
	os.Remove(otherMain)
	foundMain := FindMainProgramFile(tempDir, sub1)
	if foundMain != mainFile {
		t.Errorf("expected foundMain = %s, got %s", mainFile, foundMain)
	}
}

func TestModularExampleFiles(t *testing.T) {
	// Test on the actual examples/modular directory
	companions := FindCompanionFiles("../../examples/modular/main.for")
	if len(companions) != 2 {
		t.Errorf("expected 2 companions for modular/main.for, got %d: %v", len(companions), companions)
	}

	// Test that hello.for does not include fibonacci.for or stats.for
	standaloneCompanions := FindCompanionFiles("../../examples/hello.for")
	if len(standaloneCompanions) != 0 {
		t.Errorf("expected 0 companions for hello.for, got %d: %v", len(standaloneCompanions), standaloneCompanions)
	}
}
