C =======================================================
C  Example 2: Fibonacci Sequence Generator
C =======================================================
      PROGRAM FIBONACCI
      INTEGER N, I
      INTEGER F1, F2, NEXT
      PARAMETER (N = 15)

      PRINT *, 'First', N, 'Fibonacci Numbers:'
      PRINT *, '--------------------------------'

      F1 = 1
      F2 = 1

      PRINT *, ' 1:', F1
      PRINT *, ' 2:', F2

      DO 20 I = 3, N
          NEXT = F1 + F2
          PRINT *, I, ':', NEXT
          F1 = F2
          F2 = NEXT
   20 CONTINUE

      PRINT *, '--------------------------------'
      PRINT *, 'Calculation Complete.'
      STOP
      END
