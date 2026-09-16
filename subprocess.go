package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// Subprocess is the default Factory: it runs argv as an external process,
// with argv[0] the program and argv[1:] its arguments. It is built on the
// framework's streaming subprocess engine (the same one cmd-exec and cmd-perl
// use), so a long-running child streams its output and is torn down on a
// downstream stop or cancellation.
func Subprocess(argv []string) gloo.Command[[]byte, []byte] {
	return patterns.Subprocess(patterns.ProcessName(argv[0]), processArgs(argv[1:])...)
}

// processArgs converts the argv tail into the patterns.ProcessArg vector
// patterns.Subprocess consumes.
func processArgs(args []string) []patterns.ProcessArg {
	out := make([]patterns.ProcessArg, len(args))
	for i, a := range args {
		out[i] = patterns.ProcessArg(a)
	}
	return out
}
