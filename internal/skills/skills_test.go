package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindAndIDs(t *testing.T) {
	p, ok := Find("project-create")
	if !ok {
		t.Fatalf("expected to find project-create pack")
	}
	if p.Title == "" {
		t.Errorf("expected pack to have a title")
	}

	if _, ok := Find("does-not-exist"); ok {
		t.Errorf("expected not to find unknown pack")
	}

	ids := IDs()
	if len(ids) == 0 {
		t.Fatalf("expected at least one pack ID")
	}
}

func TestAllPacksHaveUniqueIDsAndContent(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range Packs {
		if seen[p.ID] {
			t.Errorf("duplicate pack ID: %s", p.ID)
		}
		seen[p.ID] = true

		if p.Title == "" || p.Description == "" {
			t.Errorf("pack %s missing title or description", p.ID)
		}

		content, err := p.Content()
		if err != nil {
			t.Errorf("pack %s: failed to read content: %v", p.ID, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("pack %s: expected non-empty content", p.ID)
		}
	}
}

func TestPackContent(t *testing.T) {
	p, ok := Find("project-create")
	if !ok {
		t.Fatalf("expected to find project-create pack")
	}
	content, err := p.Content()
	if err != nil {
		t.Fatalf("unexpected error reading pack content: %v", err)
	}
	if len(content) == 0 {
		t.Errorf("expected non-empty pack content")
	}
}

func TestResolvePathProjectScope(t *testing.T) {
	claude, _ := FindAgent("claude")
	path, err := claude.ResolvePath("project-create", ScopeProject, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(".claude", "skills", "vip-project-create", "SKILL.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	copilot, _ := FindAgent("copilot")
	path, err = copilot.ResolvePath("project-create", ScopeProject, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = filepath.Join(".agents", "skills", "vip-project-create", "SKILL.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolvePathGlobalScope(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claude, _ := FindAgent("claude")
	path, err := claude.ResolvePath("project-create", ScopeGlobal, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".claude", "skills", "vip-project-create", "SKILL.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	for _, id := range []string{"copilot", "cursor", "codex"} {
		agent, _ := FindAgent(id)
		path, err := agent.ResolvePath("project-create", ScopeGlobal, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(home, ".agents", "skills", "vip-project-create", "SKILL.md")
		if path != want {
			t.Errorf("%s: path = %q, want %q", id, path, want)
		}
	}
}

func TestResolvePathTargetDirOverride(t *testing.T) {
	claude, _ := FindAgent("claude")
	path, err := claude.ResolvePath("project-create", ScopeGlobal, "/custom/dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/custom/dir", "vip-project-create", "SKILL.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestSkillMDContentRewritesFrontmatter(t *testing.T) {
	pack, _ := Find("project-create")

	raw, err := pack.Content()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prepared := skillMDContent(pack, raw)
	got := string(prepared)

	wantFrontmatter := "---\nname: vip-project-create\ndescription: " + pack.Description + "\n---\n\n"
	if !strings.HasPrefix(got, wantFrontmatter) {
		n := len(got)
		if n > 200 {
			n = 200
		}
		t.Errorf("prepared content does not start with expected frontmatter.\ngot prefix: %q\nwant prefix: %q", got[:n], wantFrontmatter)
	}
	if strings.Contains(got, "\ntitle:") {
		n := len(got)
		if n > 200 {
			n = 200
		}
		t.Errorf("expected original title frontmatter key to be stripped, got: %q", got[:n])
	}
}

func TestInstallWritesFileAndRespectsForceAndDryRun(t *testing.T) {
	dir := t.TempDir()
	claude, _ := FindAgent("claude")
	pack, _ := Find("project-create")

	// Dry run: should not write anything.
	result, err := Install(claude, pack, InstallOptions{TargetDir: dir, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Written {
		t.Errorf("expected dry run to not write")
	}
	if _, statErr := os.Stat(result.Path); !os.IsNotExist(statErr) {
		t.Errorf("expected no file to exist after dry run")
	}

	// Real install.
	result, err = Install(claude, pack, InstallOptions{TargetDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Written {
		t.Errorf("expected file to be written")
	}
	if filepath.Base(result.Path) != "SKILL.md" {
		t.Errorf("expected file to be named SKILL.md, got %q", result.Path)
	}

	// Re-install without force: should skip.
	result, err = Install(claude, pack, InstallOptions{TargetDir: dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Skipped {
		t.Errorf("expected re-install without force to be skipped")
	}

	// Re-install with force: should overwrite.
	result, err = Install(claude, pack, InstallOptions{TargetDir: dir, Force: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Written {
		t.Errorf("expected forced re-install to write")
	}
}
