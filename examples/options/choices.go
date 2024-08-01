package main

import (
	"fmt"
	"strconv"

	cli "github.com/starriver/charli"
)

var choices = cli.Command{
	Name:     "choices",
	Headline: "Also string values, but user must pick from a list",
	Options: []cli.Option{
		{
			Short:    'n',
			Long:     "name",
			Choices:  []string{"zack", "garry", "ethel"},
			Metavar:  "NAME",
			Headline: "Imply how old you are",
		},
		{
			Long:     "age",
			Metavar:  "YEARS",
			Headline: "Or just straight up tell me",
		},
	},

	Run: func(r *cli.Result) bool {
		// Compare this with the r.Options["name"] check in freeform.go. If the
		// name is set here, it must be one of the above choices.
		if !r.Options["name"].IsSet {
			r.Error("I need a name...")
		}

		var age int
		if r.Options["age"].IsSet {
			var err error
			age, err = strconv.Atoi(r.Options["age"].Value)
			if err != nil || age < 0 {
				r.Errorf("%s? That ain't an age.", r.Options["age"].Value)
			}
		} else if r.Options["name"].IsSet {
			// Guess the age.
			switch r.Options["name"].Value {
			case "zack":
				age = 21
			case "garry":
				age = 52
			case "ethel":
				age = 83
			}
		}

		if len(r.Errs) != 0 {
			return false
		}

		fmt.Printf(
			"Hello %s. Blimey, %d years old?\n",
			r.Options["name"].Value, age,
		)
		return true
	},
}
