// Package project provides project creation and management functionality for viaplay-cli.
// It handles the scaffolding, configuration, and setup of new projects.
package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/git"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/template"
	"github.com/nentgroup/viaplay-cli/pkg/tmpl"
)

// creationContext holds state for a single project creation workflow.
type creationContext struct {
	opts              Options
	Summary           *Summary
	Cleanup           func()
	ProjectPath       string
	KebabName         string
	CreatedProjectDir string
	CreatedRepo       bool
	AuthenticatedUser string
	TemplateVars      *template.Variables
}

// Factory manages the project creation workflow
type Factory struct {
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
	Reporter output.Reporter

	templateVars *template.Variables
}

// NewFactory creates a new project creator with a custom progress reporter
func NewFactory(ghClient *gh.GitHubClient, reporter output.Reporter, cfg *config.Configuration) *Factory {
	// Create cache manager
	cacheManager := cache.NewManager(cfg)

	// Create template registry
	templateRegistry := registry.NewRegistry(cfg)
	err := templateRegistry.LoadTemplates()
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
func (c *Factory) Create(ctx context.Context, opts Options) (*Summary, error) {
	cctx, err := c.newCreationContext(ctx, opts)
	if err != nil {
		// Context initialisation failed; we cannot guarantee a valid summary here.
		return nil, err
	}

	c.populateAuthenticatedUser(ctx, cctx)
	c.prepareTemplateVars(ctx, cctx)

	if err := c.scaffoldIfNeeded(ctx, cctx); err != nil {
		return cctx.Summary, err
	}

	if err := c.ensureRepoIfNeeded(ctx, cctx); err != nil {
		return cctx.Summary, err
	}

	if err := c.configureGitHub(ctx, cctx); err != nil {
		return cctx.Summary, err
	}

	c.runHooksAndPublish(ctx, cctx)

	c.Reporter.Complete("Project creation", "Workflow completed successfully")
	return cctx.Summary, nil
}

// ApplyConfigurations applies repository configuration to an existing repository.
func (c *Factory) ApplyConfigurations(ctx context.Context, opts Options) error {
	if u, err := c.GitHubClient.GetUser(ctx, opts.RepoOwner); err == nil && u.Type != nil {
		switch strings.ToLower(u.GetType()) {
		case "organization":
			opts.AccountType = OrganizationAccount
		default:
			opts.AccountType = PersonalAccount
		}
	}

	c.templateVars = c.optsToTemplateVars(ctx, opts)
	if username, err := c.GitHubClient.GetAuthenticatedUser(ctx); err == nil && username != "" {
		c.templateVars.Meta.CreatedBy = username
	}

	return c.applyConfigurations(ctx, opts)
}

// ApplyEnvs applies environments to an existing repository
func (c *Factory) ApplyEnvs(ctx context.Context, owner, repo, configDir, team string) error {
	teamDir := filepath.Join(configDir, "teams", team)
	return c.applyEnvs(ctx, owner, repo, teamDir)
}

// ApplyRulesets applies rulesets to an existing repository
func (c *Factory) ApplyRulesets(ctx context.Context, owner, repo, configDir, team string) error {
	teamDir := filepath.Join(configDir, "teams", team)
	return c.applyRulesets(ctx, owner, repo, teamDir)
}

// ApplySecrets applies secrets to an existing repository
func (c *Factory) ApplySecrets(ctx context.Context, owner, repo, configDir, team string) error {
	teamDir := filepath.Join(configDir, "teams", team)
	return c.applySecrets(ctx, owner, repo, teamDir, true, true)
}

// ApplyRepoSecrets applies repository-specific secrets to an existing repository
func (c *Factory) ApplyRepoSecrets(ctx context.Context, owner, repo, secretsJSON string) error {
	return c.applyRepoSecrets(ctx, owner, repo, secretsJSON)
}

// newCreationContext initialises the creation context: account type, summary, cleanup, names and paths.
func (c *Factory) newCreationContext(ctx context.Context, opts Options) (*creationContext, error) {
	c.Reporter.Debug(fmt.Sprintf("Starting project creation with options: %+v", opts))

	u, err := c.GitHubClient.GetUser(ctx, opts.RepoOwner)
	if err != nil {
		return &creationContext{Summary: &Summary{}}, fmt.Errorf("failed to get user %s: %w", opts.RepoOwner, err)
	}

	// Normalise account type from GitHub user type
	if u.Type != nil {
		opts.AccountType = AccountType(strings.ToLower(*u.Type))
	}

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

	cctx := &creationContext{
		opts:    opts,
		Summary: summary,
	}

	// Track resources for potential cleanup
	cctx.CreatedProjectDir = ""
	cctx.CreatedRepo = false

	cctx.Cleanup = func() {
		if !cctx.opts.CleanupOnError {
			return
		}

		cctx.Summary.CleanedUp = true
		c.Reporter.Start("Cleaning up resources due to error", "")

		// 1. Delete project directory if it was created
		if cctx.CreatedProjectDir != "" && cctx.opts.Scaffold {
			c.Reporter.Progress("Cleanup", 0, fmt.Sprintf("Deleting project directory: %s", cctx.CreatedProjectDir))
			if err := os.RemoveAll(cctx.CreatedProjectDir); err != nil {
				c.Reporter.Warning("Cleanup", fmt.Sprintf("Failed to delete project directory: %v", err))
				cctx.Summary.CleanupDetails = append(cctx.Summary.CleanupDetails, fmt.Sprintf("Failed to delete project directory: %v", err))
			} else {
				c.Reporter.Progress("Cleanup", 50, "Project directory deleted")
				cctx.Summary.CleanupDetails = append(cctx.Summary.CleanupDetails, fmt.Sprintf("Project directory deleted: %s", cctx.CreatedProjectDir))
			}
		}

		// 2. Delete GitHub repository if it was created
		if cctx.CreatedRepo && !cctx.opts.SkipRepo {
			c.Reporter.Progress("Cleanup", 50, fmt.Sprintf("Deleting GitHub repository: %s/%s", cctx.opts.RepoOwner, cctx.opts.RepoName))
			if err := c.GitHubClient.DeleteRepo(ctx, cctx.opts.RepoOwner, cctx.opts.RepoName); err != nil {
				c.Reporter.Warning("Cleanup", fmt.Sprintf("Failed to delete GitHub repository: %v", err))
				cctx.Summary.CleanupDetails = append(cctx.Summary.CleanupDetails, fmt.Sprintf("Failed to delete GitHub repository: %v", err))
			} else {
				c.Reporter.Progress("Cleanup", 100, "GitHub repository deleted")
				cctx.Summary.CleanupDetails = append(cctx.Summary.CleanupDetails, fmt.Sprintf("GitHub repository deleted: %s/%s", cctx.opts.RepoOwner, cctx.opts.RepoName))
			}
		}

		c.Reporter.Complete("Cleanup", "Resources cleaned up")
	}

	// Convert project name to kebab-case for directory name
	cctx.KebabName = tmpl.ToKebabCase(opts.RepoName)

	// Determine the project path using kebab-case
	projectPath := opts.OutputDir
	if projectPath == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return cctx, fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = filepath.Join(currentDir, cctx.KebabName)
	} else {
		projectPath = filepath.Join(projectPath, cctx.KebabName)
	}
	cctx.ProjectPath = projectPath
	cctx.Summary.ProjectPath = projectPath

	return cctx, nil
}

