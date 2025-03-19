package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dckristiono/go-env"
)

// AppConfig adalah contoh struct untuk diisi dengan Parse
type AppConfig struct {
	AppName        string        `env:"APP_NAME" default:"DefaultApp"`
	Port           int           `env:"PORT" default:"8080"`
	Debug          bool          `env:"DEBUG" default:"false"`
	Timeout        time.Duration `env:"TIMEOUT" default:"30s"`
	AllowedOrigins []string      `env:"ALLOWED_ORIGINS"`
}

func main() {
	fmt.Println("=== Go-Env Example ===")

	// Contoh 1: Menggunakan Fungsi Package Level
	fmt.Println("\n=== Basic Usage ===")
	appName := env.Get("APP_NAME", "DefaultApp")
	port := env.Int("PORT", 8080)
	debug := env.Bool("DEBUG", false)

	fmt.Printf("App Name: %s\n", appName)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Debug: %t\n", debug)

	// Contoh 2: Menggunakan Fluent API
	fmt.Println("\n=== Fluent API ===")
	dbHost := env.Key("DB_HOST").Default("localhost").String()
	dbPort := env.Key("DB_PORT").IntDefault(5432)
	dbName := env.Key("DB_NAME").Required().Default("defaultdb").String()

	fmt.Printf("Database: %s:%d/%s\n", dbHost, dbPort, dbName)

	// Contoh 3: Menggunakan Prefix
	fmt.Println("\n=== Prefix ===")
	adminEmail := env.With(env.WithPrefix("ADMIN_")).Key("EMAIL").String()
	adminEnabled := env.With(env.WithPrefix("ADMIN_")).Key("ENABLED").BoolDefault(false)

	fmt.Printf("Admin Email: %s\n", adminEmail)
	fmt.Printf("Admin Enabled: %t\n", adminEnabled)

	// Contoh 4: Struct Parsing
	fmt.Println("\n=== Struct Parsing ===")
	var config AppConfig
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Printf("App: %s running on port %d\n", config.AppName, config.Port)
	fmt.Printf("Debug mode: %t, Timeout: %v\n", config.Debug, config.Timeout)
	fmt.Printf("Allowed Origins: %v\n", config.AllowedOrigins)

	// Contoh 5: Mode Environment
	fmt.Println("\n=== Environment Mode ===")
	fmt.Printf("Current Mode: %s\n", env.GetMode())

	if env.IsProduction() {
		fmt.Println("Running in production mode")
	} else if env.IsStaging() {
		fmt.Println("Running in staging mode")
	} else if env.IsDevelopment() {
		fmt.Println("Running in development mode")
	}
}
