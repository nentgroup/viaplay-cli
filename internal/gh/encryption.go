package gh

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v74/github"
	"golang.org/x/crypto/nacl/box"
)

// EncryptSecret encrypts a secret value using GitHub's public key
// Returns the encrypted value and the key ID
func EncryptSecret(publicKey *github.PublicKey, secretValue string) (string, string, error) {
	// Decode the public key
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return "", "", fmt.Errorf("failed to decode public key: %w", err)
	}

	// Convert the key to the format needed for box.SealAnonymous
	var boxKey [32]byte
	copy(boxKey[:], decodedPublicKey)

	// Encrypt the secret
	encryptedBytes, err := box.SealAnonymous(nil, []byte(secretValue), &boxKey, rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Encode the encrypted value as base64
	encryptedValue := base64.StdEncoding.EncodeToString(encryptedBytes)

	return encryptedValue, publicKey.GetKeyID(), nil
}

// ApplySecret is a helper that gets the public key, encrypts the secret, and sets it
// This simplifies the process of adding a secret to a repository or environment
func (ghc *GitHubClient) ApplySecret(ctx context.Context, owner, repo, secretName, secretValue string,
	env ...string,
) error {
	// Get the public key
	publicKey, err := ghc.GetPublicKey(ctx, owner, repo, env...)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Encrypt the secret
	encryptedValue, keyID, err := EncryptSecret(publicKey, secretValue)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Set the secret
	return ghc.SetSecret(ctx, owner, repo, secretName, encryptedValue, keyID, env...)
}
