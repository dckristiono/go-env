package env

import (
	"os"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	// Set up
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	// Test
	value := Get("TEST_VAR")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test with default
	value = Get("NON_EXISTENT_VAR", "default_value")
	if value != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value)
	}
}

func TestGetInt(t *testing.T) {
	// Set up
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	// Test
	value, err := GetInt("TEST_INT")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test with default
	value, err = GetInt("NON_EXISTENT_INT", 100)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	// Test invalid
	os.Setenv("INVALID_INT", "not_an_int")
	defer os.Unsetenv("INVALID_INT")
	_, err = GetInt("INVALID_INT")
	if err == nil {
		t.Error("Expected error for invalid int, got nil")
	}
}

func TestGetFloat64(t *testing.T) {
	// Set up
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	// Test
	value, err := GetFloat64("TEST_FLOAT")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 3.14 {
		t.Errorf("Expected 3.14, got %f", value)
	}

	// Test with default
	value, err = GetFloat64("NON_EXISTENT_FLOAT", 2.71)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 2.71 {
		t.Errorf("Expected 2.71, got %f", value)
	}

	// Test invalid
	os.Setenv("INVALID_FLOAT", "not_a_float")
	defer os.Unsetenv("INVALID_FLOAT")
	_, err = GetFloat64("INVALID_FLOAT")
	if err == nil {
		t.Error("Expected error for invalid float, got nil")
	}
}

func TestGetBool(t *testing.T) {
	// Test various boolean representations
	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"y", true},
		{"false", false},
		{"FALSE", false},
		{"False", false},
		{"0", false},
		{"no", false},
		{"n", false},
		{"anything_else", false},
	}

	for _, test := range tests {
		os.Setenv("TEST_BOOL", test.envValue)
		value := GetBool("TEST_BOOL")
		if value != test.expected {
			t.Errorf("For '%s', expected %v, got %v", test.envValue, test.expected, value)
		}
	}
	os.Unsetenv("TEST_BOOL")

	// Test with default
	value := GetBool("NON_EXISTENT_BOOL", true)
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}
}

func TestGetDuration(t *testing.T) {
	// Set up
	os.Setenv("TEST_DURATION", "5s")
	defer os.Unsetenv("TEST_DURATION")

	// Test
	value, err := GetDuration("TEST_DURATION")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 5*time.Second {
		t.Errorf("Expected 5s, got %v", value)
	}

	// Test with default
	value, err = GetDuration("NON_EXISTENT_DURATION", 10*time.Second)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 10*time.Second {
		t.Errorf("Expected 10s, got %v", value)
	}

	// Test invalid
	os.Setenv("INVALID_DURATION", "not_a_duration")
	defer os.Unsetenv("INVALID_DURATION")
	_, err = GetDuration("INVALID_DURATION")
	if err == nil {
		t.Error("Expected error for invalid duration, got nil")
	}
}

func TestGetSlice(t *testing.T) {
	// Set up
	os.Setenv("TEST_SLICE", "item1,item2,item3")
	defer os.Unsetenv("TEST_SLICE")

	// Test
	value := GetSlice("TEST_SLICE", ",")
	if len(value) != 3 {
		t.Errorf("Expected 3 items, got %d", len(value))
	}
	if value[0] != "item1" || value[1] != "item2" || value[2] != "item3" {
		t.Errorf("Unexpected values: %v", value)
	}

	// Test with default
	defaultSlice := []string{"default1", "default2"}
	value = GetSlice("NON_EXISTENT_SLICE", ",", defaultSlice)
	if len(value) != 2 {
		t.Errorf("Expected 2 items, got %d", len(value))
	}
	if value[0] != "default1" || value[1] != "default2" {
		t.Errorf("Unexpected values: %v", value)
	}

	// Test with spaces
	os.Setenv("TEST_SLICE_SPACES", "item1, item2, item3")
	defer os.Unsetenv("TEST_SLICE_SPACES")
	value = GetSlice("TEST_SLICE_SPACES", ",")
	if value[0] != "item1" || value[1] != "item2" || value[2] != "item3" {
		t.Errorf("Unexpected values: %v", value)
	}
}

