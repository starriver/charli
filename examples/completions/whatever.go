package main

import "github.com/starriver/charli"

var whatever = charli.Command{
	Name:     "whatever",
	Headline: "This command does nothing",
	Options: []charli.Option{
		{
			Short:    'f',
			Long:     "flag",
			Flag:     true,
			Headline: "This is a flag",
		},
	},
	Run: func(r *charli.Result) {
		// Nothing to do.
	},
}
