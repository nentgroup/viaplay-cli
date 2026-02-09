package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/invopop/yaml"

	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/secrets"
)

// applyConfigurations applies configurations to the GitHub repository
func (c *Factory) applyConfigurations(ctx context.Context, opts Options) error {
	// Skip all GitHub configurations if NoRepo is true
	if opts.SkipRepo {
		return nil
	}

	// Get authenticated username for personal directory path
	username, err := c.GitHubClient.GetAuthenticatedUser(ctx)
	if err != nil {
		c.Reporter.Warning("Auth", fmt.Sprintf("Failed to get authenticated username: %v", err))
		username = "" // Default to empty if we can't get the username
	}

	// Determine the appropriate configuration directory based on account type
	var configDir string

	// For organisation repos with team specified, use org team directory
	if opts.AccountType == "organization" && opts.Team != "" {
		// For organisation repositories, use the organisation-specific team directory
		configDir = c.Config.GetTeamDir(opts.Team, opts.RepoOwner)
		c.Reporter.Debug(fmt.Sprintf("Using organization-specific team directory: %s", configDir))
	} else if opts.AccountType == "user" && username != "" {
		// For personal accounts, use the personal directory
		configDir = c.Config.GetPersonalDir(username)
		c.Reporter.Debug(fmt.Sprintf("Using personal directory: %s", configDir))

		// Ensure the user directory exists
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return fmt.Errorf("failed to create personal config directory: %w", err)
			}
		}

		// Check if config files exist
		if _, err := os.Stat(filepath.Join(configDir, "envs")); os.IsNotExist(err) {
			c.Reporter.Debug(fmt.Sprintf("Personal environment configs for user %s not found, skipping", username))
		}
	} else if opts.Team != "" {
		// Fallback: For personal repositories with a team specified, use the global team directory (legacy support)
		configDir = filepath.Join(opts.ConfigDir, "teams", opts.Team)
		c.Reporter.Debug(fmt.Sprintf("Using global team directory: %s", configDir))
	} else {
		c.Reporter.Debug("No team or personal account specified, skipping configurations")
		return nil // No team or personal account specified, nothing to apply
	}

	// 1. Apply environments if requested
	if opts.ApplyEnvs {
		err := c.applyEnvs(ctx, opts.RepoOwner, opts.RepoName, configDir)
		if err != nil {
			return fmt.Errorf("failed to apply environments: %w", err)
		}
	}

	// 2. Apply rulesets if requested
	if opts.ApplyRulesets {
		if err := c.applyRulesets(ctx, opts.RepoOwner, opts.RepoName, configDir); err != nil {
			return fmt.Errorf("failed to apply rulesets: %w", err)
		}
	}

	// 3. Apply secrets if requested
	if opts.ApplySecrets {
		if err := c.applySecrets(ctx, opts.RepoOwner, opts.RepoName, configDir); err != nil {
			return fmt.Errorf("failed to apply secrets: %w", err)
		}
	}

	// 4. Apply repository-specific secrets if provided
	if opts.RepoSecrets != "" {
		if err := c.applyRepoSecrets(ctx, opts.RepoOwner, opts.RepoName, opts.RepoSecrets); err != nil {
			return fmt.Errorf("failed to apply repository-specific secrets: %w", err)
		}
	}

	return nil
}

