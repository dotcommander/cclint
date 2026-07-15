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

func Execute() {
	var app cli
	parser, err := kong.New(&app, kong.Name("cclint"), kong.Description(rootDescription), kong.UsageOnError(), kong.Exit(func(code int) { panic(parserExit(code)) }))
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
	app.apply()
	defer func() { cliChanged = nil }()
	if app.Version {
		fmt.Printf("cclint version %s\n", Version)
		return
	}
	if err := ctx.Run(); err != nil {
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

func (cmd *lintCommand) Run() error { return runRootCommand(cmd.Paths) }
func (summaryCommand) Run() error   { return runSummary() }

func (app *cli) apply() {
	cliChanged = make(map[string]bool)
	setString := func(name string, value *string, target *string) {
		if value != nil {
			*target = *value
			cliChanged[name] = true
		}
	}
	setBool := func(name string, value *bool, target *bool) {
		if value != nil {
			*target = *value
			cliChanged[name] = true
		}
	}
	setString("root", app.Root, &rootPath)
	setBool("quiet", app.Quiet, &quiet)
	setBool("verbose", app.Verbose, &verbose)
	setBool("scores", app.Scores, &showScores)
	setBool("improvements", app.Improvements, &showImprovements)
	setString("format", app.Format, &outputFormat)
	setString("output", app.Output, &outputFile)
	setString("fail-on", app.FailOn, &failOn)
	setBool("no-cycle-check", app.NoCycleCheck, &noCycleCheck)
	setBool("baseline", app.Baseline, &useBaseline)
	setBool("baseline-create", app.BaselineCreate, &createBaseline)
	setString("baseline-path", app.BaselinePath, &baselinePath)
	setString("type", app.Lint.Type, &typeFlag)
	setBool("diff", app.Lint.Diff, &diffMode)
	setBool("staged", app.Lint.Staged, &stagedMode)
	app.Fmt.apply()
}
