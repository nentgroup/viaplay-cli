package project

import (
	"os"
	"path/filepath"
	"testing"

	templatepkg "github.com/nentgroup/viaplay-cli/internal/template"
)

func TestRenderHookScriptFileRendersTemplateContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "hook.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho '{{ .Project.Name }}'\n"), 0o600); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.Chmod(scriptPath, 0o755); err != nil {
		t.Fatalf("chmod script: %v", err)
	}

	vars := templatepkg.NewTemplateVariables()
	vars.Project.Name = "demo-service"
	renderer := templatepkg.NewRenderer(vars)

	renderedPath, cleanup, err := renderHookScriptFile(renderer, scriptPath)
	if err != nil {
		t.Fatalf("renderHookScriptFile returned error: %v", err)
	}
	defer cleanup()

	renderedData, err := os.ReadFile(renderedPath)
	if err != nil {
		t.Fatalf("read rendered script: %v", err)
	}

	if string(renderedData) != "#!/bin/sh\necho 'demo-service'\n" {
		t.Fatalf("unexpected rendered script contents: %q", string(renderedData))
	}
}
