package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/campbel/run/runfile"
	"github.com/campbel/yoshi"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Workflow string `yoshi:"WORKFLOW;The workflow to run;default"`
	Runfile  string `yoshi:"--runfile,-f;The runfile to use;run.yaml"`
	List     bool   `yoshi:"--list,-l;List workflows"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {
		runfile, err := loadRunfile(options.Runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load runfile")
		}

		if options.Workflow == "" {
			for name := range runfile.Workflows {
				println(name)
			}
		}

		if options.List {
			listWorkflows(runfile.Workflows)
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

func listWorkflows(workflows map[string]runfile.Workflow) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tabwriter, "WORKFLOW\tDESCRIPTION")
	for name := range workflows {
		fmt.Fprintf(tabwriter, "%s\t%s\n", name, workflows[name].Description)
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

func runCommand(cmd string) error {
	parts := strings.Split(cmd, " ")
	command := exec.Command(parts[0], parts[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
