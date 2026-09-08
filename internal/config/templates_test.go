package config

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testOrg  = "nentgroup"
	testTeam = "gecko"
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
		DefaultOrganization: testOrg,
		Templates: map[string]map[string]*TemplateDefinition{
			"go": {
				"service": {Source: "git@github.com:nentgroup/base-template.git"},
			},
		},
	}

	teamDir := filepath.Join(dir, OrgsDirName, testOrg, TeamsDirName, testTeam)
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

	if err := cfg.ApplyTeamTemplateOverrides(testTeam, testOrg); err != nil {
		t.Fatalf("ApplyTeamTemplateOverrides returned error: %v", err)
	}

	if got := cfg.GetTemplateSource("go", "service"); got != "git@github.com:nentgroup/gecko-template.git" {
		t.Fatalf("expected overridden source, got %q", got)
	}
}

func TestGetTeamTemplateOverrideReturnsMatchingEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: testOrg,
	}

	teamDir := filepath.Join(dir, OrgsDirName, testOrg, TeamsDirName, testTeam)
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

	override, err := cfg.GetTeamTemplateOverride(testTeam, testOrg, "go", "service")
	if err != nil {
		t.Fatalf("GetTeamTemplateOverride returned error: %v", err)
	}
	if override == nil || override.Source != "git@github.com:nentgroup/gecko-template.git" {
		t.Fatalf("expected matching override, got %#v", override)
	}
}

func TestGetTeamTemplateOverrideReturnsNilForUnrelatedEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: testOrg,
	}

	teamDir := filepath.Join(dir, OrgsDirName, testOrg, TeamsDirName, testTeam)
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

	override, err := cfg.GetTeamTemplateOverride(testTeam, testOrg, "go", "worker")
	if err != nil {
		t.Fatalf("GetTeamTemplateOverride returned error: %v", err)
	}
	if override != nil {
		t.Fatalf("expected nil override for unrelated language/type, got %#v", override)
	}
}

func TestGetTeamTemplateOverrideReturnsNilWhenNoTeamConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir: dir,
		TeamsDir:  filepath.Join(dir, TeamsDirName),
	}

	override, err := cfg.GetTeamTemplateOverride(testTeam, testOrg, "go", "service")
	if err != nil {
		t.Fatalf("GetTeamTemplateOverride returned error: %v", err)
	}
	if override != nil {
		t.Fatalf("expected nil override when no team config file exists, got %#v", override)
	}
}

func TestGetTeamTemplateOverrideReturnsNilForEmptyTeam(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{}
	override, err := cfg.GetTeamTemplateOverride("", testOrg, "go", "service")
	if err != nil {
		t.Fatalf("GetTeamTemplateOverride returned error: %v", err)
	}
	if override != nil {
		t.Fatalf("expected nil override for empty team, got %#v", override)
	}
}

func TestFindTeamTemplateConfigFileReturnsExistingConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: testOrg,
	}

	teamDir := filepath.Join(dir, OrgsDirName, testOrg, TeamsDirName, testTeam)
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("mkdir team dir: %v", err)
	}
	configFile := filepath.Join(teamDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("templates:\n  go:\n    service:\n      source: src\n"), 0o600); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}

	got, err := cfg.FindTeamTemplateConfigFile(testTeam, testOrg)
	if err != nil {
		t.Fatalf("FindTeamTemplateConfigFile returned error: %v", err)
	}
	if got != configFile {
		t.Fatalf("expected %q, got %q", configFile, got)
	}
}

func TestFindTeamTemplateConfigFileReturnsPathForExistingDirWithoutConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: testOrg,
	}

	teamDir := filepath.Join(dir, OrgsDirName, testOrg, TeamsDirName, testTeam)
	if err := os.MkdirAll(teamDir, 0o755); err != nil {
		t.Fatalf("mkdir team dir: %v", err)
	}

	got, err := cfg.FindTeamTemplateConfigFile(testTeam, testOrg)
	if err != nil {
		t.Fatalf("FindTeamTemplateConfigFile returned error: %v", err)
	}
	want := filepath.Join(teamDir, "config.yaml")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFindTeamTemplateConfigFileReturnsEmptyWhenTeamDirMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir:           dir,
		TeamsDir:            filepath.Join(dir, TeamsDirName),
		DefaultOrganization: testOrg,
	}

	got, err := cfg.FindTeamTemplateConfigFile(testTeam, testOrg)
	if err != nil {
		t.Fatalf("FindTeamTemplateConfigFile returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty path when team dir doesn't exist, got %q", got)
	}
}

func TestFindTeamTemplateConfigFileReturnsEmptyForEmptyTeam(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{}
	got, err := cfg.FindTeamTemplateConfigFile("", testOrg)
	if err != nil {
		t.Fatalf("FindTeamTemplateConfigFile returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty path for empty team, got %q", got)
	}
}

func TestUpdateTemplateConfigFileCreatesMissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	if err := UpdateTemplateConfigFile(configFile, "go", "service", "github.com/org/go-service-template"); err != nil {
		t.Fatalf("UpdateTemplateConfigFile returned error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	for _, expected := range []string{"templates:", "go:", "service:", "source: github.com/org/go-service-template"} {
		if !containsLine(content, expected) {
			t.Fatalf("expected newly created config file to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestUpdateTemplateConfigFileAddsNewEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("default_team: \"gecko\"\n"), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := UpdateTemplateConfigFile(configFile, "go", "worker", "github.com/org/go-worker-template"); err != nil {
		t.Fatalf("UpdateTemplateConfigFile returned error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	for _, expected := range []string{
		"default_team: \"gecko\"",
		"templates:",
		"go:",
		"worker:",
		"source: github.com/org/go-worker-template",
	} {
		if !containsLine(content, expected) {
			t.Fatalf("expected config file to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestUpdateTemplateConfigFileOverwritesExistingEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	initial := "templates:\n  go:\n    service:\n      source: old-source\n"
	if err := os.WriteFile(configFile, []byte(initial), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := UpdateTemplateConfigFile(configFile, "go", "service", "new-source"); err != nil {
		t.Fatalf("UpdateTemplateConfigFile returned error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	if !containsLine(content, "source: new-source") {
		t.Fatalf("expected config file to contain updated source, got:\n%s", content)
	}
	if containsLine(content, "source: old-source") {
		t.Fatalf("expected old source to be replaced, got:\n%s", content)
	}
}

func TestRemoveTemplateConfigFileRemovesEntryAndEmptyLanguage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	initial := "templates:\n  go:\n    worker:\n      source: some-source\n  node:\n    service:\n      source: node-source\n"
	if err := os.WriteFile(configFile, []byte(initial), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	removed, err := RemoveTemplateConfigFile(configFile, "go", "worker")
	if err != nil {
		t.Fatalf("RemoveTemplateConfigFile returned error: %v", err)
	}
	if !removed {
		t.Fatalf("expected entry to be reported as removed")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	if containsLine(content, "worker:") || containsLine(content, "source: some-source") {
		t.Fatalf("expected go/worker entry to be removed, got:\n%s", content)
	}
	if !containsLine(content, "source: node-source") {
		t.Fatalf("expected unrelated node/service entry to survive, got:\n%s", content)
	}
}

func TestRemoveTemplateConfigFileReportsMissingEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("templates:\n  go:\n    service:\n      source: src\n"), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	removed, err := RemoveTemplateConfigFile(configFile, "go", "worker")
	if err != nil {
		t.Fatalf("RemoveTemplateConfigFile returned error: %v", err)
	}
	if removed {
		t.Fatalf("expected removed=false for a nonexistent entry")
	}
}
