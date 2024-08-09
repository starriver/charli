# c~~har~~li

[![Go Reference](https://pkg.go.dev/badge/github.com/starriver/charli.svg)](https://pkg.go.dev/github.com/starriver/charli)
[![Go Report Card](https://goreportcard.com/badge/github.com/starriver/charli)](https://goreportcard.com/report/github.com/starriver/charli)
[![Fantano Rating](https://img.shields.io/badge/fantano-10-purple
)](https://youtu.be/bLJ-zfBmChA)
[![Coverage Status](https://coveralls.io/repos/github/starriver/charli/badge.svg?branch=main)](https://coveralls.io/github/starriver/charli?branch=main)

A freestyle CLI toolkit. Here's how it looks:

![Screenshot](./.images/example.png)

charli includes a **CLI parser**, **help formatter**, and **completer** for bash & fish.

[See the code](./examples/readme/) for the above screenshot.

## Quickstart

To install:

```sh
go get github.com/starriver/charli
```

- [See the guide](./docs/tutorial.md) for usage instructions.
- [Examples](./examples) are also available.

## Who's this for?

Use charli if you want to:

- **Configure your CLI with struct data.** It doesn't use the builder pattern, struct tags or reflection.
- **Bring your own input validation.** The parser outputs a map of known options & arguments according to your configuration. It also aggregates errors caused by unknown args and bad syntax. Nothing else is transformed.
- **Have complete control over your app's I/O**. Expect no magic or surprises! None of the core functions have any side-effects.

## Design

We wrote this because very picky about how we want our CLIs to look and behave – in particular, we want to engineer complex, imperative flows for validation. The amount of hacking required on other libraries wasn't worth it for us, so we made this instead.

### Comparisons

- Its closest relative is probably [mitchellh/cli](https://github.com/mitchellh/cli) (now archived). Like charli, it has imperative operation and is configured with structs – though uses some factories.
- [urfave/cli](https://charli.urfave.org/)'s config structs (`App`, `Command` etc.) have a similar layout.


### Goals

- **Provide only necessary validation.**
	- Syntax checking only.
	- No transformation for values – only strings (and bools, in the case of flags).
	- This is to provide full control over the validation process downstream.
- **Produce as many errors as possible.**
	- Aggregate errors. Downstream can decide how to deal with them.
	- Don't give up after encountering one parse error. Keep going!
	- Allow downstream validations to continue even with parse errors.
	- However: make downstream validations aware of previous errors, so that expensive operations can be short-circuited.
- **Render a relatively sane help format.**
	- We made
	- Also, make colors optional ([fatih/color](https://github.com/fatih/color) allows turning them off).
- **Idiomatic Go.**
	- Leverage the flexibility of structs and zero values.
	- Aim for a procedural style.
	- `io.Writer` galore.

## License

[![ISC](./.images/license.jpg)](./LICENSE)
