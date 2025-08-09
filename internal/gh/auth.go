// Package gh provides GitHub API integration for viaplay-cli.
// It handles authentication, repository operations, and other GitHub-specific functionality.
package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/pkg/browser"
	"github.com/zalando/go-keyring"
)

// KeyringNamespace is the namespace used for storing tokens in the system keyring
const KeyringNamespace = "viaplaycli"

// Embed client ID in the binary during compilation with:
// go build -ldflags="-X github.com/nentgroup/viaplay-cli/internal/gh.defaultClientID=your-client-id"
var (
	// defaultClientID can be set at build time
	defaultClientID string
)

type deviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// Authenticate performs GitHub device flow, stores the token, and returns it
func Authenticate() (string, error) {
	// Check if token already exists
	token, err := keyring.Get(KeyringNamespace, "token")
	if err == nil && token != "" {
		return token, nil
	}

	// Get GitHub client ID
	clientID := getClientID()
	if clientID == "" {
		return "", fmt.Errorf("GitHub OAuth client ID not available: set GITHUB_CLIENT_ID environment variable or build with embedded client ID")
	}

	ctx := context.Background()
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", DefaultScopes)
	request, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/device/code", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating device code request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json") // Add this header to force JSON response
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("error requesting device code: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading device code response: %w", err)
	}

	var dcResp deviceCodeResponse
	if err := json.Unmarshal(bodyBytes, &dcResp); err != nil {
		return "", fmt.Errorf("error parsing device code response: %w\nRaw response: %s", err, string(bodyBytes))
	}

	// Check if we got a valid verification URI
	if dcResp.VerificationURI == "" {
		return "", fmt.Errorf("invalid device code response: missing verification_uri\nRaw response: %s", string(bodyBytes))
	}

	// Use VerificationURIComplete if available, otherwise construct the URL
	verificationURL := dcResp.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = fmt.Sprintf("%s?user_code=%s", dcResp.VerificationURI, dcResp.UserCode)
	}

	fmt.Printf("To authorize, visit: %s\n", verificationURL)
	fmt.Printf("Enter code: %s\n", dcResp.UserCode)

	if err := browser.OpenURL(verificationURL); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open browser: %v\n", err)
	}
	fmt.Println("Waiting for authorization in your browser...")

	// Step 2: Poll for access token
	form = url.Values{}
	form.Set("client_id", clientID)
	form.Set("device_code", dcResp.DeviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	var tokenResp accessTokenResponse
	for {
		time.Sleep(time.Duration(dcResp.Interval) * time.Second)
		request, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
		if err != nil {
			return "", fmt.Errorf("error creating access token request: %w", err)
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept", "application/json") // Add this header to force JSON response
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return "", fmt.Errorf("error requesting access token: %w", err)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("error reading access token response: %w", err)
		}
		if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
			return "", fmt.Errorf("error parsing access token response: %w\nRaw response: %s", err, string(bodyBytes))
		}
		if tokenResp.AccessToken != "" {
			token = tokenResp.AccessToken
			break
		}
		if tokenResp.Error != "authorization_pending" && tokenResp.Error != "slow_down" {
			return "", errors.New(tokenResp.ErrorDesc)
		}
	}

	if err := keyring.Set(KeyringNamespace, "token", token); err != nil {
		return "", fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return token, nil
}

func getClientID() string {
	// First check environment variable (for development)
	if envID := os.Getenv("GITHUB_CLIENT_ID"); envID != "" {
		return envID
	}

	// Fall back to the embedded value (for production builds)
	return defaultClientID
}

// GetToken retrieves the GitHub token from the keyring
func GetToken() (string, error) {
	token, err := keyring.Get(KeyringNamespace, "token")
	if err != nil {
		return "", err
	}
	return token, nil
}

// DeleteToken removes the GitHub token from the keyring
func DeleteToken() error {
	return keyring.Delete(KeyringNamespace, "token")
}

// GetAuthenticatedUser returns the username of the authenticated user using the token and go-github
func GetAuthenticatedUser(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("no token provided")
	}
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(token)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}
