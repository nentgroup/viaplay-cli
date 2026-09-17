// Package cli provides the command-line interface for viaplay-cli.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/scaffolding"
	"github.com/nentgroup/viaplay-cli/internal/template"
)

// TestResult represents the result of a template test
type TestResult struct {
	Success      bool   `json:"success"`
	OutputPath   string `json:"outputPath"`
	TemplatePath string `json:"templatePath"`
	Error        string `json:"error,omitempty"`
}

type templateTestCommandOptions struct {
	templatePath       string
	forceRefresh       bool
	projectName        string
	projectOwner       string
	jsonOutput         bool
	templateSet        []string
	noInput            bool
	allowTemplateHooks bool
}

// createTempDir creates a temporary directory with the given prefix
// and returns the path to that directory. The directory will be created
// inside the OS's temporary directory.
func createTempDir(prefix string) (string, error) {
	return os.MkdirTemp(os.TempDir(), prefix)
}

// NewTemplateCommand creates a template command with subcommands
func NewTemplateCommand() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Commands for working with templates",
		Long:  `Manage and test templates for project scaffolding, including local template copies.`,
	}

	// Add subcommands
	templateCmd.AddCommand(newTemplateListCommand())
	templateCmd.AddCommand(newTemplateUpdateCommand())
	templateCmd.AddCommand(newTemplatePruneCommand())
	templateCmd.AddCommand(newTemplateCleanCommand())
	templateCmd.AddCommand(newTemplateTestCommand())
	templateCmd.AddCommand(newTemplateInspectCommand())
	templateCmd.AddCommand(newTemplateAddCommand())
	templateCmd.AddCommand(newTemplateRemoveCommand())

	return templateCmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   cmdList,
		Short: "List local template copies and storage information",
		Long:  `List template copies currently stored locally for reuse, along with overall storage statistics.`,
		Run: func(cmd *cobra.Command, args []string) {
			listCache(cmd.Context())
		},
	}
}

func newTemplatePruneCommand() *cobra.Command {
	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove old local template copies",
		Long:  `Remove local template copies that have not been used in a specified number of days.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			days, err := cmd.Flags().GetInt("days")
			if err != nil {
				return fmt.Errorf("failed to get 'days' flag: %w", err)
			}

			pruneCache(days)
			return nil
		},
	}

	pruneCmd.Flags().Int("days", 30, "Prune local template copies older than specified days")
	return pruneCmd
}

func newTemplateUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Refresh all local template copies",
		Long:  `Fetch or refresh all locally stored template copies from their configured sources.`,
		Run: func(cmd *cobra.Command, args []string) {
			updateTemplates(cmd.Context())
		},
	}
}

func newTemplateCleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove all local template copies",
		Long:  `Remove all locally stored template copies.`,
		Run: func(cmd *cobra.Command, args []string) {
			cleanCache(cmd.Context())
		},
	}
}

// newTemplateTestCommand creates a new test subcommand for the template command
func newTemplateTestCommand() *cobra.Command {
	opts := &templateTestCommandOptions{}

	testCmd := &cobra.Command{
		Use:   "test [source]",
		Short: "Test template scaffolding without project creation",
		Long: `Test templates directly without creating repos or authenticating.
This command is useful for template developers who want to test their 
templates during development or in CI pipelines.

Templates are output to a temporary directory that is automatically created.

<source> accepts the same template source formats as 'project create' and
'template inspect': a GitHub address (e.g. github.com/owner/repo,
https://github.com/owner/repo, or github@owner/repo[@ref]), a local path, or
an explicit local@/url@/git@ source. The --template-path flag is kept for
backwards compatibility and is equivalent to passing <source>.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplateTestCommand(cmd, args, opts)
		},
	}

	// Add flags
	testCmd.Flags().StringVar(&opts.templatePath, "template-path", "",
		"Template source to test (deprecated; pass <source> as a positional argument instead)")
	testCmd.Flags().BoolVar(&opts.forceRefresh, "force", false, "Force refresh of local template copies")
	testCmd.Flags().StringVar(&opts.projectName, "name", "test-project", "Project name for template variables")
	testCmd.Flags().StringVar(&opts.projectOwner, "owner", "test-owner", "Project owner for template variables")
	testCmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results in JSON format for scripting")
	testCmd.Flags().StringArrayVar(&opts.templateSet, "set", nil, "Set a template option using key=value")
	testCmd.Flags().BoolVar(&opts.noInput, "no-input", false, "Do not prompt for template options; use defaults/--set values")
	testCmd.Flags().BoolVar(&opts.allowTemplateHooks, "allow-template-hooks", false,
		"Allow executing template-defined manifest hooks (hooks.post)")
	if err := testCmd.Flags().MarkDeprecated("template-path",
		"pass <source> as a positional argument instead, e.g. 'vip template test <source>'"); err != nil {
		panic(err)
	}
	if err := testCmd.Flags().MarkDeprecated("force",
		"remote templates are now always fetched fresh for 'template test' and never persisted to the shared template cache"); err != nil {
		panic(err)
	}

	return testCmd
}

