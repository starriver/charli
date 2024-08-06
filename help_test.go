package charli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// TODO: test color output.

// No default command, 2 subcommands
var testHelpApp1 = App{
	Headline:    "Headline",
	Description: "\nDescription\n",
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
}

const testHelpOutput1 = `
Headline
Usage: program [OPTIONS] COMMAND [...]

  Description

Options:
  -h/--help  Show this help

Commands:
  cmd1  Headline1
  cmd2  Headline2
`

// Default command, no description, 2 subcommands
var testHelpApp2 = App{
	Headline:       "Headline",
	DefaultCommand: "cmd1",
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
}

const testHelpOutput2 = `
Headline
Usage: program [OPTIONS] [COMMAND] [...]

Options:
  -h/--help  Show this help

Commands:
  cmd1  Headline1
  cmd2  Headline2
`

// No app headline, command help with headline, description, flags
var testHelpApp3 = App{
	Commands: []Command{
		{
			Name:        "cmd1",
			Headline:    "Headline",
			Description: "\nThis is a {description}\n",
			Options: []Option{
				{
					Short:    'a',
					Headline: "A headline",
					Choices:  []string{"a", "b", "c"},
				},
				{
					Short:    'b',
					Long:     "both",
					Metavar:  "BOTH",
					Headline: "B headline",
				},
				{
					Long: "flag",
					Flag: true,
				},
			},
		},
		{
			Name: "cmd2",
		},
	},
}

const testHelpOutput3 = `
Usage: program cmd1 [OPTIONS]

  Headline

  This is a description

Options:
  -h/--help       Show this help
  -a VALUE        A headline [a|b|c]
  -b/--both BOTH  B headline
  --flag
`

// Args only
var testHelpApp4 = App{
	Commands: []Command{
		{
			Name: "cmd1",
			Args: Args{
				Count:    3,
				Metavars: []string{"A", "B", "C"},
			},
		},
		{
			Name: "cmd2",
		},
	},
}

const testHelpOutput4 = `
Usage: program cmd1 [OPTIONS] A B C

Options:
  -h/--help  Show this help
`

// Args only, default metavars
var testHelpApp5 = App{
	Commands: []Command{
		{
			Name: "cmd1",
			Args: Args{
				Count:    3,
				Metavars: []string{"A"},
			},
		},
		{
			Name: "cmd2",
		},
	},
}

const testHelpOutput5 = `
Usage: program cmd1 [OPTIONS] A ARG ARG

Options:
  -h/--help  Show this help
`

// Args only, some required but varadic
var testHelpApp6 = App{
	Commands: []Command{
		{
			Name: "cmd1",
			Args: Args{
				Count:    2,
				Varadic:  true,
				Metavars: []string{"A", "B", "C"},
			},
		},
		{
			Name: "cmd2",
		},
	},
}

const testHelpOutput6 = `
Usage: program cmd1 [OPTIONS] A B [C...]

Options:
  -h/--help  Show this help
`

// Args only, all varadic
var testHelpApp7 = App{
	Commands: []Command{
		{
			Name: "cmd1",
			Args: Args{
				Count:    0,
				Varadic:  true,
				Metavars: []string{"A", "B", "C"},
			},
		},
		{
			Name: "cmd2",
		},
	},
}

const testHelpOutput7 = `
Usage: program cmd1 [OPTIONS] [A] [B] [C...]

Options:
  -h/--help  Show this help
`

// With global options
var testHelpApp8 = App{
	Commands: []Command{
		{
			Name: "cmd1",
			Options: []Option{
				{
					Short:    'b',
					Flag:     true,
					Headline: "Command",
				},
			},
		},
		{
			Name: "cmd2",
		},
	},
	GlobalOptions: []Option{
		{
			Short:    'a',
			Flag:     true,
			Headline: "Global",
		},
	},
}

const testHelpOutput8 = `
Usage: program cmd1 [OPTIONS]

Options:
  -h/--help  Show this help
  -a         Global
  -b         Command
`

// Command help with default command
var testHelpApp9 = App{
	Commands: []Command{
		{
			Name: "cmd1",
		},
		{
			Name: "cmd2",
		},
	},
	DefaultCommand: "cmd1",
}

const testHelpOutput9 = `
Usage: program [cmd1] [OPTIONS]

Options:
  -h/--help  Show this help
`

// Command help when only command
var testHelpApp10 = App{
	Commands: []Command{
		{},
	},
}

const testHelpOutput10 = `
Usage: program [OPTIONS]

Options:
  -h/--help  Show this help
`

// Help as command
var testHelpApp11 = App{
	Description: "\nDescription\n", // This is here to test spacing
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: HelpCommand,
}

const testHelpOutput11 = `
Usage: program COMMAND [...]

  Description

Commands:
  help  Show this help
  cmd1  Headline1
  cmd2  Headline2
`

// Help as flag & command
var testHelpApp12 = App{
	Description: "\nDescription\n", // This is here to test spacing
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: HelpFlag | HelpCommand,
}

const testHelpOutput12 = `
Usage: program [OPTIONS] COMMAND [...]

  Description

Options:
  -h/--help  Show this help

Commands:
  help  Show this help
  cmd1  Headline1
  cmd2  Headline2
`

// Command help, help as command, and no command options
var testHelpApp13 = App{
	Commands: []Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: HelpCommand,
}

const testHelpOutput13 = `
Usage: program cmd1

  Headline1
`

var testHelpCases = []struct {
	app    *App
	cmd    bool
	output string
}{
	{
		app:    &testHelpApp1,
		output: testHelpOutput1,
	},
	{
		app:    &testHelpApp2,
		output: testHelpOutput2,
	},
	{
		app:    &testHelpApp3,
		cmd:    true,
		output: testHelpOutput3,
	},
	{
		app:    &testHelpApp4,
		cmd:    true,
		output: testHelpOutput4,
	},
	{
		app:    &testHelpApp5,
		cmd:    true,
		output: testHelpOutput5,
	},
	{
		app:    &testHelpApp6,
		cmd:    true,
		output: testHelpOutput6,
	},
	{
		app:    &testHelpApp7,
		cmd:    true,
		output: testHelpOutput7,
	},
	{
		app:    &testHelpApp8,
		cmd:    true,
		output: testHelpOutput8,
	},
	{
		app:    &testHelpApp9,
		cmd:    true,
		output: testHelpOutput9,
	},
	{
		app:    &testHelpApp10,
		cmd:    true,
		output: testHelpOutput10,
	},
	{
		app:    &testHelpApp11,
		output: testHelpOutput11,
	},
	{
		app:    &testHelpApp12,
		output: testHelpOutput12,
	},
	{
		app:    &testHelpApp13,
		cmd:    true,
		output: testHelpOutput13,
	},
}

func TestHelp(t *testing.T) {
	color.NoColor = true

	for i, test := range testHelpCases {
		t.Run(fmt.Sprintf("Test %d, app: %v, cmd: %v", i, test.app, test.cmd), func(t *testing.T) {
			var cmd *Command
			if test.cmd {
				cmd = &test.app.Commands[0]
			}

			var buf bytes.Buffer
			test.app.Help(&buf, "program", cmd)
			got := buf.String()
			// Slicing test.output here to remove initial newline
			want := test.output[1:]

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
