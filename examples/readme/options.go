package main

// Common options.

import cli "github.com/starriver/charli"

var options = []cli.Option{
	{
		Short:    'g',
		Long:     "george",
		Flag:     true,
		Headline: "Dance with George",
	},
	{
		Short:    's',
		Long:     "sophie",
		Flag:     true,
		Headline: "Dance with Sophie",
	},
	{
		Short:    'r',
		Long:     "rewinds",
		Metavar:  "N",
		Headline: "Pull it back {N} times",
	},
	{
		Long:     "sweat",
		Flag:     true,
		Headline: "Whether to get sweaty",
	},
}
