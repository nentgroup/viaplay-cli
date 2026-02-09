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
	Short: "Update all templates in the cache",
	Long:  `Update all templates in the cache to their latest versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		updateCache(ctx)
	},
}

func init() {
	cacheCmd.AddCommand(cacheUpdateCmd)
}

// updateCache updates all templates in the cache
func updateCache(ctx context.Context) {
	output.Section("Template Cache Update")

	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	output.ProcessingMessage("Updating all templates in cache...")

	// Update all templates
	successCount, failCount, err := manager.UpdateAllTemplates(ctx)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to update templates: %v", err))
		return
	}

	if successCount > 0 {
		output.SuccessMessage(fmt.Sprintf("Successfully updated %d templates", successCount))
	}

	// Print summary
	fmt.Println(cache.FormatUpdateResults(successCount, failCount))
}
