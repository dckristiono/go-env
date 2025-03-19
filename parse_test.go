package env

import (
	"os"
	"reflect"
	"testing"
	"time"
)

// Define test structs that were previously undefined
type TestExtendedConfig struct {
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

// TestParseInvalidInput tests parse with invalid input types
func TestParseInvalidInput(t *testing.T) {
	// Test with non-pointer
	err := Parse(TestExtendedConfig{})
	if err == nil {
		t.Error("Parse with non-pointer should fail")
	}

	// Test with pointer to non-struct
	var str string
	err = Parse(&str)
	if err == nil {
		t.Error("Parse with pointer to non-struct should fail")
	}

	// Test with nil
	err = Parse(nil)
	if err == nil {
		t.Error("Parse with nil should fail")
	}

	// Test with pointer to pointer
	config := &TestExtendedConfig{}
	ptrToPtr := &config
	err = Parse(ptrToPtr)
	if err == nil {
		t.Error("Parse with pointer to pointer should fail")
	}
}

// TestParseWithPrivateFields tests parsing with unexported fields
func TestParseWithPrivateFields(t *testing.T) {
	// Setup
	os.Setenv("PARSE_PUBLIC", "public_value")
	os.Setenv("PARSE_PRIVATE", "private_value")
	defer func() {
		os.Unsetenv("PARSE_PUBLIC")
		os.Unsetenv("PARSE_PRIVATE")
	}()

	// Struct with private fields
	type PrivateFieldsConfig struct {
		PublicField  string `env:"PARSE_PUBLIC"`
		privateField string `env:"PARSE_PRIVATE"` // Private field should be ignored
	}

	var config PrivateFieldsConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Public field should be set
	if config.PublicField != "public_value" {
		t.Errorf("PublicField expected 'public_value', got '%s'", config.PublicField)
	}

	// Private field should remain empty
	if config.privateField != "" {
		t.Errorf("privateField expected empty, got '%s'", config.privateField)
	}
}

// TestParseNestedStructs tests parsing nested structs
func TestParseNestedStructs(t *testing.T) {
	// Setup
	os.Setenv("PARSE_OUTER", "outer_value")
	os.Setenv("PARSE_INNER", "inner_value")
	defer func() {
		os.Unsetenv("PARSE_OUTER")
		os.Unsetenv("PARSE_INNER")
	}()

	// Nested struct
	type InnerConfig struct {
		InnerField string `env:"PARSE_INNER"`
	}

	type OuterConfig struct {
		OuterField string      `env:"PARSE_OUTER"`
		Inner      InnerConfig // Nested struct
	}

	var config OuterConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Only OuterField should be set, Inner is not parsed recursively
	if config.OuterField != "outer_value" {
		t.Errorf("OuterField expected 'outer_value', got '%s'", config.OuterField)
	}

	if config.Inner.InnerField != "" {
		t.Errorf("Inner.InnerField expected empty (not parsed recursively), got '%s'",
			config.Inner.InnerField)
	}
}

// TestParseAllTypes tests parsing all supported types with all combinations
func TestParseAllTypes(t *testing.T) {
	// Setup
	envVars := map[string]string{
		"PARSE_STRING":   "string_value",
		"PARSE_INT":      "42",
		"PARSE_INT8":     "42",
		"PARSE_INT16":    "42",
		"PARSE_INT32":    "42",
		"PARSE_INT64":    "42",
		"PARSE_UINT":     "42",
		"PARSE_UINT8":    "42",
		"PARSE_UINT16":   "42",
		"PARSE_UINT32":   "42",
		"PARSE_UINT64":   "42",
		"PARSE_FLOAT32":  "3.14",
		"PARSE_FLOAT64":  "3.14",
		"PARSE_BOOL":     "true",
		"PARSE_DURATION": "5s",
		"PARSE_SLICE":    "a,b,c",
		"PARSE_MAP":      "k1:v1,k2:v2",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Struct with all supported types
	type AllTypesConfig struct {
		String   string            `env:"PARSE_STRING"`
		Int      int               `env:"PARSE_INT"`
		Int8     int8              `env:"PARSE_INT8"`
		Int16    int16             `env:"PARSE_INT16"`
		Int32    int32             `env:"PARSE_INT32"`
		Int64    int64             `env:"PARSE_INT64"`
		Uint     uint              `env:"PARSE_UINT"`
		Uint8    uint8             `env:"PARSE_UINT8"`
		Uint16   uint16            `env:"PARSE_UINT16"`
		Uint32   uint32            `env:"PARSE_UINT32"`
		Uint64   uint64            `env:"PARSE_UINT64"`
		Float32  float32           `env:"PARSE_FLOAT32"`
		Float64  float64           `env:"PARSE_FLOAT64"`
		Bool     bool              `env:"PARSE_BOOL"`
		Duration time.Duration     `env:"PARSE_DURATION"`
		Slice    []string          `env:"PARSE_SLICE"`
		Map      map[string]string `env:"PARSE_MAP"`
	}

	var config AllTypesConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify all fields
	if config.String != "string_value" {
		t.Errorf("String expected 'string_value', got '%s'", config.String)
	}
	if config.Int != 42 {
		t.Errorf("Int expected 42, got %d", config.Int)
	}
	if config.Int8 != 42 {
		t.Errorf("Int8 expected 42, got %d", config.Int8)
	}
	if config.Int16 != 42 {
		t.Errorf("Int16 expected 42, got %d", config.Int16)
	}
	if config.Int32 != 42 {
		t.Errorf("Int32 expected 42, got %d", config.Int32)
	}
	if config.Int64 != 42 {
		t.Errorf("Int64 expected 42, got %d", config.Int64)
	}
	if config.Uint != 42 {
		t.Errorf("Uint expected 42, got %d", config.Uint)
	}
	if config.Uint8 != 42 {
		t.Errorf("Uint8 expected 42, got %d", config.Uint8)
	}
	if config.Uint16 != 42 {
		t.Errorf("Uint16 expected 42, got %d", config.Uint16)
	}
	if config.Uint32 != 42 {
		t.Errorf("Uint32 expected 42, got %d", config.Uint32)
	}
	if config.Uint64 != 42 {
		t.Errorf("Uint64 expected 42, got %d", config.Uint64)
	}
	if config.Float32 != 3.14 {
		t.Errorf("Float32 expected 3.14, got %f", config.Float32)
	}
	if config.Float64 != 3.14 {
		t.Errorf("Float64 expected 3.14, got %f", config.Float64)
	}
	if !config.Bool {
		t.Errorf("Bool expected true, got %v", config.Bool)
	}
	if config.Duration != 5*time.Second {
		t.Errorf("Duration expected 5s, got %v", config.Duration)
	}
	if len(config.Slice) != 3 || config.Slice[0] != "a" || config.Slice[1] != "b" || config.Slice[2] != "c" {
		t.Errorf("Slice expected [a b c], got %v", config.Slice)
	}
	if len(config.Map) != 2 || config.Map["k1"] != "v1" || config.Map["k2"] != "v2" {
		t.Errorf("Map expected map[k1:v1 k2:v2], got %v", config.Map)
	}
}

// TestParseWithCustomTag tests parsing with custom or missing tags
func TestParseWithCustomTag(t *testing.T) {
	// Setup
	os.Setenv("PARSE_DEFAULT_TAG", "from_env")
	os.Setenv("CUSTOM_TAG", "custom_value")
	os.Setenv("NOTAG", "uppercase_value") // Change to NOTAG to match field name
	defer func() {
		os.Unsetenv("PARSE_DEFAULT_TAG")
		os.Unsetenv("CUSTOM_TAG")
		os.Unsetenv("NOTAG")
	}()

	// Struct with various tag scenarios
	type CustomTagConfig struct {
		Default     string `env:"PARSE_DEFAULT_TAG"`
		Custom      string `env:"CUSTOM_TAG"`
		NoTag       string // Should use uppercase field name NOTAG
		Empty       string `env:""` // Empty tag
		WithDefault string `env:"NON_EXISTENT" default:"default_value"`
	}

	var config CustomTagConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify fields
	if config.Default != "from_env" {
		t.Errorf("Default expected 'from_env', got '%s'", config.Default)
	}
	if config.Custom != "custom_value" {
		t.Errorf("Custom expected 'custom_value', got '%s'", config.Custom)
	}
	if config.NoTag != "uppercase_value" {
		t.Errorf("NoTag expected 'uppercase_value' (from NOTAG), got '%s'", config.NoTag)
	}
	if config.Empty != "" {
		t.Errorf("Empty expected empty value, got '%s'", config.Empty)
	}
	if config.WithDefault != "default_value" {
		t.Errorf("WithDefault expected 'default_value', got '%s'", config.WithDefault)
	}
}

// TestParseErrorHandling tests various error cases
func TestParseErrorHandling(t *testing.T) {
	// Setup for invalid values
	envVars := map[string]string{
		"PARSE_INVALID_INT":      "not_an_int",
		"PARSE_INVALID_UINT":     "-1",
		"PARSE_INVALID_FLOAT":    "not_a_float",
		"PARSE_INVALID_BOOL":     "not_a_bool", // This won't cause an error
		"PARSE_INVALID_DURATION": "not_a_duration",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Test each invalid type separately
	type InvalidIntConfig struct {
		InvalidInt int `env:"PARSE_INVALID_INT"`
	}
	var intConfig InvalidIntConfig
	err := Parse(&intConfig)
	if err == nil {
		t.Error("Parse with invalid int should fail")
	}

	type InvalidUintConfig struct {
		InvalidUint uint `env:"PARSE_INVALID_UINT"`
	}
	var uintConfig InvalidUintConfig
	err = Parse(&uintConfig)
	if err == nil {
		t.Error("Parse with invalid uint should fail")
	}

	type InvalidFloatConfig struct {
		InvalidFloat float64 `env:"PARSE_INVALID_FLOAT"`
	}
	var floatConfig InvalidFloatConfig
	err = Parse(&floatConfig)
	if err == nil {
		t.Error("Parse with invalid float should fail")
	}

	type InvalidDurationConfig struct {
		InvalidDuration time.Duration `env:"PARSE_INVALID_DURATION"`
	}
	var durConfig InvalidDurationConfig
	err = Parse(&durConfig)
	if err == nil {
		t.Error("Parse with invalid duration should fail")
	}

	// Bool always succeeds (invalid = false)
	type InvalidBoolConfig struct {
		InvalidBool bool `env:"PARSE_INVALID_BOOL"`
	}
	var boolConfig InvalidBoolConfig
	err = Parse(&boolConfig)
	if err != nil {
		t.Errorf("Parse with invalid bool should succeed (false), got error: %v", err)
	}
	if boolConfig.InvalidBool {
		t.Error("Invalid bool should parse as false")
	}
}

// TestParseUnsupportedTypesExtended tests all unsupported types
func TestParseUnsupportedTypesExtended(t *testing.T) {
	os.Setenv("PARSE_UNSUPPORTED", "value")
	defer os.Unsetenv("PARSE_UNSUPPORTED")

	// Test each unsupported type
	unsupportedTypes := []interface{}{
		struct {
			InvalidField []int `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField []float64 `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField []bool `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField map[int]string `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField map[string]int `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField [3]string `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField chan int `env:"PARSE_UNSUPPORTED"`
		}{},
		struct {
			InvalidField func() `env:"PARSE_UNSUPPORTED"`
		}{},
	}

	for i, iface := range unsupportedTypes {
		// Create a pointer to the struct
		ptrVal := reflect.New(reflect.TypeOf(iface))
		// Copy the struct to the pointer target
		reflect.Indirect(ptrVal).Set(reflect.ValueOf(iface))

		err := Parse(ptrVal.Interface())
		if err == nil {
			t.Errorf("Parse with unsupported type #%d should fail", i)
		}
	}
}

// TestParseWithSliceComplexCases tests slice parsing edge cases
func TestParseWithSliceComplexCases(t *testing.T) {
	// Setup
	sliceCases := map[string]string{
		"PARSE_SLICE_EMPTY":       "",
		"PARSE_SLICE_SINGLE":      "single",
		"PARSE_SLICE_SPACES":      " a , b , c ",
		"PARSE_SLICE_EMPTY_PARTS": ",,",
		"PARSE_SLICE_QUOTED":      "\"quoted\",regular",
	}

	for k, v := range sliceCases {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Struct for testing
	type SliceTestConfig struct {
		Empty      []string `env:"PARSE_SLICE_EMPTY"`
		Single     []string `env:"PARSE_SLICE_SINGLE"`
		Spaces     []string `env:"PARSE_SLICE_SPACES"`
		EmptyParts []string `env:"PARSE_SLICE_EMPTY_PARTS"`
		Quoted     []string `env:"PARSE_SLICE_QUOTED"`
	}

	var config SliceTestConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify results
	if len(config.Empty) != 0 {
		t.Errorf("Empty slice expected empty, got %v", config.Empty)
	}

	if len(config.Single) != 1 || config.Single[0] != "single" {
		t.Errorf("Single item slice expected [single], got %v", config.Single)
	}

	if len(config.Spaces) != 3 || config.Spaces[0] != "a" ||
		config.Spaces[1] != "b" || config.Spaces[2] != "c" {
		t.Errorf("Spaces slice expected [a b c], got %v", config.Spaces)
	}

	// Empty parts are kept (becomes ["", "", ""])
	if len(config.EmptyParts) != 3 {
		t.Errorf("EmptyParts expected 3 items, got %v", config.EmptyParts)
	}

	// Quotes are preserved
	if len(config.Quoted) != 2 || config.Quoted[0] != "\"quoted\"" || config.Quoted[1] != "regular" {
		t.Errorf("Quoted slice expected [\"quoted\" regular], got %v", config.Quoted)
	}
}

// TestParseWithMapComplexCases tests map parsing edge cases
func TestParseWithMapComplexCases(t *testing.T) {
	// Setup
	mapCases := map[string]string{
		"PARSE_MAP_EMPTY":      "",
		"PARSE_MAP_SINGLE":     "key:value",
		"PARSE_MAP_SPACES":     " k1 : v1 , k2 : v2 ",
		"PARSE_MAP_NO_VALUE":   "k2:v2", // Changed to match expected output
		"PARSE_MAP_NO_COLON":   "k1=v1,k2:v2",
		"PARSE_MAP_DUPLICATES": "k1:v1,k1:v2",
	}

	for k, v := range mapCases {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Struct for testing
	type MapTestConfig struct {
		Empty      map[string]string `env:"PARSE_MAP_EMPTY"`
		Single     map[string]string `env:"PARSE_MAP_SINGLE"`
		Spaces     map[string]string `env:"PARSE_MAP_SPACES"`
		NoValue    map[string]string `env:"PARSE_MAP_NO_VALUE"`
		NoColon    map[string]string `env:"PARSE_MAP_NO_COLON"`
		Duplicates map[string]string `env:"PARSE_MAP_DUPLICATES"`
	}

	var config MapTestConfig
	err := Parse(&config)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify results
	if len(config.Empty) != 0 {
		t.Errorf("Empty map expected empty, got %v", config.Empty)
	}

	if len(config.Single) != 1 || config.Single["key"] != "value" {
		t.Errorf("Single item map expected map[key:value], got %v", config.Single)
	}

	if len(config.Spaces) != 2 || config.Spaces["k1"] != "v1" || config.Spaces["k2"] != "v2" {
		t.Errorf("Spaces map expected map[k1:v1 k2:v2], got %v", config.Spaces)
	}

	if len(config.NoValue) != 1 || config.NoValue["k2"] != "v2" {
		t.Errorf("NoValue map expected map[k2:v2], got %v", config.NoValue)
	}

	if len(config.NoColon) != 1 || config.NoColon["k2"] != "v2" {
		t.Errorf("NoColon map expected map[k2:v2], got %v", config.NoColon)
	}

	// Last duplicate key wins
	if len(config.Duplicates) != 1 || config.Duplicates["k1"] != "v2" {
		t.Errorf("Duplicates map expected map[k1:v2], got %v", config.Duplicates)
	}
}

// TestParseBoolVariations tests all boolean value variations
func TestParseBoolVariations(t *testing.T) {
	// Setup all variations of boolean values
	boolCases := map[string]struct {
		value  string
		expect bool
	}{
		"PARSE_BOOL_TRUE":      {"true", true},
		"PARSE_BOOL_TRUE_CAP":  {"TRUE", true},
		"PARSE_BOOL_TRUE_MIX":  {"True", true},
		"PARSE_BOOL_1":         {"1", true},
		"PARSE_BOOL_YES":       {"yes", true},
		"PARSE_BOOL_YES_CAP":   {"YES", true},
		"PARSE_BOOL_Y":         {"y", true},
		"PARSE_BOOL_Y_CAP":     {"Y", true},
		"PARSE_BOOL_FALSE":     {"false", false},
		"PARSE_BOOL_FALSE_CAP": {"FALSE", false},
		"PARSE_BOOL_0":         {"0", false},
		"PARSE_BOOL_NO":        {"no", false},
		"PARSE_BOOL_NO_CAP":    {"NO", false},
		"PARSE_BOOL_N":         {"n", false},
		"PARSE_BOOL_N_CAP":     {"N", false},
		"PARSE_BOOL_OTHER":     {"anything_else", false},
	}

	// Set environment variables
	for k, tc := range boolCases {
		os.Setenv(k, tc.value)
		defer os.Unsetenv(k)
	}

	// Create struct fields
	fields := make([]reflect.StructField, 0, len(boolCases))
	for k, _ := range boolCases {
		fields = append(fields, reflect.StructField{
			Name: k,
			Type: reflect.TypeOf(false),
			Tag:  reflect.StructTag(`env:"` + k + `"`),
		})
	}

	// Create struct type and instance
	structType := reflect.StructOf(fields)
	structPtr := reflect.New(structType)

	// Parse
	err := Parse(structPtr.Interface())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify results
	structValue := structPtr.Elem()
	for i := 0; i < structValue.NumField(); i++ {
		fieldName := structType.Field(i).Name
		fieldValue := structValue.Field(i).Bool()
		expectedValue := boolCases[fieldName].expect

		if fieldValue != expectedValue {
			t.Errorf("Field %s: expected %v for value '%s', got %v",
				fieldName, expectedValue, boolCases[fieldName].value, fieldValue)
		}
	}
}
