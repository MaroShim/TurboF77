! =======================================================
! examples/modern_fibonacci.f90
! Fibonacci sequence using modern Fortran 90+ syntax
! Demonstrates: IMPLICIT NONE, DO loops, arrays, ALLOCATABLE
! =======================================================
program fibonacci
    implicit none

    integer, parameter :: n = 20
    integer, allocatable :: fib(:)
    integer :: i

    allocate(fib(n))

    fib(1) = 1
    fib(2) = 1

    do i = 3, n
        fib(i) = fib(i-1) + fib(i-2)
    end do

    print *, '======================================='
    print *, 'First', n, 'Fibonacci Numbers (F90)'
    print *, '======================================='
    do i = 1, n
        print '(a, i3, a, i10)', '  F(', i, ') = ', fib(i)
    end do

    deallocate(fib)
    stop
end program fibonacci
