package runner

import (
	"os/exec"

	"github.com/campbel/run/runfile"
)

type SkipContext struct {
	actionContext *ActionContext
	Shell         string
	Message       string
}

func NewSkipContext(actionContext *ActionContext, skip runfile.Skip) *SkipContext {
	return &SkipContext{
		actionContext: actionContext,
		Shell:         skip.Shell,
		Message:       skip.Message,
	}
}

func (ctx *SkipContext) Run(vars any) (bool, error) {
	if ctx.Shell != "" {
		subbedCommand, err := varSub(vars, ctx.Shell)
		if err != nil {
			return false, err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		command.Env = commandEnv(ctx.actionContext.Env())
		return command.Run() == nil, nil
	}
	return false, nil
}