// applyEnvs applies environments defined in the team directory
func (c *Factory) applyEnvs(ctx context.Context, owner, repo, teamDir string) error {
	mainOperation := "Applying environments"

	// Start the overall operation
	c.Reporter.Start(mainOperation, "")

	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Reporter.Failed(mainOperation, err, "Failed to get user home directory")
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	envsDir := filepath.Join(teamDir, "envs")
	c.Reporter.Debug(fmt.Sprintf("Looking for environment configs in %s", envsDir))

	// Check if the directory exists first
	if _, err := os.Stat(envsDir); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Environments directory does not exist: %s", envsDir)
		c.Reporter.Skip(mainOperation, errMsg)
		return fmt.Errorf("environments directory does not exist: %s", envsDir)
	}

	entries, err := os.ReadDir(envsDir)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, fmt.Sprintf("Failed to read envs directory: %s", envsDir))
		return fmt.Errorf("failed to read environments directory: %w", err)
	}

	if len(entries) == 0 {
		c.Reporter.Skip(mainOperation, "No environment configurations found")
		return nil
	}

	appliedCount := 0
	failedEnvs := []string{}

	// Create a renderer with the template variables
	renderer := c.getTemplateRenderer()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process yaml files
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(envsDir, entry.Name())
		c.Reporter.Debug(fmt.Sprintf("Processing environment file: %s", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read env file %s: %v", entry.Name(), err)
			c.Reporter.Warning("Environment processing", errMsg)
			failedEnvs = append(failedEnvs, errMsg)
			continue
		}

		// First, render the template variables in the raw content - this is critical for YAML processing
		c.Reporter.Debug(fmt.Sprintf("Rendering ruleset %s with template variables", entry.Name()))
		renderedData, err := renderer.RenderString(string(data))
		if err != nil {
			errMsg := fmt.Sprintf("Failed to render ruleset %s with template variables: %v", entry.Name(), err)
			c.Reporter.Warning("Ruleset rendering", errMsg)
			failedEnvs = append(failedEnvs, errMsg)
			continue
		}

		// Parse the environment configuration
		var envConfig EnvConf

		// Parse JSON
		if err := yaml.Unmarshal([]byte(renderedData), &envConfig); err != nil {
			errMsg := fmt.Sprintf("Failed to parse JSON in %s: %v", entry.Name(), err)
			c.Reporter.Warning("Environment processing", errMsg)
			failedEnvs = append(failedEnvs, errMsg)
			continue
		}

		if envConfig.Name == "" {
			c.Reporter.Skip("Environment processing", fmt.Sprintf("Environment in %s is missing a name", entry.Name()))
			continue
		}

		// Update the main operation with current environment being processed
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Creating environment: %s", envConfig.Name))

		// Create the basic environment first
		if err := c.GitHubClient.CreateEnvironment(ctx, owner, repo, envConfig.Name,
			envConfig.ToGitHubEnv()); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				errMsg := fmt.Sprintf("Failed to create environment %s: %v", envConfig.Name, err)
				c.Reporter.Warning("Environment creation", errMsg)
				failedEnvs = append(failedEnvs, errMsg)
				continue
			}
		}

		// Then apply deployment branch policy separately if present
		if envConfig.DeploymentBranchPolicy != nil {
			c.Reporter.Progress(mainOperation, 50, fmt.Sprintf("Applying branch patterns for %s", envConfig.Name))

			// In v74, we can't directly set protected vs custom branch policies
			// Instead, we'll focus on adding the branch patterns if they're provided

			// If we have branch patterns defined, add them directly to the environment
			if len(envConfig.DeploymentBranchPolicy.BranchPatterns) > 0 {
				for _, pattern := range envConfig.DeploymentBranchPolicy.BranchPatterns {
					// Extract the actual pattern string from the DeploymentBranchPolicyRequest object
					patternStr := pattern.GetName()

					c.Reporter.Progress(mainOperation, 75, fmt.Sprintf("Adding branch pattern '%s' to %s", patternStr, envConfig.Name))

					// Apply the branch pattern
					if err := c.GitHubClient.CreateCustomBranchPolicy(ctx, owner, repo, envConfig.Name,
						pattern); err != nil {
						errMsg := fmt.Sprintf("Failed to add branch pattern '%s' for %s: %v", patternStr, envConfig.Name, err)
						c.Reporter.Warning("Branch pattern", errMsg)
						// Don't fail the entire operation because of one pattern
					} else {
						c.Reporter.Debug(fmt.Sprintf("Added branch pattern '%s' to %s", patternStr, envConfig.Name))
					}
				}
			} else {
				c.Reporter.Debug(fmt.Sprintf("No branch patterns specified for %s", envConfig.Name))
			}
		}

		c.Reporter.Debug(fmt.Sprintf("Successfully created environment: %s", envConfig.Name))
		appliedCount++
	}

	// Return a summary error if any environments failed
	if len(failedEnvs) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d environments, %d failed", appliedCount, len(failedEnvs))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some environments could not be applied: %s", strings.Join(failedEnvs[:1], ", "))
	}

	// Finalise the overall operation
	if appliedCount > 0 {
		c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d environments", appliedCount))
	} else {
		c.Reporter.Skip(mainOperation, "No new environments were applied")
	}

	return nil
}

