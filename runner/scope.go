package runner

type PackageContext struct {
	Actions map[string]*ActionContext
	Imports map[string]*ActionContext
}

func NewScope() *PackageContext {
	return &PackageContext{
		Actions: make(map[string]*ActionContext),
		Imports: make(map[string]*ActionContext),
	}
}
