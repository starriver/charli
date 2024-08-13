package charli

import (
	"fmt"
	"strings"
)

// Parse parses CLI arguments, returning a [Result].
// argv should be the full list of arguments, including the executable name.
// (if in doubt, use [os.Args].)
//
// See the readme for a complete description of the syntax supported by Parse.
func (app *App) Parse(argv []string) (r Result) {
	program := argv[0]
	args := argv[1:]
	nargs := len(args)

	r.App = app

	var cmdMap map[string]*Command

	singleCmd := len(app.Commands) == 1
	if singleCmd {
		if app.DefaultCommand != "" {
			panic("Must have > 1 command when setting DefaultCommand")
		}
		r.Command = &app.Commands[0]
	} else {
		cmdMap = app.cmdMap()
	}

	ha := app.HelpAccess
	if ha == 0 {
		ha = HelpFlag
	}

	// Start by scanning for special args: -- and -h/--help/help. This is done
	// beforehand because (a) we don't want to show any other errors when
	// requesting help, (b) we need to check for -- with respect to the help
	// options so we might as well scan for it now, and (c) it's a
	// short-circuit.
	unparsedIndex := -1
	for i, arg := range args {
		if arg == "--" {
			// At this point, we want to modify the args slice. See below.
			unparsedIndex = i
			break
		}

		isHelpFlag := (ha&HelpFlag != 0) && (arg == "-h" || arg == "--help")
		isHelpCommand := (ha&HelpCommand != 0) && i < 2 && arg == "help"
		if !(isHelpFlag || isHelpCommand) {
			continue
		}

		r.Action = Help

		invalidCmd := true
		if !singleCmd && nargs > 1 {
			// Check the 1st and 2nd args for a command.
			for i := range 2 {
				// We need to disregard 'help' as a command here if necessary:
				if isOption(args[i]) || ((ha&HelpCommand != 0) && args[i] == "help") {
					continue
				}
				r.Command = cmdMap[args[i]]
				// If invalid, continue to display help anyway.
				if r.Command == nil {
					r.Error(InvalidCommandError{
						Program: program,
						Name:    args[i],
					})
				} else {
					invalidCmd = false
				}
			}
		}

		// Don't error if the user asked for help the "proper" way - with the
		// help flag, possibly a command, and nothing else.
		if !(nargs == 1 || (nargs == 2 && !invalidCmd)) {
			r.Fail = true
		}

		return
	}

	var unparsedArgs []string
	// Split off the unparsed args (if we broke out of the above loop).
	if unparsedIndex != -1 {
		unparsedArgs = args[unparsedIndex+1:]
		args = args[:unparsedIndex]
	}

	var cmdArgs []string

	if singleCmd {
		// r.Command already set.
		cmdArgs = args
	} else {
		// If we don't just have a single command, we now need to select one. If
		// a default is available, and the first arg doesn't look like a
		// command, use the default.
		possibleCommand := nargs > 0 && !isOption(args[0])
		if app.DefaultCommand != "" && !possibleCommand {
			r.Command = cmdMap[app.DefaultCommand]
			if r.Command == nil {
				panic(
					fmt.Sprintf(
						"Unknown default command '%s' configured",
						app.DefaultCommand,
					),
				)
			}
			cmdArgs = args
		} else {
			if nargs == 0 { // Implicit: app.DefaultCommand can't be set here.
				// Display help if no command or default.
				r.Action = Help
				r.Fail = true
				return
			}
			if !possibleCommand {
				// The user might've supplied flags - but no command.
				r.Error(MissingCommandError{
					Program:    program,
					HelpAccess: ha,
				})
				r.Action = Fatal
				return
			}

			r.Command = cmdMap[args[0]]
			if r.Command == nil {
				r.Error(InvalidCommandError{
					Program:     program,
					Name:        args[0],
					SuggestHelp: ha,
				})
				r.Action = Fatal
				return
			}
			cmdArgs = args[1:]
		}
	}

	// If we've reached this far, we have a valid Command and can begin
	// parsing the rest of the args within its context.

	// Start by building r.Options. Note the long and short names resolve to the
	// same struct.
	mapSizeHint := len(app.GlobalOptions)*2 + len(r.Command.Options)*2
	r.Options = make(map[string]*OptionResult, mapSizeHint)
	for _, option := range append(app.GlobalOptions, r.Command.Options...) {
		o := OptionResult{}
		o.Option = &option
		if len(option.Long) != 0 {
			_, ok := r.Options[option.Long]
			if ok {
				panic(fmt.Sprintf("Duplicate option '--%s' configured", option.Long))
			}
			r.Options[option.Long] = &o
		}
		if option.Short != 0 {
			s := string(option.Short)
			_, ok := r.Options[s]
			if ok {
				panic(fmt.Sprintf("Duplicate option '-%s' configured", s))
			}
			r.Options[s] = &o
		}
	}

	var pairedOption *OptionResult
	var pairedOptionArg string

	// This is only used twice below, but it feels just a lil too complex to
	// repeat.
	checkChoice := func(option *Option, value string, joinedArg string) bool {
		if len(option.Choices) == 0 {
			return true
		}

		for _, choice := range option.Choices {
			if value == choice {
				return true
			}
		}

		r.Error(InvalidChoiceError{
			Option:    option,
			JoinedArg: joinedArg,
			Value:     value,
		})
		return false
	}

	for _, arg := range cmdArgs {
		// Are we dealing with an arg pair?
		if pairedOption != nil {
			if !isOption(arg) {
				combinedArg := fmt.Sprintf("%s %s", pairedOptionArg, arg)
				ok := checkChoice(pairedOption.Option, arg, combinedArg)
				if ok {
					pairedOption.Value = arg
					pairedOption.IsSet = true
				}
			} else {
				r.Error(AmbiguousValueError{
					Option:    pairedOption.Option,
					OptionArg: pairedOptionArg,
					Value:     arg,
				})
			}

			pairedOption = nil
			pairedOptionArg = ""
			continue
		}

		var optionStrs []string
		var combinedShort bool
		var combinedValue string

		// Get the option name(s) out of this arg.
		if isLongOption(arg) {
			index := strings.IndexRune(arg, '=')
			if index != -1 {
				combinedValue = arg[index+1:]
				optionStrs = []string{arg[2:index]}
			} else {
				optionStrs = []string{arg[2:]}
			}
		} else if isOption(arg) {
			l := len(arg)
			if l == 1 {
				// Weird case: '-' as an option
				r.Error(InvalidOptionError{
					Arg: "-",
				})
				continue
			}

			if l > 2 {
				if strings.ContainsRune(arg, '=') {
					r.Error(CombinedEqualsError{
						Arg: arg,
					})
					continue
				}
				combinedShort = true
			}

			optionStrs = strings.Split(arg, "")[1:]
		} else {
			r.Args = append(r.Args, arg)
			continue
		}

		// Iterate through the option(s) that make up this arg. In most cases,
		// this'll just be one iteration (because this won't be combined short
		// args).
		for _, name := range optionStrs {
			o := r.Options[name]
			if o == nil {
				if combinedShort {
					r.Error(InvalidOptionError{
						Arg:         "-" + name,
						CombinedArg: arg,
					})
				} else {
					r.Error(InvalidOptionError{
						Arg: arg,
					})
				}
				continue
			}

			if o.IsSet {
				if combinedShort {
					r.Error(DuplicateOptionError{
						Option:      o,
						Arg:         "-" + name,
						CombinedArg: arg,
					})
				} else {
					r.Error(DuplicateOptionError{
						Option: o,
						Arg:    arg,
					})
				}
				continue
			}

			if o.Option.Flag {
				o.IsSet = true
			} else if combinedShort {
				r.Error(CombinedValueError{
					Option:      o.Option,
					Arg:         "-" + name,
					CombinedArg: arg,
				})
				continue
			} else if len(combinedValue) != 0 {
				ok := checkChoice(o.Option, combinedValue, arg)
				if ok {
					o.Value = combinedValue
					o.IsSet = true
				}
			} else {
				pairedOption = o
				pairedOptionArg = arg
			}
		}
	}

	if pairedOption != nil {
		metavar := pairedOption.Option.Metavar
		if metavar == "" {
			metavar = "ARG"
		}
		r.Error(MissingValueError{
			Option:  pairedOption.Option,
			Arg:     pairedOptionArg,
			Metavar: metavar,
		})
	}

	// Every option has now been processed - now we just need to validate the
	// (non-option) args count.

	// We're about to access this a lot.
	rca := &r.Command.Args

	r.Args = append(r.Args, unparsedArgs...)
	n := len(r.Args)

	if !rca.Varadic && n > rca.Count {
		r.Error(TooManyArgsError{
			Args: r.Args[rca.Count:],
		})
		r.Args = r.Args[:rca.Count]
	}

	if n < rca.Count {
		metavars := make([]string, rca.Count-n)
		for i := range len(metavars) {
			metavars[i] = "ARG"
			if n+i < len(rca.Metavars) {
				metavars[i] = rca.Metavars[n+i]
			}
		}

		r.Error(MissingArgsError{
			Metavars: metavars,
		})
	}

	// Depending on how things turned out above, Args can sometimes be a nil
	// slice. Give it an array for consistency.
	if r.Args == nil {
		r.Args = []string{}
	}

	r.Action = Proceed
	return
}

