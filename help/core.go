package help

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Default paths
const (
	DefaultSteamUndertaleApp = "~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app"
	DefaultSteamUndertaleDir = "~/Library/Application Support/Steam/steamapps/common/Undertale"
	DefaultWindowsDataPath   = "~/UMMC/windows/data.win"
	DefaultWindowsDataAlt    = "~/UMMC/windows/Undertale/data.win"
	DefaultModsDir           = "~/UMMC/mods"
	DefaultBackupsDir        = "~/UMMC/Backup"
	DefaultButterscotchDir   = "~/UMMC/runtimes/butterscotch"
	ButterscotchArm64URL     = "https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build/main/butterscotch-appkit-macos-arm64.zip"
	ButterscotchX86_64URL    = "https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build/main/butterscotch-appkit-macos-x86_64.zip"
)

type SystemStatus struct {
	UndertaleAppPath      string
	UndertaleInstalled    bool
	WindowsDataPath       string
	WindowsDataExists     bool
	XdeltaPath            string
	XdeltaAvailable       bool
	SteamcmdPath          string
	SteamcmdAvailable     bool
	ButterscotchPath      string
	ButterscotchAvailable bool
	TotalBackups          int
	TotalMods             int
}

// LogFunc is a logging callback for operations.
type LogFunc func(msg string)

func defaultLog(msg string) {
	fmt.Println(msg)
}

// CheckSystemStatus inspects the system and returns current status.
func CheckSystemStatus() SystemStatus {
	var status SystemStatus

	// Check Undertale App
	appPath := FindUndertaleApp("")
	status.UndertaleAppPath = appPath
	if appPath != "" {
		if _, err := os.Stat(appPath); err == nil {
			status.UndertaleInstalled = true
		}
	}

	// Check Windows Data
	winData := FindWindowsDataPath("")
	status.WindowsDataPath = winData
	if winData != "" {
		if _, err := os.Stat(winData); err == nil {
			status.WindowsDataExists = true
		}
	}

	// Check xdelta3
	if bin, err := exec.LookPath("xdelta3"); err == nil {
		status.XdeltaPath = bin
		status.XdeltaAvailable = true
	} else if _, err := os.Stat("/opt/homebrew/bin/xdelta3"); err == nil {
		status.XdeltaPath = "/opt/homebrew/bin/xdelta3"
		status.XdeltaAvailable = true
	} else if _, err := os.Stat("/usr/local/bin/xdelta3"); err == nil {
		status.XdeltaPath = "/usr/local/bin/xdelta3"
		status.XdeltaAvailable = true
	}

	// Check steamcmd
	if bin, err := exec.LookPath("steamcmd"); err == nil {
		status.SteamcmdPath = bin
		status.SteamcmdAvailable = true
	} else if _, err := os.Stat("/opt/homebrew/bin/steamcmd"); err == nil {
		status.SteamcmdPath = "/opt/homebrew/bin/steamcmd"
		status.SteamcmdAvailable = true
	} else if _, err := os.Stat("/usr/local/bin/steamcmd"); err == nil {
		status.SteamcmdPath = "/usr/local/bin/steamcmd"
		status.SteamcmdAvailable = true
	}

	// Check Butterscotch runtime
	bsDir := ExpandPath(DefaultButterscotchDir)
	status.ButterscotchPath = bsDir
	if fi, err := os.Stat(bsDir); err == nil && fi.IsDir() {
		entries, _ := os.ReadDir(bsDir)
		if len(entries) > 0 {
			status.ButterscotchAvailable = true
		}
	}

	// Count backups & mods
	if backups, err := GetBackups(); err == nil {
		status.TotalBackups = len(backups)
	}
	if mods, err := GetMods(); err == nil {
		status.TotalMods = len(mods)
	}

	return status
}

// FindUndertaleApp searches for the UNDERTALE.app bundle.
func FindUndertaleApp(customPath string) string {
	candidates := []string{
		customPath,
		DefaultSteamUndertaleApp,
		filepath.Join(DefaultSteamUndertaleDir, "UNDERTALE.app"),
		"/Applications/UNDERTALE.app",
		"~/Applications/UNDERTALE.app",
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		exp := ExpandPath(c)
		if strings.HasSuffix(exp, ".app") {
			if _, err := os.Stat(exp); err == nil {
				return exp
			}
		} else {
			appInDir := filepath.Join(exp, "UNDERTALE.app")
			if _, err := os.Stat(appInDir); err == nil {
				return appInDir
			}
		}
	}

	// Return standard default if not found
	return ExpandPath(DefaultSteamUndertaleApp)
}

// FindWindowsDataPath locates data.win from downloaded Windows Undertale.
func FindWindowsDataPath(customPath string) string {
	candidates := []string{
		customPath,
		DefaultWindowsDataPath,
		DefaultWindowsDataAlt,
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		exp := ExpandPath(c)
		if _, err := os.Stat(exp); err == nil {
			return exp
		}
	}
	return ExpandPath(DefaultWindowsDataPath)
}

// CreateBackupAction creates a backup of Undertale.
func CreateBackupAction(gameDir, version string, force bool, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	if gameDir == "" {
		gameDir = FindUndertaleApp("")
	} else {
		gameDir = ExpandPath(gameDir)
	}

	var appPath string
	if strings.HasSuffix(gameDir, ".app") || filepath.Base(gameDir) == "UNDERTALE.app" {
		appPath = gameDir
	} else if _, err := os.Stat(filepath.Join(gameDir, "UNDERTALE.app")); err == nil {
		appPath = filepath.Join(gameDir, "UNDERTALE.app")
	} else {
		appPath = gameDir
	}

	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("Undertale not found at %s: %w", appPath, err)
	}

	log(fmt.Sprintf("Found Undertale at: %s", appPath))

	if version == "" {
		version = "1.08"
	}

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
		log("Detected Windows data injection (winpatchdetect). Using version: " + version)
	}

	backupDir := ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", version))
	dstPath := filepath.Join(backupDir, filepath.Base(appPath))

	log(fmt.Sprintf("Copying game files to backup folder %s...", dstPath))
	if err := CopyFile(appPath, dstPath, force); err != nil {
		return fmt.Errorf("error backing up Undertale: %w", err)
	}

	if db, err := GetDB(); err == nil {
		defer db.Close()
		_, _ = db.Exec("INSERT INTO backups (version, app_path, backup_path) VALUES (?, ?, ?)", version, appPath, dstPath)
	}

	log(fmt.Sprintf("Successfully backed up Undertale version %s to %s", version, dstPath))
	return nil
}

