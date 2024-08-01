package main

import (
	"fmt"
	"os"

	cli "github.com/starriver/charli"
)

const description = `
This example demos using a single, global command. This will always be the
case when {App.Commands} only has a single element. Notice that:

- The command doesn't need a name.
- No commands are displayed in help.

Note that this description is written as a raw multiline string, justified at
80 characters, with a newline at each end. This is how charli expects it!
`

var app = cli.App{
	Commands: []cli.Command{
		{
			Description: description,
			Options: []cli.Option{
				{
					Short:    'f',
					Long:     "flag",
					Flag:     true,
					Headline: "Set a flag",
				},
			},
			Run: func(r *cli.Result) bool {
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
		fmt.Fprint(os.Stderr, app.Help(os.Args[0], r.Command))

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