func isOption(arg string) bool {
	// Note that this returns true if this is either a short or long option.
	return strings.HasPrefix(arg, "-")
}

func isLongOption(arg string) bool {
	return strings.HasPrefix(arg, "--")
}

func suggestHelpArg(ha HelpAccess) string {
	if ha&HelpFlag == 0 {
		return "help"
	}
	return "--help"
}

// InvalidCommandError indicates the user has selected a command that doesn't
// exist.
type InvalidCommandError struct {
	Program     string     // the name of the program
	Name        string     // the name of the invalid command
	SuggestHelp HelpAccess // how to suggest CLI help is accessed
}

func (err InvalidCommandError) Error() string {
	if err.SuggestHelp == 0 {
		return fmt.Sprintf("'%s' isn't a valid command.", err.Name)
	}

	return fmt.Sprintf(
		"'%s' isn't a valid command - try: `%s %s`",
		err.Name,
		err.Program,
		suggestHelpArg(err.SuggestHelp),
	)
}

// MissingCommandError indicates that the CLI requires the user to supply a
// command, yet they didn't.
//
// This error only occurs when multiple [Command]s are configured
// and [App.DefaultCommand] is blank.
type MissingCommandError struct {
	Program    string     // the name of the program
	HelpAccess HelpAccess // how to suggest CLI help is accessed
}

