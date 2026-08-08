package cmds

import (
	"UMMC/help"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed windown.sh
var windownScript []byte

var usernameWinCmdArg string

var downloadWinCmdThingy = &cobra.Command{
	Use:   "download-win [username]",
	Short: "Download Undertale for Windows via SteamCMD",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := usernameWinCmdArg
		if username == "" && len(args) > 0 {
			username = args[0]
		}

		scriptPath := "cmds/windown.sh"
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			// Fallback to embedded script in a temporary file
			tmpFile, err := os.CreateTemp("", "windown-*.sh")
			if err != nil {
				fmt.Printf("Error creating temp script file: %v\n", err)
				return
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(windownScript); err != nil {
				fmt.Printf("Error writing script: %v\n", err)
				return
			}
			tmpFile.Close()
			_ = os.Chmod(tmpFile.Name(), 0755)
			scriptPath = tmpFile.Name()
		}

		execCmd := exec.Command("bash", scriptPath)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		env := os.Environ()
		if username != "" {
			env = append(env, fmt.Sprintf("USERNAME=%s", username))
		}
		execCmd.Env = env

		if err := execCmd.Run(); err != nil {
			fmt.Printf("Error executing download script: %v\n", err)
			return
		}
	},
}

var makeWinMacVer = &cobra.Command{
	Use:   "inject [optional source data.win] [optional target game.ios or UNDERTALE.app]",
	Short: "Inject the windows data.win into your macos build of undertale as game.ios",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		winDataPath := help.ExpandPath("~/UMMC/windows/data.win")
		if len(args) > 0 {
			winDataPath = help.ExpandPath(args[0])
		}

		if _, err := os.Stat(winDataPath); os.IsNotExist(err) {
			altPath := help.ExpandPath("~/UMMC/windows/Undertale/data.win")
			if _, errAlt := os.Stat(altPath); errAlt == nil {
				winDataPath = altPath
			} else {
				fmt.Printf("Error: Windows data file not found at %s. Did you run 'UMMC download-win'?\n", winDataPath)
				return
			}
		}

		targetPath := help.ExpandPath("~/Library/Application Support/Steam/steamapps/common/Undertale/UNDERTALE.app/Contents/Resources/game.ios")
		if len(args) > 1 {
			target := help.ExpandPath(args[1])
			if strings.HasSuffix(target, ".ios") {
				targetPath = target
			} else if strings.HasSuffix(target, ".app") {
				targetPath = filepath.Join(target, "Contents/Resources/game.ios")
			} else {
				targetPath = filepath.Join(target, "UNDERTALE.app/Contents/Resources/game.ios")
			}
		}

		fmt.Printf("Injecting %s into %s...\n", winDataPath, targetPath)
		if err := help.CopyFile(winDataPath, targetPath, true); err != nil {
			fmt.Printf("Error injecting Windows data.win: %v\n", err)
			return
		}

		markerPath := filepath.Join(filepath.Dir(targetPath), "winpatchdetect")
		if err := os.WriteFile(markerPath, []byte("injected"), 0644); err != nil {
			fmt.Printf("Warning: Failed to create winpatchdetect marker file at %s: %v\n", markerPath, err)
		}

		fmt.Printf("Successfully injected Windows data.win into macOS Undertale as game.ios!\n")
		fmt.Println("Recommendation: Run 'UMMC backup create' to create a backup of this injected Windows build.")
	},
}

func init() {
	rootCmd.AddCommand(downloadWinCmdThingy)
	rootCmd.AddCommand(makeWinMacVer)
	downloadWinCmdThingy.AddCommand(makeWinMacVer)
	downloadWinCmdThingy.Flags().StringVarP(&usernameWinCmdArg, "username", "u", "", "Steam username to log into SteamCMD")
}

