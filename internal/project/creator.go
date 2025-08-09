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
	"github.com/nentgroup/viaplay-cli/internal/progress"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/template"
)

// Summary contains details about the created project to be displayed to the user
type Summary struct {
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
	CleanedUp       bool     // Whether resources were cleaned up due to an error
	CleanupDetails  []string // Details about what was cleaned up
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

	// Progress reporter for tracking operation progress
	Reporter progress.Reporter
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
	NoCache        bool   // Force update of template cache before using it

	// Error handling options
	CleanupOnError bool // Clean up resources (delete folder/repo) if errors occur
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
		Reporter:         progress.NewNoopReporter(), // Default to noop reporter
	}
}

// NewCreatorWithReporter creates a new project creator with a custom progress reporter
func NewCreatorWithReporter(ghClient *gh.GitHubClient, configDir string, reporter progress.Reporter) *Creator {
	creator := NewCreator(ghClient, configDir)
	creator.Reporter = reporter
	return creator
}

// SetReporter sets a custom reporter for the creator
func (c *Creator) SetReporter(reporter progress.Reporter) {
	if reporter == nil {
		c.Reporter = progress.NewNoopReporter()
		return
	}
	c.Reporter = reporter
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
	if err := c.Scaffolder.ScaffoldProject(outputDir, opts.Language, opts.ProjectType, templateSource, templateVars, opts.SkipHooks, opts.NoCache); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// Create handles the full project creation workflow
func (c *Creator) Create(opts CreateOptions) (*Summary, error) {
	// Debug info when available
	c.Reporter.Debug(fmt.Sprintf("Starting project creation with options: %+v", opts))

	// Initialise project summary
	summary := &Summary{
		Language:        opts.Language,
		ProjectType:     opts.ProjectType,
		Team:            opts.Team,
		AppliedEnvs:     opts.ApplyEnvs,
		AppliedRulesets: opts.ApplyRulesets,
		AppliedSecrets:  opts.ApplySecrets,
		CustomSecrets:   opts.RepoSecrets != "",
		Errors:          []string{},
	}

	// Track resources for potential cleanup
	var createdProjectDir string
	var createdRepo bool

	// Define cleanup function
	cleanup := func() {
		if !opts.CleanupOnError {
			return
		}

		summary.CleanedUp = true
		c.Reporter.Start("Cleaning up resources due to error", "")

		// 1. Delete project directory if it was created
		if createdProjectDir != "" && opts.Scaffold {
			c.Reporter.Progress("Cleanup", 0, fmt.Sprintf("Deleting project directory: %s", createdProjectDir))
			if err := os.RemoveAll(createdProjectDir); err != nil {
				c.Reporter.Warning("Cleanup", fmt.Sprintf("Failed to delete project directory: %v", err))
				summary.CleanupDetails = append(summary.CleanupDetails, fmt.Sprintf("Failed to delete project directory: %v", err))
			} else {
				c.Reporter.Progress("Cleanup", 50, "Project directory deleted")
				summary.CleanupDetails = append(summary.CleanupDetails, fmt.Sprintf("Project directory deleted: %s", createdProjectDir))
			}
		}

		// 2. Delete GitHub repository if it was created
		if createdRepo && !opts.SkipRepo {
			c.Reporter.Progress("Cleanup", 50, fmt.Sprintf("Deleting GitHub repository: %s/%s", opts.RepoOwner, opts.RepoName))
			if err := c.GitHubClient.DeleteRepo(opts.RepoOwner, opts.RepoName); err != nil {
				c.Reporter.Warning("Cleanup", fmt.Sprintf("Failed to delete GitHub repository: %v", err))
				summary.CleanupDetails = append(summary.CleanupDetails, fmt.Sprintf("Failed to delete GitHub repository: %v", err))
			} else {
				c.Reporter.Progress("Cleanup", 100, "GitHub repository deleted")
				summary.CleanupDetails = append(summary.CleanupDetails, fmt.Sprintf("GitHub repository deleted: %s/%s", opts.RepoOwner, opts.RepoName))
			}
		}

		c.Reporter.Complete("Cleanup", "Resources cleaned up")
	}

	// Get authenticated user for CreatedBy field
	c.Reporter.Start("Getting authenticated user", "")
	username, err := c.GitHubClient.GetAuthenticatedUser()
	if err != nil {
		c.Reporter.Failed("Getting authenticated user", err, "")
		summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to get authenticated username: %v", err))
	} else {
		c.Reporter.Complete("Getting authenticated user", "")
	}

	// Convert options to template variables with additional info
	templateVars := createOptionsToTemplateVariables(opts)

	// Set authenticated username if available
	if username != "" {
		templateVars.CreatedBy = username
		c.Reporter.Debug(fmt.Sprintf("Setting CreatedBy to authenticated user: %s", username))
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

	// Scaffold the project if requested
	if opts.Scaffold {
		c.Reporter.Start("Scaffolding project", "")
		if err := c.scaffoldProjectWithVariables(opts, templateVars); err != nil {
			if opts.CleanupOnError {
				cleanup()
				summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to scaffold project: %v", err))
				return summary, fmt.Errorf("failed to scaffold project: %w", err)
			}
			return nil, fmt.Errorf("failed to scaffold project: %w", err)
		}
		createdProjectDir = projectPath // Track created directory for potential cleanup
		c.Reporter.Complete("Scaffolding project", "complete!")
	}

	// Create GitHub repository if not skipped
	if opts.SkipRepo {
		c.Reporter.Skip("Creating GitHub repository", "Skipped as per user request")
	} else {
		c.Reporter.Start("Creating GitHub repository", "")
		_, err = c.createRepository(opts)
		if err != nil {
			if strings.Contains(err.Error(), "name already exists on this account") {
				c.Reporter.Skip("Creating GitHub repository", "Repository already exists")
			} else {
				c.Reporter.Failed("Creating GitHub repository", err, "")
				if opts.CleanupOnError {
					cleanup()
					summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to create repository: %v", err))
					return summary, fmt.Errorf("failed to create repository: %w", err)
				}
				return nil, fmt.Errorf("failed to create repository: %w", err)
			}
		} else {
			createdRepo = true // Track created repo for potential cleanup
			c.Reporter.Complete("Creating GitHub repository", "")
		}

		// Set the repository URL in the summary
		summary.RepoURL = fmt.Sprintf("https://github.com/%s/%s", opts.RepoOwner, opts.RepoName)
	}

	// Apply GitHub configurations
	c.Reporter.Start("Applying GitHub configurations", "")
	if err := c.applyGitHubConfigurations(opts); err != nil {
		if opts.CleanupOnError {
			cleanup()
			summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
			return summary, fmt.Errorf("failed to apply GitHub configurations: %w", err)
		}
		summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
	} else {
		c.Reporter.Complete("Applying GitHub configurations", "")
	}

	// Run post-installation hooks if scaffolding was done and hooks aren't skipped
	if opts.Scaffold && !opts.SkipHooks {
		outputDir := opts.OutputDir
		if outputDir == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				if opts.CleanupOnError {
					cleanup()
					summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to get current directory: %v", err))
					return summary, fmt.Errorf("failed to get current directory: %w", err)
				}
				return nil, fmt.Errorf("failed to get current directory: %w", err)
			}
			outputDir = filepath.Join(currentDir, opts.RepoName)
		} else {
			outputDir = filepath.Join(outputDir, opts.RepoName)
		}

		c.Reporter.Start("Running post-installation hooks \n", "")
		if err := c.Scaffolder.RunPostInstallHooks(
			outputDir,
			opts.Language,
			opts.ProjectType,
			templateVars,
		); err != nil {
			summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to run post-installation hooks: %v", err))
		} else {
			c.Reporter.Complete("Running post-installation hooks", "complete!")
		}
	} else if opts.SkipHooks {
		c.Reporter.Skip("Running post-installation hooks", "Skipped as per user request")
	}

	c.Reporter.Complete("Project creation", "Workflow completed successfully")
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

	// Skip all GitHub configurations if SkipRepo is true
	if opts.SkipRepo {
		return nil
	}

	// 1. Apply environments if requested
	if opts.ApplyEnvs {
		c.applyTeamEnvs(opts.RepoOwner, opts.RepoName, teamDir)
	} else {
		// Create default staging environment if not applying team envs
		err := c.GitHubClient.CreateEnvironment(opts.RepoOwner, opts.RepoName, "staging")
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create environment: %w", err)
			}
		}
	}

	// 2. Apply rulesets if requested
	if opts.ApplyRulesets {
		if err := c.applyTeamRulesets(opts.RepoOwner, opts.RepoName, teamDir); err != nil {
			return fmt.Errorf("failed to apply team rulesets: %w", err)
		}
	}

	// 3. Apply secrets if requested
	if opts.ApplySecrets {
		if err := c.applyTeamSecrets(opts.RepoOwner, opts.RepoName, teamDir); err != nil {
			return fmt.Errorf("failed to apply team secrets: %w", err)
		}
	}

	// 4. Apply repository-specific secrets if provided
	if opts.RepoSecrets != "" {
		if err := c.applyRepoSpecificSecrets(opts.RepoOwner, opts.RepoName, opts.RepoSecrets); err != nil {
			return fmt.Errorf("failed to apply repository-specific secrets: %w", err)
		}
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

	if err := c.Scaffolder.ScaffoldProject(outputDir, opts.Language, opts.ProjectType, templateSource, templateVars, opts.SkipHooks, opts.NoCache); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// The following methods are delegated to the appropriate handlers
// and should be implemented similarly to the functions in create.go

func (c *Creator) applyTeamEnvs(owner, repo, teamDir string) error {
	mainOperation := "Applying team environments"

	// Start the overall operation
	c.Reporter.Start(mainOperation, "")

	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Reporter.Failed(mainOperation, err, "Failed to get user home directory")
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	envsDir := filepath.Join(teamDir, "envs")
	c.Reporter.Debug(fmt.Sprintf("Looking for environment configs in %s", envsDir))

	// Check if the directory exists first
	if _, err := os.Stat(envsDir); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Environments directory does not exist: %s", envsDir)
		c.Reporter.Skip(mainOperation, errMsg)
		return fmt.Errorf(errMsg)
	}

	entries, err := os.ReadDir(envsDir)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, fmt.Sprintf("Failed to read envs directory: %s", envsDir))
		return fmt.Errorf("failed to read environments directory: %w", err)
	}

	if len(entries) == 0 {
		c.Reporter.Skip(mainOperation, "No environment configurations found")
		return nil
	}

	appliedCount := 0
	failedEnvs := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(envsDir, entry.Name())
		c.Reporter.Debug(fmt.Sprintf("Processing environment file: %s", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read env file %s: %v", entry.Name(), err)
			c.Reporter.Warning("Environment processing", errMsg)
			failedEnvs = append(failedEnvs, errMsg)
			continue
		}

		var envConfig struct {
			Name string `json:"name"`
		}

		// Parse JSON
		if err := json.Unmarshal(data, &envConfig); err != nil {
			errMsg := fmt.Sprintf("Failed to parse JSON in %s: %v", entry.Name(), err)
			c.Reporter.Warning("Environment processing", errMsg)
			failedEnvs = append(failedEnvs, errMsg)
			continue
		}

		if envConfig.Name == "" {
			c.Reporter.Skip("Environment processing", fmt.Sprintf("Environment in %s is missing a name", entry.Name()))
			continue
		}

		// Update the main operation with current environment being processed
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Creating environment: %s", envConfig.Name))

		// Create the environment
		if err := c.GitHubClient.CreateEnvironment(owner, repo, envConfig.Name); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				c.Reporter.Debug(fmt.Sprintf("Environment %s already exists", envConfig.Name))
			} else {
				errMsg := fmt.Sprintf("Failed to create environment %s: %v", envConfig.Name, err)
				c.Reporter.Warning("Environment creation", errMsg)
				failedEnvs = append(failedEnvs, errMsg)
			}
		} else {
			c.Reporter.Debug(fmt.Sprintf("Successfully created environment: %s", envConfig.Name))
			appliedCount++
		}
	}

	// Return a summary error if any environments failed
	if len(failedEnvs) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d environments, %d failed", appliedCount, len(failedEnvs))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some environments could not be applied: %s", strings.Join(failedEnvs[:1], ", "))
	}

	// Finalize the overall operation
	if appliedCount > 0 {
		c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d environments", appliedCount))
	} else {
		c.Reporter.Skip(mainOperation, "No new environments were applied")
	}

	return nil
}

