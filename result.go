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

	// All errors encountered so far. You may append your own during your own
	// validations.
	Errs []error

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

// Append an error to Result.Errs.
func (r *Result) Error(str string) {
	r.Errs = append(r.Errs, errors.New(str))
}

// Append an error to Result.Errs.
func (r *Result) Errorf(format string, a ...any) {
	r.Errs = append(r.Errs, fmt.Errorf(format, a...))
}

// Shorthand for r.Command.Run(&r).
func (r *Result) RunCommand() bool {
	return r.Command.Run(r)
}

// Print app/command help to stderr.
func (r *Result) PrintHelp() {
	fmt.Fprint(os.Stderr, r.App.Help(os.Args[0], r.Command))
}
