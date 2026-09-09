C =======================================================
C  Example 1: Classic Hello World in FORTRAN 77
C =======================================================
      PROGRAM HELLO
      CHARACTER*20 NAME
      NAME = 'Turbo F77'

      PRINT *, '**************************************'
      PRINT *, '*   Hello from ', NAME, '        *'
      PRINT *, '*   Powered by Open Watcom 2.0 F77   *'
      PRINT *, '**************************************'
      STOP
      END