func (c *Creator) applyTeamRulesets(owner, repo, teamDir string) error {
	mainOperation := "Applying team rulesets"
	// Start the overall operation
	c.Reporter.Start(mainOperation, "")

	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Reporter.Failed(mainOperation, err, "Failed to get user home directory")
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	rulesetsDir := filepath.Join(teamDir, "rulesets")
	c.Reporter.Debug(fmt.Sprintf("Looking for ruleset configs in %s", rulesetsDir))

	// Check if the directory exists first
	if _, err := os.Stat(rulesetsDir); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Rulesets directory does not exist: %s", rulesetsDir)
		c.Reporter.Skip(mainOperation, errMsg)
		return fmt.Errorf(errMsg)
	}

	files, err := os.ReadDir(rulesetsDir)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to read rulesets directory")
		return fmt.Errorf("failed to read rulesets directory: %w", err)
	}

	if len(files) == 0 {
		c.Reporter.Skip(mainOperation, "No ruleset files found")
		return nil
	}

	// Track applied and failed rulesets
	appliedCount := 0
	failedRulesets := []string{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// Only process JSON files
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(rulesetsDir, f.Name())
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Processing ruleset file: %s", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read ruleset file %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset processing", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
			continue
		}

		// Unmarshal JSON directly into the GitHub API struct
		var ruleset github.RepositoryRuleset
		if err := json.Unmarshal(data, &ruleset); err != nil {
			errMsg := fmt.Sprintf("Failed to parse JSON in %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset processing", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
			continue
		}

		// Basic validation
		if ruleset.Name == "" {
			c.Reporter.Skip("Ruleset processing", fmt.Sprintf("Ruleset in %s is missing a name", f.Name()))
			continue
		}

		if ruleset.Target == nil {
			c.Reporter.Skip("Ruleset processing", fmt.Sprintf("Ruleset in %s is missing a target", f.Name()))
			continue
		}

		// Debug output
		c.Reporter.Debug(fmt.Sprintf("Applying ruleset: %s (target: %s)", ruleset.Name, *ruleset.Target))

		// Apply the ruleset
		if err := c.GitHubClient.CreateRuleset(owner, repo, ruleset); err != nil {
			errMsg := fmt.Sprintf("Failed to apply ruleset %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset application", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
		} else {
			c.Reporter.Progress("Ruleset application", 100, fmt.Sprintf("Applied ruleset: %s", ruleset.Name))
			appliedCount++
		}
	}

	// Return a summary error if any rulesets failed
	if len(failedRulesets) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d rulesets, %d failed", appliedCount, len(failedRulesets))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some rulesets could not be applied: %s", strings.Join(failedRulesets[:1], ", "))
	}

	c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d rulesets", appliedCount))
	return nil
}

