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

		var gameDir string
		if len(args) > 0 {
			gameDir = help.ExpandPath(args[0])
		} else {
			if game == "deltarune" {
				gameDir = help.ExpandPath("~/Library/Application Support/Steam/steamapps/common/DELTARUNE/")
			} else {
				gameDir = help.ExpandPath("~/Library/Application Support/Steam/steamapps/common/Undertale/")
			}
		}

		if forceCreateCmdArg {
			fmt.Println("Using force option! Will overwrite existing backup...")
		}

		appName := "UNDERTALE.app"
		if game == "deltarune" {
			appName = "DELTARUNE.app"
		}

		var appPath string
		if strings.HasSuffix(gameDir, ".app") || filepath.Base(gameDir) == appName {
			appPath = gameDir
		} else if _, err := os.Stat(filepath.Join(gameDir, appName)); err == nil {
			appPath = filepath.Join(gameDir, appName)
		} else {
			appPath = gameDir
		}

		_, err := os.Stat(appPath)
		if errors.Is(err, os.ErrNotExist) || err != nil {
			fmt.Printf("Error: Game app not found at %s\nTry providing your own path.\n", appPath)
			return
		}

		fmt.Printf("Found game app at %s\n", appPath)

		version := versionCreateCmdArg
		if game == "deltarune" && (!cmd.Flags().Changed("version") || version == "1.08") {
			version = fmt.Sprintf("deltarune-ch%d", chapter)
		}

		isWinPatched := help.IsWinPatched(appPath, game, chapter)

		if isWinPatched {
			if !strings.HasSuffix(version, "-w") {
				version = version + "-w"
			}
			fmt.Println("Detected Windows data injection (winpatchdetect)! Appending '-w' to backup version.")
		}


		var backupDir string
		if game == "deltarune" {
			backupDir = help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/deltarune-ch%d-%s", chapter, version))
		} else {
			backupDir = help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", version))
		}
		dstPath := filepath.Join(backupDir, filepath.Base(appPath))

		if err := help.CopyFile(appPath, dstPath, forceCreateCmdArg); err != nil {
			fmt.Printf("Error backing up game app: %v\n", err)
			return
		}

		if _, err := help.AddBackup(version, appPath, dstPath, game, chapter); err != nil {
			fmt.Printf("Warning: Created backup on disk but failed to save DB entry: %v\n", err)
		}

		if game == "deltarune" {
			fmt.Printf("Successfully backed up Deltarune Chapter %d (version %s) to %s\n", chapter, version, dstPath)
		} else {
			fmt.Printf("Successfully backed up Undertale version %s to %s\n", version, dstPath)
		}
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

		targetLocation := help.GetDefaultAppPath(game)

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
		} else {
			if idToRestore != "" {
				fmt.Printf("Warning: Backup ID %s not found in database, attempting fallback path search.\n", idToRestore)
			}
			var backupSrcDir string
			if game == "deltarune" {
				backupSrcDir = help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/deltarune-ch%d-%s", chapter, versionToRestore))
			} else {
				backupSrcDir = help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", versionToRestore))
			}
			appName := "UNDERTALE.app"
			if game == "deltarune" {
				appName = "DELTARUNE.app"
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
		appName := "UNDERTALE.app"
		if game == "deltarune" {
			appName = "DELTARUNE.app"
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

		if game == "deltarune" {
			fmt.Printf("Successfully restored Deltarune (Chapter %d) version %s to %s\n", chapter, versionToRestore, dstAppPath)
		} else {
			fmt.Printf("Successfully restored Undertale version %s to %s\n", versionToRestore, dstAppPath)
		}
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
		filterChapter := DeltaruneCmdArg
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			filterGame = "undertale"
			filterChapter = 0
		} else if cmd.Flags().Changed("deltarune") || filterChapter > 0 {
			filterGame = "deltarune"
		}


		fmt.Println("=== Backups Database ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tGAME\tCHAPTER\tVERSION\tCREATED AT\tBACKUP PATH")
		fmt.Fprintln(w, "--\t----\t-------\t-------\t----------\t-----------")

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
			timeStr := rec.CreatedAt.Local().Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", rec.ID, gameDisplay, chDisplay, rec.Version, timeStr, rec.BackupPath)
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


