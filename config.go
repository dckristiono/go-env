// Package env menyediakan fungsi untuk membaca konfigurasi dari file .env
// berdasarkan mode environment (production, staging, development)
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Mode environment yang didukung
const (
	Production  = "production"
	Staging     = "staging"
	Development = "development"
)

// defaultInstance adalah instance singleton dari Config yang digunakan untuk fungsi-fungsi level package
var (
	defaultInstance *Config
	once            sync.Once
	initErr         error
)

// Config menyimpan konfigurasi environment
type Config struct {
	Mode   string
	Prefix string
}

// getDefaultInstance menginisialisasi dan mengembalikan instance singleton
func getDefaultInstance() (*Config, error) {
	once.Do(func() {
		defaultInstance, initErr = New()
	})
	return defaultInstance, initErr
}

// New membuat instance Config baru dengan opsi yang diberikan
func New(options ...ConfigOption) (*Config, error) {
	// Default config
	config := &Config{
		Mode:   determineDefaultMode(),
		Prefix: "",
	}

	// Terapkan options jika ada
	for _, option := range options {
		option(config)
	}

	// Load file .env sesuai dengan mode
	if err := config.Load(); err != nil {
		return nil, err
	}

	return config, nil
}

// determineDefaultMode menentukan mode default berdasarkan ketersediaan file
func determineDefaultMode() string {
	// Cek jika mode diatur melalui APP_ENV
	if envMode := os.Getenv("APP_ENV"); envMode != "" {
		return envMode
	}

	// Tentukan mode default berdasarkan file yang tersedia
	hasEnv := fileExists(".env")
	hasStaging := fileExists(".env.staging")
	hasDev := fileExists(".env.development")

	switch {
	case hasEnv && hasStaging && hasDev:
		return Development
	case hasEnv && hasStaging:
		return Staging
	case hasEnv:
		return Production
	default:
		return Development // Fallback ke development
	}
}

// fileExists memeriksa apakah file ada
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Load membaca file .env sesuai dengan mode environment
func (c *Config) Load() error {
	var envFile string

	switch c.Mode {
	case Production:
		envFile = ".env"
	case Staging:
		envFile = ".env.staging"
	case Development:
		envFile = ".env.development"
	default:
		return fmt.Errorf("mode environment tidak valid: %s", c.Mode)
	}

	// Periksa apakah file ada
	if !fileExists(envFile) {
		// Jika file tidak ada dan mode bukan production, berikan peringatan tapi jangan error
		if c.Mode != Production {
			fmt.Printf("Peringatan: File %s tidak ditemukan\n", envFile)
			return nil
		}
		return fmt.Errorf("file %s tidak ditemukan", envFile)
	}

	// Load file .env
	return godotenv.Load(envFile)
}

// prependPrefix menambahkan prefix ke key jika ada
func (c *Config) prependPrefix(key string) string {
	if c.Prefix == "" {
		return key
	}
	return c.Prefix + key
}

// From membuat instance baru dengan opsi untuk mendukung chaining
func (c *Config) From(options ...ConfigOption) *Config {
	newConfig := &Config{
		Mode:   c.Mode,
		Prefix: c.Prefix,
	}

	for _, option := range options {
		option(newConfig)
	}

	return newConfig
}

// Key menghasilkan result untuk key tertentu untuk mendukung chaining
func (c *Config) Key(key string) *result {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	return &result{
		config: c,
		key:    prefixedKey,
		value:  value,
		err:    nil,
	}
}

// Get mengambil nilai environment variable sebagai string
func (c *Config) Get(key string, defaultValue ...string) string {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetInt mengambil nilai environment variable sebagai integer
func (c *Config) GetInt(key string, defaultValue ...int) (int, error) {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", prefixedKey)
	}

	return strconv.Atoi(value)
}

// GetInt64 mengambil nilai environment variable sebagai int64
func (c *Config) GetInt64(key string, defaultValue ...int64) (int64, error) {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", prefixedKey)
	}

	return strconv.ParseInt(value, 10, 64)
}

// GetFloat64 mengambil nilai environment variable sebagai float64
func (c *Config) GetFloat64(key string, defaultValue ...float64) (float64, error) {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", prefixedKey)
	}

	return strconv.ParseFloat(value, 64)
}

// GetBool mengambil nilai environment variable sebagai boolean
func (c *Config) GetBool(key string, defaultValue ...bool) bool {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}

// GetDuration mengambil nilai environment variable sebagai time.Duration
func (c *Config) GetDuration(key string, defaultValue ...time.Duration) (time.Duration, error) {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, fmt.Errorf("environment variable %s tidak ditemukan", prefixedKey)
	}

	return time.ParseDuration(value)
}

// GetSlice mengambil nilai environment variable sebagai slice string
// Nilai dalam file .env harus dipisahkan dengan delimiter (defaultnya ",")
func (c *Config) GetSlice(key string, delimiter string, defaultValue ...[]string) []string {
	if delimiter == "" {
		delimiter = ","
	}

	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return []string{}
	}

	parts := strings.Split(value, delimiter)
	// Trim space dari setiap elemen
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}

// GetMap mengambil nilai environment variable sebagai map[string]string
// Format dalam file .env harus key1:value1,key2:value2
func (c *Config) GetMap(key string, defaultValue ...map[string]string) map[string]string {
	prefixedKey := c.prependPrefix(key)
	value := os.Getenv(prefixedKey)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return map[string]string{}
	}

	result := make(map[string]string)
	parts := strings.Split(value, ",")

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

