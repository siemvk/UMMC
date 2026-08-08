package cmds

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "UMMC",
	Short: "Undertale Manager Macos Cli (or Undertale macos mod cli, i forgot wich was the intended name.)",
	Long:  `A fast undertale CLI in go made for macos but it works with unix.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
