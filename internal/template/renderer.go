// Package template provides template rendering functionality for viaplay-cli.
// It handles loading, processing, and rendering project templates with variable substitution.
package template

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Template delimiter constants for template processing
const (
	DelimLeft  = "{{{"
	DelimRight = "}}}"
)

// Variables defines all variables that can be used in templates
type Variables struct {
	// Basic project information
	ProjectName        string // Name of the project/repository
	ProjectDescription string // Description of the project

	// Repository information
	RepoOwner  string // GitHub username or organisation name
	RepoName   string // Repository name (often same as ProjectName)
	RepoURL    string // Full GitHub repository URL (HTTPS)
	RepoSSHURL string // SSH URL for the repository
	IsPrivate  bool   // Whether the repository is private

	// Project language and type
	Language    string // Programming language (e.g., "go", "typescript", "python")
	ProjectType string // Type of project (e.g., "api", "library", "app")

	// Service information (for services/APIs)
	ServiceName     string // Name of the service
	ServiceOwner    string // Owner/team responsible for the service
	ServiceOwnerKey string // Key identifier for the service owner
	ServicePort     string // Port the service listens on
	ServiceType     string // Type of service (e.g., "http", "grpc", "worker")

	// Go-specific variables
	BinaryName string // Name of the compiled binary
	ModulePath string // Go module path (e.g., "github.com/nentgroup/service-name")
	GoVersion  string // Go version used (e.g., "1.20")

	// Node.js/TypeScript specific variables
	NodeVersion       string // Node.js version
	NPMPackageName    string // Name in package.json
	TypeScriptVersion string // TypeScript version

	// AWS/Cloud specific variables
	AWSRegion     string // AWS region
	AWSAccountID  string // AWS account ID
	CloudProvider string // Cloud provider name (e.g., "aws", "gcp", "azure")

	// Docker/Kubernetes variables
	DockerImageName string // Docker image name
	DockerImageTag  string // Docker image tag
	DockerRegistry  string // Docker registry URL (e.g., "ghcr.io", "docker.io")
	K8sNamespace    string // Kubernetes namespace

	// Organisation information
	Organisation string // Organisation name (added for template rendering)

	// CI/CD variables
	CIProvider string // CI provider (e.g., "github-actions", "jenkins")

	// Environment information
	DefaultEnv   string   // Default environment (e.g., "dev", "staging")
	Environments []string // List of supported environments

	// Organisational information
	Team string // Team name

	// Metadata
	CreatedAt time.Time // When the project was created
	CreatedBy string    // Username of project creator
	Year      int       // Current year (for license, copyright notices)

	// Documentation
	DocsURL    string // URL to project documentation
	APIDocsURL string // URL to API documentation

	// Miscellaneous
	License string // License type (e.g., "MIT", "Apache-2.0")
}

// NewTemplateVariables returns a template variables struct with sensible defaults
func NewTemplateVariables() *Variables {
	currentYear := time.Now().Year()

	return &Variables{
		// Set some sensible defaults
		ServicePort:    "8080",
		GoVersion:      "1.21",
		NodeVersion:    "20",
		CloudProvider:  "aws",
		CIProvider:     "github-actions",
		DefaultEnv:     "dev",
		Environments:   []string{"dev", "staging", "production"},
		Year:           currentYear,
		License:        "MIT",
		CreatedAt:      time.Now(),
		DockerRegistry: "ghcr.io", // Default to GitHub Container Registry
		DockerImageTag: "latest",  // Default image tag
	}
}

// WithProjectInfo adds basic project information to template variables
func (tv *Variables) WithProjectInfo(name, description string) *Variables {
	tv.ProjectName = name
	tv.ProjectDescription = description
	tv.RepoName = name
	tv.ServiceName = name
	tv.BinaryName = name
	tv.DockerImageName = strings.ToLower(name)

	return tv
}

// WithGoModule sets Go-specific template variables
func (tv *Variables) WithGoModule(repoOwner, repoName string) *Variables {
	// Update the relevant fields
	tv.RepoOwner = repoOwner
	tv.RepoName = repoName
	tv.ModulePath = fmt.Sprintf("github.com/%s/%s", repoOwner, repoName)
	tv.RepoURL = fmt.Sprintf("https://%s", tv.ModulePath)

	return tv
}

// WithServiceDetails sets service-specific template variables
func (tv *Variables) WithServiceDetails(serviceName, serviceOwner, port string) *Variables {
	tv.ServiceName = serviceName
	tv.ServiceOwner = serviceOwner
	tv.ServiceOwnerKey = strings.ToLower(strings.ReplaceAll(serviceOwner, " ", "-"))
	tv.ServicePort = port

	return tv
}

// FromMap creates template variables from a map of values
func (tv *Variables) FromMap(variables map[string]string) *Variables {
	// Extract common variables from the map
	projectName := variables["project_name"]
	if projectName == "" {
		projectName = variables["name"]
	}

	description := variables["description"]
	if description == "" {
		description = variables["project_description"]
	}

	repoOwner := variables["repo_owner"]
	if repoOwner == "" {
		repoOwner = variables["owner"]
	}

	// Set basic project info
	if projectName != "" {
		tv.WithProjectInfo(projectName, description)
	}

	// Set Go module path if applicable
	language := variables["language"]
	if language == "go" && repoOwner != "" && projectName != "" {
		tv.WithGoModule(repoOwner, projectName)
	}

	// Set service details
	serviceName := variables["service_name"]
	if serviceName == "" {
		serviceName = projectName
	}

	serviceOwner := variables["service_owner"]
	if serviceOwner == "" {
		serviceOwner = repoOwner
	}

	servicePort := variables["service_port"]
	if servicePort == "" {
		servicePort = tv.ServicePort // Use default
	}

	if serviceName != "" {
		tv.WithServiceDetails(serviceName, serviceOwner, servicePort)
	}

	// Set additional explicitly defined variables
	if variables["team"] != "" {
		tv.Team = variables["team"]
	}

	if variables["organization"] != "" {
		tv.Organisation = variables["organization"]
	}

	if variables["created_by"] != "" {
		tv.CreatedBy = variables["created_by"]
	}

	if variables["docker_image_tag"] != "" {
		tv.DockerImageTag = variables["docker_image_tag"]
	}

	if variables["aws_region"] != "" {
		tv.AWSRegion = variables["aws_region"]
	}

	if variables["aws_account_id"] != "" {
		tv.AWSAccountID = variables["aws_account_id"]
	}

	return tv
}

