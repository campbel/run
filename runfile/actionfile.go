package runfile

type Actionfile struct {
	Imports map[string]string `yaml:"imports"`
	Actions map[string]Action `yaml:"actions"`
}
