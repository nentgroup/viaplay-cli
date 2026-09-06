package template

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest represents an optional template manifest.
type Manifest struct {
	Schema    int                `yaml:"schema" json:"schema"`
	Options   []ManifestOption   `yaml:"options" json:"options,omitempty"`
	Files     ManifestFiles      `yaml:"files" json:"files,omitempty"`
	Metadata  ManifestMetadata   `yaml:"metadata" json:"metadata,omitempty"`
	Variables []ManifestVariable `yaml:"variables" json:"variables,omitempty"`
	Hooks     ManifestHooks      `yaml:"hooks" json:"hooks,omitempty"`
}

// IsSupported reports whether the manifest schema is supported.
func (m Manifest) IsSupported() bool { return m.Schema <= 2 }

// ManifestMetadata contains optional descriptive metadata.
type ManifestMetadata struct {
	Name        string `yaml:"name" json:"name,omitempty"`
	Description string `yaml:"description" json:"description,omitempty"`
	Version     string `yaml:"version" json:"version,omitempty"`
}

// ManifestOption defines a user-selectable template option.
type ManifestOption struct {
	Key         string           `yaml:"key" json:"key"`
	Type        string           `yaml:"type" json:"type"`
	Prompt      string           `yaml:"prompt" json:"prompt,omitempty"`
	Description string           `yaml:"description" json:"description,omitempty"`
	Default     any              `yaml:"default" json:"default,omitempty"`
	Required    bool             `yaml:"required" json:"required,omitempty"`
	Choices     []ManifestChoice `yaml:"choices" json:"choices,omitempty"`
}

// ManifestVariable defines a free-form input field.
type ManifestVariable struct {
	Key         string           `yaml:"key" json:"key"`
	Type        string           `yaml:"type" json:"type"`
	Prompt      string           `yaml:"prompt" json:"prompt,omitempty"`
	Description string           `yaml:"description" json:"description,omitempty"`
	Default     any              `yaml:"default" json:"default,omitempty"`
	Required    bool             `yaml:"required" json:"required,omitempty"`
	Validate    ManifestValidate `yaml:"validate" json:"validate,omitempty"`
}

// ManifestChoice defines a selectable option choice.
type ManifestChoice struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label" json:"label,omitempty"`
}

// ManifestFiles defines file include/exclude rules.
type ManifestFiles struct {
	Include []ManifestFileRule `yaml:"include" json:"include,omitempty"`
	Exclude []ManifestFileRule `yaml:"exclude" json:"exclude,omitempty"`
}

// ManifestFileRule defines a file path rule.
type ManifestFileRule struct {
	Path string `yaml:"path" json:"path"`
	When string `yaml:"when" json:"when,omitempty"`
}

// ManifestHooks defines hook commands.
type ManifestHooks struct {
	Post []ManifestHook `yaml:"post" json:"post,omitempty"`
}

// ManifestHook defines a post-scaffold command.
type ManifestHook struct {
	Name string `yaml:"name" json:"name,omitempty"`
	Run  string `yaml:"run" json:"run"`
	When string `yaml:"when" json:"when,omitempty"`
}

// ManifestValidate defines validation rules.
type ManifestValidate struct {
	Pattern string `yaml:"pattern" json:"pattern,omitempty"`
	Message string `yaml:"message" json:"message,omitempty"`
}

// LoadManifest loads template.yaml from the template root if present.
func LoadManifest(templateRoot string) (*Manifest, error) {
	path := filepath.Join(templateRoot, "template.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read template manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse template manifest: %w", err)
	}
	if manifest.Schema == 0 {
		manifest.Schema = 1
	}
	return &manifest, nil
}
