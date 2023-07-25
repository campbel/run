package runner

type Scope struct {
	Actions map[string]*ActionContext
	Imports map[string]*ActionContext
}

func NewScope() *Scope {
	return &Scope{
		Actions: make(map[string]*ActionContext),
		Imports: make(map[string]*ActionContext),
	}
}
