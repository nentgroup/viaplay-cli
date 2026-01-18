// Package cache provides functionality for managing the template cache.
package cache

import (
	"fmt"
	"strings"
)

// FormatPruneResults returns a formatted string summarising the results of a prune operation
func FormatPruneResults(prunedCount, totalTemplates int) string {
	if prunedCount == 0 {
		if totalTemplates == 0 {
			return "No templates found in cache."
		} else {
			return "No templates were pruned. All templates are newer than the specified age."
		}
	} else {
		return fmt.Sprintf("Successfully pruned %d templates out of %d total templates.", prunedCount, totalTemplates)
	}
}

// FormatTemplateList returns a formatted string of all templates in the cache
func FormatTemplateList(templates []Template) string {
	if len(templates) == 0 {
		return "  No templates found in cache."
	}

	var builder strings.Builder
	for _, tmpl := range templates {
		version := tmpl.Version
		if version == "" {
			version = "unknown"
		}

		remote := tmpl.RemoteURL
		if remote == "" {
			remote = "unknown"
		}

		builder.WriteString(fmt.Sprintf("  %s/%s (version: %s, remote: %s)\n",
			tmpl.Language,
			tmpl.Type,
			version,
			remote))
	}

	return builder.String()
}

// FormatUpdateResults returns a formatted string summarising the results of an update operation
func FormatUpdateResults(successCount, failCount int) string {
	var builder strings.Builder
	builder.WriteString("\nUpdate summary:\n")
	builder.WriteString(fmt.Sprintf("- %d templates updated successfully\n", successCount))

	if failCount > 0 {
		builder.WriteString(fmt.Sprintf("- %d templates failed to update\n", failCount))
	}

	if successCount == 0 && failCount == 0 {
		builder.WriteString("No templates found to update.")
	}

	return builder.String()
}
