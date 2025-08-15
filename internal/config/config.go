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
	AppName        = "viaplay-cli"
	ConfigDirName  = ".config/viaplay"
	CacheDirName   = ".cache/viaplay"
	TeamsDirName   = "teams"
	ConfigFileName = "config.yaml"
)

// Configuration stores the application configuration
type Configuration struct {
	// Basic paths
	ConfigDir  string
	ConfigFile string
	CacheDir   string
	TeamsDir   string

	// Default settings
	DefaultTeam     string
	DefaultAccount  string
	DefaultLanguage string
	DefaultType     string
	DefaultPrivate  bool
	IsOrg           bool

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

// CreateBasicConfig creates a basic configuration file with default settings
func CreateBasicConfig(configFilePath, username string) error {
	// Create the configuration directory if it doesn't exist
	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a viper instance for this configuration
	v := viper.New()

	// Set default values
	v.Set("default_team", "")
	v.Set("default_account", username)
	v.Set("is_org", false)
	v.Set("default_language", "go")
	v.Set("default_type", "service")
	v.Set("default_private", true)
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

		DefaultTeam:     viper.GetString("default_team"),
		DefaultAccount:  viper.GetString("default_account"),
		DefaultLanguage: viper.GetString("default_language"),
		DefaultType:     viper.GetString("default_type"),
		DefaultPrivate:  viper.GetBool("default_private"),
		IsOrg:           viper.GetBool("is_org"),

		// Load flag default values
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
	viper.SetDefault("default_private", true)

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
func writeBlueprintConfig(configFile, team, username string) error {
	// Always use the blueprint config as the starting point
	configBlueprint, err := GetBlueprintContent(ConfigBlueprintFile)
	if err != nil {
		return fmt.Errorf("failed to load blueprint config file: %w", err)
	}

	// Start with the blueprint content
	configContent := string(configBlueprint)

	// Replace values in the template string directly to preserve comments
	// Replace username placeholder if we have it
	if username != "" {
		configContent = strings.Replace(configContent,
			"default_account: \"\"",
			fmt.Sprintf("default_account: \"%s\"", username), 1)
	}

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
func InitializeConfigFile(configFile string, override bool, team, username string) error {
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
		return writeBlueprintConfig(configFile, team, username)
	}

	return nil
}
