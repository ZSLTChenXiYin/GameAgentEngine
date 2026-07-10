package main

import (
	"log"
	"time"

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
		store.ConfigureLogSink(store.LogSinkOptions{
			Enabled:       config.Global.Database.LogBatchEnabled,
			BatchSize:     config.Global.Database.LogBatchSize,
			FlushInterval: time.Duration(config.Global.Database.LogBatchFlushMs) * time.Millisecond,
			QueueSize:     config.Global.Database.LogBatchQueueSize,
		})
		log.Print("config OK")
		if err := store.Init(config.Global.Database.Driver, config.Global.Database.DSN); err != nil {
			log.Fatalf("db error: %v", err)
		}
		defer store.CloseLogSink()
		log.Print("db OK")
		log.Print("validation passed")
	},
}
