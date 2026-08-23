package cmds

import (
	"UMMC/help"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// Variable to hold our flag state
var forceCreateCmdArg bool
var versionCreateCmdArg string
var forceRestoreCmdArg bool
var versionRestoreCmdArg string
var idRestoreCmdArg string

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Commands related to backing up your gamefiles",
}

var createBackupCmd = &cobra.Command{
	Use:   "create [optional game location]",
	Short: "Make a backup of your gamefiles (Undertale or Deltarune)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		game, chapter := ParseGameArg(GameCmdArg, ChapterCmdArg)
		gCfg := help.GetGameConfig(game)

		var appPath string
		if len(args) > 0 {
			appPath = help.ExpandPath(args[0])
		} else {
			appPath = help.ExpandPath(gCfg.Path)
		}

		if forceCreateCmdArg {
			fmt.Println("Using force option! Will overwrite existing backup...")
		}

		_, err := os.Stat(appPath)
		if errors.Is(err, os.ErrNotExist) || err != nil {
			fmt.Printf("Error: Game app not found at %s\nTry providing your own path or updating ~/UMMC/config.json.\n", appPath)
			return
		}

		fmt.Printf("Found game app at %s\n", appPath)

		version := versionCreateCmdArg
		if !cmd.Flags().Changed("version") && gCfg.DefaultVersion != "" {
			version = gCfg.DefaultVersion
		}

		if help.IsWinPatched(appPath, game, chapter) {
			if !strings.HasSuffix(version, "-w") {
				version = version + "-w"
			}
			fmt.Println("Detected Windows data injection (winpatchdetect)! Appending '-w' to backup version.")
		}

		backupFolder := gCfg.BackupDir
		if backupFolder == "" {
			backupFolder = gCfg.ID
		}
		backupDir := help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/%s-%s", backupFolder, version))
		dstPath := filepath.Join(backupDir, filepath.Base(appPath))

		if err := help.CopyFile(appPath, dstPath, forceCreateCmdArg); err != nil {
			fmt.Printf("Error backing up game app: %v\n", err)
			return
		}

		if _, err := help.AddBackup(version, appPath, dstPath, gCfg.ID, chapter); err != nil {
			fmt.Printf("Warning: Created backup on disk but failed to save DB entry: %v\n", err)
		}

		fmt.Printf("Successfully backed up %s (version %s) to %s\n", gCfg.Name, version, dstPath)
	},
}

var restoreBackupCmd = &cobra.Command{
	Use:     "restore [optional version or id] [optional target location]",
	Short:   "Restore a game backup version",
	Aliases: []string{"vanila"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		versionToRestore := versionRestoreCmdArg
		idToRestore := idRestoreCmdArg
		game, chapter := ParseGameArg(GameCmdArg, ChapterCmdArg)
		_ = chapter
		gCfg := help.GetGameConfig(game)

		targetLocation := help.ExpandPath(gCfg.Path)

		if !cmd.Flags().Changed("version") && idToRestore != "" {
			versionToRestore = ""
		}

		if len(args) == 1 {
			if strings.Contains(args[0], "/") || strings.HasSuffix(args[0], ".app") {
				targetLocation = args[0]
			} else if _, err := strconv.Atoi(args[0]); err == nil && idToRestore == "" {
				idToRestore = args[0]
				if !cmd.Flags().Changed("version") {
					versionToRestore = ""
				}
			} else {
				versionToRestore = args[0]
			}
		} else if len(args) >= 2 {
			if _, err := strconv.Atoi(args[0]); err == nil && idToRestore == "" {
				idToRestore = args[0]
				if !cmd.Flags().Changed("version") {
					versionToRestore = ""
				}
			} else {
				versionToRestore = args[0]
			}
			targetLocation = args[1]
		}

		if forceRestoreCmdArg {
			fmt.Println("Using force option! Will overwrite existing game copy...")
		}

		var srcAppPath string
		rec, dbErr := help.GetBackupByVersionOrId(versionToRestore, idToRestore)
		if dbErr == nil && rec != nil && rec.BackupPath != "" {
			srcAppPath = rec.BackupPath
			if rec.AppPath != "" {
				targetLocation = rec.AppPath
			}
			versionToRestore = rec.Version
			game = rec.Game
			chapter = rec.Chapter
			gCfg = help.GetGameConfig(game)
		} else {
			if idToRestore != "" {
				fmt.Printf("Warning: Backup ID %s not found in database, attempting fallback path search.\n", idToRestore)
			}
			backupFolder := gCfg.BackupDir
			if backupFolder == "" {
				backupFolder = gCfg.ID
			}
			backupSrcDir := help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/%s-%s", backupFolder, versionToRestore))
			appName := filepath.Base(gCfg.Path)
			if appName == "" || !strings.HasSuffix(appName, ".app") {
				appName = "UNDERTALE.app"
			}
			if strings.HasSuffix(backupSrcDir, ".app") || filepath.Base(backupSrcDir) == appName {
				srcAppPath = backupSrcDir
			} else if _, err := os.Stat(filepath.Join(backupSrcDir, appName)); err == nil {
				srcAppPath = filepath.Join(backupSrcDir, appName)
			} else {
				srcAppPath = backupSrcDir
			}
		}

		_, err := os.Stat(srcAppPath)
		if errors.Is(err, os.ErrNotExist) || err != nil {
			fmt.Printf("Error: Backup for version %s not found at %s\nAre you sure you created the backup?\n", versionToRestore, srcAppPath)
			return
		}

		targetDir := help.ExpandPath(targetLocation)
		appName := filepath.Base(gCfg.Path)
		if appName == "" || !strings.HasSuffix(appName, ".app") {
			appName = "UNDERTALE.app"
		}
		var dstAppPath string
		if strings.HasSuffix(targetDir, ".app") || filepath.Base(targetDir) == appName {
			dstAppPath = targetDir
		} else {
			dstAppPath = filepath.Join(targetDir, appName)
		}

		if err := help.CopyFile(srcAppPath, dstAppPath, forceRestoreCmdArg); err != nil {
			fmt.Printf("Error restoring game backup: %v\n", err)
			return
		}

		fmt.Printf("Successfully restored %s version %s to %s\n", gCfg.Name, versionToRestore, dstAppPath)
	},
}

