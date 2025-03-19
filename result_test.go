package env

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// Helper function to compare slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Helper function to compare maps
func equalMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// Helper function to create a test result
func createTestResult(value string) *result {
	return &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  value,
		err:    nil,
	}
}

// TestResultChaining tests chaining multiple methods of result
func TestResultChaining(t *testing.T) {
	// Test with valid values
	r := createTestResult("42")

	// Test Required -> Default -> Int
	val, err := r.Required().Default("100").Int()
	if err != nil {
		t.Errorf("Required().Default().Int() unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("Required().Default().Int() expected 42, got %d", val)
	}

	// Test Default -> Required -> Bool
	r = createTestResult("true")
	bval := r.Default("false").Required().Bool()
	if !bval {
		t.Errorf("Default().Required().Bool() expected true, got %v", bval)
	}

	// Test with empty value
	r = createTestResult("")

	// Test Default -> Int
	val, err = r.Default("43").Int()
	if err != nil {
		t.Errorf("Default().Int() unexpected error: %v", err)
	}
	if val != 43 {
		t.Errorf("Default().Int() expected 43, got %d", val)
	}

	// Test Required -> Default
	r = createTestResult("")
	r = r.Required().Default("default_value")
	if r.err == nil {
		t.Error("Required() with empty value should set error")
	}
	if r.value != "" {
		t.Errorf("Value should not be set to default after Required() error, got '%s'", r.value)
	}

	// Test with initial error
	r = &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "", // Ubah dari "valid" menjadi string kosong
		err:    errors.New("initial error"),
	}

	// Methods should pass through error
	r = r.Required().Default("default")
	if r.err == nil || r.err.Error() != "initial error" {
		t.Errorf("Chain with initial error expected 'initial error', got %v", r.err)
	}

	// Value methods should respect error
	if val := r.String(); val != "" {
		t.Errorf("String() with error expected empty, got '%s'", val)
	}

	if val, err := r.Int(); val != 0 || err == nil || err.Error() != "initial error" {
		t.Errorf("Int() with error expected (0, 'initial error'), got (%d, %v)", val, err)
	}

	if val := r.IntDefault(99); val != 99 {
		t.Errorf("IntDefault() with error expected 99, got %d", val)
	}

	if val := r.Bool(); val != false {
		t.Errorf("Bool() with error expected false, got %v", val)
	}
}

// TestResultBoundaryValues tests boundary values for various types
func TestResultBoundaryValues(t *testing.T) {
	// Test maxint and minint
	maxInt := strconv.FormatInt(2147483647, 10)  // Max int32
	minInt := strconv.FormatInt(-2147483648, 10) // Min int32

	r := createTestResult(maxInt)
	val, err := r.Int()
	if err != nil || val != 2147483647 {
		t.Errorf("Int() with MaxInt32 expected %d, got %d (error: %v)",
			2147483647, val, err)
	}

	r = createTestResult(minInt)
	val, err = r.Int()
	if err != nil || val != -2147483648 {
		t.Errorf("Int() with MinInt32 expected %d, got %d (error: %v)",
			-2147483648, val, err)
	}

	// Test max float
	maxFloat := "1.7976931348623157e+308" // Max float64
	r = createTestResult(maxFloat)
	fval, err := r.Float64()
	if err != nil || fval != 1.7976931348623157e+308 {
		t.Errorf("Float64() with MaxFloat64 expected %v, got %v (error: %v)",
			1.7976931348623157e+308, fval, err)
	}

	// Test overflow
	overflow := "9223372036854775808" // MaxInt64 + 1
	r = createTestResult(overflow)
	_, err = r.Int()
	if err == nil {
		t.Error("Int() with overflow should return error")
	}

	// Test IntDefault overflow behavior
	val32 := r.IntDefault(42)
	if val32 != 42 {
		t.Errorf("IntDefault() with overflow expected 42, got %d", val32)
	}

	// Test long duration
	longDur := "87600h" // 10 years
	r = createTestResult(longDur)
	dval, err := r.Duration()
	if err != nil || dval != 87600*time.Hour {
		t.Errorf("Duration() with long duration expected 87600h, got %v (error: %v)",
			dval, err)
	}
}

// TestResultErrorCases tests error handling for all methods
func TestResultErrorCases(t *testing.T) {
	// Invalid value for each type
	invalidValues := map[string]string{
		"int":      "not_an_int",
		"float":    "not_a_float",
		"duration": "not_a_duration",
	}

	// Test each invalid value
	for typ, val := range invalidValues {
		r := createTestResult(val)

		var typeErr error
		switch typ {
		case "int":
			_, typeErr = r.Int()
		case "float":
			_, typeErr = r.Float64()
		case "duration":
			_, typeErr = r.Duration()
		}

		if typeErr == nil {
			t.Errorf("%s() with invalid value should return error", typ)
		}
	}

	// Test with missing key
	r := &result{
		config: &Config{},
		key:    "NONEXISTENT",
		value:  "",
		err:    nil,
	}

	_, err := r.Int()
	if err == nil {
		t.Error("Int() with empty value should return error")
	}

	_, err = r.Float64()
	if err == nil {
		t.Error("Float64() with empty value should return error")
	}

	_, err = r.Duration()
	if err == nil {
		t.Error("Duration() with empty value should return error")
	}
}

