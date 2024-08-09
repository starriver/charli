package charli

import (
	"errors"
	"fmt"
	"os"
)

// Parsing results. Returned by App.Parse(...).
type Result struct {
	// Suggested action. Note that regardless of this value, Errs may be
	// populated.
	Action Action

	// All errors encountered during Parse(...). You may choose to append your
	// errors from your own validations.
	//
	// Note that if App.ErrorHandler was set, errors won't be appended here by
	// default - but Fail will still be appropriately set. So, prefer checking
	// Fail rather than len(Errs) == 0.
	Errs []error

	// Whether any errors have occurred so far. Calling Result.Error(...) or
	// Result.Errorf(...) will set this.
	Fail bool

	// The app you called Parse(...) on.
	App *App

	// The chosen command. May be nil if Action != Proceed.
	Command *Command

	// Map of option values.
	//
	// You can access values with either the short or long option names. Don't
	// include hyphens - eg. use "o" or "opt" rather than "-o" or "--opt".
	Options map[string]*OptionResult

	// List of positional args.
	//
	// In the case of too many args being supplied, len(Args) won't be more than
	// Command.Args.Count - that is, the extraneous args will be dropped. This
	// obviously doesn't apply if Command.Args.Varadic is set.
	Args []string
}

// Suggested action after parsing.
type Action int

const (
	// Proceed to Run(...) the command. Exit 0 if Run(...) returns true, nonzero
	// otherwise.
	Proceed Action = iota

	// Display help. The user explicitly requested it, so exit 0.
	HelpOK

	// Display help. The user didn't request it (or requested it wonkily) so
	// exit nonzero.
	HelpError

	// Nothing more to do. Exit nonzero.
	Fatal
)

// Parsing results for a single option.
type OptionResult struct {
	// The original Option that this OptionResult provides for.
	Option *Option

	// The option's string value. This may be blank if an empty value was
	// supplied (eg. --opt '') - check IsSet to be sure.
	//
	// If the option is a flag, this will always be blank.
	Value string

	// Whether the option was supplied.
	IsSet bool
}

// Note: this is the only file in the lib that provides impure functions. Nice!

func (r *Result) Error(err error) {
	r.Fail = true

	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// Raise a general error. App.ErrorHandler(...) will be called if set -
// otherwise an error will be appeneded to Result.Errs.
func (r *Result) ErrorString(str string) {
	r.Fail = true

	err := errors.New(str)
	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// Raise an error. App.ErrorHandler(...) will be called if set - otherwise an
// error will be appeneded to Result.Errs.
func (r *Result) Errorf(format string, a ...any) {
	r.Fail = true

	err := fmt.Errorf(format, a...)

	if r.App.ErrorHandler != nil {
		r.App.ErrorHandler(err)
	} else {
		r.Errs = append(r.Errs, err)
	}
}

// Shorthand for r.Command.Run(&r).
func (r *Result) RunCommand() bool {
	return r.Command.Run(r)
}

// Print app/command help to stderr.
func (r *Result) PrintHelp() {
	r.App.Help(os.Stderr, os.Args[0], r.Command)
}
