// Package secrets provides functionality for managing and accessing secrets
// from various sources including keyring, environment variables, and configuration files.
//
// It handles secure storage and retrieval of sensitive values like API keys, tokens,
// and other credentials, with support for multiple secret reference formats:
//   - GitHub Actions style references: ${{ secrets.SECRET_NAME }}
//   - Environment variable references: $ENV_VAR
//   - Direct values from configuration
//   - References to other secrets
//
// The package integrates with the system keyring for secure storage and
// sanitises secret names according to GitHub's naming requirements.
//
// Key functions:
//   - ResolveSecretValue: Resolves a secret from various sources (keyring, env, config)
//   - GetSecretValueAndSource: Determines a secret's value and source
//   - GetSecret: Retrieves a secret directly from the keyring
//   - SanitizeSecretName: Ensures a secret name meets GitHub's requirements
//
// Example usage:
//
//	// Resolve a secret value that might be in different formats
//	value, sourceType, sourceKey, err := secrets.ResolveSecretValue("${{ secrets.API_KEY }}", "API_KEY")
//	if err != nil {
//	    log.Fatalf("Failed to resolve secret: %v", err)
//	}
//	fmt.Printf("Got value from %s: %s\n", sourceType, sourceKey)
//
//	// Work with a Secret struct and pre-resolved values
//	secret := secrets.Secret{Name: "TOKEN", Value: "${{ secrets.GITHUB_TOKEN }}"}
//	secretValues := map[string]string{"TOKEN": "resolved-token-value"}
//	value, source, ok := secrets.GetSecretValueAndSource(secret, secretValues)
//	if ok {
//	    fmt.Printf("Secret source: %s\n", source)
//	}
package secrets

import (
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
)

// Secret represents a secret or variable definition for use in team/repo configs
// This matches the structure used in secretsConfig.Secrets
// (duplicated here to avoid import cycles and for helper use)
type Secret struct {
	Name      string
	Value     string
	Env       string
	Type      string
	Reference string
}

// Helper functions

// Source type constants
const (
	// SourceTypeConfig represents a configuration source type
	SourceTypeConfig = "config"
	// SourceTypeKeyring represents a keyring source type
	SourceTypeKeyring = "keyring"
	// SourceTypeEnv represents an environment variable source type
	SourceTypeEnv = "env"

	KeyringServiceName = "viaplaycli" // Keyring service name used for storing secrets
)

// extractGitHubActionsSecret extracts a secret name from GitHub Actions style syntax: ${{ secrets.SECRET_NAME }}
// Returns the secret name or empty string if no match
func extractGitHubActionsSecret(value string) string {
	// Simple regex-like pattern matching: ${{ secrets.KEY_NAME }}
	value = strings.TrimSpace(value)

	// Check if it follows the pattern
	if !strings.HasPrefix(value, "${{") || !strings.HasSuffix(value, "}}") {
		return ""
	}

	// Extract the part between ${{ and }}
	inner := strings.TrimSpace(value[3 : len(value)-2])

	// Check if it starts with secrets.
	if !strings.HasPrefix(inner, "secrets.") {
		return ""
	}

	// Extract the key name (everything after secrets.)
	keyName := strings.TrimSpace(inner[8:])
	if keyName == "" {
		return ""
	}

	return keyName
}

// SanitizeSecretName ensures a secret name follows GitHub's naming requirements:
// - Can only contain alphanumeric characters or underscores
// - Must start with a letter or underscore
// - No spaces allowed
func SanitizeSecretName(name string) string {
	// Replace hyphens with underscores
	sanitized := strings.ReplaceAll(name, "-", "_")

	// Ensure the name starts with a letter or underscore
	if len(sanitized) > 0 && (!isAlpha(sanitized[0]) && sanitized[0] != '_') {
		sanitized = "_" + sanitized
	}

	// Replace any other invalid characters with underscores
	for i, char := range sanitized {
		if !isAlphaNumeric(char) && char != '_' {
			sanitized = sanitized[:i] + "_" + sanitized[i+1:]
		}
	}

	return sanitized
}

