package env

import (
	"errors"
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

func TestResultRequired(t *testing.T) {
	testCases := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"Value Present", "test_value", false},
		{"Empty Value", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)
			result := r.Required()

			if tc.expectErr && result.err == nil {
				t.Error("Expected error for required field, got nil")
			}
			if !tc.expectErr && result.err != nil {
				t.Errorf("Unexpected error: %v", result.err)
			}
		})
	}
}

func TestResultDefault(t *testing.T) {
	testCases := []struct {
		name           string
		value          string
		defaultValue   string
		expectedResult string
	}{
		{"Value Present", "original_value", "default_value", "original_value"},
		{"Empty Value", "", "default_value", "default_value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)
			result := r.Default(tc.defaultValue)

			if result.value != tc.expectedResult {
				t.Errorf("Expected %s, got %s", tc.expectedResult, result.value)
			}
		})
	}
}

func TestResultBoolMethods(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		defaultValue  bool
		expectBool    bool
		expectDefault bool
	}{
		{"True Values", "true", false, true, true},
		{"True Values (1)", "1", false, true, true},
		{"True Values (yes)", "yes", false, true, true},
		{"True Values (y)", "y", false, true, true},
		{"False Values", "false", true, false, false},
		{"False Values (0)", "0", true, false, false},
		{"False Values (no)", "no", true, false, false},
		{"False Values (n)", "n", true, false, false},
		{"Invalid Values", "invalid", true, false, false},
		{"Empty Values", "", true, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Bool()
			boolVal := r.Bool()
			if boolVal != tc.expectBool {
				t.Errorf("Bool() - Expected %v, got %v (input: %s)",
					tc.expectBool, boolVal, tc.value)
			}

			// Test BoolDefault()
			boolDefaultVal := r.BoolDefault(tc.defaultValue)
			if boolDefaultVal != tc.expectDefault {
				t.Errorf("BoolDefault() - Expected %v, got %v (input: %s, default: %v)",
					tc.expectDefault, boolDefaultVal, tc.value, tc.defaultValue)
			}
		})
	}

	// Test with existing error
	r := &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "true",
		err:    errors.New("test error"),
	}

	if r.Bool() != false {
		t.Error("Bool() should return false when error exists")
	}

	if r.BoolDefault(true) != true {
		t.Error("BoolDefault() should return default value when error exists")
	}
}

func TestResultIntMethods(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		defaultValue  int
		expectInt     int
		expectDefault int
		expectErr     bool
	}{
		{"Valid Integer", "42", 0, 42, 42, false},
		{"Negative Integer", "-100", 0, -100, -100, false},
		{"Invalid Integer", "not_an_int", 100, 0, 100, true},
		{"Empty Value", "", 100, 0, 100, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Int()
			intVal, err := r.Int()
			if tc.expectErr {
				if err == nil {
					t.Errorf("Int() - Expected error, got nil (input: %s)", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("Int() - Unexpected error: %v", err)
				}
				if intVal != tc.expectInt {
					t.Errorf("Int() - Expected %d, got %d (input: %s)",
						tc.expectInt, intVal, tc.value)
				}
			}

			// Test IntDefault()
			intDefaultVal := r.IntDefault(tc.defaultValue)
			if intDefaultVal != tc.expectDefault {
				t.Errorf("IntDefault() - Expected %d, got %d (input: %s, default: %d)",
					tc.expectDefault, intDefaultVal, tc.value, tc.defaultValue)
			}
		})
	}
}

func TestResultFloat64Methods(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		defaultValue  float64
		expectFloat   float64
		expectDefault float64
		expectErr     bool
	}{
		{"Valid Float", "3.14", 0, 3.14, 3.14, false},
		{"Negative Float", "-2.5", 0, -2.5, -2.5, false},
		{"Invalid Float", "not_a_float", 100.5, 0, 100.5, true},
		{"Empty Value", "", 100.5, 0, 100.5, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Float64()
			floatVal, err := r.Float64()
			if tc.expectErr {
				if err == nil {
					t.Errorf("Float64() - Expected error, got nil (input: %s)", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("Float64() - Unexpected error: %v", err)
				}
				if floatVal != tc.expectFloat {
					t.Errorf("Float64() - Expected %f, got %f (input: %s)",
						tc.expectFloat, floatVal, tc.value)
				}
			}

			// Test Float64Default()
			floatDefaultVal := r.Float64Default(tc.defaultValue)
			if floatDefaultVal != tc.expectDefault {
				t.Errorf("Float64Default() - Expected %f, got %f (input: %s, default: %f)",
					tc.expectDefault, floatDefaultVal, tc.value, tc.defaultValue)
			}
		})
	}
}

