package main

import (
	cli "github.com/starriver/charli"
)

const selfDescription = `
Dance to your own music. You choose to take complete responsibility for the
mood of the club.
`

var self = cli.Command{
	Name:        "self",
	Headline:    "Dance to yourself",
	Description: selfDescription,
	Options:     options,

	Run: func(r *cli.Result) bool {
		// Cmon. We're not writing logic for this nonsense.
		return len(r.Errs) == 0
	},
}
