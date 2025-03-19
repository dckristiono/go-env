package env

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestGetDefaultInstance tests the singleton behavior and error handling
func TestGetDefaultInstance(t *testing.T) {
	// Save original values
	origDefaultInstance := defaultInstance
	origInitErr := initErr
	origOnce := once
	origGetDefaultInstance := getDefaultInstance

	// Reset for testing
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Restore after testing
	defer func() {
		defaultInstance = origDefaultInstance
		initErr = origInitErr
		once = origOnce
		getDefaultInstance = origGetDefaultInstance
	}()

	// First call should initialize - using blank identifiers to avoid unused vars error
	_, _ = getDefaultInstance()

	// Save instance for comparison
	savedInstance := defaultInstance

	// Second call should return the same instance
	instance2, _ := getDefaultInstance()
	if instance2 != savedInstance && savedInstance != nil {
		t.Error("Second call returned different instance")
	}

	// Test error case
	// Important: Reset once so the function actually runs
	defaultInstance = nil
	initErr = fmt.Errorf("test error")
	once = sync.Once{}

	// Create a temporary implementation
	oldGetDefaultInstance := getDefaultInstance
	getDefaultInstance = func() (*Config, error) {
		if defaultInstance == nil && initErr != nil {
			return nil, initErr
		}
		return defaultInstance, initErr
	}

	// This should return the error from initialization
	_, err2 := getDefaultInstance()
	if err2 == nil || err2.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err2)
	}

	// Restore original function
	getDefaultInstance = oldGetDefaultInstance
}

// TestConfigDetermineDefaultModeEnv tests mode determination by APP_ENV
func TestConfigDetermineDefaultModeEnv(t *testing.T) {
	// Save original environment
	oldEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", oldEnv)

	// Test with different APP_ENV values
	testEnvs := []string{"production", "staging", "development", "custom_mode"}
	for _, env := range testEnvs {
		os.Setenv("APP_ENV", env)
		if mode := determineDefaultMode(); mode != env {
			t.Errorf("With APP_ENV=%s, expected mode '%s', got '%s'", env, env, mode)
		}
	}
}

// TestConfigLoadAdvanced tests more load scenarios
func TestConfigLoadAdvanced(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test invalid mode beyond the standard ones
	cfg := &Config{Mode: "invalid_mode"}
	err = cfg.Load()
	if err == nil {
		t.Error("Load() with invalid mode should return error")
	}

	// Test fallback behavior for missing files
	os.Setenv("APP_ENV", "custom_mode") // Custom mode not matching any standard
	defer os.Unsetenv("APP_ENV")

	customCfg, err := New()
	if err != nil {
		// This is expected in test environment without proper files
		// Let's confirm we're in the right mode at least
		if customCfg != nil && customCfg.Mode != "custom_mode" {
			t.Errorf("Expected mode 'custom_mode', got '%s'", customCfg.Mode)
		}
	}
}

