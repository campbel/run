package runfile

type Runfile struct {
	dir     string
	env     map[string]string `yaml:"env" mapstructure:"env"`
	Imports map[string]string `yaml:"imports" mapstructure:"imports"`
	Actions map[string]Action `yaml:"actions" mapstructure:"actions"`
}

func NewRunfile() *Runfile {
	return &Runfile{
		Imports: make(map[string]string),
		Actions: make(map[string]Action),
	}
}

func (r *Runfile) WithDir(dir string) *Runfile {
	r.dir = dir
	return r
}

func (r *Runfile) Dir() string {
	if r == nil {
		return ""
	}
	return r.dir
}

func (r *Runfile) Env() map[string]string {
	if r == nil {
		return make(map[string]string)
	}
	return r.env
}

func Merge(rfs ...*Runfile) *Runfile {
	rf := NewRunfile()
	for _, r := range rfs {
		if r == nil {
			continue
		}
		for name, action := range r.Actions {
			rf.Actions[name] = action
		}
		for name, path := range r.Imports {
			rf.Imports[name] = path
		}
	}
	return rf
}

type Action struct {
	Description  string            `yaml:"desc" mapstructure:"desc"`
	Dependencies []string          `yaml:"deps" mapstructure:"deps"`
	Skip         Skip              `yaml:"skip" mapstructure:"skip"`
	Vars         map[string]Var    `yaml:"vars" mapstructure:"vars"`
	Env          map[string]string `yaml:"env"  mapstructure:"env"`
	Commands     []Command         `yaml:"cmds" mapstructure:"cmds"`
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
