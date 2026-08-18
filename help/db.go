package help

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// GetDB opens or creates the SQLite database at ~/UMMC/ummc.db and initializes tables.
func GetDB() (*sql.DB, error) {
	dbPath := ExpandPath("~/UMMC/ummc.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT NOT NULL,
		app_path TEXT NOT NULL,
		backup_path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS mods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		base TEXT NOT NULL,
		name TEXT NOT NULL,
		maker TEXT NOT NULL,
		install_to_app_root INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS saves (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		
	);
	`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	_, _ = db.Exec("ALTER TABLE mods ADD COLUMN install_to_app_root INTEGER DEFAULT 0;")
	return nil
}

type BackupRecord struct {
	ID         int       `json:"id"`
	Version    string    `json:"version"`
	AppPath    string    `json:"app_path"`
	BackupPath string    `json:"backup_path"`
	CreatedAt  time.Time `json:"created_at"`
}

// GetBackupByVersionOrId queries the database for a backup by ID if provided, otherwise by version.
func GetBackupByVersionOrId(version string, id string) (*BackupRecord, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	var arg interface{}

	if id != "" {
		query = `SELECT id, version, app_path, backup_path, created_at FROM backups WHERE id = ? LIMIT 1`
		arg = id
	} else if version != "" {
		query = `SELECT id, version, app_path, backup_path, created_at FROM backups WHERE version = ? ORDER BY id DESC LIMIT 1`
		arg = version
	} else {
		return nil, fmt.Errorf("neither id nor version provided")
	}

	row := db.QueryRow(query, arg)
	var rec BackupRecord
	if err := row.Scan(&rec.ID, &rec.Version, &rec.AppPath, &rec.BackupPath, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetBackups returns all backup records ordered by newest first.
func GetBackups() ([]BackupRecord, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, version, app_path, backup_path, created_at FROM backups ORDER BY id DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []BackupRecord
	for rows.Next() {
		var rec BackupRecord
		if err := rows.Scan(&rec.ID, &rec.Version, &rec.AppPath, &rec.BackupPath, &rec.CreatedAt); err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

type ModRecord struct {
	ID               int    `json:"id"`
	Base             string `json:"base"`
	Name             string `json:"name"`
	Maker            string `json:"maker"`
	InstallToAppRoot bool   `json:"install_to_app_root"`
}

// AddMod inserts a new mod record into the database and returns the generated ID.
func AddMod(base, name, maker string, installToAppRoot bool) (int64, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	installToAppRootInt := 0
	if installToAppRoot {
		installToAppRootInt = 1
	}

	res, err := db.Exec("INSERT INTO mods (base, name, maker, install_to_app_root) VALUES (?, ?, ?, ?)", base, name, maker, installToAppRootInt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert mod into database: %w", err)
	}

	return res.LastInsertId()
}

// GetMods returns all mod records ordered by newest first.
func GetMods() ([]ModRecord, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `SELECT id, base, name, maker, install_to_app_root FROM mods ORDER BY id DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ModRecord
	for rows.Next() {
		var rec ModRecord
		var installToAppRootInt int
		if err := rows.Scan(&rec.ID, &rec.Base, &rec.Name, &rec.Maker, &installToAppRootInt); err != nil {
			continue
		}
		rec.InstallToAppRoot = (installToAppRootInt != 0)
		records = append(records, rec)
	}
	return records, nil
}

// DeleteBackup removes a backup record by ID from the database.
func DeleteBackup(id int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM backups WHERE id = ?", id)
	return err
}

// GetModByNameOrId queries the database for a mod by ID if provided, otherwise by name.
func GetModByNameOrId(name string, id string) (*ModRecord, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var query string
	var arg interface{}

	if id != "" {
		query = `SELECT id, base, name, maker, install_to_app_root FROM mods WHERE id = ? LIMIT 1`
		arg = id
	} else if name != "" {
		query = `SELECT id, base, name, maker, install_to_app_root FROM mods WHERE name = ? ORDER BY id DESC LIMIT 1`
		arg = name
	} else {
		return nil, fmt.Errorf("neither id nor name provided")
	}

	row := db.QueryRow(query, arg)
	var rec ModRecord
	var installToAppRootInt int
	if err := row.Scan(&rec.ID, &rec.Base, &rec.Name, &rec.Maker, &installToAppRootInt); err != nil {
		return nil, err
	}
	rec.InstallToAppRoot = (installToAppRootInt != 0)
	return &rec, nil
}

// DeleteMod removes a mod record by ID from the database.
func DeleteMod(id int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM mods WHERE id = ?", id)
	return err
}