var listBackupCmd = &cobra.Command{
	Use:   "list",
	Short: "List all created game backups",
	Run: func(cmd *cobra.Command, args []string) {
		records, err := help.GetBackups()
		if err != nil || len(records) == 0 {
			fmt.Println("No backups found in database.")
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


		fmt.Println("=== Backups Database ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tGAME\tVERSION\tCREATED AT\tBACKUP PATH")
		fmt.Fprintln(w, "--\t----\t-------\t----------\t-----------")

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
			timeStr := rec.CreatedAt.Local().Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", rec.ID, gameDisplay, rec.Version, timeStr, rec.BackupPath)
			count++
		}
		w.Flush()
		fmt.Printf("\nTotal backups listed: %d\n", count)
	},
}

var removeVersionCmdArg string
var removeIdCmdArg string

var removeBackupCmd = &cobra.Command{
	Use:     "remove [optional version or id]",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove a backup from database and disk",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		versionToRemove := removeVersionCmdArg
		idToRemove := removeIdCmdArg

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil && idToRemove == "" {
				idToRemove = args[0]
			} else {
				versionToRemove = args[0]
			}
		}

		if versionToRemove == "" && idToRemove == "" {
			fmt.Println("Error: No backup version or ID specified. Provide an argument or use --version / --id.")
			return
		}

		rec, err := help.GetBackupByVersionOrId(versionToRemove, idToRemove)
		if err != nil || rec == nil {
			fmt.Printf("Error: Backup not found in database: %v\n", err)
			return
		}

		if rec.BackupPath != "" {
			deletePath := help.ExpandPath(rec.BackupPath)
			if filepath.Base(deletePath) == "UNDERTALE.app" || filepath.Base(deletePath) == "DELTARUNE.app" {
				deletePath = filepath.Dir(deletePath)
			}
			if err := os.RemoveAll(deletePath); err != nil {
				fmt.Printf("Warning: Failed to delete backup files from disk at %s: %v\n", deletePath, err)
			} else {
				fmt.Printf("Deleted backup folder from disk: %s\n", deletePath)
			}
		}

		if err := help.DeleteBackup(rec.ID); err != nil {
			fmt.Printf("Error deleting backup from database: %v\n", err)
			return
		}

		fmt.Printf("Successfully removed backup version '%s' (ID: %d)!\n", rec.Version, rec.ID)
	},
}

// init() runs automatically before main()
func init() {
	// Attach this command to the root
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(createBackupCmd)
	backupCmd.AddCommand(restoreBackupCmd)
	modsCmdThingy.AddCommand(restoreBackupCmd)
	backupCmd.AddCommand(listBackupCmd)
	backupCmd.AddCommand(removeBackupCmd)

	// Define flags specific to the 'create' command
	createBackupCmd.Flags().BoolVarP(&forceCreateCmdArg, "force", "f", false, "Force overwrite if a backup already exists")
	createBackupCmd.Flags().StringVarP(&versionCreateCmdArg, "version", "v", "1.08", "The version of game we are backing up.")

	// Define flags specific to the 'restore' command
	restoreBackupCmd.Flags().BoolVarP(&forceRestoreCmdArg, "force", "f", false, "Force overwrite if game already exists at target location")
	restoreBackupCmd.Flags().StringVarP(&versionRestoreCmdArg, "version", "v", "1.08", "The version of game to restore.")
	restoreBackupCmd.Flags().StringVarP(&idRestoreCmdArg, "id", "i", "", "The ID of the backup to restore.")

	// Define flags specific to the 'remove' command
	removeBackupCmd.Flags().StringVarP(&removeVersionCmdArg, "version", "v", "", "The version of backup to remove")
	removeBackupCmd.Flags().StringVarP(&removeIdCmdArg, "id", "i", "", "The ID of the backup to remove")
}


