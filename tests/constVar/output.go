package main

import "fmt"

var foo = "'foo' is a top-level variable."

const bar = "'bar' is a top-level constant."

var baz = quux()

var (
	alpha = "'alpha' is a top-level variable in a 'var(...)' expression."
	beta  = omega()
)

func omega() string {
	return "'beta' is like 'alpha', but from a function."
}
func quux() string {
	return "'baz' is a top-level variable from a function."
}
func main() {
	fmt.Printf("%s\n%s\n%s\n%s\n%s\n", foo, bar, baz, alpha, beta)
}