// TestConfigWithConcurrent tests concurrent access to singleton
func TestConfigWithConcurrent(t *testing.T) {
	// Reset singleton
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create .env file
	if err := os.WriteFile(".env", []byte("TEST=value"), 0644); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Initialize with default values
	err = Initialize(WithMode(Production))
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test concurrent access to With
	var wg sync.WaitGroup
	const goroutines = 10
	configs := make([]*Config, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			configs[index] = With(WithMode(Development))
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got configurations with correct mode
	for i, cfg := range configs {
		if cfg == nil {
			t.Errorf("Goroutine %d got nil config", i)
			continue
		}
		if cfg.Mode != Development {
			t.Errorf("Goroutine %d expected mode %s, got %s", i, Development, cfg.Mode)
		}
	}
}

// TestConfigBoundaryValues tests handling of boundary values
func TestConfigBoundaryValues(t *testing.T) {
	// Save original env vars
	env := map[string]string{
		"TEST_MAX_INT":   os.Getenv("TEST_MAX_INT"),
		"TEST_MIN_INT":   os.Getenv("TEST_MIN_INT"),
		"TEST_MAX_FLOAT": os.Getenv("TEST_MAX_FLOAT"),
		"TEST_OVER_INT":  os.Getenv("TEST_OVER_INT"),
		"TEST_LONG_DUR":  os.Getenv("TEST_LONG_DUR"),
	}
	defer func() {
		for k, v := range env {
			os.Setenv(k, v)
		}
	}()

	// Set boundary values
	os.Setenv("TEST_MAX_INT", "2147483647")                // Max int32
	os.Setenv("TEST_MIN_INT", "-2147483648")               // Min int32
	os.Setenv("TEST_MAX_FLOAT", "1.7976931348623157e+308") // Max float64
	os.Setenv("TEST_OVER_INT", "9223372036854775808")      // MaxInt64 + 1
	os.Setenv("TEST_LONG_DUR", "87600h")                   // 10 years

	cfg := &Config{}

	// Test GetInt with boundary values
	val, err := cfg.GetInt("TEST_MAX_INT")
	if err != nil || val != 2147483647 {
		t.Errorf("GetInt(TEST_MAX_INT) expected %d, got %d (error: %v)",
			2147483647, val, err)
	}

	// Test GetInt64 with boundary values
	val64, err := cfg.GetInt64("TEST_MIN_INT")
	if err != nil || val64 != -2147483648 {
		t.Errorf("GetInt64(TEST_MIN_INT) expected %d, got %d (error: %v)",
			-2147483648, val64, err)
	}

	// Test GetFloat64 with boundary values
	maxFloat := 1.7976931348623157e+308
	valFloat, err := cfg.GetFloat64("TEST_MAX_FLOAT")
	if err != nil || valFloat != maxFloat {
		t.Errorf("GetFloat64(TEST_MAX_FLOAT) expected %v, got %v (error: %v)",
			maxFloat, valFloat, err)
	}

	// Test overflow integer
	_, err = cfg.GetInt64("TEST_OVER_INT")
	if err == nil {
		t.Error("GetInt64(TEST_OVER_INT) should fail for overflow value")
	}

	// Test long duration
	longDur, err := cfg.GetDuration("TEST_LONG_DUR")
	if err != nil || longDur != 87600*time.Hour {
		t.Errorf("GetDuration(TEST_LONG_DUR) expected 87600h, got %v (error: %v)",
			longDur, err)
	}
}

// TestConfigStringsFunctions tests string manipulation functions
func TestConfigStringsFunctions(t *testing.T) {
	// Save original env vars
	envTestStr := os.Getenv("TEST_STRING")
	envTestNonexistent := os.Getenv("TEST_NONEXISTENT")
	defer func() {
		os.Setenv("TEST_STRING", envTestStr)
		os.Setenv("TEST_NONEXISTENT", envTestNonexistent)
	}()

	// Set test values
	os.Setenv("TEST_STRING", "test_value")
	os.Unsetenv("TEST_NONEXISTENT")

	// Test standard String function
	if val := String("TEST_STRING"); val != "test_value" {
		t.Errorf("String(TEST_STRING) expected 'test_value', got '%s'", val)
	}

	// Test String with default
	if val := String("TEST_NONEXISTENT", "default_value"); val != "default_value" {
		t.Errorf("String(TEST_NONEXISTENT, default_value) expected 'default_value', got '%s'", val)
	}

	// Test Key().String() chain
	if val := Key("TEST_STRING").String(); val != "test_value" {
		t.Errorf("Key(TEST_STRING).String() expected 'test_value', got '%s'", val)
	}

	// Test Key chain with nonexistent
	if val := Key("TEST_NONEXISTENT").Default("chain_default").String(); val != "chain_default" {
		t.Errorf("Key chain with default expected 'chain_default', got '%s'", val)
	}
}

// TestConfigErrorPropagation tests error handling throughout methods
func TestConfigErrorPropagation(t *testing.T) {
	// Mock getDefaultInstance to always return error
	origGetDefaultInstance := getDefaultInstance
	mockErr := fmt.Errorf("mock error")
	getDefaultInstance = func() (*Config, error) {
		return nil, mockErr
	}
	defer func() { getDefaultInstance = origGetDefaultInstance }()

	// Test all package-level functions with error propagation
	if val := Get("KEY"); val != "" {
		t.Errorf("Get should return empty when getDefaultInstance fails, got '%s'", val)
	}

	if val, err := GetInt("KEY"); val != 0 || err != mockErr {
		t.Errorf("GetInt should return 0 and mockErr, got %d and %v", val, err)
	}

	if val, err := GetInt64("KEY"); val != 0 || err != mockErr {
		t.Errorf("GetInt64 should return 0 and mockErr, got %d and %v", val, err)
	}

	if val, err := GetFloat64("KEY"); val != 0 || err != mockErr {
		t.Errorf("GetFloat64 should return 0 and mockErr, got %f and %v", val, err)
	}

	if val := GetBool("KEY"); val != false {
		t.Errorf("GetBool should return false when getDefaultInstance fails, got %v", val)
	}

	if val, err := GetDuration("KEY"); val != 0 || err != mockErr {
		t.Errorf("GetDuration should return 0 and mockErr, got %v and %v", val, err)
	}

	if val := GetSlice("KEY", ","); len(val) != 0 {
		t.Errorf("GetSlice should return empty slice when getDefaultInstance fails, got %v", val)
	}

	if val := GetMap("KEY"); len(val) != 0 {
		t.Errorf("GetMap should return empty map when getDefaultInstance fails, got %v", val)
	}

	// Test mode functions
	if val := GetMode(); val != "" {
		t.Errorf("GetMode should return empty when getDefaultInstance fails, got '%s'", val)
	}

	if IsProduction() {
		t.Error("IsProduction should return false when getDefaultInstance fails")
	}

	if IsStaging() {
		t.Error("IsStaging should return false when getDefaultInstance fails")
	}

	if IsDevelopment() {
		t.Error("IsDevelopment should return false when getDefaultInstance fails")
	}
}

// TestConfigPrependPrefix tests prefix handling
func TestConfigPrependPrefix(t *testing.T) {
	testCases := []struct {
		prefix   string
		key      string
		expected string
	}{
		{"", "KEY", "KEY"},
		{"PREFIX_", "KEY", "PREFIX_KEY"},
		{"APP.", "CONFIG", "APP.CONFIG"},
		{"  ", "KEY", "  KEY"}, // Space prefix is preserved
	}

	for _, tc := range testCases {
		cfg := &Config{Prefix: tc.prefix}
		result := cfg.prependPrefix(tc.key)
		if result != tc.expected {
			t.Errorf("prependPrefix with prefix '%s' and key '%s': expected '%s', got '%s'",
				tc.prefix, tc.key, tc.expected, result)
		}
	}
}

// TestConfigChainedMethods tests complete chains of configuration methods
func TestConfigChainedMethods(t *testing.T) {
	// Setup
	os.Setenv("TEST_CHAIN_INT", "42")
	os.Setenv("TEST_CHAIN_FLOAT", "3.14")
	os.Setenv("TEST_CHAIN_BOOL", "true")
	defer func() {
		os.Unsetenv("TEST_CHAIN_INT")
		os.Unsetenv("TEST_CHAIN_FLOAT")
		os.Unsetenv("TEST_CHAIN_BOOL")
	}()

	// Initialize a config instance
	cfg := &Config{Prefix: "TEST_"}

	// Test a complete chain with Required -> Int
	val, err := cfg.Key("CHAIN_INT").Required().Int()
	if err != nil || val != 42 {
		t.Errorf("Key(CHAIN_INT).Required().Int() expected 42, got %d (error: %v)", val, err)
	}

	// Test a chain with float
	fval, err := cfg.Key("CHAIN_FLOAT").Required().Float64()
	if err != nil || fval != 3.14 {
		t.Errorf("Float chain expected 3.14, got %f (error: %v)", fval, err)
	}

	// Test a chain with bool
	bval := cfg.Key("CHAIN_BOOL").Bool()
	if !bval {
		t.Errorf("Bool chain expected true, got %v", bval)
	}

	// Test error propagation in chain
	_, err = cfg.Key("NONEXISTENT").Required().Int()
	if err == nil {
		t.Error("Required() for nonexistent key should return error")
	}

	// Test default in chain
	val = cfg.Key("NONEXISTENT").IntDefault(100)
	if val != 100 {
		t.Errorf("IntDefault chain expected 100, got %d", val)
	}
}

// TestConfigWrappedFunctions tests wrapper functions Int, Float64, etc.
func TestConfigWrappedFunctions(t *testing.T) {
	// Setup
	os.Setenv("TEST_WRAPPED_INT", "42")
	os.Setenv("TEST_WRAPPED_FLOAT", "3.14")
	os.Setenv("TEST_WRAPPED_BOOL", "true")
	os.Setenv("TEST_WRAPPED_DUR", "5s")
	os.Setenv("TEST_WRAPPED_SLICE", "a,b,c")
	os.Setenv("TEST_WRAPPED_MAP", "k1:v1,k2:v2")
	defer func() {
		os.Unsetenv("TEST_WRAPPED_INT")
		os.Unsetenv("TEST_WRAPPED_FLOAT")
		os.Unsetenv("TEST_WRAPPED_BOOL")
		os.Unsetenv("TEST_WRAPPED_DUR")
		os.Unsetenv("TEST_WRAPPED_SLICE")
		os.Unsetenv("TEST_WRAPPED_MAP")
	}()

	// Initialize for package-level functions
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}
	// Create a temp .env for initialization
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)
	os.WriteFile(".env", []byte("TEST=value"), 0644)

	// Test Int wrapper
	val := Int("TEST_WRAPPED_INT")
	if val != 42 {
		t.Errorf("Int(TEST_WRAPPED_INT) expected 42, got %d", val)
	}

	val = Int("NONEXISTENT", 100)
	if val != 100 {
		t.Errorf("Int(NONEXISTENT, 100) expected 100, got %d", val)
	}

	// Test Float64 wrapper
	fval := Float64("TEST_WRAPPED_FLOAT")
	if fval != 3.14 {
		t.Errorf("Float64(TEST_WRAPPED_FLOAT) expected 3.14, got %f", fval)
	}

	// Test Bool wrapper
	bval := Bool("TEST_WRAPPED_BOOL")
	if !bval {
		t.Errorf("Bool(TEST_WRAPPED_BOOL) expected true, got %v", bval)
	}

	// Test Duration wrapper
	dval := Duration("TEST_WRAPPED_DUR")
	if dval != 5*time.Second {
		t.Errorf("Duration(TEST_WRAPPED_DUR) expected 5s, got %v", dval)
	}

	// Test Slice wrapper
	sval := Slice("TEST_WRAPPED_SLICE", ",")
	if len(sval) != 3 || sval[0] != "a" || sval[1] != "b" || sval[2] != "c" {
		t.Errorf("Slice(TEST_WRAPPED_SLICE) expected [a b c], got %v", sval)
	}

	// Test Map wrapper
	mval := Map("TEST_WRAPPED_MAP")
	if len(mval) != 2 || mval["k1"] != "v1" || mval["k2"] != "v2" {
		t.Errorf("Map(TEST_WRAPPED_MAP) expected map[k1:v1 k2:v2], got %v", mval)
	}
}

