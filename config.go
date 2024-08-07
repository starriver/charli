package charli

import (
	"fmt"

	"github.com/fatih/color"
)

// CLI application configuration.
type App struct {
	// Headline of the app, displayed at the top of help output.
	//
	// Note that unlike Command.Headline, this is displayed above *everything*,
	// and in all help output (regardless of command). It also won't be
	// formatted in the same way ({} won't be highlighted).
	Headline string

	// Long description of the app, displayed below usage in help output. Should
	// both start and end with a newline \n.
	//
	// This is usually a long, multiline string. Rather the inlining this string
	// to an App literal, it's recommended to create a raw const string
	// elsewhere and use that instead, putting each backtick on its own line.
	// See the examples.
	//
	// Text inside {curly braces} will be highlighted.
	Description string

	// The app's commands, in order of display in help output.
	Commands []Command

	// Options that apply to all commands. These are prepended to every
	// Command.Options slice.
	//
	// Note that -h/--help is always available on every command, and even when
	// no command is selected. The help flag overrides any flag with -h or
	// --help that you may provide.
	GlobalOptions []Option

	// Name of the command to run if none is supplied. Leave blank to require
	// one.
	DefaultCommand string

	// How users can access help. This is a bitmask: either HelpFlag,
	// HelpCommand or both.
	//
	// If nothing is supplied, will default to HelpFlag.
	HelpAccess HelpAccess

	// Set this if you'd prefer to handle Parse(...) errors as they happen. The
	// default behaviour (if this isn't set) is to aggregate them in the
	// returned Result.Errs slice.
	ErrorHandler func(string)

	// Highlight color in help output.
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
	Run func(r *Result) bool
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
	return (app.HelpAccess & HelpCommand) != 0
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