func TestResultDurationMethods(t *testing.T) {
	testCases := []struct {
		name           string
		value          string
		defaultValue   time.Duration
		expectDuration time.Duration
		expectDefault  time.Duration
		expectErr      bool
	}{
		{"Valid Duration", "5s", 0, 5 * time.Second, 5 * time.Second, false},
		{"Long Duration", "1h30m", 0, 1*time.Hour + 30*time.Minute, 1*time.Hour + 30*time.Minute, false},
		{"Invalid Duration", "not_a_duration", 10 * time.Second, 0, 10 * time.Second, true},
		{"Empty Value", "", 10 * time.Second, 0, 10 * time.Second, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Duration()
			durationVal, err := r.Duration()
			if tc.expectErr {
				if err == nil {
					t.Errorf("Duration() - Expected error, got nil (input: %s)", tc.value)
				}
			} else {
				if err != nil {
					t.Errorf("Duration() - Unexpected error: %v", err)
				}
				if durationVal != tc.expectDuration {
					t.Errorf("Duration() - Expected %v, got %v (input: %s)",
						tc.expectDuration, durationVal, tc.value)
				}
			}

			// Test DurationDefault()
			durationDefaultVal := r.DurationDefault(tc.defaultValue)
			if durationDefaultVal != tc.expectDefault {
				t.Errorf("DurationDefault() - Expected %v, got %v (input: %s, default: %v)",
					tc.expectDefault, durationDefaultVal, tc.value, tc.defaultValue)
			}
		})
	}
}

func TestResultSliceMethods(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		delimiter     string
		defaultValue  []string
		expectSlice   []string
		expectDefault []string
	}{
		{"Basic Slice", "item1,item2,item3", ",", nil, []string{"item1", "item2", "item3"}, nil},
		{"Slice with Spaces", " item1 , item2 , item3 ", ",", nil, []string{"item1", "item2", "item3"}, nil},
		{"Custom Delimiter", "item1;item2;item3", ";", nil, []string{"item1", "item2", "item3"}, nil},
		{"Empty Slice", "", ",", []string{"default1", "default2"}, []string{}, []string{"default1", "default2"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Slice()
			if tc.defaultValue == nil {
				sliceVal := r.Slice(tc.delimiter)
				if !equalSlices(sliceVal, tc.expectSlice) {
					t.Errorf("Slice() - Expected %v, got %v (input: %s)",
						tc.expectSlice, sliceVal, tc.value)
				}
			} else {
				// Test SliceDefault()
				sliceDefaultVal := r.SliceDefault(tc.delimiter, tc.defaultValue)
				if !equalSlices(sliceDefaultVal, tc.expectDefault) {
					t.Errorf("SliceDefault() - Expected %v, got %v (input: %s)",
						tc.expectDefault, sliceDefaultVal, tc.value)
				}
			}
		})
	}
}

func TestResultMapMethods(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		defaultValue  map[string]string
		expectMap     map[string]string
		expectDefault map[string]string
	}{
		{"Basic Map", "key1:value1,key2:value2", nil, map[string]string{"key1": "value1", "key2": "value2"}, nil},
		{"Map with Spaces", " key1 : value1 , key2 : value2 ", nil, map[string]string{"key1": "value1", "key2": "value2"}, nil},
		{"Partial Map", "key1:value1,invalid", nil, map[string]string{"key1": "value1"}, nil},
		{"Empty Map", "", map[string]string{"default1": "value1"}, map[string]string{}, map[string]string{"default1": "value1"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := createTestResult(tc.value)

			// Test Map()
			if tc.defaultValue == nil {
				mapVal := r.Map()
				if !equalMaps(mapVal, tc.expectMap) {
					t.Errorf("Map() - Expected %v, got %v (input: %s)",
						tc.expectMap, mapVal, tc.value)
				}
			} else {
				// Test MapDefault()
				mapDefaultVal := r.MapDefault(tc.defaultValue)
				if !equalMaps(mapDefaultVal, tc.expectDefault) {
					t.Errorf("MapDefault() - Expected %v, got %v (input: %s)",
						tc.expectDefault, mapDefaultVal, tc.value)
				}
			}
		})
	}
}
