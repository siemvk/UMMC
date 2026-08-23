package help

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GameConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SteamID        string `json:"steam_id"`
	Path           string `json:"path"`
	GameIOSPath    string `json:"game_ios_path"`
	DataWinPath    string `json:"data_win_path"`
	BackupDir      string `json:"backup_dir"`
	DefaultVersion string `json:"default_version"`
	CoverArt       string `json:"cover_art"`
}

type Config struct {
	Games map[string]GameConfig `json:"games"`
}

// GetConfigPath returns the absolute path to ~/UMMC/config.json
func GetConfigPath() string {
	return ExpandPath("~/UMMC/config.json")
}

// DefaultConfig returns the default configuration for UMMC with separate Deltarune chapter entries.
func DefaultConfig() Config {
	return Config{
		Games: map[string]GameConfig{
			"undertale": {
				ID:             "undertale",
				Name:           "Undertale",
				SteamID:        "391540",
				Path:           "~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app",
				GameIOSPath:    "Contents/Resources/game.ios",
				DataWinPath:    "~/UMMC/windows/data.win",
				BackupDir:      "undertale",
				DefaultVersion: "1.08",
				CoverArt:       "~/UMMC/assets/cover_undertale.png",
			},
			"deltarune-ch1": {
				ID:             "deltarune-ch1",
				Name:           "Deltarune Chapter 1",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/chapter1_mac/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/chapter1_windows/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_ch1.png",
			},
			"deltarune-ch2": {
				ID:             "deltarune-ch2",
				Name:           "Deltarune Chapter 2",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/chapter2_mac/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/chapter2_windows/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_ch2.png",
			},
			"deltarune-ch3": {
				ID:             "deltarune-ch3",
				Name:           "Deltarune Chapter 3",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/chapter3_mac/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/chapter3_windows/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_ch3.png",
			},
			"deltarune-ch4": {
				ID:             "deltarune-ch4",
				Name:           "Deltarune Chapter 4",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/chapter4_mac/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/chapter4_windows/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_ch4.png",
			},
			"deltarune-ch5": {
				ID:             "deltarune-ch5",
				Name:           "Deltarune Chapter 5",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/chapter5_mac/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/chapter5_windows/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_ch5.png",
			},
			"deltarune-mus": {
				ID:             "deltarune-mus",
				Name:           "Deltarune Main Menu / Music",
				SteamID:        "1671210",
				Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
				GameIOSPath:    "Contents/Resources/game.ios",
				DataWinPath:    "~/UMMC/windows/deltarune/data.win",
				BackupDir:      "deltarune",
				DefaultVersion: "vanilla",
				CoverArt:       "~/UMMC/assets/cover_deltarune_mus.png",
			},
		},
	}
}

// ParseGameKey normalizes a game identifier and optional chapter number into a canonical game key (e.g. "deltarune-ch1") and chapter.
func ParseGameKey(gameArg string, chapterArg int) (string, int) {
	game := strings.ToLower(strings.TrimSpace(gameArg))
	chapter := chapterArg

	if game == "deltarune-mus" || game == "mus" || game == "music" || game == "deltarune-music" {
		return "deltarune-mus", 0
	}

	if game == "" {
		if chapter > 0 {
			game = fmt.Sprintf("deltarune-ch%d", chapter)
		} else {
			game = "undertale"
		}
	}

	if strings.HasPrefix(game, "deltarune") {
		remainder := strings.TrimPrefix(game, "deltarune")
		remainder = strings.TrimPrefix(remainder, "-ch")
		remainder = strings.TrimPrefix(remainder, "_ch")
		remainder = strings.TrimPrefix(remainder, "-")
		remainder = strings.TrimPrefix(remainder, ":")
		remainder = strings.TrimPrefix(remainder, "_")
		if ch, err := strconv.Atoi(remainder); err == nil && ch > 0 {
			chapter = ch
		}
		if chapter <= 0 {
			chapter = 1
		}
		game = fmt.Sprintf("deltarune-ch%d", chapter)
	}

	return game, chapter
}

// LoadConfig loads ~/UMMC/config.json. If it does not exist or lacks fields, it creates/updates it with default settings.
func LoadConfig() (*Config, error) {
	cfgPath := GetConfigPath()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := SaveConfig(&cfg); err != nil {
			return &cfg, nil
		}
		return &cfg, nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		cfg := DefaultConfig()
		return &cfg, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfgDefault := DefaultConfig()
		return &cfgDefault, err
	}

	if cfg.Games == nil {
		cfg.Games = make(map[string]GameConfig)
	}

	defaults := DefaultConfig()
	updated := false
	for k, defG := range defaults.Games {
		existing, ok := cfg.Games[k]
		if !ok {
			cfg.Games[k] = defG
			updated = true
		} else {
			if existing.GameIOSPath == "" {
				existing.GameIOSPath = defG.GameIOSPath
				updated = true
			}
			if existing.DataWinPath == "" {
				existing.DataWinPath = defG.DataWinPath
				updated = true
			}
			if existing.BackupDir == "" {
				existing.BackupDir = defG.BackupDir
				updated = true
			}
			if existing.DefaultVersion == "" {
				existing.DefaultVersion = defG.DefaultVersion
				updated = true
			}
			cfg.Games[k] = existing
		}
	}

	if _, ok := cfg.Games["deltarune"]; ok {
		delete(cfg.Games, "deltarune")
		updated = true
	}

	if updated {
		_ = SaveConfig(&cfg)
	}

	return &cfg, nil
}

// SaveConfig writes the given Config struct to ~/UMMC/config.json
func SaveConfig(cfg *Config) error {
	cfgPath := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return err
	}
	_ = os.MkdirAll(ExpandPath("~/UMMC/assets"), 0755)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

// GetGameConfig retrieves GameConfig for a game by ID or key (case-insensitive).
func GetGameConfig(game string) GameConfig {
	cfg, _ := LoadConfig()
	gameKey, _ := ParseGameKey(game, 0)

	if g, ok := cfg.Games[gameKey]; ok {
		return g
	}

	for k, g := range cfg.Games {
		if strings.ToLower(k) == gameKey || strings.ToLower(g.ID) == gameKey || strings.ToLower(g.Name) == gameKey {
			return g
		}
	}

	defaults := DefaultConfig()
	if defG, ok := defaults.Games[gameKey]; ok {
		return defG
	}

	if strings.HasPrefix(gameKey, "deltarune-ch") {
		chNum := strings.TrimPrefix(gameKey, "deltarune-ch")
		return GameConfig{
			ID:             gameKey,
			Name:           fmt.Sprintf("Deltarune Chapter %s", chNum),
			SteamID:        "1671210",
			Path:           "~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app",
			GameIOSPath:    fmt.Sprintf("Contents/Resources/chapter%s_mac/game.ios", chNum),
			DataWinPath:    fmt.Sprintf("~/UMMC/windows/deltarune/chapter%s_windows/data.win", chNum),
			BackupDir:      "deltarune",
			DefaultVersion: "vanilla",
			CoverArt:       fmt.Sprintf("~/UMMC/assets/cover_deltarune_ch%s.png", chNum),
		}
	}

	return GameConfig{
		ID:             gameKey,
		Name:           strings.Title(gameKey),
		SteamID:        "391540",
		Path:           "~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app",
		GameIOSPath:    "Contents/Resources/game.ios",
		DataWinPath:    "~/UMMC/windows/data.win",
		BackupDir:      gameKey,
		DefaultVersion: "1.08",
		CoverArt:       "~/UMMC/assets/cover_undertale.png",
	}
}
