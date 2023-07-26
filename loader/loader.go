package loader

import (
	"github.com/campbel/run/runfile"
	"github.com/campbel/run/runner"
)

type Loader struct {
	main        *runfile.Runfile
	mainContext *runner.PackageContext

	packages        map[string]*runfile.Runfile
	packagesContext map[string]*runner.PackageContext

	fetcher Fetcher
}

func NewLoader(root *runfile.Runfile, fetcher Fetcher) *Loader {
	return &Loader{
		main:            root,
		packages:        make(map[string]*runfile.Runfile),
		packagesContext: make(map[string]*runner.PackageContext),

		fetcher: fetcher,
	}
}

func (l *Loader) Load() *runner.PackageContext {
	for _, pkg := range l.main.Imports {
		l.loadPackage(pkg)
	}

	return l.loadPackageCtx(runner.NewGlobalContext(), l.main)
}

func (l *Loader) loadPackageCtx(global *runner.GlobalContext, rf *runfile.Runfile) *runner.PackageContext {
	pkg := runner.NewPackageContext()

	for name, action := range rf.Actions {
		pkg.Actions[name] = runner.NewActionContext(global, pkg, name, action)
	}

	for name, uri := range rf.Imports {
		if _, ok := l.packagesContext[uri]; !ok {
			l.packagesContext[uri] = l.loadPackageCtx(global, l.packages[uri])
		}
		pkg.Imports[name] = l.packagesContext[uri]
	}

	return pkg
}

func (l *Loader) loadPackage(uri string) error {
	if _, ok := l.packages[uri]; ok {
		return nil
	}
	runfile, err := l.fetcher.Fetch(uri)
	if err != nil {
		return err
	}
	l.packages[uri] = runfile
	for _, pkg := range runfile.Imports {
		if err := l.loadPackage(pkg); err != nil {
			return err
		}
	}
	return nil
}
