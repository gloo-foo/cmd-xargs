package command

import gloo "github.com/gloo-foo/framework"

// XargsMaxArgs is the -n flag: maximum number of fields per output line.
// When unset or <= 0, defaults to 1 (each field on its own line).
type XargsMaxArgs int

// Factory turns one group's full argv into the command to run for that group.
// argv[0] is the program/command name; argv[1:] its arguments. Because it
// returns a gloo.Command, the factory uniformly covers both an external process
// (see Subprocess) and any gloo-foo command of any composed complexity.
type Factory func(argv []string) gloo.Command[[]byte, []byte]

// XargsNull is the -0 flag: split input on NUL bytes instead of whitespace, so
// arguments may contain spaces and newlines.
type XargsNull bool

// XargsReplace is the -I flag: a replace-token substituted for each input line
// within the command template. Setting it forces one invocation per input line.
type XargsReplace string

// XargsMaxProcs is the -P flag: the maximum number of invocations to run
// concurrently. Output stays in input order regardless. Values <= 1 run one
// invocation at a time.
type XargsMaxProcs int

// XargsExec injects the factory used to build the command run for each group.
// When unset, a command named by positional arguments runs as a subprocess.
type XargsExec Factory

type flags struct {
	exec            Factory
	replace         XargsReplace
	maxArgs         XargsMaxArgs
	maxProcs        XargsMaxProcs
	isNullDelimited XargsNull
}

// with folds one option value into the flags, reporting whether o was one of
// this command's own options (as opposed to a positional argument).
func (f flags) with(o any) (flags, bool) {
	switch v := o.(type) {
	case XargsMaxArgs:
		f.maxArgs = v
	case XargsNull:
		f.isNullDelimited = v
	case XargsReplace:
		f.replace = v
	case XargsMaxProcs:
		f.maxProcs = v
	case XargsExec:
		f.exec = Factory(v)
	default:
		return f, false
	}
	return f, true
}

// parseOptions folds the command's own options out of opts, returning the
// resulting flags and the remaining (positional) arguments.
func parseOptions(opts []any) (flags, []any) {
	var f flags
	rest := make([]any, 0, len(opts))
	for _, o := range opts {
		next, isOption := f.with(o)
		if !isOption {
			rest = append(rest, o)
			continue
		}
		f = next
	}
	return f, rest
}
