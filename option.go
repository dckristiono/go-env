package env

// ConfigOption adalah sebuah option untuk fluent configuration
type ConfigOption func(*Config)

// WithMode menentukan mode environment
func WithMode(mode string) ConfigOption {
	return func(c *Config) {
		c.Mode = mode
	}
}

// WithPrefix menentukan prefix untuk environment variables
func WithPrefix(prefix string) ConfigOption {
	return func(c *Config) {
		c.Prefix = prefix
	}
}
