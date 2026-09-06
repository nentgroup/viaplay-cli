// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   defaultProjectName,
	Short: "Manage projects",
	Long: `Manage projects with GitHub integration.

This command provides subcommands to create and manipulate projects:
  - 'project create': Create a full project with scaffolding and a GitHub repository
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// NewProjectCommand returns the project command with all its subcommands
func NewProjectCommand() *cobra.Command {
	// Add the create subcommand to project
	projectCmd.AddCommand(newProjectCreateCommand())
	return projectCmd
}

// newProjectCreateCommand creates the project create command
func newProjectCreateCommand() *cobra.Command {
	opts := &CreateCommandOptions{}

	cmd := &cobra.Command{
		Use:   "create [owner/]<repo-name>",
		Short: "Create a new project with scaffolding and GitHub repository",
		Long: `Create a new project with code scaffolding and GitHub repository.

This command:
1. Creates a GitHub repository
2. Applies organization settings (environments, rulesets, secrets)
3. Scaffolds a new project from templates
4. Optionally clones the project locally

Use this for a complete project setup experience.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			// Parse owner/repo format
			parseOwnerRepoArg(args[0], opts)
			return createProjectOrRepo(ctx, opts, true)
		},
		Example: `  # Create a Go service with default settings
  vip project create my-service --language go --type service --team platform

  # Create with explicit owner
  vip project create myorg/my-app --language typescript --type service --team frontend

  # Create a local project without a GitHub repository
  vip project create local-app --language go --type cli --no-repo`,
	}

	// Add common flags without the name flag (as it's now a positional arg)
	addCommonFlags(cmd, opts)

	// Add project-specific flags
	cmd.Flags().StringVar(&opts.Language, "language", "", "Programming language (go, typescript, etc.)")
	cmd.Flags().StringVar(&opts.ProjectType, "type", "", "Project type (service, cli, lambda, etc.)")
	cmd.Flags().StringVar(&opts.TemplateSource, "template-source", "", "Custom template source")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "Directory to create the project in (defaults to current dir + repo name)")
	cmd.Flags().StringVar(&opts.BinaryName, "binary-name", "", "Name of the compiled binary (for compiled languages like Go and Rust)")
	cmd.Flags().BoolVar(&opts.NoRepo, "no-repo", false, "Skip GitHub repository creation (local project only)")
	cmd.Flags().BoolVar(&opts.NoHooks, "no-hooks", false, "Skip running post-installation hooks")
	cmd.Flags().BoolVar(&opts.NoCache, "no-cache", false, "Force update of template cache (ignore cached templates)")
	return cmd
}
