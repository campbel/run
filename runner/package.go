package runner

import "github.com/pkg/errors"

type PackageContext struct {
	Global  *GlobalContext
	Actions map[string]*ActionContext
	Imports map[string]*PackageContext
}

func NewPackageContext(global *GlobalContext) *PackageContext {
	return &PackageContext{
		Global:  global,
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
