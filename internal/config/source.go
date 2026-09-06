package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	gitpkg "github.com/nentgroup/viaplay-cli/internal/git"
)

const (
	DefaultSourceBranch   = "main"
	DefaultSourceRoot     = "."
	SharedConfigRepoName  = "vip-shared-configs"
	sourceCheckoutDirName = "source"
	yamlStrTag            = "!!str"
)

// SourceConfig defines the shared config repository used for pull operations.
type SourceConfig struct {
	Repository string `mapstructure:"repository" yaml:"repository"`
	Branch     string `mapstructure:"branch" yaml:"branch"`
	Root       string `mapstructure:"root" yaml:"root"`
}

// SourcePullSummary describes one pull operation from the shared config source.
type SourcePullSummary struct {
	SourceRoot string
	Targets    []string
	FileOps    *FileOps
}

func (s SourceConfig) normalized() SourceConfig {
	if strings.TrimSpace(s.Branch) == "" {
		s.Branch = DefaultSourceBranch
	}
	if strings.TrimSpace(s.Root) == "" {
		s.Root = DefaultSourceRoot
	}
	return s
}

// HasSource reports whether a shared config source is configured.
func (c *Configuration) HasSource() bool {
	return c != nil && strings.TrimSpace(c.Source.Repository) != ""
}

// GetSourceCheckoutDir returns the local checkout used for the shared config source.
func (c *Configuration) GetSourceCheckoutDir() string {
	return filepath.Join(c.ConfigDir, sourceCheckoutDirName)
}

// UpdateSourceConfigFile writes config_source settings into the main config file.
func UpdateSourceConfigFile(configFile string, source SourceConfig) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	source = source.normalized()

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}
	if len(doc.Content) == 0 {
		doc.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("config file %s must contain a YAML mapping at the top level", configFile)
	}

	sourceNode := ensureMappingValue(root, "config_source")
	setMappingString(sourceNode, "repository", source.Repository)
	setMappingString(sourceNode, "branch", source.Branch)
	setMappingString(sourceNode, "root", source.Root)

	var rendered bytes.Buffer
	encoder := yaml.NewEncoder(&rendered)
	encoder.SetIndent(2)
	if err := encoder.Encode(&doc); err != nil {
		return fmt.Errorf("failed to encode config file %s: %w", configFile, err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to finalize config file %s: %w", configFile, err)
	}

	if err := os.WriteFile(configFile, rendered.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configFile, err)
	}

	return nil
}

// SyncSourceRepository clones or updates the configured shared config repository.
func (c *Configuration) SyncSourceRepository(ctx context.Context) (string, error) {
	if !c.HasSource() {
		return "", fmt.Errorf("shared config source is not configured")
	}

	checkoutDir := c.GetSourceCheckoutDir()
	source := c.Source.normalized()
	if gitpkg.IsGitRepository(checkoutDir) {
		if err := gitpkg.Update(ctx, gitpkg.UpdateOptions{Directory: checkoutDir, Branch: source.Branch}); err != nil {
			return "", err
		}
	} else {
		if err := os.RemoveAll(checkoutDir); err != nil {
			return "", fmt.Errorf("failed to prepare shared config checkout %s: %w", checkoutDir, err)
		}
		if err := gitpkg.Clone(ctx, gitpkg.CloneOptions{
			URL:       source.Repository,
			Branch:    source.Branch,
			Directory: checkoutDir,
			Depth:     1,
		}); err != nil {
			return "", err
		}
	}

	sourceRoot := filepath.Clean(filepath.Join(checkoutDir, source.Root))
	info, err := os.Stat(sourceRoot)
	if err != nil {
		return "", fmt.Errorf("shared config source root %s is not available: %w", sourceRoot, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("shared config source root %s is not a directory", sourceRoot)
	}

	return sourceRoot, nil
}

// PullHooks copies the hooks from the active team's shared config directory into the local
// hooks directory. org and team fall back to the values in the main config when empty.
func (c *Configuration) PullHooks(ctx context.Context, org, team string) (*SourcePullSummary, error) {
	if org == "" {
		org = strings.TrimSpace(c.DefaultOrganization)
	}
	if team == "" {
		team = strings.TrimSpace(c.DefaultTeam)
	}
	if team == "" {
		return nil, fmt.Errorf("team is required for hook pull (set default_team in config or pass --team)")
	}

	sourceRoot, err := c.SyncSourceRepository(ctx)
	if err != nil {
		return nil, err
	}

	sourceDir, _, err := c.resolveSourceTeamDir(sourceRoot, team, org)
	if err != nil {
		return nil, err
	}

	summary := &SourcePullSummary{
		SourceRoot: sourceRoot,
		FileOps:    NewFileOps(),
	}

	found, err := c.pullHooksFromDir(filepath.Join(sourceDir, "hooks"), summary)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no hooks directory found for team %s in shared config source", team)
	}

	summary.Targets = []string{c.GetHooksDir()}
	return summary, nil
}

