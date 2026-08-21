package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateSourceConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("default_team: \"gecko\"\n"), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	err := UpdateSourceConfigFile(configFile, SourceConfig{
		Repository: "git@github.com:nentgroup/vip-shared-configs.git",
		Branch:     "main",
		Root:       "shared",
	})
	if err != nil {
		t.Fatalf("UpdateSourceConfigFile returned error: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	content := string(data)
	for _, expected := range []string{
		"default_team: \"gecko\"",
		"config_source:",
		"repository: git@github.com:nentgroup/vip-shared-configs.git",
		"branch: main",
		"root: shared",
	} {
		if !containsLine(content, expected) {
			t.Fatalf("expected config file to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestResolveSourceTeamDirDetectsUniqueOrgMatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := &Configuration{
		ConfigDir: dir,
		TeamsDir:  filepath.Join(dir, TeamsDirName),
	}

	sourceRoot := filepath.Join(dir, "source")
	teamPath := filepath.Join(sourceRoot, OrgsDirName, "nentgroup", TeamsDirName, "gecko")
	if err := os.MkdirAll(teamPath, 0o755); err != nil {
		t.Fatalf("mkdir source team path: %v", err)
	}

	sourceDir, destDir, err := cfg.resolveSourceTeamDir(sourceRoot, "gecko", "")
	if err != nil {
		t.Fatalf("resolveSourceTeamDir returned error: %v", err)
	}

	if sourceDir != teamPath {
		t.Fatalf("expected source dir %s, got %s", teamPath, sourceDir)
	}

	expectedDest := cfg.GetTeamDir("gecko", "nentgroup")
	if destDir != expectedDest {
		t.Fatalf("expected destination %s, got %s", expectedDest, destDir)
	}
}

func containsLine(content, expected string) bool {
	for _, line := range splitLines(content) {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}
	return false
}

func splitLines(value string) []string {
	lines := []string{}
	start := 0
	for i := 0; i < len(value); i++ {
		if value[i] == '\n' {
			lines = append(lines, value[start:i])
			start = i + 1
		}
	}
	if start <= len(value) {
		lines = append(lines, value[start:])
	}
	return lines
}
