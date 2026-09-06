package scaffolding

import (
	"testing"

	templ "github.com/nentgroup/viaplay-cli/internal/template"
)

const (
	dynamoRulePath = "internal/storage/dynamo/*"
	dynamoDirPath  = "internal/storage/dynamo"
)

// TestShouldSkipPath_DirectoryGating is a regression test: a manifest rule
// targeting files one level under a directory (e.g. "internal/storage/dynamo/*")
// must also gate the directory entry itself. Otherwise the directory survives
// scaffolding as an empty folder even though every file inside it is skipped.
func TestShouldSkipPath_DirectoryGating(t *testing.T) {
	manifest := &templ.Manifest{
		Files: templ.ManifestFiles{
			Include: []templ.ManifestFileRule{
				{Path: dynamoRulePath, When: "dynamo"},
				{Path: "internal/storage/memory/*", When: "!dynamo"},
			},
		},
	}

	tests := []struct {
		name     string
		relPath  string
		isDir    bool
		dynamoOn bool
		wantSkip bool
	}{
		{"dynamo dir skipped when feature off", dynamoDirPath, true, false, true},
		{"dynamo file skipped when feature off", dynamoDirPath + "/service_repository.go", false, false, true},
		{"dynamo dir kept when feature on", dynamoDirPath, true, true, false},
		{"dynamo file kept when feature on", dynamoDirPath + "/service_repository.go", false, true, false},
		{"memory dir kept when dynamo off", "internal/storage/memory", true, false, false},
		{"memory dir skipped when dynamo on", "internal/storage/memory", true, true, true},
		{"shallower ancestor never skipped", "internal/storage", true, false, false},
		{"unrelated ancestor never skipped", "internal", true, false, false},
		{"sibling file untouched by rule", "internal/storage/other.go", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := &templ.Variables{Features: templ.FeatureSet{"dynamo": tt.dynamoOn}}
			got := shouldSkipPath(tt.relPath, tt.isDir, manifest, vars)
			if got != tt.wantSkip {
				t.Fatalf("shouldSkipPath(%q, isDir=%v, dynamo=%v) = %v, want %v",
					tt.relPath, tt.isDir, tt.dynamoOn, got, tt.wantSkip)
			}
		})
	}
}

func TestDirMatchesPattern(t *testing.T) {
	tests := []struct {
		pattern string
		relPath string
		want    bool
	}{
		{dynamoRulePath, dynamoDirPath, true},
		{dynamoRulePath, "internal/storage", false},
		{dynamoRulePath, "internal", false},
		{dynamoRulePath, dynamoDirPath + "/nested", true},
		{"internal/events/*", "internal/events", true},
		{"internal/events/*", "internal/events/sub", true},
	}
	for _, tt := range tests {
		if got := dirMatchesPattern(tt.pattern, tt.relPath); got != tt.want {
			t.Errorf("dirMatchesPattern(%q, %q) = %v, want %v", tt.pattern, tt.relPath, got, tt.want)
		}
	}
}