// PullTeam copies one team directory from the shared config source into the local config,
// and merges team-level hooks into the local hooks directory.
func (c *Configuration) PullTeam(ctx context.Context, team, org string) (*SourcePullSummary, error) {
	if strings.TrimSpace(team) == "" {
		return nil, fmt.Errorf("team name is required")
	}

	sourceRoot, err := c.SyncSourceRepository(ctx)
	if err != nil {
		return nil, err
	}

	sourceDir, destDir, err := c.resolveSourceTeamDir(sourceRoot, team, org)
	if err != nil {
		return nil, err
	}

	result, err := copyTree(sourceDir, destDir)
	if err != nil {
		return nil, err
	}

	summary := &SourcePullSummary{
		SourceRoot: sourceRoot,
		Targets:    []string{destDir},
		FileOps:    result,
	}

	// Merge team-level hooks (a hooks/ subdir inside the team source dir).
	if _, err := c.pullHooksFromDir(filepath.Join(sourceDir, "hooks"), summary); err != nil {
		return nil, err
	}
	if len(summary.FileOps.Created)+len(summary.FileOps.Overrode) > len(result.Created)+len(result.Overrode) {
		summary.Targets = append(summary.Targets, c.GetHooksDir())
	}

	return summary, nil
}

// PullAll copies all shared team configs and their hooks from the source into the local config.
func (c *Configuration) PullAll(ctx context.Context) (*SourcePullSummary, error) {
	sourceRoot, err := c.SyncSourceRepository(ctx)
	if err != nil {
		return nil, err
	}

	summary := &SourcePullSummary{
		SourceRoot: sourceRoot,
		FileOps:    NewFileOps(),
	}

	if err := c.pullLegacyTeams(sourceRoot, summary); err != nil {
		return nil, err
	}
	if err := c.pullOrganizationTeams(sourceRoot, summary); err != nil {
		return nil, err
	}

	if len(summary.Targets) == 0 {
		return nil, fmt.Errorf("shared config source does not contain any team configs (teams/ or orgs/)")
	}

	return summary, nil
}

func mergePulledTree(summary *SourcePullSummary, sourceDir, destDir string) error {
	result, err := copyTree(sourceDir, destDir)
	if err != nil {
		return err
	}

	summary.Targets = append(summary.Targets, destDir)
	summary.FileOps.Merge(result)
	return nil
}

// pullHooksFromDir merges a hooks directory into the local hooks dir.
// It is a no-op when the source directory does not exist.
// Returns true if the directory was found and processed.
func (c *Configuration) pullHooksFromDir(hooksDir string, summary *SourcePullSummary) (bool, error) {
	exists, err := directoryExists(hooksDir)
	if err != nil || !exists {
		return false, err
	}

	result, err := copyTree(hooksDir, c.GetHooksDir())
	if err != nil {
		return false, err
	}

	summary.FileOps.Merge(result)
	return true, nil
}

