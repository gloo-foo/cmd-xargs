package xargs_test

import (
	"fmt"

	command "github.com/gloo-foo/cmd-xargs"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
)

func ExampleXargs_exec() {
	// echo "a b c" | xargs -n2 echo
	// A positional command runs once per argument group as a subprocess.
	output, _ := testable.Test(command.Xargs(gloo.File("echo"), command.XargsMaxArgs(2)), "a b c\n")
	fmt.Print(output)
	// Output:
	// a b
	// c
}
