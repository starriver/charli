# c~~har~~li

A small, pure(-ish) CLI parser and help formatting library.

- Reads your CLI configuration (options, flags etc.) from structs.
- Checks input syntax and lists errors.
- Captures valid option & positional argument values.
- Renders usage help (to a string, if you'd like).
- Only provides very basic validation. Bring your own!

[See the code](./examples/readme/) for the below screenshot.

![Screenshot](./.images/example.png)

## Quickstart

To install:

```sh
go get github.com/starriver/charli
```

Declare an `App` with `Command`s:

```go
var app = cli.App{
	Commands: []cli.Command{
		{
			Name: "get",
			Headline: "Download some stuff",
			Options: []cli.Option{
				Short: 'o'
				Long: "output",
				Metavar: "PATH",
				Headline: "Download to {PATH}",
			},
			Run: func(r *cli.Result) (ok bool) {
				return len(r.Errs) == 0 // TODO
			},
		},
		{
			Name: "put",
			Headline: "Upload some stuff",
			Args: cli.Args{
				Count: 1,
				Metavars: []string{"FILE"},
			},
			Run: func(r *cli.Result) (ok bool) {
				return len(r.Errs) == 0 // TODO
			},
		},
	},
}
```

Parse your args, then handle the result however you'd like:

```go
package main

import (
	"fmt"
	"os"

	cli "github.com/starriver/charli"
)

func main() {
	r := app.Parse(os.Args)

	ok := false
	switch r.Action {
	case cli.Proceed:
		ok = r.RunCommand()
	case cli.HelpOK:
		r.PrintHelp()
		ok = true
	case cli.HelpError:
		r.PrintHelp()
	case cli.Fatal:
		// Nothing to do.
	}

	for _, err := range r.Errs {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	if !ok {
		os.Exit(1)
	}
}
```

[See the examples](./examples) for more.

## Design

charli provides a very lightweight, 'toolkit' approach to CLI development.

If you've used [urfave/cli](https://charli.urfave.org/), you might notice the config structs (`App`, `Command` etc.) are similar, but that's about it. charli is functionally pure first and foremost: its workhorse functions, `App.Parse(...)` and `App.Help(...)`, have no side-effects.

Some I/O helpers are provided (eg. `Result.PrintHelp()`), but aside from parsing and help generation, the way your app responds to a parse `Result` – or a help string – is entirely up to you.

**Why did we make this?** Well, we're very picky about how we want our CLIs to look and behave – in particular, we want to engineer complex, imperative flows for validation. The amount of hacking required on other libraries wasn't worth it for us, so we made this instead.

### Goals

- **Provide only basic validation.**
	- Syntax checking only.
	- No transformation for values – only strings (and bools, in the case of flags).
	- This is to provide full control over the validation process downstream.
- **Produce as many errors as possible.**
	- Buffer errors. Downstream can decide how to deal with them.
	- Don't give up after encountering one parse error. Keep going!
	- Allow downstream validations to continue even with parse errors.
	- However: make downstream validations aware of previous errors, so that expensive operations can be short-circuited.
- **Render our preferred help format.**
	- We like what we like, we hate what we hate (but we're [oh so easily swayed](https://www.youtube.com/watch?v=7Z5kEqRFPwo)).
	- Also, make colors optional ([fatih/color](https://github.com/fatih/color) allows turning them off).

## License

[![ISC](./.images/license.jpg)](./LICENSE)