func (err MissingCommandError) Error() string {
	return fmt.Sprintf(
		"no command supplied - try: `%s %s`",
		err.Program,
		suggestHelpArg(err.HelpAccess),
	)
}

// InvalidChoiceError indicates that the user has supplied an invalid choice
// as the value for an option which has [Option.Choices] set.
type InvalidChoiceError struct {
	Option    *Option // the [Option] in question
	JoinedArg string  // the argument(s) in question, which may be concatenated
	Value     string  // the invalid value
}

func (err InvalidChoiceError) Error() string {
	return fmt.Sprintf(
		"invalid '%s': must be one of [%s]",
		err.JoinedArg,
		strings.Join(err.Option.Choices, "|"),
	)
}

// AmbiguousValueError indicates that the user has supplied a value for an
// option that looks like another option itself
// (that is, the value starts with `-`).
type AmbiguousValueError struct {
	Option    *Option // the [Option] in question
	OptionArg string  // the first argument (which triggered the [Option])
	Value     string  // the second argument (which is ambiguous)
}

func (err AmbiguousValueError) Error() string {
	var s string
	s += fmt.Sprintf(
		"missing or ambiguous option value: '%s %s'\n",
		err.OptionArg, err.Value,
	)
	s += fmt.Sprintf(
		"hint: if '%s' is meant as the value for '%s', use '=' instead:\n",
		err.Value, err.OptionArg,
	)
	s += fmt.Sprintf("  %s=%s", err.OptionArg, err.Value)
	return s
}

