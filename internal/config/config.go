// Package config provides centralised configuration management for viaplay-cli.
// It handles loading, parsing, and managing configuration settings from files and environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Default paths and constants used throughout the application
const (
	AppName         = "viaplay-cli"
	ConfigDirName   = ".config/viaplay"
	CacheDirName    = ".cache/viaplay"
	TeamsDirName    = "teams"
	OrgsDirName     = "orgs"
	PersonalDirName = "users" // Directory for user-specific configs (user folders will be inside)
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
		DefaultAccountType string                 // "personal" or "organization"
		Organizations      map[string]OrgSettings // Map of organization names to their settings
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

// OrgSettings stores organization-specific settings
type OrgSettings struct {
	DefaultTeam      string            // Default team for this organization
	Teams            []string          // Teams within this organization
	DefaultTemplates map[string]string // Default templates for this organization
}

// CreateBasicConfig creates a basic configuration file with default settings
func CreateBasicConfig(configFilePath string) error {
	// Create the configuration directory if it doesn't exist
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a viper instance for this configuration
	v := viper.New()

	// Set default values
	v.Set("default_team", "")
	v.Set("default_organization", "")
	v.Set("github.default_account_type", "personal")
	v.Set("default_language", "go")
	v.Set("default_type", "service")
	v.Set("default_visibility", "private")
	v.Set("config_dir", GetDefaultConfigDir())
	v.Set("teams_dir", GetDefaultTeamsDir())
	v.Set("config_file", configFilePath)
	v.Set("cache_dir", GetDefaultCacheDir())

	// Set the config file path and format
	v.SetConfigFile(configFilePath)
	v.SetConfigType("yaml")

	// Write the configuration to file
	if err := v.WriteConfigAs(configFilePath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}

// GetDefaultConfigDir returns the default configuration directory
func GetDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// If home directory can't be determined, use current directory
		return ConfigDirName
	}
	return filepath.Join(home, ConfigDirName)
}

// GetDefaultConfigFile returns the default configuration file path
func GetDefaultConfigFile() string {
	return filepath.Join(GetDefaultConfigDir(), ConfigFileName)
}

// GetDefaultCacheDir returns the default template cache directory
func GetDefaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// If home directory can't be determined, use current directory
		return CacheDirName
	}
	return filepath.Join(home, CacheDirName)
}

// GetDefaultTeamsDir returns the default teams directory
func GetDefaultTeamsDir() string {
	return filepath.Join(GetDefaultConfigDir(), TeamsDirName)
}

// GetDefaultPersonalDir returns the default personal configuration directory
func GetDefaultPersonalDir() string {
	return filepath.Join(GetDefaultConfigDir(), PersonalDirName)
}

