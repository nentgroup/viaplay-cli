// Package template provides template rendering functionality for viaplay-cli.
// It handles loading, processing, and rendering project templates with variable substitution.
package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/afero"

	"github.com/nentgroup/viaplay-cli/pkg/tmpl"
)

// Renderer handles template rendering
type Renderer struct {
	Variables  *Variables
	FileSystem afero.Fs
}

// NewRenderer creates a new template renderer with the given variables and filesystem
func NewRenderer(variables *Variables) *Renderer {
	if variables == nil {
		variables = NewTemplateVariables()
	}
	return &Renderer{
		Variables:  variables,
		FileSystem: afero.NewOsFs(),
	}
}

// RenderFile renders a file as a template and writes to destPath
func (r *Renderer) RenderFile(srcPath, destPath string, ensureDestDir bool) error {
	content, err := afero.ReadFile(r.FileSystem, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", srcPath, err)
	}

	rendered, err := r.RenderString(string(content))
	if err != nil {
		return fmt.Errorf("failed to render template for file %s: %w", srcPath, err)
	}

	if ensureDestDir {
		if err := r.FileSystem.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	if err := afero.WriteFile(r.FileSystem, destPath, []byte(rendered), 0o600); err != nil {
		return fmt.Errorf("failed to write rendered file %s: %w", destPath, err)
	}

	return nil
}

// RenderString processes a template string and returns the result
func (r *Renderer) RenderString(templateString string) (string, error) {
	// Protect GitHub Actions expressions from Go template parsing
	githubExprRe := regexp.MustCompile(`\$\{\{[^}]+}}`)
	templateStringSafe := githubExprRe.ReplaceAllStringFunc(templateString, func(expr string) string {
		return strings.ReplaceAll(strings.ReplaceAll(expr, "${{", "__GITHUB_EXPR_START__"), "}}", "__GITHUB_EXPR_END__")
	})

	if !strings.Contains(templateStringSafe, "{{") {
		// Restore GitHub Actions expressions before returning
		result := strings.ReplaceAll(templateStringSafe, "__GITHUB_EXPR_START__", "${{")
		result = strings.ReplaceAll(result, "__GITHUB_EXPR_END__", "}}")
		return result, nil
	}

	// Use tmpl.RenderWithLiteralUnknowns to handle missing variables gracefully
	rendered, err := tmpl.RenderWithLiteralUnknowns(templateStringSafe, r.Variables)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Restore GitHub Actions expressions before returning
	result := strings.ReplaceAll(rendered, "__GITHUB_EXPR_START__", "${{")
	result = strings.ReplaceAll(result, "__GITHUB_EXPR_END__", "}}")
	return result, nil
}

// RenderDirectoryPath processes a directory path as a template
func (r *Renderer) RenderDirectoryPath(path string) (string, error) {
	// Check if the path contains template delimiters
	if !strings.Contains(path, "{{") {
		return path, nil
	}

	// Create a new template for the path
	tmpl, err := template.New("path").
		Option("missingkey=invalid").
		Parse(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path as template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.Variables); err != nil {
		return "", fmt.Errorf("failed to execute path template: %w", err)
	}

	return buf.String(), nil
}
