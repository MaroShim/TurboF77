C =======================================================
C  Modular Fortran 77 Example: Main Program
C =======================================================
      PROGRAM MODULAR
      INTEGER A, B, SUM, CALCSUM
      A = 25
      B = 17
      SUM = CALCSUM(A, B)
      CALL PRINTSUM(A, B, SUM)
      END
