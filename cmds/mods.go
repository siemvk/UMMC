package cmds

import (
	"UMMC/help"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var qpFilenameCmdArg string
var forceQuickPatchCmdArg bool
var windowsDataQuickPatchCmdArg bool

var createModFolderCmdArg string
var createModNameCmdArg string
var createModMakerCmdArg string
var createModBaseCmdArg string
var createModWinCmdArg bool
var createModVanillaSavesCmdArg bool
var createModRootCmdArg bool
var createModButterscotchCmdArg bool
var forceCreateModCmdArg bool
var createModMacosCmdArg bool

var quickPatchCmdThingy = &cobra.Command{
	Use:   "quickpatch [filename]",
	Short: "patch a mod onto the global undertale install",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := qpFilenameCmdArg
		if filename == "" && len(args) > 0 {
			filename = args[0]
		}

		if filename == "" {
			fmt.Println("Error: No patch file specified. Provide a patch file as an argument or using --file / -p.")
			return
		}

		targetPath := help.ExpandPath("~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app/Contents/Resources/game.ios")

		if windowsDataQuickPatchCmdArg {
			winDataPath := help.ExpandPath("~/UMMC/windows/data.win")
			if _, err := os.Stat(winDataPath); err != nil {
				fmt.Printf("Error: Windows data file not found at %s. Did you run 'download-win'?\n", winDataPath)
				return
			}
			fmt.Printf("Copying %s to %s...\n", winDataPath, targetPath)
			if err := help.CopyFile(winDataPath, targetPath, true); err != nil {
				fmt.Printf("Error copying Windows data.win: %v\n", err)
				return
			}
		}

		if forceQuickPatchCmdArg {
			fmt.Println("Using force option! Disabling checksum verification...")
		}

		fmt.Printf("Patching %s with %s...\n", targetPath, filename)

		if err := help.PatchFileForce(targetPath, filename, forceQuickPatchCmdArg); err != nil {
			fmt.Printf("Error patching Undertale: %v\n", err)
			return
		}

		fmt.Printf("Successfully patched Undertale!\n")
	},
}

var modsCmdThingy = &cobra.Command{
	Use:     "mods",
	Aliases: []string{"Mods", "Mod", "mod"},
	Short:   "Manage mods.",
	Args:    cobra.NoArgs,
}

var addModCmdThingy = &cobra.Command{
	Use:     "create [folder]",
	Short:   "Create a mod from a folder.",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"add", "make"},
	Run: func(cmd *cobra.Command, args []string) {
		folder := createModFolderCmdArg
		if folder == "" && len(args) > 0 {
			folder = args[0]
		}

		if folder == "" {
			fmt.Println("Error: No mod folder specified. Provide a folder as an argument or using --folder / -d.")
			return
		}

		folderPath := help.ExpandPath(folder)
		if absPath, err := filepath.Abs(folderPath); err == nil {
			folderPath = absPath
		}
		info, err := os.Stat(folderPath)
		if err != nil || !info.IsDir() {
			fmt.Printf("Error: Mod folder not found or is not a directory at %s\n", folderPath)
			return
		}

		if createModButterscotchCmdArg && !help.IsButterscotchInstalled() {
			fmt.Println("Error: Cannot enable Butterscotch runtime because it is not installed.")
			fmt.Println("Download it first using 'UMMC download-butterscotch' or via the GUI Tools tab.")
			return
		}

		base := createModBaseCmdArg
		isMacos := createModMacosCmdArg
		if createModWinCmdArg {
			isMacos = false
			if !strings.HasSuffix(base, "-w") {
				base = base + "-w"
			}
		}

		rec, err := help.AddModAction(folderPath, createModNameCmdArg, createModMakerCmdArg, base, isMacos, createModVanillaSavesCmdArg, createModRootCmdArg, createModButterscotchCmdArg, forceCreateModCmdArg, nil)
		if err != nil {
			fmt.Printf("Error adding mod: %v\n", err)
			return
		}

		fmt.Printf("Successfully created mod '%s' (Maker: %s, Base: %s, Vanilla Saves: %v, Root Mode: %v, Butterscotch: %v)!\n", rec.Name, rec.Maker, rec.Base, rec.UseVanillaSaves, rec.InstallToAppRoot, rec.UseButterscotch)
	},
}

