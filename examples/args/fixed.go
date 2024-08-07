package main

import (
	"fmt"

	cli "github.com/starriver/charli"
)

const fixedDescription = `
This example requires 3 fixed args.

Try supplying {-f/--flag}, or {--}, or both.
`

var fixed = cli.Command{
	Name:        "fixed",
	Headline:    "Required args",
	Description: fixedDescription,
	Options: []cli.Option{
		{
			Short:    'f',
			Long:     "flag",
			Flag:     true,
			Headline: "Try supplying this flag mixed in with the args",
		},
	},
	Args: cli.Args{
		Count:    3,
		Metavars: []string{"ONE", "TWO", "THREE"},
	},

	Run: func(r *cli.Result) bool {
		cmd := r.Command

		if len(r.Errs) != 0 {
			return false
		}

		args := r.Args
		fmt.Print("All good! You supplied:\n")
		for i := range args {
			fmt.Printf("%s: %s", cmd.Args.Metavars[i], args[i])
		}

		return true
	},
}
