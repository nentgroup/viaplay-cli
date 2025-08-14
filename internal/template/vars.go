// Package template provides template rendering and variable management functionality.
package template

import (
	"fmt"
	"strings"
	"time"
)

// ProjectInfo contains basic project information
type ProjectInfo struct {
	Name        string // Name of the project/repository
	Description string // Description of the project
	Type        string // Type of project (e.g., "api", "library", "app")
	Language    string // Programming language (e.g., "go", "typescript", "python")
	License     string // License type (e.g., "MIT", "Apache-2.0")
}

// RepoInfo contains repository-specific information
type RepoInfo struct {
	Owner     string // GitHub username or organisation name
	Name      string // Repository name (often same as ProjectName)
	URL       string // Full GitHub repository URL (HTTPS)
	SSHURL    string // SSH URL for the repository
	IsPrivate bool   // Whether the repository is private
}

// ServiceInfo contains information for services/APIs
type ServiceInfo struct {
	Name     string // Name of the service
	Owner    string // Owner/team responsible for the service
	OwnerKey string // Key identifier for the service owner
	Port     string // Port the service listens on
	Type     string // Type of service (e.g., "http", "grpc", "worker")
}

// GoInfo contains Go-specific variables
type GoInfo struct {
	BinaryName string // Name of the compiled binary
	ModulePath string // Go module path (e.g., "github.com/nentgroup/service-name")
	Version    string // Go version used (e.g., "1.21")
}

// RustInfo contains Rust-specific variables
type RustInfo struct {
	BinaryName string // Name of the compiled binary
	CargoName  string // Name in Cargo.toml (may differ from repository name)
	Version    string // Rust version used (e.g., "1.75")
	Edition    string // Rust edition (e.g., "2021")
}

// NodeInfo contains Node.js/TypeScript specific variables
type NodeInfo struct {
	Version           string // Node.js version
	PackageName       string // Name in package.json
	TypeScriptVersion string // TypeScript version
}

// LambdaInfo contains AWS Lambda-specific variables
type LambdaInfo struct {
	FunctionName      string // Name of the Lambda function
	Handler           string // Handler path (e.g., "index.handler")
	Runtime           string // Lambda runtime (e.g., "nodejs18.x", "go1.x", "python3.9")
	Timeout           int    // Timeout in seconds
	MemorySize        int    // Memory size in MB
	Architecture      string // Architecture (e.g., "x86_64", "arm64")
	Layers            string // Comma-separated list of layer ARNs
	Environment       string // Environment variables as JSON string
	IAMRole           string // IAM role ARN or name
	Triggers          string // Comma-separated list of triggers (e.g., "apigateway,s3")
	DeploymentPackage string // Deployment package path (e.g., ".zip" file)
}

// CloudInfo contains cloud provider specific information
type CloudInfo struct {
	Provider     string // Cloud provider name (e.g., "aws", "gcp", "azure")
	AWSRegion    string // AWS region
	AWSAccountID string // AWS account ID
}

// DockerInfo contains Docker/Kubernetes variables
type DockerInfo struct {
	ImageName    string // Docker image name
	ImageTag     string // Docker image tag
	Registry     string // Docker registry URL (e.g., "ghcr.io", "docker.io")
	K8sNamespace string // Kubernetes namespace
}

// OrgInfo contains organisational information
type OrgInfo struct {
	Name       string // Organisation name
	Team       string // Team name
	CIProvider string // CI provider (e.g., "github-actions", "jenkins")
}

// EnvInfo contains environment information
type EnvInfo struct {
	Default      string   // Default environment (e.g., "dev", "staging")
	Environments []string // List of supported environments
}

// DocInfo contains documentation links
type DocInfo struct {
	URL    string // URL to project documentation
	APIURL string // URL to API documentation
}

// MetaInfo contains metadata about the project creation
type MetaInfo struct {
	CreatedAt time.Time // When the project was created
	CreatedBy string    // Username of project creator
	Year      int       // Current year (for license, copyright notices)
}

// Variables defines all variables that can be used in templates
type Variables struct {
	Project ProjectInfo
	Repo    RepoInfo
	Service ServiceInfo
	Go      GoInfo
	Rust    RustInfo
	Node    NodeInfo
	Lambda  LambdaInfo
	Cloud   CloudInfo
	Docker  DockerInfo
	Org     OrgInfo
	Env     EnvInfo
	Docs    DocInfo
	Meta    MetaInfo
}

// NewTemplateVariables returns a template variables struct with sensible defaults
func NewTemplateVariables() *Variables {
	currentYear := time.Now().Year()

	return &Variables{
		// Set default values for each nested structure
		Project: ProjectInfo{
			License: "MIT",
		},
		Service: ServiceInfo{
			Port: "8080",
		},
		Go: GoInfo{
			Version: "1.24",
		},
		Rust: RustInfo{
			Version: "1.88",
			Edition: "2024",
		},
		Node: NodeInfo{
			Version: "20",
		},
		Lambda: LambdaInfo{
			Timeout:      30,  // 30 seconds default timeout
			MemorySize:   512, // 512 MB default memory
			Architecture: "arm64",
			Runtime:      "nodejs22.x", // Default to current Node.js LTS
		},
		Cloud: CloudInfo{
			Provider: "aws",
		},
		Docker: DockerInfo{
			Registry: "ghcr.io",
			ImageTag: "latest",
		},
		Org: OrgInfo{
			CIProvider: "github-actions",
		},
		Env: EnvInfo{
			Default:      "dev",
			Environments: []string{"dev", "staging", "production"},
		},
		Meta: MetaInfo{
			CreatedAt: time.Now(),
			Year:      currentYear,
		},
	}
}

// WithProjectInfo adds basic project information to template variables
func (tv *Variables) WithProjectInfo(name, description string) *Variables {
	tv.Project.Name = name
	tv.Project.Description = description
	tv.Repo.Name = name
	tv.Service.Name = name
	tv.Go.BinaryName = name
	tv.Docker.ImageName = strings.ToLower(name)

	return tv
}

// WithGoModule sets Go-specific template variables
func (tv *Variables) WithGoModule(repoOwner, repoName string) *Variables {
	// Update the relevant fields
	tv.Repo.Owner = repoOwner
	tv.Repo.Name = repoName
	tv.Go.ModulePath = fmt.Sprintf("github.com/%s/%s", repoOwner, repoName)
	tv.Repo.URL = fmt.Sprintf("https://%s", tv.Go.ModulePath)

	return tv
}

// WithServiceDetails sets service-specific template variables
func (tv *Variables) WithServiceDetails(serviceName, serviceOwner, port string) *Variables {
	tv.Service.Name = serviceName
	tv.Service.Owner = serviceOwner
	tv.Service.OwnerKey = strings.ToLower(strings.ReplaceAll(serviceOwner, " ", "-"))
	tv.Service.Port = port

	return tv
}