// isAlpha checks if a byte is an alphabetic character (a-z, A-Z)
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isAlphaNumeric checks if a rune is an alphanumeric character (a-z, A-Z, 0-9)
func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// ResolveSecretValue resolves a secret value from various sources (config, keyring, env vars)
// It handles different reference formats and directly stored values.
//
// Parameters:
//   - value: The raw value or reference string from the configuration
//   - name: The name of the secret (used for error messages)
//
// Returns:
//   - string: The resolved secret value
//   - string: The source type ("keyring", "env", or "config")
//   - string: The source key (keyring name, env var name, or empty for direct values)
//   - error: Any error that occurred during resolution
//
// Resolution order:
//  1. GitHub Actions style references: ${{ secrets.KEY_NAME }} → keyring/vault lookup
//  2. Environment variable references: $ENV_VAR → environment variable lookup
//  3. Direct values from config (as-is)
func ResolveSecretValue(value, name string) (string, string, string, error) {
	// Exit early if no value is provided
	if value == "" {
		return "", "", "", fmt.Errorf("no value source provided for '%s'", name)
	}

	// Case 1a: GitHub Actions style reference - ${{ secrets.KEY_NAME }}
	if keyringKey := extractGitHubActionsSecret(value); keyringKey != "" {
		// Get from keyring/vault
		keyringValue, err := GetSecret(keyringKey)
		if err != nil {
			return "", "keyring", keyringKey, fmt.Errorf("failed to get keyring value for '%s': %w", keyringKey, err)
		}
		return keyringValue, "keyring", keyringKey, nil
	}

	// Case 1b: Environment variable reference - $ENV_VAR
	if strings.HasPrefix(value, "$") && len(value) > 1 {
		envVarName := value[1:] // Remove the $ prefix
		envVarValue := os.Getenv(envVarName)
		if envVarValue == "" {
			// We return an empty string but with a warning - this isn't a fatal error
			// as empty environment variables are valid in some cases
			fmt.Printf("Warning: Environment variable '%s' is empty or not set\n", envVarName)
		}
		return envVarValue, "env", envVarName, nil
	}

	// Case 1c: Direct value from config
	return value, "config", "", nil
}

// GetSecret retrieves a secret from the keyring
func GetSecret(key string) (string, error) {
	// Use the keyring service to get the secret
	return keyring.Get(KeyringServiceName, key)
}

// GetSecretValueAndSource determines a secret's value and its source based on the provided Secret and resolved secret values
// Parameters:
//   - s: Secret struct containing name, value, env, type, and reference information
//   - secretValues: Map of already resolved secret values, indexed by secret name
//
// Returns:
//   - string: The resolved secret value
//   - string: The source of the secret (format: "keyring:KEY", "env:ENV_VAR", "config", "reference:REF_NAME")
//   - bool: Whether the secret was successfully resolved
func GetSecretValueAndSource(s Secret, secretValues map[string]string) (string, string, bool) {
	// Case 1: Secret has a direct value
	if s.Value != "" {
		// Check if value is a GitHub Actions style reference
		keyringKey := extractGitHubActionsSecret(s.Value)
		if keyringKey != "" {
			return secretValues[s.Name], "keyring:" + keyringKey, true
		}

		// Check if value is an environment variable reference
		if strings.HasPrefix(s.Value, "$") && len(s.Value) > 1 {
			envVarName := s.Value[1:] // Remove the $ prefix
			return secretValues[s.Name], "env:" + envVarName, true
		}

		// Direct value from configuration
		return s.Value, "config", true
	}

	// Case 2: Secret references another secret
	if s.Reference != "" {
		refValue, exists := secretValues[s.Reference]
		if !exists {
			return "", "reference missing", false
		}
		return refValue, "reference:" + s.Reference, true
	}

	// Case 3: No value or reference provided
	return "", "", false
}
