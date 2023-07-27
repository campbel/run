package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/campbel/run/runfile"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

type Fetcher interface {
	Fetch(uri string) (*runfile.Runfile, error)
}

type GoGetter struct {
	client       *getter.Client
	readFile     func(string) ([]byte, error)
	filepathGlob func(string) ([]string, error)
	pwd          string
}

func NewGoGetter() *GoGetter {
	return &GoGetter{
		client:       &getter.Client{},
		readFile:     os.ReadFile,
		filepathGlob: filepath.Glob,
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

	var filepaths []string
	files, err := g.filepathGlob(filepath.Join(dst, "*.yaml"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_"+runtime.GOOS+".yaml") || strings.HasSuffix(file, "run.yaml") {
			filepaths = append(filepaths, file)
		}
	}

	sharedRunfile := runfile.NewRunfile().WithDir(dst)
	for _, filepath := range filepaths {
		if _, err := os.Stat(filepath); err == nil {
			data, err := g.readFile(filepath)
			if err != nil {
				return nil, errors.Wrapf(err, "error on read %s", filepath)
			}
			rf, err := runfile.Unmarshal(data)
			if err != nil {
				return nil, errors.Wrapf(err, "error on unmarshal %s", filepath)
			}
			sharedRunfile = runfile.Merge(sharedRunfile, rf)
		}
	}
	return sharedRunfile, nil
}

func (g *GoGetter) path(imp string) string {
	return filepath.Join(g.pwd, ".run", "imports", imp)
}
