package runner

import (
	"os/exec"
	"strings"

	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
)

type CommandContext struct {
	actionContext *ActionContext

	Action string
	Shell  string
	Args   map[string]string
}

func NewCommandContexts(actionCtx *ActionContext, commands []runfile.Command) []*CommandContext {
	var contexts []*CommandContext
	for _, command := range commands {
		contexts = append(contexts, NewCommandContext(actionCtx, command))
	}
	return contexts
}

func NewCommandContext(actionCtx *ActionContext, command runfile.Command) *CommandContext {
	return &CommandContext{
		actionContext: actionCtx,

		Action: command.Action,
		Shell:  command.Shell,
		Args:   command.Args,
	}
}

func (cmd *CommandContext) Run(input map[string]any) error {
	if cmd.Shell != "" {
		subbedCommand, err := varSub(input, cmd.Shell)
		if err != nil {
			return err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		command.Stdout = cmd.actionContext.Global.out
		command.Stderr = cmd.actionContext.Global.err
		command.Stdin = cmd.actionContext.Global.in
		if err := command.Run(); err != nil {
			return err
		}
		return nil
	}
	if cmd.Action != "" {
		if strings.Contains(cmd.Action, ".") {
			parts := strings.SplitN(cmd.Action, ".", 2)
			pkg, action := parts[0], parts[1]
			if packageCtx, exists := cmd.actionContext.Package.Imports[pkg]; exists {
				if err := packageCtx.Run(action, cmd.Args); err != nil {
					return errors.Wrap(err, "error running action")
				}
			} else {
				return errors.Errorf("no package with the name '%s'", pkg)
			}
		} else {
			if action, exists := cmd.actionContext.Package.Actions[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			} else {
				return errors.Errorf("no action with the name '%s'", cmd.Action)
			}
		}
	}
	return nil
}
