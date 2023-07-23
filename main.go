package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	Dump     bool   `yoshi:"--dump,-d;Dump the runfile"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {
		runfile, err := loadRunfile(options.Runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load runfile")
		}

		// First thing, get all imports
		rootScope, err := loadScope(runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load action context")
		}

		if options.List {
			listWorkflows(runfile.Workflows)
			return nil
		}

		if options.Dump {
			data, err := yaml.Marshal(rootScope)
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
	for _, command := range workflow.Actions {
		if action, ok := scope.Actions[command.Action]; ok {
			if err := action.Run(command.Args); err != nil {
				return errors.Wrapf(err, "error on action run %s", command.Action)
			}
			continue
		}
		if action, ok := scope.Imports[command.Action]; ok {
			if err := action.Run(command.Args); err != nil {
				return errors.Wrapf(err, "error on action run %s", command.Action)
			}
			continue
		}
	}
	return nil
}

func loadScope(runfile *runfile.Runfile, path ...string) (*Scope, error) {
	scope := NewScope()

	for name, action := range runfile.Actions {
		scope.Actions[name] = &ActionContext{
			Scope:    scope,
			Name:     strings.Join(append(path, name), "."),
			Skip:     newCommandContext(action.Skip),
			Commands: newCommandContexts(action.Commands),
		}
	}

	for namespace, imp := range runfile.Imports {
		runfile, err := fetchRunfile(imp)
		if err != nil {
			return nil, errors.Wrapf(err, "error on load action file %s", namespace)
		}
		s, err := loadScope(runfile, append(path, namespace)...)
		if err != nil {
			return nil, errors.Wrapf(err, "error on load scope %s", namespace)
		}
		for name, action := range s.Actions {
			scope.Imports[namespace+"."+name] = action
		}
	}

	return scope, nil
}

func fetchRunfile(imp string) (*runfile.Runfile, error) {
	dst := filepath.Join(pwd, ".run", "imports", imp)
	if _, err := os.Stat(dst); err != nil {
		if err := fetch(imp, dst); err != nil {
			return nil, errors.Wrapf(err, "error on fetch %s", imp)
		}
	}
	return loadRunfile(dst)
}

func loadRunfile(path string) (*runfile.Runfile, error) {
	// If the path has no extension, we assume its a directory and load both
	// the common file and the os specific file
	if filepath.Ext(path) == "" {
		sharedRunfile := runfile.NewRunfile()
		files := []string{"run.yaml", "run_" + runtime.GOOS + ".yaml"}
		for _, file := range files {
			filepath := filepath.Join(path, file)
			if _, err := os.Stat(filepath); err == nil {
				runfile, err := readRunfile(filepath)
				if err != nil {
					return nil, errors.Wrapf(err, "error on read %s", filepath)
				}
				if err := sharedRunfile.Merge(runfile); err != nil {
					return nil, errors.Wrapf(err, "error on merge %s", filepath)
				}
			}
		}
		return sharedRunfile, nil
	}

	return readRunfile(path)
}

func readRunfile(filepath string) (*runfile.Runfile, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "error on read")
	}

	var runfile runfile.Runfile
	if err := yaml.Unmarshal(data, &runfile); err != nil {
		return nil, errors.Wrap(err, "error on unmarshal")
	}

	return &runfile, nil
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
