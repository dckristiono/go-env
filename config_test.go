package env

import (
	"os"
	"testing"
	"time"
)

// ConfigOption type declarations for testing
func WithTestMode(mode string) ConfigOption {
	return func(c *Config) {
		c.Mode = mode
	}
}

func WithTestPrefix(prefix string) ConfigOption {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// Create a mock implementation for testing file existence
type mockFileSystemHelper struct {
	fileExistsFunc func(string) bool
}

func (m *mockFileSystemHelper) fileExists(filename string) bool {
	return m.fileExistsFunc(filename)
}

// TestConfigDetermineDefaultMode tests mode selection logic
func TestConfigDetermineDefaultMode(t *testing.T) {
	// Save original environment
	oldEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", oldEnv)

	// Test with APP_ENV set
	os.Setenv("APP_ENV", "staging")
	if mode := determineDefaultMode(); mode != "staging" {
		t.Errorf("Expected mode 'staging', got '%s'", mode)
	}
	os.Unsetenv("APP_ENV")

	// We can't mock fileExists directly, so we'll test the behavior
	// based on actual file creation

	// Create temporary files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Change to temporary directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test with only .env file
	if err := os.WriteFile(".env", []byte("TEST=VALUE"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}
	if mode := determineDefaultMode(); mode != Production {
		t.Errorf("With only .env file, expected mode %s, got %s", Production, mode)
	}

	// Test with .env and .env.staging
	if err := os.WriteFile(".env.staging", []byte("TEST=VALUE"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env.staging file: %v", err)
	}
	if mode := determineDefaultMode(); mode != Staging {
		t.Errorf("With .env and .env.staging files, expected mode %s, got %s", Staging, mode)
	}

	// Test with all three files
	if err := os.WriteFile(".env.development", []byte("TEST=VALUE"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env.development file: %v", err)
	}
	if mode := determineDefaultMode(); mode != Development {
		t.Errorf("With all env files, expected mode %s, got %s", Development, mode)
	}

	// Clean up by changing back to original directory
	// (temp files will be cleaned up by t.TempDir automatically)
}

// TestConfigNew tests creating a new Config
func TestConfigNew(t *testing.T) {
	// Save and restore environment
	oldEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", oldEnv)
	os.Setenv("APP_ENV", "production")

	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a minimal .env file
	if err := os.WriteFile(".env", []byte("TEST_KEY=test_value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Test default config
	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if cfg.Mode != "production" {
		t.Errorf("Expected Mode = %v, got %v", "production", cfg.Mode)
	}
	if cfg.Prefix != "" {
		t.Errorf("Expected Prefix = '', got %v", cfg.Prefix)
	}

	// Test with options
	cfg, err = New(WithTestMode(Staging), WithTestPrefix("TEST_"))
	if err != nil {
		t.Fatalf("New() with options error = %v", err)
	}
	if cfg.Mode != Staging {
		t.Errorf("Expected Mode = %v, got %v", Staging, cfg.Mode)
	}
	if cfg.Prefix != "TEST_" {
		t.Errorf("Expected Prefix = 'TEST_', got %v", cfg.Prefix)
	}
}

// TestConfigLoad tests loading environment files
func TestConfigLoad(t *testing.T) {
	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test loading in production mode with file present
	if err := os.WriteFile(".env", []byte("TEST_KEY=production_value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	cfg := &Config{Mode: Production}
	err = cfg.Load()
	if err != nil {
		t.Errorf("Load() in Production mode with file present error: %v", err)
	}

	// Test loading in production mode with file absent
	if err := os.Remove(".env"); err != nil {
		t.Fatalf("Failed to remove temp .env file: %v", err)
	}

	err = cfg.Load()
	if err == nil {
		t.Errorf("Load() in Production mode with file absent should return error")
	}

	// Test loading in staging mode with file present
	if err := os.WriteFile(".env.staging", []byte("TEST_KEY=staging_value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env.staging file: %v", err)
	}

	cfg = &Config{Mode: Staging}
	err = cfg.Load()
	if err != nil {
		t.Errorf("Load() in Staging mode with file present error: %v", err)
	}

	// Test loading in staging mode with file absent
	if err := os.Remove(".env.staging"); err != nil {
		t.Fatalf("Failed to remove temp .env.staging file: %v", err)
	}

	err = cfg.Load()
	if err != nil {
		t.Errorf("Load() in Staging mode with file absent should only warn, not error: %v", err)
	}

	// Test loading in development mode with file present
	if err := os.WriteFile(".env.development", []byte("TEST_KEY=development_value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env.development file: %v", err)
	}

	cfg = &Config{Mode: Development}
	err = cfg.Load()
	if err != nil {
		t.Errorf("Load() in Development mode with file present error: %v", err)
	}

	// Test loading in development mode with file absent
	if err := os.Remove(".env.development"); err != nil {
		t.Fatalf("Failed to remove temp .env.development file: %v", err)
	}

	err = cfg.Load()
	if err != nil {
		t.Errorf("Load() in Development mode with file absent should only warn, not error: %v", err)
	}

	// Test invalid mode
	cfg = &Config{Mode: "invalid"}
	err = cfg.Load()
	if err == nil {
		t.Errorf("Load() with invalid mode should return error")
	}
}

// TestConfigGetString tests the Get method
func TestConfigGetString(t *testing.T) {
	// Save original env vars
	oldEnv := os.Getenv("TEST_KEY")
	defer os.Setenv("TEST_KEY", oldEnv)

	// Set up test environment
	os.Setenv("TEST_KEY", "value")

	cfg := &Config{Prefix: "TEST_"}

	// Test basic Get
	if val := cfg.Get("KEY"); val != "value" {
		t.Errorf("Get(KEY) = %v, want %v", val, "value")
	}

	// Test nonexistent key
	if val := cfg.Get("NONEXISTENT"); val != "" {
		t.Errorf("Get(NONEXISTENT) = %v, want ''", val)
	}

	// Test default value
	if val := cfg.Get("NONEXISTENT", "default"); val != "default" {
		t.Errorf("Get(NONEXISTENT, default) = %v, want 'default'", val)
	}
}

// TestConfigGetInt tests the GetInt method
func TestConfigGetInt(t *testing.T) {
	// Save original env vars
	oldIntEnv := os.Getenv("TEST_INT")
	oldInvalidEnv := os.Getenv("TEST_INVALID")
	defer func() {
		os.Setenv("TEST_INT", oldIntEnv)
		os.Setenv("TEST_INVALID", oldInvalidEnv)
	}()

	os.Setenv("TEST_INT", "123")
	os.Setenv("TEST_INVALID", "abc")

	cfg := &Config{Prefix: "TEST_"}

	// Test valid int
	val, err := cfg.GetInt("INT")
	if err != nil {
		t.Errorf("GetInt(INT) unexpected error: %v", err)
	}
	if val != 123 {
		t.Errorf("GetInt(INT) = %v, want 123", val)
	}

	// Test invalid int
	if _, err := cfg.GetInt("INVALID"); err == nil {
		t.Errorf("GetInt(INVALID) expected error")
	}

	// Test missing with default
	val, err = cfg.GetInt("NONEXISTENT", 456)
	if err != nil {
		t.Errorf("GetInt(NONEXISTENT, 456) unexpected error: %v", err)
	}
	if val != 456 {
		t.Errorf("GetInt(NONEXISTENT, 456) = %v, want 456", val)
	}

	// Test missing without default
	if _, err := cfg.GetInt("NONEXISTENT"); err == nil {
		t.Errorf("GetInt(NONEXISTENT) expected error")
	}
}

// TestConfigGetInt64 tests the GetInt64 method
func TestConfigGetInt64(t *testing.T) {
	// Save original env vars
	oldInt64Env := os.Getenv("TEST_INT64")
	oldInvalidEnv := os.Getenv("TEST_INVALID")
	defer func() {
		os.Setenv("TEST_INT64", oldInt64Env)
		os.Setenv("TEST_INVALID", oldInvalidEnv)
	}()

	os.Setenv("TEST_INT64", "9223372036854775807") // Max int64
	os.Setenv("TEST_INVALID", "not a number")

	cfg := &Config{Prefix: "TEST_"}

	// Test valid int64
	val, err := cfg.GetInt64("INT64")
	if err != nil {
		t.Errorf("GetInt64(INT64) unexpected error: %v", err)
	}
	if val != 9223372036854775807 {
		t.Errorf("GetInt64(INT64) = %v, want 9223372036854775807", val)
	}

	// Test invalid int64
	if _, err := cfg.GetInt64("INVALID"); err == nil {
		t.Errorf("GetInt64(INVALID) expected error")
	}

	// Test with default
	val, err = cfg.GetInt64("NONEXISTENT", 123)
	if err != nil {
		t.Errorf("GetInt64(NONEXISTENT, 123) unexpected error: %v", err)
	}
	if val != 123 {
		t.Errorf("GetInt64(NONEXISTENT, 123) = %v, want 123", val)
	}
}

// TestConfigGetFloat64 tests the GetFloat64 method
func TestConfigGetFloat64(t *testing.T) {
	// Save original env vars
	oldFloatEnv := os.Getenv("TEST_FLOAT")
	oldInvalidEnv := os.Getenv("TEST_INVALID")
	defer func() {
		os.Setenv("TEST_FLOAT", oldFloatEnv)
		os.Setenv("TEST_INVALID", oldInvalidEnv)
	}()

	os.Setenv("TEST_FLOAT", "123.456")
	os.Setenv("TEST_INVALID", "not a number")

	cfg := &Config{Prefix: "TEST_"}

	// Test valid float
	val, err := cfg.GetFloat64("FLOAT")
	if err != nil {
		t.Errorf("GetFloat64(FLOAT) unexpected error: %v", err)
	}
	if val != 123.456 {
		t.Errorf("GetFloat64(FLOAT) = %v, want 123.456", val)
	}

	// Test invalid float
	if _, err := cfg.GetFloat64("INVALID"); err == nil {
		t.Errorf("GetFloat64(INVALID) expected error")
	}

	// Test with default
	val, err = cfg.GetFloat64("NONEXISTENT", 456.789)
	if err != nil {
		t.Errorf("GetFloat64(NONEXISTENT, 456.789) unexpected error: %v", err)
	}
	if val != 456.789 {
		t.Errorf("GetFloat64(NONEXISTENT, 456.789) = %v, want 456.789", val)
	}
}

// TestConfigGetBool tests the GetBool method
func TestConfigGetBool(t *testing.T) {
	// Save original env vars
	oldEnvs := map[string]string{
		"TEST_TRUE":  os.Getenv("TEST_TRUE"),
		"TEST_YES":   os.Getenv("TEST_YES"),
		"TEST_Y":     os.Getenv("TEST_Y"),
		"TEST_1":     os.Getenv("TEST_1"),
		"TEST_FALSE": os.Getenv("TEST_FALSE"),
		"TEST_NO":    os.Getenv("TEST_NO"),
	}
	defer func() {
		for k, v := range oldEnvs {
			os.Setenv(k, v)
		}
	}()

	// Set various boolean representations
	os.Setenv("TEST_TRUE", "true")
	os.Setenv("TEST_YES", "yes")
	os.Setenv("TEST_Y", "y")
	os.Setenv("TEST_1", "1")
	os.Setenv("TEST_FALSE", "false")
	os.Setenv("TEST_NO", "no")

	cfg := &Config{Prefix: "TEST_"}

	// Test true values
	trueKeys := []string{"TRUE", "YES", "Y", "1"}
	for _, key := range trueKeys {
		if !cfg.GetBool(key) {
			t.Errorf("GetBool(%s) = false, want true", key)
		}
	}

	// Test false values
	if cfg.GetBool("FALSE") {
		t.Errorf("GetBool(FALSE) = true, want false")
	}
	if cfg.GetBool("NO") {
		t.Errorf("GetBool(NO) = true, want false")
	}

	// Test with default
	if !cfg.GetBool("NONEXISTENT", true) {
		t.Errorf("GetBool(NONEXISTENT, true) = false, want true")
	}
	if cfg.GetBool("NONEXISTENT", false) {
		t.Errorf("GetBool(NONEXISTENT, false) = true, want false")
	}

	// Test missing with no default
	if cfg.GetBool("NONEXISTENT") {
		t.Errorf("GetBool(NONEXISTENT) = true, want false")
	}
}

// TestConfigGetDuration tests the GetDuration method
func TestConfigGetDuration(t *testing.T) {
	// Save original env vars
	oldDurationEnv := os.Getenv("TEST_DURATION")
	oldInvalidEnv := os.Getenv("TEST_INVALID")
	defer func() {
		os.Setenv("TEST_DURATION", oldDurationEnv)
		os.Setenv("TEST_INVALID", oldInvalidEnv)
	}()

	os.Setenv("TEST_DURATION", "5s")
	os.Setenv("TEST_INVALID", "not a duration")

	cfg := &Config{Prefix: "TEST_"}

	// Test valid duration
	val, err := cfg.GetDuration("DURATION")
	if err != nil {
		t.Errorf("GetDuration(DURATION) unexpected error: %v", err)
	}
	if val != 5*time.Second {
		t.Errorf("GetDuration(DURATION) = %v, want 5s", val)
	}

	// Test invalid duration
	if _, err := cfg.GetDuration("INVALID"); err == nil {
		t.Errorf("GetDuration(INVALID) expected error")
	}

	// Test with default
	val, err = cfg.GetDuration("NONEXISTENT", 10*time.Minute)
	if err != nil {
		t.Errorf("GetDuration(NONEXISTENT, 10m) unexpected error: %v", err)
	}
	if val != 10*time.Minute {
		t.Errorf("GetDuration(NONEXISTENT, 10m) = %v, want 10m", val)
	}
}

// TestConfigGetSlice tests the GetSlice method
func TestConfigGetSlice(t *testing.T) {
	// Save original env vars
	oldSliceEnv := os.Getenv("TEST_SLICE")
	oldSliceCustomEnv := os.Getenv("TEST_SLICECUSTOM")
	defer func() {
		os.Setenv("TEST_SLICE", oldSliceEnv)
		os.Setenv("TEST_SLICECUSTOM", oldSliceCustomEnv)
	}()

	os.Setenv("TEST_SLICE", "a,b,c")
	os.Setenv("TEST_SLICECUSTOM", "a|b|c")

	cfg := &Config{Prefix: "TEST_"}

	// Test with default delimiter
	slice := cfg.GetSlice("SLICE", "")
	expected := []string{"a", "b", "c"}
	if len(slice) != len(expected) {
		t.Errorf("GetSlice(SLICE, '') length = %d, want %d", len(slice), len(expected))
	}
	for i, v := range expected {
		if slice[i] != v {
			t.Errorf("GetSlice(SLICE, '')[%d] = %s, want %s", i, slice[i], v)
		}
	}

	// Test with custom delimiter
	slice = cfg.GetSlice("SLICECUSTOM", "|")
	if len(slice) != len(expected) {
		t.Errorf("GetSlice(SLICECUSTOM, '|') length = %d, want %d", len(slice), len(expected))
	}
	for i, v := range expected {
		if slice[i] != v {
			t.Errorf("GetSlice(SLICECUSTOM, '|')[%d] = %s, want %s", i, slice[i], v)
		}
	}

	// Test with default value
	defaultSlice := []string{"default1", "default2"}
	slice = cfg.GetSlice("NONEXISTENT", "", defaultSlice)
	if len(slice) != len(defaultSlice) {
		t.Errorf("GetSlice(NONEXISTENT, '', defaultSlice) length = %d, want %d", len(slice), len(defaultSlice))
	}
	for i, v := range defaultSlice {
		if slice[i] != v {
			t.Errorf("GetSlice(NONEXISTENT, '', defaultSlice)[%d] = %s, want %s", i, slice[i], v)
		}
	}

	// Test empty slice
	slice = cfg.GetSlice("NONEXISTENT", "")
	if len(slice) != 0 {
		t.Errorf("GetSlice(NONEXISTENT, '') length = %d, want 0", len(slice))
	}
}

// TestConfigGetMap tests the GetMap method
func TestConfigGetMap(t *testing.T) {
	// Save original env vars
	oldMapEnv := os.Getenv("TEST_MAP")
	oldBadMapEnv := os.Getenv("TEST_BADMAP")
	defer func() {
		os.Setenv("TEST_MAP", oldMapEnv)
		os.Setenv("TEST_BADMAP", oldBadMapEnv)
	}()

	os.Setenv("TEST_MAP", "key1:value1,key2:value2")
	os.Setenv("TEST_BADMAP", "invalid")

	cfg := &Config{Prefix: "TEST_"}

	// Test valid map
	m := cfg.GetMap("MAP")
	if len(m) != 2 {
		t.Errorf("GetMap(MAP) size = %d, want 2", len(m))
	}
	if m["key1"] != "value1" {
		t.Errorf("GetMap(MAP)[key1] = %s, want 'value1'", m["key1"])
	}
	if m["key2"] != "value2" {
		t.Errorf("GetMap(MAP)[key2] = %s, want 'value2'", m["key2"])
	}

	// Test invalid format
	m = cfg.GetMap("BADMAP")
	if len(m) != 0 {
		t.Errorf("GetMap(BADMAP) size = %d, want 0", len(m))
	}

	// Test with default
	defaultMap := map[string]string{"default": "value"}
	m = cfg.GetMap("NONEXISTENT", defaultMap)
	if len(m) != 1 {
		t.Errorf("GetMap(NONEXISTENT, defaultMap) size = %d, want 1", len(m))
	}
	if m["default"] != "value" {
		t.Errorf("GetMap(NONEXISTENT, defaultMap)[default] = %s, want 'value'", m["default"])
	}
}

// TestConfigModeChecking tests mode checking methods
func TestConfigModeChecking(t *testing.T) {
	// Test Production mode
	cfg := &Config{Mode: Production}
	if !cfg.IsProduction() {
		t.Errorf("IsProduction() = false, want true for Mode = %s", Production)
	}
	if cfg.IsStaging() {
		t.Errorf("IsStaging() = true, want false for Mode = %s", Production)
	}
	if cfg.IsDevelopment() {
		t.Errorf("IsDevelopment() = true, want false for Mode = %s", Production)
	}
	if mode := cfg.GetMode(); mode != Production {
		t.Errorf("GetMode() = %s, want %s", mode, Production)
	}

	// Test Staging mode
	cfg = &Config{Mode: Staging}
	if cfg.IsProduction() {
		t.Errorf("IsProduction() = true, want false for Mode = %s", Staging)
	}
	if !cfg.IsStaging() {
		t.Errorf("IsStaging() = false, want true for Mode = %s", Staging)
	}
	if cfg.IsDevelopment() {
		t.Errorf("IsDevelopment() = true, want false for Mode = %s", Staging)
	}

	// Test Development mode
	cfg = &Config{Mode: Development}
	if cfg.IsProduction() {
		t.Errorf("IsProduction() = true, want false for Mode = %s", Development)
	}
	if cfg.IsStaging() {
		t.Errorf("IsStaging() = true, want false for Mode = %s", Development)
	}
	if !cfg.IsDevelopment() {
		t.Errorf("IsDevelopment() = false, want true for Mode = %s", Development)
	}
}

// TestConfigFrom tests the From method
func TestConfigFrom(t *testing.T) {
	// Create initial config
	cfg := &Config{Mode: Production, Prefix: "INITIAL_"}

	// Get new config with options
	newCfg := cfg.From(WithTestMode(Staging), WithTestPrefix("NEW_"))

	// Check new config
	if newCfg.Mode != Staging {
		t.Errorf("From() Mode = %s, want %s", newCfg.Mode, Staging)
	}
	if newCfg.Prefix != "NEW_" {
		t.Errorf("From() Prefix = %s, want %s", newCfg.Prefix, "NEW_")
	}

	// Original should be unchanged
	if cfg.Mode != Production {
		t.Errorf("Original Mode changed to %s, should remain %s", cfg.Mode, Production)
	}
	if cfg.Prefix != "INITIAL_" {
		t.Errorf("Original Prefix changed to %s, should remain %s", cfg.Prefix, "INITIAL_")
	}
}

// TestConfigInitialize tests the package Initialize function
func TestConfigInitialize(t *testing.T) {
	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a .env file
	if err := os.WriteFile(".env", []byte("TEST_KEY=value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Reset singleton
	defaultInstance = nil

	// Initialize
	err = Initialize(WithTestMode(Staging), WithTestPrefix("TEST_"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check singleton was set correctly
	if defaultInstance == nil {
		t.Fatal("defaultInstance is nil after Initialize()")
	}
	if defaultInstance.Mode != Staging {
		t.Errorf("defaultInstance.Mode = %s, want %s", defaultInstance.Mode, Staging)
	}
	if defaultInstance.Prefix != "TEST_" {
		t.Errorf("defaultInstance.Prefix = %s, want %s", defaultInstance.Prefix, "TEST_")
	}
}

// TestConfigKey tests the Key method using the result type
func TestConfigKey(t *testing.T) {
	// Save original env vars
	oldStrEnv := os.Getenv("TEST_STR")
	oldIntEnv := os.Getenv("TEST_INT")
	oldBoolEnv := os.Getenv("TEST_BOOL")
	defer func() {
		os.Setenv("TEST_STR", oldStrEnv)
		os.Setenv("TEST_INT", oldIntEnv)
		os.Setenv("TEST_BOOL", oldBoolEnv)
	}()

	os.Setenv("TEST_STR", "value")
	os.Setenv("TEST_INT", "123")
	os.Setenv("TEST_BOOL", "true")

	cfg := &Config{Prefix: "TEST_"}

	// Test string value
	if val := cfg.Key("STR").String(); val != "value" {
		t.Errorf("Key(STR).String() = %s, want 'value'", val)
	}

	// Test with nonexistent key, should return empty string
	if val := cfg.Key("NONEXISTENT").String(); val != "" {
		t.Errorf("Key(NONEXISTENT).String() = %s, want ''", val)
	}

	// Test integer
	i, err := cfg.Key("INT").Int()
	if err != nil {
		t.Errorf("Key(INT).Int() unexpected error: %v", err)
	}
	if i != 123 {
		t.Errorf("Key(INT).Int() = %d, want 123", i)
	}

	// Test boolean
	if !cfg.Key("BOOL").Bool() {
		t.Errorf("Key(BOOL).Bool() = false, want true")
	}

	// Test with nonexistent key for Int, should return error
	_, err = cfg.Key("NONEXISTENT").Int()
	if err == nil {
		t.Errorf("Key(NONEXISTENT).Int() expected error, got nil")
	}
}

// TestConfigWith tests the With function
func TestConfigWith(t *testing.T) {
	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a .env file
	if err := os.WriteFile(".env", []byte("TEST_KEY=value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Reset singleton
	defaultInstance = nil

	// Initialize the default instance
	err = Initialize(WithTestMode(Production))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test With
	cfg := With(WithTestMode(Staging), WithTestPrefix("TEST_"))
	if cfg.Mode != Staging {
		t.Errorf("With() Mode = %s, want %s", cfg.Mode, Staging)
	}
	if cfg.Prefix != "TEST_" {
		t.Errorf("With() Prefix = %s, want %s", cfg.Prefix, "TEST_")
	}
}

// TestConfigPackageGetters tests all the package-level getter functions
func TestConfigPackageGetters(t *testing.T) {
	// Save original env vars
	oldStrEnv := os.Getenv("TEST_STRING")
	oldIntEnv := os.Getenv("TEST_INT")
	oldInt64Env := os.Getenv("TEST_INT64")
	oldFloatEnv := os.Getenv("TEST_FLOAT")
	oldBoolEnv := os.Getenv("TEST_BOOL")
	oldDurationEnv := os.Getenv("TEST_DURATION")
	oldSliceEnv := os.Getenv("TEST_SLICE")
	oldMapEnv := os.Getenv("TEST_MAP")
	defer func() {
		os.Setenv("TEST_STRING", oldStrEnv)
		os.Setenv("TEST_INT", oldIntEnv)
		os.Setenv("TEST_INT64", oldInt64Env)
		os.Setenv("TEST_FLOAT", oldFloatEnv)
		os.Setenv("TEST_BOOL", oldBoolEnv)
		os.Setenv("TEST_DURATION", oldDurationEnv)
		os.Setenv("TEST_SLICE", oldSliceEnv)
		os.Setenv("TEST_MAP", oldMapEnv)
	}()

	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a .env file
	if err := os.WriteFile(".env", []byte("TEST_KEY=value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Reset and initialize the default instance
	defaultInstance = nil
	err = Initialize(WithTestPrefix("TEST_"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Set variables
	os.Setenv("TEST_STRING", "value")
	os.Setenv("TEST_INT", "123")
	os.Setenv("TEST_INT64", "9223372036854775807") // Max int64
	os.Setenv("TEST_FLOAT", "123.456")
	os.Setenv("TEST_BOOL", "true")
	os.Setenv("TEST_DURATION", "5s")
	os.Setenv("TEST_SLICE", "a,b,c")
	os.Setenv("TEST_MAP", "k1:v1,k2:v2")

	// Test Get/String
	if val := Get("STRING"); val != "value" {
		t.Errorf("Get(STRING) = %s, want 'value'", val)
	}
	if val := String("STRING"); val != "value" {
		t.Errorf("String(STRING) = %s, want 'value'", val)
	}

	// Test non-existent key with default
	if val := Get("NONEXISTENT", "default"); val != "default" {
		t.Errorf("Get(NONEXISTENT, default) = %s, want 'default'", val)
	}

	// Test Int
	val, err := GetInt("INT")
	if err != nil {
		t.Errorf("GetInt(INT) unexpected error: %v", err)
	}
	if val != 123 {
		t.Errorf("GetInt(INT) = %d, want 123", val)
	}
	if val := Int("INT"); val != 123 {
		t.Errorf("Int(INT) = %d, want 123", val)
	}

	// Test Int with default
	val, err = GetInt("NONEXISTENT", 456)
	if err != nil {
		t.Errorf("GetInt(NONEXISTENT, 456) unexpected error: %v", err)
	}
	if val != 456 {
		t.Errorf("GetInt(NONEXISTENT, 456) = %d, want 456", val)
	}
	if val := Int("NONEXISTENT", 456); val != 456 {
		t.Errorf("Int(NONEXISTENT, 456) = %d, want 456", val)
	}

	// Test Int64
	int64Val, err := GetInt64("INT64")
	if err != nil {
		t.Errorf("GetInt64(INT64) unexpected error: %v", err)
	}
	if int64Val != 9223372036854775807 {
		t.Errorf("GetInt64(INT64) = %d, want 9223372036854775807", int64Val)
	}

	// Test Int64 with default
	int64Val, err = GetInt64("NONEXISTENT", 456)
	if err != nil {
		t.Errorf("GetInt64(NONEXISTENT, 456) unexpected error: %v", err)
	}
	if int64Val != 456 {
		t.Errorf("GetInt64(NONEXISTENT, 456) = %d, want 456", int64Val)
	}

	// Test Float64
	f, err := GetFloat64("FLOAT")
	if err != nil {
		t.Errorf("GetFloat64(FLOAT) unexpected error: %v", err)
	}
	if f != 123.456 {
		t.Errorf("GetFloat64(FLOAT) = %f, want 123.456", f)
	}
	if f := Float64("FLOAT"); f != 123.456 {
		t.Errorf("Float64(FLOAT) = %f, want 123.456", f)
	}

	// Test Float64 with default
	f, err = GetFloat64("NONEXISTENT", 456.789)
	if err != nil {
		t.Errorf("GetFloat64(NONEXISTENT, 456.789) unexpected error: %v", err)
	}
	if f != 456.789 {
		t.Errorf("GetFloat64(NONEXISTENT, 456.789) = %f, want 456.789", f)
	}
	if f := Float64("NONEXISTENT", 456.789); f != 456.789 {
		t.Errorf("Float64(NONEXISTENT, 456.789) = %f, want 456.789", f)
	}

	// Test Bool
	if !GetBool("BOOL") {
		t.Errorf("GetBool(BOOL) = false, want true")
	}
	if !Bool("BOOL") {
		t.Errorf("Bool(BOOL) = false, want true")
	}

	// Test Bool with default
	if GetBool("NONEXISTENT", false) {
		t.Errorf("GetBool(NONEXISTENT, false) = true, want false")
	}
	if !GetBool("NONEXISTENT", true) {
		t.Errorf("GetBool(NONEXISTENT, true) = false, want true")
	}

	// Test Duration
	d, err := GetDuration("DURATION")
	if err != nil {
		t.Errorf("GetDuration(DURATION) unexpected error: %v", err)
	}
	if d != 5*time.Second {
		t.Errorf("GetDuration(DURATION) = %v, want 5s", d)
	}
	if d := Duration("DURATION"); d != 5*time.Second {
		t.Errorf("Duration(DURATION) = %v, want 5s", d)
	}

	// Test Duration with default
	d, err = GetDuration("NONEXISTENT", 10*time.Minute)
	if err != nil {
		t.Errorf("GetDuration(NONEXISTENT, 10m) unexpected error: %v", err)
	}
	if d != 10*time.Minute {
		t.Errorf("GetDuration(NONEXISTENT, 10m) = %v, want 10m", d)
	}
	if d := Duration("NONEXISTENT", 10*time.Minute); d != 10*time.Minute {
		t.Errorf("Duration(NONEXISTENT, 10m) = %v, want 10m", d)
	}

	// Test Slice
	s := GetSlice("SLICE", "")
	if len(s) != 3 || s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Errorf("GetSlice(SLICE, '') = %v, want [a b c]", s)
	}
	s = Slice("SLICE", "")
	if len(s) != 3 || s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Errorf("Slice(SLICE, '') = %v, want [a b c]", s)
	}

	// Test Slice with default
	defaultSlice := []string{"default1", "default2"}
	s = GetSlice("NONEXISTENT", "", defaultSlice)
	if len(s) != len(defaultSlice) || s[0] != defaultSlice[0] || s[1] != defaultSlice[1] {
		t.Errorf("GetSlice(NONEXISTENT, '', defaultSlice) = %v, want %v", s, defaultSlice)
	}

	// Test Map
	m := GetMap("MAP")
	if len(m) != 2 || m["k1"] != "v1" || m["k2"] != "v2" {
		t.Errorf("GetMap(MAP) = %v, want map[k1:v1 k2:v2]", m)
	}
	m = Map("MAP")
	if len(m) != 2 || m["k1"] != "v1" || m["k2"] != "v2" {
		t.Errorf("Map(MAP) = %v, want map[k1:v1 k2:v2]", m)
	}

	// Test Map with default
	defaultMap := map[string]string{"default": "value"}
	m = GetMap("NONEXISTENT", defaultMap)
	if len(m) != 1 || m["default"] != "value" {
		t.Errorf("GetMap(NONEXISTENT, defaultMap) = %v, want %v", m, defaultMap)
	}
}

// TestConfigModeGetters tests the mode-related package functions
func TestConfigModeGetters(t *testing.T) {
	// Create temp directory and files for testing
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a .env file
	if err := os.WriteFile(".env", []byte("TEST_KEY=value"), 0644); err != nil {
		t.Fatalf("Failed to create temp .env file: %v", err)
	}

	// Reset singleton
	defaultInstance = nil

	// Test with Production mode
	err = Initialize(WithTestMode(Production))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if !IsProduction() {
		t.Errorf("IsProduction() = false, want true")
	}
	if IsStaging() {
		t.Errorf("IsStaging() = true, want false")
	}
	if IsDevelopment() {
		t.Errorf("IsDevelopment() = true, want false")
	}
	if mode := GetMode(); mode != Production {
		t.Errorf("GetMode() = %s, want %s", mode, Production)
	}

	// Test with Staging mode
	defaultInstance = nil
	err = Initialize(WithTestMode(Staging))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if IsProduction() {
		t.Errorf("IsProduction() = true, want false")
	}
	if !IsStaging() {
		t.Errorf("IsStaging() = false, want true")
	}
	if IsDevelopment() {
		t.Errorf("IsDevelopment() = true, want false")
	}

	// Test with Development mode
	defaultInstance = nil
	err = Initialize(WithTestMode(Development))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if IsProduction() {
		t.Errorf("IsProduction() = true, want false")
	}
	if IsStaging() {
		t.Errorf("IsStaging() = true, want false")
	}
	if !IsDevelopment() {
		t.Errorf("IsDevelopment() = false, want true")
	}
}
