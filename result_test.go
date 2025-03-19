package env

import (
	"errors"
	"testing"
	"time"
)

func TestResultRequired(t *testing.T) {
	// Set up
	config := &Config{}

	// Test with value present
	r := &result{
		config: config,
		key:    "TEST_KEY",
		value:  "test_value",
		err:    nil,
	}

	r = r.Required()
	if r.err != nil {
		t.Errorf("Expected no error for required field with value, got: %v", r.err)
	}

	// Test with value missing
	r = &result{
		config: config,
		key:    "TEST_KEY",
		value:  "",
		err:    nil,
	}

	r = r.Required()
	if r.err == nil {
		t.Error("Expected error for required field without value, got nil")
	}

	// Test with existing error
	existingErr := errors.New("existing error")
	r = &result{
		config: config,
		key:    "TEST_KEY",
		value:  "test_value",
		err:    existingErr,
	}

	r = r.Required()
	if r.err != existingErr {
		t.Errorf("Expected existing error to be preserved, got: %v", r.err)
	}
}

func TestResultDefault(t *testing.T) {
	// Set up
	config := &Config{}

	// Test with value present
	r := &result{
		config: config,
		key:    "TEST_KEY",
		value:  "original_value",
		err:    nil,
	}

	r = r.Default("default_value")
	if r.value != "original_value" {
		t.Errorf("Expected original value to be preserved, got: %s", r.value)
	}

	// Test with value missing
	r = &result{
		config: config,
		key:    "TEST_KEY",
		value:  "",
		err:    nil,
	}

	r = r.Default("default_value")
	if r.value != "default_value" {
		t.Errorf("Expected default value to be set, got: %s", r.value)
	}

	// Test with existing error
	existingErr := errors.New("existing error")
	r = &result{
		config: config,
		key:    "TEST_KEY",
		value:  "",
		err:    existingErr,
	}

	r = r.Default("default_value")
	if r.err != existingErr {
		t.Errorf("Expected existing error to be preserved, got: %v", r.err)
	}
}

func TestResultTypeConversions(t *testing.T) {
	// Set up
	config := &Config{}

	// Test String()
	r := &result{
		config: config,
		key:    "TEST_KEY",
		value:  "test_value",
		err:    nil,
	}

	if r.String() != "test_value" {
		t.Errorf("Expected String() to return 'test_value', got: %s", r.String())
	}

	// Test Int()
	r = &result{
		config: config,
		key:    "TEST_INT",
		value:  "42",
		err:    nil,
	}

	intVal, err := r.Int()
	if err != nil {
		t.Errorf("Unexpected error from Int(): %v", err)
	}
	if intVal != 42 {
		t.Errorf("Expected Int() to return 42, got: %d", intVal)
	}

	// Test Int() with invalid value
	r = &result{
		config: config,
		key:    "TEST_INT",
		value:  "not_an_int",
		err:    nil,
	}

	_, err = r.Int()
	if err == nil {
		t.Error("Expected error for invalid int, got nil")
	}

	// Test IntDefault()
	r = &result{
		config: config,
		key:    "TEST_INT",
		value:  "42",
		err:    nil,
	}

	intVal = r.IntDefault(100)
	if intVal != 42 {
		t.Errorf("Expected IntDefault() to return 42, got: %d", intVal)
	}

	// Test IntDefault() with invalid value
	r = &result{
		config: config,
		key:    "TEST_INT",
		value:  "not_an_int",
		err:    nil,
	}

	intVal = r.IntDefault(100)
	if intVal != 100 {
		t.Errorf("Expected IntDefault() to return default 100, got: %d", intVal)
	}

	// Test Float64()
	r = &result{
		config: config,
		key:    "TEST_FLOAT",
		value:  "3.14",
		err:    nil,
	}

	floatVal, err := r.Float64()
	if err != nil {
		t.Errorf("Unexpected error from Float64(): %v", err)
	}
	if floatVal != 3.14 {
		t.Errorf("Expected Float64() to return 3.14, got: %f", floatVal)
	}

	// Test Float64Default()
	r = &result{
		config: config,
		key:    "TEST_FLOAT",
		value:  "3.14",
		err:    nil,
	}

	floatVal = r.Float64Default(2.71)
	if floatVal != 3.14 {
		t.Errorf("Expected Float64Default() to return 3.14, got: %f", floatVal)
	}

	// Test Bool()
	r = &result{
		config: config,
		key:    "TEST_BOOL",
		value:  "true",
		err:    nil,
	}

	boolVal := r.Bool()
	if !boolVal {
		t.Error("Expected Bool() to return true, got false")
	}

	// Test BoolDefault()
	r = &result{
		config: config,
		key:    "TEST_BOOL",
		value:  "false",
		err:    nil,
	}

	boolVal = r.BoolDefault(true)
	if boolVal {
		t.Error("Expected BoolDefault() to return false, got true")
	}

	// Test Duration()
	r = &result{
		config: config,
		key:    "TEST_DURATION",
		value:  "5s",
		err:    nil,
	}

	durationVal, err := r.Duration()
	if err != nil {
		t.Errorf("Unexpected error from Duration(): %v", err)
	}
	if durationVal != 5*time.Second {
		t.Errorf("Expected Duration() to return 5s, got: %v", durationVal)
	}

	// Test DurationDefault()
	r = &result{
		config: config,
		key:    "TEST_DURATION",
		value:  "5s",
		err:    nil,
	}

	durationVal = r.DurationDefault(10 * time.Second)
	if durationVal != 5*time.Second {
		t.Errorf("Expected DurationDefault() to return 5s, got: %v", durationVal)
	}

	// Test Slice()
	r = &result{
		config: config,
		key:    "TEST_SLICE",
		value:  "item1,item2,item3",
		err:    nil,
	}

	sliceVal := r.Slice(",")
	if len(sliceVal) != 3 {
		t.Errorf("Expected Slice() to return 3 items, got: %d", len(sliceVal))
	}
	if sliceVal[0] != "item1" || sliceVal[1] != "item2" || sliceVal[2] != "item3" {
		t.Errorf("Unexpected slice values: %v", sliceVal)
	}

	// Test SliceDefault()
	r = &result{
		config: config,
		key:    "TEST_SLICE",
		value:  "",
		err:    nil,
	}

	defaultSlice := []string{"default1", "default2"}
	sliceVal = r.SliceDefault(",", defaultSlice)
	if !equalSlices(sliceVal, defaultSlice) {
		t.Errorf("Expected SliceDefault() to return default slice, got: %v", sliceVal)
	}

	// Test Map()
	r = &result{
		config: config,
		key:    "TEST_MAP",
		value:  "key1:value1,key2:value2",
		err:    nil,
	}

	mapVal := r.Map()
	if len(mapVal) != 2 {
		t.Errorf("Expected Map() to return 2 items, got: %d", len(mapVal))
	}
	if mapVal["key1"] != "value1" || mapVal["key2"] != "value2" {
		t.Errorf("Unexpected map values: %v", mapVal)
	}

	// Test MapDefault()
	r = &result{
		config: config,
		key:    "TEST_MAP",
		value:  "",
		err:    nil,
	}

	defaultMap := map[string]string{"default1": "value1", "default2": "value2"}
	mapVal = r.MapDefault(defaultMap)
	if !equalMaps(mapVal, defaultMap) {
		t.Errorf("Expected MapDefault() to return default map, got: %v", mapVal)
	}
}

