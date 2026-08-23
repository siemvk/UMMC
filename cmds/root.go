package cmds

import (
	"UMMC/help"
	"os"

	"github.com/spf13/cobra"
)

var GameCmdArg string
var ChapterCmdArg int

// ParseGameArg parses the game argument (e.g. "undertale", "deltarune-ch1", "deltarune-ch2")
// and optional chapter argument into a clean game identifier and chapter number.
func ParseGameArg(gameArg string, chapterArg int) (string, int) {
	return help.ParseGameKey(gameArg, chapterArg)
}

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
	rootCmd.PersistentFlags().StringVarP(&GameCmdArg, "game", "g", "", "Target game (e.g. undertale, deltarune-ch1, deltarune-ch2)")
	rootCmd.PersistentFlags().IntVarP(&ChapterCmdArg, "chapter", "c", 0, "Target game chapter if applicable (e.g. 1, 2)")
}
