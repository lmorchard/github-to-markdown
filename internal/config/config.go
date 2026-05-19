package config

// Config holds application configuration
type Config struct {
	// Core settings
	Verbose bool
	Debug   bool
	LogJSON bool

	// GitHub auth + target user
	Token string
	User  string

	// Time window for activity (RFC3339 or YYYY-MM-DD, or empty for default)
	Since string
	Until string

	// Output destination ("" = stdout)
	Output string

	// Path to a custom template file ("" = embedded default)
	Template string
}
