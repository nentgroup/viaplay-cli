package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallOptions configures a single pack installation.
type InstallOptions struct {
	// Scope selects project-local vs global installation. Ignored when
	// TargetDir is set.
	Scope InstallScope
	// TargetDir, when non-empty, overrides Scope and installs directly into
	// this directory.
	TargetDir string
	// Force overwrites an existing file at the resolved path.
	Force bool
	// DryRun reports what would happen without writing anything.
	DryRun bool
}

// InstallResult reports the outcome of installing one pack for one agent.
type InstallResult struct {
	Agent   Agent
	Pack    Pack
	Path    string
	Written bool
	Skipped bool
	Note    string
}

// Install writes the given pack's content, in SKILL.md format, to the
// resolved path for the given agent, according to opts. It does not
// overwrite an existing file unless opts.Force is set.
func Install(agent Agent, pack Pack, opts InstallOptions) (InstallResult, error) {
	path, err := agent.ResolvePath(pack.ID, opts.Scope, opts.TargetDir)
	if err != nil {
		return InstallResult{}, err
	}

	result := InstallResult{Agent: agent, Pack: pack, Path: path}

	if _, statErr := os.Stat(path); statErr == nil && !opts.Force {
		result.Skipped = true
		result.Note = "already exists, use --force to overwrite"
		return result, nil
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return InstallResult{}, fmt.Errorf("failed to check existing file %s: %w", path, statErr)
	}

	if opts.DryRun {
		result.Note = "dry run - not written"
		return result, nil
	}

	raw, err := pack.Content()
	if err != nil {
		return InstallResult{}, err
	}
	content := skillMDContent(pack, raw)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	if err := os.WriteFile(path, content, 0o600); err != nil {
		return InstallResult{}, fmt.Errorf("failed to write skill pack to %s: %w", path, err)
	}

	result.Written = true
	return result, nil
}
