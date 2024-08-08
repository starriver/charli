package charli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var app = App{
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
			Options: []Option{
				{
					Short:    'f',
					Flag:     true,
					Headline: "Flag",
				},
				{
					Long: "value",
				},
				{
					Short:   'c',
					Long:    "choice",
					Choices: []string{"aa", "bb"},
				},
			},
		},
		{
			Name: "cmd2",
		},
	},
	GlobalOptions: []Option{
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
	appSingleCmd.GlobalOptions = []Option{}

	appHelpCmd.HelpAccess = HelpCommand

	appSingleCmdWithHelp = appSingleCmd
	appSingleCmdWithHelp.HelpAccess = HelpCommand

	appHelpBoth.HelpAccess = HelpFlag | HelpCommand
}

func TestComplete(t *testing.T) {
	tests := []struct {
		app       App
		argv      []string
		i         int
		want      []string
		wantPanic bool
	}{
		{
			app:  app,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"cmd1", "cmd2", "-h", "--help"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "c"},
			i:    2,
			want: []string{"cmd1", "cmd2"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "-"},
			i:    2,
			want: []string{"-h", "--help"},
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
			want: []string{"cmd1"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1"},
			i:    3,
			want: []string{"-o", "-f", "--value", "-c", "--choice", "-h", "--help"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "--"},
			i:    3,
			want: []string{"--value", "--choice", "--help"},
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
			want: []string{"aa", "bb"},
		},
		{
			app:  app,
			argv: []string{"program", "_c", "cmd1", "-c", "a"},
			i:    4,
			want: []string{"aa"},
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
			want: []string{"cmd1"},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"cmd1", "cmd2", "-o", "-f", "--value", "-c", "--choice", "-h", "--help"},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c", "--v"},
			i:    2,
			want: []string{"--value"},
		},
		{
			app:  appWithDefault,
			argv: []string{"program", "_c", "cmd2"},
			i:    3,
			want: []string{"-o", "-h", "--help"},
		},
		{
			app:  appSingleCmd,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"-f", "--value", "-c", "--choice", "-h", "--help"},
		},
		{
			app:  appSingleCmd,
			argv: []string{"program", "_c", "--v"},
			i:    2,
			want: []string{"--value"},
		},
		{
			app:  appHelpCmd,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"cmd1", "cmd2", "help"},
		},
		{
			app:  appSingleCmdWithHelp,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"help", "-f", "--value", "-c", "--choice"},
		},
		{
			app:  appHelpBoth,
			argv: []string{"program", "_c"},
			i:    2,
			want: []string{"cmd1", "cmd2", "help", "-h", "--help"},
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

var wantBashCompletion = `_complete_a_a_-A0_() {
	for c in $('a+a_-A0\'' _c (( $COMP_CWORD + 1 )) $COMP_WORDS); do
		COMPREPLY+=("$c")
	done
}
complete -o bashdefault -F __complete_a_a_-A0 'a+a_-A0\''
`

func TestGenerateBashCompletions(t *testing.T) {
	// This function actually doesn't touch the App at all ftm - but it might
	// in future. GenerateFishCompletions *does* touch the App. So for
	// consistency, all of them have App as receiver.

	var buf bytes.Buffer
	// Use a really ugly name to test the identifiers + escaping.
	GenerateBashCompletions(&buf, "a+a_-A0'", "_c")
	got := buf.String()

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(got, wantBashCompletion, false)

	t.Log(dmp.DiffPrettyText(diffs))
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			t.Fail()
		}
	}
}

// TODO

func TestGenerateFishCompletions(t *testing.T) {
	tests := []struct {
		app  App
		want string
	}{
		{app, fishApp},
		{appWithDefault, fishAppWithDefault},
		{appSingleCmd, fishAppSingleCmd},
		{appHelpCmd, fishAppHelpCmd},
		{appSingleCmdWithHelp, fishAppSingleCmdWithHelp},
		{appHelpBoth, fishAppHelpBoth},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d %v", i, test), func(t *testing.T) {
			var buf bytes.Buffer
			test.app.GenerateFishCompletions(&buf, "a'")
			got := buf.String()

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(got, test.want[1:], false)

			t.Log(dmp.DiffPrettyText(diffs))
			for _, diff := range diffs {
				if diff.Type != diffmatchpatch.DiffEqual {
					t.Fail()
				}
			}
		})
	}

}
