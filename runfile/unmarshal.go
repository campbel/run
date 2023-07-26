package runfile

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func Unmarshal(content []byte) (*Runfile, error) {
	var a any
	if err := yaml.Unmarshal(content, &a); err != nil {
		return nil, errors.Wrap(err, "qqqqq unmarshal runfile")
	}

	var runfile Runfile
	return &runfile, errors.Wrap(decode(a, &runfile), "qqqqq decode runfile")
}

// decode uses mapstructure to decode the given any into the given runfile.
func decode(a any, rf *Runfile) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: decodeHook,
		Result:     rf,
	})
	if err != nil {
		return errors.Wrap(err, "error creating decoder")
	}

	return decoder.Decode(a)
}

func decodeHook(fromType, toType reflect.Type, from any) (any, error) {
	switch fromType {
	case reflect.TypeOf(""):
		switch toType {
		case reflect.TypeOf(Skip{}):
			return Skip{Shell: from.(string)}, nil
		case reflect.TypeOf(Var{}):
			return Var{Shell: from.(string)}, nil
		case reflect.TypeOf(Command{}):
			return Command{Shell: from.(string)}, nil
		}
	}
	return from, nil
}
