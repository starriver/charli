package charli

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Complete writes shell completions to w for argv.
// argv should be a slice of the arguments currently on the command line,
// including the program name, and truncated at the cursor position.
// The last element of argv may be an empty string,
// in which case all relevant completions will be written.
//
// Completions are line-separated.
// Each line contains a potential completion, followed by `\t`,
// then a description of the completion.
//
// Regarding shell completion functions that use this function's output:
//
//   - The bash function should truncate each line at the `\t`.
//   - The fish function doesn't need to truncate each line,
//     as it will display the characters after the `\t` as a description
//     for each completion.
func (app *App) Complete(w io.Writer, argv []string) {
	if len(argv) <= 2 {
		panic("argv appears truncated")
	}

	args := argv[2:]
	i := len(args) - 1

	// First ensure the arg being completed isn't beyond a --.
	for j := 0; j < i; j++ {
		if args[j] == "--" {
			return
		}
	}

	cur := args[i]
	prev := ""
	if i != 0 {
		prev = args[i-1]
	}

	helpFlags := []string{}
	if app.hasHelpFlags() {
		helpFlags = []string{"-h", "--help"}
	}

	singleCmd := len(app.Commands) == 1

	completeFor := func(word, headline, headlineDefault string) {
		if !strings.HasPrefix(word, cur) {
			return
		}
		fmt.Fprintf(w, "%s\t", word)
		if headline != "" {
			fmt.Fprint(w, headline)
		} else {
			fmt.Fprint(w, headlineDefault)
		}
		fmt.Fprintln(w)
	}

	var cmd *Command
	helpFirst := false

	if app.hasHelpFlags() && (args[0] == "-h" || args[0] == "--help") {
		helpFirst = true
	}
	if app.hasHelpCommand() && args[0] != "help" {
		helpFirst = true
	}

	// Can we complete a command (or -h/--help/help)?
	if i == 0 || (helpFirst && i == 1) {
		if !singleCmd {
			for _, cmd := range app.Commands {
				completeFor(cmd.Name, cmd.Headline, "Command")
			}
		}

		singleOrDefault := singleCmd || app.DefaultCommand != ""
		if i == 0 {
			if app.hasHelpCommand() {
				completeFor("help", "Show help", "")
			}
			if !singleOrDefault {
				for _, f := range helpFlags {
					completeFor(f, "Show help", "")
				}
			}
		}

		if !singleOrDefault {
			return
		}
	}

	if singleCmd {
		cmd = &app.Commands[0]
	} else {
		cmdArg := args[0]
		if app.DefaultCommand != "" && (args[0] == "" || isOption(args[0])) {
			cmdArg = app.DefaultCommand
		}

		for _, c := range app.Commands {
			if c.Name == cmdArg {
				cmd = &c
				break
			}
		}
		if cmd == nil {
			return
		}
	}

	// Can we complete a non-flag (maybe with choices)?
	if isOption(prev) && !strings.ContainsRune(prev, '=') {
		var opt *Option
		if isLongOption(prev) {
			long := prev[2:]
			for _, o := range cmd.Options {
				if long == o.Long {
					opt = &o
					break
				}
			}
		} else if len(prev) == 2 { // Ignore combined options
			short := rune(prev[1])
			for _, o := range cmd.Options {
				if short == o.Short {
					opt = &o
					break
				}
			}
		}

		if opt != nil {
			if len(opt.Choices) != 0 {
				for _, c := range opt.Choices {
					metavar := "ARG"
					if opt.Metavar != "" {
						metavar = opt.Metavar
					}
					completeFor(c, fmt.Sprintf("%s %s", prev, metavar), "")
				}
			}

			// If the option is expecting any value, don't complete further.
			if !opt.Flag {
				return
			}
		}
	}

	// Lastly, just complete options.
	opts := append(app.GlobalOptions, cmd.Options...)
	if app.hasHelpFlags() {
		helpOpt := fakeHelpOption
		helpOpt.Headline = "Show help"
		opts = append(opts, helpOpt)
	}
	for _, opt := range opts {
		defaultHeadline := "Option"
		if opt.Flag {
			defaultHeadline = "Flag"
		}

		if opt.Short != 0 {
			short := "-" + string(opt.Short)
			completeFor(short, opt.Headline, defaultHeadline)
		}

		if opt.Long != "" {
			long := "--" + opt.Long
			completeFor(long, opt.Headline, defaultHeadline)
		}
	}
}

func quote(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

var idRe *regexp.Regexp

// Derive a shell identifier-compatible name for program.
func shellID(program string) string {
	if idRe == nil {
		idRe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	}
	return idRe.ReplaceAllString(program, "_")
}

// GenerateBashCompletions writes a bash completion script to w.
//
// program should be the program name (which will presumably be in the user's
// PATH).
// flag should be a special trigger flag, *including* hyphen prefixes,
// which your program should use to bypass normal execution
// and generate completions instead (using [App.Complete]).
//
// flag can be anything you want, but don't use anything ambiguous to your CLI.
// If in doubt, use `--_complete`.
func GenerateBashCompletions(w io.Writer, program, flag string) {
	qprogram := quote(program)
	funcName := fmt.Sprintf("_complete_charli_%s", shellID(program))

	// Write a function that calls the program with the required completion
	// data.
	fmt.Fprintf(w, "%s() {\n", funcName)
	fmt.Fprintln(w, "\tlocal cur=\"${COMP_WORDS[$COMP_CWORD]}\"")
	fmt.Fprintln(w, "\tlocal iprev=\"$(( $COMP_CWORD - 1 ))\"")
	fmt.Fprintln(w, "\twhile IFS= read -r c; do")
	fmt.Fprintln(w, "\t\tCOMPREPLY+=(\"${c%%$'\\t'*}\")")
	fmt.Fprintf(
		w,
		"\tdone <<< \"$(%s %s ${COMP_WORDS[@]:1:$iprev} \"$cur\")\"\n",
		qprogram,
		flag,
	)
	fmt.Fprintln(w, "}")

	// Complete using the function.
	fmt.Fprintf(
		w,
		"complete -F %s %s\n",
		funcName,
		qprogram,
	)
}

// GenerateFishCompletions writes a fish completion script to w.
//
// program should be the program name (which will presumably be in the user's
// PATH).
// flag should be a special trigger flag, *including* hyphen prefixes,
// which your program should use to bypass normal execution
// and generate completions instead (using [App.Complete]).
//
// flag can be anything you want, but don't use anything ambiguous to your CLI.
// If in doubt, use `--_complete`.
func GenerateFishCompletions(w io.Writer, program, flag string) {
	qprogram := quote(program)
	funcName := fmt.Sprintf("__fish_complete_charli_%s", shellID(program))

	fmt.Fprintf(w, "function %s\n", funcName)
	fmt.Fprintln(w, "\tset -l tokens (commandline -cop)")
	fmt.Fprintln(w, "\tset -l cur (commandline -ct)")
	fmt.Fprintf(w, "\t%s %s $tokens[2..-1] \"$cur\"\n", qprogram, flag)
	fmt.Fprintln(w, "end")
	fmt.Fprintf(w, "complete -c %s -f -k -a '(%s)'\n", qprogram, funcName)
}
