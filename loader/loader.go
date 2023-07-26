package loader

import "github.com/campbel/run/runfile"

type Loader struct {
	root     *runfile.Runfile
	packages map[string]*runfile.Runfile

	fetcher Fetcher
}

func NewLoader(root *runfile.Runfile, fetcher Fetcher) *Loader {
	return &Loader{
		root:     root,
		packages: make(map[string]*runfile.Runfile),
		fetcher:  fetcher,
	}
}

func (l *Loader) Load() {
	for _, pkg := range l.root.Imports {
		l.loadPackage(pkg)
	}
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
