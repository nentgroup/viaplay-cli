// Package git provides Git operations for viaplay-cli.
// It handles repository management, including cloning, updating, and manipulation of Git repositories.
package git

import (
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
func Clone(opts CloneOptions) error {
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

	// Execute the git clone command
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
func Update(opts UpdateOptions) error {
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

	// Fetch latest updates
	fetchCmd := exec.Command("git", "fetch")
	fetchCmd.Stdout = os.Stdout
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	// If a branch is specified, check it out
	if opts.Branch != "" {
		// Validate the branch name for security
		if !validateGitBranch(opts.Branch) {
			return fmt.Errorf("invalid branch name: %s", opts.Branch)
		}

		// Checkout the branch
		checkoutCmd := exec.Command("git", "checkout", opts.Branch)
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout branch '%s': %w", opts.Branch, err)
		}

		// Pull the latest changes
		pullCmd := exec.Command("git", "pull", "origin", opts.Branch)
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return fmt.Errorf("failed to pull updates from '%s': %w", opts.Branch, err)
		}
	} else {
		// Pull the latest changes from the current branch
		pullCmd := exec.Command("git", "pull")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return fmt.Errorf("failed to pull updates: %w", err)
		}
	}

	return nil
}

// GetRepositoryInfo retrieves information about a Git repository
func GetRepositoryInfo(directory string) (*RepositoryInfo, error) {
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
	remoteCmd := exec.Command("git", "config", "--get", "remote.origin.url")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}
	remoteURL := strings.TrimSpace(string(remoteOutput))

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get last commit information
	logCmd := exec.Command("git", "log", "-1", "--pretty=format:%H|%an|%ad|%s")
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
func InitRepository(directory string) error {
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
	initCmd := exec.Command("git", "init")
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	return nil
}

// AddRemote adds a remote to a Git repository
func AddRemote(directory, name, url string) error {
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
	remoteCmd := exec.Command("git", "remote", "add", name, url)
	if err := remoteCmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// CommitAll commits all changes in a Git repository
func CommitAll(directory, message string) error {
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
	addCmd := exec.Command("git", "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
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
func Push(directory, remote, branch string) error {
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
	pushCmd := exec.Command("git", "push", "-u", remote, branch)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	return nil
}

// IsBehindRemote checks if the local repository is behind the remote branch.
// Returns true if the local repo is behind the remote, false otherwise, and any error encountered.
func IsBehindRemote(repoPath, remoteName, branch string) (bool, error) {
	// Validate the input parameters for security
	if !validateGitBranch(branch) {
		return false, fmt.Errorf("invalid branch name: %s", branch)
	}

	if !validateGitRemote(remoteName) {
		return false, fmt.Errorf("invalid remote name: %s", remoteName)
	}

	// Fetch latest from remote
	fetchCmd := exec.Command("git", "fetch", remoteName)
	fetchCmd.Dir = repoPath
	if err := fetchCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get number of commits the local branch is behind remote
	revListArg := fmt.Sprintf("HEAD..%s/%s", remoteName, branch)
	behindCmd := exec.Command("git", "rev-list", "--count", revListArg)
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
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			strings.ContainsRune(allowedSpecialChars, c)) {
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
