// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
)

// Use the default paths from the config package instead of maintaining duplicates
var (
	defaultConfigDir  string
	defaultConfigFile string
	defaultTeamsDir   string
)

func init() {
	// Initialise paths using the config package functions
	defaultConfigDir = config.GetDefaultConfigDir()
	defaultConfigFile = config.GetDefaultConfigFile()
	defaultTeamsDir = config.GetDefaultTeamsDir()
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage viaplay-cli configuration and team settings",
	Long: `Manage global and team-specific configuration for viaplay-cli.

- View and update CLI settings
- Scaffold team config folders and example JSON files
- Set up directories for rulesets, secrets, and environments
- Integrate with $HOME/.config/viaplay/config.yaml by default

Examples:
  vip config init                     # Initialize config file
  vip config init --team myteam       # Initialize with team config
  vip config get default_account      # Get a config value
  vip config set default_account user # Set a config value
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// initCmd scaffolds example config files for a team (optionally)
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize viaplay-cli configuration",
	Long: `Initialize the viaplay-cli configuration directory and config.yaml.
Optionally, scaffold example team config files (ruleset.json, secrets.json, envs/*.json)
with --team <team>. Use --override to force overwrite existing configuration files.

Examples:
  vip config init
  vip config init --team platform
  vip config init --team platform --override
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		override, err := cmd.Flags().GetBool("override")
		if err != nil {
			return fmt.Errorf("failed to get 'override' flag: %w", err)
		}
		team, err := cmd.Flags().GetString("team")
		if err != nil {
			return fmt.Errorf("failed to get 'team' flag: %w", err)
		}

		// Initialise main config file
		if err := initializeConfigFile(override, team); err != nil {
			return err
		}

		// Scaffold team config files if requested
		if team != "" {
			if err := scaffoldTeamConfig(team, override); err != nil {
				return err
			}
		}

		// Show success message and next steps
		showInitSuccessMessage(team)

		return nil
	},
}

// getCmd gets a configuration value
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a configuration value from the config file.
If no key is specified, all configuration values will be displayed.

Examples:
  vip config get default_account
  vip config get default_team
  vip config get           # Show all values
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure we're reading from the config file
		if err := ensureConfigLoaded(); err != nil {
			return err
		}

		// If no key is provided, show all config
		if len(args) == 0 {
			return showAllConfig()
		}

		// Get the specified key
		return getConfigValue(args[0])
	},
}

// setCmd sets a configuration value
var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a configuration value in the config file.

Examples:
  vip config set default_account myusername
  vip config set default_team myteam
  vip config set default_language typescript
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateConfigFile(args[0], args[1])
	},
}

// pathsCmd shows the configuration paths
var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show configuration paths",
	Long: `Show all configuration paths used by viaplay-cli.

This includes:
- Config file location
- Config directory
- Teams directory
`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfigPaths()
	},
}

func init() {
	// Add all subcommands to the config command
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(pathsCmd)

	// Define flags for the init command
	initCmd.Flags().String("team", "", "Team name to scaffold configs for (optional)")
	initCmd.Flags().Bool("override", false, "Override existing config files if they exist")
}

// Path handling functions

// ensureConfigLoaded makes sure viper has loaded the config file
func ensureConfigLoaded() error {
	// Check if config file exists
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found, run 'viaplay-cli config init' first")
	}

	// Use the config package's initialization function if viper hasn't loaded a config yet
	if viper.ConfigFileUsed() == "" {
		if err := config.InitConfig(defaultConfigFile); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
	}

	return nil
}

// showConfigPaths displays all configuration paths
func showConfigPaths() {
	fmt.Println("Configuration paths:")
	fmt.Printf("  Config file:    %s\n", defaultConfigFile)
	fmt.Printf("  Config dir:     %s\n", defaultConfigDir)
	fmt.Printf("  Teams dir:      %s\n", defaultTeamsDir)

	// Show the actual config file being used by viper
	if viper.ConfigFileUsed() != "" && viper.ConfigFileUsed() != defaultConfigFile {
		fmt.Printf("  Active config:  %s\n", viper.ConfigFileUsed())
	}
}

// Config file manipulation functions

// initializeConfigFile creates or updates the main config file
func initializeConfigFile(override bool, team string) error {
	// Get username for suggested default_account value
	username := getUsernameFromGitHub()

	// Use the centralised function from the config package
	if err := config.InitializeConfigFile(defaultConfigFile, override, team, username); err != nil {
		return err
	}

	// Display appropriate messages based on the operation
	configExists := true
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		configExists = false
	}

	if configExists && override {
		fmt.Println("Overrode existing configuration file:", defaultConfigFile)
	} else if !configExists {
		fmt.Println("Created configuration file:", defaultConfigFile)
	} else {
		fmt.Println("Configuration file already exists:", defaultConfigFile)
		fmt.Println("Use --override to replace it with a fresh configuration.")
	}

	if configExists || override {
		fmt.Println("You can edit this file to customize viaplay-cli behavior.")
	}

	return nil
}

// getUsernameFromGitHub attempts to get the GitHub username if authenticated
func getUsernameFromGitHub() string {
	username := ""
	token, err := gh.GetToken()
	if err == nil && token != "" {
		usernameResult, err := gh.GetAuthenticatedUser(token)
		if err == nil {
			username = usernameResult
		}
	}
	return username
}

// Config value functions

// showAllConfig displays all configuration values
func showAllConfig() error {
	allSettings := viper.AllSettings()
	fmt.Println("Current configuration:")
	for k, v := range allSettings {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return nil
}

// getConfigValue retrieves and displays a specific config value
func getConfigValue(key string) error {
	if !viper.IsSet(key) {
		return fmt.Errorf("key %q not found in config", key)
	}

	value := viper.Get(key)
	fmt.Printf("%v\n", value)
	return nil
}

// updateConfigFile updates a value in the configuration file while preserving comments
func updateConfigFile(key, value string) error {
	if err := ensureConfigLoaded(); err != nil {
		return err
	}

	// Use the config package's UpdateConfigValue function
	if err := config.UpdateConfigValue(defaultConfigFile, key, value); err != nil {
		return err
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

// Team config scaffolding functions

// scaffoldTeamConfig creates the team configuration files and directories
func scaffoldTeamConfig(team string, override bool) error {
	// Use the config package's team config creation functionality
	_, err := config.CreateTeamConfig(defaultTeamsDir, team, override)
	if err != nil {
		return err
	}

	fmt.Printf("\nScaffolded config files for team '%s' in %s\n", team, defaultTeamsDir)
	return nil
}

// Message display functions

// showInitSuccessMessage displays a success message after initialization
func showInitSuccessMessage(team string) {
	fmt.Println("\n✅ viaplay-cli initialized successfully!")
	fmt.Println("- Configuration directory: ", defaultConfigDir)
	fmt.Println("- Teams directory:        ", defaultTeamsDir)

	// Provide hints for next steps
	if team == "" {
		fmt.Println("\nTip: Set up a team configuration with:")
		fmt.Println("  vip config init --team <team-name>")
	}
}
