package env

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// TestKeyPackageLevelAdvanced tests the Key package level function with advanced scenarios
func TestKeyPackageLevelAdvanced(t *testing.T) {
	// Set up
	os.Setenv("TEST_KEY_PKG", "pkg_value")
	defer os.Unsetenv("TEST_KEY_PKG")

	// Create a config
	config := &Config{}

	// Mock getDefaultInstance
	origGetDefaultInstance := getDefaultInstance
	getDefaultInstance = func() (*Config, error) {
		return config, nil
	}

	defer func() {
		getDefaultInstance = origGetDefaultInstance
	}()

	// Test Key package function
	result := Key("TEST_KEY_PKG")
	if result.String() != "pkg_value" {
		t.Errorf("Expected 'pkg_value', got '%s'", result.String())
	}

	// Test with error in getDefaultInstance
	getDefaultInstance = func() (*Config, error) {
		return nil, fmt.Errorf("mock error")
	}

	result = Key("ANY_KEY")
	if result.err == nil {
		t.Error("Expected error when getDefaultInstance fails, got nil")
	}
}

// TestOptionsComposition tests advanced composition of options
func TestOptionsComposition(t *testing.T) {
	// Create a base config
	config := &Config{
		Mode:   Production,
		Prefix: "",
	}

	// Create slice of options
	options := []ConfigOption{
		WithMode(Development),
		WithPrefix("TEST_"),
	}

	// Apply them one by one
	for _, opt := range options {
		opt(config)
	}

	// Verify all options were applied
	if config.Mode != Development {
		t.Errorf("Expected Mode=%s, got %s", Development, config.Mode)
	}

	if config.Prefix != "TEST_" {
		t.Errorf("Expected Prefix=TEST_, got %s", config.Prefix)
	}
}

// TestAdvancedFromChaining tests more complex From chaining scenarios
func TestAdvancedFromChaining(t *testing.T) {
	// Create initial config
	config := &Config{
		Mode:   Production,
		Prefix: "",
	}

	// Chain multiple From calls with different options
	result := config.
		From(WithPrefix("PREFIX1_")).
		From(WithMode(Staging)).
		From(WithPrefix("PREFIX2_"))

	// Original config should be unchanged
	if config.Mode != Production || config.Prefix != "" {
		t.Errorf("Original config was modified: mode=%s, prefix=%s", config.Mode, config.Prefix)
	}

	// Result should have the last options applied
	if result.Mode != Staging || result.Prefix != "PREFIX2_" {
		t.Errorf("Expected Mode=%s, Prefix=PREFIX2_, got Mode=%s, Prefix=%s",
			Staging, result.Mode, result.Prefix)
	}
}

// TestWithFunctionEdgeCases tests edge cases for the With function
func TestWithFunctionEdgeCases(t *testing.T) {
	// Test with a mock default instance
	origGetDefaultInstance := getDefaultInstance

	// Setup a mock for success case
	mockConfig := &Config{Mode: Production, Prefix: "PROD_"}
	getDefaultInstance = func() (*Config, error) {
		return mockConfig, nil
	}

	// Create a temporary file for tests needing file access
	tmpFile := ".env"
	f, err := os.Create(tmpFile)
	if err == nil {
		f.Close()
		defer os.Remove(tmpFile)

		// Test With with a working default instance
		result := With(WithMode(Development))

		// Original mock should be unchanged
		if mockConfig.Mode != Production {
			t.Errorf("Original config modified, expected Mode=%s, got %s", Production, mockConfig.Mode)
		}

		// New config should have Development mode
		if result.Mode != Development {
			t.Errorf("Expected Mode=%s, got %s", Development, result.Mode)
		}

		// Test With when getDefaultInstance returns error
		getDefaultInstance = func() (*Config, error) {
			return nil, fmt.Errorf("mock error")
		}

		// The With function should create a new config
		result = With(WithMode(Staging))
		if result == nil {
			t.Error("With() returned nil instead of creating new config")
		} else if result.Mode != Staging {
			t.Errorf("Expected Mode=%s, got %s", Staging, result.Mode)
		}
	} else {
		t.Logf("Skipping file-dependent tests: %v", err)
	}

	// Reset the original function
	getDefaultInstance = origGetDefaultInstance
}

// TestInitializeWithMultipleOptions tests Initialize with multiple options
func TestInitializeWithMultipleOptions(t *testing.T) {
	// Save original state
	origDefaultInstance := defaultInstance
	origInitErr := initErr
	origOnce := once

	// Reset singleton state
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Restore after test
	defer func() {
		defaultInstance = origDefaultInstance
		initErr = origInitErr
		once = origOnce
	}()

	// Create temporary file for New to succeed
	tmpFile := ".env"
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Skipf("Skipping test: Failed to create temp file: %v", err)
		return
	}
	f.Close()
	defer os.Remove(tmpFile)

	// Test Initialize with multiple options
	options := []ConfigOption{
		WithMode(Staging),
		WithPrefix("STAGE_"),
	}

	err = Initialize(options...)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify all options were applied to defaultInstance
	if defaultInstance == nil {
		t.Fatal("defaultInstance is nil after Initialize")
	}

	if defaultInstance.Mode != Staging {
		t.Errorf("Expected Mode=%s, got %s", Staging, defaultInstance.Mode)
	}

	if defaultInstance.Prefix != "STAGE_" {
		t.Errorf("Expected Prefix=STAGE_, got %s", defaultInstance.Prefix)
	}
}
