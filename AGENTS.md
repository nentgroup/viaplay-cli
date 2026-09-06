# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project overview

`viaplay-cli` (binary name `vip`) is a Go CLI for scaffolding projects, creating and
configuring GitHub repositories, and applying team standards (rulesets, secrets,
environments). Built with Cobra/Viper, uses the system keyring for auth token/secret
storage, and integrates with the GitHub API via `google/go-github`.

## Repository layout

- `cmd/vip` — main entrypoint
- `internal/cli` — Cobra command definitions (`auth`, `project`, `repo`, `secrets`,
  `template`, `config`, `cache`, etc.)
- `internal/config` — configuration, blueprints, and template sources
- `internal/project`, `internal/scaffolding`, `internal/template` — project creation
  and templating logic
- `internal/gh` — GitHub API client/helpers
- `internal/git`, `internal/secrets`, `internal/cache`, `internal/registry`,
  `internal/output` — supporting subsystems
- `pkg/tmpl`, `pkg/paths` — reusable packages
- `docs` — user-facing documentation (docsify site)

## Build, test, and lint commands

This project uses [Task](https://taskfile.dev) (`Taskfile.yml`). Prefer these over
raw `go` invocations when available:

- `task build` — cross-compile binaries into `bin/`
- `task deps:install` — `go mod download && go mod tidy`
- `task lint` — `golangci-lint run --allow-parallel-runners --fix` + `actionlint`
- `task test:coverage` — run tests with coverage (`go test -tags=test -race ...`)
- `task vip -- <args>` — run the CLI via `go run ./cmd/vip`
- `task install:local` — `go install ./cmd/vip`
- `task clean` — remove build/coverage artifacts

Equivalent raw commands (if Task is unavailable):
- Build: `go build -o vip ./cmd/vip`
- Test: `go test -tags=test -v ./... -race -coverprofile=coverage/coverage.out`
- Lint: `golangci-lint run --allow-parallel-runners --fix`

Run the smallest targeted test/build/lint that covers your change; only run the full
suite when necessary.

## Code style

- Formatters: `gofumpt` and `goimports` (local prefix
  `github.com/nentgroup/viaplay-cli`) — run via `golangci-lint run --fix` or your
  editor's format-on-save.
- Linters enabled include `bodyclose`, `cyclop`, `dupl`, `errorlint`, `goconst`,
  `gosec`, `misspell` (UK locale), `nestif`, `noctx`, `revive`, `staticcheck`,
  `whitespace`, `ireturn`, `errname`. See `.golangci.yml` for full config
  (max cyclomatic complexity 20, dupl threshold 250, line length 120).
- Follow existing package conventions (exported symbols documented per `revive`
  `exported` rule).

## Commit conventions

Commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
(enforced by commitlint via lefthook on `commit-msg`):
```
feat(scope): add feature
fix(scope): fix bug
```
Header length: 10–100 chars.

## Git hooks (lefthook)

- `pre-commit`: `golangci-lint run --allow-parallel-runners --fix` (on `*.go`) and
  `actionlint` (on `.github/workflows/*.yml`)
- `pre-push`: `go test -v ./... -covermode=atomic -coverpkg=./... -race`
- `commit-msg`: `commitlint lint`

Run `task setup:lefthook` to install hooks locally.

## Notes for agents

- Don't fix unrelated pre-existing lint/test issues while making a change.
- Update `docs/` when changing user-facing CLI behavior.
- Avoid committing secrets/tokens; this tool interacts with the system keyring and
  GitHub tokens — never hardcode credentials in code or tests.

## Codebase navigation

This repository is indexed with graphify-rs.

For questions involving codebase structure, dependencies, call relationships,
architecture, or locating relevant implementation:

1. Use `graphify-rs query "<question>"` from the repository root first.
2. Use Graphify results to identify relevant symbols and files.
3. Inspect the actual source files before making changes or drawing conclusions.
4. Prefer Graphify for broad codebase discovery instead of repeatedly using
   grep/find/glob across the entire repository.
5. Graphify's generated data is stored in its default per-project location
   under `~/.graphify-rs/`; do not assume generated graph files are in this repo.

After significant structural code changes, refresh the graph with the
appropriate graphify-rs update/build command.

`.graphifyignore` defines paths that should not be indexed.