func TestGetMap(t *testing.T) {
	// Set up
	os.Setenv("TEST_MAP", "key1:value1,key2:value2,key3:value3")
	defer os.Unsetenv("TEST_MAP")

	// Test
	value := GetMap("TEST_MAP")
	if len(value) != 3 {
		t.Errorf("Expected 3 items, got %d", len(value))
	}
	if value["key1"] != "value1" || value["key2"] != "value2" || value["key3"] != "value3" {
		t.Errorf("Unexpected values: %v", value)
	}

	// Test with default
	defaultMap := map[string]string{"default1": "val1", "default2": "val2"}
	value = GetMap("NON_EXISTENT_MAP", defaultMap)
	if len(value) != 2 {
		t.Errorf("Expected 2 items, got %d", len(value))
	}
	if value["default1"] != "val1" || value["default2"] != "val2" {
		t.Errorf("Unexpected values: %v", value)
	}

	// Test with spaces
	os.Setenv("TEST_MAP_SPACES", "key1: value1, key2: value2")
	defer os.Unsetenv("TEST_MAP_SPACES")
	value = GetMap("TEST_MAP_SPACES")
	if value["key1"] != "value1" || value["key2"] != "value2" {
		t.Errorf("Unexpected values: %v", value)
	}
}

func TestFluentAPI(t *testing.T) {
	// Set up
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	// Test string
	value := Key("TEST_KEY").String()
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test with default
	value = Key("NON_EXISTENT_KEY").Default("default_value").String()
	if value != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value)
	}

	// Test int
	os.Setenv("TEST_INT_KEY", "42")
	defer os.Unsetenv("TEST_INT_KEY")
	intValue := Key("TEST_INT_KEY").IntDefault(0)
	if intValue != 42 {
		t.Errorf("Expected 42, got %d", intValue)
	}

	// Test required
	err := Key("NON_EXISTENT_REQUIRED").Required().err
	if err == nil {
		t.Error("Expected error for required key, got nil")
	}
}

func TestWithPrefix(t *testing.T) {
	// Set up
	os.Setenv("PREFIX_TEST", "prefixed_value")
	defer os.Unsetenv("PREFIX_TEST")

	// Test with prefix
	cfg := With(WithPrefix("PREFIX_"))
	value := cfg.Get("TEST")
	if value != "prefixed_value" {
		t.Errorf("Expected 'prefixed_value', got '%s'", value)
	}

	// Test fluent with prefix
	value = With(WithPrefix("PREFIX_")).Key("TEST").String()
	if value != "prefixed_value" {
		t.Errorf("Expected 'prefixed_value', got '%s'", value)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Set up
	os.Setenv("TEST_HELPER_INT", "42")
	os.Setenv("TEST_HELPER_FLOAT", "3.14")
	os.Setenv("TEST_HELPER_BOOL", "true")
	os.Setenv("TEST_HELPER_DURATION", "5s")
	defer func() {
		os.Unsetenv("TEST_HELPER_INT")
		os.Unsetenv("TEST_HELPER_FLOAT")
		os.Unsetenv("TEST_HELPER_BOOL")
		os.Unsetenv("TEST_HELPER_DURATION")
	}()

	// Test helper functions
	intValue := Int("TEST_HELPER_INT", 0)
	if intValue != 42 {
		t.Errorf("Expected 42, got %d", intValue)
	}

	floatValue := Float64("TEST_HELPER_FLOAT", 0.0)
	if floatValue != 3.14 {
		t.Errorf("Expected 3.14, got %f", floatValue)
	}

	boolValue := Bool("TEST_HELPER_BOOL", false)
	if !boolValue {
		t.Errorf("Expected true, got %v", boolValue)
	}

	durationValue := Duration("TEST_HELPER_DURATION", 0)
	if durationValue != 5*time.Second {
		t.Errorf("Expected 5s, got %v", durationValue)
	}

	// Test with defaults
	intValue = Int("NON_EXISTENT_INT", 100)
	if intValue != 100 {
		t.Errorf("Expected 100, got %d", intValue)
	}
}
