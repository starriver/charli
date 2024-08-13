package main

import (
	"fmt"

	"github.com/starriver/charli"
)

const varadicDescription = `
This example requires one arg, but the rest are optional. How they're formatted
above (with the )

Try supplying {-f/--flag}, or {--}, or both.
`

var varadic = charli.Command{
	Name:        "varadic",
	Headline:    "Varadic args",
	Description: varadicDescription,
	Options: []charli.Option{
		{
			Short:    'f',
			Long:     "flag",
			Flag:     true,
			Headline: "Try supplying this flag mixed in with the args",
		},
	},
	Args: charli.Args{
		Count:    1,
		Metavars: []string{"ONE", "TWO", "OTHERS"},
		Varadic:  true,
	},

	Run: func(r *charli.Result) bool {
		cmd := r.Command

		if len(r.Errs) != 0 {
			return false
		}

		argsCfg := &cmd.Args
		fmt.Print("All good! You supplied:\n")
		fmt.Printf("%s: %s", argsCfg.Metavars[0], r.Args[0])
		if len(r.Args) > 1 {
			fmt.Printf("%s: %s", argsCfg.Metavars[1], r.Args[1])
		}
		if len(r.Args) > 2 {
			fmt.Printf("%s: %v", argsCfg.Metavars[2], r.Args[2:])
		}

		return true
	},
}