func runTemplateTestCommand(cmd *cobra.Command, args []string, opts *templateTestCommandOptions) error {
	templateSource, err := resolveTestTemplateSource(args, opts.templatePath)
	if err != nil {
		return err
	}

	result := TestResult{TemplatePath: templateSource}
	cfg, err := config.LoadConfig()
	if err != nil {
		return emitTemplateTestFailure(&result, opts.jsonOutput, "failed to load configuration: %v", err)
	}

	outputPath, absOutputPath, err := createTemplateTestOutputPath()
	if err != nil {
		return emitTemplateTestFailure(&result, opts.jsonOutput, "failed to create temporary directory: %v", err)
	}
	result.OutputPath = outputPath

	if !opts.jsonOutput {
		fmt.Fprintln(os.Stderr, "🧪 Testing template scaffolding...")
		fmt.Fprintf(os.Stderr, "📁 Output directory: %s\n", outputPath)
	}

	vars := buildTemplateTestVariables(opts.projectName, opts.projectOwner)
	scaffolder := scaffolding.NewProjectScaffolder(cache.NewManager(cfg), cfg)
	manifest, err := scaffolder.ScaffoldProjectWithOptions(cmd.Context(), absOutputPath, "", "", templateSource, vars, false,
		opts.forceRefresh, opts.templateSet, opts.noInput)
	if err != nil {
		return emitTemplateTestFailure(&result, opts.jsonOutput, "failed to scaffold template: %v", err)
	}

	if err := maybeRunTemplateTestManifestHooks(cmd, opts, &result, manifest, vars, absOutputPath); err != nil {
		return err
	}

	result.Success = true
	if opts.jsonOutput {
		printJSONResult(result)
		return nil
	}
	fmt.Fprintf(os.Stderr, "✅ Template successfully scaffolded to: %s\n", absOutputPath)
	return nil
}

func emitTemplateTestFailure(result *TestResult, jsonOutput bool, format string, args ...any) error {
	result.Error = fmt.Sprintf(format, args...)
	if jsonOutput {
		printJSONResult(*result)
		return nil
	}
	return fmt.Errorf("%s", result.Error)
}

func createTemplateTestOutputPath() (string, string, error) {
	timestamp := time.Now().Format("20060102-150405")
	tempDirPrefix := fmt.Sprintf("vip-template-test-%s-", timestamp)
	outputPath, err := createTempDir(tempDirPrefix)
	if err != nil {
		return "", "", err
	}
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", "", err
	}
	return outputPath, absOutputPath, nil
}

func maybeRunTemplateTestManifestHooks(
	cmd *cobra.Command,
	opts *templateTestCommandOptions,
	result *TestResult,
	manifest *template.Manifest,
	vars *template.Variables,
	absOutputPath string,
) error {
	if !opts.allowTemplateHooks || manifest == nil || len(manifest.Hooks.Post) == 0 {
		return nil
	}
	if opts.noInput {
		return emitTemplateTestFailure(result, opts.jsonOutput,
			"template manifest hooks require interactive double confirmation; rerun without --no-input")
	}

	confirmed, err := template.ConfirmManifestHooksExecution(manifest, absOutputPath, os.Stdin, os.Stdout)
	if err != nil {
		return emitTemplateTestFailure(result, opts.jsonOutput,
			"failed to read template manifest hook confirmation: %v", err)
	}
	if !confirmed {
		if !opts.jsonOutput {
			fmt.Fprintln(os.Stderr, "Skipping template manifest hooks: confirmation declined")
		}
		return nil
	}

	if err := template.ExecuteManifestHooks(cmd.Context(), absOutputPath, manifest, vars, os.Stdout, os.Stderr); err != nil {
		return emitTemplateTestFailure(result, opts.jsonOutput, "failed to run template manifest hooks: %v", err)
	}
	return nil
}

