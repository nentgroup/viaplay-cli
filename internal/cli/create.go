// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/git"

	"github.com/nentgroup/viaplay-cli/internal/config"

	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/project"
)

// CreateCommandOptions contains all the options for create commands
type CreateCommandOptions struct {
	// Repository options
	RepoName    string
	RepoOwner   string // Owner of the repository (user or organization)
	IsOrg       bool   // Deprecated
	Description string
	Public      bool // false = private repo (default)
	RepoSecrets string
	SecretsFile string
	NoRepo      bool

	// Team options
	Team          string
	ApplyEnvs     bool
	ApplyRulesets bool
	ApplySecrets  bool

	// Project-specific options
	Language       string
	ProjectType    string
	TemplateSource string
	OutputDir      string
	BinaryName     string
	NoHooks        bool
	NoCache        bool // Force template cache update

	// Error handling options
	CleanupOnError bool // Clean up resources (delete folder/repo) if errors occur
}

// addCommonFlagsExceptName adds common flags to a command, excluding the name flag
func addCommonFlagsExceptName(cmd *cobra.Command, opts *CreateCommandOptions) {
	// Repository flags
	cmd.Flags().StringVar(&opts.Description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&opts.Public, "public", false, "Create a public repository (overrides --private)")
	cmd.Flags().BoolP("private", "p", false, "Create a private repository (overrides default visibility)")
	cmd.Flags().StringVar(&opts.RepoSecrets, "repo-secrets", "", "JSON string containing repository-specific secrets")
	cmd.Flags().StringVar(&opts.SecretsFile, "secrets-file", "", "Path to a JSON file containing repository-specific secrets")

	// Team/organization flags
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name for loading configuration templates")

	// Define flags without setting Viper defaults at initialization time
	cmd.Flags().BoolVar(&opts.ApplyEnvs, "apply-envs", false, "Apply environments from team configuration")
	cmd.Flags().BoolVar(&opts.ApplyRulesets, "apply-rulesets", false, "Apply rulesets from team configuration")
	cmd.Flags().BoolVar(&opts.ApplySecrets, "apply-secrets", false, "Apply secrets from team configuration")
	cmd.Flags().BoolVar(&opts.CleanupOnError, "cleanup-on-error", false, "Clean up resources on error")

	// Add a PreRun hook to set the defaults from Viper at runtime
	originalPreRun := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Only apply defaults if flag wasn't explicitly set by user
		if !cmd.Flags().Changed("team") {
			opts.Team = viper.GetString("default_team")
		}
		if !cmd.Flags().Changed("apply-envs") {
			opts.ApplyEnvs = viper.GetBool("apply_envs")
		}
		if !cmd.Flags().Changed("apply-rulesets") {
			opts.ApplyRulesets = viper.GetBool("apply_rulesets")
		}
		if !cmd.Flags().Changed("apply-secrets") {
			opts.ApplySecrets = viper.GetBool("apply_secrets")
		}
		if !cmd.Flags().Changed("cleanup-on-error") {
			opts.CleanupOnError = viper.GetBool("cleanup_on_error")
		}

		// Handle repository visibility with priority order:
		// 1. --public flag (highest priority)
		// 2. --private flag (second priority)
		// 3. default_visibility from config (lowest priority)
		isPrivateSet, _ := cmd.Flags().GetBool("private")

		// If neither flag is explicitly set, use the default_visibility from config
		if !cmd.Flags().Changed("public") && !cmd.Flags().Changed("private") {
			visibility := viper.GetString("default_visibility")
			opts.Public = visibility == "public"
		} else if cmd.Flags().Changed("public") && opts.Public {
			// --public is set to true, which takes precedence
			opts.Public = true
		} else if cmd.Flags().Changed("private") && isPrivateSet {
			// --private is set to true, make Public = false
			opts.Public = false
		}
		// In case of conflict (both flags set), --public takes precedence

		// Add the project-specific flag defaults from Viper
		if !cmd.Flags().Changed("no-repo") {
			opts.NoRepo = viper.GetBool("no_repo")
		}
		if !cmd.Flags().Changed("no-hooks") {
			opts.NoHooks = viper.GetBool("no_hooks")
		}
		if !cmd.Flags().Changed("no-cache") {
			opts.NoCache = viper.GetBool("no_cache")
		}

		// Run the original PreRun if it exists
		if originalPreRun != nil {
			return originalPreRun(cmd, args)
		}
		return nil
	}
}

