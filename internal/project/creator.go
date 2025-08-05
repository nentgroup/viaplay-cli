// Package project provides project creation and management functionality for viaplay-cli.
// It handles the scaffolding, configuration, and setup of new projects.
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/zalando/go-keyring"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/template"
)

// ProjectSummary contains details about the created project to be displayed to the user
type ProjectSummary struct {
	ProjectPath     string   // Full path to the project location
	RepoURL         string   // GitHub repository URL
	Language        string   // Programming language used
	ProjectType     string   // Type of project (service, CLI, etc.)
	Team            string   // Team assigned to the project
	AppliedEnvs     bool     // Whether environments were applied
	AppliedRulesets bool     // Whether rulesets were applied
	AppliedSecrets  bool     // Whether secrets were applied
	CustomSecrets   bool     // Whether custom secrets were applied
	Errors          []string // Any non-fatal errors that occurred
}

// Creator manages the project creation workflow
type Creator struct {
	// GitHub client for repository operations
	GitHubClient *gh.GitHubClient

	// Configuration
	Config *config.Configuration

	// Template registry for looking up templates
	TemplateRegistry *registry.Registry

	// Cache manager for template caching
	CacheManager *cache.Manager

	// Project scaffolder for applying templates
	Scaffolder *scaffolding.ProjectScaffolder
}

// CreateOptions contains all options for creating a new project
type CreateOptions struct {
	// Repository options
	RepoName        string
	RepoDescription string
	RepoOwner       string
	IsPrivate       bool
	IsOrg           bool
	SkipRepo        bool // Skip GitHub repository creation

	// Project options
	Language    string
	ProjectType string
	Team        string
	BinaryName  string // Name for compiled binary (for Go, Rust, etc.)
	SkipHooks   bool   // Skip running post-installation hooks

	// Configuration options
	ConfigDir     string
	ApplyEnvs     bool
	ApplyRulesets bool
	ApplySecrets  bool
	RepoSecrets   string // JSON string of repo-specific secrets

	// Template options
	TemplateSource string
	Scaffold       bool   // Wether to scaffold the project, always true for project creation
	OutputDir      string // Local directory for the project (if Scaffold is true)
}

// NewCreator creates a new project creator with the given GitHub client and config directory
func NewCreator(ghClient *gh.GitHubClient, configDir string) *Creator {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use default configuration if loading fails
		cfg = &config.Configuration{
			ConfigDir: config.GetDefaultConfigDir(),
			CacheDir:  config.GetDefaultCacheDir(),
			TeamsDir:  config.GetDefaultTeamsDir(),
		}
	}

	// Create cache manager
	cacheManager := cache.NewManager(cfg)

	// Create template registry
	templateRegistry := registry.NewRegistry(cfg)
	err = templateRegistry.LoadTemplates()
	if err != nil {
		fmt.Printf("Warning: failed to load templates: %v\n", err)
	}

	// Create scaffolder
	scaffolder := scaffolding.NewProjectScaffolder(cacheManager, cfg)

	return &Creator{
		GitHubClient:     ghClient,
		Config:           cfg,
		TemplateRegistry: templateRegistry,
		CacheManager:     cacheManager,
		Scaffolder:       scaffolder,
	}
}

