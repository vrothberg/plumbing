package main

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Array attributedStringSlice `toml:"array,omitempty"`
}

const (
	config_default      = `array=["1", "2", "3"]`
	config_append_front = `array=[{append=true},"4", "5", "6"]`
	config_append_mid   = `array=["7", {append=true}, "8"]`
	config_append_back  = `array=["9", {append=true}]`
	config_append_false = `array=["10", {append=false}]`
)

func loadConfigs(configs []string) (*testConfig, error) {
	var config testConfig
	for _, c := range configs {
		if _, err := toml.Decode(c, &config); err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func TestLoading(t *testing.T) {
	for _, test := range []struct {
		configs                  []string
		expected_slice           []string
		expected_append          bool
		expected_error_substring string
	}{
		// Load single configs
		{[]string{config_default}, []string{"1", "2", "3"}, false, ""},
		{[]string{config_append_front}, []string{"4", "5", "6"}, true, ""},
		{[]string{config_append_mid}, []string{"7", "8"}, true, ""},
		{[]string{config_append_back}, []string{"9"}, true, ""},
		{[]string{config_append_false}, []string{"10"}, false, ""},
		// Append=true
		{[]string{config_default, config_append_front}, []string{"1", "2", "3", "4", "5", "6"}, true, ""},
		{[]string{config_append_front, config_default}, []string{"4", "5", "6", "1", "2", "3"}, true, ""}, // The attribute is sticky unless explicitly being turned off in a later config
		{[]string{config_append_front, config_default, config_append_back}, []string{"4", "5", "6", "1", "2", "3", "9"}, true, ""},
		// Append=false
		{[]string{config_default, config_append_false}, []string{"10"}, false, ""},
		{[]string{config_default, config_append_mid, config_append_false}, []string{"10"}, false, ""},
		{[]string{config_default, config_append_false, config_append_mid}, []string{"10", "7", "8"}, true, ""}, // Append can be re-enabled by a later config

		// Error checks
		{[]string{`array=["1", false]`}, nil, false, `unsupported item in attributed string slice bool: false`},
		{[]string{`array=["1", 42]`}, nil, false, `unsupported item in attributed string slice int`}, // Stop a `int` such that it passes on 32bit as well
		{[]string{`array=["1", {foo=true}]`}, nil, false, `unsupported key "foo" in map: `},
		{[]string{`array=["1", {append="false"}]`}, nil, false, `unable to cast append to bool: `},

	} {
		result, err := loadConfigs(test.configs)
		if test.expected_error_substring != "" {
			require.Error(t, err, "test is expected to fail: %v", test)
			require.ErrorContains(t, err, test.expected_error_substring, "error does not match: %v", test)
			continue
		}
		require.NoError(t, err, "test is expected to succeeed: %v", test)
		require.NotNil(t, result, "loaded config must not be nil: %v", test)
		require.Equal(t, result.Array.slice, test.expected_slice, "slices do not match: %v", test)
		require.Equal(t, result.Array.attributes.append, test.expected_append, "append field does not match: %v", test)
	}

}
