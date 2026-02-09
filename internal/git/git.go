// Package git provides Git operations for viaplay-cli.
// It handles repository management, including cloning, updating, and manipulation of Git repositories.
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions represents options for cloning a Git repository
type CloneOptions struct {
	URL       string // Repository URL
	Branch    string // Branch to clone (optional)
	Directory string // Destination directory
	Depth     int    // Depth for shallow clones (0 for full clone)
}

// UpdateOptions represents options for updating a Git repository
type UpdateOptions struct {
	Directory string // Repository directory
	Branch    string // Branch to checkout and update (optional)
	Force     bool   // Whether to force update
}

// RepositoryInfo contains information about a Git repository
type RepositoryInfo struct {
	RemoteURL  string     // Remote URL (origin)
	Branch     string     // Current branch
	LastCommit CommitInfo // Last commit information
}

// CommitInfo contains information about a Git commit
type CommitInfo struct {
	Hash    string // Commit hash
	Message string // Commit message
	Author  string // Author name
	Date    string // Commit date
}

// Clone clones a Git repository to the specified directory
func Clone(ctx context.Context, opts CloneOptions) error {
	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(opts.Directory), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Build the clone command
	args := []string{"clone"}

	// Add depth if specified
	if opts.Depth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", opts.Depth))
	}

	// Add branch if specified
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}

	// Add URL and destination
	args = append(args, opts.URL, opts.Directory)

	// Execute the git clone command - capture output instead of sending to terminal
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w, output: %s", err, string(output))
	}

	return nil
}

// validateGitBranch checks if a branch name is valid and not malicious
func validateGitBranch(branch string) bool {
	// Branches shouldn't contain spaces, control characters, or escape sequences
	for _, c := range branch {
		if c <= 32 || c == 127 { // ASCII control characters or space
			return false
		}
	}

	// Branches shouldn't contain certain dangerous characters
	dangerousChars := []string{";", "&&", "||", ">", "<", "`", "$", "\\", "\"", "'"}
	for _, char := range dangerousChars {
		if strings.Contains(branch, char) {
			return false
		}
	}

	return true
}

// Update updates a Git repository to the latest changes
func Update(ctx context.Context, opts UpdateOptions) error {
	// Verify the directory exists and is a git repository
	if !IsGitRepository(opts.Directory) {
		return fmt.Errorf("not a git repository: %s", opts.Directory)
	}

	// Change to the repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(opts.Directory); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Fetch latest updates - capture output instead of sending to terminal
	fetchCmd := exec.CommandContext(ctx, "git", "fetch")
	fetchOutput, err := fetchCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch updates: %w, output: %s", err, string(fetchOutput))
	}

	if opts.Branch != "" {
		return updateSpecificBranch(ctx, opts.Branch)
	}

	// Pull the latest changes from the current branch - capture output instead of sending to terminal
	pullCmd := exec.CommandContext(ctx, "git", "pull")
	pullOutput, err := pullCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull updates: %w, output: %s", err, string(pullOutput))
	}

	return nil
}

// updateSpecificBranch validates and checks out a branch, then pulls latest changes.
func updateSpecificBranch(ctx context.Context, branch string) error {
	// Validate the branch name for security
	if !validateGitBranch(branch) {
		return fmt.Errorf("invalid branch name: %s", branch)
	}

	// Checkout the branch - capture output instead of sending to terminal
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", branch)
	checkoutOutput, err := checkoutCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch '%s': %w, output: %s", branch, err, string(checkoutOutput))
	}

	// Pull the latest changes - capture output instead of sending to terminal
	pullCmd := exec.CommandContext(ctx, "git", "pull", "origin", branch)
	pullOutput, err := pullCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull updates from '%s': %w, output: %s", branch, err, string(pullOutput))
	}

	return nil
}

// GetRepositoryInfo retrieves information about a Git repository
func GetRepositoryInfo(ctx context.Context, directory string) (*RepositoryInfo, error) {
	// Verify the directory exists and is a git repository
	gitDir := filepath.Join(directory, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository: %s", directory)
	}

	// Change to the repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(directory); err != nil {
		return nil, fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Get remote URL
	remoteCmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}
	remoteURL := strings.TrimSpace(string(remoteOutput))

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get last commit information
	logCmd := exec.CommandContext(ctx, "git", "log", "-1", "--pretty=format:%H|%an|%ad|%s")
	logOutput, err := logCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit information: %w", err)
	}

	// Parse commit information
	commitParts := strings.Split(string(logOutput), "|")
	commitInfo := CommitInfo{}

	if len(commitParts) >= 4 {
		commitInfo.Hash = commitParts[0]
		commitInfo.Author = commitParts[1]
		commitInfo.Date = commitParts[2]
		commitInfo.Message = commitParts[3]
	}

	return &RepositoryInfo{
		RemoteURL:  remoteURL,
		Branch:     branch,
		LastCommit: commitInfo,
	}, nil
}

