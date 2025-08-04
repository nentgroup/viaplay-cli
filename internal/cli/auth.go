package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and manage GitHub credentials for viaplay-cli.",
	Long: `Authenticate with GitHub and manage your viaplay-cli credentials.

Examples:
  vip auth login    # Authenticate with GitHub and store your token
  vip auth logout   # Remove your stored GitHub token from the keyring
  vip auth status   # Show authentication status and account info
  vip auth whoami   # Print the currently authenticated GitHub username
  vip auth token    # Print the stored GitHub token (if any) from the keyring

Subcommands:
- login: Authenticate with GitHub and store your token securely.
- logout: Remove your stored GitHub token from the keyring.
- status: Show authentication status and account info.
- whoami: Print the currently authenticated GitHub username.
- token: Print the stored GitHub token (if any) from the keyring.
`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub and store your token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		output.AuthMessage("Authenticating with GitHub...")
		_, err := gh.Authenticate()
		if err != nil {
			output.ErrorMessage("GitHub authentication failed")
			return fmt.Errorf("GitHub authentication failed: %w", err)
		}
		output.SuccessMessage("GitHub authentication complete!")
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove your stored GitHub token from the keyring.",
	RunE: func(cmd *cobra.Command, args []string) error {
		output.ProcessingMessage("Removing GitHub token from keyring")
		err := gh.DeleteToken()
		if err != nil {
			output.ErrorMessage("Failed to remove token")
			return fmt.Errorf("failed to remove token: %w", err)
		}
		output.SuccessMessage("GitHub token removed from keyring")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status and account info.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := gh.GetToken()
		if err != nil || token == "" {
			output.InfoMessage("Not authenticated. Run 'vip auth login' to authenticate.")
			return nil
		}

		output.Section("Authentication Status")
		fmt.Printf("✓ %s\n", output.Success("Authenticated with GitHub"))

		// Optionally, fetch user info using the token
		user, err := gh.GetAuthenticatedUser(token)
		if err == nil {
			fmt.Printf("● Logged in as: %s\n", output.Bold(user))
		}

		fmt.Println() // Add spacing
		return nil
	},
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print the currently authenticated GitHub username.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := gh.GetToken()
		if err != nil || token == "" {
			output.InfoMessage("Not authenticated. Run 'vip auth login' to authenticate.")
			return nil
		}

		output.ProcessingMessage("Fetching user information from GitHub")
		user, err := gh.GetAuthenticatedUser(token)
		if err != nil {
			output.ErrorMessage("Failed to fetch user info")
			return fmt.Errorf("failed to fetch user info: %w", err)
		}

		fmt.Println(output.Bold(user))
		return nil
	},
}

var authPrintTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print the stored GitHub token from the keyring (if any).",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := gh.GetToken()
		if err != nil || token == "" {
			output.InfoMessage("No token found in keyring")
			return nil
		}

		// Format token for better readability
		if len(token) > 10 {
			prefix := token[:7]
			suffix := token[len(token)-7:]
			fmt.Printf("%s...%s (%d characters)\n", prefix, suffix, len(token))

			// Show full token only if explicitly requested
			showFull, err := cmd.Flags().GetBool("full")
			if err != nil {
				fmt.Println(output.Error(fmt.Sprintf("Failed to get 'full' flag: %v", err)))
				return nil
			}
			if showFull {
				fmt.Println(output.Faint("\nFull token:"))
				fmt.Println(token)
			} else {
				fmt.Println(output.Faint("\nUse --full to display the entire token"))
			}
		} else {
			fmt.Println(token)
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authPrintTokenCmd)

	// Add flags
	authPrintTokenCmd.Flags().Bool("full", false, "Display the full token instead of a masked version")
}
