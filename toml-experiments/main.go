package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type attributedStringSlice struct {
	env        []string
	attributes struct {
		append bool
	}
}

func (ts *attributedStringSlice) UnmarshalTOML(data interface{}) error {
	iFaceSlice, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unable to cast to interface array: %v", data)
	}
	for _, x := range iFaceSlice {
		kind := reflect.ValueOf(x).Kind()
		switch kind {
		case reflect.String:
			ts.env = append(ts.env, fmt.Sprintf("%v", x))
		case reflect.Map:
			attrMap, ok := x.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to cast to map of interfaces: %v", data)
			}
			for k, v := range attrMap {
				switch k {
				case "append":
					boolVal, ok := v.(bool)
					if !ok {
						return fmt.Errorf("unable to cast to bool: %v", k)
					}
					ts.attributes.append = boolVal
				default:
					return fmt.Errorf("invalid key %q in map: %v", k, attrMap)
				}
			}
		default:
			return fmt.Errorf("unsupported kind %v: %v", kind, x)
		}
	}
	return nil
}

type configTOML struct {
	Env attributedStringSlice `toml:"env,omitempty"`
}

func main() {
	blob :=`
env=["a", "b", "c", {append= true}]
`
	var data configTOML
	_, err := toml.Decode(blob, &data)
	if err != nil {
		logrus.Errorf("%v", err)
		os.Exit(1)
	}

	logrus.Errorf("Finished loading:")
	logrus.Errorf(" - env: %v", data.Env.env)
	logrus.Errorf(" - attributes: %v", data.Env.attributes)
}
