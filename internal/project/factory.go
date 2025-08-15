// Package project provides project creation and management functionality for viaplay-cli.
// It handles the scaffolding, configuration, and setup of new projects.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/git"
	"github.com/nentgroup/viaplay-cli/internal/progress"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/template"
	"github.com/nentgroup/viaplay-cli/pkg/tmpl"
)

// NewFactory creates a new project creator with a custom progress reporter
func NewFactory(ghClient *gh.GitHubClient, reporter progress.Reporter) *Factory {
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

	return &Factory{
		GitHubClient:     ghClient,
		Config:           cfg,
		TemplateRegistry: templateRegistry,
		CacheManager:     cacheManager,
		Scaffolder:       scaffolder,
		Reporter:         reporter, // Default to noop reporter
	}
}

// Create handles the full project creation workflow
func (c *Factory) Create(opts Options) (*Summary, error) {
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
	templateVars := c.optsToTemplateVars(opts)

	// Store template variables for later use with secret resolution
	c.templateVars = templateVars

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
		if err := c.setUp(opts, templateVars); err != nil {
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
	if err := c.applyConfigurations(opts); err != nil {
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
		if err := c.RunHooks(
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

		if err := c.Publish(projectPath, sshURL); err != nil {
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

// Publish initialises a git repository in the project directory and pushes it to the remote
func (c *Factory) Publish(projectPath, repoURL string) error {
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

// setUp scaffolds a project locally with pre-populated template variables
func (c *Factory) setUp(opts Options, templateVars *template.Variables) error {
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

// valueOrEmpty returns the value or a default value if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// optsToTemplateVars converts project creation options to template variables
func (c *Factory) optsToTemplateVars(opts Options) *template.Variables {
	vars := template.NewTemplateVariables()

	// Keep original input name
	originalName := opts.RepoName

	// Convert to kebab-case for repo name and other technical identifiers
	kebabName := tmpl.ToKebabCase(originalName)

	// Convert to title case for user-friendly display name
	titleCaseName := tmpl.ToTitleCase(originalName)

	// Basic project information
	vars.Project.Name = titleCaseName // Use title case for display name
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

	// If this is an organization repo and we have a team name, try to fetch the team ID
	if opts.IsOrg && opts.Team != "" {
		c.Reporter.Debug(fmt.Sprintf("Attempting to fetch team ID for '%s' in org '%s'", opts.Team, opts.RepoOwner))
		teamID, err := c.GitHubClient.GetTeamID(opts.RepoOwner, opts.Team)
		if err != nil {
			c.Reporter.Warning("Team ID", fmt.Sprintf("Could not fetch team ID: %v", err))
		} else {
			vars.Org.TeamID = teamID
			c.Reporter.Debug(fmt.Sprintf("Successfully fetched team ID %d for team '%s'", teamID, opts.Team))
			vars.Org.Name = opts.RepoOwner // Set organization name
		}
	}

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
	switch opts.Language {
	case "go":
		vars.Go.BinaryName = binaryName
	case "rust":
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

// getTemplateRenderer creates a template renderer with the current project's template variables
func (c *Factory) getTemplateRenderer() *template.Renderer {
	// Create a new renderer using the template variables
	return template.NewRenderer(c.templateVars)
}
