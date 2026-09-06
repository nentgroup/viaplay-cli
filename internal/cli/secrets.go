package cli

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

const (
	// KeyringService is the service name used for keyring operations
	// Using the same service name as in gh package for consistency
	KeyringService = "viaplaycli"
)

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   scopeSecrets,
	Short: "Manage secrets in the system keyring",
	Long: `Manage secrets in the system keyring for use with viaplay-cli.
These secrets can be referenced in team configuration files and used 
when creating or managing GitHub repositories and environments.

Examples:
  vip secrets set MY_SECRET        # Prompts securely for the value
  vip secrets set MY_SECRET --file secret.txt  # Reads from file
  vip secrets get MY_SECRET
  vip secrets forget MY_SECRET
  vip secrets list
`,
	Run: func(cmd *cobra.Command, args []string) {
		output.Section("Secrets Management")
		fmt.Println("Use one of the subcommands: set, get, forget, or list")
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

// secretsSetCmd represents the secrets set command
var secretsSetCmd = &cobra.Command{
	Use:   "set KEY",
	Short: "Store a secret in the system keyring",
	Long: `Store a secret in the system keyring associated with viaplay-cli.
The secret will be stored securely and can be retrieved using the get command.

For security, there are two ways to provide the secret value:
1. Interactive prompt (default): You'll be prompted for the value securely (no echo to terminal)
2. File input: Use --file to read the secret from a file

Examples:
  vip secrets set GITHUB_TOKEN         # Will prompt securely
  vip secrets set API_KEY --file key.txt
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		fileFlag, err := cmd.Flags().GetString("file")
		if err != nil {
			return fmt.Errorf("failed to get 'file' flag: %w", err)
		}

		// Get secret value from file or prompt
		value, err := getSecretValue(key, fileFlag)
		if err != nil {
			return err
		}

		// Check if value is empty and confirm if needed
		if value == "" && !confirmEmptySecret() {
			return nil
		}

		// Store the secret in keyring
		return storeSecret(key, value)
	},
}

// secretsGetCmd represents the secrets get command
var secretsGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Retrieve a secret from the system keyring",
	Long: `Retrieve a secret from the system keyring that was previously stored.

Example:
  vip secrets get GITHUB_TOKEN
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		showFlag, err := cmd.Flags().GetBool("show")
		if err != nil {
			return fmt.Errorf("failed to get 'show' flag: %w", err)
		}

		// Retrieve and display the secret
		return retrieveAndDisplaySecret(key, showFlag)
	},
}

// secretsForgetCmd represents the secrets forget command
var secretsForgetCmd = &cobra.Command{
	Use:   "forget KEY",
	Short: "Remove a secret from the system keyring",
	Long: `Remove a secret from the system keyring.

Example:
  vip secrets forget GITHUB_TOKEN
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		return removeSecret(key)
	},
}

// secretsListCmd represents the secrets list command
var secretsListCmd = &cobra.Command{
	Use:   cmdList,
	Short: "List all secrets stored in the system keyring",
	Long: `List all secrets stored in the system keyring for viaplay-cli.
This command only shows the names of the secrets, not their values.

Example:
  vip secrets list
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output.InfoMessage("This feature is not yet implemented")
		fmt.Println("In the meantime, you can use your system's keyring utility to view all secrets.")
		fmt.Println("The service name used is:", output.Bold(KeyringService))
		return nil
	},
}

// getSecretValue retrieves a secret value from a file or prompts the user
func getSecretValue(key, filePath string) (string, error) {
	if filePath != "" {
		// Read from file
		output.ProcessingMessage(fmt.Sprintf("Reading secret '%s' from file '%s'", key, filePath))
		data, err := os.ReadFile(filePath)
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to read from file: %v", err))
			return "", fmt.Errorf("failed to read from file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// Interactive prompt (most secure)
	fmt.Printf("Enter value for secret '%s' (input will not be displayed): ", key)
	valueBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Add newline after password input
	if err != nil {
		output.ErrorMessage("Failed to read secret value")
		return "", fmt.Errorf("failed to read secret value: %w", err)
	}
	return strings.TrimSpace(string(valueBytes)), nil
}

// confirmEmptySecret asks for confirmation when storing an empty secret
func confirmEmptySecret() bool {
	output.WarningMessage("Secret value is empty")
	fmt.Print("Continue storing empty secret? (y/N): ")
	var confirm string
	n, err := fmt.Scanln(&confirm)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to read input: %v", err))
		return false
	}
	_ = n // ignore the number of items scanned
	if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
		output.InfoMessage("Operation canceled")
		return false
	}
	return true
}

// storeSecret stores a secret in the system keyring
func storeSecret(key, value string) error {
	output.ProcessingMessage(fmt.Sprintf("Storing secret '%s' in system keyring", key))
	if err := keyring.Set(KeyringService, key, value); err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to store secret: %v", err))
		return fmt.Errorf("failed to store secret: %w", err)
	}

	output.SuccessMessage(fmt.Sprintf("Secret '%s' stored successfully", key))
	return nil
}

// retrieveAndDisplaySecret gets a secret from the keyring and displays it if requested
func retrieveAndDisplaySecret(key string, showValue bool) error {
	output.ProcessingMessage(fmt.Sprintf("Retrieving secret '%s' from system keyring", key))
	value, err := keyring.Get(KeyringService, key)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to retrieve secret: %v", err))
		return fmt.Errorf("failed to retrieve secret: %w", err)
	}

	if showValue {
		fmt.Printf("\n%s\n", output.Bold(value))
	} else {
		output.SuccessMessage(fmt.Sprintf("Secret '%s' retrieved successfully", key))
		fmt.Printf("Use %s to display the value\n", output.Bold("--show"))
	}
	return nil
}

// removeSecret deletes a secret from the system keyring
func removeSecret(key string) error {
	output.ProcessingMessage(fmt.Sprintf("Removing secret '%s' from system keyring", key))
	if err := keyring.Delete(KeyringService, key); err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to remove secret: %v", err))
		return fmt.Errorf("failed to remove secret: %w", err)
	}

	output.SuccessMessage(fmt.Sprintf("Secret '%s' removed successfully", key))
	return nil
}

func init() {
	// Add subcommands
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsGetCmd)
	secretsCmd.AddCommand(secretsForgetCmd)
	secretsCmd.AddCommand(secretsListCmd)

	// Add flags
	secretsSetCmd.Flags().String("file", "", "Read secret value from a file")
	secretsGetCmd.Flags().Bool("show", false, "Display the secret value")
}
