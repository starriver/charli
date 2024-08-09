package charli_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	cli "github.com/starriver/charli"
)

var testParseTemplate = cli.App{
	GlobalOptions: []cli.Option{
		{
			Short: 'g',
			Flag:  true,
		},
	},

	Commands: []cli.Command{
		{
			Name: "zero",
		},
		{
			Name: "options",
			Options: []cli.Option{
				{
					Long:    "long",
					Metavar: "LONG",
				},
				{
					Short:   'c',
					Long:    "choice",
					Choices: []string{"a", "b", "c"},
				},
				{
					Short: 'f',
					Long:  "flag",
					Flag:  true,
				},
			},
		},
		{
			Name: "combined",
			Options: []cli.Option{
				{
					Short: 'a',
					Flag:  true,
				},
				{
					Short: 'b',
					Flag:  true,
				},
				{
					Short: 'c',
					Flag:  true,
				},
			},
		},
		{
			Name: "args3",
			Options: []cli.Option{
				{
					Long: "opt",
					Flag: true,
				},
			},
			Args: cli.Args{
				Count:    3,
				Metavars: []string{"A", "B", "C"},
			},
		},
		{
			Name: "args3v",
			Args: cli.Args{
				Count:    3,
				Varadic:  true,
				Metavars: []string{"A", "B", "C", "D"},
			},
		},
		{
			Name: "args0v",
			Args: cli.Args{
				Count:    0,
				Varadic:  true,
				Metavars: []string{"A"},
			},
		},
	},
}

var testParseTemplateSingle = cli.App{
	Commands: []cli.Command{
		{
			Options: []cli.Option{
				{
					Short: 'f',
					Long:  "flag",
					Flag:  true,
				},
			},
		},
	},
}

