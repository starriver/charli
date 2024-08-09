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
					Long: "long",
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
	errCount   int
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
		errCount: 1,
	},
	{
		// Invalid command
		input: []string{"nope", "-h"},
		output: cli.Result{
			Action: cli.HelpError,
		},
		errCount: 1,
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
		errCount: 1,
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
		errCount: 1,
		cmdName:  "options",
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
		errCount: 1,
		cmdName:  "options",
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
		errCount: 3,
		cmdName:  "options",
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
		errCount: 1,
		cmdName:  "combined",
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
		errCount: 1,
		cmdName:  "args3",
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
		errCount: 1,
		cmdName:  "args3",
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
		input: []string{"args3", "a", "b"},
		output: cli.Result{
			Action: cli.Proceed,
			Options: map[string]*cli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b"},
		},
		errCount: 1,
		cmdName:  "args3",
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
		errCount: 1,
		cmdName:  "args3v",
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
		cmdName:  "only",
		errCount: 1,
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
			if test.errCount != 0 {
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

			// TODO: a more proper error test. Counting them will do for now.
			if len(got.Errs) != test.errCount {
				t.Errorf(
					"error count: got %v, want %v\nerrors: %v",
					len(got.Errs),
					test.errCount,
					got.Errs,
				)
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
