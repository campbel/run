package loader

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/campbel/run/runfile"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

type Fetcher interface {
	Fetch(uri string) (*runfile.Runfile, error)
}

type GoGetter struct {
	client   *getter.Client
	readFile func(string) ([]byte, error)
	pwd      string
}

func NewGoGetter() *GoGetter {
	return &GoGetter{
		client:   &getter.Client{},
		readFile: os.ReadFile,
	}
}

func (g *GoGetter) Fetch(src string) (*runfile.Runfile, error) {
	dst := g.path(src)
	if _, err := os.Stat(dst); err != nil {
		if err := (&getter.Client{
			Src:  src,
			Dst:  g.path(src),
			Pwd:  g.pwd,
			Mode: getter.ClientModeAny,
		}).Get(); err != nil {
			return nil, err
		}
	}

	files := []string{"run.yaml", "run_" + runtime.GOOS + ".yaml"}
	sharedRunfile := runfile.NewRunfile()
	for _, file := range files {
		filepath := filepath.Join(dst, file)
		if _, err := os.Stat(filepath); err == nil {
			data, err := g.readFile(filepath)
			if err != nil {
				return nil, errors.Wrapf(err, "error on read %s", filepath)
			}
			runfile, err := runfile.Unmarshal(data)
			if err != nil {
				return nil, errors.Wrapf(err, "error on unmarshal %s", filepath)
			}
			sharedRunfile.Merge(runfile)
		}
	}
	return sharedRunfile, nil
}

func (g *GoGetter) path(imp string) string {
	return filepath.Join(g.pwd, ".run", "imports", imp)
}
