package main

import (
	cli "github.com/starriver/charli"
)

const agDescription = `
Dance to A.G. Cook's music. You may choose to have him {-p/--produce} for you.
`

var ag = cli.Command{
	Name:        "ag",
	Headline:    "Dance to A.G.",
	Description: agDescription,
	Options: append(options, cli.Option{
		Short:    'f',
		Long:     "feature",
		Flag:     true,
		Headline: "Feature on the track",
	}, cli.Option{
		Short:    'p',
		Long:     "produce",
		Flag:     true,
		Headline: "Have A.G. produce your track",
	}),

	Run: func(r *cli.Result) bool {
		// Cmon. We're not writing logic for this nonsense.
		return len(r.Errs) == 0
	},
}
