package env

import (
	"os"
	"testing"
	"time"
)

// ParseExtendedConfig untuk menguji parsing tipe-tipe tambahan
type ParseExtendedConfig struct {
	Uint8Field   uint8   `env:"PARSE_UINT8" default:"255"`
	Uint16Field  uint16  `env:"PARSE_UINT16" default:"65535"`
	Uint32Field  uint32  `env:"PARSE_UINT32" default:"4294967295"`
	Float32Field float32 `env:"PARSE_FLOAT32" default:"3.14"`
	Float64Field float64 `env:"PARSE_FLOAT64" default:"3.14159"`
	Int8Field    int8    `env:"PARSE_INT8" default:"127"`
	Int16Field   int16   `env:"PARSE_INT16" default:"32767"`
	Int32Field   int32   `env:"PARSE_INT32" default:"2147483647"`
	Int64Field   int64   `env:"PARSE_INT64" default:"9223372036854775807"`
	NoEnvTag     string  `default:"default_value"`
	EmptyDefault string  `env:"PARSE_EMPTY_DEFAULT" default:""`
}

func TestParseExtendedTypes(t *testing.T) {
	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig ParseExtendedConfig
		shouldFail     bool
	}{
		{
			name: "Parse Extended Types with Custom Values",
			envVars: map[string]string{
				"PARSE_UINT8":         "128",
				"PARSE_UINT16":        "32768",
				"PARSE_UINT32":        "2147483648",
				"PARSE_FLOAT32":       "2.718",
				"PARSE_FLOAT64":       "1.414",
				"PARSE_INT8":          "-128",
				"PARSE_INT16":         "-32768",
				"PARSE_INT32":         "-2147483648",
				"PARSE_INT64":         "-9223372036854775808",
				"PARSE_EMPTY_DEFAULT": "",
			},
			expectedConfig: ParseExtendedConfig{
				Uint8Field:   128,
				Uint16Field:  32768,
				Uint32Field:  2147483648,
				Float32Field: 2.718,
				Float64Field: 1.414,
				Int8Field:    -128,
				Int16Field:   -32768,
				Int32Field:   -2147483648,
				Int64Field:   -9223372036854775808,
				NoEnvTag:     "default_value",
				EmptyDefault: "",
			},
		},
		{
			name:    "Parse Extended Types with Default Values",
			envVars: map[string]string{},
			expectedConfig: ParseExtendedConfig{
				Uint8Field:   255,
				Uint16Field:  65535,
				Uint32Field:  4294967295,
				Float32Field: 3.14,
				Float64Field: 3.14159,
				Int8Field:    127,
				Int16Field:   32767,
				Int32Field:   2147483647,
				Int64Field:   9223372036854775807,
				NoEnvTag:     "default_value",
				EmptyDefault: "",
			},
		},
		{
			name: "Invalid Uint Parse",
			envVars: map[string]string{
				"PARSE_UINT8": "-1",
			},
			shouldFail: true,
		},
		{
			name: "Invalid Float Parse",
			envVars: map[string]string{
				"PARSE_FLOAT32": "not_a_number",
			},
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tc.envVars {
				os.Setenv(key, val)
			}
			defer func() {
				// Unset all environment variables
				for key := range tc.envVars {
					os.Unsetenv(key)
				}
			}()

			var config ParseExtendedConfig
			err := Parse(&config)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected parsing to fail for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Compare struct values
			if config.Uint8Field != tc.expectedConfig.Uint8Field {
				t.Errorf("Uint8Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Uint8Field, config.Uint8Field)
			}
			if config.Uint16Field != tc.expectedConfig.Uint16Field {
				t.Errorf("Uint16Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Uint16Field, config.Uint16Field)
			}
			if config.Uint32Field != tc.expectedConfig.Uint32Field {
				t.Errorf("Uint32Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Uint32Field, config.Uint32Field)
			}
			if config.Float32Field != tc.expectedConfig.Float32Field {
				t.Errorf("Float32Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Float32Field, config.Float32Field)
			}
			if config.Float64Field != tc.expectedConfig.Float64Field {
				t.Errorf("Float64Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Float64Field, config.Float64Field)
			}
			if config.Int8Field != tc.expectedConfig.Int8Field {
				t.Errorf("Int8Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Int8Field, config.Int8Field)
			}
			if config.Int16Field != tc.expectedConfig.Int16Field {
				t.Errorf("Int16Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Int16Field, config.Int16Field)
			}
			if config.Int32Field != tc.expectedConfig.Int32Field {
				t.Errorf("Int32Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Int32Field, config.Int32Field)
			}
			if config.Int64Field != tc.expectedConfig.Int64Field {
				t.Errorf("Int64Field mismatch. Expected %v, got %v",
					tc.expectedConfig.Int64Field, config.Int64Field)
			}
			if config.NoEnvTag != tc.expectedConfig.NoEnvTag {
				t.Errorf("NoEnvTag mismatch. Expected %v, got %v",
					tc.expectedConfig.NoEnvTag, config.NoEnvTag)
			}
			if config.EmptyDefault != tc.expectedConfig.EmptyDefault {
				t.Errorf("EmptyDefault mismatch. Expected %v, got %v",
					tc.expectedConfig.EmptyDefault, config.EmptyDefault)
			}
		})
	}
}

// ParseBoolConfig untuk menguji berbagai variasi parsing boolean
type ParseBoolConfig struct {
	TrueValues  bool `env:"PARSE_TRUE"`
	FalseValues bool `env:"PARSE_FALSE"`
	InvalidBool bool `env:"PARSE_INVALID_BOOL"`
}

func TestParseBoolValues(t *testing.T) {
	testCases := []struct {
		name           string
		envValue       string
		expectedResult bool
	}{
		{"True Literal", "true", true},
		{"True Number", "1", true},
		{"True Yes", "yes", true},
		{"True Y", "y", true},
		{"False Literal", "false", false},
		{"False Number", "0", false},
		{"False No", "no", false},
		{"False N", "n", false},
		{"Mixed Case", "True", true},
		{"Mixed Case False", "False", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("PARSE_TRUE", tc.envValue)
			defer os.Unsetenv("PARSE_TRUE")

			var config ParseBoolConfig
			err := Parse(&config)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config.TrueValues != tc.expectedResult {
				t.Errorf("Expected %v, got %v for input %s",
					tc.expectedResult, config.TrueValues, tc.envValue)
			}
		})
	}
}

// ParseComplexConfig untuk menguji kasus edge case
type ParseComplexConfig struct {
	SpacedSlice []string          `env:"PARSE_SPACED_SLICE"`
	SpacedMap   map[string]string `env:"PARSE_SPACED_MAP"`
}

func TestParseComplexTypes(t *testing.T) {
	t.Run("Parse Slice with Spaces", func(t *testing.T) {
		os.Setenv("PARSE_SPACED_SLICE", " item1 , item2 , item3 ")
		defer os.Unsetenv("PARSE_SPACED_SLICE")

		var config ParseComplexConfig
		err := Parse(&config)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		expectedSlice := []string{"item1", "item2", "item3"}
		if len(config.SpacedSlice) != len(expectedSlice) {
			t.Errorf("Expected %d items, got %d", len(expectedSlice), len(config.SpacedSlice))
			return
		}

		for i, item := range expectedSlice {
			if config.SpacedSlice[i] != item {
				t.Errorf("Mismatch at index %d. Expected %s, got %s", i, item, config.SpacedSlice[i])
			}
		}
	})

	t.Run("Parse Map with Spaces", func(t *testing.T) {
		os.Setenv("PARSE_SPACED_MAP", " key1 : value1 , key2 : value2 ")
		defer os.Unsetenv("PARSE_SPACED_MAP")

		var config ParseComplexConfig
		err := Parse(&config)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		expectedMap := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		if len(config.SpacedMap) != len(expectedMap) {
			t.Errorf("Expected %d items, got %d", len(expectedMap), len(config.SpacedMap))
			return
		}

		for k, v := range expectedMap {
			if config.SpacedMap[k] != v {
				t.Errorf("Mismatch for key %s. Expected %s, got %s", k, v, config.SpacedMap[k])
			}
		}
	})
}

// ParseUnsupportedConfig untuk menguji tipe yang tidak didukung
type ParseUnsupportedConfig struct {
	UnsupportedSlice []int          `env:"PARSE_UNSUPPORTED_SLICE"`
	UnsupportedMap   map[int]string `env:"PARSE_UNSUPPORTED_MAP"`
}

func TestParseUnsupportedTypes(t *testing.T) {
	t.Run("Unsupported Slice Type", func(t *testing.T) {
		os.Setenv("PARSE_UNSUPPORTED_SLICE", "1,2,3")
		defer os.Unsetenv("PARSE_UNSUPPORTED_SLICE")

		var config ParseUnsupportedConfig
		err := Parse(&config)

		if err == nil {
			t.Error("Expected error for unsupported slice type, got nil")
		}
	})

	t.Run("Unsupported Map Type", func(t *testing.T) {
		os.Setenv("PARSE_UNSUPPORTED_MAP", "1:value1,2:value2")
		defer os.Unsetenv("PARSE_UNSUPPORTED_MAP")

		var config ParseUnsupportedConfig
		err := Parse(&config)

		if err == nil {
			t.Error("Expected error for unsupported map type, got nil")
		}
	})
}

// ParseTimeoutConfig untuk menguji parsing durasi
type ParseTimeoutConfig struct {
	ShortTimeout   time.Duration `env:"PARSE_SHORT_TIMEOUT"`
	LongTimeout    time.Duration `env:"PARSE_LONG_TIMEOUT"`
	InvalidTimeout time.Duration `env:"PARSE_INVALID_TIMEOUT"`
}

func TestParseDurationValues(t *testing.T) {
	testCases := []struct {
		name           string
		envValue       string
		expectedResult time.Duration
		shouldFail     bool
	}{
		{"Short Duration", "5s", 5 * time.Second, false},
		{"Long Duration", "1h30m", 1*time.Hour + 30*time.Minute, false},
		{"Milliseconds", "500ms", 500 * time.Millisecond, false},
		{"Invalid Duration", "invalid", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("PARSE_SHORT_TIMEOUT", tc.envValue)
			defer os.Unsetenv("PARSE_SHORT_TIMEOUT")

			var config ParseTimeoutConfig
			err := Parse(&config)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected parsing to fail for %s", tc.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config.ShortTimeout != tc.expectedResult {
				t.Errorf("Expected %v, got %v for input %s",
					tc.expectedResult, config.ShortTimeout, tc.envValue)
			}
		})
	}
}
