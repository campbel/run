package runfile

type Runfile struct {
	Imports map[string]string `yaml:"imports" mapstructure:"imports"`
	Actions map[string]Action `yaml:"actions" mapstructure:"actions"`
}

func NewRunfile() *Runfile {
	return &Runfile{
		Imports: make(map[string]string),
		Actions: make(map[string]Action),
	}
}

func (r *Runfile) Merge(a *Runfile) {
	for name, action := range a.Actions {
		r.Actions[name] = action
	}
	for name, path := range a.Imports {
		r.Imports[name] = path
	}
}

type Action struct {
	Description  string         `yaml:"desc" mapstructure:"desc"`
	Dependencies []string       `yaml:"deps" mapstructure:"deps"`
	Skip         Skip           `yaml:"skip" mapstructure:"skip"`
	Vars         map[string]Var `yaml:"vars" mapstructure:"vars"`
	Commands     []Command      `yaml:"cmds" mapstructure:"cmds"`
}

type Skip struct {
	Shell   string `yaml:"shell" mapstructure:"shell"`
	Message string `yaml:"msg" mapstructure:"msg"`
}

type Command struct {
	Shell  string            `yaml:"shell" mapstructure:"shell"`
	Action string            `yaml:"action" mapstructure:"action"`
	Args   map[string]string `yaml:"args" mapstructure:"args"`
}

type Var struct {
	Value string `yaml:"value"`
	Shell string `yaml:"shell"`
}
