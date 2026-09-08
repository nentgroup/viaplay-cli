// Package cache provides template caching functionality.
package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/git"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/pkg/paths"
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
	// Version of the template (derived from git tags or commit hash)
	Version string
	// RemoteURL of the template repository (e.g. git@github.com:org/repo.git or https URL)
	RemoteURL string
}

// TemplateInfo represents template information for display
type TemplateInfo struct {
	Language  string
	Type      string
	Path      string
	LastUsed  time.Time
	Version   string
	RemoteURL string
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

// ParseSource parses a template source string into a Source struct
// Format:
// - GitHub repo: "github@<owner>/<repo>.git[@branch/tag]"
// - Local path: "local@/path/to/template"
// - Tarball URL: "url@https://example.com/template.tar.gz"
// - Git SSH URL: "git@github.com:<owner>/<repo>.git" or "git@github.com/<owner>/<repo>.git"
//
// For convenience, a few common shorthand forms are also accepted and expanded
// before parsing (see normalizeSourceShorthand): plain GitHub URLs/addresses
// (e.g. "https://github.com/owner/repo" or "github.com/owner/repo"), and bare
// strings with no recognised "type@" prefix, which are treated as local paths.
func ParseSource(sourceStr string) (Source, error) {
	sourceStr = normalizeSourceShorthand(sourceStr)

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
			location = paths.Expand(location)
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
		// A colon in the location means it's already a complete SSH remote for
		// some host (e.g. "git@bitbucket.org:owner/repo.git"), so it can be
		// used verbatim as the clone URL.
		if strings.Contains(location, ":") {
			return Source{
				Type:     SourceTypeSSH,
				Location: sourceStr,
			}, nil
		}
		// Otherwise, this is the "git@owner/repo[@ref]" shorthand documented
		// alongside local@/url@ (no host given), matching the convenience
		// "github@owner/repo" syntax. Expand it to a full GitHub SSH URL.
		if idx := strings.LastIndex(location, "@"); idx != -1 {
			reference = location[idx+1:]
			location = location[:idx]
		}
		location = strings.TrimSuffix(location, ".git")
		return Source{
			Type:      SourceTypeSSH,
			Location:  "git@github.com:" + location + ".git",
			Reference: reference,
		}, nil
	default:
		return Source{}, fmt.Errorf("unknown template source type: %s", sourceType)
	}
}

// normalizeSourceShorthand expands convenient shorthand forms of a template
// source into the explicit "type@value" syntax required by the rest of
// ParseSource, so users don't need to memorise or hand-construct the
// "github@owner/repo" syntax:
//   - A GitHub repository address copy-pasted from a browser or git remote
//     (e.g. "https://github.com/owner/repo", "http://github.com/owner/repo" or
//     the bare "github.com/owner/repo", each optionally suffixed with ".git"
//     and/or "@branch-or-tag") is expanded to "github@owner/repo[@ref]".
//   - Any other string without a recognised "type@" prefix is treated as a
//     local path, matching the CLI's historical default behaviour.
func normalizeSourceShorthand(sourceStr string) string {
	trimmed := strings.TrimSpace(sourceStr)
	if trimmed == "" {
		return trimmed
	}

	for _, prefix := range []string{"https://github.com/", "http://github.com/", "github.com/"} {
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		rest := strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), "/")
		rest = strings.TrimSuffix(rest, ".git")
		return "github@" + rest
	}

	if !strings.Contains(trimmed, "@") {
		return "local@" + trimmed
	}

	return trimmed
}

// GetTemplatePath returns the cache path for a specific template
func (m *Manager) GetTemplatePath(language, templateType string) string {
	return filepath.Join(m.BaseCacheDir, language, templateType)
}

// EnsureEphemeralTemplate resolves an ad-hoc template source (one not tied to
// a configured language/type, as used by `template inspect` and
// `template test --template-path`) for one-off, read-only use. Unlike
// EnsureTemplate, it never reads from or writes to the persistent template
// cache: local sources are returned as-is, and remote sources are cloned into
// a fresh temporary directory on every call. This guarantees inspecting or
// test-scaffolding a template always reflects its current state and never
// mutates cache state shared with real project creation.
//
// The returned cleanup function removes any temporary clone and must be
// called once the caller is done reading the template; it is a no-op for
// local sources.
func (m *Manager) EnsureEphemeralTemplate(ctx context.Context, sourceStr string) (path string, cleanup func(), err error) {
	noopCleanup := func() {}

	source, err := ParseSource(sourceStr)
	if err != nil {
		return "", noopCleanup, fmt.Errorf("failed to parse template source: %w", err)
	}

	// Local templates are used directly from their existing path; nothing to
	// clone or clean up.
	if source.Type == SourceTypeLocal {
		path, err := handleLocalTemplate(source)
		return path, noopCleanup, err
	}

	tempDir, err := os.MkdirTemp("", "vip-template-source-*")
	if err != nil {
		return "", noopCleanup, fmt.Errorf("failed to create temporary directory for template source: %w", err)
	}
	cleanup = func() {
		if err := os.RemoveAll(tempDir); err != nil {
			output.VerboseMessage(fmt.Sprintf("failed to clean up temporary template clone at %s: %v", tempDir, err))
		}
	}

	if _, err := handleNewTemplate(ctx, tempDir, source); err != nil {
		cleanup()
		return "", noopCleanup, err
	}

	return tempDir, cleanup, nil
}

