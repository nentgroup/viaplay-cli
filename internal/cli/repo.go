// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	projectpkg "github.com/nentgroup/viaplay-cli/internal/project"
	secretspkg "github.com/nentgroup/viaplay-cli/internal/secrets"
)

const (
	repoSecretTypeSecret   = "secret"
	repoSecretTypeVariable = "variable"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage GitHub repositories",
	Long: `Manage GitHub repositories.

This command provides subcommands to create and manipulate repositories:
  - 'repo create': Create a GitHub repository without code scaffolding
  - 'repo apply': Apply configuration to an existing repository
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// NewRepoCommand returns the repo command with all its subcommands.
func NewRepoCommand() *cobra.Command {
	repoCmd.AddCommand(newRepoCreateCommand())
	repoCmd.AddCommand(newRepoApplyCommand())
	return repoCmd
}

// newRepoCreateCommand creates the repo create command.
func newRepoCreateCommand() *cobra.Command {
	opts := &CreateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "create [owner/]<repo-name>",
		Short: "Create a GitHub repository without code scaffolding",
		Long: `Create a GitHub repository without code scaffolding.

This command:
1. Creates a GitHub repository
2. Applies organization settings (environments, rulesets, secrets)

Use this when you need to create a repository structure but will add code 
manually or migrate existing code to a new repository.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			parseOwnerRepoArg(args[0], opts)
			return createProjectOrRepo(ctx, opts, false)
		},
		Example: `  # Create a private repository with default settings
  vip repo create my-repo --team platform

  # Create with explicit owner
  vip repo create myorg/my-repo --team platform

  # Create a public repository
  vip repo create my-public-repo --public --team frontend

  # Create a repository with a specific description
  vip repo create my-repo --description "This is my custom repository description"`,
	}

	addCommonFlags(cmd, opts)
	return cmd
}

