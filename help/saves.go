package help

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// SaveInfoFileName is the metadata filename stored alongside save files.
const SaveInfoFileName = "ummc_save_info.json"

// ActiveSaveInfo represents the metadata stored in ummc_save_info.json.
type ActiveSaveInfo struct {
	Name    string    `json:"name"`
	ModName string    `json:"mod_name"`
	SavedAt time.Time `json:"saved_at,omitempty"`
}

// SaveRecord represents a saved game slot associated with a mod (or "Vanilla").
type SaveRecord struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ModName   string    `json:"mod_name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// GetUndertaleSaveDir returns the active Undertale save directory on macOS.
func GetUndertaleSaveDir() string {
	return ExpandPath("~/Library/Application Support/com.tobyfox.undertale")
}

// WriteSaveInfo writes the metadata JSON file into the specified directory.
func WriteSaveInfo(dir, name, modName string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	info := ActiveSaveInfo{
		Name:    name,
		ModName: modName,
		SavedAt: time.Now(),
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, SaveInfoFileName), data, 0644)
}

// GetActiveSaveInfo reads the metadata of the currently active save if present.
func GetActiveSaveInfo() (*ActiveSaveInfo, error) {
	infoPath := filepath.Join(GetUndertaleSaveDir(), SaveInfoFileName)
	data, err := os.ReadFile(infoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var info ActiveSaveInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// HasActiveSaveFiles checks if the active Undertale save directory contains actual game save files.
func HasActiveSaveFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == ".DS_Store" || name == SaveInfoFileName {
			continue
		}
		return true
	}
	return false
}

// clearDir removes all files and subdirectories within a directory.
func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

// initSavesTable ensures the saves table exists in the database.
func initSavesTable() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := `
	CREATE TABLE IF NOT EXISTS saves (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		mod_name TEXT NOT NULL,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, mod_name)
	);
	`
	_, err = db.Exec(query)
	return err
}

// AutoSaveCurrentGame saves whatever save is currently active before loading a new one.
func AutoSaveCurrentGame(log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	activeDir := GetUndertaleSaveDir()
	if !HasActiveSaveFiles(activeDir) {
		return nil
	}

	info, err := GetActiveSaveInfo()
	if err != nil {
		log(fmt.Sprintf("Warning: Failed to read active save info: %v", err))
	}

	var saveName, modName string
	if info != nil && info.Name != "" {
		saveName = info.Name
		modName = info.ModName
		if modName == "" {
			modName = "Vanilla"
		}
		log(fmt.Sprintf("Auto-saving active game progress to '%s' (mod: %s)...", saveName, modName))
	} else {
		saveName = fmt.Sprintf("autosave_%s", time.Now().Format("20060102_150405"))
		modName = "Vanilla"
		log(fmt.Sprintf("Detected untracked active save files. Auto-saving as '%s' (mod: %s)...", saveName, modName))
	}

	_, err = SaveCurrentGame(saveName, modName, true, log)
	return err
}

// SaveCurrentGame copies the active Undertale save files into a named save slot for a specific mod.
func SaveCurrentGame(saveName, modName string, force bool, log LogFunc) (*SaveRecord, error) {
	if log == nil {
		log = defaultLog
	}

	if saveName == "" {
		return nil, fmt.Errorf("save name cannot be empty")
	}
	if modName == "" {
		modName = "Vanilla"
	}

	srcDir := GetUndertaleSaveDir()
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("active Undertale save directory not found at %s", srcDir)
	}

	// Write save metadata to active dir first so it's always included
	if err := WriteSaveInfo(srcDir, saveName, modName); err != nil {
		log(fmt.Sprintf("Warning: Failed to write save metadata to active directory: %v", err))
	}

	destDir := filepath.Join(ExpandPath("~/UMMC/saves"), modName, saveName)
	log(fmt.Sprintf("Saving current game as '%s' for mod '%s'...", saveName, modName))

	if err := CopyDir(srcDir, destDir, force); err != nil {
		return nil, fmt.Errorf("failed to copy save files: %w", err)
	}

	// Ensure save metadata is also saved in destination
	_ = WriteSaveInfo(destDir, saveName, modName)

	if err := initSavesTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize saves table: %w", err)
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
	INSERT INTO saves (name, mod_name, path) VALUES (?, ?, ?)
	ON CONFLICT(name, mod_name) DO UPDATE SET path = excluded.path, created_at = CURRENT_TIMESTAMP;
	`
	res, err := db.Exec(query, saveName, modName, destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to update saves database: %w", err)
	}

	id, _ := res.LastInsertId()
	log(fmt.Sprintf("Successfully saved '%s' for mod '%s'!", saveName, modName))

	return &SaveRecord{
		ID:        int(id),
		Name:      saveName,
		ModName:   modName,
		Path:      destDir,
		CreatedAt: time.Now(),
	}, nil
}

