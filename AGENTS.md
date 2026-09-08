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

### graphify-rs knowledge graph

This repo has a `graphify-rs-out/` knowledge graph (nodes/edges/communities extracted from the
codebase) used to answer architecture questions and speed up navigation. Key files:
`graphify-rs-out/graph.json` (GraphRAG-ready data), `graphify-rs-out/GRAPH_REPORT.md`
(god nodes, surprising connections, suggested questions), `graphify-rs-out/graph.html`
(interactive visualization).

- **Querying**: `graphify-rs query "<question>" --graph graphify-rs-out/graph.json` (add `--dfs`
  to trace a specific path instead of broad BFS context). Prefer this over re-reading the whole
  tree when you need architecture/relationship context.
- **After modifying code**: rebuild so the graph doesn't go stale before answering questions about
  the changed code. Batch changes, then rebuild once (not after every one-line edit):
  ```console
  graphify-rs build --path . --output graphify-rs-out --no-llm --update
  ```
  `--update` only re-extracts changed files (SHA256 cache); `--no-llm` keeps it AST-only, free,
  and fast (~2-5s).
- **Stats/diff**: `graphify-rs stats graphify-rs-out/graph.json` and
  `graphify-rs diff <old-graph.json> <new-graph.json>` for comparing snapshots.
- Treat edge confidence honestly: edges are tagged EXTRACTED, INFERRED, or AMBIGUOUS in the
  graph — don't present INFERRED/AMBIGUOUS relationships as verified facts.