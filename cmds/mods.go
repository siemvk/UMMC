package cmds

import (
	"UMMC/help"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type ModConfig struct {
	ConfigVersion string `json:"config_version"`
	Metadata      struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Author      string   `json:"author"`
		Description string   `json:"description"`
		Game        string   `json:"game"`
		GameVersion string   `json:"game_version"`
		Tags        []string `json:"tags"`
	} `json:"metadata"`
	InstallToAppRoot bool                   `json:"install_to_app_root"`
	Files            map[string]interface{} `json:"files"`
}

var qpFilenameCmdArg string
var forceQuickPatchCmdArg bool
var windowsDataQuickPatchCmdArg bool

var createModFolderCmdArg string
var createModNameCmdArg string
var createModMakerCmdArg string
var createModBaseCmdArg string
var createModWinCmdArg bool
var forceCreateModCmdArg bool
var createModMacosCmdArg bool
var createModInstallToAppRootCmdArg bool
var loadModInstallToAppRootCmdArg bool

var quickPatchCmdThingy = &cobra.Command{
	Use:   "quickpatch [filename]",
	Short: "patch a mod onto the global Undertale or Deltarune install",
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

		game := "undertale"
		chapter := 0
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			game = "undertale"
		} else if cmd.Flags().Changed("deltarune") || DeltaruneCmdArg > 0 {
			game = "deltarune"
			chapter = DeltaruneCmdArg
			if chapter <= 0 {
				chapter = 1
			}
		}

		appPath := help.GetDefaultAppPath(game)
		targetPath := help.GetGameDataPath(appPath, game, chapter)

		if windowsDataQuickPatchCmdArg {
			winDataPath := help.FindWindowsDataWin(game, chapter)

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

		if game == "deltarune" {
			fmt.Printf("Patching Deltarune Chapter %d (%s) with %s...\n", chapter, targetPath, filename)
		} else {
			fmt.Printf("Patching %s with %s...\n", targetPath, filename)
		}

		if err := help.PatchFileForce(targetPath, filename, forceQuickPatchCmdArg); err != nil {
			fmt.Printf("Error patching game file: %v\n", err)
			return
		}

		if game == "deltarune" {
			fmt.Printf("Successfully patched Deltarune Chapter %d!\n", chapter)
		} else {
			fmt.Printf("Successfully patched Undertale!\n")
		}
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
			fmt.Println("Error: No mod folder specified. Provide a folder as an argument or using --folder / -F.")
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

		var modCfg *ModConfig
		cfgPath := filepath.Join(folderPath, "mod_config.json")
		if data, err := os.ReadFile(cfgPath); err == nil {
			var cfg ModConfig
			if err := json.Unmarshal(data, &cfg); err == nil {
				modCfg = &cfg
				fmt.Println("Loaded metadata from mod_config.json")
			}
		}

		game := "undertale"
		chapter := 0
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			game = "undertale"
		} else if cmd.Flags().Changed("deltarune") || DeltaruneCmdArg > 0 {
			game = "deltarune"
			chapter = DeltaruneCmdArg
			if chapter <= 0 {
				chapter = 1
			}
		} else if modCfg != nil && strings.Contains(strings.ToLower(modCfg.Metadata.Game), "deltarune") {
			game = "deltarune"
			chapter = 1
			for _, tag := range modCfg.Metadata.Tags {
				tLower := strings.ToLower(tag)
				if strings.Contains(tLower, "ch2") || strings.Contains(tLower, "chapter2") || strings.Contains(tLower, "chapter 2") {
					chapter = 2
				} else if strings.Contains(tLower, "ch1") || strings.Contains(tLower, "chapter1") || strings.Contains(tLower, "chapter 1") {
					chapter = 1
				}
			}
		}

		name := createModNameCmdArg
		if name == "" && modCfg != nil && modCfg.Metadata.Name != "" {
			name = modCfg.Metadata.Name
		}
		if name == "" {
			name = filepath.Base(folderPath)
		}

		maker := createModMakerCmdArg
		if maker == "" && modCfg != nil && modCfg.Metadata.Author != "" {
			maker = modCfg.Metadata.Author
		}
		if maker == "" {
			maker = "Unknown"
		}

		base := createModBaseCmdArg
		if base == "" && modCfg != nil && modCfg.Metadata.GameVersion != "" {
			base = modCfg.Metadata.GameVersion
		}
		if base == "" {
			if game == "deltarune" {
				base = fmt.Sprintf("deltarune-ch%d", chapter)
			} else if createModWinCmdArg {
				base = "win"
			} else {
				base = "1.08"
			}
		}

		if (createModWinCmdArg || !createModMacosCmdArg) && !strings.HasSuffix(base, "-w") {
			base = base + "-w"
		}

		dstPath := filepath.Join(help.ExpandPath("~/UMMC/mods/"), name)

		if err := help.CopyDir(folderPath, dstPath, forceCreateModCmdArg); err != nil {
			fmt.Printf("Error copying mod directory: %v\n", err)
			return
		}

		installToAppRoot := createModInstallToAppRootCmdArg || (modCfg != nil && modCfg.InstallToAppRoot)

		if installToAppRoot {
			dstCfgPath := filepath.Join(dstPath, "mod_config.json")
			var cfg map[string]interface{}
			if data, err := os.ReadFile(dstCfgPath); err == nil {
				_ = json.Unmarshal(data, &cfg)
			}
			if cfg == nil {
				cfg = make(map[string]interface{})
			}
			cfg["install_to_app_root"] = true
			if updatedData, err := json.MarshalIndent(cfg, "", "  "); err == nil {
				_ = os.WriteFile(dstCfgPath, updatedData, 0644)
			}
		}

		if _, err := help.AddMod(base, name, maker, installToAppRoot, game, chapter); err != nil {
			fmt.Printf("Warning: Created mod folder at %s but failed to save DB entry: %v\n", dstPath, err)
		}

		if game == "deltarune" {
			fmt.Printf("Successfully created mod '%s' for Deltarune Chapter %d (Maker: %s, Base: %s, InstallToAppRoot: %v)!\n", name, chapter, maker, base, installToAppRoot)
		} else {
			fmt.Printf("Successfully created mod '%s' (Maker: %s, Base: %s, InstallToAppRoot: %v)!\n", name, maker, base, installToAppRoot)
		}
		fmt.Printf("The mod is now stored in %s so you can safely delete the folder you gave as input and the mod will continue working.\n", dstPath)
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

		filterGame := ""
		filterChapter := DeltaruneCmdArg
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			filterGame = "undertale"
			filterChapter = 0
		} else if cmd.Flags().Changed("deltarune") || filterChapter > 0 {
			filterGame = "deltarune"
		}

		fmt.Println("=== Mods Database ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tGAME\tCHAPTER\tNAME\tMAKER\tBASE\tAPP ROOT")
		fmt.Fprintln(w, "--\t----\t-------\t----\t-----\t----\t--------")

		count := 0
		for _, rec := range records {
			if filterGame != "" && strings.ToLower(rec.Game) != filterGame {
				continue
			}
			if filterChapter > 0 && rec.Chapter != filterChapter {
				continue
			}
			gameDisplay := "Undertale"
			chDisplay := "-"
			if strings.ToLower(rec.Game) == "deltarune" {
				gameDisplay = "Deltarune"
				if rec.Chapter > 0 {
					chDisplay = fmt.Sprintf("Ch %d", rec.Chapter)
				} else {
					chDisplay = "Ch 1"
				}
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%t\n", rec.ID, gameDisplay, chDisplay, rec.Name, rec.Maker, rec.Base, rec.InstallToAppRoot)
			count++
		}
		w.Flush()
		fmt.Printf("\nTotal mods listed: %d\n", count)
	},
}