// createProjectOrRepo is a shared function that handles both project and repo creation
// The withScaffolding parameter determines whether to include scaffolding
func createProjectOrRepo(opts *CreateCommandOptions, withScaffolding bool) error {
	// Start timing the operation
	startTime := time.Now()
	// Setup GitHub client and get config
	ghClient, configDir, err := setupGitHubClient()
	if err != nil {
		return err
	}

	// Validate and prepare repository parameters
	repoParams, err := validateRepoParameters(opts, withScaffolding)
	if err != nil {
		output.FatalError(fmt.Sprintf("Repo parameters validation failed: %v", err))
		return nil
	}

	// Handle secrets data
	secretsData, err := getSecretsData(opts)
	if err != nil {
		return err
	}

	// Validate project directory
	projectDir, err := validateProjectDirectory(opts, repoParams.name)
	if err != nil {
		output.FatalError(fmt.Sprintf("Project directory validation failed: %v", err))
		return nil
	}
	// Make sure OutputDir is set for the project creation
	if opts.OutputDir == "" {
		opts.OutputDir = filepath.Dir(projectDir)
	}

	// Check if repository exists (if we're creating one)
	if !opts.NoRepo && !validateRepositoryDoesNotExist(ghClient, repoParams.owner, repoParams.name) {
		output.FatalError(fmt.Sprintf("Repository already exists: %s/%s", repoParams.owner, repoParams.name))
		return nil
	}

	// Debug info about command options when in verbose mode
	if viper.GetBool("verbose") {
		output.VerboseMessage(fmt.Sprintf("Command options: name=%s, language=%s, type=%s, team=%s",
			opts.RepoName, opts.Language, opts.ProjectType, opts.Team))
		output.VerboseMessage(fmt.Sprintf("Config settings: apply-envs=%t, apply-rulesets=%t, apply-secrets=%t, no-repo=%t",
			opts.ApplyEnvs, opts.ApplyRulesets, opts.ApplySecrets, opts.NoRepo))
		output.VerboseMessage(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	// Create the project using the Factory
	summary, createErr := executeProjectCreation(ghClient, configDir, repoParams, opts, secretsData, withScaffolding)

	// Calculate total execution time
	executionTime := time.Since(startTime)

	// Always print the summary if we have one, even if there was an error
	if summary != nil {
		// Print the summary
		printProjectSummary(summary, executionTime)

		// If resources were cleaned up due to an error, and cleanup was successful,
		// we should consider this a successful operation (exit code 0)
		if summary.CleanedUp && opts.CleanupOnError {
			// Return nil to indicate success (resources were cleaned up properly)
			return nil
		}
	}

	// For repo-only creation (no scaffolding) we want to initialise a local git repo
	// that points to the newly created GitHub repository.
	if !withScaffolding && !opts.NoRepo {
		if err := initLocalRepo(projectDir, repoParams); err != nil {
			return err
		}
	}

	// Return any error that occurred during creation
	if createErr != nil {
		return createErr
	}

	return nil
}

// setupGitHubClient handles GitHub authentication and client setup
func setupGitHubClient() (*gh.GitHubClient, string, error) {
	// Authenticate with GitHub
	output.VerboseMessage("Authenticating with GitHub...")
	token, err := gh.Authenticate()
	if err != nil {
		output.ErrorMessage("GitHub authentication failed")
		return nil, "", fmt.Errorf("GitHub authentication failed: %w", err)
	}
	output.VerboseMessage("GitHub authentication successful!")

	// Initialise GitHub client
	ghClient := gh.NewGitHubClient(token)

	// Get config directory
	configDir := viper.GetString("config_dir")

	return ghClient, configDir, nil
}

// repoParameters holds validated repository parameters
type repoParameters struct {
	name        string
	description string
	owner       string
	team        string
}

// validateRepoParameters validates and collects repository parameters from flags and config
func validateRepoParameters(opts *CreateCommandOptions, withScaffolding bool) (repoParameters, error) {
	var params repoParameters

	// Use RepoOwner from opts if provided, otherwise determine from config
	if opts.RepoOwner != "" {
		// Owner explicitly specified in command, use it directly
		params.owner = opts.RepoOwner
		output.VerboseMessage(fmt.Sprintf("Using specified repository owner: '%s'", params.owner))
	} else {
		// No owner specified, determine based on github.organization setting
		orgName := viper.GetString("github.organization")
		username := viper.GetString("github.username")

		// Use organization if set, otherwise fall back to personal username
		if orgName != "" {
			params.owner = orgName
			output.VerboseMessage(fmt.Sprintf("Using organization from config: '%s'", params.owner))
		} else if username != "" {
			params.owner = username
			output.VerboseMessage(fmt.Sprintf("Using username from config: '%s'", params.owner))
		} else {
			return params, fmt.Errorf("repository owner is required (use --owner flag or configure github.username/github.organization in config)")
		}
	}

	// Get repo name from flag
	params.name = opts.RepoName
	if params.name == "" {
		return params, fmt.Errorf("repository name is required (use --name flag)")
	}

	// Get or set description
	params.description = opts.Description
	if params.description == "" {
		params.description = fmt.Sprintf("Repository for %s", params.name)
	}

	// Get team name
	params.team = opts.Team
	if params.team == "" {
		params.team = viper.GetString("default_team")
	}

	// For project creation, ensure language and project type are set
	if withScaffolding {
		if err := setupScaffoldingOptions(opts); err != nil {
			return params, err
		}
	}

	return params, nil
}

// getSecretsData handles secrets data from flag or file
func getSecretsData(opts *CreateCommandOptions) (string, error) {
	secretsData := opts.RepoSecrets
	if opts.SecretsFile != "" {
		data, err := os.ReadFile(opts.SecretsFile)
		if err != nil {
			return "", fmt.Errorf("failed to read secrets file: %w", err)
		}
		secretsData = string(data)
	}
	return secretsData, nil
}

// validateProjectDirectory checks if the project directory is valid
func validateProjectDirectory(opts *CreateCommandOptions, repoName string) (string, error) {
	projectDir := opts.OutputDir
	if projectDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		projectDir = filepath.Join(currentDir, repoName)
	} else {
		projectDir = filepath.Join(projectDir, repoName)
	}

	// Always ensure the project directory does not already exist
	if _, err := os.Stat(projectDir); err == nil {
		return "", fmt.Errorf("project directory already exists: %s", projectDir)
	}

	return projectDir, nil
}

// validateRepositoryDoesNotExist checks if the repository doesn't exist on GitHub
func validateRepositoryDoesNotExist(ghClient *gh.GitHubClient, owner, repoName string) bool {
	repoExists, err := ghClient.RepositoryExists(owner, repoName)
	if err != nil {
		output.VerboseMessage(fmt.Sprintf("Error checking if repository exists: %v", err))
		return false
	}
	return !repoExists
}

// executeProjectCreation executes the project creation workflow
func executeProjectCreation(ghClient *gh.GitHubClient, configDir string, params repoParameters, opts *CreateCommandOptions, secretsData string, withScaffolding bool) (*project.Summary, error) { // Create project creator with reporter
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	creator := project.NewFactory(ghClient, output.NewCallbackReporter(output.DefaultCB, viper.GetBool("verbose")), cfg)

	// Set up options
	projectOpts := project.Options{
		// Repository options
		RepoName:        params.name,
		RepoDescription: params.description,
		RepoOwner:       params.owner,
		IsPrivate:       !opts.Public, // Convert public flag to private flag
		IsOrg:           opts.IsOrg,   // Set isOrg based on our determination
		SkipRepo:        opts.NoRepo,  // Skip GitHub repository creation if --no-repo is set

		// Project options
		Language:    opts.Language,
		ProjectType: opts.ProjectType,
		Team:        params.team,
		BinaryName:  opts.BinaryName, // Set the binary name from flag
		SkipHooks:   opts.NoHooks,    // Skip running post-installation hooks if flag is set

		// Configuration options
		ConfigDir: configDir,
		// If no-repo is set, we should skip applying environments, rulesets and secrets as they only make sense with a repo
		ApplyEnvs:     opts.ApplyEnvs && !opts.NoRepo,
		ApplyRulesets: opts.ApplyRulesets && !opts.NoRepo,
		ApplySecrets:  opts.ApplySecrets && !opts.NoRepo,
		RepoSecrets:   secretsData,

		// Template options (only used if withScaffolding is true)
		TemplateSource: opts.TemplateSource,
		Scaffold:       withScaffolding, // Always true for project, false for repo
		OutputDir:      opts.OutputDir,
		NoCache:        opts.NoCache, // Force update of template cache if flag is set

		// Error handling options
		CleanupOnError: opts.CleanupOnError, // Pass the cleanup flag to the creator
	}

	// Execute the project creation workflow
	return creator.Create(projectOpts)
}

// formatDuration formats a duration to be more human-readable
func formatDuration(d time.Duration) string {
	// For very short durations (less than a second), show milliseconds
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds()))
	}

	// For durations between 1 second and 1 minute
	if d < time.Minute {
		seconds := d.Seconds()
		return fmt.Sprintf("%.2f seconds", seconds)
	}

	// For longer durations, use minutes and seconds
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if seconds == 0 {
		return fmt.Sprintf("%d minute%s", minutes, pluralS(minutes))
	}
	return fmt.Sprintf("%d minute%s %d second%s",
		minutes, pluralS(minutes),
		seconds, pluralS(seconds))
}

