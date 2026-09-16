package command_test

import (
	"context"
	"strings"
	"testing"
	"time"

	gloo "github.com/gloo-foo/framework"
	"github.com/gloo-foo/testable"
	"github.com/gloo-foo/testable/assertion"

	command "github.com/gloo-foo/cmd-xargs"
)

// recordRun is a fake Factory: instead of running a real process it emits a
// single line recording the exact argv it was handed, so a test can assert how
// xargs grouped input and built each invocation's argv.
func recordRun(argv []string) gloo.Command[[]byte, []byte] {
	return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
		return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, _ func(error)) {
			send([]byte("run:" + strings.Join(argv, ",")))
		})
	})
}

func TestXargs_ExecTearsDownChildOnDownstreamStop(t *testing.T) {
	stopped := make(chan struct{})
	streaming := func(_ []string) gloo.Command[[]byte, []byte] {
		return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
			return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, _ func(error)) {
				running := true
				for running {
					running = send([]byte("tick")) // emit until the consumer stops
				}
				close(stopped) // send returned false: the downstream stop reached the child
			})
		})
	}
	cmd := command.Xargs(gloo.File("x"), command.XargsExec(streaming))
	out := cmd.Execute(context.Background(), gloo.StreamOf([]byte("a")))
	<-out.Chan() // pull one item, so the child is actively streaming
	out.Discard()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("child kept running after the consumer stopped reading")
	}
}

func TestXargs_ExecPropagatesChildError(t *testing.T) {
	boom := func(_ []string) gloo.Command[[]byte, []byte] {
		return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
			return gloo.Generate(ctx, func(_ context.Context, _ func([]byte) bool, sendErr func(error)) {
				sendErr(gloo.Error("child failed"))
			})
		})
	}
	cmd := command.Xargs(gloo.File("x"), command.XargsExec(boom))
	_, err := testable.TestLines(cmd, "a\n")
	assertion.ErrorContains(t, err, "child failed")
}

func TestXargs_ExecForwardsEveryOutputLineInOrder(t *testing.T) {
	echoEach := func(argv []string) gloo.Command[[]byte, []byte] {
		return gloo.FuncCommand[[]byte, []byte](func(ctx context.Context, _ gloo.Stream[[]byte]) gloo.Stream[[]byte] {
			return gloo.Generate(ctx, func(_ context.Context, send func([]byte) bool, _ func(error)) {
				for _, a := range argv {
					if !send([]byte(a)) {
						return
					}
				}
			})
		})
	}
	cmd := command.Xargs(command.XargsMaxArgs(2), command.XargsExec(echoEach))
	lines, err := testable.TestLines(cmd, "a b c\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c"})
}

func TestXargs_ExecDefaultsToSubprocessForPositionalCommand(t *testing.T) {
	cmd := command.Xargs(gloo.File("echo"), command.XargsMaxArgs(2))
	lines, err := testable.TestLines(cmd, "a b c\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a b", "c"})
}

func TestXargs_ExecPropagatesInputError(t *testing.T) {
	errIn := gloo.Generate(context.Background(), func(_ context.Context, _ func([]byte) bool, sendErr func(error)) {
		sendErr(gloo.Error("input broke"))
	})
	out := command.Xargs(gloo.File("x"), command.XargsExec(recordRun)).
		Execute(context.Background(), errIn)
	_, err := out.Collect()
	assertion.ErrorContains(t, err, "input broke")
}

func TestXargs_ExecReplaceSkipsBlankLines(t *testing.T) {
	cmd := command.Xargs(
		gloo.File("echo"), gloo.File("[{}]"),
		command.XargsReplace("{}"), command.XargsExec(recordRun),
	)
	lines, err := testable.TestLines(cmd, "a\n\nb\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,[a]", "run:echo,[b]"})
}

func TestXargs_ExecNullSkipsEmptyRuns(t *testing.T) {
	cmd := command.Xargs(gloo.File("echo"), command.XargsNull(true), command.XargsExec(recordRun))
	lines, err := testable.TestLines(cmd, "a\x00\x00b\x00")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,a,b"})
}

func TestXargs_ExecRunsOnceForAllFields(t *testing.T) {
	cmd := command.Xargs(gloo.File("echo"), command.XargsExec(recordRun))
	lines, err := testable.TestLines(cmd, "a b c\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,a,b,c"})
}

func TestXargs_ExecGroupsByMaxArgsAcrossInput(t *testing.T) {
	cmd := command.Xargs(gloo.File("echo"), command.XargsMaxArgs(2), command.XargsExec(recordRun))
	lines, err := testable.TestLines(cmd, "a b\nc d e\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,a,b", "run:echo,c,d", "run:echo,e"})
}

func TestXargs_ExecMaxProcsPreservesOutputOrder(t *testing.T) {
	cmd := command.Xargs(
		gloo.File("echo"),
		command.XargsMaxArgs(1), command.XargsMaxProcs(4), command.XargsExec(recordRun),
	)
	lines, err := testable.TestLines(cmd, "a b c d\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,a", "run:echo,b", "run:echo,c", "run:echo,d"})
}

func TestXargs_ExecNullSplitsOnNULPreservingSpaces(t *testing.T) {
	cmd := command.Xargs(
		gloo.File("echo"),
		command.XargsNull(true), command.XargsMaxArgs(2), command.XargsExec(recordRun),
	)
	lines, err := testable.TestLines(cmd, "a b\x00c d\x00e")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,a b,c d", "run:echo,e"})
}

func TestXargs_ExecReplaceSubstitutesTokenOncePerLine(t *testing.T) {
	cmd := command.Xargs(
		gloo.File("echo"), gloo.File("[{}]"),
		command.XargsReplace("{}"), command.XargsExec(recordRun),
	)
	lines, err := testable.TestLines(cmd, "a b\nc\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"run:echo,[a b]", "run:echo,[c]"})
}