// EnsureTemplate ensures a template is available in the cache
func (m *Manager) EnsureTemplate(ctx context.Context, language, templateType, sourceStr string,
	forceUpdate bool,
) (string, error) {
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

	// Check if we should force an update

	// If template exists and no force update requested, simply use the cached version
	if exists && !forceUpdate {
		output.VerboseMessage("Using cached template (use --no-cache to check for updates)")
		return cachePath, nil
	}

	// If force update is requested and template exists, remove it first
	if forceUpdate {
		output.VerboseMessage(fmt.Sprintf("Force update requested, removing existing template at: %s", cachePath))
		if err := os.RemoveAll(cachePath); err != nil {
			return "", fmt.Errorf("failed to remove existing template for force update: %w", err)
		}
	}

	// At this point, either the template doesn't exist or we removed it for a force update
	return handleNewTemplate(ctx, cachePath, source)
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

// handleNewTemplate handles cloning a new template
func handleNewTemplate(ctx context.Context, cachePath string, source Source) (string, error) {
	output.VerboseMessage(fmt.Sprintf("Template not found in cache, cloning to: %s", cachePath))

	// Prepare git URL based on source type
	gitURL, err := getGitURLFromSource(source)
	if err != nil {
		return "", err
	}

	// Clone the repository
	if err := git.Clone(ctx, git.CloneOptions{
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
func (m *Manager) ListTemplates(ctx context.Context) ([]Template, error) {
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
			if tmpl, ok := m.buildTemplateFromDir(ctx, path, info, parts); ok {
				templates = append(templates, tmpl)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing templates: %w", err)
	}

	return templates, nil
}

// buildTemplateFromDir constructs a Template from a given directory path if it is a valid git repo.
func (m *Manager) buildTemplateFromDir(ctx context.Context, path string, info os.FileInfo, parts []string) (Template, bool) {
	// Expect parts to be [language, type]
	if len(parts) != 2 || !info.IsDir() {
		return Template{}, false
	}

	// Check if it's a git repository
	if !git.IsGitRepository(path) {
		return Template{}, false
	}

	// Try to get the source information from the git repository
	source, err := getTemplateSourceFromGit(ctx, path)
	if err != nil {
		// If we can't determine the source, use a default
		source = Source{
			Type:     SourceTypeGitHub,
			Location: unknownVersion,
		}
	}

	// Derive version information from git
	version, err := git.DescribeVersion(ctx, path)
	if err != nil {
		version = unknownVersion
	}

	// Get repository info to extract remote URL
	repoInfo, err := git.GetRepositoryInfo(ctx, path)
	remoteURL := unknownVersion
	if err == nil && repoInfo != nil && repoInfo.RemoteURL != "" {
		remoteURL = repoInfo.RemoteURL

		// Normalise GitHub SSH URLs (git@github.com:org/repo.git) to HTTPS
		if strings.HasPrefix(remoteURL, "git@github.com:") {
			// Strip git@github.com: and optional .git suffix
			repoPath := strings.TrimPrefix(remoteURL, "git@github.com:")
			repoPath = strings.TrimSuffix(repoPath, ".git")
			remoteURL = "https://github.com/" + repoPath
		}
	}

	return Template{
		Language:     parts[0],
		Type:         parts[1],
		Path:         path,
		LastModified: info.ModTime(),
		Source:       source,
		Version:      version,
		RemoteURL:    remoteURL,
	}, true
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

		if info.IsDir() && !strings.Contains(rel, string(os.PathSeparator)) {
			if err := pruneTemplatesInLanguageDir(path, cutoffTime, &totalTemplates, &prunedCount); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("error while pruning cache: %w", err)
	}

	return prunedCount, totalTemplates, nil
}

// pruneTemplatesInLanguageDir inspects template-type subdirectories under a language directory
// and removes those older than cutoffTime, updating the provided counters.
func pruneTemplatesInLanguageDir(languagePath string, cutoffTime time.Time, totalTemplates, prunedCount *int) error {
	templateTypes, err := os.ReadDir(languagePath)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", languagePath, err)
	}

	for _, templateType := range templateTypes {
		if !templateType.IsDir() {
			continue
		}

		*totalTemplates++
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
			*prunedCount++
		}
	}

	return nil
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
func (m *Manager) UpdateAllTemplates(ctx context.Context) (int, int, error) {
	templates, err := m.ListTemplates(ctx)
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
			_ = os.Setenv("VIAPLAY_CLI_FORCE_UPDATE", "true")
			defer func() {
				_ = os.Unsetenv("VIAPLAY_CLI_FORCE_UPDATE")
			}()

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
			_, err := m.EnsureTemplate(ctx, t.Language, t.Type, sourceStr,
				true) // Force update when explicitly updating templates
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
func getTemplateSourceFromGit(ctx context.Context, repoPath string) (Source, error) {
	// Get repository info
	repoInfo, err := git.GetRepositoryInfo(ctx, repoPath)
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