// RestoreBackupAction restores a backup to Undertale location.
func RestoreBackupAction(versionOrId string, customTarget string, force bool, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	if versionOrId == "" {
		return errors.New("no backup version or ID specified")
	}

	var version string
	var id string
	if _, err := strconv.Atoi(versionOrId); err == nil {
		id = versionOrId
	} else {
		version = versionOrId
	}

	targetLocation := DefaultSteamUndertaleApp
	if customTarget != "" {
		targetLocation = customTarget
	}

	var srcAppPath string
	rec, dbErr := GetBackupByVersionOrId(version, id)
	if dbErr == nil && rec != nil && rec.BackupPath != "" {
		srcAppPath = rec.BackupPath
		if rec.AppPath != "" && customTarget == "" {
			targetLocation = rec.AppPath
		}
		version = rec.Version
	} else {
		backupSrcDir := ExpandPath(fmt.Sprintf("~/UMMC/Backup/undertale%s", version))
		if strings.HasSuffix(backupSrcDir, ".app") || filepath.Base(backupSrcDir) == "UNDERTALE.app" {
			srcAppPath = backupSrcDir
		} else if _, err := os.Stat(filepath.Join(backupSrcDir, "UNDERTALE.app")); err == nil {
			srcAppPath = filepath.Join(backupSrcDir, "UNDERTALE.app")
		} else {
			srcAppPath = backupSrcDir
		}
	}

	if _, err := os.Stat(srcAppPath); err != nil {
		return fmt.Errorf("backup for Undertale %s not found at %s: %w", versionOrId, srcAppPath, err)
	}

	targetDir := ExpandPath(targetLocation)
	var dstAppPath string
	if strings.HasSuffix(targetDir, ".app") || filepath.Base(targetDir) == "UNDERTALE.app" {
		dstAppPath = targetDir
	} else {
		dstAppPath = filepath.Join(targetDir, "UNDERTALE.app")
	}

	log(fmt.Sprintf("Restoring backup '%s' from %s to %s...", version, srcAppPath, dstAppPath))
	if err := CopyFile(srcAppPath, dstAppPath, force); err != nil {
		return fmt.Errorf("error restoring Undertale: %w", err)
	}

	_ = os.Remove(filepath.Join(dstAppPath, "Contents/Resources/butterscotch"))
	_ = os.Remove(filepath.Join(dstAppPath, "Contents/MacOS/butterscotch"))

	log(fmt.Sprintf("Successfully restored Undertale version %s!", version))
	return nil
}

// RemoveBackupAction deletes a backup from disk and DB.
func RemoveBackupAction(versionOrId string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	var version string
	var id string
	if _, err := strconv.Atoi(versionOrId); err == nil {
		id = versionOrId
	} else {
		version = versionOrId
	}

	rec, err := GetBackupByVersionOrId(version, id)
	if err != nil || rec == nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	if rec.BackupPath != "" {
		deletePath := ExpandPath(rec.BackupPath)
		if filepath.Base(deletePath) == "UNDERTALE.app" {
			deletePath = filepath.Dir(deletePath)
		}
		if err := os.RemoveAll(deletePath); err != nil {
			log(fmt.Sprintf("Warning: Failed to delete folder %s: %v", deletePath, err))
		} else {
			log(fmt.Sprintf("Deleted backup folder: %s", deletePath))
		}
	}

	if err := DeleteBackup(rec.ID); err != nil {
		return fmt.Errorf("error deleting backup record from DB: %w", err)
	}

	log(fmt.Sprintf("Successfully removed backup '%s' (ID: %d)", rec.Version, rec.ID))
	return nil
}

// ModConfigFile represents optional metadata in a mod directory.
type ModConfigFile struct {
	ConfigVersion string `json:"config_version"`
	Metadata      struct {
		ID               string   `json:"id"`
		Name             string   `json:"name"`
		Version          string   `json:"version"`
		Author           string   `json:"author"`
		Description      string   `json:"description"`
		Game             string   `json:"game"`
		GameVersion      string   `json:"game_version"`
		Tags             []string `json:"tags"`
		UseVanillaSaves  bool     `json:"use_vanilla_saves,omitempty"`
		InstallToAppRoot bool     `json:"install_to_app_root,omitempty"`
		UseButterscotch  bool     `json:"use_butterscotch,omitempty"`
	} `json:"metadata"`
	Files map[string]interface{} `json:"files"`
}

