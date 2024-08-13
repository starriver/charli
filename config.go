// Package charli is a small CLI toolkit.
//
// It includes a CLI parser ([App.Parse]), help formatter ([App.Help]) and shell
// completions ([App.Complete]).
//
// Configure your CLI with an [App]. Apps have one or several [Command]s, each
// with several [Option]s and [Args].
package charli

import (
	"fmt"

	"github.com/fatih/color"
)

// An App contains configuration for your CLI.
type App struct {
	// Headline is the text displayed at the top of help output,
	// above the 'Usage:' line.
	//
	// Text surrounded by {curly braces} will be highlighted.
	Headline string

	// Description is a longer description of the app,
	// which may be several paragraphs long.
	//
	// If using only a single command, this won't be displayed.
	// Instead, set [Command.Description].
	//
	// In global help output, this is displayed below the 'Usage:' line.
	// Each line will be indented by 2 spaces.
	//
	// It should start and end with a newline `\n`.
	// This is because it's designed to be written using a raw string literal,
	// with each backtick on a separate line, like this:
	//
	//  const description = `
	//  This is the first paragraph.
	//
	//  This is the second paragraph.
	//  `
	//
	// The text won't be automatically justified when printed.
	// Generally, it should be pre-justified at 78 characters
	// (to make up for the 2-space indent).
	//
	// Text surrounded by {curly braces} will be highlighted.
	Description string

	// Commands configures the app's [Command]s,
	// in the order they should be displayed in help output.
	//
	// If only one command is configured,
	// the parser won't require a command name will be supplied at all,
	// and [Command.Name] should be omitted.
	// Otherwise, all commands must have a name.
	//
	// Commands must not have duplicate names.
	Commands []Command

	// GlobalOptions configures [Option]s that apply to every command in
	// [App.Commands].
	// They are effectively prepended to every [Command.Options] slice.
	//
	// Note that unless [App.HelpAccess] explicitly removes it,
	// `-h/--help` is always available.
	// The help flags override any flag with `-h` or `--help` that you may
	// configure.
	GlobalOptions []Option

	// DefaultCommand is the name of the command to run if none is supplied.
	// If blank, the parser will require a command.
	//
	// Note that setting a default can introduce some ambiguity
	// where the first supplied argument isn't an option
	// (ie. it doesn't start with `-`).
	//
	// If only one command is configured,
	// this must be blank.
	DefaultCommand string

	// HelpAccess specifies how users can access help.
	//
	// This is a bitmask.
	// It should be [HelpFlag], [HelpCommand], or both.
	//
	// If nothing is supplied, it will default to [HelpFlag].
	HelpAccess HelpAccess

	// ErrorHandler is a callback which, if set,
	// will handle [App.Parse] errors as they happen.
	//
	// If this *isn't* set,
	// errors will be aggregated in [Result.Errs].
	ErrorHandler func(error)

	// HighlightColor is the color used for highlighting in help output.
	//
	// To disable color, don't use this.
	// Instead, set [github.com/fatih/color.NoColor] to true.
	HighlightColor color.Attribute
}

// A Command contains configuration for a single CLI command.
type Command struct {
	// Name is the command name.
	//
	// It should be short (ideally a single word, or kebab-case if not).
	Name string

	// Headline is a one-line summary of the command.
	//
	// In global help output,
	// it is displayed alongside the command name in the 'Commands:' listing.
	// In command help output,
	// it is also displayed below the 'Usage:' line.
	//
	// Text surrounded by {curly braces} will be highlighted.
	Headline string

	// Description is a longer description of the command,
	// which may be several paragraphs long.
	// In command help output, it is displayed below the 'Usage:' line.
	// Each line will be indented by 2 spaces.
	//
	// It should start and end with a newline `\n`.
	// This is because it's designed to be written using a raw string literal,
	// with each backtick on a separate line, like this:
	//
	//  const description = `
	//  This is the first paragraph.
	//
	//  This is the second paragraph.
	//  `
	//
	// The text won't be automatically justified when printed.
	// Generally, it should be pre-justified at 78 characters
	// (to make up for the 2-space indent).
	//
	// Text surrounded by {curly braces} will be highlighted.
	Description string

	// Options is a slice of this command's unique [Option]s,
	// in order of display in command help output.
	// If [App.GlobalOptions] is set, those options will effectively be
	// prepended to this slice.
	//
	// Note that unless [App.HelpAccess] explicitly removes it,
	// `-h/--help` is always available.
	// The help flags override any flag with `-h` or `--help` that you may
	// configure.
	Options []Option

	// Args is the configuration for this command's positional arguments.
	//
	// If left blank, no positional arguments will be allowed.
	Args Args

	// Run is the function to execute if this Command is chosen.
	//
	// Supplying this function is actually entirely optional,
	// as it's up to you whether to call it.
	// You may wish to have an entirely different way of reacting to the
	// command the user has chosen.
	//
	// If you are supplying a Run function, it should:
	//
	//   - Validate the option values in the passed [Result].
	//   - Call [Result.Error] (or its other `Error*` functions) in the case
	//     of errors).
	//   - Return before doing any meaningful work if [Result.Fail] is true.
	//
	// [Result.Fail] may additionally be set by this function for any reason.
	//
	// Responsibility for processing and displaying [Result.Errs] should
	// generally be with the caller,
	// or you may wish to set [App.ErrorHandler].
	Run func(r *Result)
}

