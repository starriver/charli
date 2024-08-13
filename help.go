package charli

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Note that these aren't used by Parse(...).
// -h/--help/help are treated as special symbols there.
var fakeHelpOption = Option{
	Short:    'h',
	Long:     "help",
	Flag:     true,
	Headline: "Show this help",
}
var fakeHelpCmd = Command{
	Name:     "help",
	Headline: "Show this help",
}

// Help writes CLI help to w.
//
// program should be the name of the program,
// usually the first element of [os.Args].
//
// If cmd is nil, global help will be written.
// Otherwise, command help will be written.
//
// [App.Headline] is written first, followed by a usage line.
// For global help, the following are then written:
//   - [App.Description]
//   - A list of commands
//   - A list of options
//
// For command help, the following are instead written:
//   - [Command.Description]
//   - A list of options
func (app *App) Help(w io.Writer, program string, cmd *Command) {
	print := func(str string) {
		fmt.Fprint(w, str)
	}
	printf := func(format string, a ...any) {
		fmt.Fprintf(w, format, a...)
	}

	// Default to blue.
	hiColor := app.HighlightColor
	if hiColor == 0 {
		hiColor = color.FgHiBlue
	}

	hi := color.New(hiColor).SprintfFunc()
	bold := color.New(color.Bold).SprintfFunc()
	grey := color.New(color.Faint).SprintFunc()

	// Applys color to {FOO} and {-f/--foo} in descriptions.
	re := regexp.MustCompile(`\{.+?\}`)
	highlight := func(d string) string {
		return re.ReplaceAllStringFunc(d, func(s string) string {
			slashIndex := strings.Index(s, "/")
			if slashIndex != -1 {
				return hi(s[1:slashIndex]) +
					grey("/") +
					hi(s[slashIndex+1:len(s)-1])
			}
			return hi(s[1 : len(s)-1])
		})
	}

	// Aggregate all options now - we need to know whether to print [OPTIONS] in
	// the usage line.
	options := []Option{}
	if app.hasHelpFlags() {
		options = []Option{fakeHelpOption}
	}
	options = append(options, app.GlobalOptions...)
	if cmd != nil {
		options = append(options, cmd.Options...)
	}

	if app.Headline != "" {
		printf("%s\n", highlight(app.Headline))
	}

	basename := filepath.Base(program)
	printf("%s %s", bold("Usage:"), basename)

	var description string

	if cmd == nil {
		cmdStr := hi("COMMAND")
		if app.DefaultCommand != "" {
			cmdStr = fmt.Sprintf("[%s]", cmdStr)
		}
		if len(options) != 0 {
			printf(" [%s]", hi("OPTIONS"))
		}
		printf(" %s [...]", cmdStr)

		description = app.Description
	} else {
		if len(app.Commands) != 1 {
			if cmd.Name == app.DefaultCommand {
				printf(" [%s]", cmd.Name)
			} else {
				printf(" %s", cmd.Name)
			}
		}

		if len(options) != 0 {
			printf(" [%s]", hi("OPTIONS"))
		}

		args := &cmd.Args

		argsShown := max(args.Count, len(args.Metavars))
		for i := range argsShown {
			metavar := "ARG"
			if i < len(args.Metavars) {
				metavar = args.Metavars[i]
			}

			// Is this an optional arg?
			if i >= args.Count {
				ellipsis := ""
				if args.Varadic && (i == argsShown-1) {
					ellipsis = "..."
				}
				printf(" [%s%s]", hi(metavar), ellipsis)
			} else {
				printf(" %s", hi(metavar))
			}
		}

		if cmd.Headline != "" {
			printf("\n\n  %s", bold(cmd.Headline))
		}

		description = cmd.Description
	}

	if description != "" {
		description = description[1 : len(description)-1]
		description = strings.ReplaceAll(description, "\n", "\n  ")
		description = highlight(description)
		printf("\n\n  %s", description)
	}

	// Set up a left-align. These aren't in the following block because we reuse
	// these variables to render the command listing later.
	left := make([]string, len(options))
	lengths := make([]int, len(options))
	leftMax := 0

	if len(options) != 0 {
		printf("\n\n%s", bold("Options:"))

		slash := grey("/")

		for i, option := range options {
			l := 0

			if option.Short != 0 {
				left[i] += hi("-" + string(option.Short))
				l += 2
				if option.Long != "" {
					left[i] += slash
					l += 1
				}
			}
			if option.Long != "" {
				left[i] += hi("--" + option.Long)
				l += 2 + len(option.Long)
			}

			if !option.Flag {
				metavar := option.Metavar
				if metavar == "" {
					metavar = "VALUE"
				}
				left[i] += " " + hi(metavar)
				l += 1 + len(metavar)
			}

			if l > leftMax {
				leftMax = l
			}
			lengths[i] = l
		}

		// Add 2 more spaces of padding.
		leftMax += 2

		// These may be used repeatedly for choices.
		pipe := grey("|")
		bracketOpen := grey("[")
		bracketClose := grey("]")

		for i, str := range left {
			printf("\n  %s", str)

			option := &options[i]
			hasHeadline := option.Headline != ""
			hasChoices := len(option.Choices) != 0
			if hasHeadline || hasChoices {
				print(strings.Repeat(" ", leftMax-lengths[i]))

				if hasHeadline {
					print(highlight(option.Headline))
					if hasChoices {
						print(" ")
					}
				}
				if hasChoices {
					print(bracketOpen)
					for i, c := range option.Choices {
						print(hi(c))
						if i != len(option.Choices)-1 {
							print(pipe)
						}
					}
					print(bracketClose)
				}
			}
		}

	}

	print("\n")

	if cmd != nil {
		return
	}

	printf("\n%s", bold("Commands:"))

	cmds := app.Commands
	if (app.HelpAccess & HelpCommand) != 0 {
		cmds = append([]Command{fakeHelpCmd}, cmds...)
	}

	lengths = make([]int, len(cmds))
	leftMax = 0

	for i, cmd := range cmds {
		l := len(cmd.Name)
		if l > leftMax {
			leftMax = l
		}
		lengths[i] = l
	}

	leftMax += 2

	for i, cmd := range cmds {
		printf("\n  %s", hi(cmd.Name))
		if cmd.Headline != "" {
			print(strings.Repeat(" ", leftMax-lengths[i]))
			print(highlight(cmd.Headline))
		}
	}

	print("\n")
}
