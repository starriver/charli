package main

import (
	"github.com/starriver/charli"
)

const hudmoDescription = `
We can go high we can go high we can go high we can go high we can go high we
can go high we can go high we can go high we can go high we can go high we can
go high we can go high we can go high we can go high we can go higher yeah.
`

var hudmo = charli.Command{
	Name:        "hudmo",
	Headline:    "We can go high we can go high",
	Description: hudmoDescription,
	Options:     options,

	Run: func(r *charli.Result) {
		// Cmon. We're not writing logic for this nonsense.
	},
}
