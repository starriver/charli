package charli_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	cli "github.com/starriver/charli"
)

var app = cli.App{
	Commands: []cli.Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
			Options: []cli.Option{
				{
					Short: 'f',
					Flag:  true,
				},
				{
					Long: "value",
				},
				{
					Short:    'c',
					Long:     "choice",
					Choices:  []string{"aa", "bb"},
					Metavar:  "C",
					Headline: "Choice headline",
				},
			},
		},
		{
			Name: "cmd2",
		},
	},
	GlobalOptions: []cli.Option{
		{
			Short: 'o',
			Flag:  true,
		},
	},
}

var appWithDefault = app
var appSingleCmd = app
var appHelpCmd = app
var appSingleCmdWithHelp = app
var appHelpBoth = app

func init() {
	appWithDefault.DefaultCommand = "cmd1"

	appSingleCmd.Commands = app.Commands[:1]
	appSingleCmd.GlobalOptions = []cli.Option{}

	appHelpCmd.HelpAccess = cli.HelpCommand

	appSingleCmdWithHelp = appSingleCmd
	appSingleCmdWithHelp.HelpAccess = cli.HelpCommand

	appHelpBoth.HelpAccess = cli.HelpFlag | cli.HelpCommand
}

func TestComplete(t *testing.T) {
	tests := []struct {
		app       cli.App
		argv      []string
		i         int
		want      []string
		wantPanic bool
	}{
		{
			app:  app,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"cmd1\tHeadline1",
				"cmd2\tCommand",
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "c"},
			i:    2,
			want: []string{
				"cmd1\tHeadline1",
				"cmd2\tCommand",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "-"},
			i:    2,
			want: []string{
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "x"},
			i:    2,
			want: []string{},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1"},
			i:    2,
			want: []string{"cmd1\tHeadline1"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1"},
			i:    3,
			want: []string{
				"-o\tFlag",
				"-f\tFlag",
				"--value\tOption",
				"-c\tChoice headline",
				"--choice\tChoice headline",
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "--"},
			i:    3,
			want: []string{
				"--value\tOption",
				"--choice\tChoice headline",
				"--help\tShow help",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "--", "--"},
			i:    4,
			want: []string{},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "-c"},
			i:    4,
			want: []string{
				"aa\t-c C",
				"bb\t-c C",
			},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "-c", "a"},
			i:    4,
			want: []string{"aa\t-c C"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "-x"},
			i:    3,
			want: []string{},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "-c", "a"},
			i:    2,
			want: []string{"cmd1\tHeadline1"},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"cmd1\tHeadline1",
				"cmd2\tCommand",
				"-o\tFlag",
				"-f\tFlag",
				"--value\tOption",
				"-c\tChoice headline",
				"--choice\tChoice headline",
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c", "--v"},
			i:    2,
			want: []string{"--value\tOption"},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c", "cmd2"},
			i:    3,
			want: []string{"-o\tFlag", "-h\tShow help", "--help\tShow help"},
		},
		{
			app:  appSingleCmd,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"-f\tFlag",
				"--value\tOption",
				"-c\tChoice headline",
				"--choice\tChoice headline",
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:  appSingleCmd,
			argv: []string{"program", "_c", "--v"},
			i:    2,
			want: []string{"--value\tOption"},
		},
		{
			app:  appHelpCmd,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"cmd1\tHeadline1",
				"cmd2\tCommand",
				"help\tShow help",
			},
		},
		{
			app:  appSingleCmdWithHelp,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"help\tShow help",
				"-f\tFlag",
				"--value\tOption",
				"-c\tChoice headline",
				"--choice\tChoice headline",
			},
		},
		{
			app:  appHelpBoth,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{
				"cmd1\tHeadline1",
				"cmd2\tCommand",
				"help\tShow help",
				"-h\tShow help",
				"--help\tShow help",
			},
		},
		{
			app:       app,
			argv:      []string{"program"},
			i:         1,
			wantPanic: true,
		},
		{
			app:       app,
			argv:      []string{"program", "_c"},
			i:         3,
			wantPanic: true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d %v", i, test.argv), func(t *testing.T) {
			defer func() {
				r := recover()
				if test.wantPanic {
					if r == nil {
						t.Error("expected panic")
					}
				} else if r != nil {
					panic(r)
				}
			}()

			var buf bytes.Buffer
			test.app.Complete(&buf, test.i, test.argv)
			got := strings.TrimSpace(buf.String())
			want := strings.Join(test.want, "\n")

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(got, want, false)

			t.Log(dmp.DiffPrettyText(diffs))
			for _, diff := range diffs {
				if diff.Type != diffmatchpatch.DiffEqual {
					t.Fail()
				}
			}
		})
	}
}

var wantBash = `
_complete_charli_a_a_-A0_() {
	for c in $('a+a_-A0\'' --_complete ${COMP_WORDS[@]:1:$COMP_CWORD}); do
		COMPREPLY+=("${c%%	*}")
	done
}
complete -o bashdefault -F _complete_charli_a_a_-A0_ 'a+a_-A0\''
`

var wantFish = `
function __complete_charli_a_a_-A0_
	set -l tokens (commandline -cop)
	'a+a_-A0\'' --_complete $tokens[1..-1]
end
complete -c 'a+a_-A0\'' -a "(__complete_charli_a_a_-A0_)"
`

func TestCompletionScripts(t *testing.T) {
	// Use a really ugly name to test the identifiers + escaping.
	program := "a+a_-A0'"
	flag := "--_complete"

	var buf bytes.Buffer
	cli.GenerateBashCompletions(&buf, program, flag)
	gotBash := buf.String()

	buf = bytes.Buffer{}
	cli.GenerateFishCompletions(&buf, program, flag)
	gotFish := buf.String()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"bash", gotBash, wantBash},
		{"fish", gotFish, wantFish},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(test.got, test.want[1:], false)

			t.Log(dmp.DiffPrettyText(diffs))
			for _, diff := range diffs {
				if diff.Type != diffmatchpatch.DiffEqual {
					t.Fail()
				}
			}
		})
	}
}
