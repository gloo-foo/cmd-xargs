package xargs_test

import (
	"fmt"

	"github.com/gloo-foo/testable"

	command "github.com/gloo-foo/cmd-xargs"
)

func ExampleXargs_basic() {
	// echo "one two three" | xargs -n1
	output, _ := testable.Test(command.Xargs(), "one two three\n")
	fmt.Print(output)
	// Output:
	// one
	// two
	// three
}
