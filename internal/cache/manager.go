// Package cache provides template caching functionality.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/git"
	"github.com/nentgroup/viaplay-cli/internal/output"
)

// Config provides configuration for the cache manager
type Config struct {
	BaseCacheDir string
}

// SourceType represents the type of template source
type SourceType string

const (
	// SourceTypeGitHub represents a GitHub repository template source
	SourceTypeGitHub SourceType = "github"
	// SourceTypeLocal represents a local directory template source
	SourceTypeLocal SourceType = "local"
	// SourceTypeURL represents a URL template source (e.g., tarball)
	SourceTypeURL SourceType = "url"
	// SourceTypeSSH represents an SSH Git URL
	SourceTypeSSH SourceType = "ssh"
)

// Source represents a template source
type Source struct {
	// Type of the template source (github, local, url, ssh)
	Type SourceType
	// Location of the template source (repo, path, or URL)
	Location string
	// Reference is a Git reference (branch, tag, commit) for Git sources
	Reference string
}

// Template represents a cached template
type Template struct {
	// Language of the template (go, typescript, etc.)
	Language string
	// Type of the template (service, cli, etc.)
	Type string
	// Path to the template in the cache
	Path string
	// Last time the template was modified
	LastModified time.Time
	// Last time the template was used
	LastUsed time.Time
	// Source of the template
	Source Source
}

// TemplateInfo represents template information for display
type TemplateInfo struct {
	Language string
	Type     string
	Path     string
	LastUsed time.Time
}

// Manager manages the template cache
type Manager struct {
	// Configuration for the cache
	Config *config.Configuration
	// Base directory for cached templates
	BaseCacheDir string
}

// NewManager creates a new cache manager
func NewManager(cfg *config.Configuration) *Manager {
	if cfg == nil {
		// Create a default configuration
		defaultCfg := &config.Configuration{
			CacheDir: config.GetDefaultCacheDir(),
		}
		cfg = defaultCfg
	}

	// Ensure the cache directory exists
	cacheDir := cfg.CacheDir
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		fmt.Printf("Warning: Failed to create cache directory %s: %v\n", cacheDir, err)
	}

	return &Manager{
		Config:       cfg,
		BaseCacheDir: cacheDir,
	}
}

// NewManagerFromConfig creates a cache manager from a Configuration object
func NewManagerFromConfig(cfg *config.Configuration) *Manager {
	return NewManager(cfg)
}

// NewManagerWithCacheConfig creates a cache manager from a simple CacheConfig
func NewManagerWithCacheConfig(cfg *Config) *Manager {
	if cfg == nil {
		return NewManager(nil)
	}
	// Ensure the cache directory exists
	if err := os.MkdirAll(cfg.BaseCacheDir, 0o755); err != nil {
		fmt.Printf("Warning: Failed to create cache directory %s: %v\n", cfg.BaseCacheDir, err)
	}
	return &Manager{
		BaseCacheDir: cfg.BaseCacheDir,
	}
}

// ParseSource parses a template source string into a Source struct
// Format:
// - GitHub repo: "github@<owner>/<repo>.git[@branch/tag]"
// - Local path: "local@/path/to/template"
// - Tarball URL: "url@https://example.com/template.tar.gz"
// - Git SSH URL: "git@github.com:<owner>/<repo>.git" or "git@github.com/<owner>/<repo>.git"
func ParseSource(sourceStr string) (Source, error) {
	// Special handling for git@github.com: format with colon
	if strings.HasPrefix(sourceStr, "git@github.com:") {
		// For SSH URLs, store the full URL as is
		return Source{
			Type:     SourceTypeSSH,
			Location: sourceStr,
		}, nil
	}

	// Special handling for git@github.com/ format with slash instead of colon
	if strings.HasPrefix(sourceStr, "git@github.com/") {
		// Convert to standard format with colon
		correctedURL := "git@github.com:" + sourceStr[15:]
		return Source{
			Type:     SourceTypeSSH,
			Location: correctedURL,
		}, nil
	}

	parts := strings.SplitN(sourceStr, "@", 2)
	if len(parts) < 2 {
		// If it doesn't have an @ symbol but looks like a Git URL, treat it as SSH
		if strings.HasPrefix(sourceStr, "git") {
			return Source{
				Type:     SourceTypeSSH,
				Location: sourceStr,
			}, nil
		}
		return Source{}, fmt.Errorf("invalid template source format: %s", sourceStr)
	}

	sourceType := parts[0]
	location := parts[1]
	var reference string

	switch sourceType {
	case "github":
		// Check if there's a branch/tag/reference specified
		if idx := strings.LastIndex(location, "@"); idx != -1 {
			reference = location[idx+1:]
			location = location[:idx]
		}
		return Source{
			Type:      SourceTypeGitHub,
			Location:  location,
			Reference: reference,
		}, nil
	case "local":
		// Expand tilde in local path
		if strings.HasPrefix(location, "~") {
			location = config.ExpandPath(location)
		}
		return Source{
			Type:     SourceTypeLocal,
			Location: location,
		}, nil
	case "url":
		return Source{
			Type:     SourceTypeURL,
			Location: location,
		}, nil
	case "git":
		// Handle git URLs explicitly as SSH type
		return Source{
			Type:     SourceTypeSSH,
			Location: sourceStr,
		}, nil
	default:
		return Source{}, fmt.Errorf("unknown template source type: %s", sourceType)
	}
}

