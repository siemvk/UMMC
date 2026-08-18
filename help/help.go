package help

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath converts a leading "~" into the user's home directory.
func ExpandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return filepath.Clean(path)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path // Fallback to raw input if home dir can't be read
	}

	if path == "~" {
		return home
	}

	if !strings.HasPrefix(path, "~/") {
		return filepath.Clean(path) // Ignore strings like "~filename"
	}

	return filepath.Join(home, path[2:])
}

// CopyFile copies a file or directory from srcPath to dstPath.
func CopyFile(srcPath, dstPath string, force bool) error {
	srcPath = ExpandPath(srcPath)
	dstPath = ExpandPath(dstPath)

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", srcPath, err)
	}

	if info.IsDir() {
		return CopyDir(srcPath, dstPath, force)
	}

	if dstInfo, err := os.Stat(dstPath); err == nil {
		if dstInfo.IsDir() {
			dstPath = filepath.Join(dstPath, filepath.Base(srcPath))
			if _, err := os.Stat(dstPath); err == nil && !force {
				return fmt.Errorf("safe mode active: file already exists at %s", dstPath)
			}
		} else if !force {
			return fmt.Errorf("safe mode active: file already exists at %s", dstPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat destination %s: %w", dstPath, err)
	}

	// Handle case where srcPath is a symlink to a file
	lInfo, err := os.Lstat(srcPath)
	if err == nil && lInfo.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read symlink %s: %w", srcPath, err)
		}
		_ = os.MkdirAll(filepath.Dir(dstPath), 0755)
		_ = os.Remove(dstPath)
		return os.Symlink(linkTarget, dstPath)
	}

	return copyFileContents(srcPath, dstPath, info.Mode())
}

// CopyDir recursively copies a directory tree from srcDir to dstDir.
func CopyDir(srcDir, dstDir string, force bool) error {
	srcDir = ExpandPath(srcDir)
	dstDir = ExpandPath(dstDir)

	absSrc, errSrc := filepath.Abs(srcDir)
	absDst, errDst := filepath.Abs(dstDir)
	if errSrc == nil && errDst == nil {
		if absSrc == absDst {
			return fmt.Errorf("source and destination are the same directory: %s", srcDir)
		}
		if strings.HasPrefix(absSrc, absDst+string(filepath.Separator)) {
			return fmt.Errorf("cannot copy parent directory %s into subdirectory %s", srcDir, dstDir)
		}
	}

	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", srcDir, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source %s is not a directory", srcDir)
	}

	if _, err := os.Stat(dstDir); err == nil {
		if !force {
			return fmt.Errorf("safe mode active: destination already exists at %s", dstDir)
		}
		if err := os.RemoveAll(dstDir); err != nil {
			return fmt.Errorf("failed to remove existing destination at %s: %w", dstDir, err)
		}
	}

	if err := os.MkdirAll(dstDir, srcInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dstDir, err)
	}

	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		absPath, errAbs := filepath.Abs(path)
		if errAbs == nil && errDst == nil {
			if absPath == absDst || strings.HasPrefix(absPath, absDst+string(filepath.Separator)) {
				return filepath.SkipDir
			}
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(dstDir, relPath)
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for %s: %w", path, err)
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", path, err)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for symlink at %s: %w", targetPath, err)
			}
			_ = os.Remove(targetPath)
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink at %s: %w", targetPath, err)
			}
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}

		return copyFileContents(path, targetPath, info.Mode())
	})
}

func copyFileContents(srcPath, dstPath string, mode os.FileMode) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", dstPath, err)
	}

	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data to %s: %w", dstPath, err)
	}

	if err = dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", dstPath, err)
	}

	return nil
}

// CopyOverlay copies a file or recursively overlays a directory tree from srcPath to dstPath.
// Unlike CopyDir, if dstPath is an existing directory, CopyOverlay merges files into it rather than removing dstPath.
func CopyOverlay(srcPath, dstPath string, force bool) error {
	srcPath = ExpandPath(srcPath)
	dstPath = ExpandPath(dstPath)

	info, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", srcPath, err)
	}

	if !info.IsDir() {
		return CopyFile(srcPath, dstPath, force)
	}

	absSrc, errSrc := filepath.Abs(srcPath)
	absDst, errDst := filepath.Abs(dstPath)
	if errSrc == nil && errDst == nil {
		if absSrc == absDst {
			return fmt.Errorf("source and destination are the same directory: %s", srcPath)
		}
		if strings.HasPrefix(absSrc, absDst+string(filepath.Separator)) {
			return fmt.Errorf("cannot copy parent directory %s into subdirectory %s", srcPath, dstPath)
		}
	}

	if dstInfo, err := os.Stat(dstPath); err == nil {
		if !dstInfo.IsDir() {
			if !force {
				return fmt.Errorf("safe mode active: destination %s is a file", dstPath)
			}
			if err := os.Remove(dstPath); err != nil {
				return fmt.Errorf("failed to remove existing destination file at %s: %w", dstPath, err)
			}
			if err := os.MkdirAll(dstPath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", dstPath, err)
			}
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(dstPath, info.Mode().Perm()); err != nil {
			return fmt.Errorf("failed to create destination directory %s: %w", dstPath, err)
		}
	} else {
		return fmt.Errorf("failed to stat destination %s: %w", dstPath, err)
	}

	return filepath.WalkDir(srcPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		absPath, errAbs := filepath.Abs(path)
		if errAbs == nil && errDst == nil {
			if absPath == absDst || strings.HasPrefix(absPath, absDst+string(filepath.Separator)) {
				return filepath.SkipDir
			}
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(dstPath, relPath)
		entryInfo, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for %s: %w", path, err)
		}

		if entryInfo.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", path, err)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for symlink at %s: %w", targetPath, err)
			}
			_ = os.Remove(targetPath)
			if err := os.Symlink(linkTarget, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink at %s: %w", targetPath, err)
			}
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(targetPath, entryInfo.Mode().Perm())
		}

		return copyFileContents(path, targetPath, entryInfo.Mode())
	})
}