// Helper to convert CreateOptions to *template.Variables
func createOptionsToTemplateVariables(opts CreateOptions) *template.Variables {
	vars := template.NewTemplateVariables()

	// Basic project information
	vars.ProjectName = opts.RepoName
	vars.ProjectDescription = opts.RepoDescription

	// Repository information
	vars.RepoOwner = opts.RepoOwner
	vars.RepoName = opts.RepoName
	vars.IsPrivate = opts.IsPrivate
	vars.RepoURL = fmt.Sprintf("https://github.com/%s/%s", opts.RepoOwner, opts.RepoName)
	vars.RepoSSHURL = fmt.Sprintf("git@github.com:%s/%s.git", opts.RepoOwner, opts.RepoName)

	// Project language and type
	vars.Language = opts.Language
	vars.ProjectType = opts.ProjectType
	vars.Team = opts.Team

	// Additional values
	vars.CreatedAt = time.Now()
	vars.Year = time.Now().Year()

	// Service information
	vars.ServiceName = opts.RepoName
	vars.ServiceOwner = opts.Team
	vars.ServiceOwnerKey = strings.ToLower(strings.ReplaceAll(opts.Team, " ", "-"))

	// Handle binary name for compiled languages (Go, Rust, etc.)
	binaryName := opts.RepoName
	if opts.BinaryName != "" {
		// Use the custom binary name if provided
		binaryName = opts.BinaryName
	}

	// Format BinaryName: remove spaces and special characters, convert to lowercase
	binaryName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1 // Drop the character
	}, binaryName)
	// Convert to lowercase
	binaryName = strings.ToLower(binaryName)

	// Set binary name for compiled languages
	if opts.Language == "go" || opts.Language == "rust" {
		vars.BinaryName = binaryName
	}

	// Go-specific variables
	if opts.Language == "go" {
		vars.ModulePath = fmt.Sprintf("github.com/%s/%s", opts.RepoOwner, opts.RepoName)
	}

	// Docker variables
	vars.DockerImageName = strings.ToLower(opts.RepoName)
	return vars
}

// CreateProject creates a new project based on the provided options
func (c *Creator) CreateProject(opts CreateOptions) error {
	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		// Use current directory if not specified
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Determine template source
	templateSource := opts.TemplateSource
	if templateSource == "" {
		// Use the default template for the specified language and project type
		template, err := c.TemplateRegistry.GetTemplate(opts.Language, opts.ProjectType)
		if err != nil {
			return fmt.Errorf("failed to find template for %s/%s: %w", opts.Language, opts.ProjectType, err)
		}
		templateSource = template.Source
	}

	// Convert options to template variables
	templateVars := createOptionsToTemplateVariables(opts)
	if err := c.Scaffolder.ScaffoldProjectWithOptions(outputDir, opts.Language, opts.ProjectType, templateSource, templateVars, opts.SkipHooks); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// Create handles the full project creation workflow
func (c *Creator) Create(opts CreateOptions) (*ProjectSummary, error) {
	output.VerboseMessage(fmt.Sprintf("Starting project creation with options: %+v", opts))

	// Initialize project summary
	summary := &ProjectSummary{
		Language:        opts.Language,
		ProjectType:     opts.ProjectType,
		Team:            opts.Team,
		AppliedEnvs:     opts.ApplyEnvs,
		AppliedRulesets: opts.ApplyRulesets,
		AppliedSecrets:  opts.ApplySecrets,
		CustomSecrets:   opts.RepoSecrets != "",
		Errors:          []string{},
	}

	// Get authenticated user for CreatedBy field
	username, err := c.GitHubClient.GetAuthenticatedUser()
	if err != nil {
		output.VerboseMessage(fmt.Sprintf("Failed to get authenticated username: %v", err))
		summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to get authenticated username: %v", err))
	}

	// Convert options to template variables with additional info
	templateVars := createOptionsToTemplateVariables(opts)

	// Set authenticated username if available
	if username != "" {
		templateVars.CreatedBy = username
		output.VerboseMessage(fmt.Sprintf("Setting CreatedBy to authenticated user: %s", username))
	}

	// Determine the project path
	projectPath := opts.OutputDir
	if projectPath == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = filepath.Join(currentDir, opts.RepoName)
	} else {
		projectPath = filepath.Join(projectPath, opts.RepoName)
	}
	summary.ProjectPath = projectPath

	if opts.Scaffold {
		output.VerboseMessage("Scaffolding project locally...")
		fmt.Printf("Scaffolding project...")
		if err := c.scaffoldProjectWithVariables(opts, templateVars); err != nil {
			output.VerboseMessage(fmt.Sprintf("Project scaffolding failed: %v", err))
			fmt.Println(" failed")
			return nil, fmt.Errorf("failed to scaffold project: %w", err)
		}
		output.VerboseMessage("Project scaffolding complete.")
		fmt.Println(" done")
	}

	if opts.SkipRepo {
		output.VerboseMessage("Skipping GitHub repository creation as per options.")
	} else {
		output.VerboseMessage("Creating repository on GitHub...")
		fmt.Printf("Creating repository...")
		_, err = c.createRepository(opts)
		if err != nil {
			output.VerboseMessage(fmt.Sprintf("Repository creation error: %v", err))
			if !strings.Contains(err.Error(), "name already exists on this account") {
				fmt.Println(" failed")
				return nil, fmt.Errorf("failed to create repository: %w", err)
			}
			fmt.Println(" already exists, proceeding")
		} else {
			output.VerboseMessage("Repository created successfully.")
			fmt.Println(" done")
		}

		// Set the repository URL in the summary
		summary.RepoURL = fmt.Sprintf("https://github.com/%s/%s", opts.RepoOwner, opts.RepoName)
	}

	output.VerboseMessage("Applying GitHub configurations (envs, rulesets, secrets)...")
	fmt.Printf("Applying GitHub configurations...")
	if err := c.applyGitHubConfigurations(opts); err != nil {
		output.VerboseMessage(fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
		fmt.Println(" failed")
		summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
	}
	output.VerboseMessage("GitHub configurations applied.")
	fmt.Println(" done")

	// Run post-installation hooks only if they're not skipped and we've scaffolded locally
	if opts.Scaffold && !opts.SkipHooks {
		outputDir := opts.OutputDir
		if outputDir == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				output.VerboseMessage(fmt.Sprintf("Failed to get current directory: %v", err))
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			outputDir = filepath.Join(currentDir, opts.RepoName)
		} else {
			outputDir = filepath.Join(outputDir, opts.RepoName)
		}

		output.VerboseMessage("Running post-installation hooks...")
		fmt.Printf("Running post-installation hooks...")
		if err := c.Scaffolder.RunPostInstallHooks(
			outputDir,
			opts.Language,
			opts.ProjectType,
			templateVars,
		); err != nil {
			output.VerboseMessage(fmt.Sprintf("Failed to run post-installation hooks: %v", err))
			fmt.Println(" failed")
			summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to run post-installation hooks: %v", err))
		}
		output.VerboseMessage("Post-installation hooks completed successfully.")
		fmt.Println(" done")
	} else if opts.SkipHooks {
		output.VerboseMessage("Skipping post-installation hooks as requested.")
		fmt.Printf("Skipping post-installation hooks...")
		fmt.Println(" done")
	}

	output.VerboseMessage("Project creation workflow complete.")
	return summary, nil
}

