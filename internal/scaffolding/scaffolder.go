// Package scaffolding handles project scaffolding from templates
package scaffolding

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/git"
	templ "github.com/nentgroup/viaplay-cli/internal/template"
)

// Template delimiter constants
const (
	TemplateDelimLeft  = "{{{"
	TemplateDelimRight = "}}}"
)

// ProjectScaffolder handles applying templates to create project structures
type ProjectScaffolder struct {
	CacheManager *cache.Manager
}

// NewProjectScaffolder creates a new project scaffolder
func NewProjectScaffolder(cacheManager *cache.Manager) *ProjectScaffolder {
	return &ProjectScaffolder{
		CacheManager: cacheManager,
	}
}

// ScaffoldProject creates a project structure from a template
// Parameters:
// - destPath: The path where the project should be created
// - language: The programming language (go, typescript, etc.)
// - projectType: The type of project (service, lambda, cli, etc.)
// - templateSource: The source of the template
// - variables: Map of template variables to replace in the project
func (ps *ProjectScaffolder) ScaffoldProject(destPath, language, projectType, templateSource string, variables map[string]string) error {
	// Ensure the template is available in the cache
	templatePath, err := ps.CacheManager.EnsureTemplate(language, projectType, templateSource)
	if err != nil {
		return fmt.Errorf("failed to ensure template is available: %w", err)
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Convert simple variables map to structured template variables
	templateVars := templ.NewTemplateVariables().FromMap(variables)

	// Create a renderer
	renderer := templ.NewRenderer(templateVars)

	// Copy the template files to the destination with variable substitution
	if err := ps.copyTemplateFiles(templatePath, destPath, renderer); err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	// Run any post-scaffolding commands
	if err := ps.runPostScaffoldCommands(destPath, language, projectType); err != nil {
		return fmt.Errorf("failed to run post-scaffold commands: %w", err)
	}

	return nil
}

// ScaffoldProjectWithOptions creates a project structure from a template using a strongly-typed options struct
func (ps *ProjectScaffolder) ScaffoldProjectWithOptions(destPath, language, projectType, templateSource string, opts interface{}) error {
	// Ensure the template is available in the cache
	templatePath, err := ps.CacheManager.EnsureTemplate(language, projectType, templateSource)
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

	// Run any post-scaffolding commands
	if err := ps.runPostScaffoldCommands(destPath, language, projectType); err != nil {
		return fmt.Errorf("failed to run post-scaffold commands: %w", err)
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

		// Process the file path itself as a template if it contains template markers
		destRelPath := relPath
		if strings.Contains(relPath, TemplateDelimLeft) {
			processedPath, err := renderer.RenderString(relPath)
			if err != nil {
				return fmt.Errorf("failed to parse path as template: %s: %w", relPath, err)
			}
			destRelPath = processedPath
		}

		destFilePath := filepath.Join(destPath, destRelPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destFilePath, 0o755)
		}

		// Process the file with the renderer
		// The renderer will handle binary detection and processing appropriately
		if err := renderer.RenderFile(path, destFilePath, true); err != nil {
			return fmt.Errorf("failed to render file %s: %w", relPath, err)
		}

		// Copy file mode from the template file to preserve executability
		if err := os.Chmod(destFilePath, info.Mode()); err != nil {
			fmt.Printf("Warning: Failed to set file mode for %s: %v\n", destFilePath, err)
		}

		return nil
	})
}

// runPostScaffoldCommands runs any post-scaffold commands for the template
func (ps *ProjectScaffolder) runPostScaffoldCommands(projectPath, language, projectType string) error {
	// Check for post-scaffold script
	scriptPath := filepath.Join(projectPath, ".post-scaffold.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		// Make the script executable
		if err := os.Chmod(scriptPath, 0o755); err != nil {
			return fmt.Errorf("failed to make post-scaffold script executable: %w", err)
		}

		// Run the script
		cmd := exec.Command(scriptPath)
		cmd.Dir = projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run post-scaffold script: %w", err)
		}

		// Remove the script after running
		if err := os.Remove(scriptPath); err != nil {
			fmt.Printf("Warning: Failed to remove post-scaffold script: %v\n", err)
		}
	}

	// Language-specific initialization
	switch language {
	case "go":
		if err := ps.initGoProject(projectPath); err != nil {
			return err
		}
	case "typescript", "javascript":
		if err := ps.initNodeProject(projectPath); err != nil {
			return err
		}
	}

	return nil
}

// initGoProject initialises a Go project with proper module setup
func (ps *ProjectScaffolder) initGoProject(projectPath string) error {
	// Check if go.mod already exists
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
		// Module already initialised
		return nil
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// initNodeProject initialises a Node.js project
func (ps *ProjectScaffolder) initNodeProject(projectPath string) error {
	// Check if package.json exists and node_modules doesn't
	packageJSONPath := filepath.Join(projectPath, "package.json")
	nodeModulesPath := filepath.Join(projectPath, "node_modules")

	if _, err := os.Stat(packageJSONPath); err == nil {
		if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
			// Run npm install
			cmd := exec.Command("npm", "install")
			cmd.Dir = projectPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}

	return nil
}

// CloneToRepo initialises a git repository in the project directory and pushes it to the remote
func (ps *ProjectScaffolder) CloneToRepo(projectPath, repoURL string) error {
	// Initialise git repository
	if err := git.InitRepository(projectPath); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add all files
	if err := git.CommitAll(projectPath, "Initial commit from viaplay-cli template"); err != nil {
		return fmt.Errorf("failed to commit files: %w", err)
	}

	// Add remote
	if err := git.AddRemote(projectPath, "origin", repoURL); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Push to remote
	branch := "main" // Default to main branch
	if err := git.Push(projectPath, "origin", branch); err != nil {
		// If pushing to main fails, try master
		fmt.Println("Push to 'main' failed, trying 'master' branch...")
		if err := git.Push(projectPath, "origin", "master"); err != nil {
			return fmt.Errorf("failed to push to remote: %w", err)
		}
	}

	return nil
}