func (c *Configuration) pullLegacyTeams(sourceRoot string, summary *SourcePullSummary) error {
	legacyTeamsDir := filepath.Join(sourceRoot, TeamsDirName)
	exists, err := directoryExists(legacyTeamsDir)
	if err != nil || !exists {
		return err
	}

	entries, err := os.ReadDir(legacyTeamsDir)
	if err != nil {
		return fmt.Errorf("failed to read shared teams directory %s: %w", legacyTeamsDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		teamSourceDir := filepath.Join(legacyTeamsDir, entry.Name())
		destDir := c.GetTeamDir(entry.Name(), "")
		if err := mergePulledTree(summary, teamSourceDir, destDir); err != nil {
			return err
		}

		// Team-level hooks inside legacy team source dir
		if _, err := c.pullHooksFromDir(filepath.Join(teamSourceDir, "hooks"), summary); err != nil {
			return err
		}
	}

	return nil
}

func (c *Configuration) pullOrganizationTeams(sourceRoot string, summary *SourcePullSummary) error {
	orgsRoot := filepath.Join(sourceRoot, OrgsDirName)
	exists, err := directoryExists(orgsRoot)
	if err != nil || !exists {
		return err
	}

	orgEntries, err := os.ReadDir(orgsRoot)
	if err != nil {
		return fmt.Errorf("failed to read shared organizations directory %s: %w", orgsRoot, err)
	}
	for _, orgEntry := range orgEntries {
		if !orgEntry.IsDir() {
			continue
		}

		orgDir := filepath.Join(orgsRoot, orgEntry.Name())
		teamsRoot := filepath.Join(orgDir, TeamsDirName)
		teamEntries, err := os.ReadDir(teamsRoot)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("failed to read shared teams directory %s: %w", teamsRoot, err)
		}

		for _, teamEntry := range teamEntries {
			if !teamEntry.IsDir() {
				continue
			}

			teamSourceDir := filepath.Join(teamsRoot, teamEntry.Name())
			destDir := c.GetTeamDir(teamEntry.Name(), orgEntry.Name())
			if err := mergePulledTree(summary, teamSourceDir, destDir); err != nil {
				return err
			}

			// Team-level hooks inside the team source dir
			if _, err := c.pullHooksFromDir(filepath.Join(teamSourceDir, "hooks"), summary); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Configuration) resolveSourceTeamDir(sourceRoot, team, org string) (string, string, error) {
	if org != "" {
		orgSourceDir := filepath.Join(sourceRoot, OrgsDirName, org, TeamsDirName, team)
		if exists, err := directoryExists(orgSourceDir); err != nil {
			return "", "", err
		} else if exists {
			return orgSourceDir, c.GetTeamDir(team, org), nil
		}

		legacySourceDir := filepath.Join(sourceRoot, TeamsDirName, team)
		if exists, err := directoryExists(legacySourceDir); err != nil {
			return "", "", err
		} else if exists {
			return legacySourceDir, c.GetTeamDir(team, org), nil
		}

		return "", "", fmt.Errorf("team %s not found in shared source for organization %s", team, org)
	}

	defaultOrg := strings.TrimSpace(c.DefaultOrganization)
	if defaultOrg != "" {
		defaultOrgSourceDir := filepath.Join(sourceRoot, OrgsDirName, defaultOrg, TeamsDirName, team)
		if exists, err := directoryExists(defaultOrgSourceDir); err != nil {
			return "", "", err
		} else if exists {
			return defaultOrgSourceDir, c.GetTeamDir(team, defaultOrg), nil
		}
	}

	legacySourceDir := filepath.Join(sourceRoot, TeamsDirName, team)
	if exists, err := directoryExists(legacySourceDir); err != nil {
		return "", "", err
	} else if exists {
		return legacySourceDir, c.GetTeamDir(team, ""), nil
	}

	matches, err := filepath.Glob(filepath.Join(sourceRoot, OrgsDirName, "*", TeamsDirName, team))
	if err != nil {
		return "", "", fmt.Errorf("failed to search shared source for team %s: %w", team, err)
	}
	switch len(matches) {
	case 0:
		return "", "", fmt.Errorf("team %s not found in shared source", team)
	case 1:
		orgName := filepath.Base(filepath.Dir(filepath.Dir(matches[0])))
		return matches[0], c.GetTeamDir(team, orgName), nil
	default:
		return "", "", fmt.Errorf("team %s exists in multiple organizations; use --organization", team)
	}
}

func ensureMappingValue(root *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == key {
			valueNode := root.Content[i+1]
			if valueNode.Kind != yaml.MappingNode {
				valueNode.Kind = yaml.MappingNode
				valueNode.Tag = "!!map"
				valueNode.Content = nil
			}
			return valueNode
		}
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: yamlStrTag, Value: key}
	valueNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content, keyNode, valueNode)
	return valueNode
}

func setMappingString(root *yaml.Node, key, value string) {
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == key {
			root.Content[i+1].Kind = yaml.ScalarNode
			root.Content[i+1].Tag = yamlStrTag
			root.Content[i+1].Value = value
			root.Content[i+1].Style = 0
			return
		}
	}

	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: yamlStrTag, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: yamlStrTag, Value: value},
	)
}

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	return info.IsDir(), nil
}

func copyTree(srcDir, destDir string) (*FileOps, error) {
	info, err := os.Stat(srcDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("shared config path not found: %s", srcDir)
		}
		return nil, fmt.Errorf("failed to stat shared config path %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("shared config path is not a directory: %s", srcDir)
	}

	result := NewFileOps()
	srcRoot, err := os.OpenRoot(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open shared config root %s: %w", srcDir, err)
	}
	defer srcRoot.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	destRoot, err := os.OpenRoot(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination root %s: %w", destDir, err)
	}
	defer destRoot.Close()

	if err := fs.WalkDir(srcRoot.FS(), ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return destRoot.MkdirAll(path, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to copy symlink from shared config source: %s", path)
		}

		sourceInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if !sourceInfo.Mode().IsRegular() {
			return nil
		}

		existed := true
		if _, err := destRoot.Stat(path); errors.Is(err, os.ErrNotExist) {
			existed = false
		} else if err != nil {
			return err
		}

		data, err := srcRoot.ReadFile(path)
		if err != nil {
			return err
		}
		if err := destRoot.WriteFile(path, data, sourceInfo.Mode().Perm()); err != nil {
			return err
		}
		if err := destRoot.Chmod(path, sourceInfo.Mode().Perm()); err != nil {
			return err
		}

		if existed {
			result.Overrode = append(result.Overrode, filepath.Join(destDir, path))
		} else {
			result.Created = append(result.Created, filepath.Join(destDir, path))
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to copy shared config from %s to %s: %w", srcDir, destDir, err)
	}

	return result, nil
}
