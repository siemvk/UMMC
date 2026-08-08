package help

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PatchFile applies an xdelta patch to a file at targetPath using patchPath and overwrites targetPath.
func PatchFile(targetPath, patchPath string) error {
	return PatchFileForce(targetPath, patchPath, false)
}

// PatchFileForce applies an xdelta patch to targetPath using patchPath and overwrites targetPath.
// If force is true, passes -f and -n to xdelta3 to disable checksum validation and force patching.
func PatchFileForce(targetPath, patchPath string, force bool) error {
	expandedTarget := ExpandPath(targetPath)
	expandedPatch := ExpandPath(patchPath)

	targetInfo, err := os.Stat(expandedTarget)
	if err != nil {
		return fmt.Errorf("target file not found at %s: %w", expandedTarget, err)
	}

	if _, err := os.Stat(expandedPatch); err != nil {
		return fmt.Errorf("patch file not found at %s: %w", expandedPatch, err)
	}

	// Verify xdelta3 binary availability
	xdeltaBin, err := exec.LookPath("xdelta3")
	if err != nil {
		if _, statErr := os.Stat("/opt/homebrew/bin/xdelta3"); statErr == nil {
			xdeltaBin = "/opt/homebrew/bin/xdelta3"
		} else if _, statErr := os.Stat("/usr/local/bin/xdelta3"); statErr == nil {
			xdeltaBin = "/usr/local/bin/xdelta3"
		} else {
			return fmt.Errorf("xdelta3 executable not found in PATH. Please install xdelta3 (e.g. 'brew install xdelta'): %w", err)
		}
	}

	// Create temporary file path in the same directory for patched output
	targetDir := filepath.Dir(expandedTarget)
	tempFile, err := os.CreateTemp(targetDir, "patched_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary output file in %s: %w", targetDir, err)
	}
	tempPath := tempFile.Name()
	tempFile.Close() // Close so xdelta3 CLI can write to it

	// Ensure temp file is cleaned up on failure
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			os.Remove(tempPath)
		}
	}()

	// Build xdelta3 decode command args
	args := []string{"-d"}
	if force {
		// -f forces overwrite, -n disables xdelta3 source/target checksum checks
		args = append(args, "-f", "-n")
	} else {
		args = append(args, "-f")
	}
	args = append(args, "-s", expandedTarget, expandedPatch, tempPath)

	// Execute xdelta3 decode command
	cmd := exec.Command(xdeltaBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xdelta3 patch failed: %w (output: %s)", err, string(output))
	}

	// Verify temp file size
	tempInfo, err := os.Stat(tempPath)
	if err != nil {
		return fmt.Errorf("failed to stat temporary patched file: %w", err)
	}
	if tempInfo.Size() == 0 {
		return fmt.Errorf("patched file is empty")
	}

	// Restore original file permissions to temporary file
	_ = os.Chmod(tempPath, targetInfo.Mode())

	// Overwrite original target file
	cleanupTemp = false
	if err := os.Rename(tempPath, expandedTarget); err != nil {
		if copyErr := CopyFile(tempPath, expandedTarget, true); copyErr != nil {
			os.Remove(tempPath)
			return fmt.Errorf("failed to overwrite target file %s: %w", expandedTarget, copyErr)
		}
		os.Remove(tempPath)
	}

	return nil
}
