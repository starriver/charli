package main

import (
	"fmt"
	"os"

	cli "github.com/starriver/charli"
)

const description = `
This example demos {charli}'s completions and script generation.

To install the completions for bash and fish, run the {install} command.
`

var app = cli.App{
	Description: description,
	Commands: []cli.Command{
		install,
		whatever,
	},
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--_complete" {
		app.Complete(os.Stdout, os.Args)
		return
	}

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
