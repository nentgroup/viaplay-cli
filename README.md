# vip

**vip** is the CLI for scaffolding very important projects, creating and configuring GitHub repositories, and applying 
team standards such as rulesets, secrets, and environments.

<p align="center">
  <img src="./.github/assets/vip.png" alt="vip screenshot" width="600">
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

### macOS

```bash
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask nentgroup/viaplay-cli/vip
```

### Linux

You can install `vip` on Linux in either of these ways:

#### Option 1: Install via Homebrew (Linuxbrew)

If you have [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux) installed:

```bash
brew tap nentgroup/viaplay-cli https://github.com/nentgroup/viaplay-cli
brew install --cask nentgroup/viaplay-cli/vip
```

#### Option 2: Install from the package manager

Download a package built by GoReleaser from the [GitHub Releases page](https://github.com/nentgroup/viaplay-cli/releases):

```bash
# Debian / Ubuntu
sudo dpkg -i vip_*.deb

# RPM-based distros
sudo rpm -i vip_*.rpm

# Alpine / APK-based distros
sudo apk add --allow-untrusted vip_*.apk
```

#### Option 3: Install from source or binary

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

Or download a release binary and place it on your PATH:

```bash
chmod +x ./vip
sudo install ./vip /usr/local/bin/vip
```

### Windows

You can install `vip` on Windows in either of these ways:

#### Option 1: Install via winget

```powershell
winget install nentgroup.viaplay-cli
```

#### Option 2: Download a release binary

Download the latest `vip.exe` from the [GitHub Releases page](https://github.com/nentgroup/viaplay-cli/releases), then place it in a folder already on your PATH.

#### Option 3: Build from source

```powershell
git clone https://github.com/nentgroup/viaplay-cli.git
cd viaplay-cli
go build -o vip.exe ./cmd/vip
Move-Item .\vip.exe "$env:USERPROFILE\bin\vip.exe"
```

### Cross-platform via Go

```bash
go install github.com/nentgroup/viaplay-cli/cmd/vip@latest
```

This works on macOS, Linux, and Windows when Go is installed and `$(go env GOPATH)/bin` (or `%USERPROFILE%\go\bin`) is on your PATH.

### Building from Source

When building from source, you must provide the GitHub OAuth client ID used by device auth:

```bash
export GITHUB_CLIENT_ID="your-client-id"
go build -o vip ./cmd/vip
```

On PowerShell:

```powershell
$env:GITHUB_CLIENT_ID = "your-client-id"
go build -o vip.exe .\cmd\vip
```

This value is required for `vip auth login` to work when you are not using a release build that embeds the client ID.

### macOS Security Bypass

When installing on macOS by downloading the binary directly, you may encounter security blocks. To bypass these:

```bash
chmod +x ./vip
xattr -dr com.apple.quarantine ./vip
sudo mv vip /usr/local/bin/
```

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-docs-dark.svg"><img src="./.github/assets/icon-docs.svg" alt="Documentation" width="18" height="18" aria-label="Documentation"></picture> Documentation

Full documentation is available at **[nentgroup.github.io/viaplay-cli](https://nentgroup.github.io/viaplay-cli)**

## <picture><source media="(prefers-color-scheme: dark)" srcset="./.github/assets/icon-dev-dark.svg"><img src="./.github/assets/icon-dev.svg" alt="Development" width="18" height="18" aria-label="Development"></picture> Development

### Prerequisites
- Go 1.27 or later
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

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for setup guidance, PR expectations, and contribution standards.

## Security

If you discover a security issue, please follow the disclosure process in [SECURITY.md](SECURITY.md). Please do not open a public issue for vulnerabilities.

## Code of conduct

Please review the [Code of Conduct](CODE_OF_CONDUCT.md) before participating in the project.

## License

This project is licensed under the [MIT License](LICENSE).
