// Package scaffolding handles project scaffolding from templates
package scaffolding

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
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
func (ps *ProjectScaffolder) ScaffoldProject(destPath, language, projectType, templateSource string, opts interface{}, skipHooks bool, forceUpdate bool) error {
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
func (ps *ProjectScaffolder) runPostScaffoldCommands(projectPath, language, projectType string, templateVars *templ.Variables) error {
	// Get renderer for template variables
	renderer := templ.NewRenderer(templateVars)

	// Execute hooks from configuration
	if ps.Config != nil {
		if err := ps.runConfiguredHooks(projectPath, language, projectType, renderer); err != nil {
			return fmt.Errorf("failed to run configured hooks: %w", err)
		}
	}

	return nil
}

// runConfiguredHooks executes hooks defined in the configuration
func (ps *ProjectScaffolder) runConfiguredHooks(projectPath, language, projectType string, renderer *templ.Renderer) error {
	// Check if we have hooks for this language and project type
	hooks := ps.Config.GetPostInstallHooks(language, projectType)
	if len(hooks) == 0 {
		// No hooks configured
		return nil
	}

	// Execute hooks in order (general -> language-specific -> project-type-specific)
	for _, hook := range hooks {
		// Process commands
		for _, cmd := range hook.GetAllCommands() {
			// Render template variables in the command
			renderedCmd, err := renderer.RenderString(cmd)
			if err != nil {
				return fmt.Errorf("failed to render run command template: %w", err)
			}

			// Create a command that will run in the project directory
			execCmd := exec.Command("sh", "-c", renderedCmd)
			execCmd.Dir = projectPath
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			// Run the command
			if err := execCmd.Run(); err != nil {
				return fmt.Errorf("hook command failed: %w", err)
			}
		}

		// Process scripts
		for _, scriptPath := range hook.GetAllScripts() {
			// Render template variables in the script path
			renderedScriptPath, err := renderer.RenderString(scriptPath)
			if err != nil {
				return fmt.Errorf("failed to render script path template: %w", err)
			}

			// Check if this is a relative path or absolute
			fullScriptPath := renderedScriptPath
			if !filepath.IsAbs(renderedScriptPath) {
				// If it's relative, look in the hooks directory
				fullScriptPath = filepath.Join(ps.Config.GetHooksDir(), renderedScriptPath)
			}

			// Check if script exists
			if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
				return fmt.Errorf("hook script not found: %s", fullScriptPath)
			}

			// Create a command to run the script
			execCmd := exec.Command(fullScriptPath)
			execCmd.Dir = projectPath
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr

			// Run the script
			if err := execCmd.Run(); err != nil {
				return fmt.Errorf("hook script failed: %w", err)
			}
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

// RunPostInstallHooks runs the post-installation hooks for a project
// This is separated from ScaffoldProject to allow running hooks after repository creation
func (ps *ProjectScaffolder) RunPostInstallHooks(projectPath, language, projectType string, templateVars *templ.Variables) error {
	// Create a function that will run the hooks and write output to provided writers
	runHookFn := func(stdout, stderr io.Writer) error {
		// Create command executors that use the provided writers
		cmdExecutor := func(cmd *exec.Cmd) error {
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			return cmd.Run()
		}

		// Get renderer for template variables
		renderer := templ.NewRenderer(templateVars)

		// Check if we have hooks for this language and project type
		hooks := ps.Config.GetPostInstallHooks(language, projectType)
		if len(hooks) == 0 {
			fmt.Fprintf(stdout, "No hooks configured for %s/%s\n", language, projectType)
			return nil
		}

		// Execute hooks in order (general -> language-specific -> project-type-specific)
		fmt.Println("----------------------------------------")
		for _, hook := range hooks {
			// Process commands
			for _, cmd := range hook.GetAllCommands() {
				// Render template variables in the command
				renderedCmd, err := renderer.RenderString(cmd)
				if err != nil {
					return fmt.Errorf("failed to render run command template: %w", err)
				}

				// Create a command that will run in the project directory
				execCmd := exec.Command("sh", "-c", renderedCmd)
				execCmd.Dir = projectPath

				// Run the command using our executor
				if err := cmdExecutor(execCmd); err != nil {
					return fmt.Errorf("hook command failed: %w", err)
				}
			}

			// Process scripts
			for _, scriptPath := range hook.GetAllScripts() {
				// Render template variables in the script path
				renderedScriptPath, err := renderer.RenderString(scriptPath)
				if err != nil {
					return fmt.Errorf("failed to render script path template: %w", err)
				}

				// Check if this is a relative path or absolute
				fullScriptPath := renderedScriptPath
				if !filepath.IsAbs(renderedScriptPath) {
					// If it's relative, look in the hooks directory
					fullScriptPath = filepath.Join(ps.Config.GetHooksDir(), renderedScriptPath)
				}
				// Check if script exists
				if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
					return fmt.Errorf("hook script not found: %s", fullScriptPath)
				}

				// Create a command to run the script
				execCmd := exec.Command(fullScriptPath)
				execCmd.Dir = projectPath

				// Run the script using our executor
				if err := cmdExecutor(execCmd); err != nil {
					return fmt.Errorf("hook script failed: %w", err)
				}
			}
		}
		fmt.Println("----------------------------------------")
		return nil
	}

	// Create a title for the TUI
	title := fmt.Sprintf("Post-Installation Hooks for %s/%s", language, projectType)

	// Display the hook output using our simplified UI
	err := DisplayHookOutput(title, runHookFn)
	// Display a simple message based on the result
	if err != nil {
		fmt.Printf("Hooks failed: %v\n", err)
	} else {
		fmt.Printf("Post-installation hooks completed successfully\n")
	}

	return err
}
