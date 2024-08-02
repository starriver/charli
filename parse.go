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

	// NOTE: not bothering with a map for command names, because we'll only need
	// to iterate through Commands a maximum of twice per parse. With such a
	// (likely) tiny array, it'd likely be more expensive to set up the map in
	// the first place.
	findCommand := func(str string) *Command {
		for _, cmd := range app.Commands {
			if cmd.Name == str {
				return &cmd
			}
		}
		return nil
	}

	singleCmd := len(app.Commands) == 1
	if singleCmd {
		r.Command = &app.Commands[0]
	}

	// Start by scanning for special args: -- and -h/--help. This is done
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

		if !(arg == "-h" || arg == "--help") {
			continue
		}

		// Show command help if user runs 'program -h command [...]' or
		// 'program command -h [...]' - ie. the 1st or 2nd flag mustn't look
		// like an option.
		invalidCmd := true
		if !singleCmd && nargs > 1 {
			for i := range 2 {
				if isOption(args[i]) {
					continue
				}
				r.Command = findCommand(args[i])
				// If invalid, continue to display help anyway.
				if r.Command == nil {
					r.Errorf("'%s' isn't a valid command.", args[i])
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
			r.Command = findCommand(app.DefaultCommand)
			cmdArgs = args
		} else {
			if nargs == 0 { // Implicit: app.DefaultCommand can't be set here.
				// Display help if no command or default.
				r.Action = HelpError
				return
			}
			if !possibleCommand {
				// The user might've supplied flags - but no command.
				r.Errorf("no command supplied - try `%s --help`", program)
				r.Action = Fatal
				return
			}

			r.Command = findCommand(args[0])
			if r.Command == nil {
				r.Errorf(
					"'%s' isn't a command - try `%s --help`",
					args[0], program,
				)
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
	checkChoice := func(option *Option, value string, combinedArg string) bool {
		if len(option.Choices) == 0 {
			return true
		}

		for _, choice := range option.Choices {
			if value == choice {
				return true
			}
		}

		r.Errorf(
			"invalid '%s': must be one of [%s]",
			combinedArg,
			strings.Join(option.Choices, "|"),
		)
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
				// TODO: optimise this.
				var e string
				e += fmt.Sprintf(
					"missing or ambiguous option value: '%s %s'\n",
					pairedOptionArg, arg,
				)
				e += fmt.Sprintf(
					"hint: if '%s' is meant as the value for '%s', use '=' instead:\n",
					arg, pairedOptionArg,
				)
				e += fmt.Sprintf("  '%s=%s'", pairedOptionArg, arg)
				r.Errorf("%s", e)
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
				r.Errorf("unrecognized option: '-'")
				continue
			}

			if l > 2 {
				if strings.ContainsRune(arg, '=') {
					r.Errorf("combined short option can't contain '=': '%s'", arg)
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
					r.Errorf("unrecognized option '%s' in '%s'", name, arg)
				} else {
					r.Errorf("unrecognized option: '%s'", arg)
				}
				continue
			}

			if o.IsSet {
				if combinedShort {
					r.Errorf("duplicate option '%s' in '%s'", name, arg)
				} else {
					r.Errorf("duplicate option: '%s'", arg)
				}
				continue
			}

			if o.Option.Flag {
				o.IsSet = true
			} else if combinedShort {
				r.Errorf("can't use '%s' in combined short option '%s'", name, arg)
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
		r.Errorf("missing value %s for '%s'", metavar, pairedOptionArg)
	}

	// Every option has now been processed - now we just need to validate the
	// (non-option) args count.

	// We're about to access this a lot.
	rca := &r.Command.Args

	r.Args = append(r.Args, unparsedArgs...)
	n := len(r.Args)

	if !rca.Varadic && n > rca.Count {
		r.Errorf("too many arguments: '%s'", strings.Join(r.Args, "' '"))
		r.Args = r.Args[:rca.Count]
	}

	if n < rca.Count {
		plural := ""
		if rca.Count-n > 1 {
			plural = "s"
		}

		metavars := strings.Join(rca.Metavars[n:], " ")

		r.Errorf("missing argument%s: %s", plural, metavars)
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
