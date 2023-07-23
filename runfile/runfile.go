package runfile

type Runfile struct {
	Imports   []string            `yaml:"imports"`
	Actions   map[string]Action   `yaml:"actions"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Action struct {
	Commands []string `yaml:"cmds"`
}

type Workflow struct {
	Description string   `yaml:"desc"`
	Actions     []string `yaml:"actions"`
}
