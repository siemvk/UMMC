package cmds

import (
	"UMMC/help"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type ModXmlPatch struct {
	Type  string `xml:"type,attr"`
	Patch string `xml:"patch,attr"`
	To    string `xml:"to,attr"`
}

func ParseModdingXml(xmlPath string) ([]ModXmlPatch, error) {
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, err
	}

	wrapped := append([]byte("<patches>"), data...)
	wrapped = append(wrapped, []byte("</patches>")...)

	type PatchesContainer struct {
		Patches []ModXmlPatch `xml:"patch"`
	}

	var container PatchesContainer
	if err := xml.Unmarshal(wrapped, &container); err != nil {
		return nil, err
	}

	return container.Patches, nil
}

func ResolveGameKeyFromTarget(toPath string) (string, int) {
	toLower := strings.ToLower(toPath)
	if strings.Contains(toLower, "main data") || strings.Contains(toLower, "main") || strings.Contains(toLower, "menu") || strings.Contains(toLower, "mus") || strings.Contains(toLower, "music") {
		return "deltarune-mus", 0
	}
	if strings.Contains(toLower, "ch1") || strings.Contains(toLower, "chapter1") || strings.Contains(toLower, "chapter 1") {
		return "deltarune-ch1", 1
	}
	if strings.Contains(toLower, "ch2") || strings.Contains(toLower, "chapter2") || strings.Contains(toLower, "chapter 2") {
		return "deltarune-ch2", 2
	}
	if strings.Contains(toLower, "ch3") || strings.Contains(toLower, "chapter3") || strings.Contains(toLower, "chapter 3") {
		return "deltarune-ch3", 3
	}
	if strings.Contains(toLower, "ch4") || strings.Contains(toLower, "chapter4") || strings.Contains(toLower, "chapter 4") {
		return "deltarune-ch4", 4
	}
	if strings.Contains(toLower, "ch5") || strings.Contains(toLower, "chapter5") || strings.Contains(toLower, "chapter 5") {
		return "deltarune-ch5", 5
	}
	if strings.Contains(toLower, "deltarune") {
		return "deltarune-ch1", 1
	}
	return "undertale", 0
}

