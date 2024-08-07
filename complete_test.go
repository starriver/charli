package charli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestComplete(t *testing.T) {
	app := App{
		Commands: []Command{
			{
				Name: "cmd1",
				Options: []Option{
					{
						Short: 'f',
						Flag:  true,
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

	appWithDefault := app
	appWithDefault.DefaultCommand = "cmd1"

	appSingleCmd := app
	appSingleCmd.Commands = appSingleCmd.Commands[:1]
	appSingleCmd.GlobalOptions = []Option{}

	appHelpCmd := app
	appHelpCmd.HelpAccess = HelpCommand

	appSingleCmdWithHelp := appSingleCmd
	appSingleCmdWithHelp.HelpAccess = HelpCommand

	appHelpBoth := app
	appHelpBoth.HelpAccess = HelpFlag | HelpCommand

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
		t.Run(fmt.Sprintf("Complete %d %v", i, test.argv), func(t *testing.T) {
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
			got := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			if len(got) == 1 && got[0] == "" {
				got = []string{}
			}

			if diff := deep.Equal(got, test.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

var wantBashCompletion = `_complete_aa() {
	for c in $('a+a\'' _c (( $COMP_CWORD + 1 )) $COMP_WORDS); do
		COMPREPLY+=("$c")
	done
}
complete -o bashdefault -F _complete_aa 'a+a\''
`

func TestGenerateBashCompletions(t *testing.T) {
	// This function actually doesn't touch the App at all ftm - but it might
	// in future. GenerateFishCompletions *does* touch the App. So for
	// consistency, all of them have App as receiver.
	app := App{}
	var buf bytes.Buffer
	app.GenerateBashCompletions(&buf, "a+a'", "_c")
	got := buf.String()

	if diff := deep.Equal(got, wantBashCompletion); diff != nil {
		t.Error(diff)
	}
}
