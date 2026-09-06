package cmds

import (
	"UMMC/help"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	saveCmdModFlag     string
	saveCmdForceFlag   bool
	saveCmdNameFlag    string
	saveCmdIDFlag      string
	copyFromModFlag    string
	copyToModFlag      string
	copySourceSaveFlag string
	copyTargetSaveFlag string
)

var savesCmd = &cobra.Command{
	Use:     "saves",
	Aliases: []string{"save", "savegame", "savegames"},
	Short:   "Manage Undertale save files and profiles across mods",
}

var createSaveCmd = &cobra.Command{
	Use:     "save [save_name]",
	Aliases: []string{"create", "add", "new"},
	Short:   "Save current Undertale game progress into a named slot",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := saveCmdNameFlag
		if name == "" && len(args) > 0 {
			name = args[0]
		}

		if name == "" {
			fmt.Println("Error: No save name specified. Provide a name as an argument or using --name / -n.")
			return
		}

		mod := saveCmdModFlag
		if mod == "" {
			if active, _ := help.GetActiveSaveInfo(); active != nil && active.ModName != "" {
				mod = active.ModName
			} else {
				mod = "Vanilla"
			}
		}

		rec, err := help.SaveCurrentGame(name, mod, saveCmdForceFlag, help.LogFunc(func(m string) { fmt.Println(m) }))
		if err != nil {
			fmt.Printf("Error saving game: %v\n", err)
			return
		}

		fmt.Printf("Save '%s' (mod: %s) stored successfully at: %s\n", rec.Name, rec.ModName, rec.Path)
	},
}

var loadSaveCmd = &cobra.Command{
	Use:     "load [save_name_or_id]",
	Aliases: []string{"restore", "use", "apply"},
	Short:   "Auto-save current progress and load a save slot into Undertale",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := saveCmdNameFlag
		if target == "" && len(args) > 0 {
			target = args[0]
		}

		if target == "" && saveCmdIDFlag != "" {
			target = saveCmdIDFlag
		}

		if target == "" {
			fmt.Println("Error: No save name or ID specified. Provide an argument or use --name / --id.")
			return
		}

		mod := saveCmdModFlag
		if mod == "" {
			mod = "Vanilla"
		}

		rec, err := help.GetSaveByNameOrID(target, mod)
		if err == nil && rec != nil {
			target = rec.Name
			mod = rec.ModName
		}

		if err := help.LoadSave(target, mod, help.LogFunc(func(m string) { fmt.Println(m) })); err != nil {
			fmt.Printf("Error loading save: %v\n", err)
			return
		}

		fmt.Printf("Save '%s' (mod: %s) is now active in Undertale!\n", target, mod)
	},
}

var listSavesCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all saved game profiles across mods",
	Run: func(cmd *cobra.Command, args []string) {
		records, err := help.GetSaves(saveCmdModFlag)
		if err != nil || len(records) == 0 {
			fmt.Println("No save slots found in database.")
			return
		}

		activeInfo, _ := help.GetActiveSaveInfo()

		fmt.Println("=== Undertale Saves ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tACTIVE\tSAVE NAME\tMOD\tSAVED AT\tPATH")
		fmt.Fprintln(w, "--\t------\t---------\t---\t--------\t----")

		for _, rec := range records {
			isActive := " "
			if activeInfo != nil && activeInfo.Name == rec.Name && activeInfo.ModName == rec.ModName {
				isActive = "*"
			}
			timeStr := rec.CreatedAt.Local().Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", rec.ID, isActive, rec.Name, rec.ModName, timeStr, rec.Path)
		}
		w.Flush()

		if activeInfo != nil && activeInfo.Name != "" {
			fmt.Printf("\n(*) Currently active in Undertale: '%s' (Mod: %s)\n", activeInfo.Name, activeInfo.ModName)
		} else {
			fmt.Println("\n(*) No active UMMC save tracked currently.")
		}
		fmt.Printf("Total saves: %d\n", len(records))
	},
}

var copySaveCmd = &cobra.Command{
	Use:     "copy [source_save] [new_save]",
	Aliases: []string{"cp", "duplicate"},
	Short:   "Duplicate a save slot and optionally assign it to another mod",
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		srcName := copySourceSaveFlag
		dstName := copyTargetSaveFlag

		if len(args) == 1 {
			if srcName == "" {
				srcName = args[0]
			} else {
				dstName = args[0]
			}
		} else if len(args) >= 2 {
			srcName = args[0]
			dstName = args[1]
		}

		if srcName == "" {
			fmt.Println("Error: No source save specified. Provide source name as first argument.")
			return
		}
		if dstName == "" {
			dstName = srcName + "_copy"
		}

		fromMod := copyFromModFlag
		if fromMod == "" {
			fromMod = "Vanilla"
		}

		toMod := copyToModFlag
		if toMod == "" {
			toMod = fromMod
		}

		// Support ID for source
		if id, err := strconv.Atoi(srcName); err == nil {
			if rec, errRec := help.GetSaveByID(id); errRec == nil && rec != nil {
				srcName = rec.Name
				fromMod = rec.ModName
			}
		}

		rec, err := help.CopySaveAndChangeMod(srcName, fromMod, dstName, toMod, help.LogFunc(func(m string) { fmt.Println(m) }))
		if err != nil {
			fmt.Printf("Error copying save: %v\n", err)
			return
		}

		fmt.Printf("Successfully copied save to '%s' (mod: %s)!\n", rec.Name, rec.ModName)
	},
}

