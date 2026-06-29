package command

import (
	"bytes"
	"context"
	"strings"

	"github.com/destel/rill"
	gloo "github.com/gloo-foo/framework"
)

// execMode runs the per-group command once for each argument group, streaming
// each invocation's output. It uses GenerateFrom so a downstream stop or a
// cancellation chains to the input and to every in-flight child (which runs
// under the producer scope pctx).
func execMode(f flags, tmpl []string, run Factory) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, in gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.GenerateFrom(ctx, in, func(pctx context.Context, send func([]byte) bool, sendErr func(error)) {
			groups := groupArgs(in.Chan(), f)
			out := rill.OrderedFlatMap(groups, procs(f), func(g []string) <-chan rill.Try[[]byte] {
				return run(buildArgv(tmpl, f, g)).Execute(pctx, gloo.StreamOf[[]byte]()).Chan()
			})
			forward(out, send, sendErr)
		})
	})
}

// procs is the -P concurrency: at least one invocation at a time.
func procs(f flags) int {
	if f.maxProcs > 1 {
		return int(f.maxProcs)
	}
	return 1
}

// buildArgv assembles one invocation's argv. Without -I the group's arguments
// are appended to the command template; with -I every occurrence of the
// replace-token in the template is substituted with the group's line.
func buildArgv(tmpl []string, f flags, group []string) []string {
	if f.replace != "" {
		return substitute(tmpl, string(f.replace), strings.Join(group, " "))
	}
	argv := make([]string, 0, len(tmpl)+len(group))
	argv = append(argv, tmpl...)
	argv = append(argv, group...)
	return argv
}

// substitute replaces every occurrence of token in each template element with
// value (the -I behavior).
func substitute(tmpl []string, token, value string) []string {
	out := make([]string, len(tmpl))
	for i, t := range tmpl {
		out[i] = strings.ReplaceAll(t, token, value)
	}
	return out
}

// groupArgs splits the input stream into argument groups, emitting each group
// as soon as it fills (-n) so execution can start before all input arrives.
// With no -n the whole input forms a single group (the GNU exec default).
func groupArgs(in <-chan rill.Try[[]byte], f flags) <-chan rill.Try[[]string] {
	out := make(chan rill.Try[[]string])
	split := splitter(f)
	go func() {
		defer close(out)
		g := grouper{out: out, n: groupSize(f)}
		for item := range in {
			if item.Error != nil {
				out <- rill.Try[[]string]{Error: item.Error}
				return
			}
			g.add(split(item.Value))
		}
		g.flush()
	}()
	return out
}

// splitter selects how each input line is tokenized: -I treats the whole line
// as one argument, otherwise each line is split into whitespace-separated
// fields.
func splitter(f flags) func([]byte) []string {
	switch {
	case f.replace != "":
		return wholeLine
	case bool(f.null):
		return splitNull
	default:
		return splitFields
	}
}

// groupSize is the per-invocation argument count: -I forces one line per
// invocation; -n sets the field count; otherwise all fields form one group.
func groupSize(f flags) int {
	if f.replace != "" {
		return 1
	}
	return int(f.maxArgs)
}

// wholeLine yields the trimmed line as a single argument, or nothing when the
// line is blank.
func wholeLine(line []byte) []string {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return nil
	}
	return []string{string(trimmed)}
}

// grouper batches fields into groups of n, emitting each group once it fills.
// A non-positive n means "no limit": all fields collect into a single group.
type grouper struct {
	out chan<- rill.Try[[]string]
	buf []string
	n   int
}

// add appends fields, emitting a group whenever the -n size is reached.
func (g *grouper) add(fields []string) {
	for _, field := range fields {
		g.buf = append(g.buf, field)
		if g.n > 0 && len(g.buf) == g.n {
			g.emit()
		}
	}
}

// flush emits any trailing partial group.
func (g *grouper) flush() {
	if len(g.buf) > 0 {
		g.emit()
	}
}

// emit sends the buffered group and starts a fresh buffer.
func (g *grouper) emit() {
	g.out <- rill.Try[[]string]{Value: g.buf}
	g.buf = nil
}

// splitFields splits one input line into whitespace-separated fields.
func splitFields(line []byte) []string {
	parts := bytes.Fields(line)
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = string(p)
	}
	return out
}

// splitNull splits on NUL bytes (-0), keeping each non-empty run verbatim so an
// argument may contain spaces or newlines.
func splitNull(line []byte) []string {
	var out []string
	for _, part := range bytes.Split(line, []byte{0}) {
		if len(part) > 0 {
			out = append(out, string(part))
		}
	}
	return out
}

// forward pumps each invocation's output downstream, stopping promptly when the
// consumer walks away.
func forward(out <-chan rill.Try[[]byte], send func([]byte) bool, sendErr func(error)) {
	for item := range out {
		if item.Error != nil {
			sendErr(item.Error)
			continue
		}
		if !send(item.Value) {
			return
		}
	}
}
