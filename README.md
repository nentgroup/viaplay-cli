# viaplay-cli

**viaplay-cli** is a developer tool for quickly scaffolding projects, creating and configuring GitHub repositories, and applying team or organization standards such as rulesets, secrets, and environments.

---

## ⚡ What is viaplay-cli?

viaplay-cli is a CLI tool for developers to:
- Scaffold new projects from templates (Go, TypeScript, etc.)
- Create and configure GitHub repositories (personal or organization)
- Apply team/organization settings (rulesets, secrets, environments)
- Manage secrets securely using your system keyring

---

## 🛠️ Building & Developing

### Prerequisites
- Go 1.21 or later
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

---

## 🚀 Quick Examples

Create a new project:

```bash
./vip create project --name my-service --language go --type service --team platform
```

Create a repository only:

```bash
./vip create repo --name my-repo --team platform
```

Authenticate with GitHub:

```bash
./vip auth login
```

---

## 📚 Usage & Documentation

- [Full CLI Usage & Commands](docs/cli/index.md)
- [Project Creation Guide](docs/project-creation.md)
- [Configuration Reference](docs/configuration.md)
- [Templates & Teams](docs/teams-and-templates.md)

---

