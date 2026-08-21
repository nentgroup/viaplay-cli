package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/template"
)

// RunHooks runs the post-installation hooks for a project
func (c *Factory) RunHooks(ctx context.Context, projectPath, language, projectType string,
	templateVars *template.Variables,
) error {
	// Create a function that will run the hooks and write output to provided writers
	runHookFn := func(stdout, stderr io.Writer) error {
		// Create command executors that use the provided writers
		cmdExecutor := func(cmd *exec.Cmd) error {
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			return cmd.Run()
		}

		// Get renderer for template variables
		renderer := template.NewRenderer(templateVars)

		// Check if we have hooks for this language and project type
		hooks := c.Config.GetPostInstallHooks(language, projectType)
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
				execCmd := exec.CommandContext(ctx, "sh", "-c", renderedCmd)
				execCmd.Dir = projectPath

				// Run the command using our executor
				if err := cmdExecutor(execCmd); err != nil {
					return fmt.Errorf("hook command failed: %w", err)
				}
			}

			// Process scripts
			for _, scriptPath := range hook.GetAllScripts() {
				fullScriptPath, err := c.resolveRenderedHookScriptPath(renderer, scriptPath)
				if err != nil {
					return err
				}

				renderedScriptPath, cleanup, err := renderHookScriptFile(renderer, fullScriptPath)
				if err != nil {
					return err
				}
				defer cleanup()

				// Create a command to run the rendered script
				execCmd := exec.CommandContext(ctx, renderedScriptPath)
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
	err := output.DisplayHookOutput(title, runHookFn)
	// Display a simple message based on the result
	if err != nil {
		fmt.Printf("Hooks failed: %v\n", err)
	} else {
		fmt.Printf("Post-installation hooks completed successfully\n")
	}

	return err
}

func (c *Factory) resolveRenderedHookScriptPath(renderer *template.Renderer, scriptPath string) (string, error) {
	renderedScriptPath, err := renderer.RenderString(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to render script path template: %w", err)
	}

	fullScriptPath := c.Config.ResolveHookScriptPath(renderedScriptPath)
	if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("hook script not found: %s", fullScriptPath)
	}

	return fullScriptPath, nil
}

func renderHookScriptFile(renderer *template.Renderer, scriptPath string) (string, func(), error) {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to read hook script %s: %w", scriptPath, err)
	}

	renderedData, err := renderer.RenderString(string(data))
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to render hook script %s: %w", scriptPath, err)
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to stat hook script %s: %w", scriptPath, err)
	}

	tempFile, err := os.CreateTemp("", "vip-hook-*"+filepath.Ext(scriptPath))
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary hook script: %w", err)
	}

	cleanup := func() {
		_ = os.Remove(tempFile.Name())
	}

	if _, err := tempFile.WriteString(renderedData); err != nil {
		_ = tempFile.Close()
		cleanup()
		return "", func() {}, fmt.Errorf("failed to write rendered hook script: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to close rendered hook script: %w", err)
	}

	if err := os.Chmod(tempFile.Name(), info.Mode().Perm()); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to apply permissions to rendered hook script: %w", err)
	}

	return tempFile.Name(), cleanup, nil
}