// ReadModConfig attempts to parse mod_config.json in a directory.
func ReadModConfig(dirPath string) (*ModConfigFile, error) {
	cfgPath := filepath.Join(ExpandPath(dirPath), "mod_config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg ModConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsButterscotchInstalled returns true if the Butterscotch runtime is downloaded and present.
func IsButterscotchInstalled() bool {
	dir := ExpandPath(DefaultButterscotchDir)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return false
	}
	return true
}

// FindButterscotchRunner locates the executable binary inside the Butterscotch runtime folder.
func FindButterscotchRunner() (string, error) {
	dir := ExpandPath(DefaultButterscotchDir)
	if !IsButterscotchInstalled() {
		return "", errors.New("butterscotch runtime is not installed")
	}

	candidates := []string{
		filepath.Join(dir, "butterscotch"),
		filepath.Join(dir, "Butterscotch"),
		filepath.Join(dir, "Mac_Runner"),
		filepath.Join(dir, "Butterscotch.app/Contents/MacOS/butterscotch"),
		filepath.Join(dir, "Butterscotch.app/Contents/MacOS/Butterscotch"),
		filepath.Join(dir, "Butterscotch.app/Contents/MacOS/Mac_Runner"),
	}

	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand, nil
		}
	}

	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && !strings.HasSuffix(e.Name(), ".json") && !strings.HasSuffix(e.Name(), ".txt") && !strings.HasSuffix(e.Name(), ".zip") {
				return filepath.Join(dir, e.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("no runner binary found in %s", dir)
}

// AddModAction copies a mod directory to ~/UMMC/mods/<name> and registers it.
func AddModAction(folderPath, name, maker, base string, isMacos, useVanillaSaves, installToAppRoot, useButterscotch, force bool, log LogFunc) (*ModRecord, error) {
	if log == nil {
		log = defaultLog
	}

	if useButterscotch && !IsButterscotchInstalled() {
		return nil, fmt.Errorf("butterscotch runtime is not installed. Please download it first using 'UMMC download-butterscotch' or in the GUI Tools tab")
	}

	folderPath = ExpandPath(folderPath)
	absPath, err := filepath.Abs(folderPath)
	if err == nil {
		folderPath = absPath
	}

	info, err := os.Stat(folderPath)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("mod folder not found or not a directory: %s", folderPath)
	}

	modCfg, _ := ReadModConfig(folderPath)
	if modCfg != nil {
		log("Loaded metadata from mod_config.json")
		if modCfg.Metadata.UseVanillaSaves {
			useVanillaSaves = true
		}
		if modCfg.Metadata.InstallToAppRoot {
			installToAppRoot = true
		}
		if modCfg.Metadata.UseButterscotch {
			if !IsButterscotchInstalled() {
				log("Notice: Mod config requests Butterscotch, but Butterscotch runtime is not installed.")
			} else {
				useButterscotch = true
			}
		}
	}

	if name == "" && modCfg != nil && modCfg.Metadata.Name != "" {
		name = modCfg.Metadata.Name
	}
	if name == "" {
		name = filepath.Base(folderPath)
	}

	if maker == "" && modCfg != nil && modCfg.Metadata.Author != "" {
		maker = modCfg.Metadata.Author
	}
	if maker == "" {
		maker = "Unknown"
	}

	if base == "" && modCfg != nil && modCfg.Metadata.GameVersion != "" {
		base = modCfg.Metadata.GameVersion
	}
	if base == "" {
		base = "1.08"
	}

	if !isMacos && !strings.HasSuffix(base, "-w") {
		base = base + "-w"
	} else if isMacos && strings.HasSuffix(base, "-w") {
		base = strings.TrimSuffix(base, "-w")
	}

	dstPath := filepath.Join(ExpandPath("~/UMMC/mods/"), name)
	log(fmt.Sprintf("Installing mod '%s' to %s...", name, dstPath))

	if err := CopyDir(folderPath, dstPath, force); err != nil {
		return nil, fmt.Errorf("error copying mod directory: %w", err)
	}

	// Update mod_config.json in destination
	cfgDstPath := filepath.Join(dstPath, "mod_config.json")
	var cfgToSave ModConfigFile
	if cfgData, err := os.ReadFile(cfgDstPath); err == nil {
		_ = json.Unmarshal(cfgData, &cfgToSave)
	}
	cfgToSave.Metadata.Name = name
	cfgToSave.Metadata.Author = maker
	cfgToSave.Metadata.GameVersion = base
	cfgToSave.Metadata.UseVanillaSaves = useVanillaSaves
	cfgToSave.Metadata.InstallToAppRoot = installToAppRoot
	cfgToSave.Metadata.UseButterscotch = useButterscotch
	if updatedJson, err := json.MarshalIndent(cfgToSave, "", "  "); err == nil {
		_ = os.WriteFile(cfgDstPath, updatedJson, 0644)
	}

	id, err := AddMod(base, name, maker, useVanillaSaves, installToAppRoot, useButterscotch)
	if err != nil {
		log(fmt.Sprintf("Warning: failed to write to database: %v", err))
	}

	log(fmt.Sprintf("Successfully installed mod '%s' (Maker: %s, Base: %s, Vanilla Saves: %v, Root Mode: %v, Butterscotch: %v)!", name, maker, base, useVanillaSaves, installToAppRoot, useButterscotch))
	return &ModRecord{ID: int(id), Base: base, Name: name, Maker: maker, UseVanillaSaves: useVanillaSaves, InstallToAppRoot: installToAppRoot, UseButterscotch: useButterscotch}, nil
}