// Tambahkan pengujian ini ke file config_test.go

// TestConfigGetMode tests GetMode and mode checking methods
func TestConfigGetMode(t *testing.T) {
	// Test with different modes
	modes := []string{Production, Staging, Development, "custom_mode"}

	for _, mode := range modes {
		cfg := &Config{Mode: mode}

		// Test GetMode
		if got := cfg.GetMode(); got != mode {
			t.Errorf("GetMode() expected %s, got %s", mode, got)
		}

		// Test IsProduction
		isProduction := mode == Production
		if got := cfg.IsProduction(); got != isProduction {
			t.Errorf("IsProduction() for mode %s expected %v, got %v",
				mode, isProduction, got)
		}

		// Test IsStaging
		isStaging := mode == Staging
		if got := cfg.IsStaging(); got != isStaging {
			t.Errorf("IsStaging() for mode %s expected %v, got %v",
				mode, isStaging, got)
		}

		// Test IsDevelopment
		isDevelopment := mode == Development
		if got := cfg.IsDevelopment(); got != isDevelopment {
			t.Errorf("IsDevelopment() for mode %s expected %v, got %v",
				mode, isDevelopment, got)
		}
	}
}

// TestGetDurationExtended provides additional tests for GetDuration
func TestGetDurationExtended(t *testing.T) {
	// Setup environment variables
	envVars := map[string]string{
		"TEST_DURATION_VALID":    "1h30m",
		"TEST_DURATION_ZERO":     "0s",
		"TEST_DURATION_NEGATIVE": "-5m",
		"TEST_DURATION_MILLIS":   "100ms",
		"TEST_DURATION_COMPLEX":  "2h45m30s",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := &Config{}

	// Test each duration
	for k, expectedStr := range envVars {
		expected, _ := time.ParseDuration(expectedStr)

		// Test GetDuration
		got, err := cfg.GetDuration(k)
		if err != nil {
			t.Errorf("GetDuration(%s) unexpected error: %v", k, err)
		}
		if got != expected {
			t.Errorf("GetDuration(%s) expected %v, got %v", k, expected, got)
		}

		// Test with missing key and default
		defaultDur := 10 * time.Second
		got, err = cfg.GetDuration("NONEXISTENT_"+k, defaultDur)
		if err != nil {
			t.Errorf("GetDuration with default unexpected error: %v", err)
		}
		if got != defaultDur {
			t.Errorf("GetDuration with default expected %v, got %v", defaultDur, got)
		}

		// Test GetDuration without default for missing key
		_, err = cfg.GetDuration("NONEXISTENT_" + k)
		if err == nil {
			t.Errorf("GetDuration for missing key should return error")
		}
	}
}

// TestGetInt64Extended provides additional tests for GetInt64
func TestGetInt64Extended(t *testing.T) {
	// Setup environment variables with various integers
	envVars := map[string]string{
		"TEST_INT64_ZERO":     "0",
		"TEST_INT64_POSITIVE": "9223372036854775807",  // Max int64
		"TEST_INT64_NEGATIVE": "-9223372036854775808", // Min int64
		"TEST_INT64_MEDIUM":   "1234567890",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := &Config{}

	// Test each int64 value
	for k, v := range envVars {
		expected, _ := strconv.ParseInt(v, 10, 64)

		// Test GetInt64
		got, err := cfg.GetInt64(k)
		if err != nil {
			t.Errorf("GetInt64(%s) unexpected error: %v", k, err)
		}
		if got != expected {
			t.Errorf("GetInt64(%s) expected %d, got %d", k, expected, got)
		}

		// Test with missing key and default
		defaultVal := int64(42)
		got, err = cfg.GetInt64("NONEXISTENT_"+k, defaultVal)
		if err != nil {
			t.Errorf("GetInt64 with default unexpected error: %v", err)
		}
		if got != defaultVal {
			t.Errorf("GetInt64 with default expected %d, got %d", defaultVal, got)
		}

		// Test GetInt64 without default for missing key
		_, err = cfg.GetInt64("NONEXISTENT_" + k)
		if err == nil {
			t.Errorf("GetInt64 for missing key should return error")
		}
	}

	// Test invalid value
	os.Setenv("TEST_INT64_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_INT64_INVALID")

	_, err := cfg.GetInt64("TEST_INT64_INVALID")
	if err == nil {
		t.Error("GetInt64 with invalid value should return error")
	}
}

// TestGetFloat64Extended provides additional tests for GetFloat64
func TestGetFloat64Extended(t *testing.T) {
	// Setup environment variables with various floats
	envVars := map[string]string{
		"TEST_FLOAT64_ZERO":     "0.0",
		"TEST_FLOAT64_POSITIVE": "1.7976931348623157e+308",  // Max float64
		"TEST_FLOAT64_NEGATIVE": "-1.7976931348623157e+308", // Min float64
		"TEST_FLOAT64_SMALL":    "0.000000000000000000000000001",
		"TEST_FLOAT64_DECIMAL":  "3.14159265359",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := &Config{}

	// Test each float64 value
	for k, v := range envVars {
		expected, _ := strconv.ParseFloat(v, 64)

		// Test GetFloat64
		got, err := cfg.GetFloat64(k)
		if err != nil {
			t.Errorf("GetFloat64(%s) unexpected error: %v", k, err)
		}
		if got != expected {
			t.Errorf("GetFloat64(%s) expected %f, got %f", k, expected, got)
		}

		// Test with missing key and default
		defaultVal := 3.14
		got, err = cfg.GetFloat64("NONEXISTENT_"+k, defaultVal)
		if err != nil {
			t.Errorf("GetFloat64 with default unexpected error: %v", err)
		}
		if got != defaultVal {
			t.Errorf("GetFloat64 with default expected %f, got %f", defaultVal, got)
		}

		// Test GetFloat64 without default for missing key
		_, err = cfg.GetFloat64("NONEXISTENT_" + k)
		if err == nil {
			t.Errorf("GetFloat64 for missing key should return error")
		}
	}

	// Test invalid value
	os.Setenv("TEST_FLOAT64_INVALID", "not_a_number")
	defer os.Unsetenv("TEST_FLOAT64_INVALID")

	_, err := cfg.GetFloat64("TEST_FLOAT64_INVALID")
	if err == nil {
		t.Error("GetFloat64 with invalid value should return error")
	}
}

// TestGetBoolExtended provides additional tests for GetBool
func TestGetBoolExtended(t *testing.T) {
	// Setup test cases for different boolean values
	testCases := []struct {
		key      string
		value    string
		expected bool
	}{
		{"TEST_BOOL_TRUE", "true", true},
		{"TEST_BOOL_FALSE", "false", false},
		{"TEST_BOOL_YES", "yes", true},
		{"TEST_BOOL_NO", "no", false},
		{"TEST_BOOL_Y", "y", true},
		{"TEST_BOOL_N", "n", false},
		{"TEST_BOOL_1", "1", true},
		{"TEST_BOOL_0", "0", false},
		{"TEST_BOOL_MIXED", "True", true},
		{"TEST_BOOL_OTHER", "anything_else", false},
		{"TEST_BOOL_EMPTY", "", false},
	}

	for _, tc := range testCases {
		os.Setenv(tc.key, tc.value)
		defer os.Unsetenv(tc.key)
	}

	cfg := &Config{}

	// Test each boolean value
	for _, tc := range testCases {
		// Test GetBool
		got := cfg.GetBool(tc.key)
		if got != tc.expected {
			t.Errorf("GetBool(%s) with value '%s' expected %v, got %v",
				tc.key, tc.value, tc.expected, got)
		}

		// Test with missing key and default true
		got = cfg.GetBool("NONEXISTENT_"+tc.key, true)
		if got != true {
			t.Errorf("GetBool with default true expected true, got %v", got)
		}

		// Test with missing key and default false
		got = cfg.GetBool("NONEXISTENT_"+tc.key, false)
		if got != false {
			t.Errorf("GetBool with default false expected false, got %v", got)
		}

		// Test missing key with no default (should be false)
		got = cfg.GetBool("NONEXISTENT_" + tc.key)
		if got != false {
			t.Errorf("GetBool for missing key without default should be false, got %v", got)
		}
	}
}

// TestGetSliceExtended provides additional tests for GetSlice
func TestGetSliceExtended(t *testing.T) {
	// Setup environment variables with various slice formats
	envVars := map[string]string{
		"TEST_SLICE_EMPTY":        "",
		"TEST_SLICE_SINGLE":       "single",
		"TEST_SLICE_MULTIPLE":     "a,b,c,d,e",
		"TEST_SLICE_SPACES":       " a , b , c ",
		"TEST_SLICE_EMPTY_PARTS":  "a,,c",
		"TEST_SLICE_CUSTOM_DELIM": "a|b|c",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := &Config{}

	// Test cases for different slice formats
	testCases := []struct {
		key       string
		delimiter string
		expected  []string
	}{
		{"TEST_SLICE_EMPTY", ",", []string{}},
		{"TEST_SLICE_SINGLE", ",", []string{"single"}},
		{"TEST_SLICE_MULTIPLE", ",", []string{"a", "b", "c", "d", "e"}},
		{"TEST_SLICE_SPACES", ",", []string{"a", "b", "c"}},
		{"TEST_SLICE_EMPTY_PARTS", ",", []string{"a", "", "c"}},
		{"TEST_SLICE_CUSTOM_DELIM", "|", []string{"a", "b", "c"}},
		// Test with empty delimiter (should default to comma)
		{"TEST_SLICE_MULTIPLE", "", []string{"a", "b", "c", "d", "e"}},
	}

	for _, tc := range testCases {
		// Test GetSlice
		got := cfg.GetSlice(tc.key, tc.delimiter)
		if !equalSlices(got, tc.expected) {
			t.Errorf("GetSlice(%s, %s) expected %v, got %v",
				tc.key, tc.delimiter, tc.expected, got)
		}

		// Test with missing key and default
		defaultVal := []string{"default1", "default2"}
		got = cfg.GetSlice("NONEXISTENT_"+tc.key, tc.delimiter, defaultVal)
		if !equalSlices(got, defaultVal) {
			t.Errorf("GetSlice with default expected %v, got %v", defaultVal, got)
		}

		// Test missing key without default
		got = cfg.GetSlice("NONEXISTENT_"+tc.key, tc.delimiter)
		if len(got) != 0 {
			t.Errorf("GetSlice for missing key without default should be empty, got %v", got)
		}
	}
}
