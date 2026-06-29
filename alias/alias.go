// Package alias provides unprefixed type aliases for xargs command flags.
// This allows users to import and use shorter names:
//
//	import "github.com/gloo-foo/cmd-xargs/alias"
//	alias.Xargs(alias.MaxArgs(2))
package alias

import command "github.com/gloo-foo/cmd-xargs"

// Xargs re-exports the constructor.
var Xargs = command.Xargs

// Subprocess re-exports the default external-process exec factory.
var Subprocess = command.Subprocess

type (
	// MaxArgs is the -n flag: max arguments per group.
	MaxArgs = command.XargsMaxArgs
	// Null is the -0 flag: split input on NUL bytes.
	Null = command.XargsNull
	// Replace is the -I flag: substitute a token per input line.
	Replace = command.XargsReplace
	// MaxProcs is the -P flag: max concurrent invocations.
	MaxProcs = command.XargsMaxProcs
	// Exec injects the per-group command factory.
	Exec = command.XargsExec
	// CommandFor builds the command run for one group's argv.
	CommandFor = command.Factory
)
