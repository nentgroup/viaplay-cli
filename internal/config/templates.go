package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type teamOverrideConfigFile struct {
	Templates map[string]map[string]*TemplateDefinition `yaml:"templates"`
}

func normalizeTemplateDefinitions(raw map[string]interface{}) (map[string]map[string]*TemplateDefinition, error) {
	result := make(map[string]map[string]*TemplateDefinition)
	for language, rawTypes := range raw {
		typeMap, ok := rawTypes.(map[string]interface{})
		if !ok {
			continue
		}

		result[language] = make(map[string]*TemplateDefinition)
		for projectType, rawDefinition := range typeMap {
			definition, err := normalizeTemplateDefinition(rawDefinition)
			if err != nil {
				return nil, fmt.Errorf("templates.%s.%s: %w", language, projectType, err)
			}
			if definition == nil {
				continue
			}
			result[language][projectType] = definition
		}
	}

	return result, nil
}

func normalizeTemplateDefinition(raw interface{}) (*TemplateDefinition, error) {
	switch value := raw.(type) {
	case string:
		return &TemplateDefinition{Source: value}, nil
	case map[string]interface{}:
		data, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}

		var definition TemplateDefinition
		if err := yaml.Unmarshal(data, &definition); err != nil {
			return nil, err
		}
		return &definition, nil
	default:
		return nil, fmt.Errorf("unsupported template definition type %T", raw)
	}
}

// UpdateTemplateConfigFile writes/overwrites templates.<language>.<type>.source
// in the main config file, preserving the rest of the file's structure and
// comments. Used by 'vip template add' to register a new template mapping.
func UpdateTemplateConfigFile(configFile, language, projectType, source string) error {
	doc, err := loadYAMLNodeDocument(configFile)
	if err != nil {
		return err
	}

	root, err := rootMappingNode(doc, configFile)
	if err != nil {
		return err
	}

	templatesNode := ensureMappingValue(root, "templates")
	languageNode := ensureMappingValue(templatesNode, language)
	typeNode := ensureMappingValue(languageNode, projectType)
	setMappingString(typeNode, "source", source)

	return writeYAMLNodeDocument(configFile, doc)
}

// RemoveTemplateConfigFile removes templates.<language>.<type> from the main
// config file, along with the language mapping if it becomes empty. It
// reports whether an entry was actually removed. Used by 'vip template remove'.
func RemoveTemplateConfigFile(configFile, language, projectType string) (bool, error) {
	doc, err := loadYAMLNodeDocument(configFile)
	if err != nil {
		return false, err
	}

	root, err := rootMappingNode(doc, configFile)
	if err != nil {
		return false, err
	}

	templatesNode := findMappingValue(root, "templates")
	if templatesNode == nil {
		return false, nil
	}

	languageNode := findMappingValue(templatesNode, language)
	if languageNode == nil {
		return false, nil
	}

	removed := deleteMappingKey(languageNode, projectType)
	if !removed {
		return false, nil
	}

	if len(languageNode.Content) == 0 {
		deleteMappingKey(templatesNode, language)
	}

	if err := writeYAMLNodeDocument(configFile, doc); err != nil {
		return false, err
	}

	return true, nil
}

// GetTemplate returns the configured template definition for language/type.
func (c *Configuration) GetTemplate(language, projectType string) *TemplateDefinition {
	if c == nil || c.Templates == nil {
		return nil
	}

	types := c.Templates[language]
	if types == nil {
		return nil
	}

	return types[projectType]
}

// GetTemplateSource returns the configured template source for language/type.
func (c *Configuration) GetTemplateSource(language, projectType string) string {
	definition := c.GetTemplate(language, projectType)
	if definition == nil {
		return ""
	}

	return definition.Source
}

// ApplyTeamTemplateOverrides overlays team config.yaml onto the loaded config.
func (c *Configuration) ApplyTeamTemplateOverrides(team, org string) error {
	if c == nil || team == "" {
		return nil
	}

	overrides, err := c.loadTeamOverrideConfigFile(team, org)
	if err != nil || overrides == nil {
		return err
	}

	for language, types := range overrides.Templates {
		if c.Templates[language] == nil {
			c.Templates[language] = make(map[string]*TemplateDefinition)
		}

		for projectType, override := range types {
			if override == nil {
				continue
			}
			c.Templates[language][projectType] = mergeTemplateDefinitions(c.Templates[language][projectType], override)
		}
	}

	return nil
}

