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
	_ "github.com/grafana/phlare/pkg/util/build"
)

type mainFlags struct {
	printModules bool
}

func (mf *mainFlags) registerFlags(fs *flag.FlagSet) {
	fs.BoolVar(&mf.printModules, "modules", false, "List available values that can be used as target.")
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
	fs := flag.NewFlagSet("", flag.ContinueOnError)
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
