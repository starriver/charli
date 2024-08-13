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
	// In is displayed below the 'Usage:' line in help output.
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
	// The help flags override any flag with -h or --help that you may
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

// Configuration for a single CLI (sub-)command.
type Command struct {
	// Name of the command that the user will need to specify.
	Name string

	// Summary of the command. Displayed in help, below 'Usage:'.
	//
	// Text inside {curly braces} will be highlighted.
	Headline string

	// Long description of the command, displayed below usage in help output.
	// Should both start and end with a newline \n.
	//
	// This is usually a long, multiline string. Rather the inlining this string
	// to a Command literal, it's recommended to create a raw const string
	// elsewhere and use that instead, putting each backtick on its own line.
	// See the examples.
	//
	// Text inside {curly braces} will be highlighted.
	Description string

	// This command's options, in order of display in help output. If
	// the parent App.GlobalOptions is set, those options will be prepended.
	//
	// Note that -h/--help is always available on every command, and even when
	// no command is selected. The help flag overrides any flag with -h or
	// --help that you may provide.
	Options []Option

	// This command's trailing arguments. It's safe to leave this blank if you
	// don't have any.
	Args Args

	// Function to run if this Command is chosen.
	//
	// This is expected to further validate each option's values, *appending* to
	// r.Errs. If r.Errs has any members by the end of validation, it should
	// return false before proceeding with its business. At this point, it can
	// stop appending to r.Errs, but should only return true if the operation
	// succeeded overall.
	//
	// The caller should inspect r.Errs after this function returns, and exit
	// the program nonzero if it returned false.
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
