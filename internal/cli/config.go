// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		override, err := cmd.Flags().GetBool("override")
		if err != nil {
			return fmt.Errorf("failed to get 'override' flag: %w", err)
		}

		team, err := cmd.Flags().GetString("team")
		if err != nil {
			return fmt.Errorf("failed to get 'team' flag: %w", err)
		}

		organization, err := cmd.Flags().GetString("organization")
		if err != nil {
			return fmt.Errorf("failed to get 'org' flag: %w", err)
		}

		// Create GitHub client
		ghClient := gh.NewGitHubClient(token)

		// Get authenticated user info
		username, err := ghClient.GetAuthenticatedUser()
		if err != nil {
			return fmt.Errorf("failed to get authenticated user: %w", err)
		}
		fmt.Printf("Authenticated as: %s\n", output.Bold(username))

		// If organization flag is specified but not a specific org, list available organizations
		if cmd.Flags().Changed("organization") && organization == "" {
			orgs, err := ghClient.GetUserOrganizations()
			if err != nil {
				fmt.Printf("Warning: Failed to fetch organizations: %v\n", err)
			} else if len(orgs) > 0 {
				fmt.Println("\nOrganizations you belong to:")
				for i, org := range orgs {
					fmt.Printf("  %d. %s\n", i+1, *org.Login)
				}
				fmt.Println("\nSpecify an organization with --org flag")
			} else {
				fmt.Println("You don't belong to any organizations.")
			}

			// Exit early if the user just wanted to see their organizations
			return nil
		}

		// Interactive organisation and team selection if neither team nor organisation flags are set
		if !cmd.Flags().Changed("team") && !cmd.Flags().Changed("organization") {
			// Interactive selection of organization
			selectedOrg, err := SelectOrganizationWithBubbles(ghClient)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				// Continue without organization if there's an error
			} else if selectedOrg != "" {
				organization = selectedOrg
				fmt.Printf("Selected organization: %s\n", output.Bold(organization))

				// If we have an organization, also select a team
				selectedTeam, err := SelectTeamWithBubbles(ghClient, organization)
				if err != nil {
					fmt.Printf("Warning: %v\n", err)
					// Continue without team if there's an error
				} else if selectedTeam != "" {
					team = selectedTeam
					fmt.Printf("Selected team: %s\n", output.Bold(team))
				}
			}
		}

		// Initialise main config file with authenticated user information
		if err := initializeConfigFile(override, team, organization, ghClient); err != nil {
			return err
		}

		// Always set up personal configurations for the authenticated user
		fmt.Println("\nSetting up personal account configurations...")
		if err := scaffoldPersonalConfig(username, override); err != nil {
			return err
		}

		// Set up team configurations if specified
		if team != "" {
			fmt.Printf("\nSetting up team configurations for '%s'", team)
			if organization != "" {
				fmt.Printf(" in organization '%s'", organization)
			}
			fmt.Println("...")

			if err := scaffoldTeamConfig(team, override, organization); err != nil {
				return err
			}
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
	initCmd.Flags().StringP("team", "t", "", "Team name to scaffold configs for (optional)")
	initCmd.Flags().Bool("override", false, "Override existing config files if they exist")
	initCmd.Flags().StringP("organization", "o", "", "Organization name for team configs (optional)")
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
func initializeConfigFile(override bool, team string, organization string, ghClient *gh.GitHubClient) error {
	// Get the authenticated user's GitHub account type (personal or organization)
	accountType, err := ghClient.GetAccountType()
	if err != nil {
		return fmt.Errorf("failed to determine GitHub account type: %w", err)
	}

	// Use the centralised function from the config package
	if err := config.InitializeConfigFile(defaultConfigFile, override, team); err != nil {
		return err
	}

	// If this is an organization account, set the organization field
	if accountType == "Organization" && organization != "" {
		if err := config.UpdateConfigValue(defaultConfigFile, "github.organization", organization); err != nil {
			fmt.Printf("Warning: Failed to update GitHub organization in config: %v\n", err)
		} else {
			fmt.Printf("Set GitHub organization to '%s' based on your logged-in account\n", organization)
		}
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
func scaffoldTeamConfig(team string, override bool, organization string) error {
	// Determine the appropriate base directory based on whether an organization is specified
	var baseDir string

	if organization != "" {
		// For organization teams, use the organization-specific directory
		configDir := filepath.Dir(defaultTeamsDir) // Get the .config/viaplay directory
		baseDir = filepath.Join(configDir, config.OrgsDirName, organization, config.TeamsDirName)

		// Ensure the organization directory exists
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			return fmt.Errorf("failed to create organization team directory: %w", err)
		}

		fmt.Printf("Using organization-specific path for team '%s' in organization '%s'\n", team, organization)

		// Update the default_organization in config file if not already set
		if viper.GetString("default_organization") == "" {
			if err := config.UpdateConfigValue(defaultConfigFile, "default_organization", organization); err != nil {
				fmt.Printf("Warning: Failed to update default_organization in config: %v\n", err)
			} else {
				fmt.Printf("Set default_organization = %s\n", organization)
			}
		}

		// Update or create the organization entry in github.organizations
		updateOrganizationConfig(organization, team)
	} else {
		// For regular teams, use the default teams directory
		baseDir = defaultTeamsDir
	}

	// Use the config package's team creation functionality
	result, err := config.CreateTeamConfig(baseDir, team, override)
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

// updateOrganizationConfig updates or creates an organization entry in the github.organizations section
func updateOrganizationConfig(organization, team string) {
	// Read the current config file content
	content, err := os.ReadFile(defaultConfigFile)
	if err != nil {
		fmt.Printf("Warning: Could not read config file to update organization: %v\n", err)
		return
	}

	// Check if the organization already exists in the config
	orgKey := fmt.Sprintf("github.organizations.%s", organization)
	orgExists := viper.IsSet(orgKey)

	if orgExists {
		// If the organization exists, check if the team is already in its teams list
		teams := viper.GetStringSlice(fmt.Sprintf("%s.teams", orgKey))
		teamExists := false
		for _, t := range teams {
			if t == team {
				teamExists = true
				break
			}
		}

		// If the team doesn't exist in the organization's teams list, add it
		if !teamExists {
			teams = append(teams, team)
			// The Set() method doesn't return anything, it was incorrectly treated as returning an error
			viper.Set(fmt.Sprintf("%s.teams", orgKey), teams)

			// Try to save the updated config
			if err := viper.WriteConfig(); err != nil {
				fmt.Printf("Warning: Failed to write updated config: %v\n", err)
			} else {
				fmt.Printf("Added team '%s' to organization '%s' in configuration\n", team, organization)
			}
		}
	} else {
		// If the organization doesn't exist, we need to add it with proper YAML structure
		// This is a bit tricky with viper, so we'll update the file manually with proper indentation

		// Find where to insert the organization config
		lines := strings.Split(string(content), "\n")
		orgSectionFound := false
		insertIndex := -1
		indentLevel := "  " // Default indent level for YAML

		// Look for the github.organizations section
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "organizations:") {
				orgSectionFound = true
				insertIndex = i + 1

				// Calculate the indent level based on the current line
				indent := line[:len(line)-len(trimmedLine)]
				indentLevel = indent + "  " // Add two spaces for child elements
				break
			}
		}

		// If we didn't find the organizations section, look for the github: section
		if !orgSectionFound {
			for i, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine == "github:" {
					// Insert at the end of the github section
					// First, find the end of the github section
					j := i + 1
					for ; j < len(lines); j++ {
						// If we find a line that's not indented more than the github line,
						// we've reached the end of the github section
						if !strings.HasPrefix(lines[j], "  ") && lines[j] != "" {
							break
						}
					}
					insertIndex = j
					indentLevel = "  " // Standard indent for sections under github
					break
				}
			}
		}

		// If we still haven't found where to insert, add to the end of the file
		if insertIndex == -1 {
			lines = append(lines, "", "github:", "  organizations:")
			insertIndex = len(lines)
			indentLevel = "    " // Indent for items under organizations
		}

		// Create the new organization config
		var orgConfig []string

		if orgSectionFound {
			// If the organizations section exists, add the new org directly under it
			orgConfig = []string{
				fmt.Sprintf("%s%s:", indentLevel, organization),
				fmt.Sprintf("%s  default_team: \"%s\"", indentLevel, team),
				fmt.Sprintf("%s  teams:", indentLevel),
				fmt.Sprintf("%s    - \"%s\"", indentLevel, team),
			}
		} else {
			// If we're adding to the github section, we need to include the organizations key
			orgConfig = []string{
				"  organizations:",
				fmt.Sprintf("    %s:", organization),
				fmt.Sprintf("      default_team: \"%s\"", team),
				fmt.Sprintf("      teams:"),
				fmt.Sprintf("        - \"%s\"", team),
			}
		}

		// Insert the organization config at the determined position
		if insertIndex >= len(lines) {
			lines = append(lines, orgConfig...)
		} else {
			// Insert the new lines at the insertIndex
			newLines := make([]string, 0, len(lines)+len(orgConfig))
			newLines = append(newLines, lines[:insertIndex]...)
			newLines = append(newLines, orgConfig...)
			newLines = append(newLines, lines[insertIndex:]...)
			lines = newLines
		}

		// Write the updated config back to file
		if err := os.WriteFile(defaultConfigFile, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
			fmt.Printf("Warning: Failed to write updated config file: %v\n", err)
		} else {
			fmt.Printf("Added organization '%s' with team '%s' to configuration\n", organization, team)

			// Reload viper config to reflect changes
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("Warning: Failed to reload config after updates: %v\n", err)
			}
		}
	}
}

