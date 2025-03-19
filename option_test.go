package env

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// TestConfigOptionComposition tests composing multiple options
func TestConfigOptionComposition(t *testing.T) {
	// Create a base config
	config := &Config{
		Mode:   Production,
		Prefix: "",
	}

	// Create multiple options
	options := []ConfigOption{
		WithMode(Development),
		WithMode(Staging), // Last one should win
		WithPrefix("PRE_"),
		WithPrefix("TEST_"), // Last one should win
	}

	// Apply them one by one
	for _, opt := range options {
		opt(config)
	}

	// Verify final state - last options should win
	if config.Mode != Staging {
		t.Errorf("Expected Mode=%s, got %s", Staging, config.Mode)
	}

	if config.Prefix != "TEST_" {
		t.Errorf("Expected Prefix=TEST_, got %s", config.Prefix)
	}
}

// TestConfigOptionCustom tests creating and using custom options
func TestConfigOptionCustom(t *testing.T) {
	// Create a base config
	config := &Config{
		Mode:   Production,
		Prefix: "",
	}

	// Create custom options
	customModeOption := func(c *Config) {
		c.Mode = "custom_mode"
	}

	customPrefixOption := func(c *Config) {
		c.Prefix = "CUSTOM_"
	}

	// Apply standard and custom options in sequence
	WithMode(Development)(config)
	customModeOption(config)
	WithPrefix("TEST_")(config)
	customPrefixOption(config)

	// Verify final state - last applied option wins
	if config.Mode != "custom_mode" {
		t.Errorf("Expected Mode=custom_mode, got %s", config.Mode)
	}

	if config.Prefix != "CUSTOM_" {
		t.Errorf("Expected Prefix=CUSTOM_, got %s", config.Prefix)
	}
}

// TestWithModeVariations tests WithMode with various inputs
func TestWithModeVariations(t *testing.T) {
	// Test cases
	testCases := []struct {
		mode     string
		expected string
	}{
		{Production, Production},
		{Staging, Staging},
		{Development, Development},
		{"", ""},                       // Empty string
		{"custom_mode", "custom_mode"}, // Custom mode
		{"  spaced  ", "  spaced  "},   // Spaces preserved
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Mode=%s", tc.mode), func(t *testing.T) {
			config := &Config{}
			WithMode(tc.mode)(config)

			if config.Mode != tc.expected {
				t.Errorf("WithMode(%s): expected %s, got %s",
					tc.mode, tc.expected, config.Mode)
			}
		})
	}
}

// TestWithPrefixVariations tests WithPrefix with various inputs
func TestWithPrefixVariations(t *testing.T) {
	// Test cases
	testCases := []struct {
		prefix   string
		expected string
	}{
		{"TEST_", "TEST_"},
		{"", ""},                   // Empty string
		{"app.", "app."},           // With dot
		{"123_", "123_"},           // With numbers
		{"  ", "  "},               // Spaces
		{"特殊前缀_", "特殊前缀_"}, // Unicode
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Prefix=%s", tc.prefix), func(t *testing.T) {
			config := &Config{}
			WithPrefix(tc.prefix)(config)

			if config.Prefix != tc.expected {
				t.Errorf("WithPrefix(%s): expected %s, got %s",
					tc.prefix, tc.expected, config.Prefix)
			}
		})
	}
}

// TestChainedFromWithOptions tests complex chaining with From and options
func TestChainedFromWithOptions(t *testing.T) {
	// Initial config
	config := &Config{
		Mode:   Production,
		Prefix: "PROD_",
	}

	// Complex chaining
	result := config.
		From(WithMode(Staging)).
		From(WithPrefix("STAGE_")).
		From(WithMode(Development)).
		From(WithPrefix("DEV_"))

	// Original config should be unchanged
	if config.Mode != Production || config.Prefix != "PROD_" {
		t.Errorf("Original config modified: Mode=%s, Prefix=%s",
			config.Mode, config.Prefix)
	}

	// Each From creates a new config, only last options should apply
	if result.Mode != Development {
		t.Errorf("Expected Mode=%s, got %s", Development, result.Mode)
	}

	if result.Prefix != "DEV_" {
		t.Errorf("Expected Prefix=DEV_, got %s", result.Prefix)
	}
}

