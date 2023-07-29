package runner

import "github.com/campbel/run/runfile"

type SkipContext struct {
	actionContext *ActionContext
	Shell         string
	Message       string
}

func NewSkipContext(actionContext *ActionContext, skip runfile.Skip) *SkipContext {
	return &SkipContext{
		actionContext: actionContext,
		Shell:         skip.Shell,
		Message:       skip.Message,
	}
}
