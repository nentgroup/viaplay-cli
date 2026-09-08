package template

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	manifestNameFromVipYaml = "from-vip-yaml"
	manifestNameFromVipYml  = "from-vip-yml"
	manifestNameFromLegacy  = "from-legacy"
)

func writeManifestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

func TestLoadManifest_NoManifestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest != nil {
		t.Fatalf("expected nil manifest, got: %+v", manifest)
	}
}

func TestLoadManifest_VipYaml(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestFile(t, dir, ".vip.yaml", "schema: 2\nmetadata:\n  name: from-vip-yaml\n")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest == nil || manifest.Metadata.Name != manifestNameFromVipYaml {
		t.Fatalf("expected manifest loaded from .vip.yaml, got: %+v", manifest)
	}
}

func TestLoadManifest_VipYml(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestFile(t, dir, ".vip.yml", "schema: 2\nmetadata:\n  name: from-vip-yml\n")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest == nil || manifest.Metadata.Name != manifestNameFromVipYml {
		t.Fatalf("expected manifest loaded from .vip.yml, got: %+v", manifest)
	}
}

func TestLoadManifest_LegacyTemplateYamlFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestFile(t, dir, "template.yaml", "schema: 2\nmetadata:\n  name: from-legacy\n")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest == nil || manifest.Metadata.Name != manifestNameFromLegacy {
		t.Fatalf("expected manifest loaded from legacy template.yaml, got: %+v", manifest)
	}
}

func TestLoadManifest_VipYamlTakesPriorityOverLegacy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestFile(t, dir, "template.yaml", "schema: 2\nmetadata:\n  name: from-legacy\n")
	writeManifestFile(t, dir, ".vip.yaml", "schema: 2\nmetadata:\n  name: from-vip-yaml\n")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest == nil || manifest.Metadata.Name != manifestNameFromVipYaml {
		t.Fatalf("expected .vip.yaml to take priority over legacy template.yaml, got: %+v", manifest)
	}
}

func TestLoadManifest_VipYamlTakesPriorityOverVipYml(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestFile(t, dir, ".vip.yml", "schema: 2\nmetadata:\n  name: from-vip-yml\n")
	writeManifestFile(t, dir, ".vip.yaml", "schema: 2\nmetadata:\n  name: from-vip-yaml\n")

	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if manifest == nil || manifest.Metadata.Name != manifestNameFromVipYaml {
		t.Fatalf("expected .vip.yaml to take priority over .vip.yml, got: %+v", manifest)
	}
}
