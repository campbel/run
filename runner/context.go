package runner

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
)

func varSub(vars any, command string) (string, error) {
	template, err := template.New("command").Funcs(sprig.FuncMap()).Parse(command)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := template.Execute(&buffer, vars); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
