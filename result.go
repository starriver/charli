package charli

import (
	"errors"
	"fmt"
	"os"
)

// A Result is a collection of results returned by [App.Parse].
type Result struct {
	// Action indicates what the parser suggests should happen next.
	// See [Action] for details.
	Action Action

	// Errs is a slice of all errors encountered during parsing.
	// You may choose to append errors from your own validations
	// using the [Result.Error] functions.
	//
	// Note that if [App.ErrorHandler] is set, this slice will not be appended
	// to automatically.
	// [App.ErrorHandler] may re-implement this behavior, if desired.
	Errs []error

	// Fail is true if any error has occurred during parsing.
	// If you are using the [Result.Error] functions in your own validations,
	// any call to those functions will set this to true.
	//
	// If true, it is suggested that the program exits with a failure code.
	//
	// Note also that it is valid for Fail to be true when [Result.Errs] is
	// empty - for example, when [Result.Action] is [Help], but extraneous
	// arguments were supplied.
	Fail bool

	// App is the [App] that [App.Parse] was called on.
	App *App

	// Command is the [Command] chosen by the user.
	// This may be nil when [Result.Action] != [Proceed].
	Command *Command

	// Options is a map of [Option] names to [OptionResult]s.
	//
	// Both [Option.Short] and [Option.Long] will be set as keys for the
	// [OptionResult] for a given [Option].
	// Don't prepend the hyphens in either case (ie. use `opt` as a key,
	// rather than `--opt`).
	Options map[string]*OptionResult

	// Args is a slice of the positional arguments.
	//
	// In the case of too many arguments being supplied,
	// `len(Args)` won't be more than [Args.Count] for the given [Command].
	// In other words, the extraneous args will be dropped.
	Args []string
}

// Action indicates what the parser suggests should happen next.
type Action int

const (
	Proceed Action = iota // proceed to call [Command.Run]
	Help                  // display help
	Fatal                 // nothing else to do; [Result.Fail] will always be true
)

// An OptionResult contains parsing results for a single option.
type OptionResult struct {
	// Option is the [Option] that this refers to.
	Option *Option

	// Value is the option's string value.
	// This may be blank if an empty value was supplied (like `--opt ''`) -
	// [OptionResult.IsSet] is preferable when checking whether an option
	// was supplied at all.
	//
	// If the option is a flag, this will always be blank.
	Value string

	// IsSet indicates whether the option was supplied.
	IsSet bool
}

// Error reports a pre-made [error] and sets [Result.Fail] to true.
// This is called by [App.Parse], and you can use it in your own validations.
// Unless [App.ErrorHandler] is set,
// the error will be appended to [Result.Errs].
// Otherwise, the handler will be called.
func (r *Result) Error(err error) {
	r.Fail = true

	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// ErrorString reports an [error] described by str and sets [Result.Fail] to
// true.
// This is called by [App.Parse], and you can use it in your own validations.
// Unless [App.ErrorHandler] is set,
// the error will be appended to [Result.Errs].
// Otherwise, the handler will be called.
func (r *Result) ErrorString(str string) {
	r.Fail = true

	err := errors.New(str)
	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// Errorf reports an [error] using the specified format and args and sets
// [Result.Fail] to true.
// This is called by [App.Parse], and you can use it in your own validations.
// Unless [App.ErrorHandler] is set,
// the error will be appended to [Result.Errs].
// Otherwise, the handler will be called.
func (r *Result) Errorf(format string, a ...any) {
	r.Fail = true

	err := fmt.Errorf(format, a...)

	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// RunCommand calls [Command.Run] for the command the user chose.
// This is shorthand for `r.Command.Run(&r)`.
func (r *Result) RunCommand() {
	r.Command.Run(r)
}

// PrintHelp writes global or command help to stderr, depending on whether the
// user selected a valid command.
func (r *Result) PrintHelp() {
	r.App.Help(os.Stderr, os.Args[0], r.Command)
}