var testParseCases = []struct {
	input      []string
	setDefault string
	useSingle  bool
	helpAccess cli.HelpAccess
	output     cli.Result
	cmdName    string
	errs       []string
}{
	{
		input: []string{},
		output: cli.Result{
			Action: cli.HelpError,
		},
	},
	{
		input: []string{"-h"},
		output: cli.Result{
			Action: cli.HelpOK,
		},
	},
	{
		// Invalid command
		input: []string{"nope"},
		output: cli.Result{
			Action: cli.Fatal,
		},
		errs: []string{"'nope' isn't a valid command - try: `program --help`"},
	},
	{
		// Invalid command
		input: []string{"nope", "-h"},
		output: cli.Result{
			Action: cli.HelpError,
		},
		errs: []string{"'nope' isn't a valid command."},
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "-a"},
		output: cli.Result{
			Action: cli.HelpError,
		},
	},
	{
		input: []string{"zero"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		input: []string{"-h", "zero"},
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "-h"},
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "--help"},
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "zero", "-b"},
		output: cli.Result{
			Action: cli.HelpError,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-h", "-b"},
		output: cli.Result{
			Action: cli.HelpError,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-b", "-h"},
		output: cli.Result{
			Action: cli.HelpError,
		},
		cmdName: "zero",
	},
	{
		input:      []string{},
		setDefault: "zero",
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		// Test the global opt
		input: []string{"zero", "-g"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {IsSet: true},
			},
		},
		cmdName: "zero",
	},
	{
		input:      []string{"-h"},
		setDefault: "zero",
		output: cli.Result{
			Action: cli.HelpOK,
		},
	},
	{
		input: []string{"help"},
		output: cli.Result{
			Action: cli.Fatal,
		},
		errs: []string{"'help' isn't a valid command - try: `program --help`"},
	},
	{
		input:      []string{"help"},
		helpAccess: cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpOK,
		},
	},
	{
		input:      []string{"help", "help"},
		helpAccess: cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpError,
		},
	},
	{
		input:      []string{"help", "zero"},
		helpAccess: cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "zero",
	},
	{
		input:      []string{"zero", "help"},
		helpAccess: cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "zero",
	},
	{
		input:      []string{"help", "-h"},
		helpAccess: cli.HelpFlag | cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpError,
		},
	},
	{
		input:      []string{"-h", "help"},
		helpAccess: cli.HelpFlag | cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpError,
		},
	},
	{
		input:      []string{"help", "zero", "-a"},
		helpAccess: cli.HelpCommand,
		output: cli.Result{
			Action: cli.HelpError,
		},
		cmdName: "zero",
	},
	{
		input: []string{"options"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {},
				"flag":   {},
				"g":      {},
			},
		},
		cmdName: "options",
	},
	{
		input: []string{"options", "--long", "ok", "--choice=a", "-f"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {Value: "ok", IsSet: true},
				"c":      {Value: "a", IsSet: true},
				"choice": {Value: "a", IsSet: true},
				"f":      {IsSet: true},
				"flag":   {IsSet: true},
				"g":      {},
			},
		},
		cmdName: "options",
	},
	{
		input: []string{"options", "-c", "a"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {},
				"c":      {Value: "a", IsSet: true},
				"choice": {Value: "a", IsSet: true},
				"f":      {},
				"flag":   {},
				"g":      {},
			},
		},
		cmdName: "options",
	},
	{
		// Nothing supplied for --long
		input: []string{"options", "--long", "--choice=a", "-f"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {IsSet: true},
				"flag":   {IsSet: true},
				"g":      {},
			},
		},
		cmdName: "options",
		errs: []string{
			"missing or ambiguous option value: '--long --choice=a'\nhint: if '--choice=a' is meant as the value for '--long', use '=' instead:\n  --long=--choice=a",
		},
	},
	{
		// Nothing supplied for --long (at end)
		input: []string{"options", "--long"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {},
				"flag":   {},
				"g":      {},
			},
		},
		cmdName: "options",
		errs:    []string{"missing value LONG for '--long'"},
	},
	{
		// -c=a is invalid, d is invalid choice, - isn't an option
		input: []string{"options", "-c=a", "--choice", "d", "-"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {},
				"flag":   {},
				"g":      {},
			},
		},
		cmdName: "options",
		errs: []string{
			"combined short option can't contain '=': '-c=a'",
			"invalid '--choice d': must be one of [a|b|c]",
			"unrecognized option: '-'",
		},
	},
	{
		input: []string{"combined", "-ab"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"g": {},
			},
		},
		cmdName: "combined",
	},
	{
		input: []string{"combined", "-ab", "-c"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {IsSet: true},
				"g": {},
			},
		},
		cmdName: "combined",
	},
	{
		// -x isn't an option
		input: []string{"combined", "-abx"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"g": {},
			},
		},
		cmdName: "combined",
		errs:    []string{"unrecognized option '-x' in '-abx'"},
	},
	{
		input: []string{"args3", "a", "b", "c"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b", "c"},
		},
		cmdName: "args3",
	},
	{
		// Too many args
		input: []string{"args3", "a", "b", "c", "d"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b", "c"},
		},
		cmdName: "args3",
		errs:    []string{"too many arguments: d"},
	},
	{
		// Too few args
		input: []string{"args3", "a"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a"},
		},
		cmdName: "args3",
		errs:    []string{"missing arguments: B C"},
	},
	{
		// Too few args
		input: []string{"args3", "a", "b"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args3",
		errs:    []string{"missing argument: C"},
	},
	{
		input: []string{"args3", "a", "--opt", "b", "c"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {IsSet: true},
				"g":   {},
			},
			Args: []string{"a", "b", "c"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3", "a", "b", "--", "--opt"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b", "--opt"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3v", "a", "b", "c", "d"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b", "c", "d"},
		},
		cmdName: "args3v",
	},
	{
		input: []string{"args3v", "a", "b"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args3v",
		errs:    []string{"missing argument: C"},
	},
	{
		input: []string{"args0v"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
		},
		cmdName: "args0v",
	},
	{
		input: []string{"args0v", "a", "b"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args0v",
	},
	{
		input:     []string{},
		useSingle: true,
		output: cli.Result{
			Action: cli.Proceed,

			Options: map[string]*cli.OptionResult{
				"f":    {},
				"flag": {},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-f"},
		useSingle: true,
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"f":    {IsSet: true},
				"flag": {IsSet: true},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-h"},
		useSingle: true,
		output: cli.Result{
			Action: cli.HelpOK,
		},
		cmdName: "only",
	},
	{
		input:     []string{"only"},
		useSingle: true,
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"f":    {},
				"flag": {},
			},
		},
		cmdName: "only",
		errs:    []string{"too many arguments: only"},
	},
	{
		input:     []string{"-h", "only"},
		useSingle: true,
		output: cli.Result{
			Action: cli.HelpError,
		},
		cmdName: "only",
	},
}

func TestParse(t *testing.T) {
	deep.CompareUnexportedFields = true

	for i, test := range testParseCases {
		t.Run(fmt.Sprintf("Test %d, %v", i, test.input), func(t *testing.T) {
			app := testParseTemplate
			if test.useSingle {
				app = testParseTemplateSingle
			}
			input := append([]string{"program"}, test.input...)
			if test.setDefault != "" {
				app.DefaultCommand = test.setDefault
			}
			app.HelpAccess = test.helpAccess

			got := app.Parse(input)
			want := test.output

			// Patch in some fields.
			want.App = &app
			if test.cmdName != "" {
				if test.useSingle {
					want.Command = &app.Commands[0]
				} else {
					for _, cmd := range app.Commands {
						if cmd.Name == test.cmdName {
							want.Command = &cmd
							break
						}
					}
				}
			}
			if len(test.errs) != 0 {
				want.Fail = true
			}

			// We don't care about testing whether these maps are initialised.
			if got.Options == nil {
				got.Options = make(map[string]*cli.OptionResult)
			}
			if want.Options == nil {
				want.Options = make(map[string]*cli.OptionResult)
			}

			// Expect an empty (not nil) slice where Action == cli.Proceed.
			if (want.Action == cli.Proceed) && want.Args == nil {
				want.Args = []string{}
			}

			// We test whether options are correctly mapped implicitly, so blank
			// out the pointers.
			for _, o := range got.Options {
				o.Option = nil
			}

			if test.errs == nil {
				test.errs = []string{}
			}
			gotErrStrings := make([]string, len(got.Errs))
			for i, err := range got.Errs {
				gotErrStrings[i] = err.Error()
			}
			if diff := deep.Equal(gotErrStrings, test.errs); diff != nil {
				t.Error(diff)
			}
			got.Errs = nil

			if diff := deep.Equal(got, want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

// Special cases that panic
func TestParsePanic(t *testing.T) {
	dupeCmd := cli.App{
		Commands: []cli.Command{
			{
				Name: "cmd",
			},
			{
				Name: "cmd",
			},
		},
	}

	dupeOption := cli.App{
		Commands: []cli.Command{
			{
				Options: []cli.Option{
					{
						Short: 'a',
					},
				},
			},
		},
		GlobalOptions: []cli.Option{
			{
				Short: 'a',
			},
		},
	}

	invalidDefaultCmd := cli.App{
		Commands: []cli.Command{
			{
				Name: "cmd1",
			},
			{
				Name: "cmd2",
			},
		},
		DefaultCommand: "cmd3",
	}

	defaultWithSingleCmd := cli.App{
		Commands: []cli.Command{
			{
				Name: "cmd1",
			},
		},
		DefaultCommand: "cmd1",
	}

	expectPanic := func(t *testing.T) {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}

	t.Run("dupe command", func(t *testing.T) {
		defer expectPanic(t)
		dupeCmd.Parse([]string{"program"})
	})

	t.Run("dupe option", func(t *testing.T) {
		defer expectPanic(t)
		dupeOption.Parse([]string{"program"})
	})

	t.Run("invalid default command", func(t *testing.T) {
		defer expectPanic(t)
		invalidDefaultCmd.Parse([]string{"program"})
	})

	t.Run("default command with single command", func(t *testing.T) {
		defer expectPanic(t)
		defaultWithSingleCmd.Parse([]string{"program"})
	})
}

// Special case: ErrorHandler provided
func TestErrorHandler(t *testing.T) {
	errs := 0
	app := cli.App{
		Commands: []cli.Command{
			{},
		},
		ErrorHandler: func(error) {
			errs += 1
		},
	}

	r := app.Parse([]string{"program", "--nope"})

	if errs != 1 {
		t.Errorf("got %d errors, want 1", errs)
	}
	if len(r.Errs) != 0 {
		t.Error("r.Errs should be empty")
	}
}
