package main

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/campbel/run/print"
	"github.com/campbel/run/runfile"
	"github.com/pkg/errors"
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
	Scope    *Scope                 `json:"-" yaml:"-"`
	Name     string                 `json:"name" yaml:"name"`
	Skip     *SkipContext           `json:"skip,omitempty" yaml:"skip,omitempty"`
	Vars     map[string]*VarContext `json:"vars,omitempty" yaml:"vars,omitempty"`
	Commands []*CommandContext      `json:"commands,omitempty" yaml:"commands,omitempty"`
}

type SkipContext struct {
	Shell   string `json:"shell,omitempty" yaml:"shell,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

func newSkipContext(skip runfile.Skip) *SkipContext {
	return &SkipContext{
		Shell:   skip.Shell,
		Message: skip.Message,
	}
}

type CommandContext struct {
	Action string            `json:"action,omitempty" yaml:"action,omitempty"`
	Shell  string            `json:"shell,omitempty" yaml:"shell,omitempty"`
	Args   map[string]string `json:"args,omitempty" yaml:"args,omitempty"`
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

func (ctx *ActionContext) Run(passedArgs map[string]string) error {

	info := print.StartInfoContext()

	// Variables cascade
	// The defaults are input to args
	// The defaults and args are input to vars
	input := map[string]any{
		"os":   runtime.GOOS,
		"OS":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"ARCH": runtime.GOARCH,
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

	info("running action %s", ctx.Name)
	defer info("finished action %s", ctx.Name)
	if ctx.Skip.Shell != "" {
		subbedCommand, err := varSub(input, ctx.Skip.Shell)
		if err != nil {
			return err
		}
		command := exec.Command("sh", "-c", subbedCommand)
		if err := command.Run(); err == nil {
			print.Notice(" - skipping - %s", ctx.Skip.Message)
			return nil
		}
	}

	for _, cmd := range ctx.Commands {
		if cmd.Shell != "" {
			subbedCommand, err := varSub(input, cmd.Shell)
			if err != nil {
				return err
			}
			command := exec.Command("sh", "-c", subbedCommand)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			command.Stdin = os.Stdin
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

type VarContext struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	Shell string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func newVarContexts(vars map[string]runfile.Var) map[string]*VarContext {
	contexts := make(map[string]*VarContext)
	for name, varCtx := range vars {
		contexts[name] = newVarContext(varCtx)
	}
	return contexts
}

func newVarContext(varCtx runfile.Var) *VarContext {
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

func varSub(vars any, command string) (string, error) {
	template, err := template.New("command").Funcs(sprig.FuncMap()).Parse(command)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := template.Execute(&buffer, vars); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
