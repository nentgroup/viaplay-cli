// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/project"
	secretspkg "github.com/nentgroup/viaplay-cli/internal/secrets"
	templatepkg "github.com/nentgroup/viaplay-cli/internal/template"
)

// Use the default paths from the config package instead of maintaining duplicates
var (
	defaultConfigDir  string
	defaultConfigFile string
	defaultTeamsDir   string
)

const (
	configTargetMain = "main"
	configTargetTeam = "team"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage viaplay-cli configuration and team settings",
	Long: `Manage global and team-specific configuration for viaplay-cli.

- View CLI settings
- Scaffold local team config folders and example YAML files
- Configure a shared read-only source for hooks and team config
- Pull shared hooks and team config into your local config directory

Examples:
  vip config init
  vip config init team myteam --organization nentgroup
  vip config init source --organization nentgroup
  vip config pull
  vip config get default_team
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// initCmd scaffolds example config files for a team (optionally)
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize viaplay-cli configuration",
	Long: `Initialize the viaplay-cli configuration directory and config.yaml.
Optionally scaffold team config files with --team <team> and --organization <organization>.

This command will:
1. Create the main config.yaml with settings from your authenticated GitHub account
2. Set up personal account configurations in ~/.config/viaplay/users/
3. Set up team configurations when specified with --team flag
4. Auto-configure the conventional shared source repo when available

When run without flags, it will guide you through an interactive selection of organization and team.

Examples:
  vip config init                          # Initialize config using your GitHub account
  vip config init --team platform          # Initialize with team config
  vip config init --team platform --organization myorg
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Check if the user is authenticated with GitHub
		token, err := gh.GetToken()
		if err != nil || token == "" {
			return fmt.Errorf("you must be authenticated with GitHub to initialize viaplay-cli configuration. Run 'vip auth login' first")
		}

		team, err := cmd.Flags().GetString("team")
		if err != nil {
			return fmt.Errorf("failed to get 'team' flag: %w", err)
		}

		organization, err := cmd.Flags().GetString("organization")
		if err != nil {
			return fmt.Errorf("failed to get 'org' flag: %w", err)
		}

		override, err := cmd.Flags().GetBool("override")
		if err != nil {
			return fmt.Errorf("failed to get 'override' flag: %w", err)
		}

		// Create GitHub client
		ghClient := gh.NewGitHubClient(token)

		// Get authenticated user info
		username, err := ghClient.GetAuthenticatedUser(ctx)
		if err != nil {
			return fmt.Errorf("failed to get authenticated user: %w", err)
		}
		fmt.Printf("Authenticated as: %s\n", output.Bold(username))

		// Interactive organisation and team selection if neither team nor organisation flags are set
		if !cmd.Flags().Changed("team") && !cmd.Flags().Changed("organization") {
			// Interactive selection of organization
			selectedOrg, err := SelectOrganizationWithBubbles(ctx, ghClient)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				// Continue without organization if there's an error
			} else if selectedOrg != "" {
				organization = selectedOrg

				// If we have an organization, also select a team
				selectedTeam, err := SelectTeamWithBubbles(ctx, ghClient, organization)
				if err != nil {
					fmt.Printf("Warning: %v\n", err)
					// Continue without team if there's an error
				} else if selectedTeam != "" {
					team = selectedTeam
				}
			}
		}

		// Set up team configurations if specified
		if team != "" && organization != "" {
			if err := scaffoldTeamConfig(organization, team, override); err != nil {
				return err
			}
		}

		// Always set up personal configurations for the authenticated user
		if err := scaffoldPersonalConfig(username, override); err != nil {
			return err
		}

		if err := initializeConfigFile(organization, team, override); err != nil {
			return fmt.Errorf("failed to initialize config file: %w", err)
		}

		syncSharedSourceDuringInit(ctx, configuredConfigFilePath(), team, organization)

		// Show success message and next steps
		showInitSuccessMessage(team, organization)

		return nil
	},
}

