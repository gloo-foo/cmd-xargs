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
	exec     Factory
	replace  XargsReplace
	maxArgs  XargsMaxArgs
	maxProcs XargsMaxProcs
	null     XargsNull
}

func (m XargsMaxArgs) Configure(f *flags) { f.maxArgs = m }

func (n XargsNull) Configure(f *flags) { f.null = n }

func (r XargsReplace) Configure(f *flags) { f.replace = r }

func (p XargsMaxProcs) Configure(f *flags) { f.maxProcs = p }

func (x XargsExec) Configure(f *flags) { f.exec = Factory(x) }
