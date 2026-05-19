package cmd

import (
	"fmt"
	"os"

	"github.com/lmorchard/github-to-markdown/internal/templates"
	"github.com/spf13/cobra"
)

const defaultConfigContent = `# Configuration file for github-to-markdown
# CLI flags override these values; the GITHUB_TOKEN env var also works for ` + "`token`" + `.

# GitHub personal access token (prefer the GITHUB_TOKEN env var over storing this here)
# token: "ghp_xxxxxxxxxxxxxxxxxxxx"

# GitHub username to fetch activity for. Defaults to the authenticated user.
# user: "your-github-username"

# Time window. Accepts YYYY-MM-DD or RFC3339. Defaults: since = now-7d, until = now.
# since: "2026-05-11"
# until: "2026-05-18"

# Output file. Empty / unset means stdout.
# output: "activity.md"

# Path to a custom template file. Empty / unset means the embedded default.
# template: "github-to-markdown.md"

# Logging
verbose: false
debug: false
log_json: false
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration and template files",
	Long: `Create a default configuration file and a customizable template file.

This command generates:
  - github-to-markdown.yaml (configuration file)
  - github-to-markdown.md  (customizable template; override with --template-file)

Use --force to overwrite existing files.

Example:
  github-to-markdown init
  github-to-markdown init --template-file my-template.md
  github-to-markdown init --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := GetLogger()
		force, _ := cmd.Flags().GetBool("force")
		templateFile, _ := cmd.Flags().GetString("template-file")

		configFile := "github-to-markdown.yaml"

		configExists := fileExists(configFile)
		if configExists && !force {
			return fmt.Errorf("config file %s already exists (use --force to overwrite)", configFile)
		}

		templateExists := fileExists(templateFile)
		if templateExists && !force {
			return fmt.Errorf("template file %s already exists (use --force to overwrite)", templateFile)
		}

		if err := os.WriteFile(configFile, []byte(defaultConfigContent), 0o644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		if configExists {
			log.Infof("Overwrote %s", configFile)
		} else {
			log.Infof("Created %s", configFile)
		}

		templateContent, err := templates.GetDefaultTemplate()
		if err != nil {
			return fmt.Errorf("failed to get default template: %w", err)
		}
		if err := os.WriteFile(templateFile, []byte(templateContent), 0o644); err != nil {
			return fmt.Errorf("failed to create template file: %w", err)
		}
		if templateExists {
			log.Infof("Overwrote %s", templateFile)
		} else {
			log.Infof("Created %s", templateFile)
		}

		fmt.Printf("\nInitialization complete.\n\n")
		fmt.Printf("Next steps:\n")
		fmt.Printf("  1. Edit %s (or set GITHUB_TOKEN in your environment)\n", configFile)
		fmt.Printf("  2. (Optional) Customize %s\n", templateFile)
		fmt.Printf("  3. Run: github-to-markdown --help\n\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
	initCmd.Flags().String("template-file", "github-to-markdown.md", "Name of custom template file to create")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
