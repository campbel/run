package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/campbel/run/app"
	"github.com/campbel/run/runfile"
	"github.com/campbel/run/runner"
	"github.com/campbel/yoshi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Action  string            `yoshi:"ACTION;The action to run;default"`
	Vars    map[string]string `yoshi:"--vars,-v;The vars file to use"`
	Runfile string            `yoshi:"--runfile,-f;The runfile to use;run.yaml"`
	List    bool              `yoshi:"--list,-l;List actions"`
	TUI     bool              `yoshi:"--tui,-t;Use the TUI"`
}

func main() {
	yoshi.New("run").Run(func(options Options) error {
		runfile, err := loadRunfile(options.Runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load runfile")
		}

		if options.List {
			listActions(runfile.Actions)
			return nil
		}

		global := runner.NewGlobalContext()
		rootScope, err := loadPackage(global, runfile)
		if err != nil {
			return errors.Wrap(err, "failed to load action context")
		}

		action, ok := rootScope.Actions[options.Action]
		if !ok {
			return fmt.Errorf("no action with the name '%s'", options.Action)
		}

		if !options.TUI {
			go func() {
				for event := range global.Events() {
					fmt.Println(event)
				}
			}()
			return action.Run(options.Vars)
		}

		program := tea.NewProgram(app.NewModel(), tea.WithAltScreen())
		go func(program *tea.Program) {
			start := time.Now()
			for event := range global.Events() {
				switch event.EventType {
				case runner.EventTypeOutput:
					program.Send(app.EventMsg{
						EventType: app.EventTypeOutput,
						Message:   event.Message,
					})
				case runner.EventTypeActionFinish:
					program.Send(app.EventMsg{
						EventType: app.EventTypeActionFinish,
						Duration:  time.Since(start),
						Message:   event.Message,
					})
				case runner.EventTypeActionStart:
					program.Send(app.EventMsg{
						EventType: app.EventTypeActionStart,
						Message:   event.Message,
					})
				}
			}
			program.Quit()
		}(program)
		global.WithStdout(global).WithErrout(global)
		go func() {
			action.Run(options.Vars)
			global.Done()
		}()
		_, err = program.Run()
		return err
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

func loadPackage(global *runner.GlobalContext, runfile *runfile.Runfile, path ...string) (*runner.PackageContext, error) {
	scope := runner.NewPackageContext()

	for name, action := range runfile.Actions {
		scope.Actions[name] = runner.NewActionContext(global, scope, strings.Join(append(path, name), "."), action)
	}

	for namespace, imp := range runfile.Imports {
		runfile, err := fetchRunfile(imp)
		if err != nil {
			return nil, errors.Wrapf(err, "error on load action file %s", namespace)
		}
		s, err := loadPackage(global, runfile, append(path, namespace)...)
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
