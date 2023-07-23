package runfile

type Runfile struct {
	Imports map[string]string `yaml:"imports"`
	Actions map[string]Action `yaml:"actions"`
}

func NewRunfile() *Runfile {
	return &Runfile{
		Imports: make(map[string]string),
		Actions: make(map[string]Action),
	}
}

func (r *Runfile) Merge(a *Runfile) error {
	for name, action := range a.Actions {
		r.Actions[name] = action
	}
	for name, path := range a.Imports {
		r.Imports[name] = path
	}
	return nil
}

type Action struct {
	Description string    `yaml:"desc"`
	Skip        Command   `yaml:"skip"`
	Commands    []Command `yaml:"cmds"`
}

type Command struct {
	Shell  string         `yaml:"shell"`
	Action string         `yaml:"action"`
	Args   map[string]any `yaml:"args"`
}
