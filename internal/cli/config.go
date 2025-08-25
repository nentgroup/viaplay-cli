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
	"github.com/nentgroup/viaplay-cli/internal/output"
)

// Use the default paths from the config package instead of maintaining duplicates
var (
	defaultConfigDir  string
	defaultConfigFile string
	defaultTeamsDir   string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage viaplay-cli configuration and team settings",
	Long: `Manage global and team-specific configuration for viaplay-cli.

- View CLI settings
- Scaffold team config folders and example JSON files
- Set up directories for rulesets, secrets, and environments
- Integrate with $HOME/.config/viaplay/config.yaml by default

Examples:
  vip config init                     				# Initialize config file
  vip config init --team myteam --organization nentgroup        # Initialize with team config
  vip config get default_account      				# Get a config value
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
Optionally scaffold team config files with --team <team> and --org <organization>.

This command will:
1. Create the main config.yaml with settings from your authenticated GitHub account
2. Set up personal account configurations in ~/.config/viaplay/personal/
3. Set up team configurations when specified with --team flag

When run without flags, it will guide you through an interactive selection of organization and team.

Examples:
  vip config init                          # Initialize config using your GitHub account
  vip config init --team platform          # Initialize with team config
  vip config init --team platform --org myorg
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if the user is authenticated with GitHub
		token, err := gh.GetToken()
		if err != nil || token == "" {
			return fmt.Errorf("you must be authenticated with GitHub to initialize viaplay-cli configuration. Run 'vip auth login' first")
		}

		team, err := cmd.Flags().GetString("team")
		if err != nil {
			return fmt.Errorf("failed to get 'team' flag: %w", err)
		}

		organization, err := cmd.Flags().GetString("organization")
		if err != nil {
			return fmt.Errorf("failed to get 'org' flag: %w", err)
		}

		override, err := cmd.Flags().GetBool("override")
		if err != nil {
			return fmt.Errorf("failed to get 'override' flag: %w", err)
		}

		// Create GitHub client
		ghClient := gh.NewGitHubClient(token)

		// Get authenticated user info
		username, err := ghClient.GetAuthenticatedUser()
		if err != nil {
			return fmt.Errorf("failed to get authenticated user: %w", err)
		}
		fmt.Printf("Authenticated as: %s\n", output.Bold(username))

		// Interactive organisation and team selection if neither team nor organisation flags are set
		if !cmd.Flags().Changed("team") && !cmd.Flags().Changed("organization") {
			// Interactive selection of organization
			selectedOrg, err := SelectOrganizationWithBubbles(ghClient)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				// Continue without organization if there's an error
			} else if selectedOrg != "" {
				organization = selectedOrg

				// If we have an organization, also select a team
				selectedTeam, err := SelectTeamWithBubbles(ghClient, organization)
				if err != nil {
					fmt.Printf("Warning: %v\n", err)
					// Continue without team if there's an error
				} else if selectedTeam != "" {
					team = selectedTeam
				}
			}
		}

		// Set up team configurations if specified
		if team != "" && organization != "" {
			if err := scaffoldTeamConfig(organization, team, override); err != nil {
				return err
			}
		}

		// Always set up personal configurations for the authenticated user
		if err := scaffoldPersonalConfig(username, override); err != nil {
			return err
		}

		if err := initializeConfigFile(organization, team, override); err != nil {
			return fmt.Errorf("failed to initialize config file: %w", err)
		}

		// Show success message and next steps
		showInitSuccessMessage(team, organization)

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
	fmt.Printf("  Org dir:     %s\n", defaultConfigDir)
	fmt.Printf("  Teams dir:      %s\n", defaultTeamsDir)

	// Show the actual config file being used by viper
	if viper.ConfigFileUsed() != "" && viper.ConfigFileUsed() != defaultConfigFile {
		fmt.Printf("  Active config:  %s\n", viper.ConfigFileUsed())
	}
}

// initializeConfigFile creates or updates the main config file
func initializeConfigFile(org string, team string, override bool) error { //nolint:gofumpt
	// Use the centralised function from the config package
	if err := config.InitializeConfigFile(defaultConfigFile, override, team, org); err != nil {
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

// scaffoldTeamConfig creates the team configuration files and directories
func scaffoldTeamConfig(org string, team string, override bool) error {
	// Use the config package's team setup functionality (which will create any needed directories)
	result, err := config.SetupTeam(org, team, override)
	if err != nil {
		return err
	}

	// Print a summary of what happened
	if len(result.Created) > 0 {
		fmt.Printf("\nCreated %d config files for team '%s'\n", len(result.Created), team)
	}
	if len(result.Overrode) > 0 {
		fmt.Printf("Overrode %d existing config files for team '%s'\n", len(result.Overrode), team)
	}
	if len(result.Skipped) > 0 {
		fmt.Printf("Skipped %d existing config files for team '%s'\n", len(result.Skipped), team)
	}

	return nil
}

// scaffoldPersonalConfig creates personal account configurations
func scaffoldPersonalConfig(username string, override bool) error {
	result, err := config.SetupPersonal(username, override)
	if err != nil {
		return err
	}

	// Print results
	for _, path := range result.Created {
		fmt.Printf("Created personal config file: %s\n", path)
	}
	for _, path := range result.Overrode {
		fmt.Printf("Overrode personal config file: %s\n", path)
	}
	for _, path := range result.Skipped {
		fmt.Printf("Personal config file already exists: %s\n", path)
	}

	return nil
}

// showInitSuccessMessage displays a success message with next steps
func showInitSuccessMessage(team string, organization string) {
	fmt.Printf("\n%s %s\n\n", output.ActiveIcons.Success, output.SuccessBold("Configuration initialized successfully!"))

	fmt.Println("Next steps:")

	if team != "" {
		fmt.Printf("  %s Set '%s' as your default team\n", output.ActiveIcons.Bullet, team)
		if organization != "" {
			fmt.Printf("  %s Configuration is set up for team '%s' in organization '%s'\n",
				output.ActiveIcons.Bullet, team, organization)
		}
		fmt.Printf("  %s Create your first project with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip create project --name myproject --team "+team))
	} else {
		fmt.Printf("  %s Create a team configuration with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip config init --team myteam"))
		fmt.Printf("  %s Create your first project with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip create project --name myproject"))
	}

	fmt.Printf("  %s View your config with: %s\n",
		output.ActiveIcons.Bullet,
		output.Bold("vip config get"))

	fmt.Println()
}

// init sets up the configuration command and its subcommands
func init() {
	defaultConfigDir = config.GetDefaultConfigDir()
	defaultConfigFile = config.GetDefaultConfigFile()
	defaultTeamsDir = config.GetDefaultTeamsDir()

	// Add all subcommands to the config command
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(pathsCmd)

	// Define flags for the init command
	initCmd.Flags().StringP("team", "t", "", "Team name to scaffold configs for (optional)")
	initCmd.Flags().Bool("override", false, "Override existing config files if they exist")
	initCmd.Flags().StringP("organization", "o", "", "Organization name for team configs (optional)")
}
