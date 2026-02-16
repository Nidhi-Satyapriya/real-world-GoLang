package mathops

import "testing"

func TestFactorial(t *testing.T) {
	if Factorial(0) != 1 {
		t.Error("Factorial(0) should be 1")
	}
	if Factorial(5) != 120 {
		t.Error("Factorial(5) should be 120")
	}
}

func TestFibonacci(t *testing.T) {
	if Fibonacci(0) != 0 {
		t.Error("Fibonacci(0) should be 0")
	}
	if Fibonacci(10) != 55 {
		t.Error("Fibonacci(10) should be 55")
	}
}
