package charli

import (
	"fmt"
	"io"
	"strings"
)

func quote(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

func (app *App) CompleteFish(w io.Writer, program string) {
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