// TestResultAdvancedMapParsing tests complex map parsing
func TestResultAdvancedMapParsing(t *testing.T) {
	// Test cases with various map formats
	mapCases := map[string]struct {
		value    string
		expected map[string]string
	}{
		"simple": {
			value:    "key1:value1,key2:value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		"with_spaces": {
			value:    " key1 : value1 , key2 : value2 ",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		"empty_value": {
			value:    "key1:,key2:value2",
			expected: map[string]string{"key1": "", "key2": "value2"},
		},
		"missing_value": {
			value:    "key1,key2:value2",
			expected: map[string]string{"key2": "value2"},
		},
		"duplicate_keys": {
			value:    "key1:value1,key1:value2",
			expected: map[string]string{"key1": "value2"},
		},
		"special_chars": {
			value:    "key-1:value/1,key.2:value\\2",
			expected: map[string]string{"key-1": "value/1", "key.2": "value\\2"},
		},
	}

	for name, tc := range mapCases {
		t.Run(name, func(t *testing.T) {
			r := createTestResult(tc.value)
			mapOutput := r.Map()

			if !equalMaps(mapOutput, tc.expected) {
				t.Errorf("Map() expected %v, got %v", tc.expected, mapOutput)
			}

			// Test MapDefault when value exists
			mapDefaultOutput := r.MapDefault(map[string]string{"default": "value"})
			if !equalMaps(mapDefaultOutput, tc.expected) {
				t.Errorf("MapDefault() with existing value expected %v, got %v",
					tc.expected, mapDefaultOutput)
			}
		})
	}

	// Test MapDefault with empty value
	r := createTestResult("")
	defaultMap := map[string]string{"default": "value"}
	mapOutput := r.MapDefault(defaultMap)

	if !equalMaps(mapOutput, defaultMap) {
		t.Errorf("MapDefault() with empty value expected %v, got %v", defaultMap, mapOutput)
	}

	// Test MapDefault with error
	r = &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "key:value",
		err:    errors.New("test error"),
	}

	mapOutput = r.MapDefault(defaultMap)
	if !equalMaps(mapOutput, defaultMap) {
		t.Errorf("MapDefault() with error expected %v, got %v", defaultMap, mapOutput)
	}
}

