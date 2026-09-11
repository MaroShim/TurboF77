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

func TestFindDefinitionFortranEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf77_test_edge_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file := filepath.Join(tempDir, "complex_prog.for")
	code := `C Line 1 comment: SUBROUTINE FAKE_SUB()
* Line 2 asterisk comment: FUNCTION FAKE_FUNC()
! Line 3 exclamation comment: PROGRAM FAKE_PROG
      DOUBLE PRECISION FUNCTION CALC_PI(RADIUS)
      DOUBLE PRECISION RADIUS
      CALC_PI = 3.1415926535D0 * RADIUS
      RETURN
      END

      ENTRY SUB_ENTRY(X)
      RETURN
      END

      BLOCK DATA INIT_VALUES
      COMMON /BLK/ VAL1
      DATA VAL1 /100/
      END
`
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Comments with keywords must not be matched as definitions
	_, _, _, ok := FindDefinitionInProject(file, "FAKE_SUB")
	if ok {
		t.Errorf("expected FAKE_SUB in 'C' comment to NOT be matched")
	}
	_, _, _, ok = FindDefinitionInProject(file, "FAKE_FUNC")
	if ok {
		t.Errorf("expected FAKE_FUNC in '*' comment to NOT be matched")
	}
	_, _, _, ok = FindDefinitionInProject(file, "FAKE_PROG")
	if ok {
		t.Errorf("expected FAKE_PROG in '!' comment to NOT be matched")
	}

	// 2. Multi-word type prefix: DOUBLE PRECISION FUNCTION
	f, line, _, ok := FindDefinitionInProject(file, "calc_pi")
	if !ok || line != 4 {
		t.Errorf("expected calc_pi at line 4, got %s:%d (ok=%v)", f, line, ok)
	}

	// 3. ENTRY statement
	_, line, _, ok = FindDefinitionInProject(file, "sub_entry")
	if !ok || line != 10 {
		t.Errorf("expected sub_entry at line 10, got line %d", line)
	}

	// 4. BLOCK DATA statement
	_, line, _, ok = FindDefinitionInProject(file, "init_values")
	if !ok || line != 14 {
		t.Errorf("expected init_values at line 14, got line %d", line)
	}

	// 5. Blank query & empty searches
	_, _, _, ok = FindDefinitionInProject(file, "   ")
	if ok {
		t.Errorf("expected blank query to return false")
	}
	matches := SearchInProject(file, "", false)
	if len(matches) != 0 {
		t.Errorf("expected empty search matches for empty query")
	}
}

func TestFindDefinitionModernFortran(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf_test_modern_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	modFile := filepath.Join(tempDir, "matrix_mod.f90")
	code := `! Modern Fortran Module
module matrix_ops
    implicit none
    type matrix_t
        real, allocatable :: data(:,:)
    end type matrix_t
contains
    subroutine init_matrix(m, rows, cols)
        type(matrix_t), intent(out) :: m
        integer, intent(in) :: rows, cols
        allocate(m%data(rows, cols))
    end subroutine init_matrix
end module matrix_ops
`
	if err := os.WriteFile(modFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Find MODULE definition
	f, line, _, ok := FindDefinitionInProject(modFile, "matrix_ops")
	if !ok || line != 2 {
		t.Errorf("expected matrix_ops module at line 2, got %s:%d (ok=%v)", f, line, ok)
	}

	// 2. Find TYPE definition
	f, line, _, ok = FindDefinitionInProject(modFile, "matrix_t")
	if !ok || line != 4 {
		t.Errorf("expected matrix_t type at line 4, got %s:%d (ok=%v)", f, line, ok)
	}

	// 3. Find Subroutine inside module
	f, line, _, ok = FindDefinitionInProject(modFile, "init_matrix")
	if !ok || line != 8 {
		t.Errorf("expected init_matrix subroutine at line 8, got %s:%d (ok=%v)", f, line, ok)
	}
}

