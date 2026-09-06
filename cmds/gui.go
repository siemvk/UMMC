package cmds

import (
	"UMMC/gui"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical interface (Fyne)",
	Run: func(cmd *cobra.Command, args []string) {
		gui.StartApp()
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
