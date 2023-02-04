package main

import (
	_ "embed"
	"flag"
	"fmt"
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
		mf     mainFlags
		config phlare.Config
	)

	mf.registerFlags(flag.CommandLine)

	if err := cfg.DynamicUnmarshal(&config, os.Args[1:], flag.CommandLine); err != nil {
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
