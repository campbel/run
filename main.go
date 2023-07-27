package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"github.com/campbel/run/loader"
	"github.com/campbel/run/runfile"
	"github.com/campbel/yoshi"
	"github.com/pkg/errors"
)

type Options struct {
	Action   string            `yoshi:"ACTION;The action to run;default"`
	Vars     map[string]string `yoshi:"--vars,-v;The vars file to use"`
	Runfile  string            `yoshi:"--runfile,-f;The runfile to use;run.yaml"`
	List     bool              `yoshi:"--list,-l;List actions"`
	Download bool              `yoshi:"--download,-d;Force download dependencies"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {

		runfilePath := filepath.Join(pwd, options.Runfile)
		if _, err := os.Stat(runfilePath); err != nil {
			return errors.Wrap(err, "failed to find runfile")
		}

		data, err := os.ReadFile(runfilePath)
		if err != nil {
			return errors.Wrap(err, "failed to read runfile")
		}

		runfile, err := runfile.Unmarshal(data)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal runfile")
		}

		if options.List {
			listActions(runfile.Actions)
			return nil
		}

		mainPkg := loader.NewLoader(runfile, loader.NewGoGetter(options.Download)).Load()

		action, ok := mainPkg.Actions[options.Action]
		if !ok {
			return fmt.Errorf("no action with the name '%s'", options.Action)
		}

		return action.Run(options.Vars)
	})
}

var pwd = (func() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
})()

func listActions(actions map[string]runfile.Action) {
	var actionNames []string
	for name := range actions {
		if name == "default" {
			continue
		}
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	tabwriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tabwriter, "ACTION\tDESCRIPTION")
	for _, name := range actionNames {
		fmt.Fprintf(tabwriter, "%s\t%s\n", name, actions[name].Description)
	}
	tabwriter.Flush()
}
