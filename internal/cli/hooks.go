package cli

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/output"
	projectpkg "github.com/nentgroup/viaplay-cli/internal/project"
	templatepkg "github.com/nentgroup/viaplay-cli/internal/template"
)

const sampleHookScriptName = "example-post-install.sh"

var unresolvedHookTemplatePattern = regexp.MustCompile(`{{[^}]+}}`)

type hooksCommandOptions struct {
	Language    string
	ProjectType string
	Path        string
	Name        string
	Owner       string
	Team        string
	Description string
}

type hookDoctorReport struct {
	checks   []string
	warnings []string
	errors   []string
}

func (r *hookDoctorReport) addCheck(message string) {
	r.checks = append(r.checks, message)
}

func (r *hookDoctorReport) addWarning(message string) {
	r.warnings = append(r.warnings, message)
}

func (r *hookDoctorReport) addError(message string) {
	r.errors = append(r.errors, message)
}

func (r *hookDoctorReport) hasErrors() bool {
	return len(r.errors) > 0
}

// NewHooksCommand creates the hooks command with subcommands.
func NewHooksCommand() *cobra.Command {
	hooksCmd := &cobra.Command{
		Use:   "hooks",
		Short: "Inspect and run post-install hooks",
		Long:  `Inspect, validate, run, and scaffold post-install hooks defined in template configuration.`,
	}

	hooksCmd.AddCommand(newHooksListCommand())
	hooksCmd.AddCommand(newHooksDoctorCommand())
	hooksCmd.AddCommand(newHooksRunCommand())
	hooksCmd.AddCommand(newHooksInitCommand())

	return hooksCmd
}

func newHooksListCommand() *cobra.Command {
	opts := &hooksCommandOptions{}

	cmd := &cobra.Command{
		Use:   "list [<language>/<type>]",
		Short: "List configured hooks for a template",
		Long:  `List the post-install hook commands and scripts configured for a template.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			language, projectType, err := resolveHookTemplateRef(args, opts)
			if err != nil {
				return err
			}

			cfg, err := loadConfigWithTeamOverrides(valueOrDefault(opts.Team, viper.GetString("default_team")), "")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			hooks := cfg.GetPostInstallHooks(language, projectType)
			if len(hooks) == 0 {
				output.InfoMessage(fmt.Sprintf("No hooks configured for %s/%s", language, projectType))
				return nil
			}

			renderer, err := buildHooksRenderer(opts, language, projectType)
			if err != nil {
				return err
			}

			output.Section(fmt.Sprintf("Hooks for %s/%s", language, projectType))
			rows, warnings := buildHookListRows(cfg, hooks, renderer)
			fmt.Print(output.Table([]string{"Type", "Configured", "Preview"}, rows, 0))
			printHookListPreviewNote(cmd)
			for _, warning := range warnings {
				output.WarningMessage(warning)
			}

			return nil
		},
	}

	addSharedHooksFlags(cmd, opts)
	return cmd
}

func newHooksDoctorCommand() *cobra.Command {
	opts := &hooksCommandOptions{}

	cmd := &cobra.Command{
		Use:          "doctor [<language>/<type>]",
		Short:        "Validate hooks for a template",
		Long:         `Validate that configured hook commands render correctly and referenced scripts exist and are executable.`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			language, projectType, err := resolveHookTemplateRef(args, opts)
			if err != nil {
				return err
			}

			cfg, err := loadConfigWithTeamOverrides(valueOrDefault(opts.Team, viper.GetString("default_team")), "")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			hooks := cfg.GetPostInstallHooks(language, projectType)
			if len(hooks) == 0 {
				output.InfoMessage(fmt.Sprintf("No hooks configured for %s/%s", language, projectType))
				return nil
			}

			renderer, err := buildHooksRenderer(opts, language, projectType)
			if err != nil {
				return err
			}

			report := validateHooks(cfg, hooks, renderer)
			printHookDoctorReport(language, projectType, report)
			if report.hasErrors() {
				return fmt.Errorf("hook validation failed")
			}

			return nil
		},
	}

	addSharedHooksFlags(cmd, opts)
	return cmd
}

func newHooksRunCommand() *cobra.Command {
	opts := &hooksCommandOptions{}

	cmd := &cobra.Command{
		Use:          "run [<language>/<type>]",
		Short:        "Run post-install hooks for a template",
		Long:         `Run the configured post-install hooks for a template against a target directory.`,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			language, projectType, err := resolveHookTemplateRef(args, opts)
			if err != nil {
				return err
			}

			cfg, err := loadConfigWithTeamOverrides(valueOrDefault(opts.Team, viper.GetString("default_team")), "")
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			projectPath, err := resolveHooksProjectPath(opts.Path)
			if err != nil {
				return err
			}

			vars, err := buildHookTemplateVars(projectPath, opts, language, projectType)
			if err != nil {
				return err
			}

			factory := projectpkg.NewFactory(nil, output.NewNoopReporter(), cfg)
			return factory.RunHooks(cmd.Context(), projectPath, language, projectType, vars)
		},
	}

	addSharedHooksFlags(cmd, opts)
	return cmd
}

func newHooksInitCommand() *cobra.Command {
	var override bool

	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Scaffold the hooks directory",
		Long:         `Create the global hooks directory and a sample post-install script.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			hooksDir := cfg.GetHooksDir()
			if err := os.MkdirAll(hooksDir, 0o755); err != nil {
				return fmt.Errorf("failed to create hooks directory: %w", err)
			}

			scriptPath := filepath.Join(hooksDir, sampleHookScriptName)
			if _, err := os.Stat(scriptPath); err == nil && !override {
				output.InfoMessage("Hooks directory is ready")
				fmt.Printf("Sample script already exists: %s\n", output.Bold(scriptPath))
				fmt.Printf("Use %s to replace it.\n", output.Bold("vip hooks init --override"))
				return nil
			}

			if err := os.WriteFile(scriptPath, []byte(sampleHookScriptContent), 0o600); err != nil {
				return fmt.Errorf("failed to write sample hook script: %w", err)
			}
			if err := os.Chmod(scriptPath, 0o755); err != nil {
				return fmt.Errorf("failed to mark sample hook script executable: %w", err)
			}

			output.SuccessMessage("Hooks directory scaffolded successfully")
			fmt.Printf("Hooks directory: %s\n", output.Bold(hooksDir))
			fmt.Printf("Sample script:   %s\n", output.Bold(scriptPath))
			return nil
		},
	}

	cmd.Flags().BoolVar(&override, "override", false, "Override the sample hook script if it already exists")
	return cmd
}

