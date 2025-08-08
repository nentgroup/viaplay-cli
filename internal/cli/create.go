// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/progress"
	"github.com/nentgroup/viaplay-cli/internal/project"
)

// CreateCommandOptions contains all the options for create commands
type CreateCommandOptions struct {
	// Repository options
	RepoName    string
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
	SkipHooks      bool
	NoCache        bool // Force template cache update
}

// createCmd is the parent command for all creation operations
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create projects and repositories",
	Long: `Create projects and repositories with GitHub integration.

This command provides two main creation paths:
  - 'create project': Create a full project with scaffolding and a GitHub repository
  - 'create repo': Create only a GitHub repository without code scaffolding

Both commands support applying organization settings like environments, 
rulesets, and secrets from team configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// NewCreateCommand returns a new create command
func NewCreateCommand() *cobra.Command {
	// Create project command
	projectCmd := newProjectCommand()

	// Create repo command
	repoCmd := newRepoCommand()

	// Add subcommands to the create command
	createCmd.AddCommand(projectCmd)
	createCmd.AddCommand(repoCmd)

	return createCmd
}

// newProjectCommand creates a new project command
func newProjectCommand() *cobra.Command {
	opts := &CreateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create a new project with scaffolding and GitHub repository",
		Long: `Create a new project with code scaffolding and GitHub repository.

This command:
1. Creates a GitHub repository
2. Applies organization settings (environments, rulesets, secrets)
3. Scaffolds a new project from templates
4. Optionally clones the project locally

Use this for a complete project setup experience.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProjectOrRepo(opts, true)
		},
	}

	// Add common flags
	addCommonFlags(cmd, opts)

	// Add project-specific flags
	cmd.Flags().StringVar(&opts.Language, "language", "", "Programming language (go, typescript, etc.)")
	cmd.Flags().StringVar(&opts.ProjectType, "type", "", "Project type (service, cli, lambda, etc.)")
	cmd.Flags().StringVar(&opts.TemplateSource, "template-source", "", "Custom template source")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Directory to create the project in (defaults to current dir + repo name)")
	cmd.Flags().StringVar(&opts.BinaryName, "binary-name", "", "Name of the compiled binary (for compiled languages like Go and Rust)")
	cmd.Flags().BoolVar(&opts.NoRepo, "no-repo", false, "Skip GitHub repository creation (local project only)")
	cmd.Flags().BoolVar(&opts.SkipHooks, "skip-hooks", false, "Skip running post-installation hooks")
	cmd.Flags().BoolVar(&opts.NoCache, "no-cache", false, "Force update of template cache (ignore cached templates)")

	return cmd
}