// An Option contains configuration for a single CLI option.
type Option struct {
	// Short is the short option name. This and/or [Option.Long] must be set.
	//
	// Short options have a single hyphen followed by a character (like `-a`).
	// Omit the hyphen, as this is a rune.
	//
	// `-` isn't a valid value for this,
	// because `--` is a special argument used to disable option parsing.
	Short rune

	// Long is the long option name. This and/or [Option.Short] must be set.
	//
	// Long options have a double hyphen followed by a string (like `--option`).
	// Don't include the hyphens.
	Long string

	// Flag indicates whether this option should take no value.
	// Flags are effectively boolean.
	//
	// If true, the option will be listed without [Option.Metavar]
	// in help output (like `--option` rather than `--option ARG`).
	//
	// After parsing, in this option's [OptionResult],
	// [OptionResult.IsSet] can be used to check whether the flag was supplied.
	Flag bool

	// Choices constrains this option's values to a list.
	//
	// Available choices will be appended to the option's headline in help
	// output.
	// This should only be used for simple, short sets of strings.
	//
	// This is invalid if set on flags.
	Choices []string

	// Metavar is the term for this option's value. It is shown in help output
	// after the option name(s), like `VALUE` in `-o/--option VALUE`.
	//
	// If omitted, it will default to `ARG`.
	//
	// It should be uppercase, but this is not a requirement.
	//
	// It is invalid if set with [Option.Flag].
	Metavar string

	// Headline is a one-line summary of the option,
	// shown in the 'Options:' section of help output.
	//
	// Text surrounded by {curly braces} will be highlighted.
	Headline string
}

// An Args contains configuration for positional arguments.
type Args struct {
	// Count is the number of required positional arguments.
	//
	// When parsing, if more positional arguments than Count are supplied,
	// those arguments are excluded from [Result.Args].
	//
	// If [Args.Varadic] is set, this becomes the *minimum* number of arguments.
	Count int

	// Varadic indicates whether to allow more positional arguments
	// than are specified in [Args.Count].
	Varadic bool

	// Metavars is a list of names for the positional arguments.
	//
	// Omitted metavars will default to `ARG`.
	//
	// They are displayed in the 'Usage:' line at the top of help output,
	// like `Usage: program ARG1 ARG2 ARG3`.
	//
	// They should be uppercase, but this is not a requirement.
	//
	// In help output, if [Args.Varadic] is true:
	//
	//   - `[--]` will be prepended.
	//   - Metavars with an index higher than [Args.Count] will be
	//     `[bracketed]`.
	//   - The last metavar will be ellipsized, like `ARG...`.
	Metavars []string
}

// HelpAccess indicates how help output should be accessed by the CLI user.
// This is a bitmask.
type HelpAccess uint8

const (
	HelpFlag    HelpAccess = 1 << iota // access via the `-h/--help` flags
	HelpCommand                        // access via a `help` pseudo-command
)

func (app *App) hasHelpFlags() bool {
	return app.HelpAccess == 0 || app.HelpAccess&HelpFlag != 0
}

func (app *App) hasHelpCommand() bool {
	return app.HelpAccess&HelpCommand != 0
}

func (app *App) cmdMap() (m map[string]*Command) {
	m = make(map[string]*Command, len(app.Commands))
	for _, cmd := range app.Commands {
		if _, ok := m[cmd.Name]; ok {
			panic(
				fmt.Sprintf("Duplicate command '%s' configured", cmd.Name),
			)
		}
		m[cmd.Name] = &cmd
	}
	return
}
