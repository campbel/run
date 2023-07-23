package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/campbel/run/runfile"
)

type Scope struct {
	Actions map[string]*ActionContext
}

func NewScope() *Scope {
	return &Scope{
		Actions: make(map[string]*ActionContext),
	}
}

type ActionContext struct {
	Scope    *Scope `json:"-"`
	Commands []*CommandContext
}

type CommandContext struct {
	Action string
	Shell  string
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
			parts := strings.Split(cmd.Shell, " ")
			command := exec.Command(parts[0], parts[1:]...)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			return command.Run()
		}
		if cmd.Action != "" {
			return ctx.Scope.Actions[cmd.Action].Run()
		}
	}
	return nil
}