// GetTemplatePath returns the cache path for a specific template
func (m *Manager) GetTemplatePath(language, templateType string) string {
	return filepath.Join(m.BaseCacheDir, language, templateType)
}

// EnsureTemplate ensures a template is available in the cache
func (m *Manager) EnsureTemplate(language, templateType, sourceStr string) (string, error) {
	output.VerboseMessage(fmt.Sprintf("Ensuring template for %s/%s from source: %s", language, templateType, sourceStr))

	// Parse the template source
	source, err := ParseSource(sourceStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template source: %w", err)
	}

	// Handle local templates directly
	if source.Type == SourceTypeLocal {
		return handleLocalTemplate(source)
	}

	// Get the cache path for this template
	cachePath := m.GetTemplatePath(language, templateType)
	output.VerboseMessage(fmt.Sprintf("Template cache path: %s", cachePath))

	// Check if the template exists in the cache
	exists := git.IsGitRepository(cachePath)

	// Check if force update is needed
	if shouldForceUpdate(exists) {
		output.VerboseMessage(fmt.Sprintf("Force update requested, removing existing template at: %s", cachePath))
		if err := os.RemoveAll(cachePath); err != nil {
			return "", fmt.Errorf("failed to remove existing template for force update: %w", err)
		}
		exists = false
	}

	if exists {
		return handleExistingTemplate(cachePath, source)
	}

	return handleNewTemplate(cachePath, source)
}

// handleLocalTemplate verifies and returns the path for a local template
func handleLocalTemplate(source Source) (string, error) {
	// For local templates, we don't need to clone or update anything
	// Just verify the path exists
	if _, err := os.Stat(source.Location); err != nil {
		return "", fmt.Errorf("local template path does not exist: %s", source.Location)
	}
	// Use the local path directly
	return source.Location, nil
}

// shouldForceUpdate checks if a force update is requested via environment variable
func shouldForceUpdate(exists bool) bool {
	return os.Getenv("VIAPLAY_CLI_FORCE_UPDATE") == "true" && exists
}

// handleExistingTemplate handles logic for an existing template in the cache
func handleExistingTemplate(cachePath string, source Source) (string, error) {
	output.VerboseMessage(fmt.Sprintf("Template already exists in cache at: %s", cachePath))

	// Check if the template is too old (older than 24 hours)
	// If it is, we'll force an update check
	forcedCheck := isTemplateTooOld(cachePath)
	if forcedCheck {
		output.VerboseMessage("Template is older than 24 hours, checking for updates")
	}

	// Check if the template needs to be updated
	needsUpdate, err := checkIfTemplateNeedsUpdate(cachePath)
	if err != nil && !forcedCheck {
		// If there's an error checking updates, use cached version anyway
		output.VerboseMessage(fmt.Sprintf("Error checking updates: %v, using cached template", err))
		return cachePath, nil
	}

	if needsUpdate || forcedCheck {
		output.VerboseMessage("Updating template...")
		if err := updateExistingTemplate(cachePath, source); err != nil {
			// If update fails, use cached version anyway
			output.VerboseMessage(fmt.Sprintf("Failed to update template: %v", err))
			output.VerboseMessage("Using cached template despite update failure")
		} else {
			output.VerboseMessage("Template updated successfully")
		}
	} else {
		output.VerboseMessage("Template is up to date, using cached version")
	}

	return cachePath, nil
}

