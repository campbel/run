// BEGIN: jh3d8f4g9d3h
package loader

import (
	"errors"
	"testing"

	"github.com/campbel/run/runfile"
	"github.com/stretchr/testify/assert"
)

type mockFetcher struct {
	fetch func(uri string) (*runfile.Runfile, error)
}

func (m *mockFetcher) Fetch(uri string) (*runfile.Runfile, error) {
	return m.fetch(uri)
}

func TestLoader_Load(t *testing.T) {
	root := &runfile.Runfile{
		Imports: map[string]string{
			"pkg1": "github.com/pkg1",
			"pkg2": "github.com/pkg2",
		},
	}
	pkg1 := &runfile.Runfile{
		Imports: map[string]string{
			"pkg2": "github.com/pkg2",
			"pkg3": "github.com/pkg3",
		},
	}
	pkg2 := &runfile.Runfile{}
	pkg3 := &runfile.Runfile{}

	tests := []struct {
		name     string
		fetcher  Fetcher
		expected map[string]*runfile.Runfile
		err      error
	}{
		{
			name: "success",
			fetcher: &mockFetcher{
				fetch: func(uri string) (*runfile.Runfile, error) {
					switch uri {
					case "github.com/pkg1":
						return pkg1, nil
					case "github.com/pkg2":
						return pkg2, nil
					case "github.com/pkg3":
						return pkg3, nil
					default:
						return nil, errors.New("unknown package")
					}
				},
			},
			expected: map[string]*runfile.Runfile{
				"github.com/pkg1": pkg1,
				"github.com/pkg2": pkg2,
				"github.com/pkg3": pkg3,
			},
			err: nil,
		},
		{
			name: "fetch error",
			fetcher: &mockFetcher{
				fetch: func(uri string) (*runfile.Runfile, error) {
					return nil, errors.New("fetch error")
				},
			},
			expected: make(map[string]*runfile.Runfile),
			err:      errors.New("fetch error"),
		},
		{
			name: "fetch error, 2nd level",
			fetcher: &mockFetcher{
				fetch: func(uri string) (*runfile.Runfile, error) {
					if uri == "github.com/pkg1" {
						return pkg1, nil
					}
					return nil, errors.New("fetch error")
				},
			},
			expected: map[string]*runfile.Runfile{
				"github.com/pkg1": pkg1,
			},
			err: errors.New("fetch error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			l := NewLoader(root, tt.fetcher)
			l.Load()
			assert.Equal(l.packages, tt.expected)
		})
	}
}
