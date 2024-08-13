package main

import (
	"github.com/starriver/charli"
)

const agDescription = `
Dance to A.G. Cook's music. You may choose to have him {-p/--produce} for you.
`

var ag = charli.Command{
	Name:        "ag",
	Headline:    "Dance to A.G.",
	Description: agDescription,
	Options: append(options, charli.Option{
		Short:    'f',
		Long:     "feature",
		Flag:     true,
		Headline: "Feature on the track",
	}, charli.Option{
		Short:    'p',
		Long:     "produce",
		Flag:     true,
		Headline: "Have A.G. produce your track",
	}),

	Run: func(r *charli.Result) {
		// Cmon. We're not writing logic for this nonsense.
	},
}
