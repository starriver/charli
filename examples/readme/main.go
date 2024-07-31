package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	cli "github.com/starriver/charli"
)

const description = `
When I go to the club I wanna hear those club classics. I will, however, need
to choose who to dance {to} and {with}.
`

var app = cli.App{
	Description: description,
	Commands: []cli.Command{
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

	ok = false
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
