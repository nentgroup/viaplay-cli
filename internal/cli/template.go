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
	var (
		templatePath string
		forceRefresh bool
		projectName  string
		projectOwner string
		jsonOutput   bool
		templateSet  []string
		noInput      bool
	)

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
			ctx := cmd.Context()

			templateSource, err := resolveTestTemplateSource(args, templatePath)
			if err != nil {
				return err
			}

			// Initialise result object for potential JSON output
			result := TestResult{
				TemplatePath: templateSource,
				Success:      false,
			}

			// Use a minimal configuration that's independent of user config
			cfg, err := config.LoadConfig()
			if err != nil {
				result.Error = fmt.Sprintf("failed to load configuration: %v", err)
				if jsonOutput {
					printJSONResult(result)
					return nil
				}
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialise cache manager
			cacheManager := cache.NewManager(cfg)
			// Initialise project scaffolder
			scaffolder := scaffolding.NewProjectScaffolder(cacheManager, cfg)

			// Create a temporary directory for output
			timestamp := time.Now().Format("20060102-150405")
			tempDirPrefix := fmt.Sprintf("vip-template-test-%s-", timestamp)
			outputPath, err := createTempDir(tempDirPrefix)
			if err != nil {
				result.Error = fmt.Sprintf("failed to create temporary directory: %v", err)
				if jsonOutput {
					printJSONResult(result)
					return nil
				}
				return fmt.Errorf("failed to create temporary directory: %w", err)
			}

			// Update result with output path
			result.OutputPath = outputPath

			// Create template variables
			vars := &template.Variables{
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
					CreatedAt: time.Now(),
					CreatedBy: "viaplay-cli",
					Year:      time.Now().Year(),
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
				Lambda: template.LambdaInfo{
					Timeout:      30,
					MemorySize:   512,
					Architecture: "arm64",
					Runtime:      "nodejs22.x",
				},
				Cloud: template.CloudInfo{
					Provider: "aws",
				},
				Docker: template.DockerInfo{
					Registry:  "ghcr.io",
					ImageName: projectName,
					ImageTag:  "latest",
				},
				Env: template.EnvInfo{
					Default:      "dev",
					Environments: []string{"dev", "staging", "production"},
				},
			}
			// Get absolute path for output
			absOutputPath, err := filepath.Abs(outputPath)
			if err != nil {
				result.Error = fmt.Sprintf("failed to get absolute path for output: %v", err)
				if jsonOutput {
					printJSONResult(result)
					return nil
				}
				return fmt.Errorf("failed to get absolute path for output: %w", err)
			}

			// Print progress info based on mode
			if !jsonOutput {
				fmt.Fprintf(os.Stderr, "🧪 Testing template scaffolding...\n")
				fmt.Fprintf(os.Stderr, "📁 Output directory: %s\n", outputPath)
			}

			// templateSource is expanded/normalised by cache.ParseSource (e.g. bare
			// paths default to "local@", GitHub URLs are recognised automatically).

			// Run the scaffolding. Manifest options (if any) are resolved interactively,
			// unless overridden via --set or suppressed via --no-input.
			err = scaffolder.ScaffoldProjectWithOptions(ctx, absOutputPath, "", "", templateSource, vars, true, forceRefresh,
				templateSet, noInput)
			if err != nil {
				result.Error = fmt.Sprintf("failed to scaffold template: %v", err)
				if jsonOutput {
					printJSONResult(result)
					return nil
				}
				return fmt.Errorf("failed to scaffold template: %w", err)
			}

			// Success!
			result.Success = true

			// Output in the appropriate format
			if jsonOutput {
				// JSON output always goes to stdout for piping to other tools
				printJSONResult(result)
			} else {
				// Normal mode: print success message to stderr, path to stdout
				fmt.Fprintf(os.Stderr, "✅ Template successfully scaffolded to: %s\n", absOutputPath)
			}

			return nil
		},
	}

	// Add flags
	testCmd.Flags().StringVar(&templatePath, "template-path", "",
		"Template source to test (deprecated; pass <source> as a positional argument instead)")
	testCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of local template copies")
	testCmd.Flags().StringVar(&projectName, "name", "test-project", "Project name for template variables")
	testCmd.Flags().StringVar(&projectOwner, "owner", "test-owner", "Project owner for template variables")
	testCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format for scripting")
	testCmd.Flags().StringArrayVar(&templateSet, "set", nil, "Set a template option using key=value")
	testCmd.Flags().BoolVar(&noInput, "no-input", false, "Do not prompt for template options; use defaults/--set values")
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
// scaffolding anything. Templates without a template.yaml are reported as having
// no configurable options, preserving backward compatibility.
func newTemplateInspectCommand() *cobra.Command {
	var (
		forceRefresh bool
		jsonOutput   bool
	)

	inspectCmd := &cobra.Command{
		Use:   "inspect <source>",
		Short: "Inspect a template and list the manifest-defined options it supports",
		Long: `Inspect a template's manifest (template.yaml), if present, and print the
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

// printTemplateManifestHuman prints a human-readable summary of the manifest's
// options and variables, styled consistently with 'template list' (section
// headers, tables, and shared colour helpers from internal/output).
func printTemplateManifestHuman(manifest *template.Manifest) {
	if manifest == nil {
		output.InfoMessage("This template has no manifest (template.yaml) — no configurable options.")
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
	if !manifest.IsSupported() {
		output.WarningMessage(fmt.Sprintf("manifest schema %d is not supported by this version of vip", manifest.Schema))
	}

	if len(manifest.Options) == 0 && len(manifest.Variables) == 0 {
		output.InfoMessage("This template defines a manifest but no options or variables.")
		return
	}

	headers := []string{"Key", colType, "Required", "Default", "Description"}
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