// ExpandPath expands the tilde in path to the user's home directory
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // Return original if home can't be determined
		}
		return filepath.Join(home, path[1:])
	}
	return path
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

		// GitHub configuration
		GitHub: struct {
			DefaultAccountType string
			Organizations      map[string]OrgSettings
		}{
			DefaultAccountType: viper.GetString("github.default_account_type"),
			Organizations:      make(map[string]OrgSettings),
		},

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
		config.ConfigDir = ExpandPath(config.ConfigDir)
	}

	if config.ConfigFile == "" {
		config.ConfigFile = GetDefaultConfigFile()
	} else {
		config.ConfigFile = ExpandPath(config.ConfigFile)
	}

	if config.CacheDir == "" {
		config.CacheDir = GetDefaultCacheDir()
	} else {
		config.CacheDir = ExpandPath(config.CacheDir)
	}

	if config.TeamsDir == "" {
		config.TeamsDir = GetDefaultTeamsDir()
	} else {
		config.TeamsDir = ExpandPath(config.TeamsDir)
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

	// Load organizations from config
	orgsMap := viper.GetStringMap("github.organizations")
	for orgName, settings := range orgsMap {
		if settingsMap, ok := settings.(map[string]interface{}); ok {
			org := OrgSettings{
				DefaultTemplates: make(map[string]string),
			}

			// Extract settings for this organization
			if defaultTeam, ok := settingsMap["default_team"].(string); ok {
				org.DefaultTeam = defaultTeam
			}

			// Extract teams list
			if teamsIface, ok := settingsMap["teams"].([]interface{}); ok {
				teams := make([]string, 0, len(teamsIface))
				for _, teamIface := range teamsIface {
					if team, ok := teamIface.(string); ok {
						teams = append(teams, team)
					}
				}
				org.Teams = teams
			}

			// Extract default templates
			if templatesIface, ok := settingsMap["default_templates"].(map[string]interface{}); ok {
				for key, valIface := range templatesIface {
					if val, ok := valIface.(string); ok {
						org.DefaultTemplates[key] = val
					}
				}
			}

			// Store the organization settings
			config.GitHub.Organizations[orgName] = org
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

// GetTemplateSource retrieves a template source based on language and type
func (c *Configuration) GetTemplateSource(language, projectType string) (string, error) {
	if langMap, exists := c.Templates[language]; exists {
		if source, exists := langMap[projectType]; exists {
			return source, nil
		}
	}

	return "", fmt.Errorf("no template source found for %s/%s", language, projectType)
}

// EnsureDirectoriesExist creates the necessary directories if they don't exist
func (c *Configuration) EnsureDirectoriesExist() error {
	dirs := []string{
		c.ConfigDir,
		c.CacheDir,
		c.TeamsDir,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// UpdateConfigValue updates a configuration value in the specified file while preserving comments
// and keeps Viper's in-memory state in sync with the file
func UpdateConfigValue(configFilePath, key, value string) error {
	// Ensure config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configFilePath)
	}

	// Read the existing config file
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the file line by line to find and update the key
	lines := strings.Split(string(content), "\n")
	keyFound := false
	keyPrefix := key + ":"

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "#") || trimmedLine == "" {
			continue
		}

		// Check if this line contains our key
		if strings.HasPrefix(trimmedLine, keyPrefix) {
			// Update the value while preserving indentation
			indentation := line[:len(line)-len(trimmedLine)]
			lines[i] = fmt.Sprintf("%s%s: %q", indentation, key, value)
			keyFound = true
			break
		}
	}

	// If key wasn't found, append it to the end of the file
	if !keyFound {
		lines = append(lines, fmt.Sprintf("%s: %q", key, value))
	}

	// Write the updated content back to the file
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(configFilePath, []byte(updatedContent), 0o600); err != nil {
		return fmt.Errorf("failed to write updated config file: %w", err)
	}

	// Update viper's in-memory state to match the file
	// This is important to keep using Viper for reading values
	viper.Set(key, value)

	return nil
}

// refactored helper to reduce nesting
func writeBlueprintConfig(configFile, team string) error {
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
func InitializeConfigFile(configFile string, override bool, team string) error {
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
		return writeBlueprintConfig(configFile, team)
	}

	return nil
}

// GetOrganizationTeamsDir returns the teams directory for a specific organization
func (c *Configuration) GetOrganizationTeamsDir(orgName string) string {
	if orgName == "" {
		// If no organization is specified, return the default teams directory
		return c.TeamsDir
	}

	// Create organization-specific teams directory path using the new structure:
	// ~/.config/viaplay/orgs/{org-name}/teams/
	return filepath.Join(c.ConfigDir, OrgsDirName, orgName, TeamsDirName)
}

// GetTeamDir returns the directory for a specific team, potentially within an organization
func (c *Configuration) GetTeamDir(team string, orgName string) string {
	if orgName == "" {
		// If no organization is specified, use the default teams structure
		return filepath.Join(c.TeamsDir, team)
	}

	// Use organization-specific team directory with the new structure:
	// ~/.config/viaplay/orgs/{org-name}/teams/{team-name}/
	return filepath.Join(c.GetOrganizationTeamsDir(orgName), team)
}

// HasOrganization checks if an organization is defined in the configuration
func (c *Configuration) HasOrganization(orgName string) bool {
	if orgName == "" {
		return false
	}
	_, exists := c.GitHub.Organizations[orgName]
	return exists
}

// GetOrganizationSettings retrieves settings for a specific organization
func (c *Configuration) GetOrganizationSettings(orgName string) (OrgSettings, bool) {
	if orgName == "" {
		return OrgSettings{}, false
	}

	settings, exists := c.GitHub.Organizations[orgName]
	return settings, exists
}

// GetDefaultTeamForOrg returns the default team for a specific organization
func (c *Configuration) GetDefaultTeamForOrg(orgName string) string {
	// If no organization specified, use global default
	if orgName == "" {
		return c.DefaultTeam
	}

	// Check if the organization exists and has a default team
	if settings, exists := c.GitHub.Organizations[orgName]; exists && settings.DefaultTeam != "" {
		return settings.DefaultTeam
	}

	// Fall back to global default if organization has no default team
	return c.DefaultTeam
}

// EnsureOrganizationDirectories creates organization-specific directories if they don't exist
func (c *Configuration) EnsureOrganizationDirectories(orgName string) error {
	if orgName == "" {
		return nil // Nothing to do if no organization specified
	}

	// Create main organization directory
	orgDir := c.GetOrganizationTeamsDir(orgName)
	if err := os.MkdirAll(orgDir, 0o755); err != nil {
		return fmt.Errorf("failed to create organization directory %s: %w", orgDir, err)
	}

	// If the organization has defined teams, create directories for each team
	if settings, exists := c.GitHub.Organizations[orgName]; exists {
		for _, team := range settings.Teams {
			teamDir := filepath.Join(orgDir, team)
			if err := os.MkdirAll(teamDir, 0o755); err != nil {
				return fmt.Errorf("failed to create team directory %s: %w", teamDir, err)
			}
		}
	}

	return nil
}

// GetPersonalDir returns the directory for personal account configuration
func (c *Configuration) GetPersonalDir(username string) string {
	if username == "" {
		return filepath.Join(c.ConfigDir, PersonalDirName)
	}
	return filepath.Join(c.ConfigDir, PersonalDirName, username)
}

// EnsurePersonalDirectories creates the personal account directory if it doesn't exist
func (c *Configuration) EnsurePersonalDirectories() error {
	// Create the personal directory
	personalDir := c.GetPersonalDir("")
	if err := os.MkdirAll(personalDir, 0o755); err != nil {
		return fmt.Errorf("failed to create personal directory %s: %w", personalDir, err)
	}
	return nil
}

// CreatePersonalConfig creates the personal configuration files and directories
func CreatePersonalConfig(username string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	// Get the personal directory path
	personalDir := filepath.Join(GetDefaultConfigDir(), PersonalDirName, username)
	if err := os.MkdirAll(personalDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create personal directory: %w", err)
	}

	// Create envs/ and rulesets/ subfolders
	envsDir := filepath.Join(personalDir, "envs")
	rulesetsDir := filepath.Join(personalDir, "rulesets")
	if err := os.MkdirAll(envsDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(rulesetsDir, 0o755); err != nil {
		return nil, err
	}

	// Create environment configs
	envResult, err := CreateEnvironmentConfigs(envsDir, override)
	if err != nil {
		return nil, err
	}
	result.Created = append(result.Created, envResult.Created...)
	result.Overrode = append(result.Overrode, envResult.Overrode...)
	result.Skipped = append(result.Skipped, envResult.Skipped...)

	// Create ruleset configs
	rulesetResult, err := CreateRulesetConfigs(rulesetsDir, override)
	if err != nil {
		return nil, err
	}
	result.Created = append(result.Created, rulesetResult.Created...)
	result.Overrode = append(result.Overrode, rulesetResult.Overrode...)
	result.Skipped = append(result.Skipped, rulesetResult.Skipped...)

	// Create secrets.json at the personal root
	secretsResult, err := CreateSecretsConfig(personalDir, override)
	if err != nil {
		return nil, err
	}
	result.Created = append(result.Created, secretsResult.Created...)
	result.Overrode = append(result.Overrode, secretsResult.Overrode...)
	result.Skipped = append(result.Skipped, secretsResult.Skipped...)

	return result, nil
}