// createRepository creates a GitHub repository
func (c *Creator) createRepository(opts CreateOptions) (string, error) {
	// Only print errors if needed, not process/info messages
	var org string
	if opts.IsOrg {
		org = opts.RepoOwner
	}
	return c.GitHubClient.CreateRepo(opts.RepoName, org, opts.IsPrivate, opts.RepoDescription)
}

// applyGitHubConfigurations applies configurations to the GitHub repository
func (c *Creator) applyGitHubConfigurations(opts CreateOptions) error {
	teamDir := filepath.Join(opts.ConfigDir, "teams", opts.Team)

	output.VerboseMessage(fmt.Sprintf("applyGitHubConfigurations: teamDir=%s, ApplyEnvs=%v, ApplyRulesets=%v, ApplySecrets=%v, RepoSecrets set=%v", teamDir, opts.ApplyEnvs, opts.ApplyRulesets, opts.ApplySecrets, opts.RepoSecrets != ""))

	// 1. Apply environments if requested
	if opts.ApplyEnvs {
		output.VerboseMessage("Applying team environments...")
		c.applyTeamEnvs(opts.RepoOwner, opts.RepoName, teamDir)
	} else {
		output.VerboseMessage("Creating default 'staging' environment (team envs not applied)...")
		// Create default staging environment if not applying team envs
		err := c.GitHubClient.CreateEnvironment(opts.RepoOwner, opts.RepoName, "staging")
		err = nil
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				fmt.Printf("\nFailed to create environment: %v\n", err)
			}
		}
	}

	// 2. Apply rulesets if requested
	if opts.ApplyRulesets {
		output.VerboseMessage("Applying team rulesets...")
		c.applyTeamRulesets(opts.RepoOwner, opts.RepoName, teamDir)
	}

	// 3. Apply secrets if requested
	if opts.ApplySecrets {
		output.VerboseMessage("Applying team secrets...")
		c.applyTeamSecrets(opts.RepoOwner, opts.RepoName, teamDir)
	}

	// 4. Apply repository-specific secrets if provided
	if opts.RepoSecrets != "" {
		output.VerboseMessage("Applying repository-specific secrets...")
		if err := c.applyRepoSpecificSecrets(opts.RepoOwner, opts.RepoName, opts.RepoSecrets); err != nil {
			output.VerboseMessage(fmt.Sprintf("Failed to apply repository-specific secrets: %v", err))
			return fmt.Errorf("failed to apply repository-specific secrets: %w", err)
		}
	}

	return nil
}