// FindTeamTemplateConfigFile returns the path to the team's config.yaml,
// preferring one that already exists (see findTeamOverrideConfigFile). If none
// exists yet but the team's config directory itself does (e.g. set up via
// 'vip config init team' or 'vip config pull'), it returns the path where that
// config.yaml would be created. Returns "" if the team isn't configured at all
// (no matching team directory), or team is empty.
//
// Used by 'vip template add' to register new templates directly in the team's
// config.yaml when one exists, since team overrides always take precedence over
// personal config once merged -- registering there avoids the entry being
// silently shadowed.
func (c *Configuration) FindTeamTemplateConfigFile(team, org string) (string, error) {
	if c == nil || team == "" {
		return "", nil
	}

	if existing, err := c.findTeamOverrideConfigFile(team, org); err != nil {
		return "", err
	} else if existing != "" {
		return existing, nil
	}

	for _, candidate := range c.teamOverrideConfigCandidates(team, org) {
		dir := filepath.Dir(candidate)
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to stat %s: %w", dir, err)
		}
	}

	return "", nil
}

// GetTeamTemplateOverride returns the team's templates.<language>.<type> entry,
// if the team has a config.yaml overriding it. Returns nil (no error) if the
// team has no override config, or doesn't override that specific language/type.
//
// Unlike ApplyTeamTemplateOverrides, this doesn't merge/mutate c.Templates -- it's
// used to detect whether templates.<language>.<type> is already registered in the
// team's config.yaml (e.g. by 'vip template add' before writing there).
func (c *Configuration) GetTeamTemplateOverride(team, org, language, projectType string) (*TemplateDefinition, error) {
	if c == nil || team == "" {
		return nil, nil
	}

	overrides, err := c.loadTeamOverrideConfigFile(team, org)
	if err != nil || overrides == nil {
		return nil, err
	}

	types := overrides.Templates[language]
	if types == nil {
		return nil, nil
	}

	return types[projectType], nil
}

// loadTeamOverrideConfigFile locates and parses the team's config.yaml, if one
// exists. Returns nil, nil if no team override config file is present.
func (c *Configuration) loadTeamOverrideConfigFile(team, org string) (*teamOverrideConfigFile, error) {
	configFile, err := c.findTeamOverrideConfigFile(team, org)
	if err != nil {
		return nil, err
	}
	if configFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read team override config %s: %w", configFile, err)
	}

	var overrides teamOverrideConfigFile
	if err := yaml.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("failed to parse team override config %s: %w", configFile, err)
	}

	return &overrides, nil
}

func (c *Configuration) findTeamOverrideConfigFile(team, org string) (string, error) {
	candidates := c.teamOverrideConfigCandidates(team, org)
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to stat %s: %w", candidate, err)
		}
	}

	return "", nil
}

func (c *Configuration) teamOverrideConfigCandidates(team, org string) []string {
	candidates := []string{}
	add := func(path string) {
		for _, candidate := range candidates {
			if candidate == path {
				return
			}
		}
		candidates = append(candidates, path)
	}

	if org != "" {
		add(filepath.Join(c.GetTeamDir(team, org), "config.yaml"))
	}
	if c.DefaultOrganization != "" {
		add(filepath.Join(c.GetTeamDir(team, c.DefaultOrganization), "config.yaml"))
	}
	add(filepath.Join(c.GetTeamDir(team, ""), "config.yaml"))

	return candidates
}

func mergeTemplateDefinitions(base, override *TemplateDefinition) *TemplateDefinition {
	if base == nil {
		return cloneTemplateDefinition(override)
	}
	if override == nil {
		return cloneTemplateDefinition(base)
	}

	merged := cloneTemplateDefinition(base)
	if override.Source != "" {
		merged.Source = override.Source
	}
	if override.Hooks != nil {
		merged.Hooks = cloneTemplateHooks(override.Hooks)
	}

	return merged
}

func cloneTemplateDefinition(definition *TemplateDefinition) *TemplateDefinition {
	if definition == nil {
		return nil
	}

	return &TemplateDefinition{
		Source: definition.Source,
		Hooks:  cloneTemplateHooks(definition.Hooks),
	}
}

func cloneTemplateHooks(hooks *TemplateHooks) *TemplateHooks {
	if hooks == nil {
		return nil
	}

	cloned := &TemplateHooks{}
	if hooks.Post != nil {
		cloned.Post = &TemplateHookStage{}
		if hooks.Post.Install != nil {
			cloned.Post.Install = &PostInstallHook{
				Cmd:     append([]string(nil), hooks.Post.Install.Cmd...),
				Scripts: append([]string(nil), hooks.Post.Install.Scripts...),
			}
		}
	}

	return cloned
}
