package charli

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Prints newline-separated shell completions. i should be the index of the arg
// being completed in args.
//
// When calling this function as a result of a generated completion script,
// remember to remove the special trigger flag from args beforehand.
func Complete(w io.Writer, i int, args []string) {
	// TODO
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