// applyRulesets applies rulesets defined in the team directory
func (c *Factory) applyRulesets(ctx context.Context, owner, repo, teamDir string) error {
	mainOperation := "Applying rulesets"
	// Start the overall operation
	c.Reporter.Start(mainOperation, "")

	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Reporter.Failed(mainOperation, err, "Failed to get user home directory")
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	rulesetsDir := filepath.Join(teamDir, "rulesets")
	c.Reporter.Debug(fmt.Sprintf("Looking for ruleset configs in %s", rulesetsDir))

	// Check if the directory exists first
	if _, err := os.Stat(rulesetsDir); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Rulesets directory does not exist: %s", rulesetsDir)
		c.Reporter.Skip(mainOperation, errMsg)
		return fmt.Errorf("rulesets directory does not exist: %s", rulesetsDir)
	}

	files, err := os.ReadDir(rulesetsDir)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to read rulesets directory")
		return fmt.Errorf("failed to read rulesets directory: %w", err)
	}

	if len(files) == 0 {
		c.Reporter.Skip(mainOperation, "No ruleset files found")
		return nil
	}

	// Create a renderer with the template variables
	renderer := c.getTemplateRenderer()

	// Track applied and failed rulesets
	appliedCount := 0
	failedRulesets := []string{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// Process both JSON and YAML ruleset files
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(rulesetsDir, f.Name())
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Processing ruleset file: %s", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to read ruleset file %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset processing", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
			continue
		}

		// First, render the template variables in the raw content - this is critical for YAML processing
		c.Reporter.Debug(fmt.Sprintf("Rendering ruleset %s with template variables", f.Name()))
		renderedData, err := renderer.RenderString(string(data))
		if err != nil {
			errMsg := fmt.Sprintf("Failed to render ruleset %s with template variables: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset rendering", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
			continue
		}

		// Debug the rendered data to help troubleshoot
		c.Reporter.Debug(fmt.Sprintf("Rendered ruleset data for %s: \n%s", f.Name(), renderedData))

		// Unmarshal JSON/YAML into the GitHub API struct
		var ruleset github.RepositoryRuleset

		// Unmarshal JSON into the GitHub struct
		if err := yaml.Unmarshal([]byte(renderedData), &ruleset); err != nil {
			errMsg := fmt.Sprintf("Failed to parse converted JSON from YAML in %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset processing", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
			continue
		}

		c.Reporter.Debug(fmt.Sprintf("Successfully converted YAML to GitHub ruleset structure for %s", f.Name()))

		// Basic validation
		if ruleset.Name == "" {
			c.Reporter.Skip("Ruleset processing", fmt.Sprintf("Ruleset in %s is missing a name", f.Name()))
			continue
		}

		if ruleset.Target == nil {
			c.Reporter.Skip("Ruleset processing", fmt.Sprintf("Ruleset in %s is missing a target", f.Name()))
			continue
		}

		// Debug output
		c.Reporter.Debug(fmt.Sprintf("Applying ruleset: %s (target: %s)", ruleset.Name, *ruleset.Target))

		// Apply the ruleset
		if err := c.GitHubClient.CreateRuleset(ctx, owner, repo, ruleset); err != nil {
			errMsg := fmt.Sprintf("Failed to apply ruleset %s: %v", f.Name(), err)
			c.Reporter.Warning("Ruleset application", errMsg)
			failedRulesets = append(failedRulesets, errMsg)
		} else {
			c.Reporter.Progress("Ruleset application", 100, fmt.Sprintf("Applied ruleset: %s", ruleset.Name))
			appliedCount++
		}
	}

	// Return a summary error if any rulesets failed
	if len(failedRulesets) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d rulesets, %d failed", appliedCount, len(failedRulesets))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some rulesets could not be applied: %s", strings.Join(failedRulesets[:1], ", "))
	}

	// Finalise the overall operation
	if appliedCount > 0 {
		c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d environments", appliedCount))
	} else {
		c.Reporter.Skip(mainOperation, "No new environments were applied")
	}
	return nil
}

