# viaplay-cli

**viaplay-cli** is a command-line tool for scaffolding projects, creating and configuring GitHub repositories, and applying team standards such as rulesets, secrets, and environments.

<p align="center">
  <img src="./.github/assets/vip.png" alt="viaplay-cli screenshot" width="600">
</p>

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-features-dark.svg"><img src="./.github/assets/icon-features.svg" alt="Features" width="18" height="18" aria-label="Features"></picture> Features

- **Project templates**: Generate projects from templates (Go, TypeScript, Rust, and more)
- **GitHub integration**: Create and configure repositories (personal or organization)
- **Team configurations**: Apply environments, rulesets, and secrets
- **Secrets management**: Store secret values in the system keyring
- **Extensible architecture**: Add new templates and project types

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-quickstart-dark.svg"><img src="./.github/assets/icon-quickstart.svg" alt="Quick Start" width="18" height="18" aria-label="Quick Start"></picture> Quick Start

### Authentication

First, authenticate with GitHub:

```bash
vip auth login
```

This opens your browser for GitHub device authentication and stores the token in your system keyring.

### Create a New Project

```bash
# Create a new Go service
vip project create my-service --language go --type service --team platform

# Create a TypeScript service
vip project create my-lib --language typescript --type service --team frontend
```

### Create a Repository Only

```bash
vip repo create my-repo --team platform --description "My awesome repository"
```

### Check Your Authentication Status

```bash
vip auth status
```

> **Available templates:** [Go service](https://github.com/nentgroup/go-service-template) and [Node service](https://github.com/nentgroup/node-service-template). See [Templates](https://nentgroup.github.io/viaplay-cli/#/templates) for details on using custom templates.

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-install-dark.svg"><img src="./.github/assets/icon-install.svg" alt="Install" width="18" height="18" aria-label="Install"></picture> Installation

### Homebrew (macOS & Linux)

```bash
# Set your GitHub token for accessing private repositories
export HOMEBREW_GITHUB_API_TOKEN=your_github_token

# Tap the repository and install
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask vip
```

> **Note:** The `HOMEBREW_GITHUB_API_TOKEN` is required to access private repositories. You can generate a token with the `repo` scope at [GitHub Settings > Developer Settings > Personal Access Tokens](https://github.com/settings/tokens).

### Manual Installation

#### Download Binaries

You can download pre-built binaries from the [releases page](https://github.com/nentgroup/viaplay-cli/releases).

##### macOS Security Bypass

When installing on macOS by downloading the binary directly, you may encounter security blocks. To bypass these:

```bash
# After downloading the binary
chmod +x ./vip

# Remove the quarantine attribute
xattr -dr com.apple.quarantine ./vip

# Now you can move it to your PATH
sudo mv vip /usr/local/bin/
```

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli

# Build the binary
go build -o vip ./cmd/vip

# Optionally, move to a directory in your PATH
sudo mv vip /usr/local/bin/
```

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-docs-dark.svg"><img src="./.github/assets/icon-docs.svg" alt="Documentation" width="18" height="18" aria-label="Documentation"></picture> Documentation

Full documentation is available at **[nentgroup.github.io/viaplay-cli](https://nentgroup.github.io/viaplay-cli)**

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-dev-dark.svg"><img src="./.github/assets/icon-dev.svg" alt="Development" width="18" height="18" aria-label="Development"></picture> Development

### Prerequisites
- Go 1.25 or later
- [Task](https://taskfile.dev/) (optional, for easier scripts)

### Build the CLI

```bash
go build -o vip ./cmd/vip
```

### Run the CLI (development)

```bash
go run ./cmd/vip --help
```

### Lint & Format

```bash
golangci-lint run --fix
```

### Run Tests

```bash
go test ./...
```


## License

This project is licensed under the [MIT License](LICENSE).
