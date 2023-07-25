package runner

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type VarContext struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	Shell string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func NewVarContexts(vars map[string]runfile.Var) map[string]*VarContext {
	contexts := make(map[string]*VarContext)
	for name, varCtx := range vars {
		contexts[name] = NewVarContext(varCtx)
	}
	return contexts
}

func NewVarContext(varCtx runfile.Var) *VarContext {
	return &VarContext{
		Value: varCtx.Value,
		Shell: varCtx.Shell,
	}
}

func (ctx *VarContext) GetValue(args any) (any, error) {
	if ctx.Shell != "" {
		shellCmd, error := varSub(args, ctx.Shell)
		if error != nil {
			return nil, errors.Wrap(error, "failed to substitute shell command")
		}
		command := exec.Command("sh", "-c", shellCmd)
		var buffer bytes.Buffer
		command.Stdout = &buffer
		if err := command.Run(); err != nil {
			return nil, errors.Wrap(err, "failed to run shell command")
		}
		return strings.TrimSpace(buffer.String()), nil
	}
	return ctx.Value, nil
}
