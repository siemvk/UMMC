package cmds

import (
	"UMMC/help"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage UMMC game configurations",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := help.GetConfigPath()
		cfg, err := help.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config at %s: %v\n", cfgPath, err)
			return
		}

		fmt.Printf("=== Config File Path ===\n%s\n\n", cfgPath)
		fmt.Println("=== Configured Games ===")

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting config JSON: %v\n", err)
			return
		}
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