var listModsCmdThingy = &cobra.Command{
	Use:   "list",
	Short: "List all created mods",
	Run: func(cmd *cobra.Command, args []string) {
		records, err := help.GetMods()
		if err != nil || len(records) == 0 {
			fmt.Println("No mods found in database.")
			return
		}

		fmt.Println("=== Undertale Mods ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tMAKER\tBASE\tROOT MODE\tBUTTERSCOTCH\tVANILLA SAVES")
		fmt.Fprintln(w, "--\t----\t-----\t----\t---------\t------------\t-------------")

		for _, rec := range records {
			vanillaStr := "No"
			if rec.UseVanillaSaves {
				vanillaStr = "Yes"
			}
			rootStr := "No"
			if rec.InstallToAppRoot {
				rootStr = "Yes"
			}
			bsStr := "No"
			if rec.UseButterscotch {
				bsStr = "Yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n", rec.ID, rec.Name, rec.Maker, rec.Base, rootStr, bsStr, vanillaStr)
		}
		w.Flush()
		fmt.Printf("\nTotal mods: %d\n", len(records))
	},
}

var loadModIdCmdArg string
var loadModNameCmdArg string
var loadModSaveCmdArg string
var noLaunchLoadModCmdArg bool

func promptSaveName(promptText, defaultName string) string {
	fmt.Print(promptText)
	var input string
	_, _ = fmt.Scanln(&input)
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultName
	}
	return input
}

var LoadModCmdThingy = &cobra.Command{
	Use:     "play [optional name or id]",
	Aliases: []string{"load", "run", "start"},
	Short:   "Apply a mod from database, switch to its save profile, and launch Undertale",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameToLoad := loadModNameCmdArg
		idToLoad := loadModIdCmdArg

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil && idToLoad == "" {
				idToLoad = args[0]
			} else {
				nameToLoad = args[0]
			}
		}

		if nameToLoad == "" && idToLoad == "" {
			fmt.Println("Error: No mod name or ID specified. Provide an argument or use --name / --id.")
			return
		}

		rec, err := help.GetModByNameOrId(nameToLoad, idToLoad)
		if err != nil || rec == nil {
			fmt.Printf("Error: Mod not found in database: %v\n", err)
			return
		}

		modDirPath := filepath.Join(help.ExpandPath("~/UMMC/mods/"), rec.Name)
		if _, err := os.Stat(modDirPath); os.IsNotExist(err) {
			fmt.Printf("Error: Mod directory not found on disk at %s\n", modDirPath)
			return
		}

		targetSaveMod := rec.Name
		if rec.UseVanillaSaves {
			targetSaveMod = "Vanilla"
			fmt.Printf("Notice: Mod '%s' is configured to use Vanilla save files.\n", rec.Name)
		}

		// 1. Handle active save (preserve current game progress)
		activeDir := help.GetUndertaleSaveDir()
		if help.HasActiveSaveFiles(activeDir) {
			activeInfo, _ := help.GetActiveSaveInfo()
			if activeInfo != nil && activeInfo.Name != "" {
				fmt.Printf("Auto-saving active game progress to '%s' (mod: %s)...\n", activeInfo.Name, activeInfo.ModName)
				_, _ = help.SaveCurrentGame(activeInfo.Name, activeInfo.ModName, true, nil)
			} else {
				defaultSaveName := fmt.Sprintf("untracked_%s", time.Now().Format("20060102_150405"))
				saveName := promptSaveName(fmt.Sprintf("Detected untracked active save data. Enter name to save current game [default: %s]: ", defaultSaveName), defaultSaveName)
				fmt.Printf("Saving current active progress as '%s' (mod: Vanilla)...\n", saveName)
				_, _ = help.SaveCurrentGame(saveName, "Vanilla", true, nil)
			}
		}

		// 2. Handle target mod save profile
		targetSave := loadModSaveCmdArg
		if targetSave != "" {
			fmt.Printf("Loading save '%s' for mod '%s'...\n", targetSave, targetSaveMod)
			if err := help.LoadSave(targetSave, targetSaveMod, nil); err != nil {
				fmt.Printf("Warning: Failed to load specified save '%s': %v\n", targetSave, err)
			}
		} else {
			modSaves, _ := help.GetSaves(targetSaveMod)
			if len(modSaves) > 0 {
				fmt.Printf("Loading save '%s' for mod '%s'...\n", modSaves[0].Name, targetSaveMod)
				_ = help.LoadSave(modSaves[0].Name, targetSaveMod, nil)
			} else {
				defaultNewName := "Save 1"
				newSaveName := promptSaveName(fmt.Sprintf("No save data found for '%s'. Enter new save profile name [default: %s]: ", targetSaveMod, defaultNewName), defaultNewName)
				_, _ = help.CreateEmptySave(newSaveName, targetSaveMod, nil)
				_ = help.LoadSave(newSaveName, targetSaveMod, nil)
			}
		}

		fmt.Printf("Applying mod '%s'...\n", rec.Name)
		if err := help.ApplyModAction(fmt.Sprintf("%d", rec.ID), nil); err != nil {
			fmt.Printf("Error applying mod: %v\n", err)
			return
		}

		if !noLaunchLoadModCmdArg {
			fmt.Println("Launching Undertale...")
			if err := help.LaunchUndertale(nil); err != nil {
				fmt.Printf("Error launching Undertale: %v\n", err)
			}
		}
	},
}