func (c *Creator) applyTeamSecrets(owner, repo, teamDir string) error {
	mainOperation := "Applying team secrets"
	c.Reporter.Start(mainOperation, "")

	secretsPath := filepath.Join(teamDir, "secrets.json")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		c.Reporter.Skip(mainOperation, "No secrets.json file found")
		return nil
	}

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to read secrets file")
		return fmt.Errorf("failed to read secrets file: %w", err)
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
		c.Reporter.Failed(mainOperation, err, fmt.Sprintf("Failed to parse JSON in %s", secretsPath))
		return fmt.Errorf("failed to parse secrets JSON: %w", err)
	}

	// First pass to collect all secret values
	secretValues := make(map[string]string)
	for _, s := range secretsConfig.Secrets {
		// Skip if name is empty
		if s.Name == "" {
			c.Reporter.Skip("Secret processing", "Skipping secret with missing name")
			continue
		}

		secretValue, sourceType, sourceKey, err := resolveSecretValue(s.Value, s.Name)
		if err != nil {
			c.Reporter.Warning("Secret resolution", fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err))
			continue
		}

		// Store the secret value for potential references
		secretValues[s.Name] = secretValue

		// Debug info about resolution
		if sourceType != "config" {
			c.Reporter.Debug(fmt.Sprintf("Resolved '%s' from %s: '%s'", s.Name, sourceType, sourceKey))
		}
	}

	// Second pass to apply secrets, including those with references
	appliedCount := 0
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
				c.Reporter.Warning("Secret reference", fmt.Sprintf("Referenced secret '%s' not found for '%s'", s.Reference, s.Name))
			}
			continue
		}

		if secretValue == "" {
			c.Reporter.Skip("Secret processing", fmt.Sprintf("Skipping secret '%s' with empty value", s.Name))
			continue
		}

		isVariable := s.Type == "variable"
		if err := c.applySecretOrVariable(isVariable, owner, repo, s.Name, secretValue, s.Env, valueSource); err != nil {
			c.Reporter.Warning("Secret application", fmt.Sprintf("Failed to apply %s '%s': %v",
				secret.Type, secret.Name, err))
		} else {
			c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Applied %s: %s (env: %s, source: %s)",
				valueOrEmpty(s.Type, "secret"), s.Name, valueOrEmpty(s.Env, "repo"), valueSource))
			appliedCount++
		}
	}

	if appliedCount > 0 {
		c.Reporter.Complete(mainOperation, fmt.Sprintf("Applied %d secrets/variables", appliedCount))
	} else {
		c.Reporter.Skip(mainOperation, "No secrets were applied")
	}

	return nil
}