// pluralS returns "s" if count is not 1, otherwise empty string
func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// printProjectSummary prints the project creation summary in a nice format
func printProjectSummary(summary *project.Summary, executionTime time.Duration) {
	fmt.Printf("\n%s %s\n", output.ActiveIcons.Summary, output.Bold("Project Summary:"))
	fmt.Println(output.Faint(strings.Repeat("─", 20)))

	// If resources were cleaned up due to an error, show that first and prominently
	if summary.CleanedUp {
		fmt.Printf("%s %s\n", output.ActiveIcons.Warning, output.WarningBold("Project creation failed and resources were cleaned up:"))
		for _, detail := range summary.CleanupDetails {
			fmt.Printf("   %s %s\n", output.ActiveIcons.Bullet, detail)
		}
		fmt.Println(output.Faint(strings.Repeat("─", 20)))

		// If we have any errors that triggered the cleanup, show them
		if len(summary.Errors) > 0 {
			fmt.Printf("%s %s\n", output.ActiveIcons.Error, output.ErrorBold("Errors that caused cleanup:"))
			for _, err := range summary.Errors {
				fmt.Printf("   %s %s\n", output.ActiveIcons.Bullet, err)
			}
			fmt.Println(output.Faint(strings.Repeat("─", 20)))
		}

		fmt.Printf("\nProject creation failed but all resources were cleaned up in %s\n", formatDuration(executionTime))
		return
	}

	// Only show these if we didn't clean up (i.e., project creation was successful)
	fmt.Printf("%s Project location: %s\n", output.ActiveIcons.Template, summary.ProjectPath)

	if summary.RepoURL != "" {
		fmt.Printf("%s Repository URL: %s\n", output.ActiveIcons.GitHub, summary.RepoURL)
	}

	fmt.Printf("%s Project type: %s/%s\n", output.ActiveIcons.Config, summary.Language, summary.ProjectType)

	if summary.Team != "" {
		fmt.Printf("%s Team: %s\n", output.ActiveIcons.People, summary.Team)
	}

	if summary.AppliedEnvs {
		fmt.Printf("%s Environments: Applied from team configuration\n", output.ActiveIcons.Globe)
	} else {
		fmt.Printf("%s Environments: Default staging environment\n", output.ActiveIcons.Globe)
	}

	if summary.AppliedRulesets {
		fmt.Printf("%s Rulesets: Applied from team configuration\n", output.ActiveIcons.Lock)
	}

	if summary.AppliedSecrets {
		fmt.Printf("%s Secrets: Applied from team configuration\n", output.ActiveIcons.Key)
	}

	if summary.CustomSecrets {
		fmt.Printf("%s Custom secrets: Applied\n", output.ActiveIcons.Key)
	}

	// Print any non-fatal errors that occurred during successful creation
	if len(summary.Errors) > 0 {
		fmt.Printf("\n%s %s\n", output.ActiveIcons.Warning, output.WarningBold("Warnings:"))
		for _, err := range summary.Errors {
			fmt.Printf("   %s %s\n", output.ActiveIcons.Bullet, err)
		}
	}

	fmt.Printf("\nProject creation complete in %s\n", formatDuration(executionTime))
}

