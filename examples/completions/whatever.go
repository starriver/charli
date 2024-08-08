package main

import cli "github.com/starriver/charli"

var whatever = cli.Command{
	Name:     "whatever",
	Headline: "This command does nothing",
	Options: []cli.Option{
		{
			Short:    'f',
			Long:     "flag",
			Flag:     true,
			Headline: "This is a flag",
		},
	},
	Run: func(r *cli.Result) bool {
		return !r.Fail
	},
}