func addSharedHooksFlags(cmd *cobra.Command, opts *hooksCommandOptions) {
	cmd.Flags().StringVar(&opts.Language, "language", "", "Template language (falls back to default_language)")
	cmd.Flags().StringVar(&opts.ProjectType, "type", "", "Template type (falls back to default_type)")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Target project directory")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Project name for template variable rendering (defaults to directory name)")
	cmd.Flags().StringVar(&opts.Owner, "owner", "", "Repository owner for template variable rendering")
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name for template variable rendering (falls back to default_team)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Project description for template variable rendering")
}

func printHookListPreviewNote(cmd *cobra.Command) {
	if hasExplicitHookPreviewInput(cmd) {
		output.InfoMessage("Preview shows hooks rendered with the context flags you provided.")
		return
	}

	output.InfoMessage("Preview uses inferred defaults from the current directory and config. It is not the exact runtime resolution from 'vip project create'.")
}

func hasExplicitHookPreviewInput(cmd *cobra.Command) bool {
	for _, flagName := range []string{"language", "type", "path", "name", "owner", "team", "description"} {
		if cmd.Flags().Changed(flagName) {
			return true
		}
	}

	return false
}

func parseHookTemplateRef(value string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid template reference %q (expected <language>/<type>)", value)
	}

	return parts[0], parts[1], nil
}

func resolveHookTemplateRef(args []string, opts *hooksCommandOptions) (string, string, error) {
	if len(args) == 1 {
		return parseHookTemplateRef(args[0])
	}

	language := opts.Language
	if language == "" {
		language = viper.GetString("default_language")
	}

	projectType := opts.ProjectType
	if projectType == "" {
		projectType = viper.GetString("default_type")
	}

	if language == "" || projectType == "" {
		return "", "", fmt.Errorf("template reference is required (pass <language>/<type> or set --language/--type, or configure default_language/default_type)")
	}

	return language, projectType, nil
}

func buildHooksRenderer(opts *hooksCommandOptions, language, projectType string) (*templatepkg.Renderer, error) {
	projectPath, err := resolveHooksProjectPath(opts.Path)
	if err != nil {
		return nil, err
	}

	vars, err := buildHookTemplateVars(projectPath, opts, language, projectType)
	if err != nil {
		return nil, err
	}

	return templatepkg.NewRenderer(vars), nil
}

func resolveHooksProjectPath(pathValue string) (string, error) {
	absPath, err := filepath.Abs(pathValue)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", pathValue, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat %s: %w", absPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path must be a directory: %s", absPath)
	}

	return absPath, nil
}

func buildHookTemplateVars(projectPath string, opts *hooksCommandOptions, language, projectType string) (*templatepkg.Variables, error) {
	projectName := opts.Name
	if projectName == "" {
		projectName = filepath.Base(projectPath)
	}
	if projectName == "" || projectName == "." || projectName == string(filepath.Separator) {
		projectName = "project"
	}

	owner := opts.Owner
	if owner == "" {
		owner = viper.GetString("github.username")
	}
	if owner == "" {
		currentUser, err := user.Current()
		if err == nil {
			owner = currentUser.Username
		}
	}
	if owner == "" {
		owner = "owner"
	}

	team := opts.Team
	if team == "" {
		team = viper.GetString("default_team")
	}

	description := opts.Description
	orgName := viper.GetString("default_organization")

	vars := templatepkg.NewTemplateVariables()
	vars.Project.Name = projectName
	vars.Project.Description = description
	vars.Project.Language = language
	vars.Project.Type = projectType
	vars.Repo.Owner = owner
	vars.Repo.Name = projectName
	vars.Repo.URL = fmt.Sprintf("https://github.com/%s/%s", owner, projectName)
	vars.Service.Name = projectName
	vars.Service.Owner = valueOrDefault(team, owner)
	vars.Service.OwnerKey = strings.ToLower(strings.ReplaceAll(vars.Service.Owner, " ", "-"))
	vars.Service.Type = projectType
	vars.Go.BinaryName = projectName
	vars.Go.ModulePath = fmt.Sprintf("github.com/%s/%s", owner, projectName)
	vars.Docker.ImageName = strings.ToLower(projectName)
	vars.Org.Name = orgName
	vars.Org.Team = team
	vars.Meta.CreatedBy = owner

	return vars, nil
}

