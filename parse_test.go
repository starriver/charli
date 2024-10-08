package charli_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/starriver/charli"
)

var testParseTemplate = charli.App{
	GlobalOptions: []charli.Option{
		{
			Short: 'g',
			Flag:  true,
		},
	},

	Commands: []charli.Command{
		{
			Name: "zero",
		},
		{
			Name: "options",
			Options: []charli.Option{
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
			Options: []charli.Option{
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
				{
					Short: 'v',
				},
			},
		},
		{
			Name: "args3",
			Options: []charli.Option{
				{
					Long: "opt",
					Flag: true,
				},
			},
			Args: charli.Args{
				Count:    3,
				Metavars: []string{"A", "B"},
			},
		},
		{
			Name: "args3v",
			Args: charli.Args{
				Count:    3,
				Varadic:  true,
				Metavars: []string{"A", "B", "C", "D"},
			},
		},
		{
			Name: "args0v",
			Args: charli.Args{
				Count:    0,
				Varadic:  true,
				Metavars: []string{"A"},
			},
		},
	},
}

var testParseTemplateSingle = charli.App{
	Commands: []charli.Command{
		{
			Options: []charli.Option{
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
	helpAccess charli.HelpAccess
	output     charli.Result
	cmdName    string
	errs       []string
	noErrFail  bool
}{
	{
		input: []string{},
		output: charli.Result{
			Action: charli.Help,
		},
		noErrFail: true,
	},
	{
		input: []string{"-h"},
		output: charli.Result{
			Action: charli.Help,
		},
	},
	{
		// Invalid command
		input: []string{"-x"},
		output: charli.Result{
			Action: charli.Fatal,
		},
		errs: []string{"no command supplied - try: `program --help`"},
	},
	{
		// Invalid command
		input:      []string{"nope"},
		helpAccess: charli.HelpFlag | charli.HelpCommand,
		output: charli.Result{
			Action: charli.Fatal,
		},
		errs: []string{"'nope' isn't a valid command - try: `program --help`"},
	},
	{
		// Invalid command
		input:      []string{"nope"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Fatal,
		},
		errs: []string{"'nope' isn't a valid command - try: `program help`"},
	},
	{
		// Invalid command
		input: []string{"nope", "-h"},
		output: charli.Result{
			Action: charli.Help,
		},
		errs: []string{"'nope' isn't a valid command."},
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "-a"},
		output: charli.Result{
			Action: charli.Help,
		},
		noErrFail: true,
	},
	{
		input: []string{"zero"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		input: []string{"-h", "zero"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "-h"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "zero",
	},
	{
		input: []string{"zero", "--help"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "zero",
	},
	{
		// Extraneous arg when asking for help
		input: []string{"-h", "zero", "-b"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName:   "zero",
		noErrFail: true,
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-h", "-b"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName:   "zero",
		noErrFail: true,
	},
	{
		// Extraneous arg when asking for help
		input: []string{"zero", "-b", "-h"},
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName:   "zero",
		noErrFail: true,
	},
	{
		input:      []string{},
		setDefault: "zero",
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
		},
		cmdName: "zero",
	},
	{
		// Test the global opt
		input: []string{"zero", "-g"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {IsSet: true},
			},
		},
		cmdName: "zero",
	},
	{
		input:      []string{"-h"},
		setDefault: "zero",
		output: charli.Result{
			Action: charli.Help,
		},
	},
	{
		input: []string{"help"},
		output: charli.Result{
			Action: charli.Fatal,
		},
		errs: []string{"'help' isn't a valid command - try: `program --help`"},
	},
	{
		input:      []string{"help"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
	},
	{
		input:      []string{"help", "help"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		noErrFail: true,
	},
	{
		input:      []string{"help", "zero"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "zero",
	},
	{
		input:      []string{"zero", "help"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "zero",
	},
	{
		input:      []string{"help", "-h"},
		helpAccess: charli.HelpFlag | charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		noErrFail: true,
	},
	{
		input:      []string{"-h", "help"},
		helpAccess: charli.HelpFlag | charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		noErrFail: true,
	},
	{
		input:      []string{"help", "zero", "-a"},
		helpAccess: charli.HelpCommand,
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName:   "zero",
		noErrFail: true,
	},
	{
		input: []string{"options"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		input: []string{"options", "--choice"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {},
				"flag":   {},
				"g":      {},
			},
		},
		cmdName: "options",
		errs:    []string{"missing value ARG for '--choice'"},
	},
	{
		input: []string{"options", "--flag", "--flag"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"long":   {},
				"c":      {},
				"choice": {},
				"f":      {IsSet: true},
				"flag":   {IsSet: true},
				"g":      {},
			},
		},
		cmdName: "options",
		errs:    []string{"duplicate option: '--flag'"},
	},
	{
		input: []string{"options", "-c=a", "--choice", "d", "-"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
	},
	{
		input: []string{"combined", "-ab", "-c"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {IsSet: true},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
	},
	{
		// -x isn't an option
		input: []string{"combined", "-abx"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
		errs:    []string{"unrecognized option '-x' in '-abx'"},
	},
	{
		input: []string{"combined", "-aba"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
		errs:    []string{"duplicate option '-a' in '-aba'"},
	},
	{
		input: []string{"combined", "-abv"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {IsSet: true},
				"b": {IsSet: true},
				"c": {},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
		errs:    []string{"can't use '-v' in combined short option '-abv'"},
	},
	{
		input: []string{"combined", "-ab=a"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"a": {},
				"b": {},
				"c": {},
				"v": {},
				"g": {},
			},
		},
		cmdName: "combined",
		errs:    []string{"combined short option can't contain '=': '-ab=a'"},
	},
	{
		input: []string{"args3", "a", "b", "c"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a"},
		},
		cmdName: "args3",
		errs:    []string{"missing arguments: B ARG"},
	},
	{
		// Too few args
		input: []string{"args3", "a", "b"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args3",
		errs:    []string{"missing argument: ARG"},
	},
	{
		input: []string{"args3", "a", "--opt", "b", "c"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"opt": {IsSet: true},
				"g":   {},
			},
			Args: []string{"a", "b", "c"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3", "a", "b", "--", "--opt"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"opt": {},
				"g":   {},
			},
			Args: []string{"a", "b", "--opt"},
		},
		cmdName: "args3",
	},
	{
		input: []string{"args3v", "a", "b", "c", "d"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b", "c", "d"},
		},
		cmdName: "args3v",
	},
	{
		input: []string{"args3v", "a", "b"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args3v",
		errs:    []string{"missing argument: C"},
	},
	{
		input: []string{"args0v"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
		},
		cmdName: "args0v",
	},
	{
		input: []string{"args0v", "a", "b"},
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"g": {},
			},
			Args: []string{"a", "b"},
		},
		cmdName: "args0v",
	},
	{
		input:     []string{},
		useSingle: true,
		output: charli.Result{
			Action: charli.Proceed,

			Options: map[string]*charli.OptionResult{
				"f":    {},
				"flag": {},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-f"},
		useSingle: true,
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
				"f":    {IsSet: true},
				"flag": {IsSet: true},
			},
		},
		cmdName: "only",
	},
	{
		input:     []string{"-h"},
		useSingle: true,
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName: "only",
	},
	{
		input:     []string{"only"},
		useSingle: true,
		output: charli.Result{
			Action: charli.Proceed,
			Options: map[string]*charli.OptionResult{
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
		output: charli.Result{
			Action: charli.Help,
		},
		cmdName:   "only",
		noErrFail: true,
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
			if len(test.errs) != 0 || test.noErrFail {
				want.Fail = true
			}

			// We don't care about testing whether these maps are initialised.
			if got.Options == nil {
				got.Options = make(map[string]*charli.OptionResult)
			}
			if want.Options == nil {
				want.Options = make(map[string]*charli.OptionResult)
			}

			// Expect an empty (not nil) slice where Action == charli.Proceed.
			if (want.Action == charli.Proceed) && want.Args == nil {
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
	dupeCmd := charli.App{
		Commands: []charli.Command{
			{
				Name: "cmd",
			},
			{
				Name: "cmd",
			},
		},
	}

	dupeOptionShort := charli.App{
		Commands: []charli.Command{
			{
				Options: []charli.Option{
					{
						Short: 'a',
					},
				},
			},
		},
		GlobalOptions: []charli.Option{
			{
				Short: 'a',
			},
		},
	}

	dupeOptionLong := charli.App{
		Commands: []charli.Command{
			{
				Options: []charli.Option{
					{
						Long: "long",
					},
				},
			},
		},
		GlobalOptions: []charli.Option{
			{
				Long: "long",
			},
		},
	}

	invalidDefaultCmd := charli.App{
		Commands: []charli.Command{
			{
				Name: "cmd1",
			},
			{
				Name: "cmd2",
			},
		},
		DefaultCommand: "cmd3",
	}

	defaultWithSingleCmd := charli.App{
		Commands: []charli.Command{
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

	t.Run("dupe option (short)", func(t *testing.T) {
		defer expectPanic(t)
		dupeOptionShort.Parse([]string{"program"})
	})

	t.Run("dupe option (long)", func(t *testing.T) {
		defer expectPanic(t)
		dupeOptionLong.Parse([]string{"program"})
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
	app := charli.App{
		Commands: []charli.Command{
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
