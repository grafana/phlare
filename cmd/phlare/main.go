package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/grafana/phlare/pkg/cfg"
	"github.com/grafana/phlare/pkg/phlare"
	"github.com/grafana/phlare/pkg/usage"
	_ "github.com/grafana/phlare/pkg/util/build"
)

type mainFlags struct {
	printModules bool
	printHelp    bool
	printHelpAll bool
}

func (mf *mainFlags) registerFlags(fs *flag.FlagSet) {
	fs.BoolVar(&mf.printModules, "modules", false, "List available values that can be used as target.")
	fs.BoolVar(&mf.printHelp, "h", false, "Print basic help.")
	fs.BoolVar(&mf.printHelp, "help", false, "Print basic help.")
	fs.BoolVar(&mf.printHelpAll, "help-all", false, "Print help, also including advanced and experimental parameters.")
}

func main() {
	var (
		config phlare.Config
	)

	// Register main flags and parse them first.
	mf, args := parseMainFlags(os.Args[1:])

	if err := cfg.DynamicUnmarshal(&config, args, flag.CommandLine); err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing config: %v\n", err)
		os.Exit(1)
	}

	f, err := phlare.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating phlare: %v\n", err)
		os.Exit(1)
	}

	if mf.printModules {
		allDeps := f.ModuleManager.DependenciesForModule(phlare.All)

		for _, m := range f.ModuleManager.UserVisibleModuleNames() {
			ix := sort.SearchStrings(allDeps, m)
			included := ix < len(allDeps) && allDeps[ix] == m

			if included {
				fmt.Fprintln(os.Stdout, m, "*")
			} else {
				fmt.Fprintln(os.Stdout, m)
			}
		}

		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Modules marked with * are included in target All.")
		return
	}

	if mf.printHelp || mf.printHelpAll {
		// Print available parameters to stdout, so that users can grep/less them easily.
		flag.CommandLine.SetOutput(os.Stdout)
		// Because we parse main flags separately, we need to create a dummy flagset to print help.
		var dummy mainFlags
		dummy.registerFlags(flag.CommandLine)
		if err := usage.Usage(mf.printHelpAll, &dummy, &config); err != nil {
			fmt.Fprintf(os.Stderr, "error printing usage: %s\n", err)
			os.Exit(1)
		}

		return
	}

	err = f.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed running phlare: %v\n", err)
		os.Exit(1)
	}
}

func parseMainFlags(args []string) (mainFlags, []string) {
	var mf mainFlags
	leftArgs := make([]string, 0, len(args))

	// Continue parsing flags if there is an error.
	fs := flag.NewFlagSet("main-flags", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	// Register main flags and parse them first.
	mf.registerFlags(fs)
	// Try to find all main flags in the arguments.
	// As Parsing stops on the first error, e.g. unknown flag, we simply
	// try remaining parameters until we find config flag, or there are no params left.
	// Put all other flags into leftArgs.
	for i := range args {
		if err := fs.Parse(args[i:]); err != nil {
			leftArgs = append(leftArgs, args[i])
		}
	}

	return mf, leftArgs
}
