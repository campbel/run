package runner

import "github.com/campbel/run/runfile"

type SkipContext struct {
	Shell   string `json:"shell,omitempty" yaml:"shell,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

func NewSkipContext(skip runfile.Skip) *SkipContext {
	return &SkipContext{
		Shell:   skip.Shell,
		Message: skip.Message,
	}
}