// ApplyModAction applies the specified mod to Undertale.
func ApplyModAction(nameOrId string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	var name string
	var id string
	if _, err := strconv.Atoi(nameOrId); err == nil {
		id = nameOrId
	} else {
		name = nameOrId
	}

	rec, err := GetModByNameOrId(name, id)
	if err != nil || rec == nil {
		return fmt.Errorf("mod not found: %w", err)
	}

	modDirPath := filepath.Join(ExpandPath("~/UMMC/mods/"), rec.Name)
	if _, err := os.Stat(modDirPath); os.IsNotExist(err) {
		return fmt.Errorf("mod directory not found on disk at %s", modDirPath)
	}

	targetAppPath := FindUndertaleApp("")
	if _, err := os.Stat(targetAppPath); err != nil {
		return fmt.Errorf("Undertale app not found at %s", targetAppPath)
	}
	targetGameIosPath := filepath.Join(targetAppPath, "Contents/Resources/game.ios")

	// 1. Restore base game backup if available
	backupRec, errBackup := GetBackupByVersionOrId(rec.Base, "")
	if errBackup == nil && backupRec != nil && backupRec.BackupPath != "" {
		log(fmt.Sprintf("Restoring base game backup version %s...", rec.Base))
		if err := CopyFile(backupRec.BackupPath, targetAppPath, true); err != nil {
			log(fmt.Sprintf("Warning: Failed to restore base backup: %v", err))
		}
	} else {
		cleanBase := strings.TrimSuffix(rec.Base, "-w")
		if backupRec2, errBackup2 := GetBackupByVersionOrId(cleanBase, ""); errBackup2 == nil && backupRec2 != nil && backupRec2.BackupPath != "" {
			log(fmt.Sprintf("Restoring base game backup version %s...", cleanBase))
			if err := CopyFile(backupRec2.BackupPath, targetAppPath, true); err != nil {
				log(fmt.Sprintf("Warning: Failed to restore base backup: %v", err))
			}
			if strings.HasSuffix(rec.Base, "-w") {
				winDataPath := FindWindowsDataPath("")
				if _, errStat := os.Stat(winDataPath); errStat == nil {
					log("Injecting Windows data.win for -w base...")
					_ = CopyFile(winDataPath, targetGameIosPath, true)
					_ = os.WriteFile(filepath.Join(filepath.Dir(targetGameIosPath), "winpatchdetect"), []byte("injected"), 0644)
				}
			}
		} else {
			log(fmt.Sprintf("Notice: No backup found for base '%s'. Applying mod directly onto current Undertale install.", rec.Base))
		}
	}

	entries, errDir := os.ReadDir(modDirPath)
	if errDir != nil {
		return fmt.Errorf("error reading mod directory: %w", errDir)
	}

	if rec.InstallToAppRoot {
		log("Applying mod in Root Mode (copying into UNDERTALE.app root)...")
		for _, entry := range entries {
			nameLower := strings.ToLower(entry.Name())
			if nameLower == "mod_config.json" {
				continue
			}
			src := filepath.Join(modDirPath, entry.Name())
			dst := filepath.Join(targetAppPath, entry.Name())

			if entry.IsDir() {
				errWalk := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return err
					}
					rel, _ := filepath.Rel(src, path)
					targetDest := filepath.Join(dst, rel)
					if d.IsDir() {
						return os.MkdirAll(targetDest, 0755)
					}
					if err := CopyFile(path, targetDest, true); err != nil {
						log(fmt.Sprintf("Warning: Failed to copy %s: %v", rel, err))
					}
					if strings.Contains(targetDest, "/MacOS/") || strings.HasSuffix(targetDest, "/Mac_Runner") {
						_ = os.Chmod(targetDest, 0755)
					}
					return nil
				})
				if errWalk != nil {
					log(fmt.Sprintf("Warning: Failed copying directory %s: %v", entry.Name(), errWalk))
				}
			} else {
				if strings.HasSuffix(nameLower, ".xdelta") {
					log(fmt.Sprintf("Applying patch '%s' to Undertale...", entry.Name()))
					_ = PatchFileForce(targetGameIosPath, src, true)
				} else {
					if err := CopyFile(src, dst, true); err != nil {
						log(fmt.Sprintf("Warning: Failed to copy %s: %v", entry.Name(), err))
					}
				}
			}
		}
	} else {
		// Normal mode: Assets go into Contents/Resources/
		resourcesDir := filepath.Join(targetAppPath, "Contents/Resources")

		// 2. Apply any .xdelta patch in the mod folder
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".xdelta") {
				patchFile := filepath.Join(modDirPath, entry.Name())
				log(fmt.Sprintf("Applying patch '%s' to Undertale...", entry.Name()))
				if err := PatchFileForce(targetGameIosPath, patchFile, true); err != nil {
					return fmt.Errorf("error applying mod patch: %w", err)
				}
			}
		}

		// 3. Copy all other files to Contents/Resources/
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
				dst = filepath.Join(resourcesDir, entry.Name())
			}

			if err := CopyFile(src, dst, true); err != nil {
				log(fmt.Sprintf("Warning: Failed to copy %s: %v", entry.Name(), err))
			}
		}
	}

	// 4. Inject Butterscotch runner if enabled for this mod
	resourcesDir := filepath.Join(targetAppPath, "Contents/Resources")
	macosDir := filepath.Join(targetAppPath, "Contents/MacOS")

	if rec.UseButterscotch {
		if !IsButterscotchInstalled() {
			return fmt.Errorf("mod '%s' is configured to use the Butterscotch runner, but Butterscotch is not installed. Download it first using 'UMMC download-butterscotch' or via the GUI Tools tab", rec.Name)
		}
		runnerPath, errRunner := FindButterscotchRunner()
		if errRunner != nil {
			return fmt.Errorf("failed to locate Butterscotch runner binary: %w", errRunner)
		}

		_ = os.MkdirAll(resourcesDir, 0755)
		_ = os.MkdirAll(macosDir, 0755)

		// 1. Put Butterscotch runner in the folder with data/game.ios
		bsInResources := filepath.Join(resourcesDir, "butterscotch")
		log(fmt.Sprintf("Placing Butterscotch runner in %s with game data...", resourcesDir))
		if err := CopyFile(runnerPath, bsInResources, true); err != nil {
			log(fmt.Sprintf("Warning: Failed to copy Butterscotch runner to %s: %v", bsInResources, err))
		}
		_ = os.Chmod(bsInResources, 0755)

		// Also copy as Mac_Runner in Resources
		_ = CopyFile(runnerPath, filepath.Join(resourcesDir, "Mac_Runner"), true)
		_ = os.Chmod(filepath.Join(resourcesDir, "Mac_Runner"), 0755)

		// 2. Put Butterscotch runner in Contents/MacOS/
		dstRunner := filepath.Join(macosDir, "Mac_Runner")
		log(fmt.Sprintf("Injecting Butterscotch runner into %s...", dstRunner))
		if err := CopyFile(runnerPath, dstRunner, true); err != nil {
			log(fmt.Sprintf("Warning: Failed to copy Butterscotch runner to Mac_Runner: %v", err))
		}
		_ = os.Chmod(dstRunner, 0755)

		_ = CopyFile(runnerPath, filepath.Join(macosDir, "butterscotch"), true)
		_ = os.Chmod(filepath.Join(macosDir, "butterscotch"), 0755)

		// Ensure data.ios / data.win aliases exist alongside game.ios if needed
		gameIosPath := filepath.Join(resourcesDir, "game.ios")
		dataIosPath := filepath.Join(resourcesDir, "data.ios")
		dataWinPath := filepath.Join(resourcesDir, "data.win")
		if _, err := os.Stat(gameIosPath); err == nil {
			if _, errData := os.Stat(dataIosPath); os.IsNotExist(errData) {
				_ = CopyFile(gameIosPath, dataIosPath, true)
			}
			if _, errWin := os.Stat(dataWinPath); os.IsNotExist(errWin) {
				_ = CopyFile(gameIosPath, dataWinPath, true)
			}
		}
	} else {
		// Clean up any previously copied butterscotch runner binaries if mod is not using Butterscotch
		_ = os.Remove(filepath.Join(resourcesDir, "butterscotch"))
		_ = os.Remove(filepath.Join(macosDir, "butterscotch"))
	}

	log(fmt.Sprintf("Successfully applied mod '%s'!", rec.Name))
	return nil
}

