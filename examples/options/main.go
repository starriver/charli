package main

import (
	"fmt"
	"os"

	"github.com/starriver/charli"
)

const description = `
This example demos {charli}'s CLI options. Each command demos a different kind
of option.

Note that this description is written as a raw multiline string, justified at
80 characters, with a newline at each end. This is how charli expects it!
`

var app = charli.App{
	Headline:    "Would you like an option?",
	Description: description,
	Commands: []charli.Command{
		freeform,
		choices,
		flags,
	},
	GlobalOptions: []charli.Option{
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

	switch r.Action {
	case charli.Proceed:
		// r.RunCommand() is exactly equivalent to r.Command.Run(&r). The
		// command's Run(...) func should provide further validation, then
		// (if everything passed) actually do the work.
		r.RunCommand()

	case charli.Help:
		// r.PrintHelp() is exactly equivalent to:
		//   r.App.Help(os.Stderr, os.Args[0], r.Command)
		r.PrintHelp()

	case charli.Fatal:
		// Fatal error, nothing else to do.
	}

	for _, err := range r.Errs {
		fmt.Fprintln(os.Stderr, err)
	}

	if r.Fail {
		os.Exit(1)
	}
}