func IsTargetDirectory(dirName string) bool {
	toLower := strings.ToLower(dirName)
	return strings.Contains(toLower, "chapter") || strings.Contains(toLower, "ch") || strings.Contains(toLower, "main") || strings.Contains(toLower, "mus") || strings.Contains(toLower, "music") || strings.Contains(toLower, "menu")
}

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

		game, chapter := ParseGameArg(GameCmdArg, ChapterCmdArg)
		gCfg := help.GetGameConfig(game)

		appPath := help.ExpandPath(gCfg.Path)
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

		fmt.Printf("Patching %s (%s) with %s...\n", gCfg.Name, targetPath, filename)

		if err := help.PatchFileForce(targetPath, filename, forceQuickPatchCmdArg); err != nil {
			fmt.Printf("Error patching game file: %v\n", err)
			return
		}

		fmt.Printf("Successfully patched %s!\n", gCfg.Name)
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

		var dInfoMetaName string
		var dInfoMetaAuthor string
		deltaModPath := filepath.Join(folderPath, "_deltamodinfo.json")
		if data, err := os.ReadFile(deltaModPath); err == nil {
			var dInfo struct {
				Metadata struct {
					Name   string      `json:"name"`
					Author interface{} `json:"author"`
				} `json:"metadata"`
			}
			if err := json.Unmarshal(data, &dInfo); err == nil {
				if dInfo.Metadata.Name != "" {
					dInfoMetaName = dInfo.Metadata.Name
					fmt.Printf("Loaded metadata name '%s' from _deltamodinfo.json\n", dInfoMetaName)
				}
				if authorStr, ok := dInfo.Metadata.Author.(string); ok && authorStr != "" {
					dInfoMetaAuthor = authorStr
				} else if authorSlice, ok := dInfo.Metadata.Author.([]interface{}); ok && len(authorSlice) > 0 {
					var authors []string
					for _, a := range authorSlice {
						if s, ok := a.(string); ok {
							authors = append(authors, s)
						}
					}
					dInfoMetaAuthor = strings.Join(authors, ", ")
				}
			}
		}

		hasModdingXml := false
		xmlPath := filepath.Join(folderPath, "modding.xml")
		if _, err := os.Stat(xmlPath); err == nil {
			hasModdingXml = true
			fmt.Println("Detected modding.xml manifest")
		}

		hasTargetSubdirs := false
		folderEntries, _ := os.ReadDir(folderPath)
		for _, fe := range folderEntries {
			if fe.IsDir() && IsTargetDirectory(fe.Name()) {
				hasTargetSubdirs = true
				break
			}
		}

		var game string
		var chapter int
		if cmd.Flags().Changed("game") || GameCmdArg != "" || ChapterCmdArg > 0 {
			game, chapter = ParseGameArg(GameCmdArg, ChapterCmdArg)
		} else if hasModdingXml || hasTargetSubdirs {
			game = "deltarune-ch1"
			chapter = 1
		} else if modCfg != nil && strings.Contains(strings.ToLower(modCfg.Metadata.Game), "deltarune") {
			game = "deltarune-ch1"
			chapter = 1
			for _, tag := range modCfg.Metadata.Tags {
				tLower := strings.ToLower(tag)
				if strings.Contains(tLower, "ch2") || strings.Contains(tLower, "chapter2") || strings.Contains(tLower, "chapter 2") {
					game = "deltarune-ch2"
					chapter = 2
				}
			}
		} else {
			game, chapter = ParseGameArg("", ChapterCmdArg)
		}

		gCfg := help.GetGameConfig(game)

		name := createModNameCmdArg
		if name == "" && modCfg != nil && modCfg.Metadata.Name != "" {
			name = modCfg.Metadata.Name
		}
		if name == "" && dInfoMetaName != "" {
			name = dInfoMetaName
		}
		if name == "" {
			name = filepath.Base(folderPath)
		}

		maker := createModMakerCmdArg
		if maker == "" && modCfg != nil && modCfg.Metadata.Author != "" {
			maker = modCfg.Metadata.Author
		}
		if maker == "" && dInfoMetaAuthor != "" {
			maker = dInfoMetaAuthor
		}
		if maker == "" {
			maker = "Unknown"
		}

		base := createModBaseCmdArg
		if base == "" && modCfg != nil && modCfg.Metadata.GameVersion != "" {
			base = modCfg.Metadata.GameVersion
		}
		if base == "" {
			base = gCfg.DefaultVersion
			if base == "" {
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

		if _, err := help.AddMod(base, name, maker, installToAppRoot, gCfg.ID, chapter); err != nil {
			fmt.Printf("Warning: Created mod folder at %s but failed to save DB entry: %v\n", dstPath, err)
		}

		fmt.Printf("Successfully created mod '%s' for %s (Maker: %s, Base: %s, InstallToAppRoot: %v)!\n", name, gCfg.Name, maker, base, installToAppRoot)
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
		filterChapter := 0
		if cmd.Flags().Changed("game") || GameCmdArg != "" {
			filterGame, filterChapter = ParseGameArg(GameCmdArg, ChapterCmdArg)
		} else if ChapterCmdArg > 0 {
			filterGame = "deltarune"
			filterChapter = ChapterCmdArg
		}

		fmt.Println("=== Mods Database ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tGAME\tNAME\tMAKER\tBASE\tAPP ROOT")
		fmt.Fprintln(w, "--\t----\t----\t-----\t----\t--------")

		count := 0
		for _, rec := range records {
			if filterGame != "" && strings.ToLower(rec.Game) != filterGame {
				continue
			}
			if filterChapter > 0 && rec.Chapter != filterChapter {
				continue
			}
			gameKey, _ := help.ParseGameKey(rec.Game, rec.Chapter)
			gCfg := help.GetGameConfig(gameKey)
			gameDisplay := gCfg.Name
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%t\n", rec.ID, gameDisplay, rec.Name, rec.Maker, rec.Base, rec.InstallToAppRoot)
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
		if cmd.Flags().Changed("game") || GameCmdArg != "" {
			game, chapter = ParseGameArg(GameCmdArg, ChapterCmdArg)
		} else if ChapterCmdArg > 0 {
			chapter = ChapterCmdArg
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

		handledSubdirs := make(map[string]bool)

		xmlPath := filepath.Join(modDirPath, "modding.xml")
		if _, errXml := os.Stat(xmlPath); errXml == nil {
			fmt.Println("Found modding.xml manifest, processing patch mappings...")
			patches, errParse := ParseModdingXml(xmlPath)
			if errParse != nil {
				fmt.Printf("Warning: Failed to parse modding.xml: %v\n", errParse)
			} else {
				for _, p := range patches {
					if p.Patch == "" || p.To == "" {
						continue
					}
					patchRelative := filepath.Clean(p.Patch)
					patchFile := filepath.Join(modDirPath, patchRelative)
					targetGameKey, targetCh := ResolveGameKeyFromTarget(p.To)
					targetDataPath := help.GetGameDataPath(targetAppPath, targetGameKey, targetCh)

					if _, errStat := os.Stat(targetDataPath); os.IsNotExist(errStat) {
						fmt.Printf("Notice: Skipping %s patch '%s' because target data file does not exist at %s\n", targetGameKey, patchRelative, targetDataPath)
						continue
					}

					fmt.Printf("Applying XML patch '%s' -> %s (%s)...\n", patchRelative, targetGameKey, targetDataPath)
					if err := help.PatchFileForce(targetDataPath, patchFile, true); err != nil {
						fmt.Printf("Warning: Patch %s failed: %v\n", patchRelative, err)
					}
				}
			}
		} else {
			// First: search for and process target subfolders (e.g. "Chapter 1", "Chapter 2", "main data", "mus")
			for _, entry := range entries {
				if entry.IsDir() {
					nameLower := strings.ToLower(entry.Name())
					if IsTargetDirectory(nameLower) {
						handledSubdirs[entry.Name()] = true
						targetGameKey, targetCh := ResolveGameKeyFromTarget(entry.Name())
						subDirPath := filepath.Join(modDirPath, entry.Name())
						subEntries, errSub := os.ReadDir(subDirPath)
						if errSub != nil {
							continue
						}

						targetDataPath := help.GetGameDataPath(targetAppPath, targetGameKey, targetCh)
						targetResDir := help.GetGameResourceDir(targetAppPath, targetGameKey, targetCh)

						for _, subEntry := range subEntries {
							subNameLower := strings.ToLower(subEntry.Name())
							srcFile := filepath.Join(subDirPath, subEntry.Name())
							if strings.HasSuffix(subNameLower, ".xdelta") {
								if _, errStat := os.Stat(targetDataPath); os.IsNotExist(errStat) {
									fmt.Printf("Notice: Skipping %s patch '%s' because target data file does not exist at %s\n", targetGameKey, subEntry.Name(), targetDataPath)
									continue
								}
								fmt.Printf("Applying folder patch '%s/%s' -> %s (%s)...\n", entry.Name(), subEntry.Name(), targetGameKey, targetDataPath)
								if err := help.PatchFileForce(targetDataPath, srcFile, true); err != nil {
									fmt.Printf("Warning: Patch %s/%s failed: %v\n", entry.Name(), subEntry.Name(), err)
								}
							} else {
								var dstFile string
								if subNameLower == "data.win" {
									dstFile = targetDataPath
								} else {
									dstFile = filepath.Join(targetResDir, subEntry.Name())
								}
								fmt.Printf("Copying folder overlay '%s/%s' -> %s...\n", entry.Name(), subEntry.Name(), dstFile)
								if err := help.CopyOverlay(srcFile, dstFile, true); err != nil {
									fmt.Printf("Warning: Failed to copy %s/%s: %v\n", entry.Name(), subEntry.Name(), err)
								}
							}
						}
					}
				} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".xdelta") {
					patchFile := filepath.Join(modDirPath, entry.Name())
					fmt.Printf("Applying root patch '%s' to game data...\n", entry.Name())
					if err := help.PatchFileForce(targetGameIosPath, patchFile, true); err != nil {
						fmt.Printf("Error applying mod patch: %v\n", err)
						return
					}
				}
			}
		}

		// Second: copy all non-metadata files and unhandled folders from mod directory to destination
		for _, entry := range entries {
			if handledSubdirs[entry.Name()] {
				continue
			}
			nameLower := strings.ToLower(entry.Name())
			if strings.HasSuffix(nameLower, ".xdelta") || nameLower == "mod_config.json" || nameLower == "_deltamodinfo.json" || nameLower == "modding.xml" || nameLower == "_icon.png" || nameLower == "data_patches" || nameLower == "readme.txt" || nameLower == ".ds_store" {
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

		gCfg := help.GetGameConfig(game)
		fmt.Printf("Successfully applied mod '%s' for %s!\n", rec.Name, gCfg.Name)
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

