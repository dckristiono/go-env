package env

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// result adalah struct untuk hasil operasi dengan validasi
type result struct {
	config *Config
	key    string
	value  string
	err    error
}

// Required menandai bahwa nilai harus ada
func (r *result) Required() *result {
	if r.err != nil {
		return r
	}

	if r.value == "" {
		r.err = fmt.Errorf("environment variable %s wajib diisi", r.key)
	}
	return r
}

// Default menetapkan nilai default
func (r *result) Default(defaultValue string) *result {
	if r.err != nil {
		return r
	}

	if r.value == "" {
		r.value = defaultValue
	}
	return r
}

// String mengembalikan nilai sebagai string
func (r *result) String() string {
	return r.value
}

// Int mengembalikan nilai sebagai int
func (r *result) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.value == "" {
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", r.key)
	}

	return strconv.Atoi(r.value)
}

// IntDefault mengembalikan nilai sebagai int dengan nilai default
func (r *result) IntDefault(defaultValue int) int {
	value, err := r.Int()
	if err != nil {
		return defaultValue
	}
	return value
}

// Float64 mengembalikan nilai sebagai float64
func (r *result) Float64() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.value == "" {
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", r.key)
	}

	return strconv.ParseFloat(r.value, 64)
}

// Float64Default mengembalikan nilai sebagai float64 dengan nilai default
func (r *result) Float64Default(defaultValue float64) float64 {
	value, err := r.Float64()
	if err != nil {
		return defaultValue
	}
	return value
}

// Bool mengembalikan nilai sebagai boolean
func (r *result) Bool() bool {
	if r.err != nil {
		return false
	}

	if r.value == "" {
		return false
	}

	value := strings.ToLower(r.value)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}

// BoolDefault mengembalikan nilai sebagai boolean dengan nilai default
func (r *result) BoolDefault(defaultValue bool) bool {
	if r.err != nil || r.value == "" {
		return defaultValue
	}
	return r.Bool()
}

// Duration mengembalikan nilai sebagai time.Duration
func (r *result) Duration() (time.Duration, error) {
	if r.err != nil {
		return 0, r.err
	}

	if r.value == "" {
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", r.key)
	}

	return time.ParseDuration(r.value)
}

// DurationDefault mengembalikan nilai sebagai time.Duration dengan nilai default
func (r *result) DurationDefault(defaultValue time.Duration) time.Duration {
	value, err := r.Duration()
	if err != nil {
		return defaultValue
	}
	return value
}

// Slice mengembalikan nilai sebagai slice string
func (r *result) Slice(delimiter string) []string {
	if r.err != nil {
		return []string{}
	}

	if r.value == "" {
		return []string{}
	}

	if delimiter == "" {
		delimiter = ","
	}

	parts := strings.Split(r.value, delimiter)
	// Trim space dari setiap elemen
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

// SliceDefault mengembalikan nilai sebagai slice string dengan nilai default
func (r *result) SliceDefault(delimiter string, defaultValue []string) []string {
	if r.err != nil || r.value == "" {
		return defaultValue
	}
	return r.Slice(delimiter)
}

// Map mengembalikan nilai sebagai map[string]string
func (r *result) Map() map[string]string {
	if r.err != nil {
		return map[string]string{}
	}

	if r.value == "" {
		return map[string]string{}
	}

	result := make(map[string]string)
	parts := strings.Split(r.value, ",")

	for _, part := range parts {
		keyValue := strings.SplitN(part, ":", 2)
		if len(keyValue) == 2 {
			k := strings.TrimSpace(keyValue[0])
			v := strings.TrimSpace(keyValue[1])
			result[k] = v
		}
	}

	return result
}

// MapDefault mengembalikan nilai sebagai map[string]string dengan nilai default
func (r *result) MapDefault(defaultValue map[string]string) map[string]string {
	if r.err != nil || r.value == "" {
		return defaultValue
	}
	return r.Map()
}
