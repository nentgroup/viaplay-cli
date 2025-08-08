// Package config provides post-install hooks configuration
package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// PostInstallHook represents a hook to run after project scaffolding
type PostInstallHook struct {
	Cmd     []string `mapstructure:"cmd" yaml:"cmd"`
	Scripts []string `mapstructure:"scripts" yaml:"scripts"`
}

// ProjectTypeHooks represents hooks for specific project types within a language
type ProjectTypeHooks map[string]*PostInstallHook

// LanguageHooks represents hooks for a specific language, including general language hooks
// and project-type specific hooks
type LanguageHooks struct {
	// General hooks for this language
	*PostInstallHook `mapstructure:",squash" yaml:",inline"`
	// Project-type specific hooks
	ProjectTypes ProjectTypeHooks `mapstructure:",remain" yaml:",inline"`
}

// PostInstallHooks represents all post-install hooks configuration
type PostInstallHooks struct {
	// General hooks for all languages and project types
	*PostInstallHook `mapstructure:",squash" yaml:",inline"`
	// Language-specific hooks
	Languages map[string]*LanguageHooks `mapstructure:",remain" yaml:",inline"`
}

// GetAllCommands returns all commands to run as a slice
func (h *PostInstallHook) GetAllCommands() []string {
	if h == nil {
		return nil
	}
	return h.Cmd
}

// GetAllScripts returns all script paths to run as a slice
func (h *PostInstallHook) GetAllScripts() []string {
	if h == nil {
		return nil
	}
	return h.Scripts
}

// GetPostInstallHooks returns the post-install hooks for a specific language and project type
func (c *Configuration) GetPostInstallHooks(language, projectType string) []*PostInstallHook {
	// Holds the hooks in order of execution
	var hooks []*PostInstallHook

	// Check for template-specific hooks
	// This is the approach where hooks are defined directly in the template configuration
	templateKey := fmt.Sprintf("templates.%s.%s", language, projectType)

	// Get commands using 'cmd' key
	templateHooksCmd := viper.GetStringSlice(templateKey + ".hooks.post.install.cmd")

	// Get scripts (unchanged)
	templateHooksScripts := viper.GetStringSlice(templateKey + ".hooks.post.install.scripts")

	// If we found template-specific hooks, use them
	if len(templateHooksCmd) > 0 || len(templateHooksScripts) > 0 {
		hooks = append(hooks, &PostInstallHook{
			Cmd:     templateHooksCmd,
			Scripts: templateHooksScripts,
		})
		return hooks
	}

	// No hooks found
	return nil
}

// GetHooksDir returns the directory where global hook scripts are stored
func (c *Configuration) GetHooksDir() string {
	return filepath.Join(c.ConfigDir, "hooks")
}
