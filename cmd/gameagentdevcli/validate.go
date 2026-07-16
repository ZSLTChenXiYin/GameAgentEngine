package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "校验本地配置与数据库连接",
	Run: func(cmd *cobra.Command, args []string) {
		if localConfigPath == "" {
			fail(fmt.Errorf("--config is required for validate"))
		}
		if err := config.Init(localConfigPath); err != nil {
			fail(fmt.Errorf("config error: %w", err))
		}
		fmt.Println("config OK")
		if err := initLocalStore(); err != nil {
			fail(fmt.Errorf("db error: %w", err))
		}
		defer store.CloseLogSink()
		fmt.Println("db OK")
		fmt.Println("validation passed")
	},
}