// GetMode mengembalikan mode environment saat ini
func (c *Config) GetMode() string {
	return c.Mode
}

// IsProduction memeriksa apakah mode saat ini adalah production
func (c *Config) IsProduction() bool {
	return c.Mode == Production
}

// IsStaging memeriksa apakah mode saat ini adalah staging
func (c *Config) IsStaging() bool {
	return c.Mode == Staging
}

// IsDevelopment memeriksa apakah mode saat ini adalah development
func (c *Config) IsDevelopment() bool {
	return c.Mode == Development
}

// -----------------------------
// Fungsi-fungsi level package
// -----------------------------

// Initialize secara eksplisit menginisialisasi singleton dengan opsi
func Initialize(options ...ConfigOption) error {
	config, err := New(options...)
	if err != nil {
		return err
	}

	defaultInstance = config
	initErr = nil

	return nil
}

// With mengembalikan Config dengan opsi yang ditentukan
func With(options ...ConfigOption) *Config {
	cfg, err := getDefaultInstance()
	if err != nil {
		// Buat instance baru jika ada error
		newCfg, _ := New(options...)
		return newCfg
	}

	return cfg.From(options...)
}

// Get adalah fungsi level package yang mengambil nilai string dari environment
func Get(key string, defaultValue ...string) string {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return cfg.Get(key, defaultValue...)
}

// GetInt adalah fungsi level package yang mengambil nilai int dari environment
func GetInt(key string, defaultValue ...int) (int, error) {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}
	return cfg.GetInt(key, defaultValue...)
}

// GetInt64 adalah fungsi level package yang mengambil nilai int64 dari environment
func GetInt64(key string, defaultValue ...int64) (int64, error) {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}
	return cfg.GetInt64(key, defaultValue...)
}

// GetFloat64 adalah fungsi level package yang mengambil nilai float64 dari environment
func GetFloat64(key string, defaultValue ...float64) (float64, error) {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}
	return cfg.GetFloat64(key, defaultValue...)
}

// GetBool adalah fungsi level package yang mengambil nilai bool dari environment
func GetBool(key string, defaultValue ...bool) bool {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
	return cfg.GetBool(key, defaultValue...)
}

// GetDuration adalah fungsi level package yang mengambil nilai time.Duration dari environment
func GetDuration(key string, defaultValue ...time.Duration) (time.Duration, error) {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, err
	}
	return cfg.GetDuration(key, defaultValue...)
}

// GetSlice adalah fungsi level package yang mengambil nilai []string dari environment
func GetSlice(key string, delimiter string, defaultValue ...[]string) []string {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return []string{}
	}
	return cfg.GetSlice(key, delimiter, defaultValue...)
}

// GetMap adalah fungsi level package yang mengambil nilai map[string]string dari environment
func GetMap(key string, defaultValue ...map[string]string) map[string]string {
	cfg, err := getDefaultInstance()
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return map[string]string{}
	}
	return cfg.GetMap(key, defaultValue...)
}

// GetMode adalah fungsi level package yang mengembalikan mode saat ini
func GetMode() string {
	cfg, err := getDefaultInstance()
	if err != nil {
		return ""
	}
	return cfg.GetMode()
}

// IsProduction adalah fungsi level package yang memeriksa apakah mode saat ini adalah production
func IsProduction() bool {
	cfg, err := getDefaultInstance()
	if err != nil {
		return false
	}
	return cfg.IsProduction()
}

// IsStaging adalah fungsi level package yang memeriksa apakah mode saat ini adalah staging
func IsStaging() bool {
	cfg, err := getDefaultInstance()
	if err != nil {
		return false
	}
	return cfg.IsStaging()
}

// IsDevelopment adalah fungsi level package yang memeriksa apakah mode saat ini adalah development
func IsDevelopment() bool {
	cfg, err := getDefaultInstance()
	if err != nil {
		return false
	}
	return cfg.IsDevelopment()
}

// Key adalah fungsi level package yang mengembalikan result untuk key tertentu
func Key(key string) *result {
	cfg, err := getDefaultInstance()
	if err != nil {
		return &result{err: err}
	}
	return cfg.Key(key)
}

// String mengambil nilai environment variable sebagai string
func String(key string, defaultValue ...string) string {
	return Get(key, defaultValue...)
}

// Int mengambil nilai environment variable sebagai int
func Int(key string, defaultValue ...int) int {
	value, err := GetInt(key, defaultValue...)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// Float64 mengambil nilai environment variable sebagai float64
func Float64(key string, defaultValue ...float64) float64 {
	value, err := GetFloat64(key, defaultValue...)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// Bool mengambil nilai environment variable sebagai boolean
func Bool(key string, defaultValue ...bool) bool {
	return GetBool(key, defaultValue...)
}

// Duration mengambil nilai environment variable sebagai time.Duration
func Duration(key string, defaultValue ...time.Duration) time.Duration {
	value, err := GetDuration(key, defaultValue...)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// Slice mengambil nilai environment variable sebagai []string
func Slice(key string, delimiter string, defaultValue ...[]string) []string {
	return GetSlice(key, delimiter, defaultValue...)
}

// Map mengambil nilai environment variable sebagai map[string]string
func Map(key string, defaultValue ...map[string]string) map[string]string {
	return GetMap(key, defaultValue...)
}
