// Package project provides project creation and management functionality for viaplay-cli.
// It handles the scaffolding, configuration, and setup of new projects.
package project

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/git"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/progress"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/secrets"
	"github.com/nentgroup/viaplay-cli/internal/template"
	"github.com/nentgroup/viaplay-cli/pkg/tmpl"
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

	// Keep original project name for display (could include spaces)
	originalName := opts.RepoName

	// Convert to kebab-case for repo name and other technical identifiers
	kebabName := tmpl.ToKebabCase(originalName)

	// Basic project information
	vars.Project.Name = originalName
	vars.Project.Description = opts.RepoDescription

	// Repository information (always use kebab case)
	vars.Repo.Owner = opts.RepoOwner
	vars.Repo.Name = kebabName
	vars.Repo.IsPrivate = opts.IsPrivate
	vars.Repo.URL = fmt.Sprintf("https://github.com/%s/%s", opts.RepoOwner, kebabName)
	vars.Repo.SSHURL = fmt.Sprintf("git@github.com:%s/%s.git", opts.RepoOwner, kebabName)

	// Project language and type
	vars.Project.Language = opts.Language
	vars.Project.Type = opts.ProjectType
	vars.Org.Team = opts.Team

	// Additional values
	vars.Meta.CreatedAt = time.Now()
	vars.Meta.Year = time.Now().Year()

	// Service information
	vars.Service.Name = kebabName // Use kebab case for service name
	vars.Service.Owner = opts.Team
	vars.Service.OwnerKey = strings.ToLower(strings.ReplaceAll(opts.Team, " ", "-"))

	// Handle binary name for compiled languages (Go, Rust, etc.)
	binaryName := kebabName // Start with kebab case version
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

	// Set binary name based on language
	if opts.Language == "go" {
		vars.Go.BinaryName = binaryName
	} else if opts.Language == "rust" {
		vars.Rust.BinaryName = binaryName
		vars.Rust.CargoName = strings.ReplaceAll(kebabName, "-", "_") // Cargo names conventionally use underscores
	}

	// Go-specific variables
	if opts.Language == "go" {
		vars.Go.ModulePath = fmt.Sprintf("github.com/%s/%s", opts.RepoOwner, kebabName)
	}

	// Docker variables
	vars.Docker.ImageName = strings.ToLower(kebabName)
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
		templateVars.Meta.CreatedBy = username
		c.Reporter.Debug(fmt.Sprintf("Setting CreatedBy to authenticated user: %s", username))
	}

	// Convert project name to kebab-case for directory name
	kebabName := tmpl.ToKebabCase(opts.RepoName)

	// Determine the project path using kebab-case
	projectPath := opts.OutputDir
	if projectPath == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = filepath.Join(currentDir, kebabName)
	} else {
		projectPath = filepath.Join(projectPath, kebabName)
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
		summary.RepoURL = fmt.Sprintf("https://github.com/%s/%s", opts.RepoOwner, kebabName)
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
	if opts.Scaffold && !opts.SkipHooks { //nolint:nestif
		// Use the kebab-case directory for post-installation hooks
		c.Reporter.Start("Running post-installation hooks \n", "")
		if err := c.RunPostInstallHooks(
			projectPath, // Use the consistent kebab-case project path
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

	// Initialise and push to GitHub repository if both scaffolding is done and repo was created
	if opts.Scaffold && !opts.SkipRepo && createdRepo {
		c.Reporter.Start("Initializing Git repository and pushing to GitHub", "")

		// Create SSH URL from repository information
		sshURL := fmt.Sprintf("git@github.com:%s/%s.git", opts.RepoOwner, kebabName)
		c.Reporter.Debug(fmt.Sprintf("Using SSH URL for Git operations: %s", sshURL))

		if err := c.CloneToRepo(projectPath, sshURL); err != nil {
			c.Reporter.Failed("Git repository initialization", err, "")
			summary.Errors = append(summary.Errors, fmt.Sprintf("Failed to initialize and push to Git repository: %v", err))
		} else {
			c.Reporter.Complete("Git repository initialization", "Successfully pushed project to GitHub")
		}
	} else if opts.SkipRepo {
		c.Reporter.Skip("Git repository initialization", "Skipped as no GitHub repository was created")
	} else if !opts.Scaffold {
		c.Reporter.Skip("Git repository initialization", "Skipped as no local project was scaffolded")
	}

	c.Reporter.Complete("Project creation", "Workflow completed successfully")
	return summary, nil
}

// createRepository creates a GitHub repository and adds appropriate topics and labels
func (c *Creator) createRepository(opts CreateOptions) (string, error) {
	// Only print errors if needed, not process/info messages
	var org string
	if opts.IsOrg {
		org = opts.RepoOwner
	}
	repoURL, err := c.GitHubClient.CreateRepo(opts.RepoName, org, opts.IsPrivate, opts.RepoDescription)
	if err != nil {
		return "", err
	}

	// Generate appropriate topics for the repository
	topics := []string{}

	// Add language topic
	if opts.Language != "" {
		topics = append(topics, strings.ToLower(opts.Language))
	}

	// Add project type topic
	if opts.ProjectType != "" {
		topics = append(topics, strings.ToLower(opts.ProjectType))
	}

	// Add team topic if provided
	if opts.Team != "" {
		topics = append(topics, strings.ToLower(strings.ReplaceAll(opts.Team, " ", "-")))
	}

	// Add viaplay-cli topic to identify repos created by this tool
	topics = append(topics, "viaplay-cli")

	// Add the topics to the repository
	if err := c.GitHubClient.AddTopicsToRepo(opts.RepoOwner, opts.RepoName, topics); err != nil {
		c.Reporter.Warning("Topic Creation", fmt.Sprintf("Failed to add topics to repository: %v", err))
		// Don't return an error here as topic creation is not critical to the repository creation
	} else {
		c.Reporter.Debug(fmt.Sprintf("Added topics to repository: %v", topics))
	}

	return repoURL, nil
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

// CloneToRepo initialises a git repository in the project directory and pushes it to the remote
func (c *Creator) CloneToRepo(projectPath, repoURL string) error {
	if err := git.InitRepository(projectPath); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}
	if err := git.CommitAll(projectPath, "chore: initial commit"); err != nil {
		return fmt.Errorf("failed to commit files: %w", err)
	}
	if err := git.AddRemote(projectPath, "origin", repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}
	branch := "main"
	if err := git.Push(projectPath, "origin", branch); err != nil {
		fmt.Println("Push to 'main' failed, trying 'master' branch...")
		if err := git.Push(projectPath, "origin", "master"); err != nil {
			return fmt.Errorf("failed to push to remote: %w", err)
		}
	}
	return nil
}

// InitGoProject initialises a Go project with proper module setup
func (c *Creator) InitGoProject(projectPath string) error {
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
		return nil
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InitNodeProject initialises a Node.js project
func (c *Creator) InitNodeProject(projectPath string) error {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	nodeModulesPath := filepath.Join(projectPath, "node_modules")
	if _, err := os.Stat(packageJSONPath); err == nil {
		if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
			cmd := exec.Command("npm", "install")
			cmd.Dir = projectPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}
	return nil
}

// RunPostInstallHooks runs the post-installation hooks for a project
func (c *Creator) RunPostInstallHooks(projectPath, language, projectType string, templateVars *template.Variables) error {
	// Create a function that will run the hooks and write output to provided writers
	runHookFn := func(stdout, stderr io.Writer) error {
		// Create command executors that use the provided writers
		cmdExecutor := func(cmd *exec.Cmd) error {
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			return cmd.Run()
		}

		// Get renderer for template variables
		renderer := template.NewRenderer(templateVars)

		// Check if we have hooks for this language and project type
		hooks := c.Config.GetPostInstallHooks(language, projectType)
		if len(hooks) == 0 {
			fmt.Fprintf(stdout, "No hooks configured for %s/%s\n", language, projectType)
			return nil
		}

		// Execute hooks in order (general -> language-specific -> project-type-specific)
		fmt.Println("----------------------------------------")
		for _, hook := range hooks {
			// Process commands
			for _, cmd := range hook.GetAllCommands() {
				// Render template variables in the command
				renderedCmd, err := renderer.RenderString(cmd)
				if err != nil {
					return fmt.Errorf("failed to render run command template: %w", err)
				}

				// Create a command that will run in the project directory
				execCmd := exec.Command("sh", "-c", renderedCmd)
				execCmd.Dir = projectPath

				// Run the command using our executor
				if err := cmdExecutor(execCmd); err != nil {
					return fmt.Errorf("hook command failed: %w", err)
				}
			}

			// Process scripts
			for _, scriptPath := range hook.GetAllScripts() {
				// Render template variables in the script path
				renderedScriptPath, err := renderer.RenderString(scriptPath)
				if err != nil {
					return fmt.Errorf("failed to render script path template: %w", err)
				}

				// Check if this is a relative path or absolute
				fullScriptPath := renderedScriptPath
				if !filepath.IsAbs(renderedScriptPath) {
					// If it's relative, look in the hooks directory
					fullScriptPath = filepath.Join(c.Config.GetHooksDir(), renderedScriptPath)
				}
				// Check if script exists
				if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
					return fmt.Errorf("hook script not found: %s", fullScriptPath)
				}

				// Create a command to run the script
				execCmd := exec.Command(fullScriptPath)
				execCmd.Dir = projectPath

				// Run the script using our executor
				if err := cmdExecutor(execCmd); err != nil {
					return fmt.Errorf("hook script failed: %w", err)
				}
			}
		}
		fmt.Println("----------------------------------------")
		return nil
	}

	// Create a title for the TUI
	title := fmt.Sprintf("Post-Installation Hooks for %s/%s", language, projectType)

	// Display the hook output using our simplified UI
	err := output.DisplayHookOutput(title, runHookFn)
	// Display a simple message based on the result
	if err != nil {
		fmt.Printf("Hooks failed: %v\n", err)
	} else {
		fmt.Printf("Post-installation hooks completed successfully\n")
	}

	return err
}

// valueOrEmpty returns the value or a default value if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// scaffoldProjectWithVariables scaffolds a project locally with pre-populated template variables
func (c *Creator) scaffoldProjectWithVariables(opts CreateOptions, templateVars *template.Variables) error {
	// Convert project name to kebab-case for directory name
	kebabName := tmpl.ToKebabCase(opts.RepoName)

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		// If no output directory is specified, use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		// Create a subdirectory with the kebab-case project name
		outputDir = filepath.Join(currentDir, kebabName)
	} else {
		// If output directory is specified, create a subdirectory with the kebab-case project name
		outputDir = filepath.Join(outputDir, kebabName)
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
		return fmt.Errorf("environments directory does not exist: %s", envsDir)
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

	// Finalise the overall operation
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
		return fmt.Errorf("rulesets directory does not exist: %s", rulesetsDir)
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

		secretValue, sourceType, sourceKey, err := secrets.ResolveSecretValue(s.Value, s.Name)
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
		secret := secrets.Secret{
			Name:      s.Name,
			Value:     s.Value,
			Env:       s.Env,
			Type:      s.Type,
			Reference: s.Reference,
		}

		secretValue, valueSource, ok := secrets.GetSecretValueAndSource(secret, secretValues)
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
		prefixedName := secrets.SanitizeSecretName(fmt.Sprintf("%s_%s", repo, s.Name))
		envScope := s.Env

		secretValue, _, _, err := secrets.ResolveSecretValue(s.Value, s.Name)
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
