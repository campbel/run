package runfile

type Runfile struct {
	Imports   map[string]string   `yaml:"imports"`
	Actions   map[string]Action   `yaml:"actions"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Action struct {
	Description string    `yaml:"desc"`
	Commands    []Command `yaml:"cmds"`
}

type Command struct {
	Shell  string         `yaml:"shell"`
	Action string         `yaml:"action"`
	Args   map[string]any `yaml:"args"`
}

type Workflow struct {
	Description string           `yaml:"desc"`
	Actions     []WorkflowAction `yaml:"actions"`
}

type WorkflowAction struct {
	Action string         `yaml:"action"`
	Args   map[string]any `yaml:"args"`
}