// scaffoldProject scaffolds a project locally
func (c *Creator) scaffoldProject(opts CreateOptions) error {
	outputDir := opts.OutputDir
	if outputDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		outputDir = filepath.Join(currentDir, opts.RepoName)
	}

	templateSource := opts.TemplateSource
	if templateSource == "" {
		return fmt.Errorf("template source not specified and could not be determined from config")
	}

	templateVars := createOptionsToTemplateVariables(opts)
	if err := c.Scaffolder.ScaffoldProjectWithOptions(outputDir, opts.Language, opts.ProjectType, templateSource, templateVars, opts.SkipHooks); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// scaffoldProjectWithVariables scaffolds a project locally with pre-populated template variables
func (c *Creator) scaffoldProjectWithVariables(opts CreateOptions, templateVars *template.Variables) error {
	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		// If no output directory is specified, use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		// Create a subdirectory with the project/repo name
		outputDir = filepath.Join(currentDir, opts.RepoName)
	} else {
		// If output directory is specified, create a subdirectory with the project/repo name
		outputDir = filepath.Join(outputDir, opts.RepoName)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	templateSource := opts.TemplateSource
	if templateSource == "" {
		// Use the default template for the specified language and project type
		template, err := c.TemplateRegistry.GetTemplate(opts.Language, opts.ProjectType)
		if err != nil {
			return fmt.Errorf("failed to find template for %s/%s: %w", opts.Language, opts.ProjectType, err)
		}
		templateSource = template.Source
	}

	output.VerboseMessage(fmt.Sprintf("Scaffolding project with template source: %s", templateSource))
	output.VerboseMessage(fmt.Sprintf("Template variables: ProjectName=%s, Language=%s, Type=%s, CreatedBy=%s",
		templateVars.ProjectName, templateVars.Language, templateVars.ProjectType, templateVars.CreatedBy))

	if err := c.Scaffolder.ScaffoldProjectWithOptions(outputDir, opts.Language, opts.ProjectType, templateSource, templateVars, opts.SkipHooks); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// The following methods are delegated to the appropriate handlers
// and should be implemented similarly to the functions in create.go

func (c *Creator) applyTeamEnvs(owner, repo, teamDir string) {
	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to get user home directory: %v", err))
			return
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	envsDir := filepath.Join(teamDir, "envs")

	entries, err := os.ReadDir(envsDir)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to read envs directory %s: %v", envsDir, err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(envsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to read env file %s: %v", filePath, err))
			continue
		}

		var envConfig struct {
			Name string `json:"name"`
		}

		// Parse JSON
		if err := json.Unmarshal(data, &envConfig); err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to parse JSON in %s: %v", filePath, err))
			continue
		}

		if envConfig.Name == "" {
			output.InfoMessage(fmt.Sprintf("Missing 'name' in %s, skipping", filePath))
			continue
		}

		output.ProcessingMessage(fmt.Sprintf("Creating environment: %s", envConfig.Name))
		if err := c.GitHubClient.CreateEnvironment(owner, repo, envConfig.Name); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				output.InfoMessage(fmt.Sprintf("Environment %s already exists, skipping", envConfig.Name))
			} else {
				output.ErrorMessage(fmt.Sprintf("Failed to apply env config %s: %v", filePath, err))
			}
		} else {
			output.SuccessMessage(fmt.Sprintf("Environment created: %s", envConfig.Name))
		}
	}
}

