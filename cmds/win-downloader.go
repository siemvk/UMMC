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
	Short: "Download Undertale or Deltarune for Windows via SteamCMD",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := usernameWinCmdArg
		if username == "" && len(args) > 0 {
			username = args[0]
		}

		game := "undertale"
		if cmd.Flags().Changed("undertale") || UndertaleCmdArg {
			game = "undertale"
		} else if cmd.Flags().Changed("deltarune") || DeltaruneCmdArg > 0 {
			game = "deltarune"
		}

		appID := "391540"
		targetDir := help.ExpandPath("~/UMMC/windows/")
		gameName := "Undertale"
		if game == "deltarune" {
			appID = "1671210"
			targetDir = help.ExpandPath("~/UMMC/windows/deltarune/")
			gameName = "Deltarune"
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
		env = append(env, fmt.Sprintf("APP_ID=%s", appID))
		env = append(env, fmt.Sprintf("TARGET_DIR=%s", targetDir))
		env = append(env, fmt.Sprintf("GAME_NAME=%s", gameName))
		execCmd.Env = env

		if err := execCmd.Run(); err != nil {
			fmt.Printf("Error executing download script: %v\n", err)
			return
		}
	},
}

var makeWinMacVer = &cobra.Command{
	Use:   "inject [optional source data.win] [optional target game.ios or app]",
	Short: "Inject the Windows data.win into your macOS build of Undertale or Deltarune as game.ios",
	Args:  cobra.MaximumNArgs(2),
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

		var winDataPath string
		if len(args) > 0 {
			winDataPath = help.ExpandPath(args[0])
		} else {
			winDataPath = help.FindWindowsDataWin(game, chapter)
		}

		if _, err := os.Stat(winDataPath); os.IsNotExist(err) {
			fmt.Printf("Error: Windows data file not found at %s. Did you run 'UMMC download-win'?\n", winDataPath)
			return
		}


		defaultApp := help.GetDefaultAppPath(game)
		targetPath := help.GetGameDataPath(defaultApp, game, chapter)

		if len(args) > 1 {
			target := help.ExpandPath(args[1])
			if strings.HasSuffix(target, ".ios") {
				targetPath = target
			} else if strings.HasSuffix(target, ".app") {
				targetPath = help.GetGameDataPath(target, game, chapter)
			} else {
				appName := "UNDERTALE.app"
				if game == "deltarune" {
					appName = "DELTARUNE.app"
				}
				targetPath = help.GetGameDataPath(filepath.Join(target, appName), game, chapter)
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

		if game == "deltarune" {
			fmt.Printf("Successfully injected Windows data.win into macOS Deltarune Chapter %d as game.ios!\n", chapter)
			fmt.Printf("Recommendation: Run 'UMMC backup create -d %d' to create a backup of this injected build.\n", chapter)
		} else {
			fmt.Printf("Successfully injected Windows data.win into macOS Undertale as game.ios!\n")
			fmt.Println("Recommendation: Run 'UMMC backup create' to create a backup of this injected Windows build.")
		}
	},
}


func init() {
	rootCmd.AddCommand(downloadWinCmdThingy)
	rootCmd.AddCommand(makeWinMacVer)
	downloadWinCmdThingy.AddCommand(makeWinMacVer)
	downloadWinCmdThingy.Flags().StringVarP(&usernameWinCmdArg, "username", "U", "", "Steam username to log into SteamCMD")
}


