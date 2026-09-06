package cmds

import (
	"UMMC/help"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var butterscotchArchCmdArg string
var butterscotchDirCmdArg string
var butterscotchUrlCmdArg string

var downloadButterscotchCmdThingy = &cobra.Command{
	Use:     "download-butterscotch [optional arch: arm64|x86_64]",
	Aliases: []string{"butterscotch", "download-runner", "runner"},
	Short:   "Download the experimental Butterscotch GameMaker runner for macOS",
	Long: `Downloads and extracts the experimental Butterscotch native GameMaker runner for macOS.
Supports Apple Silicon (arm64) and Intel (x86_64) builds from nightly.link.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		arch := butterscotchArchCmdArg
		if arch == "" && len(args) > 0 {
			arch = args[0]
		}
		if arch == "" {
			arch = "auto"
		}

		targetDir := butterscotchDirCmdArg
		if targetDir == "" {
			targetDir = help.DefaultButterscotchDir
		}

		detectedArch := "x86_64"
		if runtime.GOARCH == "arm64" {
			detectedArch = "arm64"
		}

		fmt.Println("=== Butterscotch Runner Downloader ===")
		fmt.Printf("System Arch: %s | Requested: %s\n", detectedArch, arch)

		dest, err := help.DownloadButterscotchRuntime(arch, targetDir, butterscotchUrlCmdArg, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Butterscotch runner installed at: %s\n", dest)
		fmt.Println("You can use this runtime for native macOS mods or custom GameMaker runner mods.")
	},
}

func init() {
	rootCmd.AddCommand(downloadButterscotchCmdThingy)

	downloadButterscotchCmdThingy.Flags().StringVarP(&butterscotchArchCmdArg, "arch", "a", "auto", "Architecture to download (arm64, x86_64, auto)")
	downloadButterscotchCmdThingy.Flags().StringVarP(&butterscotchDirCmdArg, "dir", "d", "", "Target extraction directory (default: ~/UMMC/runtimes/butterscotch)")
	downloadButterscotchCmdThingy.Flags().StringVar(&butterscotchUrlCmdArg, "url", "", "Custom direct URL to download zip from")
}