// InvalidOptionError indicates that the user supplied an option which doesn't
// exist.
type InvalidOptionError struct {
	Arg         string // the invalid option's argument
	CombinedArg string // the combined argument it is part of (if applicable)
}

func (err InvalidOptionError) Error() string {
	if len(err.CombinedArg) != 0 {
		return fmt.Sprintf(
			"unrecognized option '%s' in '%s'",
			err.Arg,
			err.CombinedArg,
		)
	}
	return fmt.Sprintf("unrecognized option: '%s'", err.Arg)
}

// CombinedEqualsError indicates that the user attempted to use `=` in a
// combined option.
//
// Combined options may only contain flags,
// meaning that `=` can't be used to set an option's value.
type CombinedEqualsError struct {
	Arg string // the combined argument in question
}

func (err CombinedEqualsError) Error() string {
	return fmt.Sprintf("combined short option can't contain '=': '%s'", err.Arg)
}

// DuplicateOptionError indicates that the user supplied the same option more
// than once.
type DuplicateOptionError struct {
	Option      *OptionResult // the [OptionResult] set in the first instance
	Arg         string        // the argument in question
	CombinedArg string        // the combined argument it is part of (if applicable)

	// nb. OptionResult is used here so downstream can see the previous value.
}

func (err DuplicateOptionError) Error() string {
	if len(err.CombinedArg) != 0 {
		return fmt.Sprintf(
			"duplicate option '%s' in '%s'",
			err.Arg,
			err.CombinedArg,
		)
	}
	return fmt.Sprintf("duplicate option: '%s'", err.Arg)
}

// CombinedValueError indicates that the user attempted to use a non-flag option
// as part of a combined option.
//
// Combined options may only contain flags.
type CombinedValueError struct {
	Option      *Option // the [Option] in question
	Arg         string  // the option's name (as used in the combined argument)
	CombinedArg string  // the combined argument it is part of
}

func (err CombinedValueError) Error() string {
	return fmt.Sprintf(
		"can't use '%s' in combined short option '%s'",
		err.Arg,
		err.CombinedArg,
	)
}

// MissingValueError indicates that the user omitted the value for an option
// from the end of the command line.
type MissingValueError struct {
	Option  *Option // the [Option] in question
	Arg     string  // the argument that triggered the option
	Metavar string  // the option's metavar
}

func (err MissingValueError) Error() string {
	return fmt.Sprintf("missing value %s for '%s'", err.Metavar, err.Arg)
}

// TooManyArgsError indicates that the user supplied more positional arguments
// than were allowed by [Args.Count].
//
// This error only occurs when [Args.Varadic] is false.
type TooManyArgsError struct {
	Args []string // the extraneous arguments
}

func (err TooManyArgsError) Error() string {
	return fmt.Sprint("too many arguments: ", strings.Join(err.Args, ""))
}

// MissingArgsError indicates that the user didn't supply enough positional
// arguments, as specified by [Args.Count].
type MissingArgsError struct {
	Metavars []string // the metavars for the missing arguments
}

func (err MissingArgsError) Error() string {
	plural := ""
	if len(err.Metavars) > 1 {
		plural = "s"
	}

	return fmt.Sprintf(
		"missing argument%s: %s",
		plural,
		strings.Join(err.Metavars, " "),
	)
}
