package charli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/starriver/charli"
)

// No default command, 2 subcommands
var testHelpApp1 = charli.App{
	Headline:    "Headline",
	Description: "\nDescription in {-h/--help}\n",
	Commands: []charli.Command{
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

  Description in -h/--help

Options:
  -h/--help  Show this help

Commands:
  cmd1  Headline1
  cmd2  Headline2
`

// Default command, no description, 2 subcommands
var testHelpApp2 = charli.App{
	Headline:       "Headline",
	DefaultCommand: "cmd1",
	Commands: []charli.Command{
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
var testHelpApp3 = charli.App{
	Commands: []charli.Command{
		{
			Name:        "cmd1",
			Headline:    "Headline",
			Description: "\nThis is a {description}\n",
			Options: []charli.Option{
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
var testHelpApp4 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Args: charli.Args{
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
var testHelpApp5 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Args: charli.Args{
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
var testHelpApp6 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Args: charli.Args{
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
var testHelpApp7 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Args: charli.Args{
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
var testHelpApp8 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Options: []charli.Option{
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
	GlobalOptions: []charli.Option{
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
var testHelpApp9 = charli.App{
	Commands: []charli.Command{
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
var testHelpApp10 = charli.App{
	Commands: []charli.Command{
		{},
	},
}

const testHelpOutput10 = `
Usage: program [OPTIONS]

Options:
  -h/--help  Show this help
`

// Help as command
var testHelpApp11 = charli.App{
	Description: "\nDescription\n", // This is here to test spacing
	Commands: []charli.Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: charli.HelpCommand,
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
var testHelpApp12 = charli.App{
	Description: "\nDescription\n", // This is here to test spacing
	Commands: []charli.Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: charli.HelpFlag | charli.HelpCommand,
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
var testHelpApp13 = charli.App{
	Commands: []charli.Command{
		{
			Name:     "cmd1",
			Headline: "Headline1",
		},
		{
			Name:     "cmd2",
			Headline: "Headline2",
		},
	},
	HelpAccess: charli.HelpCommand,
}

const testHelpOutput13 = `
Usage: program cmd1

  Headline1
`

// With global options but without command
var testHelpApp14 = charli.App{
	Commands: []charli.Command{
		{
			Name: "cmd1",
			Options: []charli.Option{
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
	GlobalOptions: []charli.Option{
		{
			Short:    'a',
			Flag:     true,
			Headline: "Global",
		},
	},
}

const testHelpOutput14 = `
Usage: program [OPTIONS] COMMAND [...]

Options:
  -h/--help  Show this help

Commands:
  cmd1
  cmd2
`

var testHelpCases = []struct {
	app    *charli.App
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
	{
		app:    &testHelpApp14,
		output: testHelpOutput14,
	},
}

func TestHelp(t *testing.T) {
	color.NoColor = true

	for i, test := range testHelpCases {
		t.Run(fmt.Sprintf("Test %d, app: %v, cmd: %v", i, test.app, test.cmd), func(t *testing.T) {
			var cmd *charli.Command
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
