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

	configFile, err := c.findTeamOverrideConfigFile(team, org)
	if err != nil {
		return err
	}
	if configFile == "" {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read team override config %s: %w", configFile, err)
	}

	var overrides teamOverrideConfigFile
	if err := yaml.Unmarshal(data, &overrides); err != nil {
		return fmt.Errorf("failed to parse team override config %s: %w", configFile, err)
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
