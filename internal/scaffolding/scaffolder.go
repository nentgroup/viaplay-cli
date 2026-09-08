package scaffolding

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	templ "github.com/nentgroup/viaplay-cli/internal/template"
)

type ProjectScaffolder struct {
	CacheManager *cache.Manager
	Config       *config.Configuration
}

func NewProjectScaffolder(cacheManager *cache.Manager, cfg *config.Configuration) *ProjectScaffolder {
	return &ProjectScaffolder{CacheManager: cacheManager, Config: cfg}
}

// ScaffoldProject applies a template using default (interactive) manifest option resolution.
// Use ScaffoldProjectWithOptions to control --set overrides and --no-input behaviour.
func (ps *ProjectScaffolder) ScaffoldProject(ctx context.Context, destPath, language, projectType, templateSource string, opts interface{}, skipHooks, forceUpdate bool) error {
	return ps.ScaffoldProjectWithOptions(ctx, destPath, language, projectType, templateSource, opts, skipHooks, forceUpdate, nil, false)
}

// ScaffoldProjectWithOptions applies a template, resolving any manifest options via --set
// overrides (templateSet) and/or interactive prompts (unless noInput is true).
func (ps *ProjectScaffolder) ScaffoldProjectWithOptions(ctx context.Context, destPath, language, projectType, templateSource string,
	opts interface{}, skipHooks, forceUpdate bool, templateSet []string, noInput bool,
) error {
	templatePath, manifest, cleanup, err := ps.resolveTemplateManifest(ctx, language, projectType, templateSource, forceUpdate)
	if err != nil {
		return err
	}
	defer cleanup()

	templateVars, ok := opts.(*templ.Variables)
	if !ok {
		return fmt.Errorf("opts must be of type *template.Variables")
	}

	// Resolve manifest options/prompts before touching the filesystem, so a
	// missing required option (e.g. under --no-input) fails fast without
	// leaving behind an empty destination directory that would need cleanup.
	if manifest != nil {
		if err := templ.ResolveManifestSelections(manifest, templateVars, templateSet, noInput); err != nil {
			return fmt.Errorf("failed to resolve template options: %w", err)
		}
	}

	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	renderer := templ.NewRenderer(templateVars)
	return ps.copyTemplateFiles(templatePath, destPath, renderer, manifest)
}

// GetTemplateManifest ensures the template is available locally and loads its manifest,
// if any, without scaffolding a project. Returns a nil manifest (and nil error) when the
// template has no template.yaml, preserving backward compatibility for plain templates.
func (ps *ProjectScaffolder) GetTemplateManifest(ctx context.Context, language, projectType, templateSource string, forceUpdate bool) (*templ.Manifest, error) {
	_, manifest, cleanup, err := ps.resolveTemplateManifest(ctx, language, projectType, templateSource, forceUpdate)
	defer cleanup()
	return manifest, err
}

// resolveTemplateManifest ensures the template is locally available and loads its
// manifest (template.yaml), checking both the template root and the "_template"
// subdirectory since manifests live at the repository root, which may differ from
// the rendered content root.
//
// Ad-hoc sources (language and projectType both empty, as used by
// `template inspect` and `template test --template-path`) are resolved via
// EnsureEphemeralTemplate: they're read-only, one-off lookups and must never
// read from or write to the persistent template cache shared with real
// project creation. Configured language/type sources continue to use the
// persistent, reusable cache via EnsureTemplate.
//
// The returned cleanup function removes any temporary clone created for an
// ad-hoc source and must always be called by the caller once done with the
// template (it is a no-op otherwise).
func (ps *ProjectScaffolder) resolveTemplateManifest(ctx context.Context, language, projectType, templateSource string, forceUpdate bool) (string, *templ.Manifest, func(), error) {
	noopCleanup := func() {}

	var (
		templatePath string
		cleanup      func()
		err          error
	)
	if language == "" && projectType == "" {
		templatePath, cleanup, err = ps.CacheManager.EnsureEphemeralTemplate(ctx, templateSource)
	} else {
		cleanup = noopCleanup
		templatePath, err = ps.CacheManager.EnsureTemplate(ctx, language, projectType, templateSource, forceUpdate)
	}
	if err != nil {
		return "", nil, noopCleanup, fmt.Errorf("failed to ensure template is available: %w", err)
	}

	manifestRoot := templatePath
	if _, err := os.Stat(filepath.Join(manifestRoot, "template.yaml")); err != nil {
		if _, err := os.Stat(filepath.Join(manifestRoot, "_template", "template.yaml")); err == nil {
			manifestRoot = filepath.Join(manifestRoot, "_template")
		}
	}
	manifest, err := templ.LoadManifest(manifestRoot)
	if err != nil {
		cleanup()
		return "", nil, noopCleanup, err
	}
	return templatePath, manifest, cleanup, nil
}

func (ps *ProjectScaffolder) copyTemplateFiles(templatePath, destPath string, renderer *templ.Renderer, manifest *templ.Manifest) error {
	specialTemplateDir := filepath.Join(templatePath, "_template")
	if stat, err := os.Stat(specialTemplateDir); err == nil && stat.IsDir() {
		templatePath = specialTemplateDir
	}

	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return ps.copyTemplateEntry(templatePath, destPath, path, info, renderer, manifest)
	})
}

