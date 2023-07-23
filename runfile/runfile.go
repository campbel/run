package runfile

type Runfile struct {
	Imports   map[string]string   `yaml:"imports"`
	Actions   map[string]Action   `yaml:"actions"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Action struct {
	Description string   `yaml:"desc"`
	Commands    []string `yaml:"cmds"`
}

type Workflow struct {
	Description string   `yaml:"desc"`
	Actions     []string `yaml:"actions"`
}