// applySecrets applies secrets defined in the team directory
func (c *Factory) applySecrets(ctx context.Context, owner, repo, teamDir string) error {
	mainOperation := "Applying secrets"
	c.Reporter.Start(mainOperation, "")

	// Expand tilde in path if it exists
	if strings.HasPrefix(teamDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			c.Reporter.Failed(mainOperation, err, "Failed to get user home directory")
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		teamDir = filepath.Join(home, teamDir[1:])
	}

	secretsPath := filepath.Join(teamDir, "secrets.yaml")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		c.Reporter.Skip(mainOperation, "No secrets.json file found")
		return nil
	}

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to read secrets file")
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	// Create a renderer with the template variables
	renderer := c.getTemplateRenderer()
	c.Reporter.Debug("Rendering secrets.json with template variables")

	// Render the secrets.json content with template variables
	renderedData, err := renderer.RenderString(string(data))
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to render secrets.yaml with template variables")
		return fmt.Errorf("failed to render secrets.json with template variables: %w", err)
	}

	var secretsConfig secrets.Config

	// Parse the rendered JSON
	if err := yaml.Unmarshal([]byte(renderedData), &secretsConfig); err != nil {
		c.Reporter.Failed(mainOperation, err, fmt.Sprintf("Failed to parse YAML in %s", secretsPath))
		return fmt.Errorf("failed to parse secrets JSON: %w", err)
	}

	// First pass to collect all secret values
	secretValues := make(map[string]string)
	for _, s := range secretsConfig.Secrets {
		// Skip if name is empty
		if s.Name == "" {
			c.Reporter.Skip("Secret processing", "Skipping secret with missing name")
			continue
		}

		secretValue, sourceType, sourceKey, err := secrets.ResolveSecretValue(s.Value, s.Name)
		if err != nil {
			c.Reporter.Warning("Secret resolution", fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err))
			continue
		}

		// Store the secret value for potential references
		secretValues[s.Name] = secretValue

		// Debug info about resolution
		if sourceType != "config" {
			c.Reporter.Debug(fmt.Sprintf("Resolved '%s' from %s: '%s'", s.Name, sourceType, sourceKey))
		}
	}

	// Second pass to apply secrets, including those with references
	appliedCount := 0
	for _, secret := range secretsConfig.Secrets {
		if secret.Name == "" {
			continue // Skip again
		}

		secretValue, valueSource, ok := secrets.GetSecretValueAndSource(secret, secretValues)
		if !ok {
			if secret.Reference != "" {
				c.Reporter.Warning("Secret reference", fmt.Sprintf("Referenced secret '%s' not found for '%s'", secret.Reference, secret.Name))
			}
			continue
		}

		if secretValue == "" {
			c.Reporter.Skip("Secret processing", fmt.Sprintf("Skipping secret '%s' with empty value", secret.Name))
			continue
		}

		isVariable := secret.Type == "variable"
		if err := c.applySecretOrVariable(ctx, isVariable, owner, repo, secret.Name, secretValue, secret.Env,
			valueSource); err != nil {
			c.Reporter.Warning("Secret application", fmt.Sprintf("Failed to apply %s '%s': %v",
				secret.Type, secret.Name, err))
		} else {
			c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Applied %s: %s (env: %s, source: %s)",
				valueOrEmpty(secret.Type, "secret"), secret.Name, valueOrEmpty(secret.Env, "repo"), valueSource))
			appliedCount++
		}
	}

	if appliedCount > 0 {
		c.Reporter.Complete(mainOperation, fmt.Sprintf("Applied %d secrets/variables", appliedCount))
	} else {
		c.Reporter.Skip(mainOperation, "No secrets were applied")
	}

	return nil
}

