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
	ConfigBlueprintFile      = "blueprints/config.yml"
	EnvironmentBlueprintFile = "blueprints/environment.json"
	RulesetBlueprintFile     = "blueprints/ruleset.json"
	SecretsBlueprintFile     = "blueprints/secrets.json"
)

// GetBlueprintContent returns the content of a blueprint file
func GetBlueprintContent(blueprintName string) ([]byte, error) {
	// Map short names to full paths
	blueprintPath := blueprintName
	switch blueprintName {
	case "config.yml":
		blueprintPath = ConfigBlueprintFile
	case "environment.json":
		blueprintPath = EnvironmentBlueprintFile
	case "ruleset.json":
		blueprintPath = RulesetBlueprintFile
	case "secrets.json":
		blueprintPath = SecretsBlueprintFile
	}

	return blueprintFiles.ReadFile(blueprintPath)
}
