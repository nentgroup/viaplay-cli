# Templates

This section explains how templates, teams, and template variables work together in viaplay-cli, and how to configure and use them for your projects.

---

## How Templates Work

Templates define the structure and content of new projects. You can use official templates, your own local templates, or remote templates from GitHub or other sources. Templates are configured in your global config file under the `templates` section. Example:

```yaml
templates:
  go:
    cli: github@github.com/nentgroup/go-cli-template.git
    lambda: github@github.com/nentgroup/go-lambda-template.git
    package: github@github.com/nentgroup/go-package-template.git
    service: github@github.com/nentgroup/go-service-template.git
  rust:
    http-service: local@/Users/alescole/.config/viaplay/templates_cache/rust/http-service
  typescript:
    service: github@github.com/nentgroup/ts-service-template.git
    lambda: github@github.com/nentgroup/ts-lambda-template.git
```

- You can add your own templates by editing this section.
- Template sources can be GitHub repos, local paths, or tarball URLs.

---

## Teams

Teams allow you to group configuration, secrets, and rulesets for a set of users or projects. Each team has its own configuration file, typically located at:

- `~/.config/viaplay/teams/<team>/config.yaml`

Team configs can specify:
- Team-specific rulesets
- Team secrets
- Default branches
- Environment settings

Example team config:
```yaml
ruleset: ~/.config/viaplay/teams/gecko/ruleset.json
secrets: ~/.config/viaplay/teams/gecko/secrets.json
default_branch: main
environments:
  staging:
    url: https://staging.example.com
    secrets:
      - name: STAGING_API_KEY
        value: ${{ secrets.STAGING_API_KEY }}
```

- Specify a team when creating a project to use its config, secrets, and rulesets:
  ```
  vp create --team gecko --repo-name my-service ...
  ```

---

## Template Variables

viaplay-cli supports a set of template variables that can be used in your project templates. These variables are replaced with actual values during project scaffolding.

Use the syntax `{{{.VarName}}}` (triple braces, leading dot) in your template files. For example:

```
# {{{.ProjectName}}}
Owner: {{{.RepoOwner}}}
Service Port: {{{.ServicePort}}}
```

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

For more information on template variables, see [Template Variables](templates.md).