func (c *Creator) applyTeamRulesets(owner, repo, teamDir string) {
	rulesetsDir := filepath.Join(teamDir, "rulesets")
	files, err := os.ReadDir(rulesetsDir)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("No rulesets directory found at %s: %v", rulesetsDir, err))
		return
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(rulesetsDir, f.Name())
		output.ProcessingMessage(fmt.Sprintf("Processing ruleset file: %s", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to read ruleset file %s: %v", f.Name(), err))
			continue
		}

		// Unmarshal JSON directly into the GitHub API struct
		var ruleset github.RepositoryRuleset
		if err := json.Unmarshal(data, &ruleset); err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to parse JSON in %s: %v", f.Name(), err))
			continue
		}

		// Basic validation
		if ruleset.Name == "" {
			output.InfoMessage(fmt.Sprintf("Ruleset in %s is missing a name, skipping", f.Name()))
			continue
		}

		if ruleset.Target == nil {
			output.InfoMessage(fmt.Sprintf("Ruleset in %s is missing a target, skipping", f.Name()))
			continue
		}

		// Debug output
		output.ProcessingMessage(fmt.Sprintf("Applying ruleset: %s (target: %s)", ruleset.Name, *ruleset.Target))

		// Apply the ruleset
		if err := c.GitHubClient.CreateRuleset(owner, repo, ruleset); err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to apply ruleset %s: %v", f.Name(), err))
		} else {
			output.SuccessMessage(fmt.Sprintf("Applied ruleset: %s", ruleset.Name))
		}
	}
}

func (c *Creator) applyTeamSecrets(owner, repo, teamDir string) {
	secretsPath := filepath.Join(teamDir, "secrets.json")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		output.InfoMessage("No secrets.json file found")
		return
	}

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to read secrets file: %v", err))
		return
	}

	var secretsConfig struct {
		Secrets []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Env       string `json:"env,omitempty"`
			Type      string `json:"type,omitempty"`      // "secret" or "variable"
			Reference string `json:"reference,omitempty"` // Reference to another secret by name
		} `json:"secrets"`
	}

	// Parse JSON
	if err := json.Unmarshal(data, &secretsConfig); err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to parse JSON in %s: %v", secretsPath, err))
		return
	}

	// First pass to collect all secret values
	secretValues := make(map[string]string)
	for _, s := range secretsConfig.Secrets {
		// Skip if name is empty
		if s.Name == "" {
			output.InfoMessage("Skipping secret with missing name")
			continue
		}

		secretValue, sourceType, sourceKey, err := resolveSecretValue(s.Value, s.Name)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err))
			continue
		}

		// Store the secret value for potential references
		secretValues[s.Name] = secretValue

		// Debug info about resolution
		if sourceType != "config" {
			output.InfoMessage(fmt.Sprintf("Resolved '%s' from %s: '%s'", s.Name, sourceType, sourceKey))
		}
	}

	// Second pass to apply secrets, including those with references
	for _, s := range secretsConfig.Secrets {
		if s.Name == "" {
			continue // Skip again
		}

		// Convert anonymous struct to Secret type for helper compatibility
		secret := Secret{
			Name:      s.Name,
			Value:     s.Value,
			Env:       s.Env,
			Type:      s.Type,
			Reference: s.Reference,
		}

		secretValue, valueSource, ok := getSecretValueAndSource(secret, secretValues)
		if !ok {
			if s.Reference != "" {
				output.ErrorMessage(fmt.Sprintf("Referenced secret '%s' not found for '%s'", s.Reference, s.Name))
			}
			continue
		}

		if secretValue == "" {
			output.InfoMessage(fmt.Sprintf("Skipping secret '%s' with empty value", s.Name))
			continue
		}

		isVariable := s.Type == "variable"
		c.applySecretOrVariable(isVariable, owner, repo, s.Name, secretValue, s.Env, valueSource)
	}
}

