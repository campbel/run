package runner

import (
	"os/exec"
	"runtime"

	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type ActionContext struct {
	Global       *GlobalContext
	Package      *PackageContext
	Name         string
	Dependencies []string
	Skip         *SkipContext
	Vars         map[string]*VarContext
	Commands     []*CommandContext
}

func NewActionContext(global *GlobalContext, pkg *PackageContext, name string, action runfile.Action) *ActionContext {
	return &ActionContext{
		Global:       global,
		Package:      pkg,
		Name:         name,
		Dependencies: action.Dependencies,
		Skip:         NewSkipContext(action.Skip),
		Vars:         NewVarContexts(action.Vars),
		Commands:     NewCommandContexts(action.Commands),
	}
}

func (ctx *ActionContext) Run(passedArgs map[string]string) error {
	ctx.Global.Emit(Event{
		EventType: EventTypeActionStart,
		Message:   ctx.Name,
	})
	defer func() {
		ctx.Global.Emit(Event{
			EventType: EventTypeActionFinish,
			Message:   ctx.Name,
		})
	}()

	for _, dep := range ctx.Dependencies {
		if action, exists := ctx.Package.Actions[dep]; exists {
			if err := action.Run(passedArgs); err != nil {
				return err
			}
			continue
		}
		if action, exists := ctx.Package.Imports[dep]; exists {
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

	if ctx.Skip.Shell != "" {
		subbedCommand, err := varSub(input, ctx.Skip.Shell)
		if err != nil {
			return err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		if err := command.Run(); err == nil {
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
			command.Stdout = ctx.Global.out
			command.Stderr = ctx.Global.err
			command.Stdin = ctx.Global.in
			if err := command.Run(); err != nil {
				return err
			}
			continue
		}
		if cmd.Action != "" {
			if action, exists := ctx.Package.Actions[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			} else if action, exists := ctx.Package.Imports[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			}
			continue
		}
	}
	return nil
}
