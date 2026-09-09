C =======================================================
C  Turbo F77 - Turbo FORTRAN 77 IDE
C  Open Watcom 2.0 FORTRAN 77 Toolchain
C =======================================================
      PROGRAM MAIN
      INTEGER I, SUM
      SUM = 0

      PRINT *, '======================================'
      PRINT *, '       WELCOME TO TURBO F77!          '
      PRINT *, '  Open Watcom 2.0 FORTRAN 77 IDE      '
      PRINT *, '======================================'

      DO 10 I = 1, 10
          SUM = SUM + I
          PRINT *, 'Step:', I, '  Accumulated Sum:', SUM
   10 CONTINUE

      PRINT *, '--------------------------------------'
      PRINT *, 'Final 1 to 10 Total Sum is:', SUM
      PRINT *, 'Press Alt+F5 to toggle User Screen!'
      STOP
      END