// RemoveModAction deletes a mod from disk and DB.
func RemoveModAction(nameOrId string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	var name string
	var id string
	if _, err := strconv.Atoi(nameOrId); err == nil {
		id = nameOrId
	} else {
		name = nameOrId
	}

	rec, err := GetModByNameOrId(name, id)
	if err != nil || rec == nil {
		return fmt.Errorf("mod not found: %w", err)
	}

	modDirPath := filepath.Join(ExpandPath("~/UMMC/mods/"), rec.Name)
	if err := os.RemoveAll(modDirPath); err != nil {
		log(fmt.Sprintf("Warning: Failed to delete mod folder %s: %v", modDirPath, err))
	} else {
		log(fmt.Sprintf("Deleted mod folder from disk: %s", modDirPath))
	}

	if err := DeleteMod(rec.ID); err != nil {
		return fmt.Errorf("error deleting mod from DB: %w", err)
	}

	log(fmt.Sprintf("Successfully removed mod '%s' (ID: %d)!", rec.Name, rec.ID))
	return nil
}

// UpdateModMetadataAction updates a mod's name, maker, base, and options in the database, renames folders, and updates config files.
func UpdateModMetadataAction(id int, newName, newMaker, newBase string, isMacos, useVanillaSaves, installToAppRoot, useButterscotch bool, log LogFunc) (*ModRecord, error) {
	if log == nil {
		log = defaultLog
	}

	if useButterscotch && !IsButterscotchInstalled() {
		return nil, fmt.Errorf("butterscotch runtime is not installed. Please download it first using 'UMMC download-butterscotch' or in the GUI Tools tab")
	}

	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, fmt.Errorf("mod name cannot be empty")
	}
	if newMaker == "" {
		newMaker = "Unknown"
	}
	if newBase == "" {
		newBase = "1.08"
	}

	if isMacos && strings.HasSuffix(newBase, "-w") {
		newBase = strings.TrimSuffix(newBase, "-w")
	} else if !isMacos && !strings.HasSuffix(newBase, "-w") {
		newBase = newBase + "-w"
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var oldName, oldMaker, oldBase string
	err = db.QueryRow("SELECT name, maker, base FROM mods WHERE id = ?", id).Scan(&oldName, &oldMaker, &oldBase)
	if err != nil {
		return nil, fmt.Errorf("mod with ID %d not found in database: %w", id, err)
	}

	// If name changed, rename mod directory and saves directory
	if oldName != newName {
		oldModDir := filepath.Join(ExpandPath("~/UMMC/mods"), oldName)
		newModDir := filepath.Join(ExpandPath("~/UMMC/mods"), newName)

		if _, err := os.Stat(oldModDir); err == nil {
			if _, err := os.Stat(newModDir); err == nil && oldModDir != newModDir {
				return nil, fmt.Errorf("a mod folder named '%s' already exists", newName)
			}
			log(fmt.Sprintf("Renaming mod folder from '%s' to '%s'...", oldName, newName))
			if err := os.Rename(oldModDir, newModDir); err != nil {
				return nil, fmt.Errorf("failed to rename mod folder: %w", err)
			}
		}

		// Rename saves folder if present
		oldSavesDir := filepath.Join(ExpandPath("~/UMMC/saves"), oldName)
		newSavesDir := filepath.Join(ExpandPath("~/UMMC/saves"), newName)
		if _, err := os.Stat(oldSavesDir); err == nil {
			log(fmt.Sprintf("Moving associated saves from '%s' to '%s'...", oldName, newName))
			_ = os.Rename(oldSavesDir, newSavesDir)
		}

		// Update saves table
		_, _ = db.Exec("UPDATE saves SET mod_name = ? WHERE mod_name = ?", newName, oldName)

		// Update active save info if currently active
		if activeInfo, _ := GetActiveSaveInfo(); activeInfo != nil && activeInfo.ModName == oldName {
			_ = WriteSaveInfo(GetUndertaleSaveDir(), activeInfo.Name, newName)
		}
	}

	vanillaSavesInt := 0
	if useVanillaSaves {
		vanillaSavesInt = 1
	}
	rootModeInt := 0
	if installToAppRoot {
		rootModeInt = 1
	}
	butterscotchInt := 0
	if useButterscotch {
		butterscotchInt = 1
	}

	// Update DB record
	_, err = db.Exec("UPDATE mods SET name = ?, maker = ?, base = ?, use_vanilla_saves = ?, install_to_app_root = ?, use_butterscotch = ? WHERE id = ?", newName, newMaker, newBase, vanillaSavesInt, rootModeInt, butterscotchInt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update mod in database: %w", err)
	}

	// Update mod_config.json if present
	modDir := filepath.Join(ExpandPath("~/UMMC/mods"), newName)
	cfgPath := filepath.Join(modDir, "mod_config.json")
	if cfgData, err := os.ReadFile(cfgPath); err == nil {
		var cfg ModConfigFile
		if err := json.Unmarshal(cfgData, &cfg); err == nil {
			cfg.Metadata.Name = newName
			cfg.Metadata.Author = newMaker
			cfg.Metadata.GameVersion = newBase
			cfg.Metadata.UseVanillaSaves = useVanillaSaves
			cfg.Metadata.InstallToAppRoot = installToAppRoot
			cfg.Metadata.UseButterscotch = useButterscotch
			if updatedJson, err := json.MarshalIndent(cfg, "", "  "); err == nil {
				_ = os.WriteFile(cfgPath, updatedJson, 0644)
			}
		}
	}

	log(fmt.Sprintf("Successfully updated mod metadata for '%s' (Maker: %s, Base: %s, Vanilla Saves: %v, Root Mode: %v, Butterscotch: %v)!", newName, newMaker, newBase, useVanillaSaves, installToAppRoot, useButterscotch))
	return &ModRecord{
		ID:               id,
		Name:             newName,
		Maker:            newMaker,
		Base:             newBase,
		UseVanillaSaves:  useVanillaSaves,
		InstallToAppRoot: installToAppRoot,
		UseButterscotch:  useButterscotch,
	}, nil
}

