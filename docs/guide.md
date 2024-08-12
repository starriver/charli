# Guide

This doc will guide you through all of charli's features, and will outline some best practices.

## Before you start…

charli isn't for everyone! Other CLI libraries give you a lot more out of the box. This library gives you a lot of control, but that control might end up costing you more work. [Remember Uncle Ben's advice.](https://en.wikipedia.org/wiki/With_great_power_comes_great_responsibility)

That said, we've tried to make this guide a quick read. Give charli a try and see what you think.

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
			Run: func(r *charli.Result) {
				fmt.Println("Hello world!")
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

Running the program should now output `Hello world!`, and you'll be able to request help with `-h/--help`. If you're using `go run`, you can just stick the flag on the end:

```sh
go run . --help
```

As for `main()`, there quite a lot to unpack here.

-   First, we check `r.Action`. This indicates what the parser thinks we should do next.
    -   `charli.Proceed` means you should proceed to call your `Command`'s `Run(...)` function. `r.RunCommand()` is shorthand for this.
    -   `charli.Help` means you should show help to the user. `r.PrintHelp()` does just that, printing help to stderr.
    -   Not shown here is `charli.Fatal`, which means you should do neither.
-   After this, we print everything in `r.Errs`.
    -   Lots of errors can be encountered during `Parse(...)`, so `r.Errs` is a slice.
-   Lastly, we exit with an error status if `r.Fail` is set.
    -   This field may be true even if `r.Errs` is empty. We'll explain why later.

---

There's one problem left with our program. Try running:

```sh
go run . --foo
```

You'll notice the output looks like this:

```
Hello world!
unrecognized option: '--foo'
```

This seems off – our command has continued to run despite the parser producing errors.

This is because our command's `Run(...)` function needs to pay attention to `r.Fail`. If it's true, we should `return` before doing any meaningful work.

So let's change it to:

```go
Run: func(r *charli.Result) {
	if r.Fail {
		return
	}

	fmt.Println("Hello world!")
},
```

It might seem counter-intuitive that the parser would tell us to call the `Run(...)` function when `r.Fail` is already true, but there's a very important reason that charli is designed this way. In the next section, we'll show you why.

## Options & flags

With charli, options (like `-f/--foo`) normally take a value:

```sh
program --foo bar
```

**Flags** are options without a value, so they're effectively boolean:

```sh
program -g
```

---

Let's add some options to our `Command`.

```go
var app = charli.App{
	Commands: []charli.Command{
		{
			Options: []charli.Option{
				{
					Short: 'n',
					Long: "name",
				},
				{
					Short: 'f',
					Long: "flag",
					Flag: true,
				}
			},

			Run: func(r *charli.Result) {
				if r.Fail {
					return
				}

				fmt.Println("Hello world!")
			},
		},
	},
}
```

Running the program again, you'll notice not much has changed. Try supplying the options:

```
$ go run . --name Calvin -f
Hello world!
```

However, using `-h/--help`, you'll notice the options are now listed:

```
$ go run . -h
Usage: main [OPTIONS] [...]

Options:
  -h/--help  Show this help
  -n/--name
  -f/--flag
```

---

Let's now work with the options. In `Run(...)`, we can use the passed `Result` to access our option values.

We can also add further input validations here. For example, we can make sure that `-n/--name` doesn't contain any spaces.

```go
Run: func(r *charli.Result) {
	name := "world"

	if r.Options["name"].IsSet {
		name = r.Options["name"].Value

		if regexp.MustCompile(`\s`).MatchString(name) {
			r.Errorf("name must be a single word: '%s'", name)
		}
	}

	if r.Fail {
		return
	}

	fmt.Printf("Hello %s!\n", name)

	if r.Options["flag"].IsSet {
		fmt.Println("You set the flag!")
	}
},
```

Running the program again, we can check our work:

```
$ go run . --name Calvin -f
Hello Calvin!
You set the flag!

$ go run . --name 'Calvin Broadus' -f
name must be a single word: 'Calvin Broadus'
```