// Helper to set up project scaffolding options
func setupScaffoldingOptions(opts *CreateCommandOptions) error {
	if opts.Language == "" {
		opts.Language = viper.GetString("default_language")
		if opts.Language == "" {
			opts.Language = "go"
		}
	}
	if opts.ProjectType == "" {
		opts.ProjectType = viper.GetString("default_type")
		if opts.ProjectType == "" {
			opts.ProjectType = "service"
		}
	}
	if opts.TemplateSource == "" {
		templateKey := fmt.Sprintf("templates.%s.%s.source", opts.Language, opts.ProjectType)
		opts.TemplateSource = viper.GetString(templateKey)
		if opts.TemplateSource == "" {
			return fmt.Errorf("no template found for %s/%s, please specify with --template-source", opts.Language, opts.ProjectType)
		}
	}
	return nil
}

// initLocalRepo initialises a local git repository in projectDir and sets origin to the new GitHub repo.
func initLocalRepo(projectDir string, params repoParameters) error {
	// Initialise git repository (this will create the directory if needed)
	if err := git.InitRepository(projectDir); err != nil {
		return fmt.Errorf("failed to initialise local git repository at %s: %w", projectDir, err)
	}

	// Configure origin remote to point at the new GitHub repository using SSH URL
	repoURL := fmt.Sprintf("git@github.com:%s/%s.git", params.owner, params.name)
	if err := git.AddRemote(projectDir, "origin", repoURL); err != nil {
		return fmt.Errorf("failed to configure origin remote for local repository at %s: %w", projectDir, err)
	}

	return nil
}
