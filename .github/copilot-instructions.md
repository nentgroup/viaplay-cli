# Copilot Instructions for viaplay-cli

`viaplay-cli` (binary `vip`) is a Go/Cobra CLI that scaffolds projects from templates,
creates/configures GitHub repositories, and applies team standards (rulesets, secrets,
environments) for Viaplay teams.

## Build, test, lint

Prefer [Task](https://taskfile.dev) (`Taskfile.yml`); raw `go`/`golangci-lint`
equivalents also work.

```bash
task deps:install        # go mod download && go mod tidy
task build               # cross-compile to bin/ for linux/windows/darwin (amd64/arm64)
task vip -- <args>       # run the CLI: go run ./cmd/vip <args>
task lint                # golangci-lint run --allow-parallel-runners --fix; actionlint
task test:coverage       # go test -tags=test -race -coverprofile=coverage/coverage.out ./...
```

Run a single test (tests use the `test` build tag, so it must be passed):
```bash
go test -tags=test ./internal/config/... -run TestSourceLoad -v
go test -tags=test ./internal/project/... -run TestName/SubTest -v -race
```

Lint config is `.golangci.yml` (gofumpt/goimports formatters, local prefix
`github.com/nentgroup/viaplay-cli`, cyclomatic complexity max 20, line length 120,
UK-locale misspell). Git hooks (lefthook) run golangci-lint + actionlint on
pre-commit, the full race-enabled test suite on pre-push, and commitlint
(Conventional Commits, e.g. `feat(scope): ...`, header 10–100 chars) on commit-msg.

## Architecture

### Command layer → Factory → subsystems
`cmd/vip/main.go` calls `internal/cli.Execute()`. All Cobra commands are registered
centrally in `internal/cli/root.go`'s `init()` (`auth`, `secrets`, `cache`, `config`,
`hooks`, `version`, `project`, `repo`, `template`). Commands under `internal/cli/`
are thin: they parse flags into an `Options` struct and delegate to
`internal/project.Factory`, which orchestrates the actual workflow.

`project.Factory` (internal/project/factory.go) wires together the subsystems needed
for project/repo operations:
- `internal/gh.GitHubClient` — GitHub API access (repos, envs, rulesets, secrets, teams)
- `internal/config.Configuration` — loaded settings, team/template mappings, hooks
- `internal/registry.Registry` — maps `language/type` → template source, backed by
  `Config.Templates` (also persisted back to the viper config file)
- `internal/cache.Manager` — caches downloaded template sources
- `internal/scaffolding.ProjectScaffolder` — applies a template to a new project dir
- `internal/output.Reporter` — progress/status reporting interface (Start/Progress/
  Complete/Failed/Skip/Warning/Info) implemented for terminal output

`Factory.Create` runs a fixed pipeline via an internal `creationContext`: resolve
account type from the GitHub user → populate authenticated user → prepare template
vars → scaffold (if needed) → create repo (if needed) → configure GitHub
(envs/rulesets/secrets) → run post-install hooks → publish summary. Each stage
appends to a `Summary` struct returned to the caller even on partial failure, and
`creationContext.Cleanup` removes created repo/dir on error unless
`opts.CleanupOnError` is false.

`Factory` also exposes standalone `Apply*` methods (`ApplyConfigurations`,
`ApplyEnvs`, `ApplyRulesets`, `ApplySecrets`, `ApplyRepoSecrets`) used by
`vip repo apply ...` to push team config to an *existing* repo without going through
project creation.

### Config-driven behaviour (blueprints + team configs)
`internal/config/blueprints/*.yaml` are `//go:embed`ded default templates
(`config.yaml`, `environment.yaml`, `ruleset.yaml`, `secrets.yaml`,
`team-config.yaml`) used to scaffold a user's `~/.config/viaplay` tree and per-team
config on first run (`internal/config/setup.go`). Team-specific overrides live under
`ConfigDir/teams/<team>/` and are what `Apply*` methods read at runtime.

A shared, git-backed config source can be configured (`internal/config/source.go`,
`SourceConfig`: repository/branch/root) so an org can distribute/pull shared team
configs via `internal/git` instead of maintaining them locally.

### Templates and hooks
`internal/template.Renderer`/`Variables` (`pkg/tmpl`) render template files with
project metadata (name, team, creator, etc.) during scaffolding.
`internal/config.PostInstallHook` (`Cmd` and/or `Scripts`) defines post-scaffold
commands, resolved per language/project-type via
`Configuration.GetPostInstallHooks`, with script paths resolved against
`ConfigDir/hooks` unless absolute. Hook output is streamed through
`output.HookOutputWriter` (bordered, titled panel) rather than raw stdout.

### Auth and secrets
`internal/gh/auth.go` implements the GitHub OAuth **device flow** (no client secret);
the client ID comes from `GITHUB_CLIENT_ID` env var or an `-ldflags -X`-embedded
build-time value. Tokens are stored/retrieved via the OS keyring
(`zalando/go-keyring`, namespace `viaplaycli`) — never persisted to disk/config
files. `internal/gh/encryption.go` encrypts secret values with a repo's GitHub public
key (libsodium/nacl box) before pushing via the Secrets API, matching GitHub's
Actions secrets requirements.

## Conventions

- Two `go-github` major versions are vendored side by side (`v74` and `v91`,
  see `go.mod`) — check which one a file already imports before adding GitHub API
  calls; don't assume there's only one.
- User-visible progress must go through `internal/output.Reporter` /
  `output.*Message` helpers (colour-coded, icon-prefixed) rather than raw
  `fmt.Println`, so behaviour stays consistent across commands and is testable via
  reporter fakes.
- Long-running/multi-step operations (project creation, `Apply*`) build a `Summary`
  and accumulate `Errors` rather than failing fast, so partial success can be
  reported and cleaned up.
- Config structs use both `mapstructure` (viper) and `yaml` tags since the same
  types are decoded from viper config and read/written as standalone YAML
  (blueprints, team configs).
- Update `docs/` (docsify site) when changing user-facing CLI behavior — it's the
  published documentation, not just README.