// QuickPatchAction applies an xdelta patch file to Undertale.
func QuickPatchAction(patchFile string, useWinData, force bool, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	patchFile = ExpandPath(patchFile)
	if _, err := os.Stat(patchFile); err != nil {
		return fmt.Errorf("patch file not found: %s", patchFile)
	}

	targetApp := FindUndertaleApp("")
	targetPath := filepath.Join(targetApp, "Contents/Resources/game.ios")

	if useWinData {
		winDataPath := FindWindowsDataPath("")
		if _, err := os.Stat(winDataPath); err != nil {
			return fmt.Errorf("Windows data file not found at %s. Please download Windows version first", winDataPath)
		}
		log(fmt.Sprintf("Copying %s to %s...", winDataPath, targetPath))
		if err := CopyFile(winDataPath, targetPath, true); err != nil {
			return fmt.Errorf("error copying Windows data.win: %w", err)
		}
	}

	log(fmt.Sprintf("Applying patch %s to %s...", filepath.Base(patchFile), targetPath))
	if err := PatchFileForce(targetPath, patchFile, force); err != nil {
		return fmt.Errorf("error patching Undertale: %w", err)
	}

	log("Successfully patched Undertale!")
	return nil
}

// InjectWindowsDataAction copies data.win into Undertale as game.ios.
func InjectWindowsDataAction(customWinData, customTargetApp string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	winDataPath := FindWindowsDataPath(customWinData)
	if _, err := os.Stat(winDataPath); err != nil {
		return fmt.Errorf("Windows data file not found at %s", winDataPath)
	}

	targetApp := FindUndertaleApp(customTargetApp)
	targetGameIos := filepath.Join(targetApp, "Contents/Resources/game.ios")

	log(fmt.Sprintf("Injecting %s into %s...", winDataPath, targetGameIos))
	if err := CopyFile(winDataPath, targetGameIos, true); err != nil {
		return fmt.Errorf("error injecting Windows data.win: %w", err)
	}

	markerPath := filepath.Join(filepath.Dir(targetGameIos), "winpatchdetect")
	_ = os.WriteFile(markerPath, []byte("injected"), 0644)

	log("Successfully injected Windows data.win into macOS Undertale as game.ios!")
	log("Tip: Create a backup now (e.g. '1.08-w') to easily switch back to Windows base.")
	return nil
}

// LaunchUndertale starts Undertale.
func LaunchUndertale(log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	appPath := FindUndertaleApp("")
	if _, err := os.Stat(appPath); err == nil {
		bsInResources := filepath.Join(appPath, "Contents/Resources/butterscotch")
		if fi, errBs := os.Stat(bsInResources); errBs == nil && !fi.IsDir() {
			_ = os.Chmod(bsInResources, 0755)
			log(fmt.Sprintf("Launching Butterscotch runner directly from %s...", bsInResources))
			cmd := exec.Command(bsInResources)
			cmd.Dir = filepath.Join(appPath, "Contents/Resources")
			return cmd.Start()
		}

		log(fmt.Sprintf("Launching Undertale app: %s", appPath))
		cmd := exec.Command("open", "-a", appPath)
		return cmd.Start()
	}

	// Fallback to steam protocol
	log("Undertale app not directly found, attempting steam://run/391540...")
	cmd := exec.Command("open", "steam://run/391540")
	return cmd.Start()
}

// OpenFolder opens a folder in macOS Finder.
func OpenFolder(folderPath string) error {
	exp := ExpandPath(folderPath)
	_ = os.MkdirAll(exp, 0755)
	return exec.Command("open", exp).Start()
}