// TestWithFunctionOptions tests the With function with various option combinations
func TestWithFunctionOptions(t *testing.T) {
	// Save original function
	origGetDefaultInstance := getDefaultInstance
	defer func() { getDefaultInstance = origGetDefaultInstance }()

	// Test with mock default instance
	mockConfig := &Config{Mode: Production, Prefix: ""}
	getDefaultInstance = func() (*Config, error) {
		return mockConfig, nil
	}

	// Test With with single option
	result1 := With(WithMode(Development))
	if result1.Mode != Development {
		t.Errorf("With(WithMode): expected %s, got %s", Development, result1.Mode)
	}

	// Test With with multiple options
	result2 := With(WithMode(Staging), WithPrefix("TEST_"))
	if result2.Mode != Staging || result2.Prefix != "TEST_" {
		t.Errorf("With(multiple): expected Mode=%s,Prefix=TEST_, got Mode=%s,Prefix=%s",
			Staging, result2.Mode, result2.Prefix)
	}

	// Original mock should be unchanged
	if mockConfig.Mode != Production || mockConfig.Prefix != "" {
		t.Errorf("Original mockConfig modified: Mode=%s, Prefix=%s",
			mockConfig.Mode, mockConfig.Prefix)
	}

	// Test With when getDefaultInstance returns error
	getDefaultInstance = func() (*Config, error) {
		return nil, fmt.Errorf("mock error")
	}

	// With should create a new config
	result3 := With(WithMode(Development))
	if result3 == nil {
		t.Error("With() returned nil instead of creating new config")
	} else if result3.Mode != Development {
		t.Errorf("With() with error: expected Mode=%s, got %s",
			Development, result3.Mode)
	}
}

// TestInitializeWithMultipleOptionsExtended tests Initialize with various option combinations
func TestInitializeWithMultipleOptionsExtended(t *testing.T) {
	// Save original state
	origDefaultInstance := defaultInstance
	origInitErr := initErr
	origOnce := once

	// Reset for testing
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Restore after test
	defer func() {
		defaultInstance = origDefaultInstance
		initErr = origInitErr
		once = origOnce
	}()

	// Create temp env file
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Skipf("Failed to get current dir: %v", err)
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Skipf("Failed to change to temp dir: %v", err)
		return
	}

	if err := os.WriteFile(".env", []byte("TEST=value"), 0644); err != nil {
		t.Skipf("Failed to create .env: %v", err)
		return
	}

	// Test with no options
	err = Initialize()
	if err != nil {
		t.Errorf("Initialize() with no options failed: %v", err)
	}

	if defaultInstance == nil {
		t.Fatal("defaultInstance is nil after Initialize()")
	}

	// Check default values
	if defaultInstance.Mode != Production &&
		defaultInstance.Mode != Development {
		// In test environment we might get either Production or Development
		// depending on what files exist
		t.Errorf("Unexpected Mode=%s", defaultInstance.Mode)
	}

	if defaultInstance.Prefix != "" {
		t.Errorf("Expected Prefix='', got %s", defaultInstance.Prefix)
	}

	// Reset singleton
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Test with multiple options
	options := []ConfigOption{
		WithMode(Staging),
		WithPrefix("STAGE_"),
	}

	err = Initialize(options...)
	if err != nil {
		t.Fatalf("Initialize with options failed: %v", err)
	}

	// Verify options were applied
	if defaultInstance == nil {
		t.Fatal("defaultInstance is nil after Initialize with options")
	}

	if defaultInstance.Mode != Staging {
		t.Errorf("Expected Mode=%s, got %s", Staging, defaultInstance.Mode)
	}

	if defaultInstance.Prefix != "STAGE_" {
		t.Errorf("Expected Prefix=STAGE_, got %s", defaultInstance.Prefix)
	}
}