// initTeamCmd scaffolds a team configuration folder and starter files.
var initTeamCmd = &cobra.Command{
	Use:   "team <name>",
	Short: "Scaffold a team configuration folder",
	Long: `Scaffold a team configuration folder with starter secrets, environments,
and rulesets files.

This command creates the standard team layout under the viaplay config directory
without requiring GitHub authentication.

Examples:
  vip config init team gecko --organization nentgroup
  vip config init team platform --override
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		team := args[0]

		override, err := cmd.Flags().GetBool("override")
		if err != nil {
			return fmt.Errorf("failed to get 'override' flag: %w", err)
		}

		organization, err := cmd.Flags().GetString("organization")
		if err != nil {
			return fmt.Errorf("failed to get 'organization' flag: %w", err)
		}

		if organization == "" {
			if _, statErr := os.Stat(configuredConfigFilePath()); statErr == nil {
				if err := config.InitConfig(configuredConfigFilePath()); err != nil {
					return fmt.Errorf("failed to initialize config: %w", err)
				}
				organization = viper.GetString("default_organization")
			} else if statErr != nil && !os.IsNotExist(statErr) {
				return fmt.Errorf("failed to stat config file: %w", statErr)
			}
		}
		if organization == "" {
			return fmt.Errorf("organization is required (use --organization or configure default_organization)")
		}

		if err := scaffoldTeamConfig(organization, team, override); err != nil {
			return err
		}

		teamDir := config.GetDefaultConfigDir()
		teamPath := fmt.Sprintf("%s/orgs/%s/teams/%s", teamDir, organization, team)

		fmt.Printf("\n%s %s\n\n", output.ActiveIcons.Success, output.SuccessBold("Team configuration scaffolded successfully!"))
		fmt.Printf("Location: %s\n\n", output.Bold(teamPath))
		fmt.Println("Next steps:")
		fmt.Printf("  %s Review and edit %s\n", output.ActiveIcons.Bullet, output.Bold(teamPath))
		fmt.Printf("  %s Set your default team with %s\n", output.ActiveIcons.Bullet, output.Bold("vip config init --team "+team+" --organization "+organization))
		fmt.Printf("  %s Apply it to a repo with %s\n", output.ActiveIcons.Bullet, output.Bold("vip repo apply <owner/repo> --team "+team))
		fmt.Println()

		return nil
	},
}

// initSourceCmd configures the shared config source repository.
var initSourceCmd = &cobra.Command{
	Use:          "source",
	Short:        "Configure the shared config source",
	SilenceUsage: true,
	Long: `Configure the read-only repository used to distribute shared team config
and hooks.

If --repository is omitted, vip looks for the conventional repository
<organization>/vip-shared-configs and configures it when found.

Examples:
  vip config init source --organization nentgroup
  vip config init source --repository git@github.com:nentgroup/vip-shared-configs.git
  vip config init source --organization nentgroup --root shared-config
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := configuredConfigFilePath()
		if err := config.InitializeConfigFile(configFile, false, "", ""); err != nil {
			return fmt.Errorf("failed to initialize config file: %w", err)
		}
		if err := config.InitConfig(configFile); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		sourceCfg, err := resolveSourceConfigInput(cmd, cfg)
		if err != nil {
			return err
		}

		if err := config.UpdateSourceConfigFile(configFile, sourceCfg); err != nil {
			return err
		}

		output.SuccessMessage("Shared config source configured")
		fmt.Printf("Repository: %s\n", output.Bold(sourceCfg.Repository))
		fmt.Printf("Branch:     %s\n", output.Bold(sourceCfg.Branch))
		fmt.Printf("Root:       %s\n", output.Bold(sourceCfg.Root))
		fmt.Printf("\nNext: %s\n", output.Bold("vip config pull"))

		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:          "pull",
	Short:        "Pull shared configuration",
	SilenceUsage: true,
	Long: `Pull shared configuration from the configured source.

Without subcommands, this pulls shared hooks and the default team when one is
configured.

Examples:
  vip config pull
  vip config pull team gecko --organization nentgroup
  vip config pull hooks
  vip config pull all
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithSource()
		if err != nil {
			return err
		}

		pulledAnything := false

		if hooksSummary, err := cfg.PullHooks(cmd.Context()); err != nil {
			if isMissingSharedPathError(err) {
				output.InfoMessage("No shared hooks found in the source repository")
			} else {
				return err
			}
		} else {
			printSourcePullSummary("hooks", hooksSummary)
			pulledAnything = true
		}

		teamName := strings.TrimSpace(cfg.DefaultTeam)
		if teamName != "" {
			teamSummary, err := cfg.PullTeam(cmd.Context(), teamName, cfg.DefaultOrganization)
			if err != nil {
				return err
			}
			printSourcePullSummary(fmt.Sprintf("team %s", teamName), teamSummary)
			pulledAnything = true
		}

		if !pulledAnything {
			return fmt.Errorf("nothing to pull: configure default_team or use 'vip config pull team <name>'")
		}

		return nil
	},
}

var pullTeamCmd = &cobra.Command{
	Use:          "team <name>",
	Short:        "Pull one shared team configuration",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithSource()
		if err != nil {
			return err
		}

		organization, err := cmd.Flags().GetString("organization")
		if err != nil {
			return fmt.Errorf("failed to get 'organization' flag: %w", err)
		}
		if organization == "" {
			organization = cfg.DefaultOrganization
		}

		summary, err := cfg.PullTeam(cmd.Context(), args[0], organization)
		if err != nil {
			return err
		}

		printSourcePullSummary(fmt.Sprintf("team %s", args[0]), summary)
		return nil
	},
}

var pullHooksCmd = &cobra.Command{
	Use:          "hooks",
	Short:        "Pull shared hooks",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithSource()
		if err != nil {
			return err
		}

		summary, err := cfg.PullHooks(cmd.Context())
		if err != nil {
			return err
		}

		printSourcePullSummary("hooks", summary)
		return nil
	},
}

var pullAllCmd = &cobra.Command{
	Use:          "all",
	Short:        "Pull all shared hooks and team configs",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithSource()
		if err != nil {
			return err
		}

		summary, err := cfg.PullAll(cmd.Context())
		if err != nil {
			return err
		}

		printSourcePullSummary("all shared config", summary)
		return nil
	},
}

// getCmd gets a configuration value
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long: `Get a configuration value from the config file.
If no key is specified, all configuration values will be displayed.

Examples:
  vip config get default_account
  vip config get default_team
  vip config get           # Show all values
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure we're reading from the config file
		if err := ensureConfigLoaded(); err != nil {
			return err
		}

		// If no key is provided, show all config
		if len(args) == 0 {
			return showAllConfig()
		}

		// Get the specified key
		return getConfigValue(args[0])
	},
}

// pathsCmd shows the configuration paths
var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show configuration paths",
	Long: `Show all configuration paths used by viaplay-cli.

This includes:
- Config file location
- Config directory
- Teams directory
`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfigPaths()
	},
}

