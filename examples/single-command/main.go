package main

import (
	"fmt"
	"os"

	"github.com/starriver/charli"
)

const description = `
This example demos using a single, global command. This will always be the
case when {App.Commands} only has a single element. Notice that:

- The command doesn't need a name.
- No commands are displayed in help.

Note that this description is written as a raw multiline string, justified at
80 characters, with a newline at each end. This is how charli expects it!
`

var app = charli.App{
	Commands: []charli.Command{
		{
			Description: description,
			Options: []charli.Option{
				{
					Short:    'f',
					Long:     "flag",
					Flag:     true,
					Headline: "Set a flag",
				},
			},
			Run: func(r *charli.Result) bool {
				if len(r.Errs) != 0 {
					return false
				}

				if r.Options["f"].IsSet {
					fmt.Println("You set the flag!")
				}

				return true
			},
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
