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
	templateCmd.AddCommand(newTemplateInfoCommand())
	templateCmd.AddCommand(newTemplateUpdateCommand())
	templateCmd.AddCommand(newTemplatePruneCommand())
	templateCmd.AddCommand(newTemplateCleanCommand())
	templateCmd.AddCommand(newTemplateTestCommand())
	templateCmd.AddCommand(newTemplateOptionsCommand())

	return templateCmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   cmdList,
		Short: "List local template copies",
		Long:  `List template copies currently stored locally for reuse.`,
		Run: func(cmd *cobra.Command, args []string) {
			listCache(cmd.Context())
		},
	}
}

func newTemplateInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show local template storage information",
		Long:  `Show detailed information about locally stored template copies including size and statistics.`,
		Run: func(cmd *cobra.Command, args []string) {
			showCacheInfo(cmd.Context())
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
		Use:   "test",
		Short: "Test template scaffolding without project creation",
		Long: `Test templates directly without creating repos or authenticating.
This command is useful for template developers who want to test their 
templates during development or in CI pipelines.

Templates are output to a temporary directory that is automatically created.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Initialise result object for potential JSON output
			result := TestResult{
				TemplatePath: templatePath,
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

			templateSource := templatePath
			if templateSource != "" && !strings.Contains(templateSource, "@") {
				templateSource = "local@" + templateSource
			}

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
	testCmd.Flags().StringVar(&templatePath, "template-path", "", "Local path to a template directory")
	testCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of local template copies")
	testCmd.Flags().StringVar(&projectName, "name", "test-project", "Project name for template variables")
	testCmd.Flags().StringVar(&projectOwner, "owner", "test-owner", "Project owner for template variables")
	testCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format for scripting")
	testCmd.Flags().StringArrayVar(&templateSet, "set", nil, "Set a template option using key=value")
	testCmd.Flags().BoolVar(&noInput, "no-input", false, "Do not prompt for template options; use defaults/--set values")

	return testCmd
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

// newTemplateOptionsCommand creates a new "options" subcommand that inspects a
// template and lists the manifest-driven options (if any) it supports, without
// scaffolding anything. Templates without a template.yaml are reported as having
// no configurable options, preserving backward compatibility.
func newTemplateOptionsCommand() *cobra.Command {
	var (
		templatePath   string
		templateSource string
		forceRefresh   bool
		jsonOutput     bool
	)

	optionsCmd := &cobra.Command{
		Use:   "options",
		Short: "List the manifest-defined options a template supports",
		Long: `Inspect a template's manifest (template.yaml), if present, and print the
options and variables it exposes (key, type, prompt, default, choices).

Use --set key=value with 'project create' or 'template test' to set these
options non-interactively. Templates without a manifest report no options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			source := templateSource
			if source == "" {
				source = templatePath
			}
			if source == "" {
				return fmt.Errorf("either --template-path or --template-source must be provided")
			}
			if !strings.Contains(source, "@") {
				source = "local@" + source
			}

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
				printTemplateOptionsJSON(manifest)
				return nil
			}
			printTemplateOptionsHuman(manifest)
			return nil
		},
	}

	optionsCmd.Flags().StringVar(&templatePath, "template-path", "", "Local path to a template directory")
	optionsCmd.Flags().StringVar(&templateSource, "template-source", "", "Template source (e.g. local@/path, github@owner/repo)")
	optionsCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of local template copies")
	optionsCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format for scripting")

	return optionsCmd
}

// printTemplateOptionsJSON prints the manifest options/variables as JSON, or an
// empty/null manifest when the template has none.
func printTemplateOptionsJSON(manifest *template.Manifest) {
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// printTemplateOptionsHuman prints a human-readable summary of the manifest's
// options and variables.
func printTemplateOptionsHuman(manifest *template.Manifest) {
	if manifest == nil {
		fmt.Println("This template has no manifest (template.yaml) — no configurable options.")
		return
	}

	if manifest.Metadata.Name != "" || manifest.Metadata.Description != "" {
		fmt.Printf("📦 %s\n", strings.TrimSpace(manifest.Metadata.Name+" "+manifest.Metadata.Description))
	}
	if !manifest.IsSupported() {
		fmt.Printf("⚠ Warning: manifest schema %d is not supported by this version of vip\n", manifest.Schema)
	}

	if len(manifest.Options) == 0 && len(manifest.Variables) == 0 {
		fmt.Println("This template defines a manifest but no options or variables.")
		return
	}

	if len(manifest.Options) > 0 {
		fmt.Println("Options (use --set key=value):")
		for _, opt := range manifest.Options {
			printManifestOptionLine(opt.Key, opt.Type, opt.Description, opt.Default, opt.Required, opt.Choices)
		}
	}

	if len(manifest.Variables) > 0 {
		fmt.Println("Variables (use --set key=value):")
		for _, v := range manifest.Variables {
			printManifestOptionLine(v.Key, v.Type, v.Description, v.Default, v.Required, nil)
		}
	}
}

func printManifestOptionLine(key, typ, description string, def any, required bool, choices []template.ManifestChoice) {
	line := fmt.Sprintf("  - %s (%s)", key, typ)
	if required {
		line += " [required]"
	}
	if def != nil {
		line += fmt.Sprintf(" default=%v", def)
	}
	fmt.Println(line)
	if description != "" {
		fmt.Printf("      %s\n", description)
	}
	for _, c := range choices {
		label := c.Label
		if label == "" {
			label = c.Value
		}
		fmt.Printf("      • %s: %s\n", c.Value, label)
	}
}
