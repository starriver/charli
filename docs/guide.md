# Guide

This doc will guide you through all of charli's features, and will outline some best practices.

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

Let's start by making a CLI that just says hello.

Edit `main.go`. We'll start by declaring an `App`:

```go
package main

import (
	"fmt"

	"github.com/starriver/charli"
)

var app = charli.App{
	Commands: []charli.Command{
		{
			Run: func(r *charli.Result) bool {
				fmt.Println("Hello world!")
				return true
			},
		},
	},
}

func main() {
	// TODO
}
```

Running this program (predictably) won't do anything yet.

We can parse the program's args in `main()` like so:

```go
func main() {
	app.Parse(os.Args)
}
```

…but again, this won't actually do anything. `Parse(...)` returns a `Result` struct which we should respond to.

Let's add some logic to `main()` to deal with the `Result`:

```go
func main() {
	r := app.Parse(os.Args)

	switch r.Action {
	case charli.Proceed:
		r.RunCommand()
	case charli.Help:
		r.PrintHelp()
	}

	for _, err := range r.Errs {
		fmt.Fprintln(os.Stderr, err)
	}

	if r.Fail {
		os.Exit(1)
	}
}
```

Running the program should now output `Hello world!`, but there's quite a lot to unpack here.

-
