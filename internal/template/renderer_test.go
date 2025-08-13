// Package template provides unit tests for the template renderer.
package template

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/spf13/afero"
)

func TestRenderer_VariableSubstitution(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		vars     *Variables
		expected string
	}{
		{
			name: "Basic project info",
			tmpl: "Project: {{.Project.Name}}, Description: {{.Project.Description}}",
			vars: &Variables{
				Project: ProjectInfo{
					Name:        "MyApp",
					Description: "A test app",
				},
			},
			expected: "Project: MyApp, Description: A test app",
		},
		{
			name: "Service info",
			tmpl: "Service: {{.Service.Name}} on port {{.Service.Port}}",
			vars: &Variables{
				Service: ServiceInfo{
					Name: "api",
					Port: "8080",
				},
			},
			expected: "Service: api on port 8080",
		},
		{
			name: "Go-specific",
			tmpl: "Go version: {{.Go.Version}}, Module: {{.Go.ModulePath}}",
			vars: &Variables{
				Go: GoInfo{
					Version:    "1.21",
					ModulePath: "github.com/example/app",
				},
			},
			expected: "Go version: 1.21, Module: github.com/example/app",
		},
		{
			name: "Node.js-specific",
			tmpl: "Node: {{.Node.Version}}, NPM: {{.Node.PackageName}}",
			vars: &Variables{
				Node: NodeInfo{
					Version:     "18.0.0",
					PackageName: "my-npm-app",
				},
			},
			expected: "Node: 18.0.0, NPM: my-npm-app",
		},
		{
			name: "AWS-specific",
			tmpl: "Region: {{.Cloud.AWSRegion}}, Account: {{.Cloud.AWSAccountID}}",
			vars: &Variables{
				Cloud: CloudInfo{
					AWSRegion:    "eu-west-1",
					AWSAccountID: "123456789012",
				},
			},
			expected: "Region: eu-west-1, Account: 123456789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Parse(tt.tmpl)
			if err != nil {
				t.Fatalf("template parse error: %v", err)
			}
			var buf bytes.Buffer
			err = tmpl.Execute(&buf, tt.vars)
			if err != nil {
				t.Fatalf("template execute error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("got %q, want %q", buf.String(), tt.expected)
			}
		})
	}
}

