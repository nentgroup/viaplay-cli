// Package config provides centralised configuration management for viaplay-cli.
// It handles loading, parsing, and managing configuration settings from files and environment variables
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/pkg/paths"
)

// Default paths and constants used throughout the application
const (
	AppName         = "viaplay-cli"
	ConfigDirName   = ".config/viaplay"
	CacheDirName    = ".cache/viaplay"
	TeamsDirName    = "teams"
	OrgsDirName     = "orgs"
	PersonalDirName = "users" // Directory for user-specific configs
	ConfigFileName  = "config.yaml"
)

// Configuration stores the application configuration
type Configuration struct {
	// Basic paths
	ConfigDir  string
	ConfigFile string
	CacheDir   string
	TeamsDir   string

	// Default settings
	DefaultTeam         string
	DefaultLanguage     string
	DefaultType         string
	DefaultVisibility   string // Repository visibility: "private" or "public"
	DefaultOrganization string // Default GitHub organization name

	// GitHub configuration
	GitHub struct {
		// No fields needed here anymore, but keeping the struct for backward compatibility
	}

	// Default flags for project creation
	ApplyEnvs      bool // Default for applying environments
	ApplySecrets   bool // Default for applying secrets
	ApplyRulesets  bool // Default for applying rulesets
	CleanupOnError bool // Default for cleanup on error
	NoHooks        bool // Default for skipping post-installation hooks
	NoRepo         bool // Default for skipping repository creation
	NoCache        bool // Default for disabling caching

	// Template mappings
	Templates map[string]map[string]string
}

// getHomeBasedPath returns a path based on the user's home directory
// or fallback to a relative path if home directory can't be determined
func getHomeBasedPath(subPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		// If home directory can't be determined, use current directory
		return subPath
	}
	return filepath.Join(home, subPath)
}

// GetDefaultConfigDir returns the default configuration directory
func GetDefaultConfigDir() string {
	return getHomeBasedPath(ConfigDirName)
}

// GetDefaultConfigFile returns the default configuration file path
func GetDefaultConfigFile() string {
	return filepath.Join(GetDefaultConfigDir(), ConfigFileName)
}

// GetDefaultCacheDir returns the default template cache directory
func GetDefaultCacheDir() string {
	return getHomeBasedPath(CacheDirName)
}

// GetDefaultTeamsDir returns the default teams directory
func GetDefaultTeamsDir() string {
	return filepath.Join(GetDefaultConfigDir(), TeamsDirName)
}

// LoadConfig loads the configuration from viper
func LoadConfig() (*Configuration, error) {
	config := &Configuration{
		ConfigDir:  viper.GetString("config_dir"),
		ConfigFile: viper.GetString("config_file"),
		CacheDir:   viper.GetString("cache_dir"),
		TeamsDir:   viper.GetString("teams_dir"),

		DefaultTeam:         viper.GetString("default_team"),
		DefaultLanguage:     viper.GetString("default_language"),
		DefaultType:         viper.GetString("default_type"),
		DefaultVisibility:   viper.GetString("default_visibility"),
		DefaultOrganization: viper.GetString("default_organization"),

		// Default flags for project creation
		ApplyEnvs:      viper.GetBool("apply_envs"),
		ApplySecrets:   viper.GetBool("apply_secrets"),
		ApplyRulesets:  viper.GetBool("apply_rulesets"),
		CleanupOnError: viper.GetBool("cleanup_on_error"),
		NoHooks:        viper.GetBool("no_hooks"),
		NoRepo:         viper.GetBool("no_repo"),
		NoCache:        viper.GetBool("no_cache"),

		Templates: make(map[string]map[string]string),
	}

	// Set default paths if not provided
	if config.ConfigDir == "" {
		config.ConfigDir = GetDefaultConfigDir()
	} else {
		config.ConfigDir = paths.Expand(config.ConfigDir)
	}

	if config.ConfigFile == "" {
		config.ConfigFile = GetDefaultConfigFile()
	} else {
		config.ConfigFile = paths.Expand(config.ConfigFile)
	}

	if config.CacheDir == "" {
		config.CacheDir = GetDefaultCacheDir()
	} else {
		config.CacheDir = paths.Expand(config.CacheDir)
	}

	if config.TeamsDir == "" {
		config.TeamsDir = GetDefaultTeamsDir()
	} else {
		config.TeamsDir = paths.Expand(config.TeamsDir)
	}

	// Load template mappings
	templatesMap := viper.GetStringMap("templates")
	for lang, types := range templatesMap {
		if typesMap, ok := types.(map[string]interface{}); ok {
			config.Templates[lang] = make(map[string]string)
			for typeName, source := range typesMap {
				if sourceStr, ok := source.(string); ok {
					config.Templates[lang][typeName] = sourceStr
				}
			}
		}
	}

	return config, nil
}