func buildTemplateTestVariables(projectName, projectOwner string) *template.Variables {
	now := time.Now()
	return &template.Variables{
		Project: template.ProjectInfo{
			Name:        projectName,
			Description: "test-description",
		},
		Repo: template.RepoInfo{
			Name:  projectName,
			Owner: projectOwner,
		},
		Service: template.ServiceInfo{
			Name:  projectName,
			Owner: projectOwner,
			Port:  "8080",
			Type:  "http",
		},
		Org: template.OrgInfo{
			Name:       "test-org",
			Team:       "test-team",
			CIProvider: "github",
		},
		Meta: template.MetaInfo{
			CreatedAt: now,
			CreatedBy: "viaplay-cli",
			Year:      now.Year(),
		},
		Go: template.GoInfo{
			Version:    "1.24",
			ModulePath: fmt.Sprintf("github.com/%s/%s", projectOwner, projectName),
			BinaryName: projectName,
		},
		Rust: template.RustInfo{
			Version: "1.88",
			Edition: "2024",
		},
		Node: template.NodeInfo{
			Version: "20",
		},
		Cloud: template.CloudInfo{
			Provider: template.DefaultCloudProvider,
		},
		Docker: template.DockerInfo{
			ImageName: projectName,
			ImageTag:  "latest",
		},
		Env: template.EnvInfo{
			Default:      "dev",
			Environments: []string{"dev", "staging", "production"},
		},
	}
}

// resolveTestTemplateSource determines the template source for 'template test'
// from either the new positional <source> argument or the deprecated
// --template-path flag, rejecting the case where both are given with
// different values to avoid silently picking one.
func resolveTestTemplateSource(args []string, templatePathFlag string) (string, error) {
	var positional string
	if len(args) > 0 {
		positional = args[0]
	}

	switch {
	case positional != "" && templatePathFlag != "" && positional != templatePathFlag:
		return "", fmt.Errorf("both <source> (%q) and --template-path (%q) were given with different values; use only one",
			positional, templatePathFlag)
	case positional != "":
		return positional, nil
	case templatePathFlag != "":
		return templatePathFlag, nil
	default:
		return "", fmt.Errorf("a template source is required; pass it as <source> (e.g. 'vip template test <source>')")
	}
}

