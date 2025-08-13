// Package scaffolding handles project scaffolding from templates
package scaffolding

import (
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

// ProjectScaffolder handles applying templates to create project structures
type ProjectScaffolder struct {
	CacheManager *cache.Manager
	Config       *config.Configuration
}

// NewProjectScaffolder creates a new project scaffolder
func NewProjectScaffolder(cacheManager *cache.Manager, cfg *config.Configuration) *ProjectScaffolder {
	return &ProjectScaffolder{
		CacheManager: cacheManager,
		Config:       cfg,
	}
}

// ScaffoldProject creates a project structure from a template
// Parameters:
// - destPath: The path where the project should be created
// - language: The programming language (go, typescript, etc.)
// - projectType: The type of project (service, lambda, cli, etc.)
// - templateSource: The source of the template
// - variables: Map of template variables to replace in the project
// - forceUpdate: If true, forces update of the template cache
func (ps *ProjectScaffolder) ScaffoldProject(destPath, language, projectType, templateSource string, opts interface{}, skipHooks, forceUpdate bool) error {
	// Ensure the template is available in the cache
	var templatePath string
	var err error

	templatePath, err = ps.CacheManager.EnsureTemplate(language, projectType, templateSource, forceUpdate)
	if err != nil {
		return fmt.Errorf("failed to ensure template is available: %w", err)
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Assert opts to *template.Variables
	templateVars, ok := opts.(*templ.Variables)
	if !ok {
		return fmt.Errorf("opts must be of type *template.Variables")
	}

	renderer := templ.NewRenderer(templateVars)

	// Copy the template files to the destination with variable substitution
	if err := ps.copyTemplateFiles(templatePath, destPath, renderer); err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	return nil
}

// copyTemplateFiles copies files from the template directory to the destination
// with variable substitution using the template renderer
func (ps *ProjectScaffolder) copyTemplateFiles(templatePath, destPath string, renderer *templ.Renderer) error {
	// List of directories to skip
	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".idea":        true,
	}

	// List of files to skip
	skipFiles := map[string]bool{
		".DS_Store": true,
		"Thumbs.db": true,
		".env":      true, // Skip actual .env files (but allow .env.example)
		".npmrc":    true, // Skip actual .npmrc files with tokens
		".yarnrc":   true, // Skip actual .yarnrc files with tokens
	}

	// Walk through the template directory
	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		baseName := filepath.Base(path)

		// Skip directories in the skipDirs list
		if info.IsDir() && skipDirs[baseName] {
			return filepath.SkipDir
		}

		// Skip files in the skipFiles list
		if !info.IsDir() && skipFiles[baseName] {
			return nil
		}

		// Compute the relative path from the template root
		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}

		// Skip if it's the root directory
		if relPath == "." {
			return nil
		}

		destRelPath, err := renderer.RenderDirectoryPath(relPath)
		if err != nil {
			return fmt.Errorf("failed to parse path as template: %s: %w", relPath, err)
		}

		destFilePath := filepath.Join(destPath, destRelPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destFilePath, 0o755)
		}

		isBinary, err := isBinaryFile(path)
		if err != nil {
			return fmt.Errorf("failed to check if file is binary: %w", err)
		}

		if isBinary {
			// Use helper to copy binary file
			if err := copyBinaryFile(path, destFilePath, info.Mode()); err != nil {
				return err
			}
		} else {
			// Always render non-binary files as templates
			if err := renderer.RenderFile(path, destFilePath, true); err != nil {
				return fmt.Errorf("failed to render template file %s: %w", relPath, err)
			}
		}

		// Copy file mode from the template file to preserve executability
		if err := os.Chmod(destFilePath, info.Mode()); err != nil {
			fmt.Printf("Warning: Failed to set file mode for %s: %v\n", destFilePath, err)
		}

		return nil
	})
}

// copyBinaryFile copies a binary file from srcPath to destPath, preserving file mode
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

// isBinaryFile checks if a file is likely binary by examining its content
// It uses a common heuristic: if a file contains NUL bytes or a high proportion
// of non-printable characters, it's likely binary
func isBinaryFile(path string) (bool, error) {
	binaryExtensions := map[string]bool{
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".ico":   true,
		".pdf":   true,
		".zip":   true,
		".tar":   true,
		".gz":    true,
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
		".eot":   true,
		".otf":   true,
		".svg":   true,
		".mp3":   true,
		".mp4":   true,
		".avi":   true,
		".mov":   true,
		".webm":  true,
		".webp":  true,
		".doc":   true,
		".docx":  true,
		".xls":   true,
		".xlsx":  true,
		".ppt":   true,
		".pptx":  true,
	}

	ext := strings.ToLower(filepath.Ext(path))
	if binaryExtensions[ext] {
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
	buf = buf[:n]

	nullCount := 0
	nonPrintableCount := 0
	for _, b := range buf {
		if b == 0 {
			nullCount++
		} else if b < 32 && !isAllowedNonPrintable(b) {
			nonPrintableCount++
		}
	}

	if nullCount > 0 {
		return true, nil
	}

	if n > 0 && float64(nonPrintableCount)/float64(n) > 0.3 {
		return true, nil
	}

	return false, nil
}

func isAllowedNonPrintable(b byte) bool {
	switch b {
	case '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}
