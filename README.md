# Go Env

![Go Test](https://github.com/dckristiono/go-env/workflows/Go%20Test/badge.svg)
![Lint](https://github.com/dckristiono/go-env/workflows/Lint/badge.svg)
[![codecov](https://codecov.io/gh/dckristiono/go-env/branch/main/graph/badge.svg)](https://codecov.io/gh/dckristiono/go-env)
[![Go Report Card](https://goreportcard.com/badge/github.com/dckristiono/go-env)](https://goreportcard.com/report/github.com/dckristiono/go-env)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dckristiono/go-env)](https://pkg.go.dev/github.com/dckristiono/go-env)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Env adalah modul Golang ringan dan fleksibel untuk mengelola konfigurasi environment berdasarkan mode aplikasi (production, staging, development). Modul ini menyediakan beberapa pendekatan yang berbeda untuk mengakses variabel environment.

## Fitur

- 🚀 **Deteksi Mode Otomatis**: Deteksi mode environment berdasarkan ketersediaan file (.env, .env.staging, .env.development)
- 🔄 **Multiple Environment Files**: Support untuk file konfigurasi terpisah (.env, .env.staging, .env.development)
- 🔧 **API Fleksibel**: Berbagai cara untuk mengakses konfigurasi (Fluent API, Helper Functions, Struct Parsing)
- 📌 **Pengelompokan dengan Prefix**: Dukungan untuk prefiks pada variabel environment
- ✅ **Validasi**: Kemampuan memvalidasi field yang required
- 🧩 **Type-Safe dengan Struct Parsing**: Parse environment variables ke struct dengan validasi tipe dan nilai default
- 🔗 **Method Chaining**: Fluent API dengan method chaining untuk akses yang lebih ergonomis

## Instalasi

```bash
go get github.com/dckristiono/go-env
```

## Quickstart

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/dckristiono/go-env"
)

func main() {
	// Menggunakan fungsi package level sederhana
	appName := env.Get("APP_NAME", "DefaultApp")
	fmt.Printf("App Name: %s\n", appName)

	// Menggunakan helper tanpa error handling
	port := env.Int("PORT", 8080)
	debug := env.Bool("DEBUG", false)
	fmt.Printf("Port: %d, Debug: %t\n", port, debug)
}
```

### Fluent API

```go
package main

import (
	"fmt"
	"github.com/dckristiono/go-env"
)

func main() {
	// Menggunakan Fluent API
	dbHost := env.Key("DB_HOST").Required().String()
	dbPort := env.Key("DB_PORT").IntDefault(5432)

	fmt.Printf("Database: %s:%d\n", dbHost, dbPort)

	// Dengan prefix
	adminEmail := env.With(env.WithPrefix("ADMIN_")).Key("EMAIL").String()
	fmt.Printf("Admin Email: %s\n", adminEmail)
}
```

### Struct Parsing

```go
package main

import (
	"fmt"
	"log"
	"time"
	"github.com/dckristiono/go-env"
)

type AppConfig struct {
	AppName       string        `env:"APP_NAME" default:"DefaultApp"`
	Port          int           `env:"PORT" default:"8080"`
	Debug         bool          `env:"DEBUG" default:"false"`
	Timeout       time.Duration `env:"TIMEOUT" default:"30s"`
	AllowedOrigins []string     `env:"ALLOWED_ORIGINS"`
}

func main() {
	var config AppConfig

	if err := env.Parse(&config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Printf("App: %s running on port %d\n", config.AppName, config.Port)
	fmt.Printf("Debug mode: %t, Timeout: %v\n", config.Debug, config.Timeout)
	fmt.Printf("Allowed Origins: %v\n", config.AllowedOrigins)
}
```

## Mode Environment

Modul ini mendukung 3 mode environment: `production`, `staging`, dan `development`, yang menentukan file konfigurasi mana yang akan digunakan:

- **Production**: Menggunakan file `.env`
- **Staging**: Menggunakan file `.env.staging`
- **Development**: Menggunakan file `.env.development`

Mode default ditentukan berdasarkan ketersediaan file:

1. Jika ketiga file ada (.env, .env.staging, .env.development), default ke `development`
2. Jika hanya ada .env dan .env.staging, default ke `staging`
3. Jika hanya ada .env, default ke `production`

Mode juga dapat diatur secara eksplisit melalui variabel environment `APP_ENV` atau saat inisialisasi:

```go
// Inisialisasi dengan mode explicit
env.Initialize(env.WithMode("production"))

// Periksa mode saat ini
if env.IsProduction() {
// Production logic
} else if env.IsStaging() {
// Staging logic
} else if env.IsDevelopment() {
// Development logic
}
```

## Dokumentasi API Lengkap

### Package Level Functions

```go
// Fungsi Get dasar
env.Get("KEY", "default")                // string
env.GetInt("KEY", 42)                    // (int, error)
env.GetFloat64("KEY", 3.14)              // (float64, error)
env.GetBool("KEY", false)                // bool
env.GetDuration("KEY", 30*time.Second)   // (time.Duration, error)
env.GetSlice("KEY", ",", []string{})     // []string
env.GetMap("KEY", map[string]string{})   // map[string]string

// Helper functions (tanpa error)
env.String("KEY", "default")             // string
env.Int("KEY", 42)                       // int
env.Float64("KEY", 3.14)                 // float64
env.Bool("KEY", false)                   // bool
env.Duration("KEY", 30*time.Second)      // time.Duration
env.Slice("KEY", ",", []string{})        // []string
env.Map("KEY", map[string]string{})      // map[string]string

// Mode functions
env.GetMode()                            // string
env.IsProduction()                       // bool
env.IsStaging()                          // bool
env.IsDevelopment()                      // bool

// Parse ke struct
env.Parse(&config)                       // error
```

### Fluent API

```go
// Fluent API dengan method chaining
env.Key("KEY")                           // *result
.Required()                            // *result (validasi)
.Default("default")                    // *result (nilai default)
.String()                              // string (hasil akhir)

// Tipe hasil lainnya  
env.Key("KEY").Int()                     // (int, error)
env.Key("KEY").IntDefault(42)            // int
env.Key("KEY").Float64()                 // (float64, error)
env.Key("KEY").Float64Default(3.14)      // float64
env.Key("KEY").Bool()                    // bool
env.Key("KEY").BoolDefault(false)        // bool
env.Key("KEY").Duration()                // (time.Duration, error)
env.Key("KEY").DurationDefault(30*time.Second) // time.Duration
env.Key("KEY").Slice(",")                // []string
env.Key("KEY").SliceDefault(",", []string{}) // []string
env.Key("KEY").Map()                     // map[string]string
env.Key("KEY").MapDefault(map[string]string{}) // map[string]string
```

### Configuration Options

```go
// Menggunakan options
env.With(env.WithMode("staging"))        // *Config
env.With(env.WithPrefix("APP_"))         // *Config

// Menggunakan options bersama
env.With(
env.WithMode("staging"),
env.WithPrefix("DB_"),
)                                        // *Config

// Inisialisasi ulang instance default
env.Initialize(env.WithMode("production")) // error
```

## Contoh Format File .env

```
# .env (Production)
APP_NAME=MyApp
PORT=80
DB_HOST=db.example.com
DEBUG=false
TIMEOUT=60s
ALLOWED_ORIGINS=example.com,api.example.com
FEATURES=darkMode:false,analytics:true

# .env.staging (Staging)
APP_NAME=MyApp-Staging
PORT=8080
DB_HOST=staging-db.example.com
DEBUG=true
TIMEOUT=30s
ALLOWED_ORIGINS=staging.example.com,localhost:3000
FEATURES=darkMode:true,analytics:true

# .env.development (Development)
APP_NAME=MyApp-Dev
PORT=3000
DB_HOST=localhost
DEBUG=true
TIMEOUT=10s
ALLOWED_ORIGINS=localhost:3000,localhost:8080
FEATURES=darkMode:true,analytics:false
```

## Lisensi

Projek ini dibawah lisensi MIT - lihat file [LICENSE](LICENSE) untuk detail.