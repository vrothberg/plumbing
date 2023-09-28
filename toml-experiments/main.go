package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

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
	logrus.Errorf(" - env: %v", data.Env.slice)
	logrus.Errorf(" - attributes: %v", data.Env.attributes)
}
