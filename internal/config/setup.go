// Package config provides centralised configuration management for viaplay-cli.
// This file contains functions for setting up configuration files and directories.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

// FileOps represents the result of configuration file operations
type FileOps struct {
	Created  []string // Files that were created
	Overrode []string // Files that were overridden
	Skipped  []string // Files that were skipped (already exist)
}

// Merge combines another FileOps into this one
func (r *FileOps) Merge(other *FileOps) {
	if other == nil {
		return
	}
	r.Created = append(r.Created, other.Created...)
	r.Overrode = append(r.Overrode, other.Overrode...)
	r.Skipped = append(r.Skipped, other.Skipped...)
}

// PrintVerbose outputs the operation results using verbose logging
func (r *FileOps) PrintVerbose(prefix string) {
	for _, path := range r.Created {
		output.VerboseMessage(fmt.Sprintf("%s created: %s", prefix, path))
	}
	for _, path := range r.Overrode {
		output.VerboseMessage(fmt.Sprintf("%s overrode: %s", prefix, path))
	}
	for _, path := range r.Skipped {
		output.VerboseMessage(fmt.Sprintf("%s skipped: %s", prefix, path))
	}
}

// NewFileOps initializes an empty FileOps
func NewFileOps() *FileOps {
	return &FileOps{
		Created:  []string{},
		Overrode: []string{},
		Skipped:  []string{},
	}
}

// SetupDirs creates the standard config directory structure with environments,
// rulesets, and secrets configurations
func SetupDirs(baseDir string, isTeam, override bool) (*FileOps, error) {
	result := NewFileOps()

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

	// Create configuration files
	blueprintConfigs := []struct {
		blueprintFile string
		destPath      string
		description   string
	}{
		{EnvironmentBlueprintFile, filepath.Join(envsDir, "default.yaml"), "environment"},
		{RulesetBlueprintFile, filepath.Join(rulesetsDir, "block-dev-branch.yaml"), "ruleset"},
		{SecretsBlueprintFile, filepath.Join(baseDir, "secrets.yaml"), "secrets"},
	}

	for _, config := range blueprintConfigs {
		configResult, err := createFileFromBlueprint(config.blueprintFile, config.destPath, isTeam, override)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s config: %w", config.description, err)
		}
		result.Merge(configResult)
	}

	return result, nil
}

// SetupTeam creates the team configuration files and directories
func SetupTeam(org, team string, override bool) (*FileOps, error) {
	// teamDir := filepath.Join(teamsDir, team)
	teamDir := filepath.Join(GetDefaultConfigDir(), OrgsDirName, org, TeamsDirName, team)
	result, err := SetupDirs(teamDir, true, override)
	if err != nil {
		return nil, fmt.Errorf("failed to create team config structure: %w", err)
	}

	// Print results
	result.PrintVerbose("Team config")
	return result, nil
}

// SetupPersonal creates the personal configuration files and directories
func SetupPersonal(username string, override bool) (*FileOps, error) {
	// Get the personal directory path
	personalDir := filepath.Join(GetDefaultConfigDir(), PersonalDirName, username)

	// Use the common function to create the config structure
	result, err := SetupDirs(personalDir, false, override)
	if err != nil {
		return nil, fmt.Errorf("failed to create personal config structure: %w", err)
	}

	result.PrintVerbose("Personal config")
	return result, nil
}

// createFileFromBlueprint creates a file from a blueprint template, handling file existence checks
// and returning appropriate operation results
func createFileFromBlueprint(blueprintFile string, destPath string, isTeam, override bool) (*FileOps, error) { //nolint:gofumpt
	result := NewFileOps()

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

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory for %s: %w", destPath, err)
	}

	t, err := template.New("example").Delims("{{{", "}}}").Parse(string(blueprintData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", blueprintFile, err)
	}

	data := map[string]bool{
		"IsTeam": isTeam,
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", blueprintFile, err)
	}

	err = os.WriteFile(destPath, buf.Bytes(), 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", destPath, err)
	}

	// Update the result based on whether we created or overrode the file
	if fileExists {
		result.Overrode = append(result.Overrode, destPath)
	} else {
		result.Created = append(result.Created, destPath)
	}

	return result, nil
}