// IsGitRepository checks if a directory is a Git repository
func IsGitRepository(directory string) bool {
	gitDir := filepath.Join(directory, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// InitRepository initialises a new Git repository
func InitRepository(ctx context.Context, directory string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Change to the directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("failed to change to directory: %w", err)
	}

	// Initialise the repository
	initCmd := exec.CommandContext(ctx, "git", "init")
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	return nil
}

// AddRemote adds a remote to a Git repository
func AddRemote(ctx context.Context, directory, name, url string) error {
	// Verify the directory exists and is a git repository
	if !IsGitRepository(directory) {
		return fmt.Errorf("not a git repository: %s", directory)
	}

	// Change to the repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Add the remote
	remoteCmd := exec.CommandContext(ctx, "git", "remote", "add", name, url)
	if err := remoteCmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// CommitAll commits all changes in a Git repository
func CommitAll(ctx context.Context, directory, message string) error {
	// Verify the directory exists and is a git repository
	if !IsGitRepository(directory) {
		return fmt.Errorf("not a git repository: %s", directory)
	}

	// Change to the repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Add all files
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Commit
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		// Check if there's nothing to commit
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil // No error if there's nothing to commit
		}
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// Push pushes changes to a remote branch
func Push(ctx context.Context, directory, remote, branch string) error {
	// Verify the directory exists and is a git repository
	if !IsGitRepository(directory) {
		return fmt.Errorf("not a git repository: %s", directory)
	}

	// Change to the repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			fmt.Printf("Warning: Failed to return to original directory: %v\n", err)
		}
	}()

	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Push to the remote
	pushCmd := exec.CommandContext(ctx, "git", "push", "-u", remote, branch)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	return nil
}

// IsBehindRemote checks if the local repository is behind the remote branch.
// Returns true if the local repo is behind the remote, false otherwise, and any error encountered.
func IsBehindRemote(ctx context.Context, repoPath, remoteName, branch string) (bool, error) {
	// Validate the input parameters for security
	if !validateGitBranch(branch) {
		return false, fmt.Errorf("invalid branch name: %s", branch)
	}

	if !validateGitRemote(remoteName) {
		return false, fmt.Errorf("invalid remote name: %s", remoteName)
	}

	// Fetch latest from remote - capture output instead of sending to terminal
	fetchCmd := exec.CommandContext(ctx, "git", "fetch", remoteName)
	fetchCmd.Dir = repoPath
	fetchOutput, err := fetchCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w, output: %s", err, string(fetchOutput))
	}

	// Get number of commits the local branch is behind remote
	revListArg := fmt.Sprintf("HEAD..%s/%s", remoteName, branch)
	behindCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", revListArg)
	behindCmd.Dir = repoPath
	behindOutput, err := behindCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check commits behind: %w", err)
	}

	// Parse the output (should be a number)
	behindCount := strings.TrimSpace(string(behindOutput))
	// If the count is greater than 0, the local branch is behind remote
	return behindCount != "0", nil
}

// validateGitRemote checks if a remote name is valid and not malicious
func validateGitRemote(remote string) bool {
	// Remote names should follow git naming conventions
	// Only alphanumeric characters and some special characters are allowed
	allowedSpecialChars := "-._"

	for _, c := range remote {
		if (c < 'a' || c > 'z') &&
			(c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			!strings.ContainsRune(allowedSpecialChars, c) {
			return false
		}
	}

	// Dangerous characters should be rejected
	dangerousChars := []string{";", "&&", "||", ">", "<", "`", "$", "\\", "\"", "'", " "}
	for _, char := range dangerousChars {
		if strings.Contains(remote, char) {
			return false
		}
	}

	return true
}

// ConvertToSSHURL converts an HTTPS GitHub URL to SSH format
// Example: https://github.com/user/repo.git -> git@github.com:user/repo.git
func ConvertToSSHURL(url string) string {
	// If it's already an SSH URL, return it as is
	if strings.HasPrefix(url, "git@") {
		return url
	}

	// Convert HTTPS URL to SSH
	// Format: https://github.com/user/repo.git -> git@github.com:user/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		path := strings.TrimPrefix(url, "https://github.com/")
		return "git@github.com:" + path
	}

	// If it's not a recognised format, return the original URL
	return url
}

// DescribeVersion returns a human-friendly version string for the repository at repoPath.
// It prefers the latest tag reachable from HEAD; if none is found, it falls back to a
// short commit hash. On failure it returns an error so callers can decide how to
// represent unknown versions.
func DescribeVersion(ctx context.Context, repoPath string) (string, error) {
	if !IsGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Try to get the latest tag reachable from HEAD (similar to `git describe --tags --abbrev=0`).
	describeTagCmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
	describeTagCmd.Dir = repoPath
	tagOutput, err := describeTagCmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(tagOutput))
		if version != "" {
			return version, nil
		}
	}

	// Fallback: use short commit hash of HEAD.
	shaCmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	shaCmd.Dir = repoPath
	shaOutput, err := shaCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to determine version for repo %s: %w", repoPath, err)
	}

	version := strings.TrimSpace(string(shaOutput))
	if version == "" {
		return "", fmt.Errorf("empty version string for repo %s", repoPath)
	}

	return version, nil
}
