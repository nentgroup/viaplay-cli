package scaffolding

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	templ "github.com/nentgroup/viaplay-cli/internal/template"
)

func TestManualE2E_NoInputMissingRequiredLeavesNoDir(t *testing.T) {
	tplDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tplDir, "template.yaml"), []byte("variables:\n  - key: shortName\n    type: string\n    required: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tplDir, "README.md"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	cacheDir := t.TempDir()
	cfg := &config.Configuration{CacheDir: cacheDir}
	cm := cache.NewManager(cfg)
	scaffolder := NewProjectScaffolder(cm, cfg)

	destRoot := t.TempDir()
	destPath := filepath.Join(destRoot, "myproj")

	vars := &templ.Variables{}
	_, err := scaffolder.ScaffoldProjectWithOptions(context.Background(), destPath, "go", "service", "local@"+tplDir, vars, true, false, nil, true)
	if err == nil {
		t.Fatal("expected error due to missing required shortName")
	}
	t.Logf("got expected error: %v", err)

	if _, statErr := os.Stat(destPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected destPath to NOT exist, but stat returned: %v", statErr)
	}
}