// newRepoApplyCommand creates the repo apply command.
func newRepoApplyCommand() *cobra.Command {
	applyOpts := &RepoApplyOptions{}

	cmd := &cobra.Command{
		Use:   "apply [owner/]<repo-name>",
		Short: "Apply team configuration to an existing GitHub repository",
		Long: `Apply team configuration (environments, rulesets, secrets) to an existing GitHub repository.

This command reuses the same configuration as 'repo create' but without creating a new repository.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			parts := strings.SplitN(args[0], "/", 2)
			if len(parts) == 2 {
				applyOpts.Owner = parts[0]
				applyOpts.Repo = parts[1]
			} else {
				applyOpts.Repo = parts[0]
			}
			if applyOpts.Team == "" {
				applyOpts.Team = viper.GetString("default_team")
			}
			return applyRepoConfig(ctx, applyOpts)
		},
		Example: `  # Apply all configured scopes
  vip repo apply myorg/myrepo --team platform

  # Apply only secrets
  vip repo apply myorg/myrepo --only secrets --team platform

  # Apply repo-specific secrets only
  vip repo apply myorg/myrepo --only repo-secrets --secrets-file secrets.json

  # Dry run
  vip repo apply myorg/myrepo --dry-run --team platform`,
	}

	cmd.Flags().StringVar(&applyOpts.Team, "team", "", "Team name for loading configuration (uses default_team from config if not specified)")
	cmd.Flags().StringVar(&applyOpts.Owner, "owner", "", "Repository owner (required if not specified in repo argument)")
	cmd.Flags().StringVar(&applyOpts.OnlyScopes, "only", "", "Comma-separated list of scopes to apply (envs,rulesets,secrets,variables,repo-secrets)")
	cmd.Flags().StringVar(&applyOpts.SkipScopes, "skip", "", "Comma-separated list of scopes to skip")
	cmd.Flags().StringVar(&applyOpts.RepoSecrets, "repo-secrets", "", "JSON string containing repository-specific secrets/variables")
	cmd.Flags().StringVar(&applyOpts.SecretsFile, "secrets-file", "", "Path to a JSON file containing repository-specific secrets/variables")
	cmd.Flags().BoolVar(&applyOpts.DryRun, "dry-run", false, "Show what would be applied without applying it")

	cmd.AddCommand(newRepoSecretApplyCommand())
	cmd.AddCommand(newRepoVariableApplyCommand())
	return cmd
}

// RepoApplyOptions contains options for applying config to a repo.
type RepoApplyOptions struct {
	Owner       string
	Repo        string
	Team        string
	OnlyScopes  string
	SkipScopes  string
	RepoSecrets string
	SecretsFile string
	DryRun      bool
}

// applyRepoConfig applies configuration to an existing repository.
func applyRepoConfig(ctx context.Context, opts *RepoApplyOptions) error {
	output.Section("Applying Configuration to Repository")

	if err := validateRepoApplyOptions(opts); err != nil {
		return err
	}
	if err := populateRepoOwner(opts); err != nil {
		return err
	}

	ghClient, configDir, err := setupGitHubClient()
	if err != nil {
		return err
	}
	if err := ensureRepositoryExists(ctx, ghClient, opts.Owner, opts.Repo); err != nil {
		return err
	}

	repoSecretsStr, err := loadRepoSecretsInput(opts)
	if err != nil {
		return err
	}
	scopesToApply := buildRepoApplyScopes(opts, repoSecretsStr != "")
	projectOpts := buildRepoApplyProjectOptions(opts, configDir, repoSecretsStr, scopesToApply)

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if opts.DryRun {
		printRepoApplyDryRun(ctx, ghClient, cfg, opts, repoSecretsStr, scopesToApply)
		return nil
	}

	reporter := output.NewCallbackReporter(output.DefaultCB, viper.GetBool("verbose"))
	factory := projectpkg.NewFactory(ghClient, reporter, cfg)
	if err := factory.ApplyConfigurations(ctx, projectOpts); err != nil {
		return err
	}

	output.SuccessMessage(fmt.Sprintf("Applied configuration to %s/%s", opts.Owner, opts.Repo))
	return nil
}

func validateRepoApplyOptions(opts *RepoApplyOptions) error {
	if opts.Repo == "" {
		return fmt.Errorf("repository name is required")
	}
	return nil
}

func populateRepoOwner(opts *RepoApplyOptions) error {
	if opts.Owner != "" {
		return nil
	}
	opts.Owner = viper.GetString("github.organization")
	if opts.Owner == "" {
		opts.Owner = viper.GetString("github.username")
	}
	if opts.Owner == "" {
		return fmt.Errorf("repository owner is required (use --owner or configure github.username/github.organization)")
	}
	return nil
}

func ensureRepositoryExists(ctx context.Context, ghClient *gh.GitHubClient, owner, repo string) error {
	exists, err := ghClient.RepositoryExists(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to check if repository exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("repository %s/%s does not exist", owner, repo)
	}
	return nil
}

func loadRepoSecretsInput(opts *RepoApplyOptions) (string, error) {
	if opts.SecretsFile == "" {
		return opts.RepoSecrets, nil
	}
	data, err := os.ReadFile(opts.SecretsFile)
	if err != nil {
		return "", fmt.Errorf("failed to read secrets file: %w", err)
	}
	return string(data), nil
}

func buildRepoApplyScopes(opts *RepoApplyOptions, hasRepoSecrets bool) map[string]bool {
	scopesToApply := map[string]bool{
		"envs":         true,
		"rulesets":     true,
		"secrets":      true,
		"variables":    true,
		"repo-secrets": hasRepoSecrets,
	}
	applyScopeList(scopesToApply, opts.OnlyScopes, true)
	applyScopeList(scopesToApply, opts.SkipScopes, false)
	return scopesToApply
}

func applyScopeList(scopes map[string]bool, raw string, enabled bool) {
	if raw == "" {
		return
	}
	if enabled {
		for scope := range scopes {
			scopes[scope] = false
		}
	}
	for _, scope := range strings.Split(raw, ",") {
		scopes[strings.TrimSpace(scope)] = enabled
	}
}

func buildRepoApplyProjectOptions(opts *RepoApplyOptions, configDir, repoSecretsStr string, scopes map[string]bool) projectpkg.Options {
	return projectpkg.Options{
		RepoOwner:       opts.Owner,
		RepoName:        opts.Repo,
		RepoDescription: fmt.Sprintf("Repository for %s", opts.Repo),
		Team:            opts.Team,
		ConfigDir:       configDir,
		ApplyEnvs:       scopes["envs"],
		ApplyRulesets:   scopes["rulesets"],
		ApplySecrets:    scopes["secrets"],
		ApplyVariables:  scopes["variables"],
		RepoSecrets:     repoSecretsStr,
	}
}

func printRepoApplyDryRun(ctx context.Context, ghClient *gh.GitHubClient, cfg *config.Configuration, opts *RepoApplyOptions, repoSecretsStr string, scopesToApply map[string]bool) {
	output.WarningMessage("DRY RUN: No changes will be applied")
	fmt.Printf("Repository: %s/%s\n", opts.Owner, opts.Repo)
	if opts.Team != "" {
		fmt.Printf("Team: %s\n", opts.Team)
	}
	configDir := determineRepoApplyConfigDir(ctx, ghClient, cfg, opts)
	if configDir != "" {
		fmt.Printf("Config directory: %s\n", configDir)
	}
	fmt.Println("Plan:")
	for _, scope := range []string{"envs", "rulesets", "secrets", "variables", "repo-secrets"} {
		if scopesToApply[scope] {
			printRepoApplyScopePlan(scope, configDir, repoSecretsStr)
		}
	}
}

func determineRepoApplyConfigDir(ctx context.Context, ghClient *gh.GitHubClient, cfg *config.Configuration, opts *RepoApplyOptions) string {
	if opts.Team != "" {
		u, err := ghClient.GetUser(ctx, opts.Owner)
		if err == nil && strings.EqualFold(u.GetType(), "Organization") {
			return cfg.GetTeamDir(opts.Team, opts.Owner)
		}
	}
	username, err := ghClient.GetAuthenticatedUser(ctx)
	if err == nil && username != "" {
		return cfg.GetPersonalDir(username)
	}
	if opts.Team != "" {
		return cfg.GetTeamDir(opts.Team, "")
	}
	return cfg.GetPersonalDir(opts.Owner)
}

func printRepoApplyScopePlan(scope, configDir, repoSecretsStr string) {
	fmt.Printf("  - %s\n", scope)
	switch scope {
	case "envs":
		printRepoApplyEnvPlan(configDir)
	case "rulesets":
		printRepoApplyRulesetPlan(configDir)
	case "secrets":
		printRepoApplySecretsPlan(configDir, false)
	case "variables":
		printRepoApplySecretsPlan(configDir, true)
	case "repo-secrets":
		printRepoApplyRepoSecretsPlan(repoSecretsStr)
	}
}

func printRepoApplyEnvPlan(configDir string) {
	printRepoApplyFileList(filepath.Join(configDir, "envs"), []string{".yaml"}, "environment")
}

func printRepoApplyRulesetPlan(configDir string) {
	printRepoApplyFileList(filepath.Join(configDir, "rulesets"), []string{".yaml", ".yml"}, "ruleset")
}

func printRepoApplyFileList(dir string, exts []string, label string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("      source: %s (unavailable: %v)\n", dir, err)
		return
	}
	fmt.Printf("      source: %s\n", dir)
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !repoApplyHasExt(entry.Name(), exts) {
			continue
		}
		fmt.Printf("      %s file: %s\n", label, entry.Name())
		count++
	}
	if count == 0 {
		fmt.Printf("      no %s files found\n", label)
	}
}

func repoApplyHasExt(name string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, allowed := range exts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func printRepoApplySecretsPlan(configDir string, variablesOnly bool) {
	secretsPath := filepath.Join(configDir, "secrets.yaml")
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		fmt.Printf("      source: %s (unavailable: %v)\n", secretsPath, err)
		return
	}
	fmt.Printf("      source: %s\n", secretsPath)
	var secretsConfig secretspkg.Config
	if err := yaml.Unmarshal(data, &secretsConfig); err != nil {
		fmt.Printf("      could not parse secrets file: %v\n", err)
		return
	}
	printRepoApplySecretEntries(secretsConfig.Secrets, variablesOnly)
}

func printRepoApplyRepoSecretsPlan(repoSecretsStr string) {
	if repoSecretsStr == "" {
		fmt.Println("      no repo-specific secrets input provided")
		return
	}
	fmt.Println("      source: --repo-secrets/--secrets-file")
	var secretsConfig secretspkg.Config
	if err := json.Unmarshal([]byte(repoSecretsStr), &secretsConfig); err != nil {
		fmt.Printf("      could not parse repo-specific secrets: %v\n", err)
		return
	}
	printRepoApplySecretEntries(secretsConfig.Secrets, false)
}

func repoApplyValueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func printRepoApplySecretEntries(entries []secretspkg.Secret, variablesOnly bool) {
	count := 0
	for _, secret := range entries {
		if secret.Name == "" {
			continue
		}
		isVariable := secret.Type == repoSecretTypeVariable
		if variablesOnly != isVariable {
			continue
		}
		fmt.Printf("      %s %s (%s)\n", repoApplyValueOrDefault(secret.Type, repoSecretTypeSecret), secret.Name, repoApplyValueOrDefault(secret.Env, "repo"))
		count++
	}
	if count == 0 {
		if variablesOnly {
			fmt.Println("      no variable entries found")
			return
		}
		fmt.Println("      no secret entries found")
	}
}

// newRepoSecretApplyCommand creates the repo apply secret command.
func newRepoSecretApplyCommand() *cobra.Command {
	secretOpts := &RepoSecretApplyOptions{}

	cmd := &cobra.Command{
		Use:   "secret [owner/]<repo-name> <secret-name>",
		Short: "Apply a single secret to an existing GitHub repository",
		Long: `Apply a single secret to an existing GitHub repository.

By default this resolves from team config first, then falls back to a keyring entry with the same name.
Use --from for direct one-off sources.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			populateRepoSecretApplyTarget(secretOpts, args[0], args[1])
			if secretOpts.Team == "" {
				secretOpts.Team = viper.GetString("default_team")
			}
			return applyRepoSecret(ctx, secretOpts)
		},
		Example: `  # Apply a team-configured secret
  vip repo apply secret myorg/myrepo MY_SECRET --team platform

  # Apply to a specific environment
  vip repo apply secret myorg/myrepo API_KEY --env staging --team platform

  # Apply a secret from stdin
  echo -n 'super-secret' | vip repo apply secret myorg/myrepo TOKEN --from stdin`,
	}

	addRepoSecretApplyFlags(cmd, secretOpts, true)
	return cmd
}

