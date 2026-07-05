// Package main 实现 GameAgentEngine 的主命令行入口。
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "GameAgentEngine",
	Aliases: []string{"gameagentengine"},
	Short:   "GameAgentEngine - 游戏 Agent 制作与运行引擎",
	Long:    "GameAgentEngine 是一个面向游戏开发者的专业 AI Agent 制作与运行引擎。",
	Version: version.Version,
}

// init 注册根命令的全局参数。
func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
}

// main 执行根命令并处理退出码。
func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(inspectCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
