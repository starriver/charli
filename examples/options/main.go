package main

import (
	"fmt"
	"os"

	cli "github.com/starriver/charli"
)

const description = `
This example demos {charli}'s CLI options. Each command demos a different kind
of option.

Note that this description is written as a raw multiline string, justified at
80 characters, with a newline at each end. This is how charli expects it!
`

var app = cli.App{
	Headline:    "Would you like an option?",
	Description: description,
	Commands: []cli.Command{
		freeform,
		choices,
		flags,
	},
	GlobalOptions: []cli.Option{
		{
			Short:    'g',
			Long:     "global",
			Flag:     true,
			Headline: "This is a global option",
		},
	},
}

func main() {
	r := app.Parse(os.Args)

	ok := false
	switch r.Action {
	case cli.Proceed:
		// r.RunCommand() is exactly equivalent to r.Command.Run(&r). The
		// command's Run(...) func should provide further validation, then
		// (if everything passed) actually do the work.
		ok = r.RunCommand()

	case cli.HelpOK:
		// User asked for help explicitly. This isn't an error.
		r.PrintHelp()
		ok = true

	case cli.HelpError:
		// User didn't ask for help (or asked wonkily) - but display it anyway.
		// Note that r.PrintHelp() is exactly equivalent to this line:
		app.Help(os.Stderr, os.Args[0], r.Command)

	case cli.Fatal:
		// Fatal error, nothing else to do.
	}

	for _, err := range r.Errs {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	if !ok {
		os.Exit(1)
	}
}