// pathCmd resolves a specific configuration path for scripting and quick navigation.
var pathCmd = &cobra.Command{
	Use:   "path <team|user|hooks|templates>",
	Short: "Print a single configuration path",
	Long: `Print a single configuration path for scripting or quick navigation.

Supported path types:
- team: team configuration directory
- user: personal user configuration directory
- hooks: global hooks directory
- templates: template cache directory

Examples:
  vip config path team --team gecko --organization nentgroup
  vip config path user
  vip config path hooks
  vip config path templates
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(defaultConfigFile); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		pathType := strings.ToLower(args[0])
		resolvedPath, err := resolveConfigPath(cmd, cfg, pathType)
		if err != nil {
			return err
		}

		fmt.Println(resolvedPath)
		return nil
	},
}

// editCmd opens a config file or directory in the user's editor.
var editCmd = &cobra.Command{
	Use:          "edit [main|team|user|hooks|templates]",
	Short:        "Open configuration in your editor",
	SilenceUsage: true,
	Long: `Open a viaplay-cli configuration file or directory in your configured editor.

When no target is provided, the main config file is opened.

Supported targets:
- main: main config.yaml file
- team: resolved team configuration directory
- user: resolved personal configuration directory
- hooks: global hooks directory
- templates: template cache directory

Examples:
  vip config edit
  vip config edit team --team gecko --organization nentgroup
  vip config edit hooks
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(configuredConfigFilePath()); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		target := configTargetMain
		if len(args) == 1 {
			target = strings.ToLower(args[0])
		}

		targetPath, err := resolveEditTargetPath(cmd, cfg, target)
		if err != nil {
			return err
		}

		if err := ensureEditableTarget(target, targetPath); err != nil {
			return err
		}

		return openEditor(cmd, targetPath)
	},
}

