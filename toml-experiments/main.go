package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type attributedStringSlice struct { // A "mixed-type array" in TOML.
	slice      []string
	attributes struct { // Using a struct allows for adding more attributes in the feature.
		append bool
	}
}

func (ts *attributedStringSlice) UnmarshalTOML(data interface{}) error {
	iFaceSlice, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unable to cast to interface array: %v", data)
	}
	for _, x := range iFaceSlice { // Iterate over each item in the slice.
		kind := reflect.ValueOf(x).Kind()
		switch kind {
		case reflect.String: // Strings are directly appended to the slice.
			ts.env = append(ts.slice, fmt.Sprintf("%v", x))
		case reflect.Map: // The attrivute struct is represented as a map.
			attrMap, ok := x.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to cast to map of interfaces: %v", data)
			}
			for k, v := range attrMap { // Iterate over all _supported_ keys.
				switch k {
				case "append":
					boolVal, ok := v.(bool)
					if !ok {
						return fmt.Errorf("unable to cast to bool: %v", k)
					}
					ts.attributes.append = boolVal
				default: // Unsupported map key.
					return fmt.Errorf("unsupported key %q in map: %v", k, attrMap)
				}
			}
		default: // Unsupported item.
			return fmt.Errorf("unsupported item in attributed string slice %v: %v", kind, x)
		}
	}
	return nil
}

type configTOML struct {
	Env attributedStringSlice `toml:"env,omitempty"`
}

func main() {
	// TOML supports so-called "mixed-type arrays".  A feature that we can
	// exploit to attribute string slices.  In it's curent form attributes
	// can be specified in the struct notation `{ ... }`.
	blob := `
env=["a", "b", "c", {append=true}]
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