func (c *Creator) applyRepoSpecificSecrets(owner, repo, secretsJSON string) error {
	var repoSecretsConfig struct {
		Secrets []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Env       string `json:"env,omitempty"`
			Type      string `json:"type,omitempty"`      // "secret" or "variable"
			Reference string `json:"reference,omitempty"` // Reference to another secret by name
		} `json:"secrets"`
	}

	// Parse the JSON string
	if err := json.Unmarshal([]byte(secretsJSON), &repoSecretsConfig); err != nil {
		return fmt.Errorf("failed to parse repository secrets JSON: %w", err)
	}

	// Apply the secrets
	for _, s := range repoSecretsConfig.Secrets {
		prefixedName := sanitizeSecretName(fmt.Sprintf("%s_%s", repo, s.Name))
		envScope := s.Env
		secretValue, _, _, err := resolveSecretValue(s.Value, s.Name)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err))
			continue
		}
		if secretValue == "" {
			output.InfoMessage(fmt.Sprintf("Skipping secret '%s' with empty value", s.Name))
			continue
		}
		isVariable := s.Type == "variable"
		c.applySecretOrVariable(isVariable, owner, repo, prefixedName, secretValue, envScope, "repo-secrets")
	}

	return nil
}

// pushToRepository initialises a Git repository in the local directory and pushes it to the remote GitHub repository
func (c *Creator) pushToRepository(localDir, remoteURL string) error {
	// TODO: Implement Git operations to push the local repository to GitHub
	// For now, just display a message about manual pushing
	output.InfoMessage(fmt.Sprintf("Repository created at: %s", localDir))
	output.InfoMessage("To manually push to GitHub, run the following commands:")

	// Use indentation for better readability of the commands
	fmt.Printf("  cd %s\n", localDir)
	fmt.Printf("  git init\n")
	fmt.Printf("  git add .\n")
	fmt.Printf("  git commit -m \"Initial commit\"\n")
	fmt.Printf("  git remote add origin %s\n", remoteURL)
	fmt.Printf("  git push -u origin main\n")

	return nil
}

// Secret represents a secret or variable definition for use in team/repo configs
// This matches the structure used in secretsConfig.Secrets
// (duplicated here to avoid import cycles and for helper use)
type Secret struct {
	Name      string
	Value     string
	Env       string
	Type      string
	Reference string
}

// Helper functions

// valueOrEmpty returns the value or a default value if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// extractGitHubActionsSecret extracts a secret name from GitHub Actions style syntax: ${{ secrets.SECRET_NAME }}
// Returns the secret name or empty string if no match
func extractGitHubActionsSecret(value string) string {
	// Simple regex-like pattern matching: ${{ secrets.KEY_NAME }}
	value = strings.TrimSpace(value)

	// Check if it follows the pattern
	if !strings.HasPrefix(value, "${{") || !strings.HasSuffix(value, "}}") {
		return ""
	}

	// Extract the part between ${{ and }}
	inner := strings.TrimSpace(value[3 : len(value)-2])

	// Check if it starts with secrets.
	if !strings.HasPrefix(inner, "secrets.") {
		return ""
	}

	// Extract the key name (everything after secrets.)
	keyName := strings.TrimSpace(inner[8:])
	if keyName == "" {
		return ""
	}

	return keyName
}