func (c *Creator) applyRepoSpecificSecrets(owner, repo, secretsJSON string) error {
	mainOperation := "Applying repository-specific secrets"
	c.Reporter.Start(mainOperation, "")

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
		c.Reporter.Failed(mainOperation, err, "Failed to parse repository secrets JSON")
		return fmt.Errorf("failed to parse repository secrets JSON: %w", err)
	}

	if len(repoSecretsConfig.Secrets) == 0 {
		c.Reporter.Skip(mainOperation, "No repository-specific secrets found")
		return nil
	}

	// Apply the secrets
	appliedCount := 0
	failedSecrets := []string{}

	for _, s := range repoSecretsConfig.Secrets {
		prefixedName := sanitizeSecretName(fmt.Sprintf("%s_%s", repo, s.Name))
		envScope := s.Env

		secretValue, _, _, err := resolveSecretValue(s.Value, s.Name)
		if err != nil {
			errMsg := fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err)
			c.Reporter.Warning("Secret resolution", errMsg)
			failedSecrets = append(failedSecrets, errMsg)
			continue
		}

		if secretValue == "" {
			c.Reporter.Skip("Secret processing", fmt.Sprintf("Skipping secret '%s' with empty value", s.Name))
			continue
		}

		isVariable := s.Type == "variable"
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Setting %s: %s", valueOrEmpty(s.Type, "secret"), prefixedName))

		if err := c.applySecretOrVariable(isVariable, owner, repo, prefixedName, secretValue, envScope, "repo-secrets"); err != nil {
			errMsg := fmt.Sprintf("Failed to apply %s '%s': %v", valueOrEmpty(s.Type, "secret"), prefixedName, err)
			c.Reporter.Warning("Secret application", errMsg)
			failedSecrets = append(failedSecrets, errMsg)
		} else {
			appliedCount++
		}
	}

	// Return a summary error if any secrets failed
	if len(failedSecrets) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d repository-specific secrets, %d failed", appliedCount, len(failedSecrets))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some repository-specific secrets could not be applied: %s", strings.Join(failedSecrets[:1], ", "))
	}

	c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d repository-specific secrets", appliedCount))
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

