// Package config provides post-install hooks configuration
package config

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// PostInstallHook represents a hook to run after project scaffolding
type PostInstallHook struct {
	Run     []string `mapstructure:"run" yaml:"run"`
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
	return h.Run
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
	// Holds the hooks in order of execution (general to specific)
	var hooks []*PostInstallHook

	// Check if we have hooks configuration
	hooksConfig := viper.Get("hooks")
	if hooksConfig == nil {
		return nil
	}

	// Try to get post-install hooks
	postHooks := viper.GetStringMap("hooks.post.install")
	if len(postHooks) == 0 {
		return nil
	}

	// 1. Get general hooks that apply to all languages and project types
	generalRun := viper.GetStringSlice("hooks.post.install.run")
	if len(generalRun) > 0 {
		hooks = append(hooks, &PostInstallHook{Run: generalRun})
	}

	generalScripts := viper.GetStringSlice("hooks.post.install.scripts")
	if len(generalScripts) > 0 {
		if len(hooks) > 0 && hooks[0].Run != nil {
			// Add scripts to existing hook with commands
			hooks[0].Scripts = generalScripts
		} else {
			// Create a new hook for scripts
			hooks = append(hooks, &PostInstallHook{Scripts: generalScripts})
		}
	}

	// 2. Get language-specific hooks
	langKey := "hooks.post.install." + language

	// Get language-level commands
	langRun := viper.GetStringSlice(langKey + ".run")
	langScripts := viper.GetStringSlice(langKey + ".scripts")

	// Add language-level hook if we have commands or scripts
	if len(langRun) > 0 || len(langScripts) > 0 {
		hooks = append(hooks, &PostInstallHook{
			Run:     langRun,
			Scripts: langScripts,
		})
	}

	// 3. Get project-type specific hooks within this language
	projectTypeKey := langKey + "." + projectType

	// Get project-type level commands
	projectRun := viper.GetStringSlice(projectTypeKey + ".run")
	projectScripts := viper.GetStringSlice(projectTypeKey + ".scripts")

	// Add project-type level hook if we have commands or scripts
	if len(projectRun) > 0 || len(projectScripts) > 0 {
		hooks = append(hooks, &PostInstallHook{
			Run:     projectRun,
			Scripts: projectScripts,
		})
	}

	return hooks
}

// GetHooksDir returns the directory where global hook scripts are stored
func (c *Configuration) GetHooksDir() string {
	return filepath.Join(c.ConfigDir, "hooks")
}