// applyRepoSecrets applies repository-specific secrets from a JSON string
func (c *Factory) applyRepoSecrets(ctx context.Context, owner, repo, secretsJSON string) error {
	mainOperation := "Applying repository-specific secrets"
	c.Reporter.Start(mainOperation, "")

	// Create a renderer with the template variables
	renderer := c.getTemplateRenderer()
	c.Reporter.Debug("Rendering repository secrets JSON with template variables")

	// Render the secrets JSON with template variables
	renderedJSON, err := renderer.RenderString(secretsJSON)
	if err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to render repository secrets JSON with template variables")
		return fmt.Errorf("failed to render repository secrets JSON with template variables: %w", err)
	}

	var repoSecretsConfig struct {
		Secrets []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Env       string `json:"env,omitempty"`
			Type      string `json:"type,omitempty"`      // "secret" or "variable"
			Reference string `json:"reference,omitempty"` // Reference to another secret by name
		} `json:"secrets"`
	}

	// Parse the rendered JSON string
	if err := json.Unmarshal([]byte(renderedJSON), &repoSecretsConfig); err != nil {
		c.Reporter.Failed(mainOperation, err, "Failed to parse repository secrets JSON")
		return fmt.Errorf("failed to parse repository secrets JSON: %w", err)
	}

	if len(repoSecretsConfig.Secrets) == 0 {
		c.Reporter.Skip(mainOperation, "No repository-specific secrets found")
		return nil
	}

	// Apply the secrets
	appliedCount := 0
	failedSecrets := []string{}

	for _, s := range repoSecretsConfig.Secrets {
		prefixedName := secrets.SanitizeSecretName(fmt.Sprintf("%s_%s", repo, s.Name))
		envScope := s.Env

		secretValue, sourceType, sourceKey, err := secrets.ResolveSecretValue(s.Value, s.Name)
		if err != nil {
			errMsg := fmt.Sprintf("Error resolving secret value for '%s': %v", s.Name, err)
			c.Reporter.Warning("Secret resolution", errMsg)
			failedSecrets = append(failedSecrets, errMsg)
			continue
		}

		if secretValue == "" {
			c.Reporter.Skip("Secret processing", fmt.Sprintf("Skipping secret '%s' with empty value", s.Name))
			continue
		}

		// Debug info about resolution
		if sourceType != "config" {
			c.Reporter.Debug(fmt.Sprintf("Resolved '%s' from %s: '%s'", s.Name, sourceType, sourceKey))
		}

		isVariable := s.Type == "variable"
		c.Reporter.Progress(mainOperation, 0, fmt.Sprintf("Setting %s: %s", valueOrEmpty(s.Type, "secret"), prefixedName))

		if err := c.applySecretOrVariable(ctx, isVariable, owner, repo, prefixedName, secretValue, envScope,
			"repo-secrets"); err != nil {
			errMsg := fmt.Sprintf("Failed to apply %s '%s': %v", valueOrEmpty(s.Type, "secret"), prefixedName, err)
			c.Reporter.Warning("Secret application", errMsg)
			failedSecrets = append(failedSecrets, errMsg)
		} else {
			appliedCount++
		}
	}

	// Return a summary error if any secrets failed
	if len(failedSecrets) > 0 {
		summaryMessage := fmt.Sprintf("Applied %d repository-specific secrets, %d failed", appliedCount, len(failedSecrets))
		c.Reporter.Complete(mainOperation, summaryMessage)
		return fmt.Errorf("some repository-specific secrets could not be applied: %s", strings.Join(failedSecrets[:1], ", "))
	}

	c.Reporter.Complete(mainOperation, fmt.Sprintf("Successfully applied %d repository-specific secrets", appliedCount))
	return nil
}

