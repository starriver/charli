package charli

import (
	"fmt"
	"strings"
)

// Parse argv and populate the user-specified command's values. argv must be the
// full list of args, including the executable name (you probably want os.Args).
//
// Returns a Result struct, populated with its findings.
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
		if nargs == 1 || (nargs == 2 && !invalidCmd) {
			r.Action = HelpOK
		} else {
			r.Action = HelpError
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
				r.Action = HelpError
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
		metavars := rca.Metavars[n:]
		if rca.Varadic {
			metavars = rca.Metavars[n:rca.Count]
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

type InvalidCommandError struct {
	Program     string
	Name        string
	SuggestHelp HelpAccess
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

type MissingCommandError struct {
	Program    string
	HelpAccess HelpAccess
}

func (err MissingCommandError) Error() string {
	return fmt.Sprintf(
		"no command supplied - try: `%s %s`",
		err.Program,
		suggestHelpArg(err.HelpAccess),
	)
}

type InvalidChoiceError struct {
	Option    *Option
	JoinedArg string
	Value     string
}

func (err InvalidChoiceError) Error() string {
	return fmt.Sprintf(
		"invalid '%s': must be one of [%s]",
		err.JoinedArg,
		strings.Join(err.Option.Choices, "|"),
	)
}

type AmbiguousValueError struct {
	Option    *Option
	OptionArg string
	Value     string
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

type InvalidOptionError struct {
	Arg         string
	CombinedArg string
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

type CombinedEqualsError struct {
	Arg string
}

func (err CombinedEqualsError) Error() string {
	return fmt.Sprintf("combined short option can't contain '=': '%s'", err.Arg)
}

type DuplicateOptionError struct {
	// nb. OptionResult here so downstream can see previous value
	Option      *OptionResult
	Arg         string
	CombinedArg string
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

type CombinedValueError struct {
	Option      *Option
	Arg         string
	CombinedArg string
}

func (err CombinedValueError) Error() string {
	return fmt.Sprintf(
		"can't use '%s' in combined short option '%s'",
		err.Arg,
		err.CombinedArg,
	)
}

type MissingValueError struct {
	Option  *Option
	Arg     string
	Metavar string
}

func (err MissingValueError) Error() string {
	return fmt.Sprintf("missing value %s for '%s'", err.Metavar, err.Arg)
}

type TooManyArgsError struct {
	Args []string
}

func (err TooManyArgsError) Error() string {
	return fmt.Sprint("too many arguments: ", strings.Join(err.Args, ""))
}

type MissingArgsError struct {
	Metavars []string
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
