// Package config provides centralised configuration management for viaplay-cli.
// This file contains embedded blueprint files used for scaffolding configurations.
package config

import (
	"embed"
)

//go:embed blueprints
var blueprintFiles embed.FS

// Blueprint file names
const (
	ConfigBlueprintFile      = "blueprints/config.yaml" // Changed from config.yml to config.yaml
	EnvironmentBlueprintFile = "blueprints/environment.yaml"
	RulesetBlueprintFile     = "blueprints/ruleset.yaml"
	SecretsBlueprintFile     = "blueprints/secrets.yaml"
	TeamConfigBlueprintFile  = "blueprints/team-config.yaml"
)

// GetBlueprintContent returns the content of a blueprint file
func GetBlueprintContent(blueprintName string) ([]byte, error) {
	// Map short names to full paths
	blueprintPath := blueprintName
	switch blueprintName {
	case "config.yml", "config.yaml": // Accept both extensions
		blueprintPath = ConfigBlueprintFile
	case "environment.yaml":
		blueprintPath = EnvironmentBlueprintFile
	case "ruleset.yaml":
		blueprintPath = RulesetBlueprintFile
	case "secrets.yaml":
		blueprintPath = SecretsBlueprintFile
	case "team-config.yaml":
		blueprintPath = TeamConfigBlueprintFile
	}

	return blueprintFiles.ReadFile(blueprintPath)
}