// GetDefaultAppPath returns the default macOS Steam app path for Undertale or Deltarune.
func GetDefaultAppPath(game string) string {
	if strings.ToLower(game) == "deltarune" {
		return ExpandPath("~/Library/Application Support/Steam/steamapps/common/DELTARUNE/DELTARUNE.app")
	}
	return ExpandPath("~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app")
}

// GetDeltaruneChapterDir returns the path to chapter directory inside DELTARUNE.app (e.g. chapter1_mac).
func GetDeltaruneChapterDir(appPath string, chapter int) string {
	if chapter <= 0 {
		chapter = 1
	}
	resDir := filepath.Join(appPath, "Contents/Resources")
	candidate1 := filepath.Join(resDir, fmt.Sprintf("chapter%d_mac", chapter))
	if _, err := os.Stat(candidate1); err == nil {
		return candidate1
	}
	candidate2 := filepath.Join(resDir, fmt.Sprintf("chapter%d", chapter))
	if _, err := os.Stat(candidate2); err == nil {
		return candidate2
	}
	return candidate1
}

// GetGameDataPath returns game.ios path for Undertale or Deltarune (chapter specific).
func GetGameDataPath(appPath string, game string, chapter int) string {
	if strings.ToLower(game) == "deltarune" {
		return filepath.Join(GetDeltaruneChapterDir(appPath, chapter), "game.ios")
	}
	return filepath.Join(appPath, "Contents/Resources/game.ios")
}

// GetGameResourceDir returns the target resource directory for mods (chapter folder for Deltarune, Contents/Resources for Undertale).
func GetGameResourceDir(appPath string, game string, chapter int) string {
	if strings.ToLower(game) == "deltarune" {
		return GetDeltaruneChapterDir(appPath, chapter)
	}
	return filepath.Join(appPath, "Contents/Resources")
}

// FindWindowsDataWin finds the Windows data.win file for Undertale or Deltarune (chapter specific).
func FindWindowsDataWin(game string, chapter int) string {
	if strings.ToLower(game) == "deltarune" {
		if chapter <= 0 {
			chapter = 1
		}
		candidates := []string{
			ExpandPath(fmt.Sprintf("~/UMMC/windows/deltarune/chapter%d_windows/data.win", chapter)),
			ExpandPath(fmt.Sprintf("~/UMMC/windows/deltarune/chapter%d_win/data.win", chapter)),
			ExpandPath(fmt.Sprintf("~/UMMC/windows/deltarune/chapter%d/data.win", chapter)),
			ExpandPath(fmt.Sprintf("~/UMMC/windows/deltarune/ch%d/data.win", chapter)),
			ExpandPath("~/UMMC/windows/deltarune/data.win"),
			ExpandPath("~/UMMC/windows/data.win"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		return candidates[0]
	}

	candidates := []string{
		ExpandPath("~/UMMC/windows/data.win"),
		ExpandPath("~/UMMC/windows/Undertale/data.win"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

// IsWinPatched checks if the given app path or its chapter directory contains a winpatchdetect marker.
func IsWinPatched(appPath string, game string, chapter int) bool {
	appPath = ExpandPath(appPath)

	if strings.ToLower(game) == "deltarune" {
		if chapter <= 0 {
			chapter = 1
		}
		chDir := GetDeltaruneChapterDir(appPath, chapter)
		if _, err := os.Stat(filepath.Join(chDir, "winpatchdetect")); err == nil {
			return true
		}
		return false
	}

	standardPaths := []string{
		filepath.Join(appPath, "Contents/Resources/winpatchdetect"),
		filepath.Join(appPath, "winpatchdetect"),
		filepath.Join(filepath.Dir(appPath), "winpatchdetect"),
	}

	for _, p := range standardPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}

	return false
}
