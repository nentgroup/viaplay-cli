package template

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest represents an optional template manifest.
type Manifest struct {
	// Schema is the manifest format version (see docs/templates.md#manifest-schema-versioning
	// for the user-facing behaviour). Maintainer guidance for vip contributors on
	// when to bump this: only for changes that could make an older vip
	// misinterpret the manifest (e.g. repurposing an existing field's meaning,
	// changing a field's type, or making previously-optional structure required).
	// Purely additive, optional changes (e.g. metadata.language/metadata.type)
	// don't require a bump: older vip releases ignore unrecognised keys, and
	// newer vip releases parsing an older manifest without them see empty defaults.
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
	// Language is the programming language this template targets (e.g. "go",
	// "typescript"), used by 'vip template add' to register the template
	// under 'templates.<language>.<type>' without requiring --language.
	Language string `yaml:"language" json:"language,omitempty"`
	// Type is the project type this template produces (e.g. "service",
	// "lambda", "cli"), used by 'vip template add' alongside Language.
	Type string `yaml:"type" json:"type,omitempty"`
	// Authors lists the template's maintainers. Purely informational; shown
	// in 'template show'/'template list'.
	Authors []ManifestAuthor `yaml:"authors" json:"authors,omitempty"`
}

// ManifestAuthor identifies a template maintainer.
type ManifestAuthor struct {
	Name string `yaml:"name" json:"name,omitempty"`
	// Email is validated (RFC 5322 address) by ValidateManifest when present.
	Email string `yaml:"email" json:"email,omitempty"`
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
//
// NOTE: reserved for future use. Not currently read/executed anywhere during
// scaffolding -- declaring hooks in a manifest has no effect today. Executing
// commands declared by a template's own manifest (which may come from an
// untrusted/arbitrary source, e.g. via 'template test'/'template add') is a
// deliberate design gap, not an oversight: it would let any template source
// run arbitrary commands on the user's machine without explicit opt-in. Use
// config.PostInstallHook (templates.<language>.<type>.hooks in the CLI
// config, defined by the user/team, not the template author) for working
// post-scaffold automation.
type ManifestHooks struct {
	Post []ManifestHook `yaml:"post" json:"post,omitempty"`
}

// ManifestHook defines a post-scaffold command. See ManifestHooks: not
// currently executed by vip.
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

// ManifestFileNames lists the manifest filenames vip looks for in a template's
// root directory, in priority order. template.yaml is only supported as a
// deprecated fallback for existing templates; new templates should use
// .vip.yaml or .vip.yml.
var ManifestFileNames = []string{".vip.yaml", ".vip.yml", "template.yaml"}

// FindManifestPath returns the path to the first manifest file found in dir
// (per ManifestFileNames), and whether one was found.
func FindManifestPath(dir string) (string, bool) {
	for _, name := range ManifestFileNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// LoadManifest loads the template manifest (.vip.yaml, .vip.yml, or the
// deprecated template.yaml) from the template root, if present.
func LoadManifest(templateRoot string) (*Manifest, error) {
	path, found := FindManifestPath(templateRoot)
	if !found {
		return nil, nil
	}
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
