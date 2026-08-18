package cmds

import (
	"os"

	"github.com/spf13/cobra"
)

var DeltaruneCmdArg int
var UndertaleCmdArg bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "UMMC",
	Short: "Undertale & Deltarune Manager Macos CLI",
	Long:  `A fast Undertale and Deltarune CLI in Go made for macOS (should work with any Unix).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&DeltaruneCmdArg, "deltarune", "d", 0, "Target Deltarune chapter (e.g. 1 for Chapter 1, 2 for Chapter 2)")
	rootCmd.PersistentFlags().BoolVarP(&UndertaleCmdArg, "undertale", "u", false, "Target Undertale")
}
