package command_test

import (
	"testing"

	"github.com/gloo-foo/testable"
	"github.com/gloo-foo/testable/assertion"

	command "github.com/gloo-foo/cmd-xargs"
)

func TestXargs_DefaultSplitsEachField(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(), "one two three\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"one", "two", "three"})
}

func TestXargs_N1ExplicitSameAsDefault(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(1)), "a b c\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a", "b", "c"})
}

func TestXargs_N2GroupsPairs(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(2)), "a b c d\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a b", "c d"})
}

func TestXargs_N2OddRemainder(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(2)), "a b c\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a b", "c"})
}

func TestXargs_N3GroupsTriples(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(3)), "1 2 3 4 5 6 7\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"1 2 3", "4 5 6", "7"})
}

func TestXargs_EmptyInput(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(), "")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

func TestXargs_SingleField(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(), "only\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"only"})
}

func TestXargs_SingleFieldN2(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(2)), "only\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"only"})
}

func TestXargs_MultiLineInput(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(command.XargsMaxArgs(2)), "a b\nc d\n")
	assertion.NoError(t, err)
	assertion.Lines(t, lines, []string{"a b", "c d"})
}

func TestXargs_WhitespaceOnlyLine(t *testing.T) {
	lines, err := testable.TestLines(command.Xargs(), "   \n")
	assertion.NoError(t, err)
	assertion.Empty(t, lines)
}

func TestXargs_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		opts     []any
		input    string
		expected []string
	}{
		{"default-3-fields", nil, "x y z\n", []string{"x", "y", "z"}},
		{"n2-even", []any{command.XargsMaxArgs(2)}, "a b c d\n", []string{"a b", "c d"}},
		{"n2-odd", []any{command.XargsMaxArgs(2)}, "a b c\n", []string{"a b", "c"}},
		{"n4-fewer", []any{command.XargsMaxArgs(4)}, "a b\n", []string{"a b"}},
		{"tabs-and-spaces", nil, "a\tb  c\n", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := testable.TestLines(command.Xargs(tt.opts...), tt.input)
			assertion.NoError(t, err)
			assertion.Lines(t, lines, tt.expected)
		})
	}
}
