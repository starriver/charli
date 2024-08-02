package charli

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

var testParseTemplate = App{
	GlobalOptions: []Option{
		{
			Short: 'g',
			Flag:  true,
		},
	},

	Commands: []Command{
		{
			Name: "zero",
		},
		{
			Name: "options",
			Options: []Option{
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
			Options: []Option{
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
			Options: []Option{
				{
					Long: "opt",
					Flag: true,
				},
			},
			Args: Args{
				Count:    3,
				Metavars: []string{"A", "B", "C"},
			},
		},
		{
			Name: "args3v",
			Args: Args{
				Count:    3,
				Varadic:  true,
				Metavars: []string{"A", "B", "C", "D"},
			},
		},
		{
			Name: "args0v",
			Args: Args{
				Count:    0,
				Varadic:  true,
				Metavars: []string{"A"},
			},
		},
	},
}

var testParseTemplateSingle = App{
	Commands: []Command{
		{
			Options: []Option{
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
	output     Result
	cmdName    string
	errCount   int
}{
	{
		input: []string{},
		output: Result{
			Action: HelpError,
		},
	},
	{
		input: []string{"-h"},
		output: Result{
			Action: HelpOK,
		},
	},
	{
		// Invalid command
		input: []string{"nope"},
		output: Result{
			Action: Fatal,
		},
		errCount: 1,
	},
	{
		// Invalid command
		input: []string{"nope", "-h"},
		output: Result{
			Action: HelpError,
		},
		errCount: 1,
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "-a"},
		output: Result{
			Action: HelpError,
		},
	},
	{
		input: []string{"zero"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		input: []string{"-h", "zero"},
		output: Result{
			Action: HelpOK,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "-h"},
		output: Result{
			Action: HelpOK,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "--help"},
		output: Result{
			Action: HelpOK,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "zero", "-b"},
		output: Result{
			Action: HelpError,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-h", "-b"},
		output: Result{
			Action: HelpError,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-b", "-h"},
		output: Result{
			Action: HelpError,
		},
		cmdName: "zero",
	},
	{
		input:      []string{},
		setDefault: "zero",
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		// Test the global opt
		input: []string{"zero", "-g"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {IsSet: true},
			},
		},
		cmdName: "zero",
	},
	{
		input:      []string{"-h"},
		setDefault: "zero",
		output: Result{
			Action: HelpOK,
		},
	},
	{
		input: []string{"options"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"opt": {IsSet: true},
				"g":   {},
			},
			Args: []string{"a", "b", "c"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3", "a", "b", "--", "--opt"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b", "--opt"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3", "a", "b"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
			Args: []string{"a", "b", "c", "d"},
		},
		cmdName: "args3v",
	},
	{
		input: []string{"args3v", "a", "b"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		errCount: 1,
		cmdName:  "args3v",
	},
	{
		input: []string{"args0v"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
		},
		cmdName: "args0v",
	},
	{
		input: []string{"args0v", "a", "b"},
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args0v",
	},
	{
		input:     []string{},
		useSingle: true,
		output: Result{
			Action: Proceed,

			Options: map[string]*OptionResult{
				"f":    {},
				"flag": {},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-f"},
		useSingle: true,
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
				"f":    {IsSet: true},
				"flag": {IsSet: true},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-h"},
		useSingle: true,
		output: Result{
			Action: HelpOK,
		},
		cmdName: "only",
	},
	{
		input:     []string{"only"},
		useSingle: true,
		output: Result{
			Action: Proceed,
			Options: map[string]*OptionResult{
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
		output: Result{
			Action: HelpError,
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
				got.Options = make(map[string]*OptionResult)
			}
			if want.Options == nil {
				want.Options = make(map[string]*OptionResult)
			}

			// Expect an empty (not nil) slice where Action == Proceed.
			if (want.Action == Proceed) && want.Args == nil {
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

// Special case: panicking when duplicate options configured
func TestParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	app := App{
		Commands: []Command{
			{
				Options: []Option{
					{
						Short: 'a',
					},
				},
			},
		},
		GlobalOptions: []Option{
			{
				Short: 'a',
			},
		},
	}

	app.Parse([]string{"program"})
}

// Special case: ErrorHandler provided
func TestErrorHandler(t *testing.T) {
	errs := 0
	app := App{
		Commands: []Command{
			{},
		},
		ErrorHandler: func(str string) {
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
