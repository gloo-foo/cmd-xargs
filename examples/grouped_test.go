package xargs_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-xargs"
)

func ExampleXargs_grouped() {
	// echo "a b c d e" | xargs -n2
	output, _ := testable.Test(command.Xargs(command.XargsMaxArgs(2)), "a b c d e\n")
	fmt.Print(output)
	// Output:
	// a b
	// c d
	// e
}
