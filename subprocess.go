package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// Subprocess is the default CommandFor: it runs argv as an external process,
// with argv[0] the program and argv[1:] its arguments. It is built on the
// framework's streaming subprocess engine (the same one cmd-exec and cmd-perl
// use), so a long-running child streams its output and is torn down on a
// downstream stop or cancellation.
func Subprocess(argv []string) gloo.Command[[]byte, []byte] {
	return patterns.Subprocess(argv[0], argv[1:]...)
}
