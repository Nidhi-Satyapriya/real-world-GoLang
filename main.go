package main

import (
	"fmt"

	"github.com/nisatyap/golearn/mathops"
)

func main() {
	var n int

	fmt.Print("Enter a number: ")
	fmt.Scan(&n)

	fmt.Printf("Factorial(%d) = %d\n", n, mathops.Factorial(n))
	fmt.Printf("Fibonacci(%d) = %d\n", n, mathops.Fibonacci(n))
}