var removeModNameCmdArg string
var removeModIdCmdArg string

var removeModCmdThingy = &cobra.Command{
	Use:     "remove [optional name or id]",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove a mod from database and disk",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameToRemove := removeModNameCmdArg
		idToRemove := removeModIdCmdArg

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil && idToRemove == "" {
				idToRemove = args[0]
			} else {
				nameToRemove = args[0]
			}
		}

		if nameToRemove == "" && idToRemove == "" {
			fmt.Println("Error: No mod name or ID specified. Provide an argument or use --name / --id.")
			return
		}

		rec, err := help.GetModByNameOrId(nameToRemove, idToRemove)
		if err != nil || rec == nil {
			fmt.Printf("Error: Mod not found in database: %v\n", err)
			return
		}

		modDirPath := filepath.Join(help.ExpandPath("~/UMMC/mods/"), rec.Name)
		if err := os.RemoveAll(modDirPath); err != nil {
			fmt.Printf("Warning: Failed to delete mod directory from disk at %s: %v\n", modDirPath, err)
		} else {
			fmt.Printf("Deleted mod folder from disk: %s\n", modDirPath)
		}

		if err := help.DeleteMod(rec.ID); err != nil {
			fmt.Printf("Error deleting mod from database: %v\n", err)
			return
		}

		fmt.Printf("Successfully removed mod '%s' (ID: %d)!\n", rec.Name, rec.ID)
	},
}

var editModNameCmdArg string
var editModIdCmdArg string
var editModNewNameCmdArg string
var editModNewMakerCmdArg string
var editModNewBaseCmdArg string
var editModMacosCmdArg bool
var editModWinCmdArg bool
var editModVanillaSavesCmdArg bool
var editModRootCmdArg bool
var editModButterscotchCmdArg bool

