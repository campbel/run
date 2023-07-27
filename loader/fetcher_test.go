package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/campbel/run/runfile"
	"github.com/stretchr/testify/assert"
)

func TestGoGetter_Fetch(t *testing.T) {
	t.Run("fetches runfile", func(t *testing.T) {
		gg := NewGoGetter()
		gg.pwd = t.TempDir()

		rf, err := gg.Fetch("github.com/campbel/run/loader/testdata/simple")
		assert.NoError(t, err)

		expected := &runfile.Runfile{
			Imports: make(map[string]string),
			Actions: map[string]runfile.Action{
				"test": {
					Commands: []runfile.Command{
						{Shell: "echo \"hello world\""},
					},
				},
				"test-os": {
					Commands: []runfile.Command{
						{Shell: "echo \"hello " + runtime.GOOS + "\""},
					},
				},
			},
		}
		assert.Equal(t, expected, rf)
	})

	t.Run("fetches platform specific runfile", func(t *testing.T) {
		gg := NewGoGetter()
		gg.pwd = t.TempDir()

		err := os.MkdirAll(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple"), 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple", "run_"+runtime.GOOS+".yaml"), []byte(`
actions:
  test:
    cmds:
      - echo "hello `+runtime.GOOS+`"
`), 0644)
		assert.NoError(t, err)

		rf, err := gg.Fetch("github.com/campbel/run/loader/testdata/simple")
		assert.NoError(t, err)

		expected := &runfile.Runfile{
			Imports: make(map[string]string),
			Actions: map[string]runfile.Action{
				"test": {
					Commands: []runfile.Command{
						{Shell: "echo \"hello " + runtime.GOOS + "\""},
					},
				},
			},
		}
		assert.Equal(t, expected, rf)
	})

	t.Run("returns error when runfile is not found", func(t *testing.T) {
		gg := NewGoGetter()
		gg.pwd = t.TempDir()

		_, err := gg.Fetch("github.com/campbel/run/does/not/exist")
		assert.Error(t, err)
	})

	t.Run("invalid yaml error", func(t *testing.T) {
		gg := NewGoGetter()
		gg.pwd = t.TempDir()

		err := os.MkdirAll(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple"), 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple", "run_darwin.yaml"), []byte(`{}}"`), 0644)
		assert.NoError(t, err)

		rf, err := gg.Fetch("github.com/campbel/run/loader/testdata/simple")
		assert.Error(t, err)
		assert.Nil(t, rf)
	})

	t.Run("invalid yaml error", func(t *testing.T) {
		gg := NewGoGetter()
		gg.readFile = func(string) ([]byte, error) {
			return nil, os.ErrNotExist
		}
		gg.pwd = t.TempDir()

		err := os.MkdirAll(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple"), 0755)
		assert.NoError(t, err)

		err = os.WriteFile(filepath.Join(gg.pwd, ".run", "imports", "github.com/campbel/run/loader/testdata/simple", "run.yaml"), []byte(``), 0644)
		assert.NoError(t, err)

		rf, err := gg.Fetch("github.com/campbel/run/loader/testdata/simple")
		assert.Error(t, err)
		assert.Nil(t, rf)
	})
}