// newRepoCommand creates a new repo command
func newRepoCommand() *cobra.Command {
	opts := &CreateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Create a GitHub repository without code scaffolding",
		Long: `Create a GitHub repository without code scaffolding.

This command:
1. Creates a GitHub repository
2. Applies organization settings (environments, rulesets, secrets)

Use this when you need to create a repository structure but will add code 
manually or migrate existing code to a new repository.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProjectOrRepo(opts, false)
		},
	}

	// Add common flags
	addCommonFlags(cmd, opts)

	return cmd
}

// addCommonFlags adds common flags to a command
func addCommonFlags(cmd *cobra.Command, opts *CreateCommandOptions) {
	// Repository flags
	cmd.Flags().StringVar(&opts.RepoName, "name", "", "Repository name (required)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&opts.Public, "public", false, "Create a public repository (default is private)")

	// Team flags
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name to use for configs (overrides default)")
	cmd.Flags().BoolVar(&opts.ApplyEnvs, "apply-envs", false, "Apply environment configs from team settings")
	cmd.Flags().BoolVar(&opts.ApplyRulesets, "apply-rulesets", false, "Apply ruleset configs from team settings")
	cmd.Flags().BoolVar(&opts.ApplySecrets, "apply-secrets", false, "Apply secret configs from team settings")

	// Secret flags
	cmd.Flags().StringVar(&opts.RepoSecrets, "secrets", "", "JSON string with repository-specific secrets")
	cmd.Flags().StringVar(&opts.SecretsFile, "secrets-file", "", "Path to JSON file with repository-specific secrets")

	// Mark required flags
	if err := cmd.MarkFlagRequired("name"); err != nil {
		fmt.Printf("Failed to mark 'name' flag as required: %v\n", err)
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
		return err
	}

	// Handle secrets data
	secretsData, err := getSecretsData(opts)
	if err != nil {
		return err
	}

	// Validate project directory if we're scaffolding
	if withScaffolding {
		projectDir, err := validateProjectDirectory(opts, repoParams.name, withScaffolding)
		if err != nil {
			return err
		}
		// Make sure OutputDir is set for the project creation
		if opts.OutputDir == "" {
			opts.OutputDir = filepath.Dir(projectDir)
		}
	}

	// Check if repository exists (if we're creating one)
	if !opts.NoRepo && !validateRepositoryDoesNotExist(ghClient, repoParams.owner, repoParams.name) {
		return fmt.Errorf("repository already exists: %s/%s", repoParams.owner, repoParams.name)
	}

	// Create the project using the Creator
	summary, err := executeProjectCreation(ghClient, configDir, repoParams, opts, secretsData, withScaffolding)
	if err != nil {
		return err
	}

	// Calculate total execution time
	executionTime := time.Since(startTime)

	// Print the project summary with execution time
	printProjectSummary(summary, executionTime)

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

	// Get repo owner from config
	params.owner = viper.GetString("default_account")
	if params.owner == "" {
		return params, fmt.Errorf("repository owner is required (set default_account in config)")
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
func validateProjectDirectory(opts *CreateCommandOptions, repoName string, withScaffolding bool) (string, error) {
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

	// Check if the project directory already exists when scaffolding
	if withScaffolding {
		if _, err := os.Stat(projectDir); err == nil {
			return "", fmt.Errorf("project directory already exists: %s", projectDir)
		}
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
	creator := project.NewCreatorWithReporter(
		ghClient,
		configDir,
		progress.NewCallbackReporter(progress.DefaultCB, viper.GetBool("verbose")),
	)

	// Set up options
	projectOpts := project.CreateOptions{
		// Repository options
		RepoName:        params.name,
		RepoDescription: params.description,
		RepoOwner:       params.owner,
		IsPrivate:       !opts.Public, // Convert public flag to private flag
		IsOrg:           viper.GetBool("is_org"),
		SkipRepo:        opts.NoRepo, // Skip GitHub repository creation if --no-repo is set

		// Project options
		Language:    opts.Language,
		ProjectType: opts.ProjectType,
		Team:        params.team,
		BinaryName:  opts.BinaryName, // Set the binary name from flag
		SkipHooks:   opts.SkipHooks,  // Skip running post-installation hooks if flag is set

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
	fmt.Println("\n📋 Project Summary:")
	fmt.Println("-------------------")

	fmt.Printf("📁 Project location: %s\n", summary.ProjectPath)

	if summary.RepoURL != "" {
		fmt.Printf("🔗 Repository URL: %s\n", summary.RepoURL)
	}

	fmt.Printf("⚙️  Project type: %s/%s\n", summary.Language, summary.ProjectType)

	if summary.Team != "" {
		fmt.Printf("👥 Team: %s\n", summary.Team)
	}

	if summary.AppliedEnvs {
		fmt.Printf("🌍 Environments: Applied from team configuration\n")
	} else {
		fmt.Printf("🌍 Environments: Default staging environment\n")
	}

	if summary.AppliedRulesets {
		fmt.Printf("🔒 Rulesets: Applied from team configuration\n")
	}

	if summary.AppliedSecrets {
		fmt.Printf("🔑 Secrets: Applied from team configuration\n")
	}

	if summary.CustomSecrets {
		fmt.Printf("🔑 Custom secrets: Applied\n")
	}

	// Print any non-fatal errors that occurred
	if len(summary.Errors) > 0 {
		fmt.Println("\n⚠️ Warnings:")
		for _, err := range summary.Errors {
			fmt.Printf("   - %s\n", err)
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