// LoadSave auto-saves the active game and restores a saved game slot into the active Undertale save directory.
func LoadSave(saveName, modName string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	if modName == "" {
		modName = "Vanilla"
	}

	srcDir := filepath.Join(ExpandPath("~/UMMC/saves"), modName, saveName)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("save '%s' for mod '%s' not found at %s", saveName, modName, srcDir)
	}

	// 1. Auto-save current active game first if any save files exist
	if err := AutoSaveCurrentGame(log); err != nil {
		log(fmt.Sprintf("Warning: Auto-saving current game before load encountered an error: %v", err))
	}

	targetDir := GetUndertaleSaveDir()
	log(fmt.Sprintf("Loading save '%s' (mod: %s) into Undertale...", saveName, modName))

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create Undertale save directory: %w", err)
	}

	// 2. Clear target directory to prevent mixing save files between slots
	if err := clearDir(targetDir); err != nil {
		return fmt.Errorf("failed to clean active save directory: %w", err)
	}

	// 3. Restore save files into active directory
	if err := CopyDir(srcDir, targetDir, true); err != nil {
		return fmt.Errorf("failed to restore save files: %w", err)
	}

	// 4. Ensure active save info is set in active directory
	if err := WriteSaveInfo(targetDir, saveName, modName); err != nil {
		log(fmt.Sprintf("Warning: Failed to update active save metadata: %v", err))
	}

	log(fmt.Sprintf("Successfully loaded save '%s' for mod '%s'!", saveName, modName))
	return nil
}