// checkIfTemplateNeedsUpdate checks if a template needs to be updated
func checkIfTemplateNeedsUpdate(cachePath string) (bool, error) {
	// Try to determine if the repository needs an update by checking Git
	repoInfo, err := git.GetRepositoryInfo(cachePath)
	if err != nil {
		return false, fmt.Errorf("error getting repository info: %w", err)
	}

	// Check if the local repository is behind the remote
	// Use "origin" as the default remote name since RepositoryInfo doesn't have RemoteName field
	isBehind, err := git.IsBehindRemote(cachePath, "origin", repoInfo.Branch)
	if err != nil {
		return false, fmt.Errorf("error checking if repository is behind remote: %w", err)
	}

	return isBehind, nil
}

// updateExistingTemplate updates an existing template in the cache
func updateExistingTemplate(cachePath string, source Source) error {
	return git.Update(git.UpdateOptions{
		Directory: cachePath,
		Branch:    source.Reference,
		Force:     false,
	})
}

// handleNewTemplate handles cloning a new template
func handleNewTemplate(cachePath string, source Source) (string, error) {
	output.VerboseMessage(fmt.Sprintf("Template not found in cache, cloning to: %s", cachePath))

	// Prepare git URL based on source type
	gitURL, err := getGitURLFromSource(source)
	if err != nil {
		return "", err
	}

	// Clone the repository
	if err := git.Clone(git.CloneOptions{
		URL:       gitURL,
		Branch:    source.Reference,
		Directory: cachePath,
	}); err != nil {
		// If cloning fails, provide a clear error message for offline scenarios
		return "", fmt.Errorf("failed to clone template: %w (if you're offline, you need a cached template first)", err)
	}

	output.VerboseMessage("Template cloned successfully")
	return cachePath, nil
}

// getGitURLFromSource constructs a git URL from a source
func getGitURLFromSource(source Source) (string, error) {
	switch source.Type {
	case SourceTypeGitHub:
		return fmt.Sprintf("git@github.com:%s", source.Location), nil
	case SourceTypeSSH:
		return source.Location, nil
	case SourceTypeURL:
		return "", fmt.Errorf("URL template sources not yet implemented")
	default:
		return "", fmt.Errorf("unsupported template source type: %s", source.Type)
	}
}

