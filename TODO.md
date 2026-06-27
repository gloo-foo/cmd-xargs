# cmd.xargs — Unimplemented Features

The original wave-3 features are now implemented:

- ~~Subprocess execution: run a command with grouped args~~ — done (`Xargs` exec mode; positional command, or an injected `XargsExec` factory; see `Subprocess`).
- ~~`-0` null-terminated input~~ — done (`XargsNull`).
- ~~`-I` replace string~~ — done (`XargsReplace`).
- ~~`-P` parallel execution~~ — done (`XargsMaxProcs`, order-preserving).

Remaining GNU `xargs` features not yet covered:

- `-L max-lines`: use at most N input lines per command line.
- `-d delim`: custom input delimiter (besides whitespace and `-0`).
- `-r` / `--no-run-if-empty`: skip running the command when there is no input.
- `-t`: echo each command line to stderr before running it.
- Exit-status 123/124/125 semantics on child failure.
