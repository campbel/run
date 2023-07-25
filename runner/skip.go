package runner

import "github.com/campbel/run/runfile"

type SkipContext struct {
	Shell   string
	Message string
}

func NewSkipContext(skip runfile.Skip) *SkipContext {
	return &SkipContext{
		Shell:   skip.Shell,
		Message: skip.Message,
	}
}
