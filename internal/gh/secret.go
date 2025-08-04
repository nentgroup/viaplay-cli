package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
)

// SetSecret sets a repository or environment secret
func (ghc *GitHubClient) SetSecret(ctx context.Context, owner, repo, secretName, encryptedValue, keyID string,
	env ...string,
) error {
	secret := &github.EncryptedSecret{
		Name:           secretName,
		EncryptedValue: encryptedValue,
		KeyID:          keyID,
	}

	repox, _, err := ghc.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return err
	}
	repoID := repox.GetID()

	if len(env) > 0 && env[0] != "" {
		///(s *ActionsService) CreateOrUpdateEnvSecret(ctx context.Context, repoID int, env string, eSecret *EncryptedSecret)
		_, err = ghc.client.Actions.CreateOrUpdateEnvSecret(ctx, int(repoID), env[0], secret)
		return err
	}
	_, err = ghc.client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
	return err
}

// GetPublicKey retrieves the public key for a repository, used for encrypting secrets
func (ghc *GitHubClient) GetPublicKey(owner, repo string, env ...string) (*github.PublicKey, error) {
	if len(env) > 0 && env[0] != "" {
		// Get environment public key
		repox, _, err := ghc.client.Repositories.Get(ghc.ctx, owner, repo)
		if err != nil {
			return nil, err
		}
		repoID := repox.GetID()
		key, _, err := ghc.client.Actions.GetEnvPublicKey(ghc.ctx, int(repoID), env[0])
		return key, err
	}

	// Get repository public key
	key, _, err := ghc.client.Actions.GetRepoPublicKey(ghc.ctx, owner, repo)
	return key, err
}

// ListSecrets returns all secrets for a repository or environment
func (ghc *GitHubClient) ListSecrets(owner, repo string, env ...string) ([]*github.Secret, error) {
	if len(env) > 0 && env[0] != "" {
		// List environment secrets
		repox, _, err := ghc.client.Repositories.Get(ghc.ctx, owner, repo)
		if err != nil {
			return nil, err
		}
		repoID := repox.GetID()

		secrets, _, err := ghc.client.Actions.ListEnvSecrets(ghc.ctx, int(repoID), env[0], nil)
		if err != nil {
			return nil, err
		}
		return secrets.Secrets, nil
	}

	// List repository secrets
	secrets, _, err := ghc.client.Actions.ListRepoSecrets(ghc.ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	return secrets.Secrets, nil
}

// DeleteSecret deletes a secret from a repository or environment
func (ghc *GitHubClient) DeleteSecret(owner, repo, secretName string, env ...string) error {
	if len(env) > 0 && env[0] != "" {
		// Delete environment secret
		repox, _, err := ghc.client.Repositories.Get(ghc.ctx, owner, repo)
		if err != nil {
			return err
		}
		repoID := repox.GetID()

		_, err = ghc.client.Actions.DeleteEnvSecret(ghc.ctx, int(repoID), env[0], secretName)
		return err
	}

	// Delete repository secret
	_, err := ghc.client.Actions.DeleteRepoSecret(ghc.ctx, owner, repo, secretName)
	return err
}

// SetVariable sets a repository or environment variable
func (ghc *GitHubClient) SetVariable(owner, repo, name, value string, env ...string) error {
	variable := &github.ActionsVariable{
		Name:  name,
		Value: value,
	}

	if len(env) > 0 && env[0] != "" {
		// Set environment variable
		_, err := ghc.client.Actions.CreateEnvVariable(ghc.ctx, owner, repo, env[0], variable)
		return err
	}

	// Set repository variable
	_, err := ghc.client.Actions.CreateRepoVariable(ghc.ctx, owner, repo, variable)
	return err
}

// ListVariables returns all variables for a repository or environment
func (ghc *GitHubClient) ListVariables(owner, repo string, env ...string) ([]*github.ActionsVariable, error) {
	if len(env) > 0 && env[0] != "" {
		// List environment variables
		vars, _, err := ghc.client.Actions.ListEnvVariables(ghc.ctx, owner, repo, env[0], nil)
		if err != nil {
			return nil, err
		}
		return vars.Variables, nil
	}

	// List repository variables
	vars, _, err := ghc.client.Actions.ListRepoVariables(ghc.ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	return vars.Variables, nil
}

// DeleteVariable deletes a variable from a repository or environment
func (ghc *GitHubClient) DeleteVariable(owner, repo, name string, env ...string) error {
	if len(env) > 0 && env[0] != "" {
		// Delete environment variable
		_, err := ghc.client.Actions.DeleteEnvVariable(ghc.ctx, owner, repo, env[0], name)
		return err
	}

	// Delete repository variable
	_, err := ghc.client.Actions.DeleteRepoVariable(ghc.ctx, owner, repo, name)
	return err
}
