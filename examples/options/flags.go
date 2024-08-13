package main

import (
	"fmt"

	"github.com/starriver/charli"
)

var flags = charli.Command{
	Name:     "flags",
	Headline: "For boolean values",
	Options: []charli.Option{
		{
			Short:    'f',
			Long:     "flag",
			Flag:     true,
			Headline: "Yes, it's a flag",
		},
		{
			Long:     "red",
			Flag:     true,
			Headline: "Sorry, it ain't gonna work",
		},
	},

	Run: func(r *charli.Result) bool {
		if len(r.Errs) != 0 {
			return false
		}

		if r.Options["f"].IsSet {
			fmt.Println("You set the flag!")
		}
		if r.Options["red"].IsSet {
			fmt.Println("I'mmmm gonna go.")
		}

		return true
	},
}