var removeSaveCmd = &cobra.Command{
	Use:     "remove [save_name_or_id]",
	Aliases: []string{"delete", "rm"},
	Short:   "Delete a save slot from disk and database",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := saveCmdNameFlag
		if target == "" && len(args) > 0 {
			target = args[0]
		}
		if target == "" && saveCmdIDFlag != "" {
			target = saveCmdIDFlag
		}

		if target == "" {
			fmt.Println("Error: No save name or ID specified. Provide an argument or use --name / --id.")
			return
		}

		mod := saveCmdModFlag
		if mod == "" {
			mod = "Vanilla"
		}

		if id, err := strconv.Atoi(target); err == nil {
			if rec, errRec := help.GetSaveByID(id); errRec == nil && rec != nil {
				target = rec.Name
				mod = rec.ModName
			}
		}

		if err := help.DeleteSave(target, mod, help.LogFunc(func(m string) { fmt.Println(m) })); err != nil {
			fmt.Printf("Error deleting save: %v\n", err)
			return
		}

		fmt.Printf("Successfully deleted save '%s' (mod: %s)!\n", target, mod)
	},
}

var activeSaveCmd = &cobra.Command{
	Use:     "active",
	Aliases: []string{"current", "status"},
	Short:   "Display information about the currently active Undertale save",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := help.GetActiveSaveInfo()
		if err != nil {
			fmt.Printf("Error checking active save: %v\n", err)
			return
		}

		activeDir := help.GetUndertaleSaveDir()
		hasFiles := help.HasActiveSaveFiles(activeDir)

		fmt.Println("=== Active Undertale Save Status ===")
		fmt.Printf("Save Directory: %s\n", activeDir)
		fmt.Printf("Has Save Files: %v\n", hasFiles)

		if info != nil && info.Name != "" {
			fmt.Printf("Active Save Name: %s\n", info.Name)
			fmt.Printf("Associated Mod:   %s\n", info.ModName)
			if !info.SavedAt.IsZero() {
				fmt.Printf("Last Synced:      %s\n", info.SavedAt.Local().Format("2006-01-02 15:04:05"))
			}
		} else {
			if hasFiles {
				fmt.Println("Active Save:      Untracked save files present in Undertale directory.")
			} else {
				fmt.Println("Active Save:      No save files found.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(savesCmd)

	savesCmd.AddCommand(createSaveCmd)
	savesCmd.AddCommand(loadSaveCmd)
	savesCmd.AddCommand(listSavesCmd)
	savesCmd.AddCommand(copySaveCmd)
	savesCmd.AddCommand(removeSaveCmd)
	savesCmd.AddCommand(activeSaveCmd)

	// Save flags
	createSaveCmd.Flags().StringVarP(&saveCmdNameFlag, "name", "n", "", "The name of the save slot")
	createSaveCmd.Flags().StringVarP(&saveCmdModFlag, "mod", "m", "", "The mod this save belongs to (default: Vanilla or active mod)")
	createSaveCmd.Flags().BoolVarP(&saveCmdForceFlag, "force", "f", true, "Force overwrite if save slot already exists")

	// Load flags
	loadSaveCmd.Flags().StringVarP(&saveCmdNameFlag, "name", "n", "", "The name of the save slot to load")
	loadSaveCmd.Flags().StringVarP(&saveCmdIDFlag, "id", "i", "", "The ID of the save slot to load")
	loadSaveCmd.Flags().StringVarP(&saveCmdModFlag, "mod", "m", "Vanilla", "The mod the save slot belongs to")

	// List flags
	listSavesCmd.Flags().StringVarP(&saveCmdModFlag, "mod", "m", "", "Filter saves by mod name")

	// Copy flags
	copySaveCmd.Flags().StringVarP(&copySourceSaveFlag, "source", "s", "", "Source save name")
	copySaveCmd.Flags().StringVarP(&copyTargetSaveFlag, "target", "t", "", "New save name")
	copySaveCmd.Flags().StringVar(&copyFromModFlag, "from-mod", "Vanilla", "Mod of source save")
	copySaveCmd.Flags().StringVar(&copyToModFlag, "to-mod", "Vanilla", "Target mod for copied save")

	// Remove flags
	removeSaveCmd.Flags().StringVarP(&saveCmdNameFlag, "name", "n", "", "The name of the save to remove")
	removeSaveCmd.Flags().StringVarP(&saveCmdIDFlag, "id", "i", "", "The ID of the save to remove")
	removeSaveCmd.Flags().StringVarP(&saveCmdModFlag, "mod", "m", "Vanilla", "The mod the save belongs to")
}
