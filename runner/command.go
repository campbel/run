package runner

import "github.com/campbel/run/runfile"

type CommandContext struct {
	Action string
	Shell  string
	Args   map[string]string
}

func NewCommandContexts(commands []runfile.Command) []*CommandContext {
	var contexts []*CommandContext
	for _, command := range commands {
		contexts = append(contexts, NewCommandContext(command))
	}
	return contexts
}

func NewCommandContext(command runfile.Command) *CommandContext {
	return &CommandContext{
		Action: command.Action,
		Shell:  command.Shell,
		Args:   command.Args,
	}
}