// populateAuthenticatedUser fetches the authenticated username and updates the context and summary.
func (c *Factory) populateAuthenticatedUser(ctx context.Context, cctx *creationContext) {
	c.Reporter.Start("Getting authenticated user", "")
	username, err := c.GitHubClient.GetAuthenticatedUser(ctx)
	if err != nil {
		c.Reporter.Failed("Getting authenticated user", err, "")
		cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to get authenticated username: %v", err))
		return
	}
	c.Reporter.Complete("Getting authenticated user", "")
	cctx.AuthenticatedUser = username
}

// prepareTemplateVars initialises template variables and attaches them to the context and factory.
func (c *Factory) prepareTemplateVars(ctx context.Context, cctx *creationContext) {
	templateVars := c.optsToTemplateVars(ctx, cctx.opts)
	cctx.TemplateVars = templateVars
	c.templateVars = templateVars

	if cctx.AuthenticatedUser != "" {
		cctx.TemplateVars.Meta.CreatedBy = cctx.AuthenticatedUser
		c.Reporter.Debug(fmt.Sprintf("Setting CreatedBy to authenticated user: %s", cctx.AuthenticatedUser))
	}
}

// scaffoldIfNeeded performs project scaffolding when requested and updates tracking variables.
func (c *Factory) scaffoldIfNeeded(ctx context.Context, cctx *creationContext) error {
	if !cctx.opts.Scaffold {
		return nil
	}

	c.Reporter.Start("Scaffolding project", "")
	// Set CreatedProjectDir before attempting to scaffold (rather than only on
	// success) so Cleanup can remove any partially-created output directory if
	// setUp fails partway through (e.g. a file-copy error after the directory
	// was already created). os.RemoveAll on a path that was never created is a
	// harmless no-op, so this is safe even when setUp fails before creating
	// anything on disk.
	cctx.CreatedProjectDir = cctx.ProjectPath
	if err := c.setUp(ctx, cctx.opts, cctx.TemplateVars); err != nil {
		if cctx.opts.CleanupOnError {
			cctx.Cleanup()
			cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to scaffold project: %v", err))
			return fmt.Errorf("failed to scaffold project: %w", err)
		}
		cctx.CreatedProjectDir = ""
		return fmt.Errorf("failed to scaffold project: %w", err)
	}
	c.Reporter.Complete("Scaffolding project", "complete!")
	return nil
}

