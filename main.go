package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/campbel/run/runfile"
	"github.com/campbel/yoshi"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Workflow string `yoshi:"WORKFLOW;The workflow to run;default"`
	Runfile  string `yoshi:"--runfile,-f;The runfile to use;run.yaml"`
	List     bool   `yoshi:"--list,-l;List workflows"`
	Actions  bool   `yoshi:"--actions,-a;List actions"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {
		runfile, err := loadRunfile(options.Runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load runfile")
		}

		// First thing, get all imports
		imports, err := fetchImports(runfile.Imports)
		if err != nil {
			return errors.Wrap(err, "failed to fetch imports")
		}
		for namespace, actionfile := range imports {
			for name, action := range actionfile.Actions {
				runfile.Actions[namespace+"/"+name] = action
			}
		}

		if options.List {
			listWorkflows(runfile.Workflows)
			return nil
		}

		if options.Actions {
			listActions(runfile.Actions)
			return nil
		}

		workflow, ok := runfile.Workflows[options.Workflow]
		if !ok {
			return fmt.Errorf("no workflow with the name '%s'", options.Workflow)
		}

		if err := runWorkflow(workflow, runfile.Actions); err != nil {
			return errors.Wrapf(err, "error on run workflow %s", options.Workflow)
		}

		return nil
	})
}

func loadActionFile(path string) (*runfile.Actionfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "error on read")
	}

	var actionfile runfile.Actionfile
	if err := yaml.Unmarshal(data, &actionfile); err != nil {
		return nil, errors.Wrap(err, "error on unmarshal")
	}

	return &actionfile, nil
}

func loadRunfile(path string) (*runfile.Runfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "error on read")
	}

	var runfile runfile.Runfile
	if err := yaml.Unmarshal(data, &runfile); err != nil {
		return nil, errors.Wrap(err, "error on unmarshal")
	}

	return &runfile, nil
}

var pwd = (func() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
})()

func fetchImports(imports map[string]string) (map[string]*runfile.Actionfile, error) {
	var merged = make(map[string]*runfile.Actionfile)
	for name, target := range imports {
		err := (&getter.Client{
			Src:  target,
			Dst:  filepath.Join(".run", "imports", name),
			Pwd:  pwd,
			Mode: getter.ClientModeAny,
		}).Get()
		if err != nil {
			return nil, errors.Wrapf(err, "error on get %s", target)
		}
		actionFile, err := loadActionFile(filepath.Join(".run", "imports", name, "run.yaml"))
		if err != nil {
			return nil, errors.Wrapf(err, "error on load action file %s", target)
		}
		merged[name] = actionFile
		results, err := fetchImports(actionFile.Imports)
		if err != nil {
			return nil, errors.Wrapf(err, "error on fetch imports %s", target)
		}
		for name, actionfile := range results {
			merged[name] = actionfile
		}
	}
	return merged, nil
}

func listWorkflows(workflows map[string]runfile.Workflow) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tabwriter, "WORKFLOW\tDESCRIPTION")
	for name := range workflows {
		fmt.Fprintf(tabwriter, "%s\t%s\n", name, workflows[name].Description)
	}
	tabwriter.Flush()
}

func listActions(actions map[string]runfile.Action) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tabwriter, "ACTION\tDESCRIPTION")
	for name, action := range actions {
		fmt.Fprintf(tabwriter, "%s\t%s\n", name, action.Description)
	}
	tabwriter.Flush()
}

func runWorkflow(workflow runfile.Workflow, actions map[string]runfile.Action) error {
	for _, actionName := range workflow.Actions {
		action, ok := actions[actionName]
		if !ok {
			return errors.Errorf("no workflow by that name %s", actionName)
		}

		if err := runAction(action); err != nil {
			return errors.Wrapf(err, "error on run action %s", actionName)
		}
	}
	return nil
}

func runAction(action runfile.Action) error {
	for _, cmd := range action.Commands {
		if err := runCommand(cmd); err != nil {
			return errors.Wrapf(err, "error on run command %s", cmd)
		}
	}
	return nil
}

func runCommand(cmd runfile.Command) error {
	if cmd.Shell != "" {
		parts := strings.Split(cmd.Shell, " ")
		command := exec.Command(parts[0], parts[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		return command.Run()
	} else if cmd.Action != "" {
		// return runActionCommand(cmd.Action)
	}
	return nil
}
