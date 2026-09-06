package cmds

import (
	"UMMC/help"
	"fmt"

	"github.com/spf13/cobra"
)

var deleteGameFilesNukeArg bool

var nukeCmd = &cobra.Command{
	Use:     "nuke",
	Aliases: []string{"reset", "wipe"},
	Short:   "Wipe all UMMC data, mods, backups, and database for a clean slate",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Warning: This will permanently delete all UMMC data (mods, backups, windows files, and database).")
		err := help.NukeAllDataAction(deleteGameFilesNukeArg, nil)
		if err != nil {
			fmt.Printf("Error resetting UMMC: %v\n", err)
			return
		}
		fmt.Println("UMMC has been completely reset to a clean state.")
	},
}

func init() {
	rootCmd.AddCommand(nukeCmd)
	nukeCmd.Flags().BoolVar(&deleteGameFilesNukeArg, "game-files", false, "Also delete UNDERTALE.app from Steam directory")
}
