package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDefinitionInFortranProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf77_test_def_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "main.for")
	file2 := filepath.Join(tempDir, "math_sub.for")
	file3 := filepath.Join(tempDir, "io_sub.for")

	code1 := `C Main program
      PROGRAM TESTMOD
      INTEGER A, B, SUM, CALCSUM
      A = 10
      B = 20
      SUM = CALCSUM(A, B)
      CALL PRINTSUM(A, B, SUM)
      END
`
	code2 := `C Math Companion Function
      INTEGER FUNCTION CALCSUM(X, Y)
      INTEGER X, Y
      CALCSUM = X + Y
      RETURN
      END
`
	code3 := `C IO Companion Subroutine
      SUBROUTINE PRINTSUM(X, Y, S)
      INTEGER X, Y, S
      PRINT *, 'X=', X, ' Y=', Y, ' S=', S
      RETURN
      END
`

	if err := os.WriteFile(file1, []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file3, []byte(code3), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Find CALCSUM function definition
	f, line, col, ok := FindDefinitionInProject(file1, "CALCSUM")
	if !ok {
		t.Fatalf("expected to find definition for CALCSUM")
	}
	if filepath.Base(f) != "math_sub.for" {
		t.Errorf("expected math_sub.for, got %s", f)
	}
	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}
	if col < 1 {
		t.Errorf("invalid col %d", col)
	}

	// Test 2: Find PRINTSUM subroutine definition (case-insensitive test: "printsum")
	f, line, _, ok = FindDefinitionInProject(file1, "printsum")
	if !ok {
		t.Fatalf("expected to find definition for printsum (case-insensitive)")
	}
	if filepath.Base(f) != "io_sub.for" {
		t.Errorf("expected io_sub.for, got %s", f)
	}
	if line != 2 {
		t.Errorf("expected line 2, got %d", line)
	}

	// Test 3: Search in project
	matches := SearchInProject(file1, "CALCSUM", false)
	if len(matches) < 2 {
		t.Errorf("expected at least 2 matches for CALCSUM, got %d", len(matches))
	}
}
