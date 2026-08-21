// Package registry provides template registry functionality
package registry

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/config"
)

// TemplateInfo represents metadata about a template
type TemplateInfo struct {
	// Unique identifier for the template (language/type)
	ID string
	// Programming language of the template
	Language string
	// Type of the template (service, cli, etc.)
	Type string
	// Description of the template
	Description string
	// Source URL or path for the template
	Source string
	// Tags for the template
	Tags []string
}

// Registry manages template discovery and metadata
type Registry struct {
	// Configuration
	Config *config.Configuration
	// Map of templates by ID (language/type)
	Templates map[string]*TemplateInfo
}

// NewRegistry creates a new template registry
func NewRegistry(cfg *config.Configuration) *Registry {
	if cfg == nil {
		// Use default configuration
		defaultCfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Warning: failed to load config: %v\n", err)
		}
		cfg = defaultCfg
	}

	return &Registry{
		Config:    cfg,
		Templates: make(map[string]*TemplateInfo),
	}
}

// LoadTemplates loads templates from configuration
func (r *Registry) LoadTemplates() error {
	// Load templates from config
	for language, types := range r.Config.Templates {
		for templateType, definition := range types {
			if definition == nil || definition.Source == "" {
				continue
			}
			templateID := fmt.Sprintf("%s/%s", language, templateType)

			r.Templates[templateID] = &TemplateInfo{
				ID:          templateID,
				Language:    language,
				Type:        templateType,
				Source:      definition.Source,
				Description: fmt.Sprintf("%s %s template", language, templateType),
				Tags:        []string{language, templateType},
			}
		}
	}

	return nil
}

// GetTemplate returns template info for a specific language and type
func (r *Registry) GetTemplate(language, templateType string) (*TemplateInfo, error) {
	templateID := fmt.Sprintf("%s/%s", language, templateType)

	// Check if the template exists in the registry
	if template, exists := r.Templates[templateID]; exists {
		return template, nil
	}

	// If not loaded yet, try to load templates
	if len(r.Templates) == 0 {
		if err := r.LoadTemplates(); err != nil {
			return nil, err
		}

		// Check again after loading
		if template, exists := r.Templates[templateID]; exists {
			return template, nil
		}
	}

	return nil, fmt.Errorf("template not found for %s/%s", language, templateType)
}

// GetAllTemplates returns all available templates
func (r *Registry) GetAllTemplates() []*TemplateInfo {
	// Load templates if not already loaded
	if len(r.Templates) == 0 {
		if err := r.LoadTemplates(); err != nil {
			fmt.Printf("Warning: failed to load templates: %v\n", err)
		}
	}

	// Convert map to slice
	templates := make([]*TemplateInfo, 0, len(r.Templates))
	for _, template := range r.Templates {
		templates = append(templates, template)
	}

	return templates
}

// GetTemplatesByLanguage returns all templates for a specific language
func (r *Registry) GetTemplatesByLanguage(language string) []*TemplateInfo {
	// Load templates if not already loaded
	if len(r.Templates) == 0 {
		if err := r.LoadTemplates(); err != nil {
			fmt.Printf("Warning: failed to load templates: %v\n", err)
		}
	}

	// Filter templates by language
	var templates []*TemplateInfo
	for _, template := range r.Templates {
		if template.Language == language {
			templates = append(templates, template)
		}
	}

	return templates
}

// RegisterTemplate adds a template to the registry
func (r *Registry) RegisterTemplate(template *TemplateInfo) error {
	// Validate template
	if template.Language == "" || template.Type == "" || template.Source == "" {
		return fmt.Errorf("invalid template: language, type, and source are required")
	}

	// Set ID if not already set
	if template.ID == "" {
		template.ID = fmt.Sprintf("%s/%s", template.Language, template.Type)
	}

	// Add to registry
	r.Templates[template.ID] = template

	// Update configuration
	if r.Config.Templates == nil {
		r.Config.Templates = make(map[string]map[string]*config.TemplateDefinition)
	}

	if r.Config.Templates[template.Language] == nil {
		r.Config.Templates[template.Language] = make(map[string]*config.TemplateDefinition)
	}

	r.Config.Templates[template.Language][template.Type] = &config.TemplateDefinition{Source: template.Source}

	// Update viper configuration
	viperKey := fmt.Sprintf("templates.%s.%s.source", template.Language, template.Type)
	viper.Set(viperKey, template.Source)

	return nil
}

// UnregisterTemplate removes a template from the registry
func (r *Registry) UnregisterTemplate(language, templateType string) error {
	templateID := fmt.Sprintf("%s/%s", language, templateType)

	// Remove from registry
	delete(r.Templates, templateID)

	// Remove from configuration
	if r.Config.Templates != nil && r.Config.Templates[language] != nil {
		delete(r.Config.Templates[language], templateType)

		// Clean up empty language map
		if len(r.Config.Templates[language]) == 0 {
			delete(r.Config.Templates, language)
		}
	}

	// Update viper configuration
	viperKey := fmt.Sprintf("templates.%s.%s.source", language, templateType)
	viper.Set(viperKey, nil)

	return nil
}

// SaveTemplates saves the template registry to the configuration file
func (r *Registry) SaveTemplates() error {
	// Write to viper config
	for _, template := range r.Templates {
		viperKey := fmt.Sprintf("templates.%s.%s.source", template.Language, template.Type)
		viper.Set(viperKey, template.Source)
	}

	// Write viper config to file
	return viper.WriteConfig()
}
