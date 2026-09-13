# Contributing to vip

Thanks for your interest in contributing to vip.

We welcome issues, feature ideas, bug fixes, docs improvements, and template contributions.

## Code of conduct

By participating in this project, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Reporting issues

Before opening a new issue, please:

- search existing issues to avoid duplicates
- include reproduction steps for bugs
- include the exact command, expected result, and actual result
- include the relevant version of the CLI and your OS

## Development setup

Requirements:

- Go 1.27+
- Task (optional, but recommended)
- `golangci-lint`

Quick start:

```bash
go build -o vip ./cmd/vip
go test ./...
golangci-lint run --allow-parallel-runners --fix
```

Or with Task:

```bash
task build
task test:coverage
task lint
```

## Branching and pull requests

- create a feature branch from `main`
- keep changes focused and small
- write tests for behavior changes
- run the relevant validation commands before opening a PR
- keep commit messages in Conventional Commits format

## Pull request checklist

Before opening a PR, confirm:

- code builds successfully
- relevant tests pass
- `golangci-lint` passes
- docs are updated for user-facing changes
- commit messages follow Conventional Commits

## Contribution types

### Bug fixes

Include:

- root cause summary
- reproduction steps
- affected commands or files
- validation performed

### Features

Before implementing a feature, open or reference an issue describing the problem and expected behavior. This keeps scope clear and makes review easier.

### Documentation

We welcome improvements to docs, examples, and template guidance. If a change affects how users work with the CLI, update `docs/` or the relevant README pages.

## Template contributions

If you are contributing a reusable template:

- include a `.vip.yaml` manifest when possible
- keep the template self-explanatory and consistent with existing structures
- document required variables and assumptions
- validate the template with the local CLI commands before submitting

## Security disclosures

Please do not open public issues for security vulnerabilities. Follow the steps in [SECURITY.md](SECURITY.md).
