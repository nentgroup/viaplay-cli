# Template Variables

viaplay-cli supports a set of template variables that can be used in your project templates. These variables are replaced with actual values during project scaffolding.

---

## How to Reference Variables

Use the syntax `{{{.VarName}}}` (triple braces, leading dot) in your template files. For example:

```
# {{{.ProjectName}}}
Owner: {{{.RepoOwner}}}
Service Port: {{{.ServicePort}}}
```

---

## Supported Template Variables

The following variables are available for use in your templates:

- `{{{.ProjectName}}}` — Name of the project/repository
- `{{{.ProjectDescription}}}` — Description of the project
- `{{{.RepoOwner}}}` — GitHub username or organization name
- `{{{.RepoName}}}` — Repository name (often same as ProjectName)
- `{{{.RepoURL}}}` — Full GitHub repository URL
- `{{{.IsPrivate}}}` — Whether the repository is private

**Service-related variables:**
- `{{{.ServiceName}}}` — Name of the service
- `{{{.ServiceOwner}}}` — Owner/team responsible for the service
- `{{{.ServiceOwnerKey}}}` — Key identifier for the service owner
- `{{{.ServicePort}}}` — Port the service listens on
- `{{{.ServiceType}}}` — Type of service (e.g., "http", "grpc", "worker")

**Go-specific variables:**
- `{{{.BinaryName}}}` — Name of the compiled binary
- `{{{.ModulePath}}}` — Go module path (e.g., "github.com/org/service")
- `{{{.GoVersion}}}` — Go version used (e.g., "1.20")

**Node.js/TypeScript-specific variables:**
- `{{{.NodeVersion}}}` — Node.js version
- `{{{.NPMPackageName}}}` — Name in package.json
- `{{{.TypeScriptVersion}}}` — TypeScript version

**Cloud/AWS-specific variables:**
- `{{{.AWSRegion}}}` — AWS region
- `{{{.AWSAccountID}}}` — AWS account ID
- `{{{.CloudProvider}}}` — Cloud provider name (e.g., "aws", "gcp")

**Docker/Kubernetes variables:**
- `{{{.DockerImageName}}}` — Docker image name

---

## Example Usage in a Template

### Markdown Example

```
# {{{.ProjectName}}}

{{{.ProjectDescription}}}

Maintained by: {{{.ServiceOwner}}}
Service port: {{{.ServicePort}}}
```

### Go Example

```go
package main

import "fmt"

func main() {
    fmt.Println("Service {{{.ServiceName}}} ({{{.ServiceType}}}) running on port {{{.ServicePort}}}")
}
```

### Rust Example

```rust
fn main() {
    println!("Service {{{.ServiceName}}} ({{{.ServiceType}}}) running on port {{{.ServicePort}}}");
}
```

### Node.js Example

```js
console.log(`Service {{{.ServiceName}}} ({{{.ServiceType}}}) running on port {{{.ServicePort}}}`);
```

---

## Advanced Go Template Features

The viaplay-cli template engine is based on Go's `text/template` package, which supports advanced features such as:

- **Conditionals:**
  ```
  {{if .IsPrivate}}
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
  Project: {{upper .ProjectName}}
  Owner: {{title .RepoOwner}}
  ```
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
