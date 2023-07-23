package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	Dump     bool   `yoshi:"--dump,-d;Dump the runfile"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {
		runfile, err := loadRunfile(options.Runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load runfile")
		}

		// First thing, get all imports
		rootScope, err := loadRootScope(runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load action context")
		}

		if options.List {
			listWorkflows(runfile.Workflows)
			return nil
		}

		if options.Dump {
			data, err := json.Marshal(rootScope)
			if err != nil {
				return errors.Wrap(err, "error on marshal")
			}
			fmt.Println(string(data))
			return nil
		}

		workflow, ok := runfile.Workflows[options.Workflow]
		if !ok {
			return fmt.Errorf("no workflow with the name '%s'", options.Workflow)
		}

		if err := runWorkflow(workflow, rootScope); err != nil {
			return errors.Wrapf(err, "error on run workflow %s", options.Workflow)
		}

		return nil
	})
}

func loadActionFile(imp string) (*runfile.Actionfile, error) {
	dst := path(imp)
	if err := fetch(imp, dst); err != nil {
		return nil, errors.Wrapf(err, "error on fetch %s", imp)
	}

	data, err := os.ReadFile(dst + "/run.yaml")
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

func listWorkflows(workflows map[string]runfile.Workflow) {
	tabwriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tabwriter, "WORKFLOW\tDESCRIPTION")
	for name := range workflows {
		fmt.Fprintf(tabwriter, "%s\t%s\n", name, workflows[name].Description)
	}
	tabwriter.Flush()
}

func runWorkflow(workflow runfile.Workflow, scope *Scope) error {
	for _, actionName := range workflow.Actions {
		action, ok := scope.Actions[actionName]
		if !ok {
			return errors.Errorf("no action by that name %s", actionName)
		}
		if err := action.Run(); err != nil {
			return errors.Wrapf(err, "error on action run %s", actionName)
		}
	}
	return nil
}

func loadRootScope(root *runfile.Runfile) (*Scope, error) {
	rootAction := &runfile.Actionfile{
		Imports: root.Imports,
		Actions: root.Actions,
	}
	return loadScope(rootAction)
}

func loadScope(actionfile *runfile.Actionfile) (*Scope, error) {
	scope := NewScope()

	for name, action := range actionfile.Actions {
		scope.Actions[name] = &ActionContext{
			Scope:    scope,
			Commands: newCommandContexts(action.Commands),
		}
	}

	for namespace, imp := range actionfile.Imports {
		actionfile, err := loadActionFile(imp)
		if err != nil {
			return nil, errors.Wrapf(err, "error on load action file %s", namespace)
		}
		s, err := loadScope(actionfile)
		if err != nil {
			return nil, errors.Wrapf(err, "error on load scope %s", namespace)
		}
		for name, action := range s.Actions {
			scope.Actions[namespace+"."+name] = action
		}
	}

	return scope, nil
}

func path(imp string) string {
	return filepath.Join(pwd, ".run", "imports", imp)
}

func fetch(src, dst string) error {
	if err := (&getter.Client{
		Src:  src,
		Dst:  dst,
		Pwd:  pwd,
		Mode: getter.ClientModeAny,
	}).Get(); err != nil {
		return err
	}
	return nil
}