var editModCmdThingy = &cobra.Command{
	Use:     "edit [optional name or id]",
	Aliases: []string{"update", "modify"},
	Short:   "Edit mod metadata and options (name, author/maker, base, macos-mod, root-mode, vanilla saves, butterscotch)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nameToEdit := editModNameCmdArg
		idToEdit := editModIdCmdArg

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil && idToEdit == "" {
				idToEdit = args[0]
			} else {
				nameToEdit = args[0]
			}
		}

		if nameToEdit == "" && idToEdit == "" {
			fmt.Println("Error: No mod name or ID specified. Provide an argument or use --name / --id.")
			return
		}

		rec, err := help.GetModByNameOrId(nameToEdit, idToEdit)
		if err != nil || rec == nil {
			fmt.Printf("Error: Mod not found in database: %v\n", err)
			return
		}

		newName := editModNewNameCmdArg
		if newName == "" {
			newName = rec.Name
		}
		newMaker := editModNewMakerCmdArg
		if newMaker == "" {
			newMaker = rec.Maker
		}
		newBase := editModNewBaseCmdArg
		if newBase == "" {
			newBase = rec.Base
		}

		isMacos := !strings.HasSuffix(rec.Base, "-w")
		if cmd.Flags().Changed("macos-mod") {
			isMacos = editModMacosCmdArg
		}
		if cmd.Flags().Changed("win") && editModWinCmdArg {
			isMacos = false
		}

		useVanilla := rec.UseVanillaSaves
		if cmd.Flags().Changed("vanilla-saves") {
			useVanilla = editModVanillaSavesCmdArg
		}

		rootMode := rec.InstallToAppRoot
		if cmd.Flags().Changed("install-to-app-root") || cmd.Flags().Changed("root") {
			rootMode = editModRootCmdArg
		}

		useButterscotch := rec.UseButterscotch
		if cmd.Flags().Changed("use-butterscotch") || cmd.Flags().Changed("butterscotch") {
			useButterscotch = editModButterscotchCmdArg
			if useButterscotch && !help.IsButterscotchInstalled() {
				fmt.Println("Error: Cannot enable Butterscotch runtime because it is not installed.")
				fmt.Println("Download it first using 'UMMC download-butterscotch' or via the GUI Tools tab.")
				return
			}
		}

		updated, err := help.UpdateModMetadataAction(rec.ID, newName, newMaker, newBase, isMacos, useVanilla, rootMode, useButterscotch, nil)
		if err != nil {
			fmt.Printf("Error updating mod: %v\n", err)
			return
		}

		fmt.Printf("Successfully updated mod '%s' (Maker: %s, Base: %s, Vanilla Saves: %v, Root Mode: %v, Butterscotch: %v)!\n", updated.Name, updated.Maker, updated.Base, updated.UseVanillaSaves, updated.InstallToAppRoot, updated.UseButterscotch)
	},
}