// Message display functions

// showInitSuccessMessage displays a success message after initialization
func showInitSuccessMessage(team string, organization string) {
	fmt.Println("\n✅ viaplay-cli initialized successfully!")
	fmt.Println("- Configuration directory: ", defaultConfigDir)
	fmt.Println("- Teams directory:        ", defaultTeamsDir)

	// Provide hints for next steps
	if team == "" {
		fmt.Println("\nTip: Set up a team configuration with:")
		fmt.Println("  vip config init --team <team-name>")
	}
}

// scaffoldPersonalConfig creates personal account configuration files and directories
func scaffoldPersonalConfig(username string, override bool) error {
	fmt.Printf("Setting up personal account configurations for '%s'\n", username)

	// Use the config package's personal config creation functionality
	result, err := config.CreatePersonalConfig(username, override)
	if err != nil {
		return fmt.Errorf("failed to create personal configuration: %w", err)
	}

	// Print a summary of what happened
	if len(result.Created) > 0 {
		fmt.Printf("\nCreated %d config files for personal account\n", len(result.Created))
	}
	if len(result.Overrode) > 0 {
		fmt.Printf("Overrode %d existing config files for personal account\n", len(result.Overrode))
	}
	if len(result.Skipped) > 0 {
		fmt.Printf("Skipped %d existing config files for personal account\n", len(result.Skipped))
	}

	return nil
}
