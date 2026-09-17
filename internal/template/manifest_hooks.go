package template

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const manifestHookConfirmationPhrase = "run-template-hooks"

// ConfirmManifestHooksExecution asks for explicit double confirmation before
// template-defined manifest hooks are executed.
func ConfirmManifestHooksExecution(manifest *Manifest, projectPath string, in io.Reader, out io.Writer) (bool, error) {
	if manifest == nil || len(manifest.Hooks.Post) == 0 {
		return true, nil
	}
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprintf(out, "\n⚠️  Template manifest declares %d post-scaffold hook(s) for %s:\n", len(manifest.Hooks.Post), projectPath)
	for i, hook := range manifest.Hooks.Post {
		name := strings.TrimSpace(hook.Name)
		if name == "" {
			name = fmt.Sprintf("hook-%d", i+1)
		}
		fmt.Fprintf(out, "  %d. %s\n     run: %s\n", i+1, name, strings.TrimSpace(hook.Run))
		if when := strings.TrimSpace(hook.When); when != "" {
			fmt.Fprintf(out, "     when: %s\n", when)
		}
	}

	reader := bufio.NewReader(in)
	first, err := readConfirmationLine(reader, out, "Do you want to run these template manifest hooks? (y/N): ")
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !strings.EqualFold(first, "y") && !strings.EqualFold(first, "yes") {
		return false, nil
	}

	second, err := readConfirmationLine(reader, out,
		fmt.Sprintf("Type %q to confirm: ", manifestHookConfirmationPhrase))
	if err != nil {
		return false, fmt.Errorf("failed to read second confirmation: %w", err)
	}
	if second != manifestHookConfirmationPhrase {
		return false, nil
	}

	return true, nil
}

func readConfirmationLine(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	fmt.Fprint(out, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// ExecuteManifestHooks runs manifest-declared post-scaffold hooks in order.
func ExecuteManifestHooks(ctx context.Context, projectPath string, manifest *Manifest, vars *Variables, stdout, stderr io.Writer) error {
	if manifest == nil || len(manifest.Hooks.Post) == 0 {
		return nil
	}

	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	renderer := NewRenderer(vars)
	for index, hook := range manifest.Hooks.Post {
		name := strings.TrimSpace(hook.Name)
		if name == "" {
			name = fmt.Sprintf("hook-%d", index+1)
		}
		command := strings.TrimSpace(hook.Run)
		if command == "" {
			return fmt.Errorf("template manifest hook %q has an empty run command", name)
		}

		ok, err := EvalCondition(hook.When, vars)
		if err != nil {
			return fmt.Errorf("template manifest hook %q has invalid when condition %q: %w", name, hook.When, err)
		}
		if !ok {
			fmt.Fprintf(stdout, "Skipping template manifest hook %q (condition not met)\n", name)
			continue
		}

		renderedCommand, err := renderer.RenderString(command)
		if err != nil {
			return fmt.Errorf("failed to render template manifest hook %q: %w", name, err)
		}

		fmt.Fprintf(stdout, "Running template manifest hook %q: %s\n", name, renderedCommand)
		execCmd := exec.CommandContext(ctx, "sh", "-c", renderedCommand)
		execCmd.Dir = projectPath
		execCmd.Stdout = stdout
		execCmd.Stderr = stderr
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("template manifest hook %q failed: %w", name, err)
		}
	}

	return nil
}
