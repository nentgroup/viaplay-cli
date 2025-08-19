# Template Variables

viaplay-cli supports a set of template variables that can be used in your project templates. These variables are replaced with actual values during project scaffolding.

---

## How to Reference Variables

Use the syntax `{{ .Namespace.VarName }}` (double braces, leading dot) in your template files. For example:

```
# {{ .Project.Name }}
Owner: {{ .Repo.Owner }}
Service Port: {{ .Service.Port }}
```

Template variables can be used in:
- File content (as shown above)
- Filenames (e.g., `{{.Project.Name}}.md`, `{{.Service.Name}}-config.yaml`)
- Folder names (e.g., `src/{{.Service.Name}}`, `{{kebab .Project.Name}}/lib`)

This allows you to dynamically name files and folders based on the project attributes.

---

## Supported Template Variables

The following variables are organized into namespaces for easier reference and to prevent naming collisions:

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

---

## Example Usage in a Template

### Markdown Example

```
# {{ .Project.Name }}

{{ .Project.Description }}

Maintained by: {{ .Service.Owner }}
Service port: {{ .Service.Port }}
```

### Go Example

```go
package main

import "fmt"

func main() {
    fmt.Println("Service {{ .Service.Name }} ({{ .Service.Type }}) running on port {{ .Service.Port }}")
}
```

### Rust Example

```rust
fn main() {
    println!("Service {{ .Service.Name }} ({{ .Service.Type }}) running on port {{ .Service.Port }}");
}
```

### Node.js Example

```js
console.log(`Service {{ .Service.Name }} ({{ .Service.Type }}) running on port {{ .Service.Port }}`);
```

---

## Advanced Go Template Features

The viaplay-cli template engine is based on Go's `text/template` package, which supports advanced features such as:

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
- **Functions:**
  You can use built-in functions like `upper`, `lower`, `title`, and more:
  ```
  Project: {{upper .Project.Name}}
  Owner: {{title .Repo.Owner}}
  ```

- **Custom Formatting Functions:**
  viaplay-cli provides special formatting functions to help with naming conventions:
  ```
  {{pascal .Project.Name}}    → Converts to PascalCase (e.g., "my-service" → "MyService")
  {{kebab .Project.Name}}     → Converts to kebab-case (e.g., "MyService" → "my-service")
  {{title .Project.Name}}     → Converts to Title Case (e.g., "my-service" → "My Service")
  ```
  These are particularly useful for generating code, filenames, and configuration that requires
  specific naming formats.

- **Nested Variables:**
  If your variables are structured, you can access nested fields:
  ```
  {{.Team.Name}}
  {{.Team.Members}}
  ```

For a full list of available functions, see the [Go template documentation](https://pkg.go.dev/text/template) and [Sprig functions](https://masterminds.github.io/sprig/).

---

> **Note:** Some variables are language or platform specific and will only be set if relevant to your project type (e.g., Go, Node.js, AWS, Docker).

For more on templates and teams, see the [Configuration](configuration.md) and [Project & Repo Creation](project-creation.md) sections.