func init() {
	rootCmd.AddCommand(quickPatchCmdThingy)
	rootCmd.AddCommand(modsCmdThingy)
	quickPatchCmdThingy.Flags().BoolVarP(&forceQuickPatchCmdArg, "force", "f", false, "Force patch even when checksums error")
	quickPatchCmdThingy.Flags().StringVarP(&qpFilenameCmdArg, "file", "p", "", "The patch file")
	quickPatchCmdThingy.Flags().BoolVarP(&windowsDataQuickPatchCmdArg, "windows-data", "w", false, "Copy Windows data.win to game.ios before patching")

	modsCmdThingy.AddCommand(addModCmdThingy)
	modsCmdThingy.AddCommand(listModsCmdThingy)
	modsCmdThingy.AddCommand(removeModCmdThingy)
	modsCmdThingy.AddCommand(LoadModCmdThingy)
	modsCmdThingy.AddCommand(editModCmdThingy)
	modsCmdThingy.AddCommand(restoreBackupCmd)

	addModCmdThingy.Flags().StringVarP(&createModFolderCmdArg, "folder", "d", "", "The folder that contains all the mod data")
	addModCmdThingy.Flags().StringVarP(&createModNameCmdArg, "name", "n", "", "The name of the mod")
	addModCmdThingy.Flags().StringVarP(&createModMakerCmdArg, "maker", "m", "", "The author/maker of the mod")
	addModCmdThingy.Flags().StringVarP(&createModBaseCmdArg, "base", "b", "", "The base game version/type for the mod")
	addModCmdThingy.Flags().BoolVarP(&createModWinCmdArg, "win", "w", false, "Set base to Windows version (appends -w)")
	addModCmdThingy.Flags().BoolVar(&createModMacosCmdArg, "macos-mod", false, "Specify that this is a macOS mod (prevents appending '-w' to base)")
	addModCmdThingy.Flags().BoolVar(&createModVanillaSavesCmdArg, "vanilla-saves", false, "Make this mod use and share vanilla Undertale save files")
	addModCmdThingy.Flags().BoolVar(&createModRootCmdArg, "install-to-app-root", false, "Install mod directly into UNDERTALE.app root instead of Contents/Resources")
	addModCmdThingy.Flags().BoolVar(&createModRootCmdArg, "root", false, "Alias for --install-to-app-root")
	addModCmdThingy.Flags().BoolVar(&createModButterscotchCmdArg, "use-butterscotch", false, "Use the experimental Butterscotch GameMaker runner for this mod")
	addModCmdThingy.Flags().BoolVar(&createModButterscotchCmdArg, "butterscotch", false, "Alias for --use-butterscotch")
	addModCmdThingy.Flags().BoolVarP(&forceCreateModCmdArg, "force", "f", false, "Force overwrite if mod already exists")

	removeModCmdThingy.Flags().StringVarP(&removeModNameCmdArg, "name", "n", "", "The name of the mod to remove")
	removeModCmdThingy.Flags().StringVarP(&removeModIdCmdArg, "id", "i", "", "The ID of the mod to remove")

	editModCmdThingy.Flags().StringVarP(&editModNameCmdArg, "name", "n", "", "The current name of the mod to edit")
	editModCmdThingy.Flags().StringVarP(&editModIdCmdArg, "id", "i", "", "The ID of the mod to edit")
	editModCmdThingy.Flags().StringVar(&editModNewNameCmdArg, "new-name", "", "New name for the mod")
	editModCmdThingy.Flags().StringVar(&editModNewMakerCmdArg, "maker", "", "New author/maker for the mod")
	editModCmdThingy.Flags().StringVar(&editModNewBaseCmdArg, "base", "", "New base version for the mod")
	editModCmdThingy.Flags().BoolVar(&editModMacosCmdArg, "macos-mod", false, "Set whether this is a macOS mod")
	editModCmdThingy.Flags().BoolVarP(&editModWinCmdArg, "win", "w", false, "Set whether this is a Windows mod (appends -w)")
	editModCmdThingy.Flags().BoolVar(&editModVanillaSavesCmdArg, "vanilla-saves", false, "Set whether this mod uses Vanilla save files")
	editModCmdThingy.Flags().BoolVar(&editModRootCmdArg, "install-to-app-root", false, "Set whether to install mod into UNDERTALE.app root")
	editModCmdThingy.Flags().BoolVar(&editModRootCmdArg, "root", false, "Alias for --install-to-app-root")
	editModCmdThingy.Flags().BoolVar(&editModButterscotchCmdArg, "use-butterscotch", false, "Set whether to use the Butterscotch runner for this mod")
	editModCmdThingy.Flags().BoolVar(&editModButterscotchCmdArg, "butterscotch", false, "Alias for --use-butterscotch")

	LoadModCmdThingy.Flags().StringVarP(&loadModNameCmdArg, "name", "n", "", "The name of the mod to play")
	LoadModCmdThingy.Flags().StringVarP(&loadModIdCmdArg, "id", "i", "", "The ID of the mod to play")
	LoadModCmdThingy.Flags().StringVarP(&loadModSaveCmdArg, "save", "s", "", "The specific save profile to load with the mod")
	LoadModCmdThingy.Flags().BoolVar(&noLaunchLoadModCmdArg, "no-launch", false, "Apply mod and switch save without launching Undertale")
}
