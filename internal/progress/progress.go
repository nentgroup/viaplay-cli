// Package progress provides types and functions for tracking operation progress
// throughout the application in a presentation-agnostic way.
package progress

import (
	"fmt"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

// Status represents the current status of an operation
type Status string

// Pre-defined status constants
const (
	StatusStarted    Status = "started"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusSkipped    Status = "skipped"
	StatusWarning    Status = "warning"
	StatusDebug      Status = "debug" // For debug messages
	StatusInfo       Status = "info"  // For info messages
	StatusError      Status = "error" // For error messages
)

// Reporter defines an interface for reporting operation progress
type Reporter interface {
	// Start reports that an operation has started
	Start(operation, details string)

	// Progress reports an update on a running operation
	Progress(operation string, percentComplete int, details string)

	// Complete reports that an operation has completed successfully
	Complete(operation, details string)

	// Failed reports that an operation has failed
	Failed(operation string, err error, details string)

	// Skip reports that an operation was skipped
	Skip(operation, reason string)

	// Warning reports a warning during an operation
	Warning(operation, message string)

	// Info reports general information
	Info(message string)

	// Debug reports debug information (only shown in verbose mode)
	Debug(message string)

	Error(message string)
}

// Callback is a function that receives operation updates
type Callback func(operation string, status Status, details string, err error)

// CallbackReporter implements Reporter by calling a callback function
type CallbackReporter struct {
	callback Callback
	debug    bool
}

// NewCallbackReporter creates a new CallbackReporter
func NewCallbackReporter(callback Callback, debug bool) *CallbackReporter {
	return &CallbackReporter{
		callback: callback,
		debug:    debug,
	}
}

// Start implements Reporter.Start
func (r *CallbackReporter) Start(operation, details string) {
	r.callback(operation, StatusStarted, details, nil)
}

// Progress implements Reporter.Progress
func (r *CallbackReporter) Progress(operation string, percentComplete int, details string) {
	r.callback(operation, StatusInProgress, details, nil)
}

// Complete implements Reporter.Complete
func (r *CallbackReporter) Complete(operation, details string) {
	r.callback(operation, StatusCompleted, details, nil)
}

// Failed implements Reporter.Failed
func (r *CallbackReporter) Failed(operation string, err error, details string) {
	r.callback(operation, StatusFailed, details, err)
}

// Skip implements Reporter.Skip
func (r *CallbackReporter) Skip(operation, reason string) {
	r.callback(operation, StatusSkipped, reason, nil)
}

// Warning implements Reporter.Warning
func (r *CallbackReporter) Warning(operation, message string) {
	r.callback(operation, StatusWarning, message, nil)
}

// Info implements Reporter.Info
func (r *CallbackReporter) Info(message string) {
	r.callback("info", StatusInfo, message, nil)
}

// Debug implements Reporter.Debug
func (r *CallbackReporter) Debug(message string) {
	if r.debug {
		r.callback("debug", StatusDebug, message, nil)
	}
}

func (r *CallbackReporter) Error(message string) {
	r.callback("error", StatusError, message, fmt.Errorf("error: %s", message))
}

// NoopReporter is a Reporter implementation that does nothing
type NoopReporter struct{}

// NewNoopReporter creates a new NoopReporter
func NewNoopReporter() *NoopReporter {
	return &NoopReporter{}
}

var _ Reporter = &NoopReporter{}

// Start implements Reporter.Start
func (r *NoopReporter) Start(operation, details string) {}

// Progress implements Reporter.Progress
func (r *NoopReporter) Progress(operation string, percentComplete int, details string) {}

// Complete implements Reporter.Complete
func (r *NoopReporter) Complete(operation, details string) {}

// Failed implements Reporter.Failed
func (r *NoopReporter) Failed(operation string, err error, details string) {}

// Skip implements Reporter.Skip
func (r *NoopReporter) Skip(operation, reason string) {}

// Warning implements Reporter.Warning
func (r *NoopReporter) Warning(operation, message string) {}

// Info implements Reporter.Info
func (r *NoopReporter) Info(message string) {}

// Debug implements Reporter.Debug
func (r *NoopReporter) Debug(message string) {}

func (r *NoopReporter) Error(message string) {}

// DefaultCB is a default callback implementation that uses the output.Spinner
// to display progress updates with appropriate icons and formatting.
// It maintains a single spinner instance across multiple callbacks.
var spinnerInstance *output.Spinner

func DefaultCB(operation string, status Status, details string, err error) {
	// Initialise spinner once
	if spinnerInstance == nil {
		spinnerInstance = output.NewSpinner()
	}

	switch status {
	case StatusDebug:
		output.VerboseMessage(details)
		return
	case StatusInfo:
		output.InfoMessage(details)
		return
	case StatusError:
		output.ErrorMessage(details)
		return
	case StatusStarted:
		spinnerInstance.Start(operation)
	case StatusInProgress:
		if details != "" {
			spinnerInstance.Update(fmt.Sprintf("%s: %s", operation, details))
		}
	case StatusCompleted:
		successMsg := operation
		if details != "" {
			successMsg = fmt.Sprintf("%s: %s", operation, details)
		}
		spinnerInstance.Success(successMsg)
	case StatusFailed:
		failMsg := fmt.Sprintf("%s failed", operation)
		if err != nil {
			failMsg = fmt.Sprintf("%s: %v", failMsg, err)
		}
		spinnerInstance.Fail(failMsg)
	case StatusSkipped:
		skipMsg := fmt.Sprintf("%s skipped", operation)
		if details != "" {
			skipMsg = fmt.Sprintf("%s: %s", skipMsg, details)
		}
		spinnerInstance.Skip(skipMsg)
	case StatusWarning:
		// Handle warnings as spinner updates but don't change spinner state
		//output.WarningMessage(fmt.Sprintf("%s: %s", operation, details))
		skipMsg := fmt.Sprintf("%s skipped", operation)
		if details != "" {
			skipMsg = fmt.Sprintf("%s: %s", skipMsg, details)
		}
		spinnerInstance.Warn(skipMsg)
	}
}

// NewDefaultReporter creates a new reporter with the default callback
// and automatically sets debug mode based on the output package's verbose setting
func NewDefaultReporter() Reporter {
	// Get verbose status from output package
	isVerbose := output.IsVerboseEnabled()
	return NewCallbackReporter(DefaultCB, isVerbose)
}
