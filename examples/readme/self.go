package main

import (
	"github.com/starriver/charli"
)

const selfDescription = `
Dance to your own music. You choose to take complete responsibility for the
mood of the club.
`

var self = charli.Command{
	Name:        "self",
	Headline:    "Dance to yourself",
	Description: selfDescription,
	Options:     options,

	Run: func(r *charli.Result) {
		// Cmon. We're not writing logic for this nonsense.
	},
}