var loadModIdCmdArg string
var loadModNameCmdArg string
var noLaunchLoadModCmdArg bool

var LoadModCmdThingy = &cobra.Command{
	Use:     "play [optional name or id]",
	Aliases: []string{"load", "run", "start"},
	Short:   "Apply a mod from database and launch game",
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

		game := rec.Game
		chapter := rec.Chapter
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			game = "undertale"
			chapter = 0
		} else if cmd.Flags().Changed("deltarune") || DeltaruneCmdArg > 0 {
			game = "deltarune"
			if DeltaruneCmdArg > 0 {
				chapter = DeltaruneCmdArg
			}
		}

		if strings.ToLower(game) == "deltarune" && chapter <= 0 {
			chapter = 1
		}

		targetAppPath := help.GetDefaultAppPath(game)
		if _, err := os.Stat(targetAppPath); err != nil {
			fmt.Printf("Error: Game app not found at %s\n", targetAppPath)
			return
		}
		targetGameIosPath := help.GetGameDataPath(targetAppPath, game, chapter)

		// Restore base game backup if available
		backupRec, errBackup := help.GetBackupByVersionOrId(rec.Base, "")
		if errBackup == nil && backupRec != nil && backupRec.BackupPath != "" {
			fmt.Printf("Restoring base game backup version %s...\n", rec.Base)
			if err := help.CopyFile(backupRec.BackupPath, targetAppPath, true); err != nil {
				fmt.Printf("Warning: Failed to restore base backup: %v\n", err)
			}
		} else {
			cleanBase := strings.TrimSuffix(rec.Base, "-w")
			if backupRec2, errBackup2 := help.GetBackupByVersionOrId(cleanBase, ""); errBackup2 == nil && backupRec2 != nil && backupRec2.BackupPath != "" {
				fmt.Printf("Restoring base game backup version %s...\n", cleanBase)
				if err := help.CopyFile(backupRec2.BackupPath, targetAppPath, true); err != nil {
					fmt.Printf("Warning: Failed to restore base backup: %v\n", err)
				}
				if strings.HasSuffix(rec.Base, "-w") {
					winDataPath := help.FindWindowsDataWin(game, chapter)
					if _, errStat := os.Stat(winDataPath); errStat == nil {
						fmt.Printf("Injecting Windows data.win (%s) for -w base...\n", winDataPath)
						_ = help.CopyFile(winDataPath, targetGameIosPath, true)
						_ = os.WriteFile(filepath.Join(filepath.Dir(targetGameIosPath), "winpatchdetect"), []byte("injected"), 0644)
					}
				}
			} else {
				fmt.Printf("Notice: No backup found for base '%s'. Applying mod directly onto current game install.\n", rec.Base)
				if strings.HasSuffix(rec.Base, "-w") {
					winDataPath := help.FindWindowsDataWin(game, chapter)
					if _, errStat := os.Stat(winDataPath); errStat == nil {
						fmt.Printf("Injecting Windows data.win (%s) for -w base...\n", winDataPath)
						_ = help.CopyFile(winDataPath, targetGameIosPath, true)
						_ = os.WriteFile(filepath.Join(filepath.Dir(targetGameIosPath), "winpatchdetect"), []byte("injected"), 0644)
					}
				}
			}
		}


		// Check install to app root preference
		installToAppRoot := loadModInstallToAppRootCmdArg || rec.InstallToAppRoot
		if !installToAppRoot {
			cfgPath := filepath.Join(modDirPath, "mod_config.json")
			if data, err := os.ReadFile(cfgPath); err == nil {
				var cfg ModConfig
				if err := json.Unmarshal(data, &cfg); err == nil {
					if cfg.InstallToAppRoot {
						installToAppRoot = true
					}
				}
			}
		}

		entries, errDir := os.ReadDir(modDirPath)
		if errDir != nil {
			fmt.Printf("Error reading mod directory: %v\n", errDir)
			return
		}

		var destDir string
		if installToAppRoot {
			destDir = targetAppPath
			fmt.Println("Installing mod files to app root directory...")
		} else {
			destDir = help.GetGameResourceDir(targetAppPath, game, chapter)
		}

		// First: search for and apply any .xdelta patch file in the mod folder
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".xdelta") {
				patchFile := filepath.Join(modDirPath, entry.Name())
				fmt.Printf("Applying patch '%s' to game data...\n", entry.Name())
				if err := help.PatchFileForce(targetGameIosPath, patchFile, true); err != nil {
					fmt.Printf("Error applying mod patch: %v\n", err)
					return
				}
			}
		}

		// Second: copy all files and folders from mod directory to destination
		for _, entry := range entries {
			nameLower := strings.ToLower(entry.Name())
			if strings.HasSuffix(nameLower, ".xdelta") || nameLower == "mod_config.json" {
				continue
			}

			src := filepath.Join(modDirPath, entry.Name())
			var dst string
			if nameLower == "data.win" {
				dst = targetGameIosPath
			} else {
				dst = filepath.Join(destDir, entry.Name())
			}

			if err := help.CopyOverlay(src, dst, true); err != nil {
				fmt.Printf("Warning: Failed to copy %s: %v\n", entry.Name(), err)
			}
		}

		if strings.ToLower(game) == "deltarune" {
			fmt.Printf("Successfully applied mod '%s' for Deltarune Chapter %d!\n", rec.Name, chapter)
		} else {
			fmt.Printf("Successfully applied mod '%s'!\n", rec.Name)
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
	modsCmdThingy.AddCommand(restoreBackupCmd)

	addModCmdThingy.Flags().StringVarP(&createModFolderCmdArg, "folder", "F", "", "The folder that contains all the mod data")
	addModCmdThingy.Flags().StringVarP(&createModNameCmdArg, "name", "n", "", "The name of the mod")
	addModCmdThingy.Flags().StringVarP(&createModMakerCmdArg, "maker", "m", "", "The author/maker of the mod")
	addModCmdThingy.Flags().StringVarP(&createModBaseCmdArg, "base", "b", "", "The base game version/type for the mod")
	addModCmdThingy.Flags().BoolVarP(&createModWinCmdArg, "win", "w", false, "Set base to 'win' (Windows version)")
	addModCmdThingy.Flags().BoolVar(&createModMacosCmdArg, "macos-mod", false, "Specify that this is a macOS mod (prevents appending '-w' to base)")
	addModCmdThingy.Flags().BoolVarP(&forceCreateModCmdArg, "force", "f", false, "Force overwrite if mod already exists")
	addModCmdThingy.Flags().BoolVar(&createModInstallToAppRootCmdArg, "install-to-app-root", false, "Install mod files to app root instead of chapter/resources directory")

	removeModCmdThingy.Flags().StringVarP(&removeModNameCmdArg, "name", "n", "", "The name of the mod to remove")
	removeModCmdThingy.Flags().StringVarP(&removeModIdCmdArg, "id", "i", "", "The ID of the mod to remove")

	LoadModCmdThingy.Flags().StringVarP(&loadModNameCmdArg, "name", "n", "", "The name of the mod to play")
	LoadModCmdThingy.Flags().StringVarP(&loadModIdCmdArg, "id", "i", "", "The ID of the mod to play")
	LoadModCmdThingy.Flags().BoolVar(&loadModInstallToAppRootCmdArg, "install-to-app-root", false, "Install mod files to app root instead of chapter/resources directory")
}

