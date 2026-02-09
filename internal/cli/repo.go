// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage GitHub repositories",
	Long: `Manage GitHub repositories.

This command provides subcommands to create and manipulate repositories:
  - 'repo create': Create a GitHub repository without code scaffolding
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// NewRepoCommand returns the repo command with all its subcommands
func NewRepoCommand() *cobra.Command {
	// Add the create subcommand to repo
	repoCmd.AddCommand(newRepoCreateCommand())
	return repoCmd
}

// newRepoCreateCommand creates the repo create command
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
			// Parse owner/repo format
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

	// Add common flags without the name flag (as it's now a positional arg)
	addCommonFlags(cmd, opts)

	return cmd
}
