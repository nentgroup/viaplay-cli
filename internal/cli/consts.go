package cli

// Shared string constants used across cli commands to avoid goconst violations.
const (
	cmdList            = "list"
	colType            = "Type"
	pathHooks          = "hooks"
	configTargetUser   = "user"
	flagOwner          = "owner"
	defaultProjectName = "project"
	yamlExt            = ".yaml"

	scopeEnvs        = "envs"
	scopeRulesets    = "rulesets"
	scopeSecrets     = "secrets"
	scopeVariables   = "variables"
	scopeRepoSecrets = "repo-secrets"
)
