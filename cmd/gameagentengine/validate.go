package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "校验配置",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Init(cfgFile); err != nil {
			log.Fatalf("config error: %v", err)
		}
		log.Print("config OK")
		if err := store.Init(config.Global.Database.Driver, config.Global.Database.DSN); err != nil {
			log.Fatalf("db error: %v", err)
		}
		log.Print("db OK")
		log.Print("validation passed")
	},
}