// TestKeyPackageLevelWithError tests the Key package function with error handling
func TestKeyPackageLevelWithError(t *testing.T) {
	// Save original function
	origGetDefaultInstance := getDefaultInstance
	defer func() { getDefaultInstance = origGetDefaultInstance }()

	// Mock getDefaultInstance to return error
	mockErr := fmt.Errorf("mock error")
	getDefaultInstance = func() (*Config, error) {
		return nil, mockErr
	}

	// Test Key function
	result := Key("ANY_KEY")
	if result.err != mockErr {
		t.Errorf("Expected error %v, got %v", mockErr, result.err)
	}

	// Test all result methods after error
	if result.String() != "" {
		t.Errorf("Expected empty string, got '%s'", result.String())
	}

	if val, err := result.Int(); val != 0 || err != mockErr {
		t.Errorf("Expected (0, error), got (%d, %v)", val, err)
	}

	if val := result.IntDefault(42); val != 42 {
		t.Errorf("Expected default 42, got %d", val)
	}

	if val := result.Bool(); val != false {
		t.Errorf("Expected false, got %v", val)
	}

	if val := result.BoolDefault(true); val != true {
		t.Errorf("Expected default true, got %v", val)
	}
}

// TestWithInitializeRace tests race conditions with Initialize and With
func TestWithInitializeRace(t *testing.T) {
	// Skip if -short flag is set
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Save original state
	origDefaultInstance := defaultInstance
	origInitErr := initErr
	origOnce := once

	// Reset for testing
	defaultInstance = nil
	initErr = nil
	once = sync.Once{}

	// Restore after test
	defer func() {
		defaultInstance = origDefaultInstance
		initErr = origInitErr
		once = origOnce
	}()

	// Create temp env file
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Skipf("Failed to get current dir: %v", err)
		return
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Skipf("Failed to change to temp dir: %v", err)
		return
	}

	if err := os.WriteFile(".env", []byte("TEST=value"), 0644); err != nil {
		t.Skipf("Failed to create .env: %v", err)
		return
	}

	// Run Initialize and With concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 10)
	resultChan := make(chan *Config, 10)

	// 5 goroutines for Initialize
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Initialize(WithMode(Staging))
			errChan <- err
		}()
	}

	// 5 goroutines for With
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultChan <- With(WithPrefix("TEST_"))
		}()
	}

	wg.Wait()
	close(errChan)
	close(resultChan)

	// Check Initialize results
	for err := range errChan {
		if err != nil {
			t.Errorf("Initialize returned error: %v", err)
		}
	}

	// Check With results
	for cfg := range resultChan {
		if cfg == nil {
			t.Error("With returned nil config")
		}
	}

	// Final singleton should be in a valid state
	if defaultInstance == nil {
		t.Error("defaultInstance is nil after concurrent operations")
	} else {
		// Either the Initialize or With mode should have won
		if defaultInstance.Mode != Staging {
			t.Errorf("Expected final Mode=%s, got %s", Staging, defaultInstance.Mode)
		}
	}
}

// TestOptionChaining tests creating option chains
func TestOptionChaining(t *testing.T) {
	// Create an option chain function
	chainOptions := func(options ...ConfigOption) ConfigOption {
		return func(c *Config) {
			for _, opt := range options {
				opt(c)
			}
		}
	}

	// Create a combined option
	combinedOption := chainOptions(
		WithMode(Staging),
		WithPrefix("COMBINED_"),
	)

	// Apply to a config
	config := &Config{}
	combinedOption(config)

	// Verify both options applied
	if config.Mode != Staging {
		t.Errorf("Expected Mode=%s, got %s", Staging, config.Mode)
	}

	if config.Prefix != "COMBINED_" {
		t.Errorf("Expected Prefix=COMBINED_, got %s", config.Prefix)
	}

	// Override with a regular option
	WithMode(Development)(config)
	if config.Mode != Development {
		t.Errorf("Expected Mode=%s, got %s", Development, config.Mode)
	}

	// Prefix should remain unchanged
	if config.Prefix != "COMBINED_" {
		t.Errorf("Expected Prefix=COMBINED_, got %s", config.Prefix)
	}
}
