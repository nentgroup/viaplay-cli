package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/output"
)

var cacheUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh all local template copies",
	Long:  `Fetch or refresh all locally stored template copies from their configured sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		updateTemplates(ctx)
	},
}

func init() {
	cacheCmd.AddCommand(cacheUpdateCmd)
}

// updateTemplates refreshes all locally stored template copies.
func updateTemplates(ctx context.Context) {
	output.Section("Updating Local Templates")

	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	templates, err := manager.ListTemplates(ctx)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to list templates: %v", err))
		return
	}

	if len(templates) == 0 {
		output.InfoMessage("No local templates found")
		return
	}

	output.ProcessingMessage("Refreshing all local template copies...")

	// Update all templates
	successCount, failCount, err := manager.UpdateAllTemplates(ctx)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to update local templates: %v", err))
		return
	}

	if successCount > 0 {
		output.SuccessMessage(fmt.Sprintf("Successfully updated %d local template(s)", successCount))
	}

	// Print summary
	fmt.Println(cache.FormatUpdateResults(successCount, failCount))
}
