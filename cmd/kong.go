package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

const rootDescription = `Lint Claude Code projects.

Run without a subcommand to lint all components, component types, files, or directories:
  cclint
  cclint agents commands
  cclint --type agent ./custom/file.md

Run "cclint lint --help" for lint-only flags such as --type, --diff, and --staged.`

type cli struct {
	Root           *string        `short:"r" type:"path" help:"Project root directory."`
	Quiet          *bool          `short:"q" help:"Suppress non-essential output."`
	Verbose        *bool          `short:"v" help:"Enable verbose output."`
	Scores         *bool          `short:"s" help:"Show quality scores."`
	Improvements   *bool          `short:"i" help:"Show suggested improvements."`
	Format         *string        `short:"f" enum:"console,json,markdown" help:"Report output format."`
	Output         *string        `short:"o" type:"path" help:"Report output file."`
	FailOn         *string        `name:"fail-on" enum:"error,warning,suggestion" help:"Failure threshold."`
	NoCycleCheck   *bool          `name:"no-cycle-check" help:"Disable circular dependency detection."`
	Baseline       *bool          `help:"Filter known baseline issues."`
	BaselineCreate *bool          `name:"baseline-create" help:"Create or update the baseline."`
	BaselinePath   *string        `name:"baseline-path" help:"Baseline file path relative to the project root."`
	Version        bool           `short:"V" help:"Print version information."`
	Fmt            fmtCommand     `cmd:"" help:"Format component files canonically."`
	Summary        summaryCommand `cmd:"" help:"Show quality summary across all components."`
	Lint           lintCommand    `cmd:"" default:"withargs" help:"Lint files or directories."`
}

type summaryCommand struct{}
type lintCommand struct {
	Type   *string  `short:"t" enum:"agent,command,skill,settings,context,plugin,rule,output-style" help:"Force component type."`
	Diff   *bool    `help:"Lint only uncommitted changes (staged and unstaged)."`
	Staged *bool    `help:"Lint only staged files."`
	Paths  []string `arg:"" optional:"" name:"files-or-dirs"`
}
type parserExit int

type executionOptions struct {
	root           *string
	quiet          *bool
	verbose        *bool
	scores         *bool
	improvements   *bool
	format         *string
	output         *string
	failOn         *string
	noCycleCheck   *bool
	baseline       *bool
	baselineCreate *bool
	baselinePath   *string
}

func Execute() {
	var app cli
	parser, err := kong.New(
		&app,
		kong.Name("cclint"),
		kong.Description(rootDescription),
		kong.UsageOnError(),
		kong.Exit(func(code int) { panic(parserExit(code)) }),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:   true,
			Tree:      true,
			Summary:   true,
			FlagsLast: true,
		}),
	)
	if err != nil {
		exitWithError(err)
		return
	}
	ctx, err := parseCLI(parser, os.Args[1:])
	if err != nil {
		exitWithError(err)
		return
	}
	if ctx == nil {
		return
	}
	if app.Version {
		fmt.Printf("cclint version %s\n", Version)
		return
	}
	if err := ctx.Run(app.executionOptions()); err != nil {
		exitWithError(err)
	}
}

func parseCLI(parser *kong.Kong, args []string) (ctx *kong.Context, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(parserExit); ok {
				ctx, err = nil, nil
				return
			}
			panic(recovered)
		}
	}()
	return parser.Parse(args)
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	exitFunc(1)
}

func (cmd *lintCommand) Run(opts executionOptions) error { return runRootCommand(opts, cmd) }
func (summaryCommand) Run(opts executionOptions) error   { return runSummary(opts) }

func (app *cli) executionOptions() executionOptions {
	return executionOptions{
		root:           app.Root,
		quiet:          app.Quiet,
		verbose:        app.Verbose,
		scores:         app.Scores,
		improvements:   app.Improvements,
		format:         app.Format,
		output:         app.Output,
		failOn:         app.FailOn,
		noCycleCheck:   app.NoCycleCheck,
		baseline:       app.Baseline,
		baselineCreate: app.BaselineCreate,
		baselinePath:   app.BaselinePath,
	}
}
