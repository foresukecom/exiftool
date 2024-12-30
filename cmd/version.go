package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"  // デフォルト値
	commit  = "none" // デフォルト値
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "バージョン情報の表示",
	Long:  "このツールのバージョンとビルド情報を表示します。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("バージョン: %s\n", version)
		fmt.Printf("コミットハッシュ: %s\n", commit)
		fmt.Printf("ビルド日付: %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
