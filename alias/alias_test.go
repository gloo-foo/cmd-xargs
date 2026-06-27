package alias_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	xargs "github.com/gloo-foo/cmd-xargs/alias"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
)

// record is a fake exec factory: it emits the argv it was handed as one line,
// so an alias test can prove the exec/grouping re-exports are wired correctly.
func record(argv []string) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, _ func(error)) {
			send([]byte(strings.Join(argv, " ")))
		})
	})
}

// The alias package re-exports the constructor and the -n flag type under
// unprefixed names. A mis-wired re-export (Xargs bound to the wrong function, or
// MaxArgs aliased to an unrelated type) compiles cleanly, so only behavior can
// prove the wiring. Each test exercises one re-export and asserts the GNU xargs
// output it must produce.

const groupingInput = "a b c d e\n"

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAlias_DefaultSplitsEachField(t *testing.T) {
	// With no flag, xargs emits one field per line (equivalent to -n1).
	lines, err := testable.TestLines(xargs.Xargs(), groupingInput)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"a", "b", "c", "d", "e"})
}

func TestAlias_MaxArgsGroupsFields(t *testing.T) {
	// MaxArgs(2) packs at most two fields per line; the odd field trails alone.
	lines, err := testable.TestLines(xargs.Xargs(xargs.MaxArgs(2)), groupingInput)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"a b", "c d", "e"})
}

func TestAlias_MaxArgsOneMatchesDefault(t *testing.T) {
	// MaxArgs(1) is the explicit form of the default: one field per line.
	lines, err := testable.TestLines(xargs.Xargs(xargs.MaxArgs(1)), groupingInput)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"a", "b", "c", "d", "e"})
}

func TestAlias_MaxArgsNonPositiveMatchesDefault(t *testing.T) {
	// A non-positive MaxArgs normalizes to one field per line, like the default.
	lines, err := testable.TestLines(xargs.Xargs(xargs.MaxArgs(0)), groupingInput)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"a", "b", "c", "d", "e"})
}

func TestAlias_ExecGroupsAndRunsPerGroup(t *testing.T) {
	// Exec + MaxArgs + MaxProcs: two fields per invocation, run concurrently but
	// in input order.
	cmd := xargs.Xargs(gloo.File("echo"), xargs.MaxArgs(2), xargs.MaxProcs(4), xargs.Exec(record))
	lines, err := testable.TestLines(cmd, "a b c\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"echo a b", "echo c"})
}

func TestAlias_ReplaceSubstitutesPerLine(t *testing.T) {
	// Replace runs once per line, substituting the token in the template.
	cmd := xargs.Xargs(gloo.File("echo"), gloo.File("[{}]"), xargs.Replace("{}"), xargs.Exec(record))
	lines, err := testable.TestLines(cmd, "a b\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"echo [a b]", "echo [c]"})
}

func TestAlias_NullSplitsOnNUL(t *testing.T) {
	// Null splits input on NUL, keeping embedded spaces in each argument.
	cmd := xargs.Xargs(gloo.File("echo"), xargs.Null(true), xargs.Exec(record))
	lines, err := testable.TestLines(cmd, "a b\x00c")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"echo a b c"})
}

func TestAlias_SubprocessRunsExternalProgram(t *testing.T) {
	// Subprocess is the default external-process factory.
	cmd := xargs.Xargs(gloo.File("echo"), xargs.Exec(xargs.Subprocess))
	lines, err := testable.TestLines(cmd, "x y\n")
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, lines, []string{"x y"})
}
