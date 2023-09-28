package main

import (
	"bytes"
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

var (
	bTrue  = true
	bFalse = false
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

func TestAttributedStringSliceLoading(t *testing.T) {
	for _, test := range []struct {
		configs                  []string
		expected_slice           []string
		expected_append          *bool
		expected_error_substring string
	}{
		// Load single configs
		{[]string{config_default}, []string{"1", "2", "3"}, nil, ""},
		{[]string{config_append_front}, []string{"4", "5", "6"}, &bTrue, ""},
		{[]string{config_append_mid}, []string{"7", "8"}, &bTrue, ""},
		{[]string{config_append_back}, []string{"9"}, &bTrue, ""},
		{[]string{config_append_false}, []string{"10"}, &bFalse, ""},
		// Append=true
		{[]string{config_default, config_append_front}, []string{"1", "2", "3", "4", "5", "6"}, &bTrue, ""},
		{[]string{config_append_front, config_default}, []string{"4", "5", "6", "1", "2", "3"}, &bTrue, ""}, // The attribute is sticky unless explicitly being turned off in a later config
		{[]string{config_append_front, config_default, config_append_back}, []string{"4", "5", "6", "1", "2", "3", "9"}, &bTrue, ""},
		// Append=false
		{[]string{config_default, config_append_false}, []string{"10"}, &bFalse, ""},
		{[]string{config_default, config_append_mid, config_append_false}, []string{"10"}, &bFalse, ""},
		{[]string{config_default, config_append_false, config_append_mid}, []string{"10", "7", "8"}, &bTrue, ""}, // Append can be re-enabled by a later config

		// Error checks
		{[]string{`array=["1", false]`}, nil, nil, `unsupported item in attributed string slice bool: false`},
		{[]string{`array=["1", 42]`}, nil, nil, `unsupported item in attributed string slice int`}, // Stop a `int` such that it passes on 32bit as well
		{[]string{`array=["1", {foo=true}]`}, nil, nil, `unsupported key "foo" in map: `},
		{[]string{`array=["1", {append="false"}]`}, nil, nil, `unable to cast append to bool: `},
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

func TestAttributedStringSliceEncoding(t *testing.T) {
	for _, test := range []struct {
		configs         []string
		marshalled_data string
		expected_slice  []string
		expected_append *bool
	}{
		{
			[]string{config_default},
			"array = [\"1\", \"2\", \"3\"]\n",
			[]string{"1", "2", "3"},
			nil,
		},
		{
			[]string{config_append_front},
			"array = [\"4\", \"5\", \"6\", {append = true}]\n",
			[]string{"4", "5", "6"},
			&bTrue,
		},
		{
			[]string{config_append_front,config_append_false},
			"array = [\"10\", {append = false}]\n",
			[]string{"10"},
			&bFalse,
		},
	} {
		// 1) Load the configs
		result, err := loadConfigs(test.configs)
		require.NoError(t, err, "loading config must succeed")
		require.NotNil(t, result, "loaded config must not be nil")
		require.Equal(t, result.Array.slice, test.expected_slice, "slices do not match: %v", test)
		require.Equal(t, result.Array.attributes.append, test.expected_append, "append field does not match: %v", test)

		// 2) Marshal the config to emulate writing it to disk
		buf := new(bytes.Buffer)
		enc := toml.NewEncoder(buf)
		encErr := enc.Encode(result)
		require.NoError(t, encErr, "encoding config must work")
		require.Equal(t, buf.String(), test.marshalled_data)

		// 3) Reload the marshaled config to make sure that data is preserved
		var reloadedConfig testConfig
		_, decErr := toml.Decode(buf.String(), &reloadedConfig)
		require.NoError(t, decErr, "loading config must succeed")
		require.NotNil(t, reloadedConfig, "re-loaded config must not be nil")
		require.Equal(t, reloadedConfig.Array.slice, test.expected_slice, "slices do not match: %v", test)
		require.Equal(t, reloadedConfig.Array.attributes.append, test.expected_append, "append field does not match: %v", test)
	}
}
