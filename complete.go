package charli

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Prints newline-separated completions. i should be the index of the arg being
// completed in argv (as requested by the shell) - note that this number should
// be 1 less than the index of the arg being completed, because of the special
// trigger flag.
func (app *App) Complete(w io.Writer, i int, argv []string) {
	if len(argv) < 3 {
		panic("argv appears truncated")
	}
	args := argv[2:]
	nargs := len(args)
	i -= 2

	if i < 0 || i >= nargs {
		panic("Completion index out of bounds")
	}

	// First ensure the arg being completed isn't beyond a --.
	for j := 0; j < i; j++ {
		if args[j] == "--" {
			return
		}
	}

	cur := args[i]
	prev := ""
	if i != 0 {
		prev = args[i-1]
	}

	helpFlags := []string{}
	if app.hasHelpFlags() {
		helpFlags = []string{"-h", "--help"}
	}

	singleCmd := len(app.Commands) == 1

	var cmd *Command
	helpFirst := false

	if app.hasHelpFlags() && (args[0] == "-h" || args[0] == "--help") {
		helpFirst = true
	}
	if app.hasHelpCommand() && args[0] != "help" {
		helpFirst = true
	}

	// Can we complete a command (or -h/--help)?
	if i == 0 || (helpFirst && i == 1) {
		if !singleCmd {
			for _, cmd := range app.Commands {
				if strings.HasPrefix(cmd.Name, cur) {
					fmt.Fprintln(w, cmd.Name)
				}
			}
		}

		if i == 0 {
			for _, f := range helpFlags {
				if strings.HasPrefix(f, cur) {
					fmt.Fprintln(w, f)
				}
			}
			if app.hasHelpCommand() && strings.HasPrefix("help", cur) {
				fmt.Fprintln(w, "help")
			}
		}
		return
	}

	if singleCmd {
		cmd = &app.Commands[0]
	} else {
		cmdArg := args[0]
		if app.DefaultCommand != "" && (args[0] == "" || isOption(args[0])) {
			cmdArg = app.DefaultCommand
		}

		for _, c := range app.Commands {
			if c.Name == cmdArg {
				cmd = &c
				break
			}
		}
		if cmd == nil {
			return
		}
	}

	// Can we complete choices?
	if isOption(prev) && !strings.ContainsRune(prev, '=') {
		var opt *Option
		if isLongOption(prev) {
			long := prev[2:]
			for _, o := range cmd.Options {
				if long == o.Long {
					opt = &o
					break
				}
			}
		} else if len(prev) == 2 { // Ignore combined options
			short := rune(prev[1])
			for _, o := range cmd.Options {
				if short == o.Short {
					opt = &o
					break
				}
			}
		}

		if opt != nil && len(opt.Choices) != 0 {
			for _, c := range opt.Choices {
				if strings.HasPrefix(c, args[i]) {
					fmt.Fprintln(w, c)
				}
			}
			return
		}
	}

	// Lastly, just complete options.
	opts := cmd.Options
	if app.hasHelpFlags() {
		opts = append(opts, fakeHelpOption)
	}
	for _, opt := range opts {
		if opt.Short != 0 {
			short := "-" + string(opt.Short)
			if strings.HasPrefix(short, cur) {
				fmt.Fprintln(w, short)
			}
		}
		if opt.Long != "" {
			long := "--" + opt.Long
			if strings.HasPrefix(long, cur) {
				fmt.Fprintln(w, long)
			}
		}
	}
}

func quote(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

// Generate a bash completion script.
//
// program should be the program name (which will presumably be in the user's
// PATH). flag should be a special trigger flag, *including* hyphen prefixes,
// which your program should use to bypass normal execution and generate
// completions instead (presumably using Complete(...)).
//
// flag can be anything you want, but don't use anything ambiguous to your CLI.
// If in doubt, use "--_complete".
func (app *App) GenerateBashCompletions(w io.Writer, program, flag string) {
	program = quote(program)

	// We want a valid name for the completion function, so strip illegal
	// characters from the program name.
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	funcName := fmt.Sprintf(
		"_complete_%s",
		re.ReplaceAllString(program, ""),
	)

	// Write a function that calls the program with the required completion
	// data.
	fmt.Fprintf(w, "%s() {\n", funcName)
	fmt.Fprintf(
		w,
		"\tfor c in $(%s %s $COMP_CWORD $COMP_WORDS); do\n",
		program,
		flag,
	)
	fmt.Fprintln(w, "\t\tCOMPREPLY+=(\"$c\")")
	fmt.Fprintln(w, "\tdone")
	fmt.Fprintln(w, "}")

	// Complete using the function.
	fmt.Fprintf(
		w,
		"complete -o bashdefault -F %s %s\n",
		funcName,
		program,
	)
}

// Generate Fish completions. These are pretty comprehensive and often don't
// need to call the binary at all.
func (app *App) GenerateFishCompletions(w io.Writer, program string) {
	prefix := fmt.Sprintf("complete -c %s -k", quote(program))

	describeCmd := func(cmd *Command) {
		fmt.Fprintf(
			w,
			"%s -n __fish_cmdname_needs_subcommand -a %s",
			prefix,
			quote(cmd.Name),
		)
		if len(cmd.Headline) != 0 {
			fmt.Fprintf(w, " -d %s", quote(cmd.Headline))
		}
		fmt.Fprint(w, "\n")
	}

	// Note that this only provides the complete flags without the prefix. Needs
	// to be flexible enough for both single- and multi-command operation (see
	// below).
	describeOpt := func(opt *Option) {
		if opt.Short != 0 {
			fmt.Fprintf(w, " -s %s", quote(string(opt.Short)))
		}
		if opt.Long != "" {
			fmt.Fprintf(w, " -l %s", quote(opt.Long))
		}
		if opt.Headline != "" {
			fmt.Fprintf(w, " -d %s", quote(opt.Headline))
		}

		if opt.Flag {
			fmt.Fprint(w, " -f")
		} else {
			fmt.Fprint(w, " -r")
		}

		if len(opt.Choices) != 0 {
			fmt.Fprintf(w, " -f -a %s", quote(strings.Join(opt.Choices, " ")))
		}
	}

	if (app.HelpAccess & HelpCommand) != 0 {
		describeCmd(&fakeHelpCmd)
	}

	var withHelpOption []Option
	if (app.HelpAccess & HelpFlag) != 0 {
		withHelpOption = []Option{fakeHelpOption}
	}

	if len(app.Commands) == 1 {
		for _, opt := range append(withHelpOption, app.Commands[0].Options...) {
			fmt.Fprint(w, prefix)
			describeOpt(&opt)
			fmt.Fprint(w, "\n")
		}
		return
	}

	for _, cmd := range app.Commands {
		describeCmd(&cmd)

		// Yes this is horrifically ugly.
		optPrefix := fmt.Sprintf(" -n %s", quote(
			fmt.Sprintf("__fish_cmdname_using_subcommand %s", cmd.Name),
		))

		for _, opt := range append(withHelpOption, cmd.Options...) {
			fmt.Fprint(w, prefix, optPrefix)
			describeOpt(&opt)
			fmt.Fprint(w, "\n")
		}
	}
}
