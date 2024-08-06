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

	if len(app.Commands) == 1 {
		for _, opt := range app.Commands[0].Options {
			fmt.Fprint(w, prefix)
			describeOpt(&opt)
			fmt.Fprint(w, "\n")
		}
		return
	}

	for _, cmd := range app.Commands {
		fmt.Fprintf(
			w,
			"%s -n __fish_cmdname_needs_subcommand -a %s\n",
			prefix,
			quote(cmd.Name),
		)

		// Yes this is horrifically ugly.
		optPrefix := fmt.Sprintf(" -n %s", quote(
			fmt.Sprintf("__fish_cmdname_using_subcommand %s", cmd.Name),
		))

		for _, opt := range cmd.Options {
			fmt.Fprint(w, prefix, optPrefix)
			describeOpt(&opt)
			fmt.Fprint(w, "\n")
		}
	}
}
