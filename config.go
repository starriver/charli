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
	// In command help output, it is also displayed below the 'Usage:' line.
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

// Configuration for a single option, providing a value or a flag.
type Option struct {
	// Short option name, used for single-dashed args (like '-a'). Don't include
	// the hyphen. This is optional, but one of either Short of Long must be
	// specified.
	Short rune

	// Long option name, used for double-dashed args (like '--all'). Don't
	// include the hyphens. This is optional, but one of either Short of Long
	// must be specified.
	Long string

	// Set to true if this option is a flag, ie. it takes no value.
	Flag bool

	// If set, adds a simple from-a-list constraint to this option. The
	// available choices will be appended to the option's headline. Has no
	// effect on flags.
	Choices []string

	// Sets a term for this option's value - usually uppercase (like
	// '-p/--person NAME'). Optional but recommended.
	Metavar string

	// Summary of the option. Displayed in the command help, and at the top of
	// the command's help output. Text inside {curly braces} will be
	// highlighted.
	Headline string
}

// Configuration for positional arguments.
type Args struct {
	// Number of expected positional args. If Varadic is set, this is the
	// *minimum* number of positional args.
	Count int

	// Whether to accept more args than specified in Count.
	Varadic bool

	// Sets term(s) for the args. These are displayed in the usage line at the
	// top of help. They should be uppercase. '[--]' will be prepended, and
	// names will be [bracketed] appropriately if Varadic is set.
	Metavars []string
}

type HelpAccess uint8

const (
	HelpFlag HelpAccess = 1 << iota
	HelpCommand
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
