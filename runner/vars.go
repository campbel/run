package runner

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type VarContext struct {
	actionContext *ActionContext
	Value         string
	Shell         string
}

func NewVarContexts(actionContext *ActionContext, vars map[string]runfile.Var) map[string]*VarContext {
	contexts := make(map[string]*VarContext)
	for name, varCtx := range vars {
		contexts[name] = NewVarContext(actionContext, varCtx)
	}
	return contexts
}

func NewVarContext(actionContex *ActionContext, varCtx runfile.Var) *VarContext {
	return &VarContext{
		actionContext: actionContex,
		Value:         varCtx.Value,
		Shell:         varCtx.Shell,
	}
}

func (ctx *VarContext) GetValue(args any) (any, error) {
	if ctx.Shell != "" {
		shellCmd, error := varSub(args, ctx.Shell)
		if error != nil {
			return nil, errors.Wrap(error, "failed to substitute shell command")
		}
		command := exec.Command("sh", "-c", shellCmd)
		command.Env = commandEnv(ctx.actionContext.Env())
		var buffer bytes.Buffer
		command.Stdout = &buffer
		if err := command.Run(); err != nil {
			return nil, errors.Wrap(err, "failed to run shell command")
		}
		return strings.TrimSpace(buffer.String()), nil
	}
	return ctx.Value, nil
}
