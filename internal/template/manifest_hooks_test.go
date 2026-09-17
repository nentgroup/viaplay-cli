package template

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfirmManifestHooksExecution(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Hooks: ManifestHooks{
			Post: []ManifestHook{
				{Name: "deps", Run: "npm ci", When: "node"},
			},
		},
	}

	t.Run("confirmed twice", func(t *testing.T) {
		in := strings.NewReader("y\nrun-template-hooks\n")
		out := &bytes.Buffer{}
		ok, err := ConfirmManifestHooksExecution(manifest, "/tmp/project", in, out)
		if err != nil {
			t.Fatalf("ConfirmManifestHooksExecution returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected confirmation to succeed")
		}
	})

	t.Run("denied on first prompt", func(t *testing.T) {
		in := strings.NewReader("n\n")
		out := &bytes.Buffer{}
		ok, err := ConfirmManifestHooksExecution(manifest, "/tmp/project", in, out)
		if err != nil {
			t.Fatalf("ConfirmManifestHooksExecution returned error: %v", err)
		}
		if ok {
			t.Fatal("expected confirmation to be denied")
		}
	})

	t.Run("wrong second phrase", func(t *testing.T) {
		in := strings.NewReader("y\nrun-it\n")
		out := &bytes.Buffer{}
		ok, err := ConfirmManifestHooksExecution(manifest, "/tmp/project", in, out)
		if err != nil {
			t.Fatalf("ConfirmManifestHooksExecution returned error: %v", err)
		}
		if ok {
			t.Fatal("expected confirmation to fail with wrong phrase")
		}
	})
}

func TestExecuteManifestHooks_RunsAndSkipsByCondition(t *testing.T) {
	t.Parallel()

	projectPath := t.TempDir()
	markerPath := filepath.Join(projectPath, "hook-marker.txt")
	manifest := &Manifest{
		Hooks: ManifestHooks{
			Post: []ManifestHook{
				{Name: "enabled", Run: "printf 'ok' > hook-marker.txt", When: "runHook"},
				{Name: "disabled", Run: "printf 'nope' > should-not-exist.txt", When: "!runHook"},
			},
		},
	}
	vars := &Variables{Features: FeatureSet{"runHook": true}}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	if err := ExecuteManifestHooks(context.Background(), projectPath, manifest, vars, stdout, stderr); err != nil {
		t.Fatalf("ExecuteManifestHooks returned error: %v", err)
	}

	data, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("failed to read hook marker: %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("unexpected hook marker content: %q", string(data))
	}

	if _, err := os.Stat(filepath.Join(projectPath, "should-not-exist.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected disabled hook output to not exist, got: %v", err)
	}

	if !strings.Contains(stdout.String(), "Running template manifest hook") {
		t.Fatalf("expected run output, got: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Skipping template manifest hook") {
		t.Fatalf("expected skip output, got: %q", stdout.String())
	}
}

func TestExecuteManifestHooks_EmptyRunFails(t *testing.T) {
	t.Parallel()

	manifest := &Manifest{
		Hooks: ManifestHooks{
			Post: []ManifestHook{
				{Name: "broken", Run: "   "},
			},
		},
	}

	err := ExecuteManifestHooks(context.Background(), t.TempDir(), manifest, &Variables{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected an error for empty hook run command")
	}
}
