package xargs_test

import (
	"fmt"

	command "github.com/gloo-foo/cmd-xargs"
	"github.com/gloo-foo/testable"
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
