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
)

// Config provides configuration for the cache manager
// (renamed from CacheConfig to avoid stutter)
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
// - Git SSH URL: "git@github.com:<owner>/<repo>.git"
func ParseSource(sourceStr string) (Source, error) {
	// Special handling for git@github.com format
	if strings.HasPrefix(sourceStr, "git@github.com:") {
		// For SSH URLs, store the full URL as is
		return Source{
			Type:     SourceTypeSSH,
			Location: sourceStr,
		}, nil
	}

	parts := strings.SplitN(sourceStr, "@", 2)
	if len(parts) < 2 {
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
	fmt.Printf("Ensuring template for %s/%s from source: %s\n", language, templateType, sourceStr)

	// Parse the template source
	source, err := ParseSource(sourceStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template source: %w", err)
	}

	// Handle local templates directly
	if source.Type == SourceTypeLocal {
		// For local templates, we don't need to clone or update anything
		// Just verify the path exists
		if _, err := os.Stat(source.Location); err != nil {
			return "", fmt.Errorf("local template path does not exist: %s", source.Location)
		}
		// Use the local path directly
		return source.Location, nil
	}

	// Get the cache path for this template
	cachePath := m.GetTemplatePath(language, templateType)
	fmt.Printf("Template cache path: %s\n", cachePath)

	// Check if the template exists in the cache
	exists := git.IsGitRepository(cachePath)

	// If force update is needed, remove existing template entirely
	if forceUpdate := os.Getenv("VIAPLAY_CLI_FORCE_UPDATE") == "true"; forceUpdate && exists {
		fmt.Printf("Force update requested, removing existing template at: %s\n", cachePath)
		if err := os.RemoveAll(cachePath); err != nil {
			return "", fmt.Errorf("failed to remove existing template for force update: %w", err)
		}
		exists = false
	}

	if exists {
		fmt.Printf("Template already exists in cache at: %s\n", cachePath)

		// Check if the template is fresh enough (less than 24 hours old)
		info, err := os.Stat(cachePath)
		if err != nil || time.Since(info.ModTime()) > 24*time.Hour {
			fmt.Println("Template is too old, updating...")

			// Template exists but needs to be updated
			err := git.Update(git.UpdateOptions{
				Directory: cachePath,
				Branch:    source.Reference,
				Force:     false,
			})
			if err != nil {
				return "", fmt.Errorf("failed to update template: %w", err)
			}

			fmt.Println("Template updated successfully")
		} else {
			fmt.Println("Template is fresh, using cached version")
		}
	} else {
		fmt.Printf("Template not found in cache, cloning to: %s\n", cachePath)

		// Prepare git URL based on source type
		var gitURL string

		switch source.Type {
		case SourceTypeGitHub:
			gitURL = fmt.Sprintf("git@github.com:%s", source.Location)
		case SourceTypeSSH:
			gitURL = source.Location
		case SourceTypeURL:
			// For URL templates, we need to download and extract
			return "", fmt.Errorf("URL template sources not yet implemented")
		default:
			return "", fmt.Errorf("unsupported template source type: %s", source.Type)
		}

		// Clone the repository
		err := git.Clone(git.CloneOptions{
			URL:       gitURL,
			Branch:    source.Reference,
			Directory: cachePath,
		})
		if err != nil {
			return "", fmt.Errorf("failed to clone template: %w", err)
		}

		fmt.Println("Template cloned successfully")
	}

	return cachePath, nil
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
