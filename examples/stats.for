C =======================================================
C  Example 3: Array Statistics & Root Mean Square
C =======================================================
      PROGRAM STATS
      INTEGER N, I
      PARAMETER (N = 5)
      REAL A(N), SUM, AVG, SQSUM, RMS

      DATA A / 12.5, 18.2, 9.8, 25.1, 14.4 /

      SUM = 0.0
      SQSUM = 0.0

      PRINT *, 'Computing array statistics for:'
      DO 10 I = 1, N
          PRINT *, 'A(', I, ') =', A(I)
          SUM = SUM + A(I)
          SQSUM = SQSUM + A(I)*A(I)
   10 CONTINUE

      AVG = SUM / FLOAT(N)
      RMS = SQRT(SQSUM / FLOAT(N))

      PRINT *, '--------------------------------'
      PRINT *, 'Sum     =', SUM
      PRINT *, 'Average =', AVG
      PRINT *, 'RMS     =', RMS
      PRINT *, '--------------------------------'
      STOP
      END
