package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

// Version information
var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "none"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show viaplay-cli version information",
	Long:  `Display version, build date, and commit information for viaplay-cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		output.Section("Viaplay CLI Version Information")
		fmt.Printf("Version:    %s\n", output.Bold(Version))
		fmt.Printf("Build Date: %s\n", output.Bold(BuildDate))
		fmt.Printf("Commit:     %s\n", output.Bold(Commit))
	},
}
