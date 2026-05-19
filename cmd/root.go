package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/lmorchard/github-to-markdown/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	log     = logrus.New()
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "github-to-markdown",
	Short: "Fetch recent GitHub activity and render it as Markdown",
	Long: `github-to-markdown fetches the authenticated user's recent activity from
GitHub's events feed and renders it as a Markdown document.

By default, output goes to stdout; use --output to write to a file.
Use 'init' to generate a config file and customizable template.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		setupLogging()
	},
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE:          runFetch,
}

// Execute adds all child commands to the root command and sets appropriate flags.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./github-to-markdown.yaml)")

	// Logging flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "debug output")
	rootCmd.PersistentFlags().Bool("log-json", false, "output logs in JSON format")

	// Fetch flags (on root since fetch is the default action)
	rootCmd.Flags().String("token", "", "GitHub personal access token (or set GITHUB_TOKEN)")
	rootCmd.Flags().String("user", "", "GitHub username (defaults to the authenticated user)")
	rootCmd.Flags().String("since", "", "start of time window (YYYY-MM-DD or RFC3339; default: 7 days ago)")
	rootCmd.Flags().String("until", "", "end of time window (YYYY-MM-DD or RFC3339; default: now)")
	rootCmd.Flags().StringP("output", "o", "", "write output to file (default: stdout)")
	rootCmd.Flags().String("template", "", "path to custom template file (default: embedded)")
	rootCmd.Flags().Bool("include-private", false, "include events from private repositories")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("log_json", rootCmd.PersistentFlags().Lookup("log-json"))
	_ = viper.BindPFlag("token", rootCmd.Flags().Lookup("token"))
	_ = viper.BindPFlag("user", rootCmd.Flags().Lookup("user"))
	_ = viper.BindPFlag("since", rootCmd.Flags().Lookup("since"))
	_ = viper.BindPFlag("until", rootCmd.Flags().Lookup("until"))
	_ = viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("template", rootCmd.Flags().Lookup("template"))
	_ = viper.BindPFlag("include_private", rootCmd.Flags().Lookup("include-private"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("github-to-markdown")
	}

	// Set defaults
	viper.SetDefault("verbose", false)
	viper.SetDefault("debug", false)
	viper.SetDefault("log_json", false)

	// Read in environment variables that match. Every config key is reachable
	// via GITHUB_<KEY> (e.g. token -> GITHUB_TOKEN, include_private -> GITHUB_INCLUDE_PRIVATE).
	viper.SetEnvPrefix("GITHUB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// Only error if config was explicitly specified
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}

// setupLogging configures the logger based on configuration
func setupLogging() {
	if viper.GetBool("log_json") {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	if viper.GetBool("debug") {
		log.SetLevel(logrus.DebugLevel)
	} else if viper.GetBool("verbose") {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.WarnLevel)
	}
}

// GetConfig returns the application configuration, loading it if necessary
func GetConfig() *config.Config {
	if cfg == nil {
		cfg = &config.Config{
			Verbose:        viper.GetBool("verbose"),
			Debug:          viper.GetBool("debug"),
			LogJSON:        viper.GetBool("log_json"),
			Token:          viper.GetString("token"),
			User:           viper.GetString("user"),
			Since:          viper.GetString("since"),
			Until:          viper.GetString("until"),
			Output:         viper.GetString("output"),
			Template:       viper.GetString("template"),
			IncludePrivate: viper.GetBool("include_private"),
		}
	}
	return cfg
}

// GetLogger returns the configured logger
func GetLogger() *logrus.Logger {
	return log
}