// printJSONResult outputs the test result as JSON to stdout
func printJSONResult(result TestResult) {
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// newTemplateInspectCommand creates a new "inspect" subcommand that inspects a
// template and lists the manifest-driven options (if any) it supports, without
// scaffolding anything. Templates without a manifest file are reported as having
// no configurable options, preserving backward compatibility.
func newTemplateInspectCommand() *cobra.Command {
	var (
		forceRefresh bool
		jsonOutput   bool
	)

	inspectCmd := &cobra.Command{
		Use:   "inspect <source>",
		Short: "Inspect a template and list the manifest-defined options it supports",
		Long: `Inspect a template's manifest (.vip.yaml), if present, and print the
options and variables it exposes (key, type, prompt, default, choices).

<source> accepts the same template source formats as 'project create' and
'template test': a GitHub address (e.g. github.com/owner/repo,
https://github.com/owner/repo, or github@owner/repo[@ref]), a local path, or
an explicit local@/url@/git@ source.

Use --set key=value with 'project create' or 'template test' to set these
options non-interactively. Templates without a manifest report no options.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			source := args[0]
			// source is expanded/normalised by cache.ParseSource (e.g. bare paths
			// default to "local@", GitHub URLs are recognised automatically).

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			cacheManager := cache.NewManager(cfg)
			scaffolder := scaffolding.NewProjectScaffolder(cacheManager, cfg)

			manifest, err := scaffolder.GetTemplateManifest(ctx, "", "", source, forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to inspect template: %w", err)
			}

			if jsonOutput {
				printTemplateManifestJSON(manifest)
				return nil
			}
			printTemplateManifestHuman(manifest)
			return nil
		},
	}

	inspectCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of local template copies")
	inspectCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format for scripting")
	if err := inspectCmd.Flags().MarkDeprecated("force",
		"templates are now always fetched fresh for 'template inspect' and never persisted to the shared template cache"); err != nil {
		panic(err)
	}

	return inspectCmd
}

// newTemplateAddCommand registers a template source under
// templates.<language>.<type> in the user's personal config file.
func newTemplateAddCommand() *cobra.Command {
	var (
		language  string
		typeFlag  string
		team      string
		force     bool
		skipCache bool
	)

	addCmd := &cobra.Command{
		Use:   "add <source>",
		Short: "Register a template source under templates.<language>.<type>",
		Long: `Register a template source under templates.<language>.<type> and cache it
locally, so it can be used with 'vip project create --language <language>
--type <type>'. Registers under your team's config if one is set up
(--team/default_team), otherwise your personal config.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			source := args[0]

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			cacheManager := cache.NewManager(cfg)
			scaffolder := scaffolding.NewProjectScaffolder(cacheManager, cfg)

			manifest, err := scaffolder.GetTemplateManifest(ctx, "", "", source, false)
			if err != nil {
				return fmt.Errorf("failed to inspect template: %w", err)
			}
			if issues := template.ValidateManifest(manifest); len(issues) > 0 {
				return fmt.Errorf("refusing to register %q: template manifest (.vip.yaml) is invalid:\n  - %s",
					source, strings.Join(issues, "\n  - "))
			}

			resolvedLanguage := strings.TrimSpace(language)
			resolvedType := strings.TrimSpace(typeFlag)
			if manifest != nil {
				if resolvedLanguage == "" {
					resolvedLanguage = strings.TrimSpace(manifest.Metadata.Language)
				}
				if resolvedType == "" {
					resolvedType = strings.TrimSpace(manifest.Metadata.Type)
				}
			}

			if resolvedLanguage == "" || resolvedType == "" {
				return fmt.Errorf(
					"could not determine language/type for %q: pass --language and --type, "+
						"or add metadata.language/metadata.type to the template's .vip.yaml", source)
			}

			resolvedTeam := valueOrDefault(team, viper.GetString("default_team"))
			resolvedOrg := resolveTeamConfigOrganization(cfg, "")

			targetFile := cfg.ConfigFile
			targetDescription := fmt.Sprintf("personal config file (%s)", cfg.ConfigFile)
			var existing *config.TemplateDefinition

			teamConfigFile, err := cfg.FindTeamTemplateConfigFile(resolvedTeam, resolvedOrg)
			if err != nil {
				return fmt.Errorf("failed to locate team config file: %w", err)
			}
			if teamConfigFile != "" {
				targetFile = teamConfigFile
				targetDescription = fmt.Sprintf("team %q config file (%s)", resolvedTeam, teamConfigFile)
				existing, err = cfg.GetTeamTemplateOverride(resolvedTeam, resolvedOrg, resolvedLanguage, resolvedType)
				if err != nil {
					return fmt.Errorf("failed to check existing team template config: %w", err)
				}
			} else if resolvedTeam != "" {
				output.InfoMessage(fmt.Sprintf(
					"Team %q has no config directory set up; registering in your personal config file instead. "+
						"Run 'vip config init team %s' (or 'vip config pull team %s') to set one up.",
					resolvedTeam, resolvedTeam, resolvedTeam))
				existing = cfg.GetTemplate(resolvedLanguage, resolvedType)
			} else {
				existing = cfg.GetTemplate(resolvedLanguage, resolvedType)
			}

			if existing != nil && !force {
				return fmt.Errorf("templates.%s.%s is already registered in the %s (source: %s); pass --force to overwrite",
					resolvedLanguage, resolvedType, targetDescription, existing.Source)
			}

			if err := config.UpdateTemplateConfigFile(targetFile, resolvedLanguage, resolvedType, source); err != nil {
				return fmt.Errorf("failed to update config file: %w", err)
			}

			output.SuccessMessage(fmt.Sprintf(
				"Registered %s/%s -> %s in the %s", resolvedLanguage, resolvedType, source, targetDescription))

			// Local sources are used directly from their path and never go through
			// the cache (see cache.Manager.EnsureTemplate), so there's nothing to warm.
			isLocalSource := false
			if parsedSource, err := cache.ParseSource(source); err == nil {
				isLocalSource = parsedSource.Type == cache.SourceTypeLocal
			}

			if !skipCache && !isLocalSource {
				if _, err := cacheManager.EnsureTemplate(ctx, resolvedLanguage, resolvedType, source, true); err != nil {
					output.WarningMessage(fmt.Sprintf(
						"Failed to cache the template now: %v (it will be cloned automatically on first use)", err))
				} else {
					output.InfoMessage(fmt.Sprintf(
						"Cached %s/%s locally (see 'vip template list')", resolvedLanguage, resolvedType))
				}
			}

			fmt.Printf("Use it with: vip project create <name> --language %s --type %s\n", resolvedLanguage, resolvedType)
			return nil
		},
	}

	addCmd.Flags().StringVar(&language, "language", "",
		"Programming language to register the template under (overrides manifest metadata)")
	addCmd.Flags().StringVar(&typeFlag, "type", "",
		"Project type to register the template under (overrides manifest metadata)")
	addCmd.Flags().StringVar(&team, "team", "",
		"Team to register the template under, if it has a config directory (falls back to default_team, then personal config)")
	addCmd.Flags().BoolVar(&force, "force", false,
		"Overwrite an existing templates.<language>.<type> entry")
	addCmd.Flags().BoolVar(&skipCache, "skip-cache", false,
		"Register the mapping without cloning it into the local template cache now")

	return addCmd
}