// CopySaveAndChangeMod creates a duplicate of an existing save and associates it with a new mod.
func CopySaveAndChangeMod(saveName, currentModName, newSaveName, newModName string, log LogFunc) (*SaveRecord, error) {
	if log == nil {
		log = defaultLog
	}

	if currentModName == "" {
		currentModName = "Vanilla"
	}
	if newModName == "" {
		newModName = "Vanilla"
	}
	if newSaveName == "" {
		newSaveName = saveName
	}

	srcDir := filepath.Join(ExpandPath("~/UMMC/saves"), currentModName, saveName)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("source save '%s' for mod '%s' not found at %s", saveName, currentModName, srcDir)
	}

	destDir := filepath.Join(ExpandPath("~/UMMC/saves"), newModName, newSaveName)
	log(fmt.Sprintf("Copying save '%s' (mod: %s) to '%s' (mod: %s)...", saveName, currentModName, newSaveName, newModName))

	if err := CopyDir(srcDir, destDir, true); err != nil {
		return nil, fmt.Errorf("failed to duplicate save directory: %w", err)
	}

	// Update metadata in new copied folder
	if err := WriteSaveInfo(destDir, newSaveName, newModName); err != nil {
		log(fmt.Sprintf("Warning: Failed to write save metadata in copy: %v", err))
	}

	if err := initSavesTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize saves table: %w", err)
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
	INSERT INTO saves (name, mod_name, path) VALUES (?, ?, ?)
	ON CONFLICT(name, mod_name) DO UPDATE SET path = excluded.path, created_at = CURRENT_TIMESTAMP;
	`
	res, err := db.Exec(query, newSaveName, newModName, destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to register copied save in database: %w", err)
	}

	id, _ := res.LastInsertId()
	log(fmt.Sprintf("Successfully copied save to '%s' under mod '%s'!", newSaveName, newModName))

	return &SaveRecord{
		ID:        int(id),
		Name:      newSaveName,
		ModName:   newModName,
		Path:      destDir,
		CreatedAt: time.Now(),
	}, nil
}

// DeleteSave removes a save slot from disk and database.
func DeleteSave(saveName, modName string, log LogFunc) error {
	if log == nil {
		log = defaultLog
	}

	if modName == "" {
		modName = "Vanilla"
	}

	saveDir := filepath.Join(ExpandPath("~/UMMC/saves"), modName, saveName)
	log(fmt.Sprintf("Deleting save '%s' for mod '%s'...", saveName, modName))

	if err := os.RemoveAll(saveDir); err != nil {
		log(fmt.Sprintf("Warning: Failed to delete save directory %s: %v", saveDir, err))
	} else {
		log(fmt.Sprintf("Deleted save folder: %s", saveDir))
	}

	if err := initSavesTable(); err == nil {
		if db, errDB := GetDB(); errDB == nil {
			defer db.Close()
			_, _ = db.Exec("DELETE FROM saves WHERE name = ? AND mod_name = ?", saveName, modName)
		}
	}

	// If active save info corresponds to this deleted save, remove the metadata file
	if info, err := GetActiveSaveInfo(); err == nil && info != nil {
		if info.Name == saveName && info.ModName == modName {
			_ = os.Remove(filepath.Join(GetUndertaleSaveDir(), SaveInfoFileName))
		}
	}

	log(fmt.Sprintf("Successfully removed save '%s' (mod: %s)!", saveName, modName))
	return nil
}

// GetSaves returns all save records, optionally filtered by mod name (pass "" for all).
func GetSaves(filterMod string) ([]SaveRecord, error) {
	if err := initSavesTable(); err != nil {
		return nil, err
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	var rows *sql.Rows
	var errQuery error

	if filterMod != "" {
		query = `SELECT id, name, mod_name, path, created_at FROM saves WHERE mod_name = ? ORDER BY created_at DESC`
		rows, errQuery = db.Query(query, filterMod)
	} else {
		query = `SELECT id, name, mod_name, path, created_at FROM saves ORDER BY mod_name ASC, created_at DESC`
		rows, errQuery = db.Query(query)
	}

	if errQuery != nil {
		return nil, errQuery
	}
	defer rows.Close()

	var records []SaveRecord
	for rows.Next() {
		var rec SaveRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.ModName, &rec.Path, &rec.CreatedAt); err != nil {
			continue
		}
		records = append(records, rec)
	}

	return records, nil
}

// GetSaveByID returns a save record by its database ID.
func GetSaveByID(id int) (*SaveRecord, error) {
	if err := initSavesTable(); err != nil {
		return nil, err
	}
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var rec SaveRecord
	err = db.QueryRow("SELECT id, name, mod_name, path, created_at FROM saves WHERE id = ?", id).
		Scan(&rec.ID, &rec.Name, &rec.ModName, &rec.Path, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetSaveByNameAndMod returns a save record by name and mod name.
func GetSaveByNameAndMod(name, modName string) (*SaveRecord, error) {
	if err := initSavesTable(); err != nil {
		return nil, err
	}
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if modName == "" {
		modName = "Vanilla"
	}

	var rec SaveRecord
	err = db.QueryRow("SELECT id, name, mod_name, path, created_at FROM saves WHERE name = ? AND mod_name = ?", name, modName).
		Scan(&rec.ID, &rec.Name, &rec.ModName, &rec.Path, &rec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetSaveByNameOrID looks up a save by ID or by name + optional mod filter.
func GetSaveByNameOrID(nameOrID, modName string) (*SaveRecord, error) {
	if modName == "" {
		modName = "Vanilla"
	}
	// Try parsing as ID first
	if id, err := strconv.Atoi(nameOrID); err == nil {
		if rec, errRec := GetSaveByID(id); errRec == nil && rec != nil {
			return rec, nil
		}
	}
	return GetSaveByNameAndMod(nameOrID, modName)
}

// CreateEmptySave initializes a fresh, empty save slot for a mod.
func CreateEmptySave(saveName, modName string, log LogFunc) (*SaveRecord, error) {
	if log == nil {
		log = defaultLog
	}

	if saveName == "" {
		return nil, fmt.Errorf("save name cannot be empty")
	}
	if modName == "" {
		modName = "Vanilla"
	}

	destDir := filepath.Join(ExpandPath("~/UMMC/saves"), modName, saveName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}

	if err := WriteSaveInfo(destDir, saveName, modName); err != nil {
		return nil, fmt.Errorf("failed to write save metadata: %w", err)
	}

	if err := initSavesTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize saves table: %w", err)
	}

	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
	INSERT INTO saves (name, mod_name, path) VALUES (?, ?, ?)
	ON CONFLICT(name, mod_name) DO UPDATE SET path = excluded.path, created_at = CURRENT_TIMESTAMP;
	`
	res, err := db.Exec(query, saveName, modName, destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to update saves database: %w", err)
	}

	id, _ := res.LastInsertId()
	log(fmt.Sprintf("Initialized fresh save slot '%s' for mod '%s'.", saveName, modName))

	return &SaveRecord{
		ID:        int(id),
		Name:      saveName,
		ModName:   modName,
		Path:      destDir,
		CreatedAt: time.Now(),
	}, nil
}

// GetMostRecentSave returns the most recent save record for a specific mod, or nil if none found.
func GetMostRecentSave(modName string) (*SaveRecord, error) {
	if err := initSavesTable(); err != nil {
		return nil, err
	}
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if modName == "" {
		modName = "Vanilla"
	}

	var rec SaveRecord
	err = db.QueryRow("SELECT id, name, mod_name, path, created_at FROM saves WHERE mod_name = ? ORDER BY created_at DESC LIMIT 1", modName).
		Scan(&rec.ID, &rec.Name, &rec.ModName, &rec.Path, &rec.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}