// newRepoVariableApplyCommand creates the repo apply variable command.
func newRepoVariableApplyCommand() *cobra.Command {
	variableOpts := &RepoSecretApplyOptions{Type: repoSecretTypeVariable}

	cmd := &cobra.Command{
		Use:   "variable [owner/]<repo-name> <name>",
		Short: "Apply a single variable to an existing GitHub repository",
		Long: `Apply a single variable to an existing GitHub repository.

By default this resolves from team config first, then falls back to a keyring entry with the same name.
Use --from for direct one-off sources.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			populateRepoSecretApplyTarget(variableOpts, args[0], args[1])
			if variableOpts.Team == "" {
				variableOpts.Team = viper.GetString("default_team")
			}
			variableOpts.Type = repoSecretTypeVariable
			return applyRepoSecret(ctx, variableOpts)
		},
		Example: `  # Apply a team-configured variable
  vip repo apply variable myorg/myrepo SERVICE_URL --team platform

  # Apply to a specific environment
  vip repo apply variable myorg/myrepo API_URL --env staging --team platform

  # Apply a variable from an environment variable
  vip repo apply variable myorg/myrepo SERVICE_URL --from env --value SERVICE_URL`,
	}

	addRepoSecretApplyFlags(cmd, variableOpts, false)
	return cmd
}

func populateRepoSecretApplyTarget(opts *RepoSecretApplyOptions, ownerRepo, secretName string) {
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) == 2 {
		opts.Owner = parts[0]
		opts.Repo = parts[1]
	} else {
		opts.Repo = parts[0]
	}
	opts.SecretName = secretName
}

func addRepoSecretApplyFlags(cmd *cobra.Command, opts *RepoSecretApplyOptions, includeType bool) {
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name for loading configuration")
	cmd.Flags().StringVar(&opts.Owner, "owner", "", "Repository owner (required if not specified in repo argument)")
	cmd.Flags().StringVar(&opts.Env, "env", "", "Target environment (leave empty for repository scope)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Direct source: keyring, env, file, or stdin")
	if includeType {
		cmd.Flags().StringVar(&opts.Type, "type", "", "Value type: secret or variable (defaults to team config value or secret)")
	}
	cmd.Flags().StringVar(&opts.Value, "value", "", "Literal value, or source selector when used with --from env|file|keyring")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be applied without applying it")
}

// RepoSecretApplyOptions contains options for applying a single secret.
type RepoSecretApplyOptions struct {
	Owner      string
	Repo       string
	SecretName string
	Team       string
	Env        string
	From       string
	Type       string
	Value      string
	DryRun     bool
}

// applyRepoSecret applies a single secret or variable to a repository.
func applyRepoSecret(ctx context.Context, opts *RepoSecretApplyOptions) error {
	output.Section("Applying Secret or Variable to Repository")

	if opts.Repo == "" {
		return fmt.Errorf("repository name is required")
	}
	if opts.SecretName == "" {
		return fmt.Errorf("secret name is required")
	}
	if opts.Owner == "" {
		opts.Owner = viper.GetString("github.organization")
		if opts.Owner == "" {
			opts.Owner = viper.GetString("github.username")
		}
		if opts.Owner == "" {
			return fmt.Errorf("repository owner is required (use --owner or configure github.username/github.organization)")
		}
	}

	ghClient, _, err := setupGitHubClient()
	if err != nil {
		return err
	}

	exists, err := ghClient.RepositoryExists(ctx, opts.Owner, opts.Repo)
	if err != nil {
		return fmt.Errorf("failed to check if repository exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("repository %s/%s does not exist", opts.Owner, opts.Repo)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	value, valueType, valueSource, err := resolveRepoSecretValue(ctx, ghClient, cfg, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		output.WarningMessage("DRY RUN: No changes will be applied")
		fmt.Printf("Repository: %s/%s\n", opts.Owner, opts.Repo)
		fmt.Printf("Name: %s\n", opts.SecretName)
		fmt.Printf("Type: %s\n", valueType)
		fmt.Printf("Source: %s\n", valueSource)
		if opts.Env != "" {
			fmt.Printf("Environment: %s\n", opts.Env)
		}
		return nil
	}

	sanitizedName := secretspkg.SanitizeSecretName(opts.SecretName)
	switch valueType {
	case repoSecretTypeVariable:
		err = ghClient.SetVariable(ctx, opts.Owner, opts.Repo, sanitizedName, value, opts.Env)
	default:
		err = ghClient.ApplySecret(ctx, opts.Owner, opts.Repo, sanitizedName, value, opts.Env)
	}
	if err != nil {
		return fmt.Errorf("failed to apply %s '%s': %w", valueType, sanitizedName, err)
	}

	output.SuccessMessage(fmt.Sprintf("Applied %s '%s' to %s/%s", valueType, sanitizedName, opts.Owner, opts.Repo))
	return nil
}

func resolveRepoSecretValue(ctx context.Context, ghClient *gh.GitHubClient, cfg *config.Configuration, opts *RepoSecretApplyOptions) (string, string, string, error) {
	valueType, err := normalizeRepoSecretType(opts.Type)
	if err != nil {
		return "", "", "", err
	}

	if opts.From != "" {
		value, source, err := resolveRepoSecretValueFromDirectSource(opts)
		if err != nil {
			return "", "", "", err
		}
		if valueType == "" {
			valueType = repoSecretTypeSecret
		}
		return value, valueType, source, nil
	}

	if opts.Value != "" {
		if valueType == "" {
			valueType = repoSecretTypeSecret
		}
		return opts.Value, valueType, "literal", nil
	}

	if opts.Team != "" {
		value, configType, source, err := resolveRepoSecretValueFromTeamConfig(ctx, ghClient, cfg, opts)
		if err == nil {
			if valueType == "" {
				valueType = configType
			}
			return value, valueType, source, nil
		}
	}

	value, err := secretspkg.GetSecret(opts.SecretName)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to resolve '%s' from team config or keyring: %w", opts.SecretName, err)
	}
	if valueType == "" {
		valueType = repoSecretTypeSecret
	}
	return value, valueType, "keyring:" + opts.SecretName, nil
}

func resolveRepoSecretValueFromDirectSource(opts *RepoSecretApplyOptions) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(opts.From)) {
	case "keyring":
		key := opts.Value
		if key == "" {
			key = opts.SecretName
		}
		value, err := secretspkg.GetSecret(key)
		if err != nil {
			return "", "", fmt.Errorf("failed to read keyring secret '%s': %w", key, err)
		}
		return value, "keyring:" + key, nil
	case "env":
		envName := opts.Value
		if envName == "" {
			envName = opts.SecretName
		}
		value, ok := os.LookupEnv(envName)
		if !ok {
			return "", "", fmt.Errorf("environment variable '%s' is not set", envName)
		}
		return value, "env:" + envName, nil
	case "file":
		if opts.Value == "" {
			return "", "", fmt.Errorf("--value must point to a file when --from file is used")
		}
		data, err := os.ReadFile(opts.Value)
		if err != nil {
			return "", "", fmt.Errorf("failed to read secret file '%s': %w", opts.Value, err)
		}
		return strings.TrimSpace(string(data)), "file:" + opts.Value, nil
	case "stdin":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("failed to read secret from stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), "stdin", nil
	default:
		return "", "", fmt.Errorf("unsupported --from value %q (expected keyring, env, file, or stdin)", opts.From)
	}
}

func resolveRepoSecretValueFromTeamConfig(ctx context.Context, ghClient *gh.GitHubClient, cfg *config.Configuration, opts *RepoSecretApplyOptions) (string, string, string, error) {
	paths := secretConfigPathsForRepo(ctx, ghClient, cfg, opts)
	if len(paths) == 0 {
		return "", "", "", fmt.Errorf("no team or personal configuration directory found")
	}

	var errs []string
	for _, dir := range paths {
		secretPath := filepath.Join(dir, "secrets.yaml")
		data, err := os.ReadFile(secretPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			errs = append(errs, fmt.Sprintf("%s: %v", secretPath, err))
			continue
		}

		var secretsConfig secretspkg.Config
		if err := yaml.Unmarshal(data, &secretsConfig); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", secretPath, err))
			continue
		}

		value, valueType, source, err := selectRepoSecretFromConfig(secretPath, secretsConfig, opts)
		if err == nil {
			return value, valueType, source, nil
		}
		errs = append(errs, err.Error())
	}

	if len(errs) == 0 {
		return "", "", "", fmt.Errorf("secret '%s' was not found in team config", opts.SecretName)
	}
	return "", "", "", fmt.Errorf("%s", strings.Join(errs, "; "))
}

func secretConfigPathsForRepo(ctx context.Context, ghClient *gh.GitHubClient, cfg *config.Configuration, opts *RepoSecretApplyOptions) []string {
	seen := map[string]bool{}
	paths := make([]string, 0, 3)
	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}

	if opts.Team != "" {
		if u, err := ghClient.GetUser(ctx, opts.Owner); err == nil && strings.EqualFold(u.GetType(), "Organization") {
			add(cfg.GetTeamDir(opts.Team, opts.Owner))
		}
		add(cfg.GetTeamDir(opts.Team, ""))
	}
	add(cfg.GetPersonalDir(opts.Owner))
	return paths
}

func selectRepoSecretFromConfig(secretPath string, secretsConfig secretspkg.Config, opts *RepoSecretApplyOptions) (string, string, string, error) {
	secretValues := buildResolvedSecretValues(secretsConfig)
	matches, repoLevelMatches := collectConfiguredSecretMatches(secretsConfig, opts.SecretName)

	selected, err := chooseConfiguredSecret(secretPath, opts, matches, repoLevelMatches)
	if err != nil {
		return "", "", "", err
	}
	if usesTemplateValue(selected.Value) {
		return "", "", "", fmt.Errorf("secret '%s' in %s uses Go template values; use 'vip repo apply' or pass --value/--from explicitly", opts.SecretName, secretPath)
	}

	value, source, ok := secretspkg.GetSecretValueAndSource(selected, secretValues)
	if !ok {
		return "", "", "", fmt.Errorf("secret '%s' could not be resolved from %s", opts.SecretName, secretPath)
	}

	valueType := selected.Type
	if valueType == "" {
		valueType = repoSecretTypeSecret
	}
	return value, valueType, source + " (" + secretPath + ")", nil
}

func buildResolvedSecretValues(secretsConfig secretspkg.Config) map[string]string {
	secretValues := make(map[string]string)
	for _, s := range secretsConfig.Secrets {
		if s.Name == "" || usesTemplateValue(s.Value) {
			continue
		}
		resolvedValue, _, _, err := secretspkg.ResolveSecretValue(s.Value, s.Name)
		if err == nil {
			secretValues[s.Name] = resolvedValue
		}
	}
	return secretValues
}

func collectConfiguredSecretMatches(secretsConfig secretspkg.Config, secretName string) ([]secretspkg.Secret, []secretspkg.Secret) {
	matches := make([]secretspkg.Secret, 0)
	repoLevelMatches := make([]secretspkg.Secret, 0)
	for _, s := range secretsConfig.Secrets {
		if s.Name != secretName {
			continue
		}
		matches = append(matches, s)
		if s.Env == "" {
			repoLevelMatches = append(repoLevelMatches, s)
		}
	}
	return matches, repoLevelMatches
}

func chooseConfiguredSecret(secretPath string, opts *RepoSecretApplyOptions, matches, repoLevelMatches []secretspkg.Secret) (secretspkg.Secret, error) {
	if opts.Env != "" {
		return chooseConfiguredEnvSecret(secretPath, opts, matches)
	}
	if len(repoLevelMatches) == 1 {
		return repoLevelMatches[0], nil
	}
	if len(repoLevelMatches) > 1 {
		return secretspkg.Secret{}, fmt.Errorf("multiple repo-level secret entries found for '%s' in %s", opts.SecretName, secretPath)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return secretspkg.Secret{}, fmt.Errorf("multiple secret entries found for '%s' in %s; use --env to disambiguate", opts.SecretName, secretPath)
	}
	return secretspkg.Secret{}, fmt.Errorf("secret '%s' not found in %s", opts.SecretName, secretPath)
}

func chooseConfiguredEnvSecret(secretPath string, opts *RepoSecretApplyOptions, matches []secretspkg.Secret) (secretspkg.Secret, error) {
	filtered := make([]secretspkg.Secret, 0)
	for _, s := range matches {
		if s.Env == opts.Env {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return secretspkg.Secret{}, fmt.Errorf("secret '%s' with env '%s' not found in %s", opts.SecretName, opts.Env, secretPath)
	}
	if len(filtered) > 1 {
		return secretspkg.Secret{}, fmt.Errorf("multiple secret entries found for '%s' with env '%s' in %s", opts.SecretName, opts.Env, secretPath)
	}
	return filtered[0], nil
}

func usesTemplateValue(value string) bool {
	return strings.Contains(value, "{{.") || strings.Contains(value, "{{ .")
}

func normalizeRepoSecretType(valueType string) (string, error) {
	if valueType == "" {
		return "", nil
	}
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case repoSecretTypeSecret:
		return repoSecretTypeSecret, nil
	case repoSecretTypeVariable:
		return repoSecretTypeVariable, nil
	default:
		return "", fmt.Errorf("unsupported --type value %q (expected secret or variable)", valueType)
	}
}
