// Package cli provides the command-line interface for viaplay-cli.
package cli

import (
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
		outputPath   string
		forceRefresh bool
		projectName  string
		projectOwner string
	)

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test template scaffolding without project creation",
		Long: `Test templates directly without creating repos or authenticating.
This command is useful for template developers who want to test their 
templates during development or in CI pipelines.

By default, scaffolded templates are output to a temporary directory.
You can override this with the --output flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use a minimal configuration that's independent of user config
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialise cache manager
			cacheManager := cache.NewManager(cfg)
			// Initialise project scaffolder
			scaffolder := scaffolding.NewProjectScaffolder(cacheManager, cfg)

			// If no outputPath is provided, create a temporary directory
			if outputPath == "" {
				// Generate a temp directory name based on current time
				timestamp := time.Now().Format("20060102-150405")
				tempDirPrefix := fmt.Sprintf("vip-template-test-%s-", timestamp)
				tempDir, err := createTempDir(tempDirPrefix)
				if err != nil {
					return fmt.Errorf("failed to create temporary directory: %w", err)
				}
				outputPath = tempDir
			}

			// Create a clean output directory
			if err := os.MkdirAll(outputPath, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

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

			fmt.Printf("🧪 Testing template scaffolding...\n")
			fmt.Printf("📁 Output directory: %s\n", outputPath)

			// Create the output directory if it doesn't exist
			absOutputPath, err := filepath.Abs(outputPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for output: %w", err)
			}

			// Run the scaffolding
			err = scaffolder.ScaffoldProject(absOutputPath, "", "", templatePath, vars, true, forceRefresh)
			if err != nil {
				return fmt.Errorf("failed to scaffold template: %w", err)
			}

			fmt.Printf("✅ Template successfully scaffolded to: %s\n", absOutputPath)
			return nil
		},
	}

	// Add flags
	testCmd.Flags().StringVar(&templatePath, "template-path", "", "Local path to a template directory")
	testCmd.Flags().StringVar(&outputPath, "output", "", "Directory where the scaffolded template will be output (defaults to a temporary directory)")
	testCmd.Flags().BoolVar(&forceRefresh, "force", false, "Force refresh of template cache")
	testCmd.Flags().StringVar(&projectName, "name", "test-project", "Project name for template variables")
	testCmd.Flags().StringVar(&projectOwner, "owner", "test-owner", "Project owner for template variables")

	return testCmd
}
