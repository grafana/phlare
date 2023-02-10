package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/grafana/phlare/pkg/cfg"
	"github.com/grafana/phlare/pkg/phlare"
	"github.com/grafana/phlare/pkg/usage"
	_ "github.com/grafana/phlare/pkg/util/build"
)

func main() {
	var config phlare.Config

	if err := cfg.DynamicUnmarshal(&config, os.Args[1:], flag.CommandLine); err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing config: %v\n", err)
		os.Exit(1)
	}

	f, err := phlare.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating phlare: %v\n", err)
		os.Exit(1)
	}

	if config.MainFlags.PrintModules {
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

	if config.MainFlags.PrintHelp || config.MainFlags.PrintHelpAll {
		// Print available parameters to stdout, so that users can grep/less them easily.
		flag.CommandLine.SetOutput(os.Stdout)
		if err := usage.Usage(config.MainFlags.PrintHelpAll, &config); err != nil {
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
