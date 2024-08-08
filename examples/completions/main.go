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

	ok := false
	switch r.Action {
	case cli.Proceed:
		ok = r.RunCommand()

	case cli.HelpOK:
		r.PrintHelp()
		ok = true

	case cli.HelpError:
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
