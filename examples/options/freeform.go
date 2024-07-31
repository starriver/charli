package main

import (
	"fmt"
	"strconv"

	cli "github.com/starriver/charli"
)

var freeform = cli.Command{
	Name:     "freeform",
	Headline: "These take any string (including an empty string)",

	Options: []cli.Option{
		{
			Short:    'n',
			Long:     "name",
			Metavar:  "NAME",
			Headline: "Tell me what to call you",
		},
		{
			Long:     "age",
			Metavar:  "YEARS",
			Headline: "Tell me how many {YEARS} old you are",
		},
	},

	Run: func(r *cli.Result) (ok bool) {
		if r.Options["name"].Value == "" {
			r.Error("I need a name...")
		}

		var age int
		if r.Options["age"].IsSet {
			var err error
			age, err = strconv.Atoi(r.Options["age"].Value)
			if err != nil || age < 0 {
				r.Errorf("%s? That ain't an age.", r.Options["age"].Value)
			}
		}

		if len(r.Errs) != 0 {
			return false
		}

		fmt.Printf("Hello %s.", r.Options["name"].Value)
		if age != 0 {
			fmt.Printf(" Blimey, %d years old?", age)
		}
		fmt.Print("\n")
		return true
	},
}