// sanitizeSecretName ensures a secret name follows GitHub's naming requirements:
// - Can only contain alphanumeric characters or underscores
// - Must start with a letter or underscore
// - No spaces allowed
func sanitizeSecretName(name string) string {
	// Replace hyphens with underscores
	sanitized := strings.ReplaceAll(name, "-", "_")

	// Ensure the name starts with a letter or underscore
	if len(sanitized) > 0 && !((sanitized[0] >= 'a' && sanitized[0] <= 'z') ||
		(sanitized[0] >= 'A' && sanitized[0] <= 'Z') ||
		sanitized[0] == '_') {
		sanitized = "_" + sanitized
	}

	// Replace any other invalid characters with underscores
	for i, char := range sanitized {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			sanitized = sanitized[:i] + "_" + sanitized[i+1:]
		}
	}

	return sanitized
}

// resolveSecretValue resolves a secret value from various sources (config, keyring, env vars)
// Returns the resolved value, source type, source key, and any error that occurred
func resolveSecretValue(value, name string) (string, string, string, error) {
	var secretValue string
	var sourceType, sourceKey string

	// Priority order for value resolution:
	// 1. Direct value in the config (which may contain references)
	// 2. Reference to another secret (handled separately)

	if value != "" {
		// 1a. Check if value contains GitHub Actions style reference: ${{ secrets.KEY_NAME }}
		if keyringKey := extractGitHubActionsSecret(value); keyringKey != "" {
			// Get from keyring/vault
			keyringValue, err := GetSecret(keyringKey)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to get keyring value for '%s': %w", keyringKey, err)
			}
			secretValue = keyringValue
			sourceType = "keyring"
			sourceKey = keyringKey
		} else if strings.HasPrefix(value, "$") && len(value) > 1 {
			// 1b. Simple $ENV_VAR syntax - get from environment variables
			envVarName := value[1:] // Remove the $ prefix
			envVarValue := os.Getenv(envVarName)
			if envVarValue == "" {
				fmt.Printf("Warning: Environment variable '%s' is empty or not set\n", envVarName)
			}
			secretValue = envVarValue
			sourceType = "env"
			sourceKey = envVarName
		} else {
			// 1c. Regular direct value
			secretValue = value
			sourceType = "config"
		}
	} else {
		return "", "", "", fmt.Errorf("no value source provided for '%s'", name)
	}

	return secretValue, sourceType, sourceKey, nil
}

// GetSecret retrieves a secret from the keyring
func GetSecret(key string) (string, error) {
	// Use the keyring service to get the secret
	return keyring.Get("viaplaycli", key)
}

// Helper to determine secret value and source
func getSecretValueAndSource(s Secret, secretValues map[string]string) (string, string, bool) {
	if s.Value != "" {
		if keyringKey := extractGitHubActionsSecret(s.Value); keyringKey != "" {
			return secretValues[s.Name], "keyring:" + keyringKey, true
		} else if strings.HasPrefix(s.Value, "$") && len(s.Value) > 1 {
			return secretValues[s.Name], "env:" + s.Value[1:], true
		} else {
			return s.Value, "config", true
		}
	} else if s.Reference != "" {
		refValue, exists := secretValues[s.Reference]
		if !exists {
			return "", "reference missing", false
		}
		return refValue, "reference:" + s.Reference, true
	}
	return "", "", false
}

// Helper to apply a secret or variable
func (c *Creator) applySecretOrVariable(isVariable bool, owner, repo, name, value, env, valueSource string) {
	if isVariable {
		output.ProcessingMessage(fmt.Sprintf("Setting variable: %s", name))
		err := c.GitHubClient.SetVariable(owner, repo, name, value, env)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to set variable '%s': %v", name, err))
			return
		}
		output.SuccessMessage(fmt.Sprintf("Applied variable: %s (env: %s, source: %s)", name, valueOrEmpty(env, "repo"), valueSource))
	} else {
		output.ProcessingMessage(fmt.Sprintf("Setting secret: %s", name))
		err := c.GitHubClient.ApplySecret(owner, repo, name, value, env)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to set secret '%s': %v", name, err))
			return
		}
		output.SuccessMessage(fmt.Sprintf("Applied secret: %s (env: %s, source: %s)", name, valueOrEmpty(env, "repo"), valueSource))
	}
}