// ensureRepoIfNeeded creates the GitHub repository if required and updates the context and summary.
func (c *Factory) ensureRepoIfNeeded(ctx context.Context, cctx *creationContext) error {
	if cctx.opts.SkipRepo {
		c.Reporter.Skip("Creating GitHub repository", "Skipped as per user request")
		return nil
	}

	c.Reporter.Start("Creating GitHub repository", "")
	_, err := c.createRepository(ctx, cctx.opts)
	if err != nil {
		if strings.Contains(err.Error(), "name already exists on this account") {
			c.Reporter.Skip("Creating GitHub repository", "Repository already exists")
		} else {
			c.Reporter.Failed("Creating GitHub repository", err, "")
			if cctx.opts.CleanupOnError {
				cctx.Cleanup()
				cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to create repository: %v", err))
				return fmt.Errorf("failed to create repository: %w", err)
			}
			return fmt.Errorf("failed to create repository: %w", err)
		}
		return nil
	}

	cctx.CreatedRepo = true
	c.Reporter.Complete("Creating GitHub repository", "")

	// Set the repository URL in the summary
	cctx.Summary.RepoURL = fmt.Sprintf("https://github.com/%s/%s", cctx.opts.RepoOwner, cctx.KebabName)
	return nil
}

// configureGitHub applies GitHub configurations with appropriate reporting and cleanup.
func (c *Factory) configureGitHub(ctx context.Context, cctx *creationContext) error {
	c.Reporter.Start("Applying GitHub configurations", "")
	if err := c.applyConfigurations(ctx, cctx.opts); err != nil {
		if cctx.opts.CleanupOnError {
			cctx.Cleanup()
			cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
			return fmt.Errorf("failed to apply GitHub configurations: %w", err)
		}
		cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to apply GitHub configurations: %v", err))
	} else {
		c.Reporter.Complete("Applying GitHub configurations", "")
	}
	return nil
}

