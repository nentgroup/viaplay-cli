package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeTemplateDefinitionsPreservesBaseAndOverridesFields(t *testing.T) {
	t.Parallel()

	base := &TemplateDefinition{
		Source: "git@github.com:nentgroup/base-template.git",
		Hooks: &TemplateHooks{
			Post: &TemplateHookStage{
				Install: &PostInstallHook{
					Cmd: []string{"go mod tidy"},
				},
			},
		},
	}
	override := &TemplateDefinition{
		Hooks: &TemplateHooks{
			Post: &TemplateHookStage{
				Install: &PostInstallHook{
					Cmd: []string{"task team:setup"},
				},
			},
		},
	}

	merged := mergeTemplateDefinitions(base, override)
	if merged.Source != base.Source {
		t.Fatalf("expected source %q, got %q", base.Source, merged.Source)
	}
	if len(merged.Hooks.Post.Install.Cmd) != 1 || merged.Hooks.Post.Install.Cmd[0] != "task team:setup" {
		t.Fatalf("expected overridden hooks, got %#v", merged.Hooks.Post.Install.Cmd)
	}
}

func TestApplyTeamTemplateOverridesLoadsTeamConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: "nentgroup",
		Templates: map[string]map[string]*TemplateDefinition{
			"go": {
				"service": {Source: "git@github.com:nentgroup/base-template.git"},
			},
		},
	}

	teamDir := filepath.Join(dir, OrgsDirName, "nentgroup", TeamsDirName, "gecko")
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("mkdir team dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(teamDir, "config.yaml"), []byte(`
templates:
  go:
    service:
      source: git@github.com:nentgroup/gecko-template.git
`), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	if err := cfg.ApplyTeamTemplateOverrides("gecko", "nentgroup"); err != nil {
		t.Fatalf("ApplyTeamTemplateOverrides returned error: %v", err)
	}

	if got := cfg.GetTemplateSource("go", "service"); got != "git@github.com:nentgroup/gecko-template.git" {
		t.Fatalf("expected overridden source, got %q", got)
	}
}