// Renderer handles template rendering
type Renderer struct {
	Variables *Variables
}

// NewRenderer creates a new template renderer with the given variables
func NewRenderer(variables *Variables) *Renderer {
	if variables == nil {
		variables = NewTemplateVariables()
	}

	return &Renderer{
		Variables: variables,
	}
}

// RenderFile processes a template file and writes the result to the destination
// It handles binary file detection, template processing, and file copying
func (r *Renderer) RenderFile(srcPath, destPath string, ensureDestDir bool) error {
	// First, check if the file is binary
	isBinary, err := isBinaryFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to check if file is binary %s: %w", srcPath, err)
	}

	// If binary, just copy without processing
	if isBinary {
		return r.copyFile(srcPath, destPath, ensureDestDir)
	}

	// Read the source file
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", srcPath, err)
	}

	// Check if the file contains template delimiters
	contentStr := string(content)
	if !strings.Contains(contentStr, DelimLeft) {
		// Not a template file, just copy it
		return r.copyFile(srcPath, destPath, ensureDestDir)
	}

	// Create a new template
	tmplName := filepath.Base(srcPath)
	tmpl, err := template.New(tmplName).
		Delims(DelimLeft, DelimRight).
		Option("missingkey=zero").
		Parse(contentStr)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", srcPath, err)
	}

	// Create the destination directory if requested
	if ensureDestDir {
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	// Execute the template with the provided variables
	if err := tmpl.Execute(destFile, r.Variables); err != nil {
		return fmt.Errorf("failed to execute template for file %s: %w", srcPath, err)
	}

	return nil
}

// RenderString processes a template string and returns the result
func (r *Renderer) RenderString(templateString string) (string, error) {
	// Check if the string contains template delimiters
	if !strings.Contains(templateString, DelimLeft) {
		return templateString, nil
	}

	// Create a new template
	tmpl, err := template.New("string").
		Delims(DelimLeft, DelimRight).
		Option("missingkey=zero").
		Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template string: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.Variables); err != nil {
		return "", fmt.Errorf("failed to execute template string: %w", err)
	}

	return buf.String(), nil
}

// RenderDirectoryPath processes a directory path as a template
func (r *Renderer) RenderDirectoryPath(path string) (string, error) {
	// Check if the path contains template delimiters
	if !strings.Contains(path, DelimLeft) {
		return path, nil
	}

	// Create a new template for the path
	tmpl, err := template.New("path").
		Delims(DelimLeft, DelimRight).
		Option("missingkey=zero").
		Parse(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path as template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.Variables); err != nil {
		return "", fmt.Errorf("failed to execute path template: %w", err)
	}

	return buf.String(), nil
}

// copyFile copies a file without template processing
func (r *Renderer) copyFile(srcPath, destPath string, ensureDestDir bool) error {
	// Read source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", srcPath, err)
	}

	// Create the destination directory if requested
	if ensureDestDir {
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Write to destination
	return os.WriteFile(destPath, data, 0o600)
}

// isBinaryFile checks if a file is likely binary by examining its content
// It uses a common heuristic: if a file contains NUL bytes or a high proportion
// of non-printable characters, it's likely binary
func isBinaryFile(path string) (bool, error) {
	// Common binary file extensions that should be copied directly
	binaryExtensions := map[string]bool{
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".gif":   true,
		".ico":   true,
		".pdf":   true,
		".zip":   true,
		".tar":   true,
		".gz":    true,
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
		".eot":   true,
		".otf":   true,
		".svg":   true, // Even though SVG is XML, it's often better to not template it
		".mp3":   true,
		".mp4":   true,
		".avi":   true,
		".mov":   true,
		".webm":  true,
		".webp":  true,
		".doc":   true,
		".docx":  true,
		".xls":   true,
		".xlsx":  true,
		".ppt":   true,
		".pptx":  true,
	}

	// Check file extension first
	ext := strings.ToLower(filepath.Ext(path))
	if binaryExtensions[ext] {
		return true, nil
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the first 512 bytes (same as http.DetectContentType)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	buf = buf[:n]

	// Check for NUL bytes or high ratio of non-printable characters
	nullCount := 0
	nonPrintableCount := 0
	for _, b := range buf {
		if b == 0 {
			nullCount++
		} else if b < 32 && !isAllowedNonPrintable(b) {
			nonPrintableCount++
		}
	}

	// If we have any NUL bytes, it's definitely binary
	if nullCount > 0 {
		return true, nil
	}

	// If >30% of the bytes are non-printable, consider it binary
	if n > 0 && float64(nonPrintableCount)/float64(n) > 0.3 {
		return true, nil
	}

	return false, nil
}

// isAllowedNonPrintable checks if a byte is a non-printable character that's common in text files
// like tab, newline, carriage return
func isAllowedNonPrintable(b byte) bool {
	switch b {
	case '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}