func TestRenderer_RenderString(t *testing.T) {
	r := &Renderer{
		Variables: &Variables{
			Project: ProjectInfo{
				Name:        "TestProject",
				Description: "A sample project",
			},
			Service: ServiceInfo{
				Port: "8080",
			},
			Go: GoInfo{
				Version: "1.21",
			},
			Cloud: CloudInfo{
				AWSRegion: "us-east-1",
			},
			Docker: DockerInfo{
				ImageName: "test/image",
				ImageTag:  "latest",
			},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No template",
			input:    "plain string",
			expected: "plain string",
		},
		{
			name:     "ProjectName variable",
			input:    "{{.Project.Name}}",
			expected: "TestProject",
		},
		{
			name:     "Multiple variables",
			input:    "{{.Project.Name}} - {{.Project.Description}}",
			expected: "TestProject - A sample project",
		},
		{
			name:     "ServicePort variable",
			input:    "Port: {{.Service.Port}}",
			expected: "Port: 8080",
		},
		{
			name:     "GoVersion variable",
			input:    "Go version: {{.Go.Version}}",
			expected: "Go version: 1.21",
		},
		{
			name:     "AWSRegion variable",
			input:    "Region: {{.Cloud.AWSRegion}}",
			expected: "Region: us-east-1",
		},
		{
			name:     "DockerImageName variable",
			input:    "Image: {{.Docker.ImageName}}:{{.Docker.ImageTag}}",
			expected: "Image: test/image:latest",
		},
		{
			name:     "Missing variable",
			input:    "{{.NotSet}}",
			expected: "{{.NotSet}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.RenderString(tt.input)
			//if tt.name == "Missing variable" {
			//	if err == nil {
			//		t.Errorf("expected error for missing variable, got %s", got)
			//	}
			//	return
			//}
			if err != nil {
				t.Fatalf("RenderString error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRenderer_AllTemplateVariables(t *testing.T) {
	r := &Renderer{
		Variables: &Variables{
			Project: ProjectInfo{
				Name:        "MyProject",
				Description: "A sample project",
				Language:    "go",
				Type:        "service",
			},
			Repo: RepoInfo{
				Owner:     "octocat",
				Name:      "my-repo",
				URL:       "https://github.com/octocat/my-repo",
				SSHURL:    "git@github.com:octocat/my-repo.git",
				IsPrivate: true,
			},
			Service: ServiceInfo{
				Name:     "api",
				Owner:    "backend-team",
				OwnerKey: "bt",
				Port:     "8080",
				Type:     "http",
			},
			Go: GoInfo{
				BinaryName: "myproject",
				ModulePath: "github.com/octocat/myproject",
				Version:    "1.21",
			},
			Node: NodeInfo{
				Version:           "18.0.0",
				PackageName:       "my-npm-app",
				TypeScriptVersion: "5.0",
			},
			Cloud: CloudInfo{
				Provider:     "aws",
				AWSRegion:    "eu-west-1",
				AWSAccountID: "123456789012",
			},
			Docker: DockerInfo{
				ImageName:    "octocat/myproject",
				ImageTag:     "latest",
				Registry:     "ghcr.io",
				K8sNamespace: "default",
			},
			Org: OrgInfo{
				Name:       "OctoOrg",
				Team:       "backend-team",
				CIProvider: "github-actions",
			},
			Env: EnvInfo{
				Default:      "staging",
				Environments: []string{"dev", "staging", "prod"},
			},
			Meta: MetaInfo{
				CreatedAt: time.Date(2025, 8, 10, 12, 0, 0, 0, time.UTC),
				CreatedBy: "octocat",
				Year:      2025,
			},
		},
	}

	templateStr := `
Project: {{.Project.Name}}
Description: {{.Project.Description}}
Owner: {{.Repo.Owner}}
Repo Name: {{.Repo.Name}}
Repo URL: {{.Repo.URL}}
Repo SSH: {{.Repo.SSHURL}}
Is Private: {{.Repo.IsPrivate}}
Language: {{.Project.Language}}
Project Type: {{.Project.Type}}
Service Name: {{.Service.Name}}
Service Owner: {{.Service.Owner}}
Service Owner Key: {{.Service.OwnerKey}}
Service Port: {{.Service.Port}}
Service Type: {{.Service.Type}}
Binary Name: {{.Go.BinaryName}}
Module Path: {{.Go.ModulePath}}
Go Version: {{.Go.Version}}
Node Version: {{.Node.Version}}
NPM Package Name: {{.Node.PackageName}}
TypeScript Version: {{.Node.TypeScriptVersion}}
AWS Region: {{.Cloud.AWSRegion}}
AWS Account ID: {{.Cloud.AWSAccountID}}
Cloud Provider: {{.Cloud.Provider}}
Docker Image Name: {{.Docker.ImageName}}
Docker Image Tag: {{.Docker.ImageTag}}
Docker Registry: {{.Docker.Registry}}
K8s Namespace: {{.Docker.K8sNamespace}}
Organisation: {{.Org.Name}}
Team: {{.Org.Team}}
CI Provider: {{.Org.CIProvider}}
Default Env: {{.Env.Default}}
Environments: {{range .Env.Environments}}{{.}} {{end}}
Created At: {{.Meta.CreatedAt}}
Created By: {{.Meta.CreatedBy}}
Year: {{.Meta.Year}}
`

	expected := `
Project: MyProject
Description: A sample project
Owner: octocat
Repo Name: my-repo
Repo URL: https://github.com/octocat/my-repo
Repo SSH: git@github.com:octocat/my-repo.git
Is Private: true
Language: go
Project Type: service
Service Name: api
Service Owner: backend-team
Service Owner Key: bt
Service Port: 8080
Service Type: http
Binary Name: myproject
Module Path: github.com/octocat/myproject
Go Version: 1.21
Node Version: 18.0.0
NPM Package Name: my-npm-app
TypeScript Version: 5.0
AWS Region: eu-west-1
AWS Account ID: 123456789012
Cloud Provider: aws
Docker Image Name: octocat/myproject
Docker Image Tag: latest
Docker Registry: ghcr.io
K8s Namespace: default
Organisation: OctoOrg
Team: backend-team
CI Provider: github-actions
Default Env: staging
Environments: dev staging prod 
Created At: 2025-08-10 12:00:00 +0000 UTC
Created By: octocat
Year: 2025
`
	got, err := r.RenderString(templateStr)
	if err != nil {
		t.Fatalf("RenderString error: %v", err)
	}
	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func TestRenderer_GitHubActionsSyntaxIsIgnored(t *testing.T) {
	r := &Renderer{
		Variables: &Variables{
			Project: ProjectInfo{
				Name: "MyProject",
			},
		},
	}

	templateStr := `
name: Release
on:
  push:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.21
      - name: Print project name
        run: echo "Project: {{.Project.Name}}"
      - name: Use GitHub Actions output
        run: echo "Release created: ${{ steps.release.outputs.release_created }}"
`

	expected := `
name: Release
on:
  push:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.21
      - name: Print project name
        run: echo "Project: MyProject"
      - name: Use GitHub Actions output
        run: echo "Release created: ${{ steps.release.outputs.release_created }}"
`

	got, err := r.RenderString(templateStr)
	if err != nil {
		t.Fatalf("RenderString error: %v", err)
	}
	if got != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func TestRenderer_RenderDirectoryPath(t *testing.T) {
	r := &Renderer{
		Variables: &Variables{
			Project: ProjectInfo{
				Name: "MyProject",
			},
			Service: ServiceInfo{
				Port: "8080",
			},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No template",
			input:    "plain/path",
			expected: "plain/path",
		},
		{
			name:     "ProjectName variable",
			input:    "projects/{{.Project.Name}}/service",
			expected: "projects/MyProject/service",
		},
		{
			name:     "ServicePort variable",
			input:    "services/{{.Service.Port}}/api",
			expected: "services/8080/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.RenderDirectoryPath(tt.input)
			if err != nil {
				t.Fatalf("RenderDirectoryPath error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRenderer_RenderFile_MockFS(t *testing.T) {
	memFS := afero.NewMemMapFs()
	_ = afero.WriteFile(memFS, "/tmp/template.txt", []byte("Hello, {{.Project.Name}}!"), 0o600)
	_ = afero.WriteFile(memFS, "/tmp/plain.txt", []byte("Just plain text."), 0o600)
	r := &Renderer{
		Variables:  &Variables{Project: ProjectInfo{Name: "TestProject"}},
		FileSystem: memFS,
	}

	// Test rendering a template file
	destPath := "/tmp/out.txt"
	err := r.RenderFile("/tmp/template.txt", destPath, false)
	if err != nil {
		t.Fatalf("RenderFile error: %v", err)
	}
	data, _ := afero.ReadFile(memFS, destPath)
	if string(data) != "Hello, TestProject!" {
		t.Errorf("got %q, want %q", string(data), "Hello, TestProject!")
	}

	// Test copying a plain file (now always rendered, so content should match)
	destPathPlain := "/tmp/out_plain.txt"
	err = r.RenderFile("/tmp/plain.txt", destPathPlain, false)
	if err != nil {
		t.Fatalf("RenderFile error: %v", err)
	}
	dataPlain, _ := afero.ReadFile(memFS, destPathPlain)
	if string(dataPlain) != "Just plain text." {
		t.Errorf("got %q, want %q", string(dataPlain), "Just plain text.")
	}
}