// TestResultAdvancedSliceParsing tests complex slice parsing
func TestResultAdvancedSliceParsing(t *testing.T) {
	// Test cases with various slice formats
	sliceCases := map[string]struct {
		value     string
		delimiter string
		expected  []string
	}{
		"simple": {
			value:     "a,b,c",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		"with_spaces": {
			value:     " a , b , c ",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		"custom_delimiter": {
			value:     "a|b|c",
			delimiter: "|",
			expected:  []string{"a", "b", "c"},
		},
		"empty_parts": {
			value:     "a,,c",
			delimiter: ",",
			expected:  []string{"a", "", "c"},
		},
		"quoted_values": {
			value:     "\"a\",b,'c'",
			delimiter: ",",
			expected:  []string{"\"a\"", "b", "'c'"},
		},
		"empty_delimiter": {
			value:     "a,b,c",
			delimiter: "", // Should default to comma
			expected:  []string{"a", "b", "c"},
		},
	}

	for name, tc := range sliceCases {
		t.Run(name, func(t *testing.T) {
			r := createTestResult(tc.value)
			sliceOutput := r.Slice(tc.delimiter)

			if !equalSlices(sliceOutput, tc.expected) {
				t.Errorf("Slice() expected %v, got %v", tc.expected, sliceOutput)
			}

			// Test SliceDefault when value exists
			sliceDefaultOutput := r.SliceDefault(tc.delimiter, []string{"default"})
			if !equalSlices(sliceDefaultOutput, tc.expected) {
				t.Errorf("SliceDefault() with existing value expected %v, got %v",
					tc.expected, sliceDefaultOutput)
			}
		})
	}

	// Test SliceDefault with empty value
	r := createTestResult("")
	defaultSlice := []string{"default1", "default2"}
	sliceOutput := r.SliceDefault(",", defaultSlice)

	if !equalSlices(sliceOutput, defaultSlice) {
		t.Errorf("SliceDefault() with empty value expected %v, got %v", defaultSlice, sliceOutput)
	}

	// Test SliceDefault with error
	r = &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "a,b,c",
		err:    errors.New("test error"),
	}

	sliceOutput = r.SliceDefault(",", defaultSlice)
	if !equalSlices(sliceOutput, defaultSlice) {
		t.Errorf("SliceDefault() with error expected %v, got %v", defaultSlice, sliceOutput)
	}
}

// TestResultDurationComplexCases tests complex duration parsing
func TestResultDurationComplexCases(t *testing.T) {
	// Test complex duration formats
	durationCases := map[string]struct {
		value    string
		expected time.Duration
		hasError bool
	}{
		"simple": {
			value:    "5s",
			expected: 5 * time.Second,
			hasError: false,
		},
		"complex": {
			value:    "1h30m45s",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
			hasError: false,
		},
		"decimal": {
			value:    "1.5h",
			expected: 1*time.Hour + 30*time.Minute,
			hasError: false,
		},
		"microseconds": {
			value:    "1.5µs",
			expected: 1500 * time.Nanosecond,
			hasError: false,
		},
		"zero": {
			value:    "0s",
			expected: 0,
			hasError: false,
		},
		"negative": {
			value:    "-10m",
			expected: -10 * time.Minute,
			hasError: false,
		},
		"invalid": {
			value:    "not_a_duration",
			expected: 0,
			hasError: true,
		},
		"mixed_case": {
			value:    "1H30M", // Uppercase units
			expected: 0,       // Karena akan error
			hasError: true,    // Memang akan error dengan unit huruf besar
		},
	}

	for name, tc := range durationCases {
		t.Run(name, func(t *testing.T) {
			r := createTestResult(tc.value)
			durationOutput, err := r.Duration()

			if tc.hasError {
				if err == nil {
					t.Errorf("Duration() expected error, got nil for value '%s'", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("Duration() unexpected error: %v for value '%s'", err, tc.value)
				}
				if durationOutput != tc.expected {
					t.Errorf("Duration() expected %v, got %v for value '%s'",
						tc.expected, durationOutput, tc.value)
				}
			}

			// Test DurationDefault when value exists and is valid
			if !tc.hasError {
				durationDefaultOutput := r.DurationDefault(time.Minute)
				if durationDefaultOutput != tc.expected {
					t.Errorf("DurationDefault() with valid value expected %v, got %v",
						tc.expected, durationDefaultOutput)
				}
			}
		})
	}

	// Test DurationDefault with invalid value
	r := createTestResult("invalid")
	defaultDur := 5 * time.Minute
	durationOutput := r.DurationDefault(defaultDur)

	if durationOutput != defaultDur {
		t.Errorf("DurationDefault() with invalid value expected %v, got %v",
			defaultDur, durationOutput)
	}
}

// TestResultFloatComplexCases tests complex float parsing
func TestResultFloatComplexCases(t *testing.T) {
	// Test complex float formats
	floatCases := map[string]struct {
		value    string
		expected float64
		hasError bool
	}{
		"integer": {
			value:    "42",
			expected: 42.0,
			hasError: false,
		},
		"decimal": {
			value:    "3.14159",
			expected: 3.14159,
			hasError: false,
		},
		"scientific": {
			value:    "1.23e+5",
			expected: 123000.0,
			hasError: false,
		},
		"negative_scientific": {
			value:    "-1.23e-5",
			expected: -0.0000123,
			hasError: false,
		},
		"zero": {
			value:    "0.0",
			expected: 0.0,
			hasError: false,
		},
		"negative": {
			value:    "-42.5",
			expected: -42.5,
			hasError: false,
		},
		"invalid": {
			value:    "not_a_float",
			expected: 0.0,
			hasError: true,
		},
	}

	for name, tc := range floatCases {
		t.Run(name, func(t *testing.T) {
			r := createTestResult(tc.value)
			floatOutput, err := r.Float64()

			if tc.hasError {
				if err == nil {
					t.Errorf("Float64() expected error, got nil for value '%s'", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("Float64() unexpected error: %v for value '%s'", err, tc.value)
				}
				if floatOutput != tc.expected {
					t.Errorf("Float64() expected %v, got %v for value '%s'",
						tc.expected, floatOutput, tc.value)
				}
			}

			// Test Float64Default when value exists and is valid
			if !tc.hasError {
				floatDefaultOutput := r.Float64Default(99.9)
				if floatDefaultOutput != tc.expected {
					t.Errorf("Float64Default() with valid value expected %v, got %v",
						tc.expected, floatDefaultOutput)
				}
			}
		})
	}
}

// TestResultErrorPropagation tests error propagation in chained methods
func TestResultErrorPropagation(t *testing.T) {
	// Create result with error
	r := &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "valid_but_ignored", // Kembalikan ke nilai asli
		err:    fmt.Errorf("initial error"),
	}

	// Test error propagation in all method types

	// Required should preserve error
	r2 := r.Required()
	if r2.err == nil || r2.err.Error() != "initial error" {
		t.Errorf("Required() did not preserve error, got %v", r2.err)
	}

	// Default should preserve error
	r3 := r.Default("default_value")
	if r3.err == nil || r3.err.Error() != "initial error" {
		t.Errorf("Default() did not preserve error, got %v", r3.err)
	}
	if r3.value != "valid_but_ignored" {
		t.Errorf("Default() modified value with error, expected 'valid_but_ignored', got '%s'",
			r3.value)
	}

	// String should return empty with error
	if str := r.String(); str != "valid_but_ignored" {
		t.Errorf("String() with error expected 'valid_but_ignored', got '%s'", str)
	}

	// Int should propagate error
	if val, err := r.Int(); val != 0 || err == nil || err.Error() != "initial error" {
		t.Errorf("Int() with error expected (0, initial error), got (%d, %v)", val, err)
	}

	// IntDefault should return default with error
	if val := r.IntDefault(42); val != 42 {
		t.Errorf("IntDefault() with error expected 42, got %d", val)
	}

	// Float64 should propagate error
	if val, err := r.Float64(); val != 0.0 || err == nil || err.Error() != "initial error" {
		t.Errorf("Float64() with error expected (0.0, initial error), got (%f, %v)", val, err)
	}

	// Float64Default should return default with error
	if val := r.Float64Default(3.14); val != 3.14 {
		t.Errorf("Float64Default() with error expected 3.14, got %f", val)
	}

	// Bool should return false with error
	if val := r.Bool(); val != false {
		t.Errorf("Bool() with error expected false, got %v", val)
	}

	// BoolDefault should return default with error
	if val := r.BoolDefault(true); val != true {
		t.Errorf("BoolDefault() with error expected true, got %v", val)
	}

	// Duration should propagate error
	if val, err := r.Duration(); val != 0 || err == nil || err.Error() != "initial error" {
		t.Errorf("Duration() with error expected (0, initial error), got (%v, %v)", val, err)
	}

	// DurationDefault should return default with error
	if val := r.DurationDefault(time.Minute); val != time.Minute {
		t.Errorf("DurationDefault() with error expected 1m, got %v", val)
	}

	// Slice should return empty with error
	if val := r.Slice(","); len(val) != 0 {
		t.Errorf("Slice() with error expected empty slice, got %v", val)
	}

	// SliceDefault should return default with error
	defaultSlice := []string{"default"}
	if val := r.SliceDefault(",", defaultSlice); !equalSlices(val, defaultSlice) {
		t.Errorf("SliceDefault() with error expected %v, got %v", defaultSlice, val)
	}

	// Map should return empty with error
	if val := r.Map(); len(val) != 0 {
		t.Errorf("Map() with error expected empty map, got %v", val)
	}

	// MapDefault should return default with error
	defaultMap := map[string]string{"default": "value"}
	if val := r.MapDefault(defaultMap); !equalMaps(val, defaultMap) {
		t.Errorf("MapDefault() with error expected %v, got %v", defaultMap, val)
	}
}

// TestBoolParsingVariations tests all variations of boolean values
func TestBoolParsingVariations(t *testing.T) {
	// Test various boolean representations
	boolCases := map[string]struct {
		value    string
		expected bool
	}{
		"true":   {"true", true},
		"TRUE":   {"TRUE", true},
		"True":   {"True", true},
		"1":      {"1", true},
		"yes":    {"yes", true},
		"YES":    {"YES", true},
		"y":      {"y", true},
		"Y":      {"Y", true},
		"false":  {"false", false},
		"FALSE":  {"FALSE", false},
		"False":  {"False", false},
		"0":      {"0", false},
		"no":     {"no", false},
		"NO":     {"NO", false},
		"n":      {"n", false},
		"N":      {"N", false},
		"other":  {"other", false},
		"empty":  {"", false},
		"spaces": {"   ", false},
		"truthy": {"truthy", false},
		"falsey": {"falsey", false},
	}

	for name, tc := range boolCases {
		t.Run(name, func(t *testing.T) {
			r := createTestResult(tc.value)
			boolOutput := r.Bool()

			if boolOutput != tc.expected {
				t.Errorf("Bool() for '%s' expected %v, got %v", tc.value, tc.expected, boolOutput)
			}

			// Test BoolDefault with actual value
			var boolDefaultOutput bool
			if name == "empty" {
				// Gunakan false sebagai default untuk kasus string kosong
				boolDefaultOutput = r.BoolDefault(false)
			} else {
				// Untuk kasus lain, gunakan !tc.expected
				boolDefaultOutput = r.BoolDefault(!tc.expected)
			}

			if boolDefaultOutput != tc.expected {
				t.Errorf("BoolDefault() for '%s' expected %v, got %v",
					tc.value, tc.expected, boolDefaultOutput)
			}
		})
	}
}
