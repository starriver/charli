# Guide

This doc will guide you through all of charli's features, and will outline some best practices.

1. Simplest CLI possible
1. Validating your input
1. Displaying help
1. Subcommands
1. Completions

## Initial setup

We'll start with a blank Go project:

```sh
mkdir cli
cd cli
go mod init cli
touch main.go
```

Install charli:

```sh
go get github.com/starriver/charli
```

## Hello world

Now let's edit `main.go`. We'll start by declaring a `charli.App`:

```go
package main

import "github.com/starriver/charli"

var app = charli.App{

}

func main() {
	// TODO
}
```

Declare an `App` with `Command`s to configure your CLI.

```go
var app = cli.App{
	Commands: []cli.Command{
		get,
		put,
	},
}

var get = cli.Command{
	Name: "get",
	Headline: "Download some stuff",
	Options: []cli.Option{
		Short: 'o'
		Long: "output",
		Metavar: "PATH",
		Headline: "Download to {PATH}",
	},
	Run: func(r *cli.Result) bool {
		return len(r.Errs) == 0 // TODO
	},
}

var put = cli.Command{
	Name: "put",
	Headline: "Upload some stuff",
	Args: cli.Args{
		Count: 1,
		Metavars: []string{"FILE"},
	},
	Run: func(r *cli.Result) bool {
		return len(r.Errs) == 0 // TODO
	},
}
```

### Implement `Run` functions

`Command.Run(...)` functions are where you do your own validations, and – if they pass – proceed to actually do that command's work.

Here's an example for the `get` command above, which wants to download a file to the path specified by the `--output` flag.

```go
Run: func(r* cli.Result) bool {
	v := r.Options["output"].Value
	if v == "" {
		r.Error("blank path supplied")
	} else if _, err := os.Stat(v); err == nil {
		r.Error("file already exists")
	}

	if len(r.Errs) != 0 {
		return false
	}

	// TODO: actually download some stuff.

	return true
}
```

### Parsing

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

[See the examples](./examples) and [the docs](https://pkg.go.dev/github.com/starriver/charli) for more.
