// Package config provides centralised configuration management for viaplay-cli.
// This file contains functions for creating team configuration files from blueprints.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// CreationResult represents the result of a configuration creation operation
type CreationResult struct {
	Created  []string // Files that were created
	Overrode []string // Files that were overridden
	Skipped  []string // Files that were skipped (already exist)
}

// CreateTeamConfig creates the team configuration files and directories
func CreateTeamConfig(teamsDir, team string, override bool) (*CreationResult, error) {
	result := &CreationResult{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}

	teamDir := filepath.Join(teamsDir, team)
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		return nil, err
	}

	// Create envs/ and rulesets/ subfolders
	envsDir := filepath.Join(teamDir, "envs")
	rulesetsDir := filepath.Join(teamDir, "rulesets")
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

	// Create secrets.json at the team root
	secretsResult, err := CreateSecretsConfig(teamDir, override)
	if err != nil {
		return nil, err
	}
	result.Created = append(result.Created, secretsResult.Created...)
	result.Overrode = append(result.Overrode, secretsResult.Overrode...)
	result.Skipped = append(result.Skipped, secretsResult.Skipped...)

	// Print results
	for _, path := range result.Created {
		fmt.Printf("Created config file: %s\n", path)
	}
	for _, path := range result.Overrode {
		fmt.Printf("Overrode config file: %s\n", path)
	}
	for _, path := range result.Skipped {
		fmt.Printf("Config file already exists: %s\n", path)
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
		"staging.json": "staging",
		"prod.json":    "production",
	}

	for fname, envName := range exampleEnvs {
		fpath := filepath.Join(envsDir, fname)
		fileExists := true
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			fileExists = false
		}

		if !fileExists || override {
			// Use the embedded template with the appropriate environment name
			if err := CreateEnvironmentFromBlueprint(fpath, envName); err != nil {
				return nil, fmt.Errorf("failed to create environment from blueprint: %w", err)
			}

			if fileExists && override {
				result.Overrode = append(result.Overrode, fpath)
			} else {
				result.Created = append(result.Created, fpath)
			}
		} else {
			result.Skipped = append(result.Skipped, fpath)
		}
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
		"block-dev-branch.json":  "Block dev branch creation",
		"branch-protection.json": "Standard branch protection rules",
	}

	for fname, ruleName := range exampleRulesets {
		fpath := filepath.Join(rulesetsDir, fname)
		fileExists := true
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			fileExists = false
		}

		if !fileExists || override {
			// Use the embedded template
			if err := CreateRulesetFromBlueprint(fpath, ruleName); err != nil {
				return nil, fmt.Errorf("failed to create ruleset from blueprint: %w", err)
			}

			if fileExists && override {
				result.Overrode = append(result.Overrode, fpath)
			} else {
				result.Created = append(result.Created, fpath)
			}
		} else {
			result.Skipped = append(result.Skipped, fpath)
		}
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

	secretsPath := filepath.Join(teamDir, "secrets.json")
	secretsExists := true
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		secretsExists = false
	}

	if !secretsExists || override {
		if err := CreateSecretsFromBlueprint(secretsPath); err != nil {
			return nil, fmt.Errorf("failed to create secrets from blueprint: %w", err)
		}

		if secretsExists && override {
			result.Overrode = append(result.Overrode, secretsPath)
		} else {
			result.Created = append(result.Created, secretsPath)
		}
	} else {
		result.Skipped = append(result.Skipped, secretsPath)
	}

	return result, nil
}

// CreateEnvironmentFromBlueprint creates an environment configuration file from the embedded blueprint
func CreateEnvironmentFromBlueprint(destPath, envName string) error {
	// Read the environment blueprint
	blueprintData, err := GetBlueprintContent(EnvironmentBlueprintFile)
	if err != nil {
		return fmt.Errorf("failed to read environment blueprint: %w", err)
	}

	// Parse the blueprint
	tmpl, err := template.New("environment").Parse(string(blueprintData))
	if err != nil {
		return fmt.Errorf("failed to parse environment blueprint: %w", err)
	}

	// Create the destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create environment file: %w", err)
	}
	defer f.Close()

	// Execute the blueprint with the environment name
	err = tmpl.Execute(f, map[string]string{
		"Name": envName,
	})
	if err != nil {
		return fmt.Errorf("failed to execute environment blueprint: %w", err)
	}

	return nil
}

// CreateRulesetFromBlueprint creates a ruleset configuration file from the embedded blueprint
func CreateRulesetFromBlueprint(destPath, ruleName string) error {
	// Read the ruleset blueprint
	blueprintData, err := GetBlueprintContent(RulesetBlueprintFile)
	if err != nil {
		return fmt.Errorf("failed to read ruleset blueprint: %w", err)
	}

	// Parse the blueprint
	tmpl, err := template.New("ruleset").Parse(string(blueprintData))
	if err != nil {
		return fmt.Errorf("failed to parse ruleset blueprint: %w", err)
	}

	// Create the destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create ruleset file: %w", err)
	}
	defer f.Close()

	// Execute the blueprint with the ruleset name
	err = tmpl.Execute(f, map[string]string{
		"Name": ruleName,
	})
	if err != nil {
		return fmt.Errorf("failed to execute ruleset blueprint: %w", err)
	}

	return nil
}

// CreateSecretsFromBlueprint creates a secrets configuration file from the embedded blueprint
func CreateSecretsFromBlueprint(destPath string) error {
	// Read the secrets blueprint
	blueprintData, err := GetBlueprintContent(SecretsBlueprintFile)
	if err != nil {
		return fmt.Errorf("failed to read secrets blueprint: %w", err)
	}

	// Create the destination file
	err = os.WriteFile(destPath, blueprintData, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create secrets file: %w", err)
	}

	return nil
}
