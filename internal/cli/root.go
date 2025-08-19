// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"fmt"
	"os"

	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nentgroup/viaplay-cli/internal/output"
)

var (
	cfgFile     string
	verboseFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vip",
	Short: "Scaffold, configure, and manage Go projects and GitHub repositories for Viaplay teams.",
	Long: `viaplay-cli is a developer tool for quickly scaffolding Go projects from templates, 
creating and configuring GitHub repositories (organization or personal), and applying 
team or organization standards such as rulesets, secrets, and environments.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Define help template function for showing banner
	cobra.AddTemplateFunc("ShowBanner", func() string {
		printBanner()
		return ""
	})

	// Modify help template to show banner before help content
	helpTemplate := rootCmd.HelpTemplate()
	rootCmd.SetHelpTemplate(`{{ShowBanner}}` + helpTemplate)

	// Set a pre-run hook for the root command to display the banner
	// when running just 'vip' with no subcommands
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Only show the banner if it's the root command with no args
		if cmd.Name() == rootCmd.Name() && len(args) == 0 && !cmd.Flags().Changed("help") {
			printBanner()
		}
	}

	err := rootCmd.Execute()
	if err != nil {
		output.VerboseMessage("Root command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initVerbose)

	// Register all commands in one central place
	rootCmd.AddCommand(
		authCmd,             // Authentication
		secretsCmd,          // Secrets management
		cacheCmd,            // Cache management
		configCmd,           // Configuration
		versionCmd,          // Version information
		NewProjectCommand(), // Project management
		NewRepoCommand(),    // Repository management
	)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/viaplay/config.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output for debugging and troubleshooting")
}

func initVerbose() {
	// Set verbose mode for the output package
	output.SetVerbose(verboseFlag)
	output.VerboseMessage("Verbose mode enabled")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		output.VerboseMessage("Using config file from flag: " + cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		output.VerboseMessage("Searching config in home directory: " + home)
		// Search config in home directory
		viper.AddConfigPath(fmt.Sprintf("%s/.config/viaplay", home))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		output.VerboseMessage("Config file loaded: " + viper.ConfigFileUsed())
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		output.VerboseMessage("Could not read config file: " + err.Error())
		fmt.Fprintln(os.Stderr, "Warning: Could not read config file:", err)
	}
}

func printBanner() {
	bannerLines := []string{
		`
██╗   ██╗██╗ █████╗ ██████╗ ██╗      █████╗ ██╗   ██╗      ██████╗██╗     ██╗
██║   ██║██║██╔══██╗██╔══██╗██║     ██╔══██╗╚██╗ ██╔╝     ██╔════╝██║     ██║
██║   ██║██║███████║██████╔╝██║     ███████║ ╚████╔╝█████╗██║     ██║     ██║
╚██╗ ██╔╝██║██╔══██║██╔═══╝ ██║     ██╔══██║  ╚██╔╝ ╚════╝██║     ██║     ██║
 ╚████╔╝ ██║██║  ██║██║     ███████╗██║  ██║   ██║        ╚██████╗███████╗██║
  ╚═══╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚══════╝╚═╝  ╚═╝   ╚═╝         ╚═════╝╚══════╝╚═╝
`,
	}

	// Add the version as the last line of the banner
	versionLine := fmt.Sprintf("                                                                v%s", Version)
	bannerLines = append(bannerLines, versionLine, "\n")

	p := termenv.ColorProfile()
	colors := []string{"#e6007a", "#ff4e50"}
	for i, line := range bannerLines {
		color := colors[i%len(colors)]
		fmt.Println(termenv.String(line).Foreground(p.Color(color)).Bold())
	}
}
