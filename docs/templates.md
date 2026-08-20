# Templates

viaplay-cli uses project templates to scaffold new repositories. Templates can be local directories or remote Git repositories. During project creation, template files are copied and variables are replaced with values you provide.

---

## Available Templates

The following official templates are available out of the box:

| Language | Type | Repository |
|----------|------|------------|
| Go | service | [go-service-template](https://github.com/nentgroup/go-service-template) |
| Node | service | [node-service-template](https://github.com/nentgroup/node-service-template) |

More templates are planned. You can also add your own templates in your `config.yaml` under the `templates` section — see [Configuration](configuration.md) for details.

> **Built a template?** If you've created a reusable template that could benefit others, open a PR to add it to this list!

---

## Template Sources

- **Remote Git:** Use a Git URL (e.g., `git@github.com:nentgroup/go-service-template.git`).
- **Local Directory:** Use a local path for custom templates.

### Using a Custom Template Source

Specify the template source when creating a project:

```bash
# Remote template
vip project create my-service --language go --type service --template-source git@github.com:your-org/your-template.git

# Local template
vip project create my-service --language go --type service --template-source ~/my-templates/go-service
```

---

## Template Structure

A template typically contains:
- Project files (README, source code, configs)
- Placeholders for variables using `{{ .Namespace.VarName }}` syntax

### Raw Files (no rendering)

Add the `.raw` suffix to any template file you want copied verbatim. The file is copied as-is and the `.raw` suffix is stripped in the generated project.

Example: `template.go.tmpl.raw` is copied to `template.go.tmpl` without any rendering.

---

## Template Caching

Remote templates are cached locally for faster reuse. Use `vip template list`, `vip template info`, `vip template update`, `vip template prune`, or `vip template clean` to manage those local copies, or pass `--no-cache` when creating a project to force a fresh download.

---

## Template Variables

Template variables are replaced with actual values during project scaffolding. Use the syntax `{{ .Namespace.VarName }}` (double braces, leading dot) in your template files.

Variables can be used in:
- **File content** (e.g., `{{ .Project.Name }}`)
- **Filenames** (e.g., `{{.Project.Name}}.md`, `{{.Service.Name}}-config.yaml`)
- **Folder names** (e.g., `src/{{.Service.Name}}`, `{{kebab .Project.Name}}/lib`)

### Project Information
- `{{ .Project.Name }}` — Name of the project/repository
- `{{ .Project.Description }}` — Description of the project
- `{{ .Project.Type }}` — Type of project (e.g., "api", "library", "app")
- `{{ .Project.Language }}` — Programming language (e.g., "go", "typescript", "python")
- `{{ .Project.License }}` — License type (e.g., "MIT", "Apache-2.0")

### Repository Information
- `{{ .Repo.Owner }}` — GitHub username or organization name
- `{{ .Repo.Name }}` — Repository name
- `{{ .Repo.URL }}` — Full GitHub repository URL
- `{{ .Repo.SSHURL }}` — SSH URL for the repository
- `{{ .Repo.IsPrivate }}` — Whether the repository is private

### Service Information
- `{{ .Service.Name }}` — Name of the service
- `{{ .Service.Owner }}` — Owner/team responsible for the service
- `{{ .Service.OwnerKey }}` — Key identifier for the service owner
- `{{ .Service.Port }}` — Port the service listens on
- `{{ .Service.Type }}` — Type of service (e.g., "http", "grpc", "worker")

### Go-specific Variables
- `{{ .Go.BinaryName }}` — Name of the compiled binary
- `{{ .Go.ModulePath }}` — Go module path (e.g., "github.com/org/service")
- `{{ .Go.Version }}` — Go version used (e.g., "1.21")

### Rust-specific Variables
- `{{ .Rust.BinaryName }}` — Name of the compiled binary
- `{{ .Rust.CargoName }}` — Name in Cargo.toml (often uses underscores instead of dashes)
- `{{ .Rust.Version }}` — Rust version used (e.g., "1.75")
- `{{ .Rust.Edition }}` — Rust edition (e.g., "2021")

### AWS Lambda-specific Variables
- `{{ .Lambda.FunctionName }}` — Name of the Lambda function
- `{{ .Lambda.Handler }}` — Handler path (e.g., "index.handler")
- `{{ .Lambda.Runtime }}` — Lambda runtime (e.g., "nodejs18.x", "go1.x", "python3.9")
- `{{ .Lambda.Timeout }}` — Timeout in seconds
- `{{ .Lambda.MemorySize }}` — Memory size in MB
- `{{ .Lambda.Architecture }}` — Architecture (e.g., "x86_64", "arm64")
- `{{ .Lambda.Layers }}` — Comma-separated list of layer ARNs
- `{{ .Lambda.Environment }}` — Environment variables as JSON string
- `{{ .Lambda.IAMRole }}` — IAM role ARN or name
- `{{ .Lambda.Triggers }}` — Event triggers (e.g., "apigateway,s3")
- `{{ .Lambda.DeploymentPackage }}` — Deployment package path

### Node.js/TypeScript-specific Variables
- `{{ .Node.Version }}` — Node.js version
- `{{ .Node.PackageName }}` — Name in package.json
- `{{ .Node.TypeScriptVersion }}` — TypeScript version

### Cloud/AWS Information
- `{{ .Cloud.Provider }}` — Cloud provider name (e.g., "aws", "gcp", "azure")
- `{{ .Cloud.AWSRegion }}` — AWS region
- `{{ .Cloud.AWSAccountID }}` — AWS account ID

### Docker/Kubernetes Variables
- `{{ .Docker.ImageName }}` — Docker image name
- `{{ .Docker.ImageTag }}` — Docker image tag
- `{{ .Docker.Registry }}` — Docker registry URL
- `{{ .Docker.K8sNamespace }}` — Kubernetes namespace

### Organization Information
- `{{ .Org.Name }}` — Organization name
- `{{ .Org.Team }}` — Team name
- `{{ .Org.CIProvider }}` — CI provider (e.g., "github-actions", "jenkins")

### Environment Information
- `{{ .Env.Default }}` — Default environment (e.g., "dev", "staging")
- `{{ .Env.Environments }}` — List of supported environments

### Documentation Links
- `{{ .Docs.URL }}` — URL to project documentation
- `{{ .Docs.APIURL }}` — URL to API documentation

### Metadata
- `{{ .Meta.CreatedAt }}` — When the project was created
- `{{ .Meta.CreatedBy }}` — Username of project creator
- `{{ .Meta.Year }}` — Current year (for license, copyright notices)

> **Note:** Some variables are language or platform specific and will only be set if relevant to your project type (e.g., Go, Node.js, AWS, Docker).

---

## Example Usage in a Template

### Markdown

```
# {{ .Project.Name }}

{{ .Project.Description }}

Maintained by: {{ .Service.Owner }}
Service port: {{ .Service.Port }}
```

### Go

```go
package main

import "fmt"

func main() {
    fmt.Println("Service {{ .Service.Name }} ({{ .Service.Type }}) running on port {{ .Service.Port }}")
}
```

### Node.js

```js
console.log(`Service {{ .Service.Name }} ({{ .Service.Type }}) running on port {{ .Service.Port }}`);
```

---

## Advanced Template Features

The template engine is based on Go's `text/template` package, which supports:

- **Conditionals:**
  ```
  {{if .Repo.IsPrivate}}
  This repository is private.
  {{else}}
  This repository is public.
  {{end}}
  ```
- **Loops:**
  ```
  {{range .Contributors}}
  - {{.}}
  {{end}}
  ```
- **Built-in Functions:**
  ```
  Project: {{upper .Project.Name}}
  Owner: {{title .Repo.Owner}}
  ```

- **Custom Formatting Functions:**
  ```
  {{pascal .Project.Name}}    → PascalCase (e.g., "my-service" → "MyService")
  {{kebab .Project.Name}}     → kebab-case (e.g., "MyService" → "my-service")
  {{title .Project.Name}}     → Title Case (e.g., "my-service" → "My Service")
  ```
  These are useful for generating code, filenames, and configuration that requires specific naming formats.

For a full list of available functions, see the [Go template documentation](https://pkg.go.dev/text/template) and [Sprig functions](https://masterminds.github.io/sprig/).

---

For more on team configuration and project creation, see [Configuration](configuration.md) and [Project & Repo Creation](project-creation.md).
