package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/starriver/charli"
)

// This is the example shown in the screenshot in the readme.

const description = `
When I go to the club I wanna hear those club classics. I will, however, need
to choose who to dance {to} and {with}.
`

var app = charli.App{
	Description: description,
	Commands: []charli.Command{
		self,
		ag,
		hudmo,
	},
}

func main() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build info")
	}
	app.Headline = fmt.Sprintf(
		"%s v%s - club CLI toolchain\n",
		color.New(color.FgHiBlue, color.Bold).Sprint("Dancing"),
		bi.Main.Version,
	)

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
