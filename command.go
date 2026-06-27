package command

import (
	"bytes"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/framework/patterns"
)

// fieldCount is the maximum number of input fields packed into one output line
// (the -n argument). It is always >= 1 after normalization.
type fieldCount int

// Xargs returns a Command with two modes. With no command to run it regroups:
// it splits each input line into fields and groups them into output lines of at
// most N fields (-n mode; default -n1, each field on its own line). With a
// command — positional arguments naming a program, or an injected XargsExec
// factory — it runs that command once per argument group, streaming each
// invocation's output (the GNU xargs exec behavior).
func Xargs(opts ...any) gloo.Command[[]byte, []byte] {
	params := gloo.NewParameters[gloo.File, flags](opts...)
	f := params.Flags
	tmpl := template(params.Typed)
	run := resolveRunner(f, tmpl)
	if run == nil {
		return patterns.Expand(group(normalize(f.maxArgs)))
	}
	return execMode(f, tmpl, run)
}

// template renders the positional command arguments as plain strings: the
// program name followed by its fixed initial arguments.
func template(files []gloo.File) []string {
	out := make([]string, len(files))
	for i, fl := range files {
		out[i] = string(fl)
	}
	return out
}

// resolveRunner chooses the per-group command factory: an explicitly injected
// one, else a subprocess of the positional command, else nil (regroup mode).
func resolveRunner(f flags, tmpl []string) CommandFor {
	switch {
	case f.exec != nil:
		return f.exec
	case len(tmpl) > 0:
		return Subprocess
	default:
		return nil
	}
}

// normalize coerces the raw -n value to a usable per-line field count. Any
// non-positive value means "one field per line" (GNU xargs default behavior).
func normalize(n XargsMaxArgs) fieldCount {
	if n <= 0 {
		return 1
	}
	return fieldCount(n)
}

// group builds the Expand callback that packs each line's fields into output
// lines of at most n fields. Empty (whitespace-only) lines expand to nothing.
func group(n fieldCount) func([]byte) ([][]byte, error) {
	return func(line []byte) ([][]byte, error) {
		return chunk(bytes.Fields(line), n), nil
	}
}

// chunk joins fields into space-separated lines of at most n fields each.
func chunk(fields [][]byte, n fieldCount) [][]byte {
	var lines [][]byte
	for _, batch := range batches(fields, n) {
		lines = append(lines, bytes.Join(batch, []byte(" ")))
	}
	return lines
}

// batches splits fields into consecutive slices of at most n elements.
func batches(fields [][]byte, n fieldCount) [][][]byte {
	var out [][][]byte
	for i := 0; i < len(fields); i += int(n) {
		out = append(out, fields[i:min(i+int(n), len(fields))])
	}
	return out
}