func buildHookListRows(
	cfg *config.Configuration,
	hooks []*config.PostInstallHook,
	renderer *templatepkg.Renderer,
) ([][]string, []string) {
	rows := [][]string{}
	warnings := []string{}

	for _, hook := range hooks {
		for _, command := range hook.GetAllCommands() {
			renderedCommand, err := renderer.RenderString(command)
			if err != nil {
				renderedCommand = "render error: " + err.Error()
			} else if containsUnresolvedHookTemplate(renderedCommand) {
				warnings = append(warnings, fmt.Sprintf("command contains unresolved template expressions: %s", command))
			}

			rows = append(rows, []string{"cmd", command, renderedCommand})
		}

		for _, scriptPath := range hook.GetAllScripts() {
			renderedScriptPath, err := renderer.RenderString(scriptPath)
			resolvedPath := ""
			if err != nil {
				resolvedPath = "render error: " + err.Error()
			} else {
				resolvedPath = cfg.ResolveHookScriptPath(renderedScriptPath)
				if containsUnresolvedHookTemplate(renderedScriptPath) {
					warnings = append(warnings, fmt.Sprintf("script path contains unresolved template expressions: %s", scriptPath))
				}
			}

			rows = append(rows, []string{"script", scriptPath, resolvedPath})
		}
	}

	return rows, warnings
}

func validateHooks(
	cfg *config.Configuration,
	hooks []*config.PostInstallHook,
	renderer *templatepkg.Renderer,
) *hookDoctorReport {
	report := &hookDoctorReport{}

	for _, hook := range hooks {
		for _, command := range hook.GetAllCommands() {
			renderedCommand, err := renderer.RenderString(command)
			if err != nil {
				report.addError(fmt.Sprintf("command failed to render: %s (%v)", command, err))
				continue
			}

			report.addCheck(fmt.Sprintf("command renders: %s", renderedCommand))
			if containsUnresolvedHookTemplate(renderedCommand) {
				report.addWarning(fmt.Sprintf("command contains unresolved template expressions: %s", command))
			}
		}

		for _, scriptPath := range hook.GetAllScripts() {
			renderedScriptPath, err := renderer.RenderString(scriptPath)
			if err != nil {
				report.addError(fmt.Sprintf("script path failed to render: %s (%v)", scriptPath, err))
				continue
			}

			resolvedPath := cfg.ResolveHookScriptPath(renderedScriptPath)
			info, err := os.Stat(resolvedPath)
			if err != nil {
				report.addError(fmt.Sprintf("script not found: %s", resolvedPath))
				continue
			}

			report.addCheck(fmt.Sprintf("script exists: %s", resolvedPath))
			if info.Mode().Perm()&0o111 == 0 {
				report.addError(fmt.Sprintf("script is not executable: %s", resolvedPath))
			} else {
				report.addCheck(fmt.Sprintf("script is executable: %s", resolvedPath))
			}

			data, err := os.ReadFile(resolvedPath)
			if err != nil {
				report.addError(fmt.Sprintf("failed to read script: %s (%v)", resolvedPath, err))
				continue
			}

			renderedScript, err := renderer.RenderString(string(data))
			if err != nil {
				report.addError(fmt.Sprintf("script failed to render: %s (%v)", resolvedPath, err))
				continue
			}

			report.addCheck(fmt.Sprintf("script renders: %s", resolvedPath))
			if containsUnresolvedHookTemplate(renderedScript) {
				report.addWarning(fmt.Sprintf("script contains unresolved template expressions: %s", resolvedPath))
			}
		}
	}

	return report
}

func printHookDoctorReport(language, projectType string, report *hookDoctorReport) {
	output.Section(fmt.Sprintf("Hook Validation for %s/%s", language, projectType))
	for _, check := range report.checks {
		fmt.Printf("  [ok] %s\n", check)
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

func containsUnresolvedHookTemplate(value string) bool {
	return unresolvedHookTemplatePattern.MatchString(value)
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

const sampleHookScriptContent = `#!/bin/sh
set -eu

echo "Running post-install hook for {{ .Project.Name }}"
echo "Template: {{ .Project.Language }}/{{ .Project.Type }}"
`
