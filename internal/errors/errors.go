// Package errors provides specialised error handling for viaplay-cli.
// It defines error types, sentinel errors, and error formatting utilities.
package errors

import (
	"errors"
	"fmt"
)

// Standard errors package functions that we re-export
var (
	As     = errors.As
	Is     = errors.Is
	New    = errors.New
	Unwrap = errors.Unwrap
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	// CategoryUnknown represents an uncategorized error
	CategoryUnknown ErrorCategory = iota
	// CategoryUser represents a user error (input validation, missing flags, etc.)
	CategoryUser
	// CategorySystem represents a system error (file system issues, permissions, etc.)
	CategorySystem
	// CategoryNetwork represents a network error (HTTP request failures, etc.)
	CategoryNetwork
	// CategoryGitHub represents a GitHub API error
	CategoryGitHub
	// CategoryTemplate represents a template processing error
	CategoryTemplate
	// CategoryConfig represents a configuration error
	CategoryConfig
	// CategoryPermission represents a permission error
	CategoryPermission
)

// String returns the string representation of an ErrorCategory
func (c ErrorCategory) String() string {
	switch c {
	case CategoryUser:
		return "user error"
	case CategorySystem:
		return "system error"
	case CategoryNetwork:
		return "network error"
	case CategoryGitHub:
		return "GitHub error"
	case CategoryTemplate:
		return "template error"
	case CategoryConfig:
		return "configuration error"
	case CategoryPermission:
		return "permission error"
	default:
		return "unknown error"
	}
}

// AppError represents an application-specific error with context
type AppError struct {
	// Original is the original error
	Original error
	// Message is a human-readable error message
	Message string
	// Category is the error category
	Category ErrorCategory
	// Code is an optional error code for programmatic handling
	Code string
	// Suggestion is an optional suggestion for fixing the error
	Suggestion string
	// Data contains any additional contextual data
	Data map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Original)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Original
}

// WithCategory sets the error category and returns the error
func (e *AppError) WithCategory(category ErrorCategory) *AppError {
	e.Category = category
	return e
}

// WithCode sets the error code and returns the error
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithSuggestion sets a suggestion and returns the error
func (e *AppError) WithSuggestion(suggestion string) *AppError {
	e.Suggestion = suggestion
	return e
}

// WithData adds contextual data and returns the error
func (e *AppError) WithData(key string, value interface{}) *AppError {
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
	e.Data[key] = value
	return e
}

// GetCategory returns the error category
func GetCategory(err error) ErrorCategory {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category
	}
	return CategoryUnknown
}

// GetSuggestion returns the error suggestion if available
func GetSuggestion(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Suggestion
	}
	return ""
}

// IsUserError returns true if the error is a user error
func IsUserError(err error) bool {
	return GetCategory(err) == CategoryUser
}

// IsSystemError returns true if the error is a system error
func IsSystemError(err error) bool {
	return GetCategory(err) == CategorySystem
}

// IsNetworkError returns true if the error is a network error
func IsNetworkError(err error) bool {
	return GetCategory(err) == CategoryNetwork
}

// IsGitHubError returns true if the error is a GitHub API error
func IsGitHubError(err error) bool {
	return GetCategory(err) == CategoryGitHub
}

// IsTemplateError returns true if the error is a template processing error
func IsTemplateError(err error) bool {
	return GetCategory(err) == CategoryTemplate
}

// IsConfigError returns true if the error is a configuration error
func IsConfigError(err error) bool {
	return GetCategory(err) == CategoryConfig
}

// IsPermissionError returns true if the error is a permission error
func IsPermissionError(err error) bool {
	return GetCategory(err) == CategoryPermission
}

// E creates a new AppError with the given message and original error
func E(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryUnknown,
	}
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return E(message, err)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return E(fmt.Sprintf(format, args...), err)
}

// User creates a new user error
func User(message string) *AppError {
	return &AppError{
		Message:  message,
		Category: CategoryUser,
	}
}

// Userf creates a new formatted user error
func Userf(format string, args ...interface{}) *AppError {
	return &AppError{
		Message:  fmt.Sprintf(format, args...),
		Category: CategoryUser,
	}
}

// System creates a new system error
func System(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategorySystem,
	}
}

// Systemf creates a new formatted system error
func Systemf(original error, format string, args ...interface{}) *AppError {
	return &AppError{
		Message:  fmt.Sprintf(format, args...),
		Original: original,
		Category: CategorySystem,
	}
}

// Network creates a new network error
func Network(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryNetwork,
	}
}

// GitHub creates a new GitHub API error
func GitHub(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryGitHub,
	}
}

// Template creates a new template processing error
func Template(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryTemplate,
	}
}

// Config creates a new configuration error
func Config(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryConfig,
	}
}

// Permission creates a new permission error
func Permission(message string, original error) *AppError {
	return &AppError{
		Message:  message,
		Original: original,
		Category: CategoryPermission,
	}
}

// Sentinel errors for common scenarios
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = New("resource not found")

	// ErrInvalidInput is returned when user input is invalid
	ErrInvalidInput = New("invalid input")

	// ErrPermissionDenied is returned when the user does not have permission
	ErrPermissionDenied = New("permission denied")

	// ErrNotAuthenticated is returned when the user is not authenticated
	ErrNotAuthenticated = New("not authenticated")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = New("operation timed out")

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists = New("resource already exists")

	// ErrNotImplemented is returned when a feature is not implemented
	ErrNotImplemented = New("feature not implemented")

	// ErrConfiguration is returned when there is a configuration error
	ErrConfiguration = New("configuration error")

	// ErrGitHubAPI is returned when there is a GitHub API error
	ErrGitHubAPI = New("GitHub API error")

	// ErrTemplateProcessing is returned when there is a template processing error
	ErrTemplateProcessing = New("template processing error")
)

// FormatError formats an error for display to the user
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		// Format specialised app error
		var result string

		switch appErr.Category {
		case CategoryUser:
			result = fmt.Sprintf("User error: %s", appErr.Message)
		case CategorySystem:
			result = fmt.Sprintf("System error: %s", appErr.Message)
		case CategoryNetwork:
			result = fmt.Sprintf("Network error: %s", appErr.Message)
		case CategoryGitHub:
			result = fmt.Sprintf("GitHub error: %s", appErr.Message)
		case CategoryTemplate:
			result = fmt.Sprintf("Template error: %s", appErr.Message)
		case CategoryConfig:
			result = fmt.Sprintf("Configuration error: %s", appErr.Message)
		case CategoryPermission:
			result = fmt.Sprintf("Permission error: %s", appErr.Message)
		default:
			result = fmt.Sprintf("Error: %s", appErr.Message)
		}

		// Add suggestion if available
		if appErr.Suggestion != "" {
			result += fmt.Sprintf("\nSuggestion: %s", appErr.Suggestion)
		}

		return result
	}

	// Regular error
	return fmt.Sprintf("Error: %v", err)
}
