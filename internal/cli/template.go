// Package cli provides the command-line interface for viaplay-cli.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		Long:  `Manage and test templates for project scaffolding.`,
	}

	// Add subcommands
	templateCmd.AddCommand(newTemplateTestCommand())

	return templateCmd
}

// newTemplateTestCommand creates a new test subcommand for the template command
func newTemplateTestCommand() *cobra.Command {
	var (
		templatePath string
		forceRefresh bool
		projectName  string
		projectOwner string
		jsonOutput   bool
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

			// Run the scaffolding
			err = scaffolder.ScaffoldProject(ctx, absOutputPath, "", "", templatePath, vars, true, forceRefresh)
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
	testCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of template cache")
	testCmd.Flags().StringVar(&projectName, "name", "test-project", "Project name for template variables")
	testCmd.Flags().StringVar(&projectOwner, "owner", "test-owner", "Project owner for template variables")
	testCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format for scripting")

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
