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
	Use:   "create [optional undertale location]",
	Short: "Make a backup of your undertale gamefiles",
	Args:  cobra.MaximumNArgs(1), // Ensures the user provides at most one argument
	Run: func(cmd *cobra.Command, args []string) {
		var gameDir string
		if len(args) > 0 {
			gameDir = help.ExpandPath(args[0])
		} else {
			gameDir = help.ExpandPath("~/Library/Application Support/Steam/steamapps/common/Undertale/")
		}

		if forceCreateCmdArg {
			fmt.Println("Using force option! Will overwrite existing backup...")
		}

		var appPath string
		if strings.HasSuffix(gameDir, ".app") || filepath.Base(gameDir) == "UNDERTALE.app" {
			appPath = gameDir
		} else if _, err := os.Stat(filepath.Join(gameDir, "UNDERTALE.app")); err == nil {
			appPath = filepath.Join(gameDir, "UNDERTALE.app")
		} else {
			appPath = gameDir
		}

		_, err := os.Stat(appPath)
		if errors.Is(err, os.ErrNotExist) || err != nil {
			fmt.Printf("Error: Undertale not found at %s\nTry providing your own path to Undertale.\n", appPath)
			return
		}

		fmt.Printf("Found Undertale at %s\n", appPath)

		version := versionCreateCmdArg
		winDetectPaths := []string{
			filepath.Join(appPath, "Contents/Resources/winpatchdetect"),
			filepath.Join(appPath, "winpatchdetect"),
			filepath.Join(filepath.Dir(appPath), "winpatchdetect"),
		}

		isWinPatched := false
		for _, detectPath := range winDetectPaths {
			if _, err := os.Stat(detectPath); err == nil {
				isWinPatched = true
				break
			}
		}

		if isWinPatched {
			if !strings.HasSuffix(version, "-w") {
				version = version + "-w"
			}
			fmt.Println("Detected Windows data injection (winpatchdetect)! Appending '-w' to backup version.")
		}

		backupDir := help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", version))
		dstPath := filepath.Join(backupDir, filepath.Base(appPath))

		if err := help.CopyFile(appPath, dstPath, forceCreateCmdArg); err != nil {
			fmt.Printf("Error backing up Undertale: %v\n", err)
			return
		}

		if db, err := help.GetDB(); err == nil {
			defer db.Close()
			_, _ = db.Exec("INSERT INTO backups (version, app_path, backup_path) VALUES (?, ?, ?)", version, appPath, dstPath)
		}

		fmt.Printf("Successfully backed up Undertale version %s to %s\n", version, dstPath)
	},
}

var restoreBackupCmd = &cobra.Command{
	Use:     "restore [optional version or id] [optional undertale target location]",
	Short:   "Restore an undertale backup version",
	Aliases: []string{"vanila"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		versionToRestore := versionRestoreCmdArg
		idToRestore := idRestoreCmdArg
		targetLocation := "~/Library/Application Support/Steam/steamapps/common/Undertale/"

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
			fmt.Println("Using force option! Will overwrite existing undertale copy...")
		}

		var srcAppPath string
		rec, dbErr := help.GetBackupByVersionOrId(versionToRestore, idToRestore)
		if dbErr == nil && rec != nil && rec.BackupPath != "" {
			srcAppPath = rec.BackupPath
			if rec.AppPath != "" {
				targetLocation = rec.AppPath
			}
			versionToRestore = rec.Version
		} else {
			if idToRestore != "" {
				fmt.Printf("Warning: Backup ID %s not found in database, attempting fallback path search.\n", idToRestore)
			}
			backupSrcDir := help.ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", versionToRestore))
			if strings.HasSuffix(backupSrcDir, ".app") || filepath.Base(backupSrcDir) == "UNDERTALE.app" {
				srcAppPath = backupSrcDir
			} else if _, err := os.Stat(filepath.Join(backupSrcDir, "UNDERTALE.app")); err == nil {
				srcAppPath = filepath.Join(backupSrcDir, "UNDERTALE.app")
			} else {
				srcAppPath = backupSrcDir
			}
		}

		_, err := os.Stat(srcAppPath)
		if errors.Is(err, os.ErrNotExist) || err != nil {
			fmt.Printf("Error: Backup for Undertale version %s not found at %s\nAre you sure you created the backup?\n", versionToRestore, srcAppPath)
			return
		}

		targetDir := help.ExpandPath(targetLocation)
		var dstAppPath string
		if strings.HasSuffix(targetDir, ".app") || filepath.Base(targetDir) == "UNDERTALE.app" {
			dstAppPath = targetDir
		} else {
			dstAppPath = filepath.Join(targetDir, "UNDERTALE.app")
		}

		if err := help.CopyFile(srcAppPath, dstAppPath, forceRestoreCmdArg); err != nil {
			fmt.Printf("Error restoring Undertale: %v\n", err)
			return
		}

		fmt.Printf("Successfully restored Undertale version %s to %s\n", versionToRestore, dstAppPath)
	},
}

var listBackupCmd = &cobra.Command{
	Use:   "list",
	Short: "List all created undertale backups",
	Run: func(cmd *cobra.Command, args []string) {
		records, err := help.GetBackups()
		if err != nil || len(records) == 0 {
			fmt.Println("No backups found in database.")
			return
		}

		fmt.Println("=== Undertale Backups ===")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tVERSION\tCREATED AT\tBACKUP PATH")
		fmt.Fprintln(w, "--\t-------\t----------\t-----------")

		for _, rec := range records {
			timeStr := rec.CreatedAt.Local().Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", rec.ID, rec.Version, timeStr, rec.BackupPath)
		}
		w.Flush()
		fmt.Printf("\nTotal backups: %d\n", len(records))
	},
}

var removeVersionCmdArg string
var removeIdCmdArg string

var removeBackupCmd = &cobra.Command{
	Use:     "remove [optional version or id]",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove an undertale backup from database and disk",
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
			if filepath.Base(deletePath) == "UNDERTALE.app" {
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
	createBackupCmd.Flags().StringVarP(&versionCreateCmdArg, "version", "v", "1.08", "The version of undertale we are backing up.")

	// Define flags specific to the 'restore' command
	restoreBackupCmd.Flags().BoolVarP(&forceRestoreCmdArg, "force", "f", false, "Force overwrite if Undertale already exists at target location")
	restoreBackupCmd.Flags().StringVarP(&versionRestoreCmdArg, "version", "v", "1.08", "The version of undertale to restore.")
	restoreBackupCmd.Flags().StringVarP(&idRestoreCmdArg, "id", "i", "", "The ID of the backup of undertale to restore.")

	// Define flags specific to the 'remove' command
	removeBackupCmd.Flags().StringVarP(&removeVersionCmdArg, "version", "v", "", "The version of undertale backup to remove")
	removeBackupCmd.Flags().StringVarP(&removeIdCmdArg, "id", "i", "", "The ID of the undertale backup to remove")
}