// TestResultWithExistingError tests that an existing error isn't overwritten
func TestResultWithExistingError(t *testing.T) {
	// Set up a result with an existing error
	existingErr := errors.New("existing error")
	r := &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "",
		err:    existingErr,
	}

	// Test that Int() preserves the error
	_, err := r.Int()
	if err != existingErr {
		t.Errorf("Expected existing error to be returned, got: %v", err)
	}

	// Test that Float64() preserves the error
	_, err = r.Float64()
	if err != existingErr {
		t.Errorf("Expected existing error to be returned, got: %v", err)
	}

	// Test that Duration() preserves the error
	_, err = r.Duration()
	if err != existingErr {
		t.Errorf("Expected existing error to be returned, got: %v", err)
	}
}

// TestResultEmptyValue tests behavior with empty values
func TestResultEmptyValue(t *testing.T) {
	// Set up a result with an empty value
	r := &result{
		config: &Config{},
		key:    "TEST_KEY",
		value:  "",
		err:    nil,
	}

	// Empty value should return error for Int()
	_, err := r.Int()
	if err == nil {
		t.Error("Expected error for empty value in Int(), got nil")
	}

	// Empty value should return error for Float64()
	_, err = r.Float64()
	if err == nil {
		t.Error("Expected error for empty value in Float64(), got nil")
	}

	// Empty value should return error for Duration()
	_, err = r.Duration()
	if err == nil {
		t.Error("Expected error for empty value in Duration(), got nil")
	}

	// Empty value should return empty string for String()
	if r.String() != "" {
		t.Errorf("Expected empty string from String(), got: '%s'", r.String())
	}

	// Empty value should return false for Bool()
	if r.Bool() != false {
		t.Error("Expected false from Bool() with empty value, got true")
	}

	// Empty value should return empty slice from Slice()
	slice := r.Slice(",")
	if len(slice) != 0 {
		t.Errorf("Expected empty slice from Slice(), got: %v", slice)
	}

	// Empty value should return empty map from Map()
	m := r.Map()
	if len(m) != 0 {
		t.Errorf("Expected empty map from Map(), got: %v", m)
	}
}

// TestResultInvalidValues tests behavior with invalid values
func TestResultInvalidValues(t *testing.T) {
	// Invalid int
	rInt := &result{
		config: &Config{},
		key:    "TEST_INT",
		value:  "not-an-int",
		err:    nil,
	}

	_, err := rInt.Int()
	if err == nil {
		t.Error("Expected error for invalid int, got nil")
	}

	// Using IntDefault should return the default value for invalid int
	intVal := rInt.IntDefault(42)
	if intVal != 42 {
		t.Errorf("Expected default value 42 for invalid int, got %d", intVal)
	}

	// Invalid float
	rFloat := &result{
		config: &Config{},
		key:    "TEST_FLOAT",
		value:  "not-a-float",
		err:    nil,
	}

	_, err = rFloat.Float64()
	if err == nil {
		t.Error("Expected error for invalid float, got nil")
	}

	// Using Float64Default should return the default value for invalid float
	floatVal := rFloat.Float64Default(3.14)
	if floatVal != 3.14 {
		t.Errorf("Expected default value 3.14 for invalid float, got %f", floatVal)
	}

	// Invalid duration
	rDuration := &result{
		config: &Config{},
		key:    "TEST_DURATION",
		value:  "not-a-duration",
		err:    nil,
	}

	_, err = rDuration.Duration()
	if err == nil {
		t.Error("Expected error for invalid duration, got nil")
	}

	// Using DurationDefault should return the default value for invalid duration
	durationVal := rDuration.DurationDefault(5 * time.Second)
	if durationVal != 5*time.Second {
		t.Errorf("Expected default value 5s for invalid duration, got %v", durationVal)
	}
}

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