// Source type constants
const (
	// SourceTypeConfig represents a configuration source type
	SourceTypeConfig = "config"
	// SourceTypeKeyring represents a keyring source type
	SourceTypeKeyring = "keyring"
	// SourceTypeEnv represents an environment variable source type
	SourceTypeEnv = "env"
)

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
	if len(sanitized) > 0 && (!isAlpha(sanitized[0]) && sanitized[0] != '_') {
		sanitized = "_" + sanitized
	}

	// Replace any other invalid characters with underscores
	for i, char := range sanitized {
		if !isAlphaNumeric(char) && char != '_' {
			sanitized = sanitized[:i] + "_" + sanitized[i+1:]
		}
	}

	return sanitized
}

// isAlpha checks if a byte is an alphabetic character (a-z, A-Z)
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isAlphaNumeric checks if a rune is an alphanumeric character (a-z, A-Z, 0-9)
func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
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
func (c *Creator) applySecretOrVariable(isVariable bool, owner, repo, name, value, env, valueSource string) error {
	operation := "Setting variable"
	resourceType := "variable"
	if !isVariable {
		operation = "Setting secret"
		resourceType = "secret"
	}

	c.Reporter.Progress(operation, 0, name)

	var err error
	if isVariable {
		err = c.GitHubClient.SetVariable(owner, repo, name, value, env)
	} else {
		err = c.GitHubClient.ApplySecret(owner, repo, name, value, env)
	}

	if err != nil {
		c.Reporter.Warning(operation, fmt.Sprintf("Failed to set %s '%s': %v", resourceType, name, err))
		return err
	}

	c.Reporter.Debug(fmt.Sprintf("Applied %s: %s (env: %s, source: %s)",
		resourceType, name, valueOrEmpty(env, "repo"), valueSource))
	return nil
}