// InitConfig initialises viper configuration
func InitConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Search config in home directory
		viper.AddConfigPath(filepath.Join(home, ".config/viaplay"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Set defaults
	viper.SetDefault("config_dir", GetDefaultConfigDir())
	viper.SetDefault("config_file", GetDefaultConfigFile())
	viper.SetDefault("cache_dir", GetDefaultCacheDir())
	viper.SetDefault("teams_dir", GetDefaultTeamsDir())
	viper.SetDefault("default_language", "go")
	viper.SetDefault("default_type", "service")
	viper.SetDefault("default_visibility", "private")

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Config file was found but another error occurred
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is not a critical error
	}

	return nil
}

// refactored helper to reduce nesting
func writeBlueprintConfig(configFile, team, org string) error {
	// Always use the blueprint config as the starting point
	configBlueprint, err := GetBlueprintContent(ConfigBlueprintFile)
	if err != nil {
		return fmt.Errorf("failed to load blueprint config file: %w", err)
	}

	// Start with the blueprint content
	configContent := string(configBlueprint)

	// Set default_team from the team flag if provided
	if team != "" {
		configContent = strings.Replace(configContent,
			"default_team: \"\"",
			fmt.Sprintf("default_team: \"%s\"", team), 1)
	}

	if org != "" {
		// Set default_organization from the org flag if provided
		configContent = strings.Replace(configContent,
			"default_organization: \"\"",
			fmt.Sprintf("default_organization: \"%s\"", org), 1)
	}

	// Check if path-related settings already exist in the blueprint
	hasConfigDir := strings.Contains(configContent, "config_dir:")
	hasConfigFile := strings.Contains(configContent, "config_file:")
	hasTeamsDir := strings.Contains(configContent, "teams_dir:")

	// Only add path-related settings if they don't exist in the blueprint
	if !hasConfigDir || !hasConfigFile || !hasTeamsDir {
		pathConfig := "\n\n# Path configuration (added by viaplay-cli)"
		if !hasConfigDir {
			pathConfig += fmt.Sprintf("\nconfig_dir: %s", GetDefaultConfigDir())
		}
		if !hasConfigFile {
			pathConfig += fmt.Sprintf("\nconfig_file: %s", configFile)
		}
		if !hasTeamsDir {
			pathConfig += fmt.Sprintf("\nteams_dir: %s", GetDefaultTeamsDir())
		}
		configContent = fmt.Sprintf("%s%s", configContent, pathConfig)
	}

	// Write the modified template directly to the config file
	if err := os.WriteFile(configFile, []byte(configContent), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// After writing the config file, initialise viper to use it
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: Config file created but couldn't be loaded: %v\n", err)
	}
	return nil
}

// InitializeConfigFile creates or updates the main config file with values from the blueprint
func InitializeConfigFile(configFile string, override bool, team, org string) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Check if config already exists
	configExists := true
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configExists = false
	}

	if !configExists || override {
		return writeBlueprintConfig(configFile, team, org)
	}

	return nil
}

// GetOrganizationTeamsDir returns the teams directory for a specific organization
func (c *Configuration) GetOrganizationTeamsDir(orgName string) string {
	if orgName == "" {
		// If no organization is specified, return the default teams directory
		return c.TeamsDir
	}
	return filepath.Join(c.ConfigDir, OrgsDirName, orgName, TeamsDirName)
}

// GetPersonalDir returns the directory for personal account configuration
func (c *Configuration) GetPersonalDir(username string) string {
	if username == "" {
		return filepath.Join(c.ConfigDir, PersonalDirName)
	}
	return filepath.Join(c.ConfigDir, PersonalDirName, username)
}

// GetTeamDir returns the directory for a specific team, potentially within an organization
func (c *Configuration) GetTeamDir(team, orgName string) string {
	if orgName == "" {
		// If no organization is specified, use the default teams structure
		return filepath.Join(c.TeamsDir, team)
	}
	return filepath.Join(c.GetOrganizationTeamsDir(orgName), team)
}