// Helper to apply a secret or variable
func (c *Factory) applySecretOrVariable(ctx context.Context, isVariable bool, owner, repo, name, value, env,
	valueSource string,
) error {
	operation := "Setting variable"
	resourceType := "variable"
	if !isVariable {
		operation = "Setting secret"
		resourceType = "secret"
	}

	c.Reporter.Progress(operation, 0, name)

	var err error
	if isVariable {
		err = c.GitHubClient.SetVariable(ctx, owner, repo, name, value, env)
	} else {
		err = c.GitHubClient.ApplySecret(ctx, owner, repo, name, value, env)
	}

	if err != nil {
		c.Reporter.Warning(operation, fmt.Sprintf("Failed to set %s '%s': %v", resourceType, name, err))
		return err
	}

	c.Reporter.Debug(fmt.Sprintf("Applied %s: %s (env: %s, source: %s)",
		resourceType, name, valueOrEmpty(env, "repo"), valueSource))
	return nil
}

// createRepository creates a GitHub repository and adds appropriate topics and labels
func (c *Factory) createRepository(ctx context.Context, opts Options) (string, error) {
	// Only print errors if needed, not process/info messages
	var org string
	if opts.AccountType == "organization" {
		org = opts.RepoOwner
	}
	repoURL, err := c.GitHubClient.CreateRepo(ctx, opts.RepoName, org, opts.IsPrivate, opts.RepoDescription)
	if err != nil {
		return "", err
	}

	// If this is an organization repository and we have a team, add it as admin to the repository
	if opts.AccountType == "organization" && opts.Team != "" {
		c.Reporter.Progress("Repository Setup", 50, fmt.Sprintf("Adding team '%s' as admin to repository", opts.Team))
		err := c.GitHubClient.AddTeamToRepository(ctx, org, opts.RepoName, opts.Team, gh.TeamPermissionAdmin)
		if err != nil {
			c.Reporter.Warning("Team Permission", fmt.Sprintf("Failed to add team '%s' as admin: %v", opts.Team, err))
			// Don't fail the entire operation - this is a non-critical enhancement
		} else {
			c.Reporter.Debug(fmt.Sprintf("Successfully added team '%s' as admin to repository", opts.Team))
		}
	}

	// Generate appropriate topics for the repository
	topics := []string{}

	// Add language topic
	if opts.Language != "" {
		topics = append(topics, strings.ToLower(opts.Language))
	}

	// Add project type topic
	if opts.ProjectType != "" {
		topics = append(topics, strings.ToLower(opts.ProjectType))
	}

	// Add team topic if provided
	if opts.Team != "" {
		topics = append(topics, strings.ToLower(strings.ReplaceAll(opts.Team, " ", "-")))
	}

	// Add viaplay-cli topic to identify repos created by this tool
	topics = append(topics, "viaplay-cli")

	// Add the topics to the repository
	if err := c.GitHubClient.AddTopicsToRepo(ctx, opts.RepoOwner, opts.RepoName, topics); err != nil {
		c.Reporter.Warning("Topic Creation", fmt.Sprintf("Failed to add topics to repository: %v", err))
		// Don't return an error here as topic creation is not critical to the repository creation
	} else {
		c.Reporter.Debug(fmt.Sprintf("Added topics to repository: %v", topics))
	}

	return repoURL, nil
}
