package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/agent"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/api"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/llm"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Init(cfgFile); err != nil {
			log.Fatalf("config: %v", err)
		}
		store.ConfigureLogSink(store.LogSinkOptions{
			Enabled:       config.Global.Database.LogBatchEnabled,
			BatchSize:     config.Global.Database.LogBatchSize,
			FlushInterval: time.Duration(config.Global.Database.LogBatchFlushMs) * time.Millisecond,
			QueueSize:     config.Global.Database.LogBatchQueueSize,
		})
		store.ConfigureWriteRetry(store.WriteRetryOptions{
			Enabled:     config.Global.Database.WriteRetryEnabled,
			MaxAttempts: config.Global.Database.WriteRetryMaxAttempts,
			BaseDelay:   time.Duration(config.Global.Database.WriteRetryBaseDelayMs) * time.Millisecond,
			MaxDelay:    time.Duration(config.Global.Database.WriteRetryMaxDelayMs) * time.Millisecond,
		})
		store.ConfigureMigrationsEnabled(config.Global.Database.MigrationsEnabled)
		service.ConfigureWorldLocks(config.Global.Engine.WorldLockEnabled)
		if err := store.Init(config.Global.Database.Driver, config.Global.Database.DSN); err != nil {
			log.Fatalf("db: %v", err)
		}
		defer func() {
			if err := store.CloseLogSink(); err != nil {
				log.Printf("close log sink: %v", err)
			}
		}()
		// Debug 模式下自动启用详细日志
		if config.ExecutionMode() == "debug" {
			log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
			log.Printf("[debug] execution_mode=debug, verbose logging enabled")
		}

		log.Printf("DB: %s (%s)", config.Global.Database.Driver, config.Global.Database.DSN)

		var provider engine.LLMProvider
		if config.Global.LLM.APIKey != "" {
			provider = llm.NewOpenAIProvider(
				config.Global.LLM.APIKey,
				config.Global.LLM.BaseURL,
				config.Global.LLM.Model,
			)
			log.Printf("LLM: %s (%s)", config.Global.LLM.Model, config.Global.LLM.BaseURL)
		} else {
			log.Print("LLM: no API key configured, using mock provider")
			provider = llm.NewMockProvider()
		}

		pipeline := engine.NewPipeline(provider)
		schedulerCtx, stopScheduler := context.WithCancel(context.Background())
		defer stopScheduler()
		if config.Global.Engine.AutonomousSchedulerEnabled {
			interval := time.Duration(config.Global.Engine.AutonomousSchedulerIntervalSeconds) * time.Second
			limit := config.Global.Engine.AutonomousSchedulerMaxNodesPerWorld
			log.Printf("autonomous scheduler enabled: interval=%s limit=%d", interval, limit)
			go agent.NewScheduler(pipeline, interval, limit).Start(schedulerCtx)
		}
		mux := api.NewRouter(pipeline)

		var handler http.Handler = mux
		handler = api.RequestAuth(handler, config.Global.Auth)
		handler = api.CORSMiddleware(handler)

		addr := fmt.Sprintf("%s:%d", config.Global.Server.Host, config.Global.Server.Port)
		log.Printf("listen on %s", addr)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		srv := &http.Server{Addr: addr, Handler: handler}
		go func() {
			<-stop
			log.Print("shutting down...")
			stopScheduler()
			srv.Close()
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	},
}
