// Package template provides template rendering and variable management functionality.
package template

import (
	"fmt"
	"strings"
	"time"
)

// FeatureSet contains template feature selections.
type FeatureSet map[string]any

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

// CloudInfo contains cloud provider specific information.
type CloudInfo struct {
	Provider string // Cloud provider name (e.g., "aws", "gcp", "azure")
}

// DockerInfo contains Docker variables.
type DockerInfo struct {
	ImageName string // Docker image name
	ImageTag  string // Docker image tag
}

// OrgInfo contains organisational information
type OrgInfo struct {
	Name       string // Organisation name
	Team       string // Team name
	TeamID     int64  // GitHub team ID (used for repository permissions)
	CIProvider string // CI provider (e.g., "github-actions", "jenkins")
}

// EnvInfo contains environment information
type EnvInfo struct {
	Default      string   // Default environment (e.g., "dev", "staging")
	Environments []string // List of supported environments
}

// MetaInfo contains metadata about the project creation
type MetaInfo struct {
	CreatedAt time.Time // When the project was created
	CreatedBy string    // Username of project creator
	Year      int       // Current year (for license, copyright notices)
}

// DefaultCloudProvider is the default cloud provider supplied to templates.
const DefaultCloudProvider = "aws"

// Variables defines all variables that can be used in templates
type Variables struct {
	Project  ProjectInfo
	Repo     RepoInfo
	Service  ServiceInfo
	Go       GoInfo
	Rust     RustInfo
	Node     NodeInfo
	Cloud    CloudInfo
	Docker   DockerInfo
	Org      OrgInfo
	Env      EnvInfo
	Meta     MetaInfo
	Features FeatureSet
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
		Cloud: CloudInfo{
			Provider: DefaultCloudProvider,
		},
		Docker: DockerInfo{
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
		Features: FeatureSet{},
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

// SetFeature records a template feature selection.
func (tv *Variables) SetFeature(key string, value any) {
	if tv.Features == nil {
		tv.Features = FeatureSet{}
	}
	tv.Features[key] = value
}