// copyTemplateEntry processes a single file/directory entry found while walking a
// template's source tree: it applies skip rules (well-known dirs/files, manifest
// gating), renders the destination path, and copies/renders the file contents.
func (ps *ProjectScaffolder) copyTemplateEntry(templatePath, destPath, path string, info os.FileInfo, renderer *templ.Renderer, manifest *templ.Manifest) error {
	skipDirs := map[string]bool{".git": true, "node_modules": true, "vendor": true, "dist": true, "build": true, ".idea": true}
	skipFiles := map[string]bool{".DS_Store": true, "Thumbs.db": true, ".env": true, ".npmrc": true, ".yarnrc": true}

	relPath, err := filepath.Rel(templatePath, path)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}
	if relPath == "." {
		return nil
	}
	baseName := filepath.Base(path)
	if info.IsDir() && skipDirs[baseName] {
		return filepath.SkipDir
	}
	if !info.IsDir() && skipFiles[baseName] {
		return nil
	}
	if manifest != nil && shouldSkipPath(relPath, info.IsDir(), manifest, renderer.Variables) {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}
	destRelPath, err := renderer.RenderDirectoryPath(relPath)
	if err != nil {
		return fmt.Errorf("failed to parse path as template: %s: %w", relPath, err)
	}
	if strings.HasSuffix(destRelPath, ".env.example") {
		destRelPath = strings.TrimSuffix(destRelPath, ".example")
	}
	destFilePath := filepath.Join(destPath, destRelPath)
	if info.IsDir() {
		return os.MkdirAll(destFilePath, 0o755)
	}
	destFilePath, err = ps.copyOrRenderFile(path, destFilePath, relPath, baseName, info, renderer)
	if err != nil {
		return err
	}
	if err := os.Chmod(destFilePath, info.Mode()); err != nil {
		fmt.Printf("Warning: Failed to set file mode for %s: %v\n", destFilePath, err)
	}
	return nil
}

// copyOrRenderFile copies a file as-is (binary or ".raw"-suffixed) or renders it as a
// Go template, writing the result to destFilePath. It returns the actual path the
// file was written to (which may differ from destFilePath when a ".raw" suffix is
// trimmed) so the caller can apply the correct file mode.
func (ps *ProjectScaffolder) copyOrRenderFile(path, destFilePath, relPath, baseName string, info os.FileInfo, renderer *templ.Renderer) (string, error) {
	isBinary, err := isBinaryFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to check if file is binary: %w", err)
	}
	switch {
	case isBinary:
		return destFilePath, copyBinaryFile(path, destFilePath, info.Mode())
	case strings.HasSuffix(baseName, ".raw"):
		destFilePath = strings.TrimSuffix(destFilePath, ".raw")
		return destFilePath, copyBinaryFile(path, destFilePath, info.Mode())
	default:
		if err := renderer.RenderFile(path, destFilePath, true); err != nil {
			return "", fmt.Errorf("failed to render template file %s: %w", relPath, err)
		}
		return destFilePath, nil
	}
}

// shouldSkipPath decides whether a given relative path should be skipped during
// scaffolding based on the manifest's file rules. Semantics:
//   - exclude rules skip a matching path when their condition evaluates to true.
//   - include rules gate a matching path: skip it when their condition evaluates to false.
//   - paths that don't match any rule are never skipped (rules only affect the
//     specific paths they target, they are not a whitelist for the whole tree).
//
// Directory entries are matched against a rule if the directory is (or is an
// ancestor of) something the rule's glob pattern would match, e.g. a rule
// targeting "internal/storage/dynamo/*" also matches the directory
// "internal/storage/dynamo" itself — otherwise the directory would still get
// created (via os.MkdirAll further down the walk) and left empty even though
// every file inside it was correctly skipped.
func shouldSkipPath(relPath string, isDir bool, manifest *templ.Manifest, vars *templ.Variables) bool {
	matches := func(pattern string) bool {
		if isDir {
			return dirMatchesPattern(pattern, relPath)
		}
		match, err := filepath.Match(pattern, relPath)
		return err == nil && match
	}
	for _, rule := range manifest.Files.Exclude {
		if matches(rule.Path) {
			ok, err := templ.EvalCondition(rule.When, vars)
			if err == nil && ok {
				return true
			}
		}
	}
	for _, rule := range manifest.Files.Include {
		if matches(rule.Path) {
			ok, err := templ.EvalCondition(rule.When, vars)
			if err == nil && !ok {
				return true
			}
		}
	}
	return false
}

// dirMatchesPattern reports whether relPath (a directory) is the immediate
// parent directory targeted by pattern's final wildcard segment, or matches
// the pattern's full depth outright, by comparing path segments
// component-by-component with filepath.Match. This lets a file-glob rule
// like "internal/storage/dynamo/*" also gate the "internal/storage/dynamo"
// directory entry itself — but not shallower ancestors like
// "internal/storage" or "internal", which may hold sibling entries governed
// by different (or no) rules.
func dirMatchesPattern(pattern, relPath string) bool {
	patternParts := strings.Split(pattern, "/")
	relParts := strings.Split(relPath, "/")
	if len(relParts) != len(patternParts)-1 && len(relParts) != len(patternParts) {
		return false
	}
	for i, part := range relParts {
		match, err := filepath.Match(patternParts[i], part)
		if err != nil || !match {
			return false
		}
	}
	return true
}

func copyBinaryFile(srcPath, destPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open binary file: %w", err)
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy binary file: %w", err)
	}
	return nil
}

func isBinaryFile(path string) (bool, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".exe", ".dll", ".so", ".dylib", ".woff", ".woff2", ".ttf", ".eot", ".otf", ".svg", ".mp3", ".mp4", ".avi", ".mov", ".webm", ".webp", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return true, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}
