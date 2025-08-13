package project

import (
	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/progress"
	"github.com/nentgroup/viaplay-cli/internal/registry"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
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
	Reporter progress.Reporter
}

// Options contains all options for creating a new project
type Options struct {
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
