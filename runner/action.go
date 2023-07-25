package runner

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/campbel/run/print"
	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type ActionContext struct {
	Scope        *Scope                 `json:"-" yaml:"-"`
	Name         string                 `json:"name" yaml:"name"`
	Dependencies []string               `json:"deps,omitempty" yaml:"deps,omitempty"`
	Skip         *SkipContext           `json:"skip,omitempty" yaml:"skip,omitempty"`
	Vars         map[string]*VarContext `json:"vars,omitempty" yaml:"vars,omitempty"`
	Commands     []*CommandContext      `json:"commands,omitempty" yaml:"commands,omitempty"`
}

func NewActionContext(scope *Scope, name string, action runfile.Action) *ActionContext {
	return &ActionContext{
		Scope:        scope,
		Name:         name,
		Dependencies: action.Dependencies,
		Skip:         NewSkipContext(action.Skip),
		Vars:         NewVarContexts(action.Vars),
		Commands:     NewCommandContexts(action.Commands),
	}
}

func (ctx *ActionContext) Run(passedArgs map[string]string) error {

	info := print.StartInfoContext()

	for _, dep := range ctx.Dependencies {
		info("running dependency %s", dep)
		if action, exists := ctx.Scope.Actions[dep]; exists {
			if err := action.Run(passedArgs); err != nil {
				return err
			}
			continue
		}
		if action, exists := ctx.Scope.Imports[dep]; exists {
			if err := action.Run(passedArgs); err != nil {
				return err
			}
			continue
		}
		return errors.Errorf("no action with the name '%s'", dep)
	}

	// Variables cascade
	// The defaults are input to args
	// The defaults and args are input to vars
	input := map[string]any{
		"os":   runtime.GOOS,
		"OS":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"ARCH": runtime.GOARCH,
	}

	args := make(map[string]any)
	for name, arg := range passedArgs {
		subbedArg, err := varSub(input, arg)
		if err != nil {
			return err
		}
		args[name] = subbedArg
	}
	input["args"] = args
	input["ARGS"] = args

	vars := make(map[string]any)
	for name, varCtx := range ctx.Vars {
		value, err := varCtx.GetValue(input)
		if err != nil {
			return errors.Wrap(err, "error geting value for var")
		}
		vars[name] = value
		vars[name] = value
	}
	input["vars"] = vars
	input["VARS"] = vars

	info("running action %s", ctx.Name)
	defer info("finished action %s", ctx.Name)
	if ctx.Skip.Shell != "" {
		subbedCommand, err := varSub(input, ctx.Skip.Shell)
		if err != nil {
			return err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		if err := command.Run(); err == nil {
			print.Notice(" - skipping - %s", ctx.Skip.Message)
			return nil
		}
	}

	for _, cmd := range ctx.Commands {
		if cmd.Shell != "" {
			subbedCommand, err := varSub(input, cmd.Shell)
			if err != nil {
				return err
			}
			command := exec.Command("sh", "-c", subbedCommand)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			command.Stdin = os.Stdin
			if err := command.Run(); err != nil {
				return err
			}
			continue
		}
		if cmd.Action != "" {
			if action, exists := ctx.Scope.Actions[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			} else if action, exists := ctx.Scope.Imports[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			}
			continue
		}
	}
	return nil
}