// newTemplateRemoveCommand removes a templates.<language>.<type> mapping from
// the user's personal config file, or a team's config file when --team (or
// default_team) resolves to one, mirroring 'vip template add' targeting. It
// also removes any locally cached clone for that language/type.
func newTemplateRemoveCommand() *cobra.Command {
	var (
		team      string
		keepCache bool
	)

	removeCmd := &cobra.Command{
		Use:   "remove <language>/<type>",
		Short: "Remove a registered template mapping",
		Long: `Remove a templates.<language>.<type> entry and its cached clone. Removes
from your team's config if --team/default_team applies, otherwise your
personal config.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			language, projectType, err := parseHookTemplateRef(args[0])
			if err != nil {
				return err
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			resolvedTeam := valueOrDefault(team, viper.GetString("default_team"))
			resolvedOrg := resolveTeamConfigOrganization(cfg, "")

			targetFile := cfg.ConfigFile
			targetDescription := fmt.Sprintf("personal config file (%s)", cfg.ConfigFile)

			teamConfigFile, err := cfg.FindTeamTemplateConfigFile(resolvedTeam, resolvedOrg)
			if err != nil {
				return fmt.Errorf("failed to locate team config file: %w", err)
			}
			if teamConfigFile != "" {
				targetFile = teamConfigFile
				targetDescription = fmt.Sprintf("team %q config file (%s)", resolvedTeam, teamConfigFile)
			} else if resolvedTeam != "" {
				output.InfoMessage(fmt.Sprintf(
					"Team %q has no config directory set up; looking in your personal config file instead.",
					resolvedTeam))
			}

			removed, err := config.RemoveTemplateConfigFile(targetFile, language, projectType)
			if err != nil {
				return fmt.Errorf("failed to update config file: %w", err)
			}
			if !removed {
				return fmt.Errorf("no templates.%s.%s entry found in the %s", language, projectType, targetDescription)
			}

			output.SuccessMessage(fmt.Sprintf("Removed %s/%s from the %s", language, projectType, targetDescription))

			if !keepCache {
				cacheManager := cache.NewManager(cfg)
				cacheRemoved, err := cacheManager.RemoveTemplate(language, projectType)
				if err != nil {
					output.WarningMessage(fmt.Sprintf("Failed to remove cached template copy: %v", err))
				} else if cacheRemoved {
					output.InfoMessage(fmt.Sprintf("Removed cached copy of %s/%s", language, projectType))
				}
			}

			return nil
		},
	}

	removeCmd.Flags().StringVar(&team, "team", "",
		"Team to remove the template from, if it has a config directory (falls back to default_team, then personal config)")
	removeCmd.Flags().BoolVar(&keepCache, "keep-cache", false,
		"Do not remove the locally cached clone for this language/type")

	return removeCmd
}

// printTemplateManifestJSON prints the manifest options/variables as JSON, or an
// empty/null manifest when the template has none.
func printTemplateManifestJSON(manifest *template.Manifest) {
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// manifestAuthorNames formats each author as "Name <email>", "Name", or just
// the email when no name is set, for display in 'template show'.
func manifestAuthorNames(authors []template.ManifestAuthor) []string {
	names := make([]string, 0, len(authors))
	for _, a := range authors {
		switch {
		case a.Name != "" && a.Email != "":
			names = append(names, fmt.Sprintf("%s <%s>", a.Name, a.Email))
		case a.Name != "":
			names = append(names, a.Name)
		case a.Email != "":
			names = append(names, a.Email)
		}
	}
	return names
}

// printTemplateManifestHuman prints a human-readable summary of the manifest's
// options and variables, styled consistently with 'template list' (section
// headers, tables, and shared colour helpers from internal/output).
func printTemplateManifestHuman(manifest *template.Manifest) {
	if manifest == nil {
		output.InfoMessage("This template has no manifest (.vip.yaml) — no configurable options.")
		return
	}

	title := strings.TrimSpace(manifest.Metadata.Name)
	if title == "" {
		title = "Template Options"
	}
	output.Section(title)

	if manifest.Metadata.Description != "" {
		fmt.Println(manifest.Metadata.Description)
	}
	if len(manifest.Metadata.Authors) > 0 {
		fmt.Printf("Authors: %s\n", strings.Join(manifestAuthorNames(manifest.Metadata.Authors), ", "))
	}
	if issues := template.ValidateManifest(manifest); len(issues) > 0 {
		output.WarningMessage("This template's manifest (.vip.yaml) has validation issues " +
			"and will be rejected by 'project create'/'template test':")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	if len(manifest.Options) == 0 && len(manifest.Variables) == 0 {
		output.InfoMessage("This template defines a manifest but no options or variables.")
		return
	}

	headers := []string{tableKeyHeader, colType, "Required", "Default", "Description"}
	if len(manifest.Options) > 0 {
		printManifestTable("Options (use --set key=value)", headers, manifestOptionRows(manifest.Options))
	}
	if len(manifest.Variables) > 0 {
		printManifestTable("Variables (use --set key=value)", headers, manifestVariableRows(manifest.Variables))
	}
}

// printManifestTable prints a titled table of manifest options or variables.
func printManifestTable(title string, headers []string, rows [][]string) {
	fmt.Printf("\n%s\n", output.Bold(title))
	fmt.Print(output.Table(headers, rows, 2))
}

// manifestOptionRows builds table rows for manifest options, folding each
// option's choices into the description column since they don't fit their
// own column alongside free-form variables.
func manifestOptionRows(options []template.ManifestOption) [][]string {
	rows := make([][]string, 0, len(options))
	for _, opt := range options {
		rows = append(rows, []string{
			output.Bold(opt.Key),
			opt.Type,
			requiredCell(opt.Required),
			defaultCell(opt.Default),
			descriptionWithChoices(opt.Description, opt.Choices),
		})
	}
	return rows
}

// manifestVariableRows builds table rows for manifest variables.
func manifestVariableRows(vars []template.ManifestVariable) [][]string {
	rows := make([][]string, 0, len(vars))
	for _, v := range vars {
		rows = append(rows, []string{
			output.Bold(v.Key),
			v.Type,
			requiredCell(v.Required),
			defaultCell(v.Default),
			v.Description,
		})
	}
	return rows
}

// requiredCell renders the "Required" column, colour-coded for quick scanning.
func requiredCell(required bool) string {
	if required {
		return output.Warning("yes")
	}
	return output.Faint("no")
}

// defaultCell renders the "Default" column, using a placeholder when unset.
func defaultCell(def any) string {
	if def == nil {
		return output.Faint("-")
	}
	return fmt.Sprintf("%v", def)
}

// descriptionWithChoices appends an option's choices (value and, if
// different, label) to its description for display in a single column.
func descriptionWithChoices(description string, choices []template.ManifestChoice) string {
	if len(choices) == 0 {
		return description
	}

	labels := make([]string, 0, len(choices))
	for _, c := range choices {
		if c.Label == "" || c.Label == c.Value {
			labels = append(labels, c.Value)
		} else {
			labels = append(labels, fmt.Sprintf("%s (%s)", c.Value, c.Label))
		}
	}

	choicesNote := output.Faint(fmt.Sprintf("[choices: %s]", strings.Join(labels, ", ")))
	if description == "" {
		return choicesNote
	}
	return description + " " + choicesNote
}