// ListTemplates lists all templates in the cache
func (m *Manager) ListTemplates() ([]Template, error) {
	templates := []Template{}

	// Walk through the cache directory to find templates
	err := filepath.Walk(m.BaseCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == m.BaseCacheDir {
			return nil
		}

		// Get relative path from the cache directory
		rel, err := filepath.Rel(m.BaseCacheDir, path)
		if err != nil {
			return err
		}

		// We want to list only language/template_type directories
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) == 2 && info.IsDir() {
			// Check if it's a git repository
			if git.IsGitRepository(path) {
				// Try to get the source information from the git repository
				source, err := getTemplateSourceFromGit(path)
				if err != nil {
					// If we can't determine the source, use a default
					source = Source{
						Type:     SourceTypeGitHub,
						Location: "unknown",
					}
				}

				templates = append(templates, Template{
					Language:     parts[0],
					Type:         parts[1],
					Path:         path,
					LastModified: info.ModTime(),
					Source:       source,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing templates: %w", err)
	}

	return templates, nil
}

// PruneCache removes templates from the cache that are older than the specified days
func (m *Manager) PruneCache(days int) (int, int, error) {
	// Verify the cache directory exists
	if _, err := os.Stat(m.BaseCacheDir); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("cache directory does not exist: %s", m.BaseCacheDir)
	}

	// Get the cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// Track how many templates were pruned
	prunedCount := 0
	totalTemplates := 0

	// Walk through the templates cache directory
	err := filepath.Walk(m.BaseCacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip the root directory
		if path == m.BaseCacheDir {
			return nil
		}

		// Only process directories at the language level (one level deep)
		rel, err := filepath.Rel(m.BaseCacheDir, path)
		if err != nil {
			return fmt.Errorf("could not determine relative path for %s: %w", path, err)
		}

		// If it's a directory and the modification time is older than the cutoff
		if info.IsDir() && !strings.Contains(rel, string(os.PathSeparator)) {
			// Check each template type within this language directory
			languagePath := path
			templateTypes, err := os.ReadDir(languagePath)
			if err != nil {
				return fmt.Errorf("could not read directory %s: %w", languagePath, err)
			}

			for _, templateType := range templateTypes {
				if !templateType.IsDir() {
					continue
				}

				totalTemplates++
				templatePath := filepath.Join(languagePath, templateType.Name())
				templateInfo, err := os.Stat(templatePath)
				if err != nil {
					continue
				}

				// If the template is older than the cutoff, remove it
				if templateInfo.ModTime().Before(cutoffTime) {
					if err := os.RemoveAll(templatePath); err != nil {
						return fmt.Errorf("error removing template %s: %w", templatePath, err)
					}
					prunedCount++
				}
			}
		}

		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error while pruning cache: %w", err)
	}

	return prunedCount, totalTemplates, nil
}

// CleanCache removes all templates from the cache
func (m *Manager) CleanCache() error {
	// Remove the entire cache directory
	if err := os.RemoveAll(m.BaseCacheDir); err != nil {
		return fmt.Errorf("error cleaning cache: %w", err)
	}

	// Recreate the cache directory
	if err := os.MkdirAll(m.BaseCacheDir, 0o755); err != nil {
		return fmt.Errorf("error recreating cache directory: %w", err)
	}

	return nil
}

// UpdateAllTemplates updates all templates in the cache
func (m *Manager) UpdateAllTemplates() (int, int, error) {
	templates, err := m.ListTemplates()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list templates: %w", err)
	}

	// Track success and failure counts
	var (
		successCount int
		failCount    int
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	for _, tmpl := range templates {
		wg.Add(1)
		go func(t Template) {
			defer wg.Done()

			// Force update by setting environment variable
			os.Setenv("VIAPLAY_CLI_FORCE_UPDATE", "true")
			defer os.Unsetenv("VIAPLAY_CLI_FORCE_UPDATE")

			// Generate source string
			var sourceStr string

			switch t.Source.Type {
			case SourceTypeGitHub:
				if t.Source.Reference != "" {
					sourceStr = fmt.Sprintf("github@%s@%s", t.Source.Location, t.Source.Reference)
				} else {
					sourceStr = fmt.Sprintf("github@%s", t.Source.Location)
				}
			case SourceTypeSSH:
				sourceStr = t.Source.Location
			default:
				// Skip unknown source types
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			fmt.Printf("Updating template: %s/%s from %s\n", t.Language, t.Type, sourceStr)

			// Use EnsureTemplate to update the template
			_, err := m.EnsureTemplate(t.Language, t.Type, sourceStr)
			if err != nil {
				fmt.Printf("Error updating template %s/%s: %v\n", t.Language, t.Type, err)
				mu.Lock()
				failCount++
				mu.Unlock()
			} else {
				fmt.Printf("Successfully updated template: %s/%s\n", t.Language, t.Type)
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(tmpl)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return successCount, failCount, nil
}

// GetCacheDir returns the base cache directory path
func (m *Manager) GetCacheDir() string {
	return m.BaseCacheDir
}

// getTemplateSourceFromGit attempts to determine the template source from a git repository
func getTemplateSourceFromGit(repoPath string) (Source, error) {
	// Get repository info
	repoInfo, err := git.GetRepositoryInfo(repoPath)
	if err != nil {
		return Source{}, err
	}

	// Extract source information from the remote URL
	remoteURL := repoInfo.RemoteURL

	// Handle different remote URL formats
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		// Format: git@github.com:owner/repo.git
		parts := strings.SplitN(remoteURL, ":", 2)
		if len(parts) != 2 {
			return Source{
				Type:     SourceTypeSSH,
				Location: remoteURL,
			}, nil
		}

		repoPath := parts[1]
		// Remove .git suffix if present
		repoPath = strings.TrimSuffix(repoPath, ".git")
		return Source{
			Type:      SourceTypeGitHub,
			Location:  repoPath,
			Reference: repoInfo.Branch,
		}, nil
	}
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		repoPath := strings.TrimPrefix(remoteURL, "https://github.com/")
		repoPath = strings.TrimSuffix(repoPath, ".git")
		return Source{
			Type:      SourceTypeGitHub,
			Location:  repoPath,
			Reference: repoInfo.Branch,
		}, nil
	}

	// For other formats, use the SSH type with the full URL
	return Source{
		Type:     SourceTypeSSH,
		Location: remoteURL,
	}, nil
}

// isTemplateTooOld checks if a template's age exceeds the freshness threshold
// It returns true if the template is older than 24 hours
func isTemplateTooOld(templatePath string) bool {
	// Check if the template path exists
	info, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		// Template path doesn't exist, consider it too old
		return true
	} else if err != nil {
		// Error checking template path, treat as old
		return true
	}

	// Calculate the age of the template
	age := time.Since(info.ModTime())

	// Check if the age exceeds 24 hours
	return age > 24*time.Hour
}
