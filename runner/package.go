package runner

import (
	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type PackageContext struct {
	Global  *GlobalContext
	Dir     string
	env     map[string]string
	Actions map[string]*ActionContext
	Imports map[string]*PackageContext
}

func NewPackageContext(global *GlobalContext, rf *runfile.Runfile) *PackageContext {
	return &PackageContext{
		Global:  global,
		Dir:     rf.Dir(),
		env:     rf.Env,
		Actions: make(map[string]*ActionContext),
		Imports: make(map[string]*PackageContext),
	}
}

func (ctx *PackageContext) Run(actionName string, passedArgs map[string]string) error {
	if action, exists := ctx.Actions[actionName]; exists {
		return action.Run(passedArgs)
	}
	return errors.Errorf("no action with the name '%s'", actionName)
}

func (ctx *PackageContext) Env() map[string]string {
	return ctx.env
}
