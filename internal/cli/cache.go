package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/nentgroup/viaplay-cli/internal/cache"
	"github.com/nentgroup/viaplay-cli/internal/config"
	"github.com/nentgroup/viaplay-cli/internal/output"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the template cache",
	Long:  `Manage the template cache used by viaplay-cli, including pruning old templates and cleaning the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Display help information by default
		if err := cmd.Help(); err != nil {
			fmt.Printf("Failed to show help: %v\n", err)
		}
	},
}

var cachePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old template files from the cache",
	Long:  `Remove templates from the cache that haven't been used in a specified number of days.`,
	Run: func(cmd *cobra.Command, args []string) {
		days, err := cmd.Flags().GetInt("days")
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to get 'days' flag: %v", err))
			return
		}
		pruneCache(days)
	},
}

var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the entire template cache",
	Long:  `Remove all template files from the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		cleanCache()
	},
}

var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates in the cache",
	Long:  `List all templates currently stored in the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		listCache()
	},
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about the cache",
	Long:  `Show detailed information about the template cache including size and statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		showCacheInfo()
	},
}

func init() {
	// Command registration moved to root.go
	cacheCmd.AddCommand(cachePruneCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheInfoCmd)

	// Add flags for the prune command
	cachePruneCmd.Flags().Int("days", 30, "Prune templates older than specified days")
}

// getCacheManager returns a configured cache manager
func getCacheManager() (*cache.Manager, error) {
	// Load configuration from the centralised config package
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Create a cache manager from the configuration
	cacheManager := cache.NewManagerFromConfig(cfg)

	return cacheManager, nil
}

// pruneCache removes templates from the cache that are older than the specified days
func pruneCache(days int) {
	output.CacheMessage(fmt.Sprintf("Pruning templates older than %d days", days))

	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	// Execute the prune operation
	prunedCount, totalTemplates, err := manager.PruneCache(days)
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to prune cache: %v", err))
		return
	}

	if prunedCount == 0 {
		output.InfoMessage("No templates were pruned")
		fmt.Printf("All %d templates are newer than %d days old\n", totalTemplates, days)
	} else {
		output.SuccessMessage(fmt.Sprintf("Successfully pruned %d/%d templates", prunedCount, totalTemplates))

		// Show more details
		fmt.Printf("\nTemplates removed: %s\n", output.PrimaryBold(fmt.Sprintf("%d", prunedCount)))
		fmt.Printf("Templates kept:   %s\n", output.PrimaryBold(fmt.Sprintf("%d", totalTemplates-prunedCount)))
		fmt.Printf("Age threshold:    %s\n", output.PrimaryBold(fmt.Sprintf("%d days", days)))
	}
}

// cleanCache removes all templates from the cache
func cleanCache() {
	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	// Get current templates to show what will be removed
	templates, err := manager.ListTemplates()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to list templates: %v", err))
		return
	}

	if len(templates) == 0 {
		output.InfoMessage("Cache is already empty")
		return
	}

	// Confirm before cleaning the entire cache
	output.WarningMessage("This will remove ALL cached templates")
	fmt.Print("Are you sure you want to clean the entire template cache? (y/N): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to read input: %v", err))
		return
	}

	if response != "y" && response != "Y" {
		output.InfoMessage("Cache cleaning cancelled")
		return
	}

	// List what will be removed
	output.Section("Templates to be removed")
	rows := [][]string{}
	totalSize := int64(0)

	for _, t := range templates {
		size := int64(0)
		err := filepath.Walk(t.Path, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				size += info.Size()
			}
			return nil
		})
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to walk path %s: %v", t.Path, err))
			continue
		}
		totalSize += size
		rows = append(rows, []string{
			t.Language,
			t.Type,
			output.FormatSize(size),
			t.LastUsed.Format("2006-01-02"),
		})
	}

	// Sort by language and type
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0] == rows[j][0] {
			return rows[i][1] < rows[j][1]
		}
		return rows[i][0] < rows[j][0]
	})

	fmt.Print(output.Table(
		[]string{"Language", "Type", "Size", "Last Used"},
		rows,
		2,
	))
	fmt.Printf("\nTotal size to be freed: %s\n\n", output.Bold(output.FormatSize(totalSize)))

	// Clean the cache
	startTime := time.Now()
	if err := manager.CleanCache(); err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to clean cache: %v", err))
		return
	}

	duration := time.Since(startTime)
	output.SuccessMessage("Cache cleaned successfully")
	fmt.Printf("Removed %s templates (%s) in %s\n",
		output.Bold(fmt.Sprintf("%d", len(templates))),
		output.Bold(output.FormatSize(totalSize)),
		output.Bold(output.Duration(duration)))
}

// listCache lists all templates in the cache
func listCache() {
	output.Section("Templates in Cache")

	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	// Get the templates
	templates, err := manager.ListTemplates()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to list templates: %v", err))
		return
	}

	if len(templates) == 0 {
		output.InfoMessage("No templates found in cache")
		return
	}

	// Prepare data for table
	rows := [][]string{}
	totalSize := int64(0)
	languages := make(map[string]struct{})
	types := make(map[string]struct{})

	for _, t := range templates {
		size := int64(0)
		err := filepath.Walk(t.Path, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				size += info.Size()
			}
			return nil
		})
		if err != nil {
			output.ErrorMessage(fmt.Sprintf("Failed to walk path %s: %v", t.Path, err))
			continue
		}
		totalSize += size

		version := t.Version
		if version == "" {
			version = "unknown"
		}

		remote := t.RemoteURL
		if remote == "" {
			remote = "unknown"
		}

		rows = append(rows, []string{
			t.Language,
			t.Type,
			output.FormatSize(size),
			version,
			remote,
		})

		languages[t.Language] = struct{}{}
		types[t.Type] = struct{}{}
	}

	// Sort by language and type
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0] == rows[j][0] {
			return rows[i][1] < rows[j][1]
		}
		return rows[i][0] < rows[j][0]
	})

	// Print the table
	fmt.Print(output.Table(
		[]string{"Language", "Type", "Size", "Version", "Repository"},
		rows,
		0,
	))

	// Print summary
	fmt.Println()
	fmt.Printf("Total templates: %s\n", output.Bold(fmt.Sprintf("%d", len(templates))))
	fmt.Printf("Total size:      %s\n", output.Bold(output.FormatSize(totalSize)))
	fmt.Printf("Languages:       %s\n", output.Bold(fmt.Sprintf("%d", len(languages))))
	fmt.Printf("Template types:  %s\n", output.Bold(fmt.Sprintf("%d", len(types))))
}

// showCacheInfo displays detailed information about the cache
func showCacheInfo() {
	output.Section("Cache Information")

	// Get the cache manager
	manager, err := getCacheManager()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to initialize cache manager: %v", err))
		return
	}

	// Get cache directory and ensure it exists
	cacheDir := manager.GetCacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		output.InfoMessage("Cache directory does not exist yet")
		fmt.Printf("Cache directory: %s\n", cacheDir)
		return
	}

	// Calculate cache size and stats
	var totalSize int64
	var fileCount int
	var dirCount int
	oldestFile := time.Now()
	newestFile := time.Time{}

	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
			totalSize += info.Size()

			// Track oldest and newest files
			modTime := info.ModTime()
			if modTime.Before(oldestFile) {
				oldestFile = modTime
			}
			if modTime.After(newestFile) {
				newestFile = modTime
			}
		}
		return nil
	})
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to analyze cache: %v", err))
		return
	}

	// Print general information
	fmt.Printf("Cache directory:    %s\n", output.Bold(cacheDir))
	fmt.Printf("Total size:         %s\n", output.Bold(output.FormatSize(totalSize)))
	fmt.Printf("Files:              %s\n", output.Bold(fmt.Sprintf("%d", fileCount)))
	fmt.Printf("Directories:        %s\n", output.Bold(fmt.Sprintf("%d", dirCount)))

	if !oldestFile.Equal(time.Now()) && !newestFile.Equal(time.Time{}) {
		fmt.Printf("Oldest cached file:  %s (%s ago)\n",
			output.Bold(oldestFile.Format("2006-01-02 15:04:05")),
			output.Duration(time.Since(oldestFile)))
		fmt.Printf("Newest cached file:  %s (%s ago)\n",
			output.Bold(newestFile.Format("2006-01-02 15:04:05")),
			output.Duration(time.Since(newestFile)))
	}

	// Get and display template counts
	templates, err := manager.ListTemplates()
	if err != nil {
		output.ErrorMessage(fmt.Sprintf("Failed to list templates: %v", err))
		return
	}

	// Group templates by language
	templatesByLang := make(map[string][]cache.TemplateInfo)
	for _, t := range templates {
		// Convert Template to TemplateInfo
		templateInfo := cache.TemplateInfo{
			Language: t.Language,
			Type:     t.Type,
			Path:     t.Path,
			LastUsed: t.LastUsed,
		}
		templatesByLang[t.Language] = append(templatesByLang[t.Language], templateInfo)
	}

	// Print template counts by language
	if len(templatesByLang) > 0 {
		fmt.Println("\nTemplates by language:")
		for lang, templates := range templatesByLang {
			fmt.Printf("  %s: %s\n", output.Bold(lang), output.Primary(fmt.Sprintf("%d", len(templates))))
		}
	}
}
