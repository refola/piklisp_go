package main

import "fmt"

func square(n int) int {
	return (n * n)
}

func main() {
	fmt.Printf("5^2==%v\n", square(5))
}
