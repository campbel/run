package runner

import (
	"runtime"
	"strings"

	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type ActionContext struct {
	Global       *GlobalContext
	Package      *PackageContext
	Dependencies []string
	Skip         *SkipContext
	Vars         map[string]*VarContext
	env          map[string]string
	Commands     []*CommandContext
}

func NewActionContext(global *GlobalContext, pkg *PackageContext, action runfile.Action) *ActionContext {
	actionContext := &ActionContext{
		Global:       global,
		Package:      pkg,
		Dependencies: action.Dependencies,
		env:          action.Env,
	}

	actionContext.Skip = NewSkipContext(actionContext, action.Skip)
	actionContext.Vars = NewVarContexts(actionContext, action.Vars)
	actionContext.Commands = NewCommandContexts(actionContext, action.Commands)

	return actionContext
}

func (ctx *ActionContext) Run(passedArgs map[string]string) error {
	for _, dep := range ctx.Dependencies {
		if strings.Contains(dep, ".") {
			parts := strings.SplitN(dep, ".", 2)
			pkg, action := parts[0], parts[1]
			if packageCtx, exists := ctx.Package.Imports[pkg]; exists {
				if err := packageCtx.Run(action, passedArgs); err != nil {
					return errors.Wrap(err, "error running action")
				}
			} else {
				return errors.Errorf("no package with the name '%s'", pkg)
			}
		} else {
			if action, exists := ctx.Package.Actions[dep]; exists {
				if err := action.Run(passedArgs); err != nil {
					return err
				}
			} else {
				return errors.Errorf("no action with the name '%s'", dep)
			}
		}
	}

	// Variables cascade
	// The defaults are input to args
	// The defaults and args are input to vars
	input := map[string]any{
		"os":      runtime.GOOS,
		"OS":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"ARCH":    runtime.GOARCH,
		"pkg_dir": ctx.Package.Dir,
		"PKG_DIR": ctx.Package.Dir,
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

	if skip, err := ctx.Skip.Run(input); skip || err != nil {
		return err
	}

	for _, cmd := range ctx.Commands {
		if err := cmd.Run(input); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *ActionContext) Env() map[string]string {
	merged := make(map[string]string)
	for name, value := range ctx.Package.Env() {
		merged[name] = value
	}
	for name, value := range ctx.env {
		merged[name] = value
	}
	return merged
}
