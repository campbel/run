package main

import (
	"os"
	"os/exec"

	"github.com/campbel/run/runfile"
)

type Scope struct {
	Actions map[string]*ActionContext
	Imports map[string]*ActionContext
}

func NewScope() *Scope {
	return &Scope{
		Actions: make(map[string]*ActionContext),
		Imports: make(map[string]*ActionContext),
	}
}

type ActionContext struct {
	Scope    *Scope            `json:"-" yaml:"-"`
	Commands []*CommandContext `json:"commands" yaml:"commands,omitempty"`
}

type CommandContext struct {
	Action string `json:"action,omitempty" yaml:"action,omitempty"`
	Shell  string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func newCommandContexts(commands []runfile.Command) []*CommandContext {
	var contexts []*CommandContext
	for _, command := range commands {
		contexts = append(contexts, newCommandContext(command))
	}
	return contexts
}

func newCommandContext(command runfile.Command) *CommandContext {
	return &CommandContext{
		Action: command.Action,
		Shell:  command.Shell,
	}
}

func (ctx *ActionContext) Run() error {
	for _, cmd := range ctx.Commands {
		if cmd.Shell != "" {
			command := exec.Command("sh", "-c", cmd.Shell)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				return err
			}
			continue
		}
		if cmd.Action != "" {
			if action, exists := ctx.Scope.Actions[cmd.Action]; exists {
				if err := action.Run(); err != nil {
					return err
				}
			} else if action, exists := ctx.Scope.Imports[cmd.Action]; exists {
				if err := action.Run(); err != nil {
					return err
				}
			}
			continue
		}
	}
	return nil
}
