package main

import (
	"bytes"
	"os"
	"os/exec"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/campbel/run/print"
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
	Name     string            `json:"name" yaml:"name"`
	Skip     *CommandContext   `json:"skip,omitempty" yaml:"skip,omitempty"`
	Commands []*CommandContext `json:"commands,omitempty" yaml:"commands,omitempty"`
}

type CommandContext struct {
	Action string         `json:"action,omitempty" yaml:"action,omitempty"`
	Shell  string         `json:"shell,omitempty" yaml:"shell,omitempty"`
	Args   map[string]any `json:"args,omitempty" yaml:"args,omitempty"`
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
		Args:   command.Args,
	}
}

func (ctx *ActionContext) Run(args map[string]any) error {
	print.Info("running action %s", ctx.Name)
	if ctx.Skip.Shell != "" {
		subbedCommand, err := argSub(args, ctx.Skip.Shell)
		if err != nil {
			return err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		if err := command.Run(); err == nil {
			print.Notice("skipping action %s", ctx.Name)
			return nil
		}
	}

	for _, cmd := range ctx.Commands {
		if cmd.Shell != "" {
			subbedCommand, err := argSub(args, cmd.Shell)
			if err != nil {
				return err
			}
			command := exec.Command("sh", "-c", subbedCommand)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				return err
			}
			continue
		}
		if cmd.Action != "" {
			if action, exists := ctx.Scope.Actions[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			} else if action, exists := ctx.Scope.Imports[cmd.Action]; exists {
				if err := action.Run(cmd.Args); err != nil {
					return err
				}
			}
			continue
		}
	}
	return nil
}

func argSub(args map[string]any, command string) (string, error) {
	template, err := template.New("command").Funcs(sprig.FuncMap()).Parse(command)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := template.Execute(&buffer, args); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
