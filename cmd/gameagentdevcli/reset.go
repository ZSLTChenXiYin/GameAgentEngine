package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "清空本地引擎数据库中的全部数据",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initLocalStore(); err != nil {
			fail(err)
		}
		if err := store.ResetAll(); err != nil {
			fail(fmt.Errorf("reset error: %w", err))
		}
		fmt.Println("Local engine database reset completed")
	},
}