// validateCmd validates the main config file and optional team/user config directories.
var validateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Validate configuration files",
	SilenceUsage: true,
	Long: `Validate the main viaplay-cli configuration and optionally team or personal
configuration directories.

By default, this validates the active config file. If a default team is configured,
its team directory is validated too. Use flags to validate a specific team, user,
or every discovered team config.

Examples:
  vip config validate
  vip config validate --team gecko --organization nentgroup
  vip config validate --all-teams
  vip config validate --user alecoletti
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		report := &configValidationReport{}

		configFile := configuredConfigFilePath()
		if err := validateMainConfig(configFile, report); err != nil {
			printConfigValidationReport(report)
			return err
		}

		if err := config.InitConfig(configFile); err != nil {
			report.addError(fmt.Sprintf("failed to initialize config: %v", err))
			printConfigValidationReport(report)
			return fmt.Errorf("configuration validation failed")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			report.addError(fmt.Sprintf("failed to load configuration: %v", err))
			printConfigValidationReport(report)
			return fmt.Errorf("configuration validation failed")
		}

		targets, err := collectValidationTargets(cmd, cfg)
		if err != nil {
			return err
		}

		for _, target := range targets {
			validateConfigDirectory(target, report)
		}

		printConfigValidationReport(report)
		if report.hasErrors() {
			return fmt.Errorf("configuration validation failed")
		}

		return nil
	},
}

// Path handling functions

type configValidationTarget struct {
	label string
	path  string
	kind  string
}

type configValidationReport struct {
	checked  []string
	warnings []string
	errors   []string
}

func (r *configValidationReport) addChecked(message string) {
	r.checked = append(r.checked, message)
}

func (r *configValidationReport) addWarning(message string) {
	r.warnings = append(r.warnings, message)
}

func (r *configValidationReport) addError(message string) {
	r.errors = append(r.errors, message)
}

func (r *configValidationReport) hasErrors() bool {
	return len(r.errors) > 0
}

// ensureConfigLoaded makes sure viper has loaded the config file
func ensureConfigLoaded() error {
	configFile := configuredConfigFilePath()

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found, run 'viaplay-cli config init' first")
	}

	// Use the config package's initialization function if viper hasn't loaded a config yet
	if viper.ConfigFileUsed() == "" {
		if err := config.InitConfig(configFile); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
	}

	return nil
}

func configuredConfigFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return defaultConfigFile
}

func loadConfigWithSource() (*config.Configuration, error) {
	configFile := configuredConfigFilePath()
	if err := config.InitConfig(configFile); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	if !cfg.HasSource() {
		return nil, fmt.Errorf("shared config source is not configured (run 'vip config init source')")
	}

	return cfg, nil
}

func resolveSourceConfigInput(cmd *cobra.Command, cfg *config.Configuration) (config.SourceConfig, error) {
	repository, err := cmd.Flags().GetString("repository")
	if err != nil {
		return config.SourceConfig{}, fmt.Errorf("failed to get 'repository' flag: %w", err)
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return config.SourceConfig{}, fmt.Errorf("failed to get 'branch' flag: %w", err)
	}
	root, err := cmd.Flags().GetString("root")
	if err != nil {
		return config.SourceConfig{}, fmt.Errorf("failed to get 'root' flag: %w", err)
	}
	organization, err := cmd.Flags().GetString("organization")
	if err != nil {
		return config.SourceConfig{}, fmt.Errorf("failed to get 'organization' flag: %w", err)
	}

	if organization == "" && cfg != nil {
		organization = cfg.DefaultOrganization
	}

	if repository == "" {
		repository, err = detectSharedSourceRepository(cmd.Context(), organization)
		if err != nil {
			return config.SourceConfig{}, err
		}
	}
	if branch == "" {
		if cfg != nil && cfg.Source.Branch != "" {
			branch = cfg.Source.Branch
		} else {
			branch = config.DefaultSourceBranch
		}
	}
	if root == "" {
		if cfg != nil && cfg.Source.Root != "" {
			root = cfg.Source.Root
		} else {
			root = config.DefaultSourceRoot
		}
	}

	return config.SourceConfig{
		Repository: repository,
		Branch:     branch,
		Root:       root,
	}, nil
}

func detectSharedSourceRepository(ctx context.Context, organization string) (string, error) {
	organization = strings.TrimSpace(organization)
	if organization == "" {
		return "", fmt.Errorf("organization is required when --repository is omitted")
	}

	token, err := gh.GetToken()
	if err != nil || token == "" {
		return "", fmt.Errorf("GitHub authentication is required to detect %s/%s automatically; run 'vip auth login' or pass --repository",
			organization, config.SharedConfigRepoName)
	}

	ghClient := gh.NewGitHubClient(token)
	exists, err := ghClient.RepositoryExists(ctx, organization, config.SharedConfigRepoName)
	if err != nil {
		return "", fmt.Errorf("failed to check for shared config repository %s/%s: %w", organization, config.SharedConfigRepoName, err)
	}
	if !exists {
		return "", fmt.Errorf("shared config repository not found: %s/%s", organization, config.SharedConfigRepoName)
	}

	return fmt.Sprintf("git@github.com:%s/%s.git", organization, config.SharedConfigRepoName), nil
}

func syncSharedSourceDuringInit(ctx context.Context, configFile, team, organization string) {
	cfg, err := loadConfigWithSource()
	if err == nil && cfg.HasSource() {
		pullSharedSourceDuringInit(ctx, cfg, team, organization)
		return
	}

	if strings.TrimSpace(organization) == "" {
		return
	}

	sourceCfg := config.SourceConfig{
		Branch: config.DefaultSourceBranch,
		Root:   config.DefaultSourceRoot,
	}
	sourceCfg.Repository, err = detectSharedSourceRepository(ctx, organization)
	if err != nil {
		return
	}

	if err := config.UpdateSourceConfigFile(configFile, sourceCfg); err != nil {
		output.WarningMessage(fmt.Sprintf("Detected shared config source but failed to save it: %v", err))
		return
	}

	output.InfoMessage(fmt.Sprintf("Configured shared source from %s/%s", organization, config.SharedConfigRepoName))

	cfg, err = loadConfigWithSource()
	if err != nil {
		output.WarningMessage(fmt.Sprintf("Shared source was configured but could not be loaded: %v", err))
		return
	}

	pullSharedSourceDuringInit(ctx, cfg, team, organization)
}

func pullSharedSourceDuringInit(ctx context.Context, cfg *config.Configuration, team, organization string) {
	if hooksSummary, err := cfg.PullHooks(ctx); err == nil {
		printSourcePullSummary("hooks", hooksSummary)
	} else if !isMissingSharedPathError(err) {
		output.WarningMessage(fmt.Sprintf("Failed to pull shared hooks: %v", err))
	}

	team = strings.TrimSpace(team)
	if team == "" {
		team = cfg.DefaultTeam
	}
	if team == "" {
		return
	}

	organization = strings.TrimSpace(organization)
	if organization == "" {
		organization = cfg.DefaultOrganization
	}

	teamSummary, err := cfg.PullTeam(ctx, team, organization)
	if err != nil {
		output.WarningMessage(fmt.Sprintf("Failed to pull shared team config for %s: %v", team, err))
		return
	}

	printSourcePullSummary(fmt.Sprintf("team %s", team), teamSummary)
}

func isMissingSharedPathError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "shared config path not found")
}

func printSourcePullSummary(label string, summary *config.SourcePullSummary) {
	if summary == nil || summary.FileOps == nil {
		return
	}

	output.SuccessMessage(fmt.Sprintf("Pulled %s", label))
	fmt.Printf("Source root: %s\n", output.Bold(summary.SourceRoot))
	for _, target := range summary.Targets {
		fmt.Printf("Target:      %s\n", output.Bold(target))
	}
	fmt.Printf("Created:     %d\n", len(summary.FileOps.Created))
	fmt.Printf("Updated:     %d\n", len(summary.FileOps.Overrode))
	fmt.Println()
}

// showConfigPaths displays all configuration paths
func showConfigPaths() {
	fmt.Println("Configuration paths:")
	fmt.Printf("  Config file:    %s\n", configuredConfigFilePath())
	fmt.Printf("  Org dir:     %s\n", defaultConfigDir)
	fmt.Printf("  Teams dir:      %s\n", defaultTeamsDir)

	// Show the actual config file being used by viper
	if viper.ConfigFileUsed() != "" && viper.ConfigFileUsed() != defaultConfigFile {
		fmt.Printf("  Active config:  %s\n", viper.ConfigFileUsed())
	}
}

func resolveConfigPath(cmd *cobra.Command, cfg *config.Configuration, pathType string) (string, error) {
	switch pathType {
	case "team":
		return resolveTeamConfigPath(cmd, cfg)
	case "user":
		return resolveUserConfigPath(cmd, cfg)
	case "hooks":
		return cfg.GetHooksDir(), nil
	case "templates":
		return cfg.CacheDir, nil
	default:
		return "", fmt.Errorf("unsupported path type %q (expected team, user, hooks, or templates)", pathType)
	}
}

func resolveEditTargetPath(cmd *cobra.Command, cfg *config.Configuration, target string) (string, error) {
	switch target {
	case configTargetMain:
		return configuredConfigFilePath(), nil
	case configTargetTeam, "user", "hooks", "templates":
		return resolveConfigPath(cmd, cfg, target)
	default:
		return "", fmt.Errorf("unsupported edit target %q (expected main, team, user, hooks, or templates)", target)
	}
}

func resolveTeamConfigPath(cmd *cobra.Command, cfg *config.Configuration) (string, error) {
	teamName, err := cmd.Flags().GetString("team")
	if err != nil {
		return "", fmt.Errorf("failed to get 'team' flag: %w", err)
	}
	if teamName == "" {
		teamName = viper.GetString("default_team")
	}
	if teamName == "" {
		return "", fmt.Errorf("team name is required (use --team or configure default_team)")
	}

	organization, err := cmd.Flags().GetString("organization")
	if err != nil {
		return "", fmt.Errorf("failed to get 'organization' flag: %w", err)
	}
	if organization == "" {
		organization = viper.GetString("default_organization")
	}

	return cfg.GetTeamDir(teamName, organization), nil
}

func resolveUserConfigPath(cmd *cobra.Command, cfg *config.Configuration) (string, error) {
	username, err := cmd.Flags().GetString("user")
	if err != nil {
		return "", fmt.Errorf("failed to get 'user' flag: %w", err)
	}
	if username == "" {
		username = viper.GetString("github.username")
	}
	if username == "" {
		currentUser, err := user.Current()
		if err == nil {
			username = currentUser.Username
		}
	}

	return cfg.GetPersonalDir(username), nil
}

func ensureEditableTarget(target, targetPath string) error {
	if target == configTargetMain {
		return config.InitializeConfigFile(targetPath, false, "", "")
	}

	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("failed to create %s path %s: %w", target, targetPath, err)
	}

	return nil
}

func openEditor(cmd *cobra.Command, targetPath string) error {
	editorCommand, args, err := editorCommandParts()
	if err != nil {
		return err
	}

	editorArgs := append(args, targetPath)
	editorCmd := exec.CommandContext(cmd.Context(), editorCommand, editorArgs...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open %s with %s: %w", targetPath, editorCommand, err)
	}

	return nil
}

func editorCommandParts() (string, []string, error) {
	for _, envKey := range []string{"VISUAL", "EDITOR"} {
		if raw := strings.TrimSpace(os.Getenv(envKey)); raw != "" {
			parts := strings.Fields(raw)
			if len(parts) == 0 {
				continue
			}
			return parts[0], parts[1:], nil
		}
	}

	switch runtime.GOOS {
	case "darwin":
		return "open", nil, nil
	case "linux":
		return "xdg-open", nil, nil
	case "windows":
		return "cmd", []string{"/c", "start", ""}, nil
	default:
		return "", nil, fmt.Errorf("no editor configured; set $VISUAL or $EDITOR")
	}
}

func validateMainConfig(configFile string, report *configValidationReport) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		report.addError(fmt.Sprintf("main config: failed to read %s: %v", configFile, err))
		return fmt.Errorf("configuration validation failed")
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		report.addError(fmt.Sprintf("main config: invalid YAML in %s: %v", configFile, err))
		return fmt.Errorf("configuration validation failed")
	}

	report.addChecked(fmt.Sprintf("main config: %s", configFile))
	validateMainConfigSettings(raw, report)

	if report.hasErrors() {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

func validateMainConfigSettings(raw map[string]interface{}, report *configValidationReport) {
	validateDefaultVisibility(raw["default_visibility"], report)
	validateSourceConfig(raw["config_source"], report)

	for _, issue := range collectTemplateValidationIssues(raw["templates"]) {
		report.addError(issue)
	}
}

func validateDefaultVisibility(value interface{}, report *configValidationReport) {
	if value == nil {
		return
	}

	visibility, ok := value.(string)
	if !ok {
		report.addError("main config: default_visibility must be a string")
		return
	}

	trimmed := strings.TrimSpace(visibility)
	if trimmed == "" {
		return
	}

	if !slices.Contains([]string{"private", "public"}, trimmed) {
		report.addError(fmt.Sprintf("main config: default_visibility must be \"private\" or \"public\", got %q", visibility))
	}
}

func validateSourceConfig(value interface{}, report *configValidationReport) {
	if value == nil {
		return
	}

	sourceMap, ok := toStringAnyMap(value)
	if !ok {
		report.addError("main config: config_source must be a map")
		return
	}

	if repository, exists := sourceMap["repository"]; exists {
		repositoryValue, ok := repository.(string)
		if !ok || strings.TrimSpace(repositoryValue) == "" {
			report.addError("main config: config_source.repository must be a non-empty string when set")
		}
	}

	if branch, exists := sourceMap["branch"]; exists {
		branchValue, ok := branch.(string)
		if !ok || strings.TrimSpace(branchValue) == "" {
			report.addError("main config: config_source.branch must be a non-empty string when set")
		}
	}

	if root, exists := sourceMap["root"]; exists {
		rootValue, ok := root.(string)
		if !ok || strings.TrimSpace(rootValue) == "" {
			report.addError("main config: config_source.root must be a non-empty string when set")
		}
	}
}

func collectTemplateValidationIssues(rawTemplates interface{}) []string {
	if rawTemplates == nil {
		return nil
	}

	languages, ok := toStringAnyMap(rawTemplates)
	if !ok {
		return []string{"main config: templates must be a map"}
	}

	var issues []string
	for language, rawTypes := range languages {
		types, ok := toStringAnyMap(rawTypes)
		if !ok {
			issues = append(issues, fmt.Sprintf("main config: templates.%s must be a map", language))
			continue
		}

		for templateType, rawEntry := range types {
			switch entry := rawEntry.(type) {
			case string:
				if strings.TrimSpace(entry) == "" {
					issues = append(issues, fmt.Sprintf("main config: templates.%s.%s must not be empty", language, templateType))
				}
			default:
				entryMap, ok := toStringAnyMap(entry)
				if !ok {
					issues = append(issues, fmt.Sprintf("main config: templates.%s.%s must be a string or map", language, templateType))
					continue
				}

				source, ok := entryMap["source"].(string)
				if !ok || strings.TrimSpace(source) == "" {
					issues = append(issues, fmt.Sprintf("main config: templates.%s.%s.source must be a non-empty string", language, templateType))
				}
			}
		}
	}

	sort.Strings(issues)
	return issues
}

func toStringAnyMap(value interface{}) (map[string]interface{}, bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed, true
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			keyString, ok := key.(string)
			if !ok {
				return nil, false
			}
			result[keyString] = item
		}
		return result, true
	default:
		return nil, false
	}
}

func collectValidationTargets(cmd *cobra.Command, cfg *config.Configuration) ([]configValidationTarget, error) {
	validateAllTeams, err := cmd.Flags().GetBool("all-teams")
	if err != nil {
		return nil, fmt.Errorf("failed to get 'all-teams' flag: %w", err)
	}
	if validateAllTeams {
		return discoverTeamValidationTargets(cfg)
	}

	targets := make([]configValidationTarget, 0, 2)

	teamTarget, err := resolveExplicitOrDefaultTeamValidationTarget(cmd, cfg)
	if err != nil {
		return nil, err
	}
	if teamTarget != nil {
		targets = append(targets, *teamTarget)
	}

	userTarget, err := resolveUserValidationTarget(cmd, cfg)
	if err != nil {
		return nil, err
	}
	if userTarget != nil {
		targets = append(targets, *userTarget)
	}

	return targets, nil
}

func resolveExplicitOrDefaultTeamValidationTarget(cmd *cobra.Command, cfg *config.Configuration) (*configValidationTarget, error) {
	teamName, err := cmd.Flags().GetString("team")
	if err != nil {
		return nil, fmt.Errorf("failed to get 'team' flag: %w", err)
	}

	if teamName == "" {
		teamName = viper.GetString("default_team")
	}
	if teamName == "" {
		return nil, nil
	}

	organization, err := cmd.Flags().GetString("organization")
	if err != nil {
		return nil, fmt.Errorf("failed to get 'organization' flag: %w", err)
	}
	if organization == "" {
		organization = viper.GetString("default_organization")
	}

	label := fmt.Sprintf("team config (%s)", teamName)
	if organization != "" {
		label = fmt.Sprintf("team config (%s/%s)", organization, teamName)
	}

	return &configValidationTarget{
		label: label,
		path:  cfg.GetTeamDir(teamName, organization),
		kind:  "team",
	}, nil
}

func resolveUserValidationTarget(cmd *cobra.Command, cfg *config.Configuration) (*configValidationTarget, error) {
	username, err := cmd.Flags().GetString("user")
	if err != nil {
		return nil, fmt.Errorf("failed to get 'user' flag: %w", err)
	}
	if username == "" {
		return nil, nil
	}

	return &configValidationTarget{
		label: fmt.Sprintf("user config (%s)", username),
		path:  cfg.GetPersonalDir(username),
		kind:  "user",
	}, nil
}

func discoverTeamValidationTargets(cfg *config.Configuration) ([]configValidationTarget, error) {
	targetMap := map[string]configValidationTarget{}

	if err := collectLegacyTeamValidationTargets(cfg, targetMap); err != nil {
		return nil, err
	}
	if err := collectOrgTeamValidationTargets(cfg, targetMap); err != nil {
		return nil, err
	}

	targets := make([]configValidationTarget, 0, len(targetMap))
	for _, target := range targetMap {
		targets = append(targets, target)
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].path < targets[j].path
	})

	return targets, nil
}

func collectLegacyTeamValidationTargets(cfg *config.Configuration, targetMap map[string]configValidationTarget) error {
	entries, err := os.ReadDir(cfg.TeamsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read teams directory %s: %w", cfg.TeamsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		teamPath := filepath.Join(cfg.TeamsDir, entry.Name())
		targetMap[teamPath] = configValidationTarget{
			label: fmt.Sprintf("team config (%s)", entry.Name()),
			path:  teamPath,
			kind:  configTargetTeam,
		}
	}

	return nil
}

func collectOrgTeamValidationTargets(cfg *config.Configuration, targetMap map[string]configValidationTarget) error {
	orgsRoot := filepath.Join(cfg.ConfigDir, config.OrgsDirName)
	orgEntries, err := os.ReadDir(orgsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read organizations directory %s: %w", orgsRoot, err)
	}

	for _, orgEntry := range orgEntries {
		if !orgEntry.IsDir() {
			continue
		}

		teamsRoot := filepath.Join(orgsRoot, orgEntry.Name(), config.TeamsDirName)
		teamEntries, err := os.ReadDir(teamsRoot)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read teams directory %s: %w", teamsRoot, err)
		}

		for _, teamEntry := range teamEntries {
			if !teamEntry.IsDir() {
				continue
			}

			teamPath := filepath.Join(teamsRoot, teamEntry.Name())
			targetMap[teamPath] = configValidationTarget{
				label: fmt.Sprintf("team config (%s/%s)", orgEntry.Name(), teamEntry.Name()),
				path:  teamPath,
				kind:  configTargetTeam,
			}
		}
	}

	return nil
}

func validateConfigDirectory(target configValidationTarget, report *configValidationReport) {
	info, err := os.Stat(target.path)
	if err != nil {
		if os.IsNotExist(err) {
			report.addError(fmt.Sprintf("%s: path does not exist: %s", target.label, target.path))
			return
		}
		report.addError(fmt.Sprintf("%s: failed to stat %s: %v", target.label, target.path, err))
		return
	}
	if !info.IsDir() {
		report.addError(fmt.Sprintf("%s: expected a directory, got %s", target.label, target.path))
		return
	}

	report.addChecked(fmt.Sprintf("%s: %s", target.label, target.path))
	validateSecretsConfigFile(target, report)
	validateTeamOverrideConfigFile(target, report)
	validateYAMLConfigDir(target, "envs", validateEnvironmentConfigFile, report)
	validateYAMLConfigDir(target, "rulesets", validateRulesetConfigFile, report)
}

func validateSecretsConfigFile(target configValidationTarget, report *configValidationReport) {
	secretsPath := filepath.Join(target.path, "secrets.yaml")
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.addWarning(fmt.Sprintf("%s: secrets.yaml not found", target.label))
			return
		}
		report.addError(fmt.Sprintf("%s: failed to read %s: %v", target.label, secretsPath, err))
		return
	}

	rendered, err := renderConfigTemplate(data)
	if err != nil {
		report.addError(fmt.Sprintf("%s: failed to render %s: %v", target.label, secretsPath, err))
		return
	}

	var secretsConfig secretspkg.Config
	if err := yaml.Unmarshal([]byte(rendered), &secretsConfig); err != nil {
		report.addError(fmt.Sprintf("%s: invalid YAML in %s: %v", target.label, secretsPath, err))
		return
	}

	for index, secret := range secretsConfig.Secrets {
		entryPath := fmt.Sprintf("%s: secrets.yaml entry %d", target.label, index+1)
		if strings.TrimSpace(secret.Name) == "" {
			report.addError(entryPath + " is missing name")
		}
		if secret.Type != "" && !slices.Contains([]string{"secret", "variable"}, secret.Type) {
			report.addError(fmt.Sprintf("%s has unsupported type %q", entryPath, secret.Type))
		}
	}

	report.addChecked(fmt.Sprintf("%s: %s", target.label, secretsPath))
}

func validateTeamOverrideConfigFile(target configValidationTarget, report *configValidationReport) {
	configPath := filepath.Join(target.path, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.addWarning(fmt.Sprintf("%s: config.yaml not found", target.label))
			return
		}
		report.addError(fmt.Sprintf("%s: failed to read %s: %v", target.label, configPath, err))
		return
	}

	var overrideFile struct {
		Templates map[string]map[string]*config.TemplateDefinition `yaml:"templates"`
	}
	if err := yaml.Unmarshal(data, &overrideFile); err != nil {
		report.addError(fmt.Sprintf("%s: invalid YAML in %s: %v", target.label, configPath, err))
		return
	}

	for language, types := range overrideFile.Templates {
		for projectType, definition := range types {
			if definition == nil {
				continue
			}
			if definition.Source == "" && definition.Hooks == nil {
				report.addError(fmt.Sprintf("%s: templates.%s.%s must define source and/or hooks", target.label, language, projectType))
			}
		}
	}

	report.addChecked(fmt.Sprintf("%s: %s", target.label, configPath))
}

func validateYAMLConfigDir(
	target configValidationTarget,
	dirName string,
	validateFile func(configValidationTarget, string, *configValidationReport),
	report *configValidationReport,
) {
	dirPath := filepath.Join(target.path, dirName)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			report.addWarning(fmt.Sprintf("%s: %s directory not found", target.label, dirPath))
			return
		}
		report.addError(fmt.Sprintf("%s: failed to read %s: %v", target.label, dirPath, err))
		return
	}

	yamlFound := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isYAMLFile(entry.Name()) {
			continue
		}

		yamlFound = true
		validateFile(target, filepath.Join(dirPath, entry.Name()), report)
	}

	if !yamlFound {
		report.addWarning(fmt.Sprintf("%s: no YAML files found in %s", target.label, dirPath))
	}
}

func validateEnvironmentConfigFile(target configValidationTarget, filePath string, report *configValidationReport) {
	rendered, err := renderConfigTemplateFile(filePath)
	if err != nil {
		report.addError(fmt.Sprintf("%s: failed to render %s: %v", target.label, filePath, err))
		return
	}

	var envConfig project.EnvConf
	if err := yaml.Unmarshal([]byte(rendered), &envConfig); err != nil {
		report.addError(fmt.Sprintf("%s: invalid YAML in %s: %v", target.label, filePath, err))
		return
	}
	if strings.TrimSpace(envConfig.Name) == "" {
		report.addError(fmt.Sprintf("%s: environment in %s is missing name", target.label, filePath))
		return
	}

	report.addChecked(fmt.Sprintf("%s: %s", target.label, filePath))
}

func validateRulesetConfigFile(target configValidationTarget, filePath string, report *configValidationReport) {
	rendered, err := renderConfigTemplateFile(filePath)
	if err != nil {
		report.addError(fmt.Sprintf("%s: failed to render %s: %v", target.label, filePath, err))
		return
	}

	var ruleset map[string]interface{}
	if err := yaml.Unmarshal([]byte(rendered), &ruleset); err != nil {
		report.addError(fmt.Sprintf("%s: invalid YAML in %s: %v", target.label, filePath, err))
		return
	}
	name := ""
	if value, ok := ruleset["name"].(string); ok {
		name = value
	}
	if strings.TrimSpace(name) == "" {
		report.addError(fmt.Sprintf("%s: ruleset in %s is missing name", target.label, filePath))
	}
	targetValue, ok := ruleset["target"].(string)
	if !ok || strings.TrimSpace(targetValue) == "" {
		report.addError(fmt.Sprintf("%s: ruleset in %s is missing target", target.label, filePath))
	}

	report.addChecked(fmt.Sprintf("%s: %s", target.label, filePath))
}

func renderConfigTemplateFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return renderConfigTemplate(data)
}

func renderConfigTemplate(data []byte) (string, error) {
	vars := templatepkg.NewTemplateVariables()
	vars.Org.Name = "org"
	vars.Org.Team = "team"
	vars.Org.TeamID = 1
	vars.Repo.Owner = "owner"
	vars.Repo.Name = "repo"
	vars.Project.Name = "project"

	return templatepkg.NewRenderer(vars).RenderString(string(data))
}

func isYAMLFile(name string) bool {
	extension := strings.ToLower(filepath.Ext(name))
	return extension == ".yaml" || extension == ".yml"
}

func printConfigValidationReport(report *configValidationReport) {
	fmt.Println("Configuration validation")

	for _, checked := range report.checked {
		fmt.Printf("  [ok] %s\n", checked)
	}
	for _, warning := range report.warnings {
		fmt.Printf("  [warn] %s\n", warning)
	}
	for _, validationError := range report.errors {
		fmt.Printf("  [error] %s\n", validationError)
	}

	if report.hasErrors() {
		fmt.Printf("\nResult: %d error(s), %d warning(s)\n", len(report.errors), len(report.warnings))
		return
	}

	fmt.Printf("\nResult: valid (%d warning(s))\n", len(report.warnings))
}

// initializeConfigFile creates or updates the main config file
func initializeConfigFile(org, team string, override bool) error { //nolint:gofumpt
	// Use the centralised function from the config package
	if err := config.InitializeConfigFile(defaultConfigFile, override, team, org); err != nil {
		return err
	}

	// Display appropriate messages based on the operation
	configExists := true
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		configExists = false
	}

	if configExists && override {
		fmt.Println("Overrode existing configuration file:", defaultConfigFile)
	} else if !configExists {
		fmt.Println("Created configuration file:", defaultConfigFile)
	} else {
		fmt.Println("Configuration file already exists:", defaultConfigFile)
		fmt.Println("Use --override to replace it with a fresh configuration.")
	}

	if configExists || override {
		fmt.Println("You can edit this file to customize viaplay-cli behavior.")
	}

	return nil
}

// showAllConfig displays all configuration values
func showAllConfig() error {
	allSettings := viper.AllSettings()
	fmt.Println("Current configuration:")
	for k, v := range allSettings {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return nil
}

// getConfigValue retrieves and displays a specific config value
func getConfigValue(key string) error {
	if !viper.IsSet(key) {
		return fmt.Errorf("key %q not found in config", key)
	}

	value := viper.Get(key)
	fmt.Printf("%v\n", value)
	return nil
}

// scaffoldTeamConfig creates the team configuration files and directories
func scaffoldTeamConfig(org, team string, override bool) error {
	// Use the config package's team setup functionality (which will create any needed directories)
	result, err := config.SetupTeam(org, team, override)
	if err != nil {
		return err
	}

	// Print a summary of what happened
	if len(result.Created) > 0 {
		fmt.Printf("\nCreated %d config files for team '%s'\n", len(result.Created), team)
	}
	if len(result.Overrode) > 0 {
		fmt.Printf("Overrode %d existing config files for team '%s'\n", len(result.Overrode), team)
	}
	if len(result.Skipped) > 0 {
		fmt.Printf("Skipped %d existing config files for team '%s'\n", len(result.Skipped), team)
	}

	return nil
}

// scaffoldPersonalConfig creates personal account configurations
func scaffoldPersonalConfig(username string, override bool) error {
	result, err := config.SetupPersonal(username, override)
	if err != nil {
		return err
	}

	// Print results
	for _, path := range result.Created {
		fmt.Printf("Created personal config file: %s\n", path)
	}
	for _, path := range result.Overrode {
		fmt.Printf("Overrode personal config file: %s\n", path)
	}
	for _, path := range result.Skipped {
		fmt.Printf("Personal config file already exists: %s\n", path)
	}

	return nil
}

// showInitSuccessMessage displays a success message with next steps
func showInitSuccessMessage(team, organization string) {
	fmt.Printf("\n%s %s\n\n", output.ActiveIcons.Success, output.SuccessBold("Configuration initialized successfully!"))

	fmt.Println("Next steps:")

	if team != "" {
		fmt.Printf("  %s Set '%s' as your default team\n", output.ActiveIcons.Bullet, team)
		if organization != "" {
			fmt.Printf("  %s Configuration is set up for team '%s' in organization '%s'\n",
				output.ActiveIcons.Bullet, team, organization)
		}
		fmt.Printf("  %s Create your first project with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip create project --name myproject --team "+team))
	} else {
		fmt.Printf("  %s Create a team configuration with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip config init --team myteam"))
		fmt.Printf("  %s Create your first project with: %s\n",
			output.ActiveIcons.Bullet,
			output.Bold("vip create project --name myproject"))
	}

	fmt.Printf("  %s View your config with: %s\n",
		output.ActiveIcons.Bullet,
		output.Bold("vip config get"))

	fmt.Println()
}

// init sets up the configuration command and its subcommands
func init() {
	defaultConfigDir = config.GetDefaultConfigDir()
	defaultConfigFile = config.GetDefaultConfigFile()
	defaultTeamsDir = config.GetDefaultTeamsDir()

	// Add all subcommands to the config command
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(editCmd)
	configCmd.AddCommand(pathCmd)
	configCmd.AddCommand(pathsCmd)
	configCmd.AddCommand(pullCmd)
	configCmd.AddCommand(validateCmd)
	initCmd.AddCommand(initTeamCmd)
	initCmd.AddCommand(initSourceCmd)
	pullCmd.AddCommand(pullTeamCmd)
	pullCmd.AddCommand(pullHooksCmd)
	pullCmd.AddCommand(pullAllCmd)

	// Define flags for the init command
	initCmd.Flags().StringP("team", "t", "", "Team name to scaffold configs for (optional)")
	initCmd.Flags().Bool("override", false, "Override existing config files if they exist")
	initCmd.Flags().StringP("organization", "o", "", "Organization name for team configs (optional)")

	initTeamCmd.Flags().Bool("override", false, "Override existing team config files if they exist")
	initTeamCmd.Flags().StringP("organization", "o", "", "Organization name for the team config")

	initSourceCmd.Flags().StringP("repository", "r", "", "Shared config repository URL")
	initSourceCmd.Flags().String("branch", "", "Branch to use from the shared config repository (defaults to current config or main)")
	initSourceCmd.Flags().String("root", "", "Root path inside the shared config repository (defaults to current config or .)")
	initSourceCmd.Flags().StringP("organization", "o", "", "Organization used to detect the conventional vip-shared-configs repository")

	pathCmd.Flags().StringP("team", "t", "", "Team name for the team config path (falls back to default_team)")
	pathCmd.Flags().StringP("organization", "o", "", "Organization name for the team config path (falls back to default_organization)")
	pathCmd.Flags().StringP("user", "u", "", "Username for the personal config path")

	editCmd.Flags().StringP("team", "t", "", "Team name for the team config path (falls back to default_team)")
	editCmd.Flags().StringP("organization", "o", "", "Organization name for the team config path (falls back to default_organization)")
	editCmd.Flags().StringP("user", "u", "", "Username for the personal config path")

	validateCmd.Flags().StringP("team", "t", "", "Team name to validate (falls back to default_team)")
	validateCmd.Flags().StringP("organization", "o", "", "Organization name for the team config path (falls back to default_organization)")
	validateCmd.Flags().StringP("user", "u", "", "Username for the personal config path")
	validateCmd.Flags().Bool("all-teams", false, "Validate every discovered team configuration directory")

	pullTeamCmd.Flags().StringP("organization", "o", "", "Organization name for the team config path (falls back to default_organization)")
}