// runHooksAndPublish initialises the local git repository (if applicable), runs post-installation
// hooks, and finally commits/pushes to GitHub. Git is initialised before hooks run so tools that
// expect a git repository (e.g. lefthook, commitlint git-hook wiring) don't emit spurious warnings.
func (c *Factory) runHooksAndPublish(ctx context.Context, cctx *creationContext) {
	willPublish := cctx.opts.Scaffold && !cctx.opts.SkipRepo && cctx.CreatedRepo

	// Initialise the local git repository first (if we're going to publish) so that any
	// post-installation hooks which wire up git hooks (lefthook, commitlint, etc.) find a
	// valid .git directory instead of warning that scaffolding isn't yet a git repository.
	if willPublish {
		c.Reporter.Start("Initializing Git repository", "")
		if err := git.InitRepository(ctx, cctx.ProjectPath); err != nil {
			c.Reporter.Failed("Git repository initialization", err, "")
			cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to initialize Git repository: %v", err))
			willPublish = false
		} else {
			c.Reporter.Complete("Git repository initialization", "")
		}
	}

	// Run post-installation hooks if scaffolding was done and hooks aren't skipped
	if cctx.opts.Scaffold && !cctx.opts.SkipHooks { //nolint:nestif
		c.Reporter.Start("Running post-installation hooks \n", "")
		if err := c.RunHooks(ctx, cctx.ProjectPath, cctx.opts.Language, cctx.opts.ProjectType,
			cctx.TemplateVars); err != nil {
			cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to run post-installation hooks: %v", err))
		} else {
			c.Reporter.Complete("Running post-installation hooks", "complete!")
		}
	} else if cctx.opts.SkipHooks {
		c.Reporter.Skip("Running post-installation hooks", "Skipped as per user request")
	}

	// Commit and push to GitHub repository if both scaffolding is done and repo was created
	if willPublish {
		c.Reporter.Start("Committing and pushing to GitHub", "")

		sshURL := fmt.Sprintf("git@github.com:%s/%s.git", cctx.opts.RepoOwner, cctx.KebabName)
		c.Reporter.Debug(fmt.Sprintf("Using SSH URL for Git operations: %s", sshURL))

		if err := c.Publish(ctx, cctx.ProjectPath, sshURL); err != nil {
			c.Reporter.Failed("Git repository initialization", err, "")
			cctx.Summary.Errors = append(cctx.Summary.Errors, fmt.Sprintf("Failed to initialize and push to Git repository: %v", err))
		} else {
			c.Reporter.Complete("Git repository initialization", "Successfully pushed project to GitHub")
		}
	} else if cctx.opts.SkipRepo {
		c.Reporter.Skip("Git repository initialization", "Skipped as no GitHub repository was created")
	} else if !cctx.opts.Scaffold {
		c.Reporter.Skip("Git repository initialization", "Skipped as no local project was scaffolded")
	}
}

// Publish commits all files in the project directory, adds the remote, and pushes to it.
// The git repository itself must already be initialised (see runHooksAndPublish) before calling this.
func (c *Factory) Publish(ctx context.Context, projectPath, repoURL string) error {
	if err := git.CommitAll(ctx, projectPath, "chore: initial commit"); err != nil {
		return fmt.Errorf("failed to commit files: %w", err)
	}
	if err := git.AddRemote(ctx, projectPath, "origin", repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}
	branch := "main"
	if err := git.Push(ctx, projectPath, "origin", branch); err != nil {
		fmt.Println("Push to 'main' failed, trying 'master' branch...")
		if err := git.Push(ctx, projectPath, "origin", "master"); err != nil {
			return fmt.Errorf("failed to push to remote: %w", err)
		}
	}
	return nil
}

// setUp scaffolds a project locally with pre-populated template variables
func (c *Factory) setUp(ctx context.Context, opts Options, templateVars *template.Variables) error {
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
		templatef, err := c.TemplateRegistry.GetTemplate(opts.Language, opts.ProjectType)
		if err != nil {
			return fmt.Errorf("failed to find template for %s/%s: %w", opts.Language, opts.ProjectType, err)
		}
		templateSource = templatef.Source
	}

	if err := c.Scaffolder.ScaffoldProjectWithOptions(ctx, outputDir, opts.Language, opts.ProjectType, templateSource, templateVars,
		opts.SkipHooks, opts.NoCache, opts.TemplateSet, opts.NoInput); err != nil {
		return fmt.Errorf("failed to scaffold project: %w", err)
	}

	return nil
}

// optsToTemplateVars converts project creation options to template variables
func (c *Factory) optsToTemplateVars(ctx context.Context, opts Options) *template.Variables {
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

	// Only set organization team if this is an organization repository
	if opts.Team != "" && opts.AccountType == OrganizationAccount {
		// Add the team and organization details
		vars.Org.Team = opts.Team
		vars.Org.Name = opts.RepoOwner // Set organization name

		// Try to fetch the team ID
		c.Reporter.Debug(fmt.Sprintf("Attempting to fetch team ID for '%s' in org '%s'", opts.Team, opts.RepoOwner))
		teamID, err := c.GitHubClient.GetTeamID(ctx, opts.RepoOwner, opts.Team)
		if err != nil {
			c.Reporter.Warning("Team ID", fmt.Sprintf("Could not fetch team ID: %v", err))
		} else {
			vars.Org.TeamID = teamID
			c.Reporter.Debug(fmt.Sprintf("Successfully fetched team ID %d for team '%s'", teamID, opts.Team))
		}
	}

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

// valueOrEmpty returns the value or a default value if empty
func valueOrEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
