package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GameAgentEngine %s (min compatible: %s)\n", version.Version, version.MinCompatibleVersion)
	},
}
