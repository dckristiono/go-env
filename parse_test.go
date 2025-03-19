package env

import (
	"os"
	"testing"
	"time"
)

// TestConfig untuk test parsing
type TestConfig struct {
	AppName        string            `env:"TEST_APP_NAME" default:"DefaultApp"`
	Port           int               `env:"TEST_PORT" default:"8080"`
	Debug          bool              `env:"TEST_DEBUG" default:"false"`
	Timeout        time.Duration     `env:"TEST_TIMEOUT" default:"30s"`
	AllowedOrigins []string          `env:"TEST_ALLOWED_ORIGINS"`
	Features       map[string]string `env:"TEST_FEATURES"`
	NoTagField     string
	privateField   string
}

func TestParse(t *testing.T) {
	// Set up
	os.Setenv("TEST_APP_NAME", "TestApp")
	os.Setenv("TEST_PORT", "9000")
	os.Setenv("TEST_DEBUG", "true")
	os.Setenv("TEST_TIMEOUT", "45s")
	os.Setenv("TEST_ALLOWED_ORIGINS", "localhost,127.0.0.1")
	os.Setenv("TEST_FEATURES", "feature1:enabled,feature2:disabled")
	os.Setenv("NOTAG_FIELD", "should_not_be_set")
	defer func() {
		os.Unsetenv("TEST_APP_NAME")
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_DEBUG")
		os.Unsetenv("TEST_TIMEOUT")
		os.Unsetenv("TEST_ALLOWED_ORIGINS")
		os.Unsetenv("TEST_FEATURES")
		os.Unsetenv("NOTAG_FIELD")
	}()

	// Test
	var config TestConfig
	err := Parse(&config)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check values
	if config.AppName != "TestApp" {
		t.Errorf("Expected AppName 'TestApp', got '%s'", config.AppName)
	}

	if config.Port != 9000 {
		t.Errorf("Expected Port 9000, got %d", config.Port)
	}

	if !config.Debug {
		t.Errorf("Expected Debug true, got %v", config.Debug)
	}

	if config.Timeout != 45*time.Second {
		t.Errorf("Expected Timeout 45s, got %v", config.Timeout)
	}

	if len(config.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
	} else {
		if config.AllowedOrigins[0] != "localhost" || config.AllowedOrigins[1] != "127.0.0.1" {
			t.Errorf("Unexpected allowed origins: %v", config.AllowedOrigins)
		}
	}

	if len(config.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(config.Features))
	} else {
		if config.Features["feature1"] != "enabled" || config.Features["feature2"] != "disabled" {
			t.Errorf("Unexpected features: %v", config.Features)
		}
	}

	// Fields without tag should be empty
	if config.NoTagField != "" {
		t.Errorf("Expected NoTagField to be empty, got '%s'", config.NoTagField)
	}
}

func TestParseDefaults(t *testing.T) {
	// Make sure environment variables are unset
	os.Unsetenv("TEST_APP_NAME")
	os.Unsetenv("TEST_PORT")
	os.Unsetenv("TEST_DEBUG")
	os.Unsetenv("TEST_TIMEOUT")
	os.Unsetenv("TEST_ALLOWED_ORIGINS")
	os.Unsetenv("TEST_FEATURES")

	// Test
	var config TestConfig
	err := Parse(&config)

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check default values
	if config.AppName != "DefaultApp" {
		t.Errorf("Expected default AppName 'DefaultApp', got '%s'", config.AppName)
	}

	if config.Port != 8080 {
		t.Errorf("Expected default Port 8080, got %d", config.Port)
	}

	if config.Debug {
		t.Errorf("Expected default Debug false, got %v", config.Debug)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default Timeout 30s, got %v", config.Timeout)
	}

	// Arrays and maps should be empty
	if len(config.AllowedOrigins) != 0 {
		t.Errorf("Expected empty allowed origins, got %v", config.AllowedOrigins)
	}

	if len(config.Features) != 0 {
		t.Errorf("Expected empty features, got %v", config.Features)
	}
}

// NonStructConfig bukan struct
type NonStructType int

func TestParseNonStructType(t *testing.T) {
	// Test dengan non-struct
	var nonStruct NonStructType
	err := Parse(&nonStruct)

	// Should return error
	if err == nil {
		t.Error("Expected error for non-struct type, got nil")
	}

	// Test dengan non-pointer
	var config TestConfig
	err = Parse(config) // Passing by value, not pointer

	// Should return error
	if err == nil {
		t.Error("Expected error for non-pointer type, got nil")
	}
}

// InvalidTypeConfig untuk test tipe yang tidak didukung
type InvalidTypeConfig struct {
	Complex complex128 `env:"TEST_COMPLEX"`
}

func TestParseInvalidType(t *testing.T) {
	os.Setenv("TEST_COMPLEX", "1+2i")
	defer os.Unsetenv("TEST_COMPLEX")

	var config InvalidTypeConfig
	err := Parse(&config)

	// Should return error
	if err == nil {
		t.Error("Expected error for unsupported type, got nil")
	}
}
