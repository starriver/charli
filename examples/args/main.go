package main

import (
	"fmt"
	"os"

	cli "github.com/starriver/charli"
)

const description = `
This example demos {charli}'s positional args. Try running each of the commands
below.

Note that this description is written as a raw multiline string, justified at
80 characters, with a newline at each end. This is how charli expects it!
`

var app = cli.App{
	Description: description,
	Commands: []cli.Command{
		fixed,
		varadic,
	},
}

func main() {
	r := app.Parse(os.Args)

	switch r.Action {
	case cli.Proceed:
		// r.RunCommand() is exactly equivalent to r.Command.Run(&r). The
		// command's Run(...) func should provide further validation, then
		// (if everything passed) actually do the work.
		r.RunCommand()

	case cli.Help:
		// r.PrintHelp() is exactly equivalent to:
		//   r.App.Help(os.Stderr, os.Args[0], r.Command)
		r.PrintHelp()

	case cli.Fatal:
		// Fatal error, nothing else to do.
	}

	for _, err := range r.Errs {
		fmt.Fprintln(os.Stderr, err)
	}

	if r.Fail {
		os.Exit(1)
	}
}
