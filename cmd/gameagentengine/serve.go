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

		// Restore pending world change plans from the previous server run
		if err := engine.GlobalPlanReview.RestorePendingPlansFromDB(); err != nil {
			log.Printf("[warn] restore pending plans: %v", err)
		}

		// Debug 模式下自动启用详细日志
		if config.ExecutionMode() == "debug" {
			log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
			log.Printf("[debug] execution_mode=debug, verbose logging enabled")
		}

		log.Printf("DB: %s (%s)", config.Global.Database.Driver, config.Global.Database.DSN)
		// Validate external integration configurations at startup
		for name, intCfg := range config.Global.ExternalIntegrations {
			knownTypes := map[string]bool{"http_adapter": true, "rpc_adapter": true, "websocket_adapter": true}
			if !knownTypes[intCfg.Type] {
				log.Printf("[warn] external integration %q has unsupported type %q", name, intCfg.Type)
			} else {
				log.Printf("external integration %q: type=%s", name, intCfg.Type)
			}
		}

		var provider engine.LLMProvider
		switch config.Global.LLM.Provider {
		case "fixture":
			fixtureProvider, err := llm.NewFixtureProvider(config.Global.LLM.Model, config.Global.LLM.FixtureFile)
			if err != nil {
				log.Fatalf("llm fixture provider: %v", err)
			}
			provider = fixtureProvider
			log.Printf("LLM: fixture provider (%s) file=%s", config.Global.LLM.Model, config.Global.LLM.FixtureFile)
		case "mock":
			provider = llm.NewMockProvider()
			log.Print("LLM: using mock provider by config")
		case "", "openai":
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
		default:
			log.Fatalf("unsupported llm.provider: %s", config.Global.LLM.Provider)
		}

		pipeline := engine.NewPipeline(provider)
		schedulerCtx, stopScheduler := context.WithCancel(context.Background())
		defer stopScheduler()
		governorCtx, stopGovernor := context.WithCancel(context.Background())
		defer stopGovernor()
		if config.Global.Engine.AutonomousSchedulerEnabled {
			interval := time.Duration(config.Global.Engine.AutonomousSchedulerIntervalSeconds) * time.Second
			limit := config.Global.Engine.AutonomousSchedulerMaxNodesPerWorld
			log.Printf("autonomous scheduler enabled: interval=%s limit=%d", interval, limit)
			go agent.NewScheduler(pipeline, interval, limit).Start(schedulerCtx)
		}
		governanceInterval := time.Duration(config.Global.Engine.RuntimeTaskGovernanceIntervalSeconds) * time.Second
		heartbeatTimeout := time.Duration(config.Global.Engine.RuntimeTaskHeartbeatTimeoutSeconds) * time.Second
		if governanceInterval > 0 && heartbeatTimeout > 0 {
			governor := service.NewRuntimeTaskGovernor(governanceInterval, service.RuntimeTaskGovernanceOptions{
				HeartbeatTimeout:  heartbeatTimeout,
				AutoRequeue:       config.Global.Engine.RuntimeTaskAutoRequeueEnabled,
				AutoRequeueLimit:  config.Global.Engine.RuntimeTaskAutoRequeueLimit,
				AutoRequeueDelay:  time.Duration(config.Global.Engine.RuntimeTaskAutoRequeueDelayMs) * time.Millisecond,
				AutoRequeueReason: "auto requeue after heartbeat timeout",
			})
			log.Printf("runtime task governor enabled: interval=%s heartbeat_timeout=%s auto_requeue=%t", governanceInterval, heartbeatTimeout, config.Global.Engine.RuntimeTaskAutoRequeueEnabled)
			go governor.Start(governorCtx)
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
			stopGovernor()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Fatalf("shutdown: %v", err)
			}
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	},
}
