// Package config provides centralised configuration management for viaplay-cli.
// This file contains functions for creating team configuration files from blueprints.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

// CreationResult represents the result of a configuration creation operation
type CreationResult struct {
	Created  []string // Files that were created
	Overrode []string // Files that were overridden
	Skipped  []string // Files that were skipped (already exist)
}

// CreateConfigStructure creates a standard configuration directory structure with
// environments, rulesets, and secrets configurations
func CreateConfigStructure(baseDir string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", baseDir, err)
	}

	// Create envs/ and rulesets/ subfolders
	envsDir := filepath.Join(baseDir, "envs")
	rulesetsDir := filepath.Join(baseDir, "rulesets")
	if err := os.MkdirAll(envsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create envs directory: %w", err)
	}
	if err := os.MkdirAll(rulesetsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create rulesets directory: %w", err)
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

	// Create secrets.json at the root
	secretsResult, err := CreateSecretsConfig(baseDir, override)
	if err != nil {
		return nil, err
	}
	result.Created = append(result.Created, secretsResult.Created...)
	result.Overrode = append(result.Overrode, secretsResult.Overrode...)
	result.Skipped = append(result.Skipped, secretsResult.Skipped...)

	return result, nil
}

// CreateTeamConfig creates the team configuration files and directories
func CreateTeamConfig(teamsDir, team string, override bool) (*CreationResult, error) {
	teamDir := filepath.Join(teamsDir, team)
	result, err := CreateConfigStructure(teamDir, override)
	if err != nil {
		return nil, fmt.Errorf("failed to create team config structure: %w", err)
	}

	// Print results
	for _, path := range result.Created {
		output.VerboseMessage(fmt.Sprintf("Created config file %s", path))
	}
	for _, path := range result.Overrode {
		output.VerboseMessage(fmt.Sprintf("Overrode config file: %s\n", path))
	}
	for _, path := range result.Skipped {
		output.VerboseMessage(fmt.Sprintf("Skipped config file: %s\n", path))
	}

	return result, nil
}

// CreateEnvironmentConfigs creates environment config files
func CreateEnvironmentConfigs(envsDir string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	exampleEnvs := map[string]string{
		"staging.yaml": "staging",
		"prod.yaml":    "production",
	}

	for fname, envName := range exampleEnvs {
		fpath := filepath.Join(envsDir, fname)

		// Create the environment blueprint
		envResult, err := CreateEnvironmentFromBlueprint(fpath, envName, override)
		if err != nil {
			return nil, fmt.Errorf("failed to create environment from blueprint: %w", err)
		}

		// Merge results
		result.Created = append(result.Created, envResult.Created...)
		result.Overrode = append(result.Overrode, envResult.Overrode...)
		result.Skipped = append(result.Skipped, envResult.Skipped...)
	}

	return result, nil
}

// CreateRulesetConfigs creates ruleset config files
func CreateRulesetConfigs(rulesetsDir string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	exampleRulesets := map[string]string{
		"block-dev-branch.yaml": "Block dev branch creation",
	}

	for fname, ruleName := range exampleRulesets {
		fpath := filepath.Join(rulesetsDir, fname)

		// Create the ruleset blueprint
		rulesetResult, err := CreateRulesetFromBlueprint(fpath, ruleName, override)
		if err != nil {
			return nil, fmt.Errorf("failed to create ruleset from blueprint: %w", err)
		}

		// Merge results
		result.Created = append(result.Created, rulesetResult.Created...)
		result.Overrode = append(result.Overrode, rulesetResult.Overrode...)
		result.Skipped = append(result.Skipped, rulesetResult.Skipped...)
	}

	return result, nil
}

// CreateSecretsConfig creates a secrets config file
func CreateSecretsConfig(teamDir string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	secretsPath := filepath.Join(teamDir, "secrets.yaml")

	// Create the secrets blueprint
	secretsResult, err := CreateSecretsFromBlueprint(secretsPath, override)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets from blueprint: %w", err)
	}

	// Merge results
	result.Created = append(result.Created, secretsResult.Created...)
	result.Overrode = append(result.Overrode, secretsResult.Overrode...)
	result.Skipped = append(result.Skipped, secretsResult.Skipped...)

	return result, nil
}

// handleBlueprintCreation is a helper function that handles the common pattern of
// creating a file from a blueprint and processing the result
func handleBlueprintCreation(blueprintFile string, destPath string, templateData interface{}, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	// Check if the file exists first
	fileExists := true
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		fileExists = false
	}

	// Skip if file exists and override is false
	if fileExists && !override {
		result.Skipped = append(result.Skipped, destPath)
		return result, nil
	}

	// Read the blueprint content
	blueprintData, err := GetBlueprintContent(blueprintFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint %s: %w", blueprintFile, err)
	}

	// If no template data is provided, just write the file directly
	if templateData == nil {
		err = os.WriteFile(destPath, blueprintData, 0o600)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", destPath, err)
		}
	} else {
		// Parse the blueprint as a template
		templateName := filepath.Base(blueprintFile)
		tmpl, err := template.New(templateName).Parse(string(blueprintData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse blueprint %s: %w", blueprintFile, err)
		}

		// Create the destination file
		f, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", destPath, err)
		}
		defer f.Close()

		// Execute the blueprint with the provided data
		err = tmpl.Execute(f, templateData)
		if err != nil {
			return nil, fmt.Errorf("failed to execute blueprint %s: %w", blueprintFile, err)
		}
	}

	// Update the result based on whether we created or overrode the file
	if fileExists {
		result.Overrode = append(result.Overrode, destPath)
	} else {
		result.Created = append(result.Created, destPath)
	}

	return result, nil
}

// CreateEnvironmentFromBlueprint creates an environment configuration file from the embedded blueprint
func CreateEnvironmentFromBlueprint(destPath, envName string, override bool) (*CreationResult, error) {
	return handleBlueprintCreation(EnvironmentBlueprintFile, destPath, map[string]string{
		"Name": envName,
	}, override)
}

// CreateRulesetFromBlueprint creates a ruleset configuration file from the embedded blueprint
func CreateRulesetFromBlueprint(destPath, ruleName string, override bool) (*CreationResult, error) {
	return handleBlueprintCreation(RulesetBlueprintFile, destPath, map[string]string{
		"Name": ruleName,
	}, override)
}

// CreateSecretsFromBlueprint creates a secrets configuration file from the embedded blueprint
func CreateSecretsFromBlueprint(destPath string, override bool) (*CreationResult, error) {
	return handleBlueprintCreation(SecretsBlueprintFile, destPath, nil, override)
}
