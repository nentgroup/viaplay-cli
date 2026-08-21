package cli

import (
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/config"
)

func loadConfigWithTeamOverrides(team, orgHint string) (*config.Configuration, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.ApplyTeamTemplateOverrides(team, resolveTeamConfigOrganization(cfg, orgHint)); err != nil {
		return nil, err
	}

	return cfg, nil
}

func resolveTeamConfigOrganization(cfg *config.Configuration, orgHint string) string {
	if cfg != nil && cfg.DefaultOrganization != "" {
		return cfg.DefaultOrganization
	}
	if organization := viper.GetString("github.organization"); organization != "" {
		return organization
	}
	return orgHint
}
