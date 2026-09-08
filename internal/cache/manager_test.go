package cache

import (
	"context"
	"os"
	"testing"

	"github.com/nentgroup/viaplay-cli/internal/config"
)

const (
	testRepoSlug  = "nentgroup/go-service-template"
	testLocalPath = "/path/to/template"
	testSSHURL    = "git@github.com:" + testRepoSlug + ".git"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		wantType     SourceType
		wantLocation string
		wantRef      string
		wantErr      bool
	}{
		{
			name:         "explicit github prefix",
			source:       "github@" + testRepoSlug,
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "explicit github prefix with ref",
			source:       "github@" + testRepoSlug + "@main",
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
			wantRef:      "main",
		},
		{
			name:         "bare github.com address",
			source:       "github.com/" + testRepoSlug,
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "https github url",
			source:       "https://github.com/" + testRepoSlug,
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "https github url with .git suffix",
			source:       "https://github.com/" + testRepoSlug + ".git",
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "https github url with trailing slash",
			source:       "https://github.com/" + testRepoSlug + "/",
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "http github url",
			source:       "http://github.com/" + testRepoSlug,
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
		},
		{
			name:         "github.com address with ref",
			source:       "github.com/" + testRepoSlug + "@v1.2.3",
			wantType:     SourceTypeGitHub,
			wantLocation: testRepoSlug,
			wantRef:      "v1.2.3",
		},
		{
			name:         "explicit local prefix",
			source:       "local@" + testLocalPath,
			wantType:     SourceTypeLocal,
			wantLocation: testLocalPath,
		},
		{
			name:         "bare path defaults to local",
			source:       testLocalPath,
			wantType:     SourceTypeLocal,
			wantLocation: testLocalPath,
		},
		{
			name:         "explicit url prefix",
			source:       "url@https://example.com/template.tar.gz",
			wantType:     SourceTypeURL,
			wantLocation: "https://example.com/template.tar.gz",
		},
		{
			name:         "ssh url with colon",
			source:       testSSHURL,
			wantType:     SourceTypeSSH,
			wantLocation: testSSHURL,
		},
		{
			name:         "ssh url with slash",
			source:       "git@github.com/" + testRepoSlug + ".git",
			wantType:     SourceTypeSSH,
			wantLocation: testSSHURL,
		},
		{
			name:         "git@ shorthand with owner/repo (no host)",
			source:       "git@" + testRepoSlug,
			wantType:     SourceTypeSSH,
			wantLocation: testSSHURL,
		},
		{
			name:         "git@ shorthand with owner/repo.git (no host)",
			source:       "git@" + testRepoSlug + ".git",
			wantType:     SourceTypeSSH,
			wantLocation: testSSHURL,
		},
		{
			name:         "git@ shorthand with owner/repo and ref (no host)",
			source:       "git@" + testRepoSlug + "@main",
			wantType:     SourceTypeSSH,
			wantLocation: testSSHURL,
			wantRef:      "main",
		},
		{
			name:         "git@ full SSH URL for a non-GitHub host is left untouched",
			source:       "git@bitbucket.org:" + testRepoSlug + ".git",
			wantType:     SourceTypeSSH,
			wantLocation: "git@bitbucket.org:" + testRepoSlug + ".git",
		},
		{
			name:    "unknown prefix errors",
			source:  "ftp@example.com/template",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.source)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSource(%q) expected error, got none", tt.source)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSource(%q) unexpected error: %v", tt.source, err)
			}
			if got.Type != tt.wantType {
				t.Errorf("ParseSource(%q).Type = %q, want %q", tt.source, got.Type, tt.wantType)
			}
			if got.Location != tt.wantLocation {
				t.Errorf("ParseSource(%q).Location = %q, want %q", tt.source, got.Location, tt.wantLocation)
			}
			if got.Reference != tt.wantRef {
				t.Errorf("ParseSource(%q).Reference = %q, want %q", tt.source, got.Reference, tt.wantRef)
			}
		})
	}
}

// TestEnsureEphemeralTemplateLocalSourceDoesNotTouchCache verifies that
// resolving a local template source through EnsureEphemeralTemplate returns
// the path directly and never reads from or writes to the persistent
// template cache directory, since local templates are already used in place.
func TestEnsureEphemeralTemplateLocalSourceDoesNotTouchCache(t *testing.T) {
	templateDir := t.TempDir()
	baseCacheDir := t.TempDir()
	m := &Manager{Config: &config.Configuration{}, BaseCacheDir: baseCacheDir}

	got, cleanup, err := m.EnsureEphemeralTemplate(context.Background(), "local@"+templateDir)
	if err != nil {
		t.Fatalf("EnsureEphemeralTemplate unexpected error: %v", err)
	}
	defer cleanup()

	if got != templateDir {
		t.Errorf("EnsureEphemeralTemplate path = %q, want %q", got, templateDir)
	}

	entries, err := os.ReadDir(baseCacheDir)
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected cache dir %q to remain empty, found entries: %v", baseCacheDir, entries)
	}
}

// TestEnsureEphemeralTemplateInvalidSourceReturnsError ensures parse errors
// are surfaced with a no-op cleanup, rather than a nil cleanup that would
// panic if called.
func TestEnsureEphemeralTemplateInvalidSourceReturnsError(t *testing.T) {
	m := &Manager{Config: &config.Configuration{}, BaseCacheDir: t.TempDir()}

	_, cleanup, err := m.EnsureEphemeralTemplate(context.Background(), "ftp@example.com/template")
	if err == nil {
		t.Fatal("EnsureEphemeralTemplate expected error for unknown source type, got none")
	}
	if cleanup == nil {
		t.Fatal("EnsureEphemeralTemplate returned nil cleanup, want a no-op function")
	}
	cleanup() // must not panic
}