// DownloadWindowsAction runs steamcmd to download the Windows depot of Undertale.
func DownloadWindowsAction(username string, outputWriter io.Writer, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	targetDir := ExpandPath("~/UMMC/windows/")
	_ = os.MkdirAll(targetDir, 0755)

	status := CheckSystemStatus()
	if !status.SteamcmdAvailable {
		return errors.New("steamcmd is not installed or not found in PATH. Install via 'brew install steamcmd'")
	}

	args := []string{
		"+@sSteamCmdForcePlatformType", "windows",
		"+force_install_dir", targetDir,
	}

	if username != "" {
		args = append(args, "+login", username)
	} else {
		args = append(args, "+login", "anonymous")
	}

	args = append(args, "+app_update", "391540", "validate", "+quit")

	log("Running SteamCMD for Undertale (App ID 391540)...")
	cmd := exec.Command(status.SteamcmdPath, args...)
	if outputWriter != nil {
		cmd.Stdout = outputWriter
		cmd.Stderr = outputWriter
	}

	return cmd.Run()
}

// LaunchSteamCMDInTerminal opens an interactive Terminal.app window to run SteamCMD directly.
// This works anywhere without relying on working directory or binary location (e.g. within a macOS .app).
func LaunchSteamCMDInTerminal(username string) error {
	targetDir := ExpandPath("~/UMMC/windows/")
	_ = os.MkdirAll(targetDir, 0755)

	loginArg := `+login ""`
	if username != "" {
		loginArg = fmt.Sprintf("+login %s", username)
	}

	shellCmd := fmt.Sprintf(
		`export PATH="$PATH:/opt/homebrew/bin:/usr/local/bin"; clear; echo "=========================================="; echo " UMMC: Undertale Windows SteamCMD Downloader"; echo "=========================================="; if ! command -v steamcmd &> /dev/null; then echo "Error: steamcmd is not installed or not found in PATH."; echo "Please install steamcmd (brew install steamcmd) and try again."; echo ""; read -p "Press Enter to exit..."; exit 1; fi; echo "Downloading Undertale Windows depot into ~/UMMC/windows/..."; steamcmd +@sSteamCmdForcePlatformType windows +force_install_dir "%s" %s +app_update 391540 validate +quit; echo ""; echo "=========================================="; echo "Download complete! You can close this window."; echo "=========================================="; read -p "Press Enter to exit..."`,
		targetDir,
		loginArg,
	)

	// Escape double quotes and backslashes for AppleScript string literal
	escapedCmd := strings.ReplaceAll(shellCmd, `\`, `\\`)
	escapedCmd = strings.ReplaceAll(escapedCmd, `"`, `\"`)

	appleScript := fmt.Sprintf(`tell application "Terminal"
		activate
		do script "%s"
	end tell`, escapedCmd)

	cmd := exec.Command("osascript", "-e", appleScript)
	return cmd.Start()
}

// NukeAllDataAction wipes all UMMC data, mods, backups, windows files, database, and optionally deletes game files.
func NukeAllDataAction(deleteGameFiles bool, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	ummcDir := ExpandPath("~/UMMC")
	log(fmt.Sprintf("Nuking UMMC directory: %s...", ummcDir))
	if err := os.RemoveAll(ummcDir); err != nil {
		log(fmt.Sprintf("Warning: Failed to completely remove %s: %v", ummcDir, err))
	} else {
		log("Deleted ~/UMMC (all mods, backups, windows data, and database).")
	}

	if deleteGameFiles {
		gamePath := FindUndertaleApp("")
		if _, err := os.Stat(gamePath); err == nil {
			log(fmt.Sprintf("Deleting Undertale game files at: %s...", gamePath))
			if err := os.RemoveAll(gamePath); err != nil {
				log(fmt.Sprintf("Warning: Failed to delete game app at %s: %v", gamePath, err))
			} else {
				log("Deleted UNDERTALE.app from Steam directory. You can verify integrity or redownload in Steam.")
			}
		}
	}

	// Re-create empty base ~/UMMC directory and fresh DB
	_ = os.MkdirAll(ummcDir, 0755)
	if db, err := GetDB(); err == nil {
		db.Close()
	}

	log("Complete reset finished successfully!")
	return nil
}

// GetDefaultSplashImage searches for the official Undertale splash.png in backups, windows depot, and steam app.
func GetDefaultSplashImage() string {
	candidates := []string{
		"~/UMMC/Backup/undertale1.08/UNDERTALE.app/Contents/Resources/splash.png",
		"~/UMMC/Backup/undertale1.08-w/UNDERTALE.app/Contents/Resources/splash.png",
		"~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app/Contents/Resources/splash.png",
		"~/UMMC/windows/splash.png",
	}

	for _, c := range candidates {
		exp := ExpandPath(c)
		if _, err := os.Stat(exp); err == nil {
			return exp
		}
	}

	// Check any backup record directory
	if backups, err := GetBackups(); err == nil {
		for _, b := range backups {
			if b.BackupPath != "" {
				p := filepath.Join(ExpandPath(b.BackupPath), "Contents/Resources/splash.png")
				if _, err := os.Stat(p); err == nil {
					return p
				}
				p2 := filepath.Join(ExpandPath(b.BackupPath), "splash.png")
				if _, err := os.Stat(p2); err == nil {
					return p2
				}
			}
		}
	}

	return ""
}

// GetModSplashImage searches for a mod's custom splash/banner/cover/icon image, falling back to default Undertale splash.
func GetModSplashImage(modName string) string {
	if modName == "" || strings.EqualFold(modName, "Vanilla") || strings.EqualFold(modName, "Undertale (Original)") {
		return GetDefaultSplashImage()
	}

	modDir := filepath.Join(ExpandPath("~/UMMC/mods"), modName)
	splashCandidates := []string{
		filepath.Join(modDir, "splash.png"),
		filepath.Join(modDir, "splash.jpg"),
		filepath.Join(modDir, "splash.jpeg"),
		filepath.Join(modDir, "banner.png"),
		filepath.Join(modDir, "banner.jpg"),
		filepath.Join(modDir, "cover.png"),
		filepath.Join(modDir, "cover.jpg"),
		filepath.Join(modDir, "icon.png"),
		filepath.Join(modDir, "icon.jpg"),
	}

	for _, cand := range splashCandidates {
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}

	return GetDefaultSplashImage()
}

// GetSaveSplashImage finds the splash image for a save slot based on its associated mod.
func GetSaveSplashImage(saveName, modName string) string {
	// First check if save folder itself has a custom splash or screenshot
	if saveName != "" && modName != "" {
		saveDir := filepath.Join(ExpandPath("~/UMMC/saves"), modName, saveName)
		saveCandidates := []string{
			filepath.Join(saveDir, "splash.png"),
			filepath.Join(saveDir, "screenshot.png"),
			filepath.Join(saveDir, "cover.png"),
			filepath.Join(saveDir, "icon.png"),
		}
		for _, cand := range saveCandidates {
			if _, err := os.Stat(cand); err == nil {
				return cand
			}
		}
	}

	// Fallback to the mod's splash image (which falls back to default Undertale splash)
	return GetModSplashImage(modName)
}

// ResolveButterscotchURL determines the download URL and architecture name based on system or input.
func ResolveButterscotchURL(arch string) (string, string) {
	cleanArch := strings.ToLower(strings.TrimSpace(arch))
	if cleanArch == "" || cleanArch == "auto" {
		if runtime.GOARCH == "arm64" {
			cleanArch = "arm64"
		} else {
			cleanArch = "x86_64"
		}
	}

	if cleanArch == "arm64" || cleanArch == "aarch64" || strings.Contains(cleanArch, "arm") || strings.Contains(cleanArch, "apple") || strings.Contains(cleanArch, "silicon") {
		return ButterscotchArm64URL, "arm64"
	}
	return ButterscotchX86_64URL, "x86_64"
}

// DownloadButterscotchRuntime downloads and extracts the Butterscotch GameMaker runner archive for the specified arch.
func DownloadButterscotchRuntime(arch, customTargetDir, customURL string, log LogFunc) (string, error) {
	if log == nil {
		log = defaultLog
	}

	targetDir := DefaultButterscotchDir
	if customTargetDir != "" {
		targetDir = customTargetDir
	}
	targetDir = ExpandPath(targetDir)

	downloadURL, resolvedArch := ResolveButterscotchURL(arch)
	if customURL != "" {
		downloadURL = customURL
	}

	log(fmt.Sprintf("Preparing Butterscotch runner (%s)...", resolvedArch))
	log(fmt.Sprintf("Download URL: %s", downloadURL))

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}

	urlsToTry := []string{downloadURL}
	if customURL == "" {
		// Fallback candidate URLs in case GitHub workflow path was renamed
		if resolvedArch == "arm64" {
			urlsToTry = append(urlsToTry,
				"https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build-macos/main/butterscotch-appkit-macos-arm64.zip",
				"https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build-macos.yml/main/butterscotch-appkit-macos-arm64.zip",
			)
		} else {
			urlsToTry = append(urlsToTry,
				"https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build-macos/main/butterscotch-appkit-macos-x86_64.zip",
				"https://nightly.link/ButterscotchRunner/Butterscotch/workflows/build-macos.yml/main/butterscotch-appkit-macos-x86_64.zip",
			)
		}
	}

	var resp *http.Response
	var lastErr error
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	for _, tryURL := range urlsToTry {
		req, err := http.NewRequest("GET", tryURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "UMMC-Undertale-Mod-Manager/1.0")

		r, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if r.StatusCode == http.StatusOK {
			resp = r
			downloadURL = tryURL
			break
		}
		r.Body.Close()
		lastErr = fmt.Errorf("HTTP %d: %s", r.StatusCode, r.Status)
	}

	if resp == nil {
		return "", fmt.Errorf("failed to download Butterscotch from %s: %w", downloadURL, lastErr)
	}
	defer resp.Body.Close()

	tmpZip, err := os.CreateTemp("", "butterscotch-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary download file: %w", err)
	}
	defer os.Remove(tmpZip.Name())
	defer tmpZip.Close()

	log("Downloading Butterscotch archive...")
	written, err := io.Copy(tmpZip, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error saving download archive: %w", err)
	}
	log(fmt.Sprintf("Download finished (%d bytes).", written))

	log(fmt.Sprintf("Extracting files to %s...", targetDir))
	if err := extractZipArchive(tmpZip.Name(), targetDir, log); err != nil {
		return "", fmt.Errorf("failed to extract Butterscotch archive: %w", err)
	}

	log(fmt.Sprintf("Successfully installed Butterscotch runtime to %s!", targetDir))
	return targetDir, nil
}

func extractZipArchive(zipPath, destDir string, log LogFunc) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	destDir = filepath.Clean(destDir)

	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		targetPath := filepath.Join(destDir, cleanName)

		// Security check: Zip Slip prevention
		if !strings.HasPrefix(targetPath, destDir+string(filepath.Separator)) && targetPath != destDir {
			return fmt.Errorf("illegal path in zip archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to read zip entry %s: %w", f.Name, err)
		}

		mode := f.Mode()
		if mode.Perm() == 0 {
			mode = 0644
		}
		nameLower := strings.ToLower(f.Name)
		if strings.Contains(nameLower, "macos") || strings.HasSuffix(nameLower, "butterscotch") || strings.HasSuffix(nameLower, "mac_runner") || mode&0111 != 0 {
			mode = 0755
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file %s: %w", targetPath, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", targetPath, err)
		}

		_ = os.Chmod(targetPath, mode)
	}

	return nil
}
