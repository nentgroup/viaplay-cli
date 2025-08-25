package project

import (
	"github.com/google/go-github/v74/github"
	"github.com/nentgroup/viaplay-cli/internal/output"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
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

// Options contains all options for creating a new project
type Options struct {
	// Repository options
	RepoName        string
	RepoDescription string
	RepoOwner       string
	IsPrivate       bool
	IsOrg           bool   // Deprecated: use AccountType instead
	AccountType     string // "personal" or "organization"
	SkipRepo        bool   // Skip GitHub repository creation

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

// EnvConf represents the configuration for a GitHub environment.
// Notice that it contains Also BranchPatterns for convenience, nut shall not be confused with
// github.CreateUpdateEnvironment which is used to create environments in GitHub.
type EnvConf struct {
	Name                   string                 `json:"name"`
	WaitTimer              int                    `json:"wait_timer,omitempty"`
	PreventSelfReview      bool                   `json:"prevent_self_review,omitempty"`
	DeploymentBranchPolicy *CustomBranchPolicy    `json:"deployment_branch_policy,omitempty"`
	Reviewers              []*github.EnvReviewers `json:"reviewers,omitempty"`
}

type CustomBranchPolicy struct {
	ProtectedBranches    bool                                    `json:"protected_branches,omitempty"`
	CustomBranchPolicies bool                                    `json:"custom_branch_policies,omitempty"`
	BranchPatterns       []*github.DeploymentBranchPolicyRequest `json:"branch_patterns,omitempty"`
}

func (e EnvConf) ToGitHubEnv() *github.CreateUpdateEnvironment {
	env := &github.CreateUpdateEnvironment{
		WaitTimer:         github.Ptr(e.WaitTimer),
		PreventSelfReview: github.Ptr(e.PreventSelfReview),
		Reviewers:         e.Reviewers,
	}

	if e.DeploymentBranchPolicy != nil {
		env.DeploymentBranchPolicy = &github.BranchPolicy{
			ProtectedBranches:    github.Ptr(e.DeploymentBranchPolicy.ProtectedBranches),
			CustomBranchPolicies: github.Ptr(e.DeploymentBranchPolicy.CustomBranchPolicies),
		}
	}

	return env
}
