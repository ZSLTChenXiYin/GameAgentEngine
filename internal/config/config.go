// Package config 负责加载和持有应用配置。
// 配置默认读取 gameagentengine.conf.yaml，并允许通过环境变量覆盖。
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 是应用的总配置对象。
type Config struct {
	Server               ServerConfig                         `mapstructure:"server"`
	Database             DatabaseConfig                       `mapstructure:"database"`
	Auth                 AuthConfig                           `mapstructure:"auth"`
	LLM                  LLMConfig                            `mapstructure:"llm"`
	Engine               EngineConfig                         `mapstructure:"engine"`
	ExternalIntegrations map[string]ExternalIntegrationConfig `mapstructure:"external_integrations"`
	ExternalInterfaces   map[string]ExternalInterfaceConfig   `mapstructure:"external_interfaces"`
}

// ServerConfig 定义 HTTP 服务监听参数。
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig 定义数据库驱动和连接串。
type DatabaseConfig struct {
	Driver                string `mapstructure:"driver"`
	DSN                   string `mapstructure:"dsn"`
	MigrationsEnabled     bool   `mapstructure:"migrations_enabled"`
	WriteRetryEnabled     bool   `mapstructure:"write_retry_enabled"`
	WriteRetryMaxAttempts int    `mapstructure:"write_retry_max_attempts"`
	WriteRetryBaseDelayMs int    `mapstructure:"write_retry_base_delay_ms"`
	WriteRetryMaxDelayMs  int    `mapstructure:"write_retry_max_delay_ms"`
	LogBatchEnabled       bool   `mapstructure:"log_batch_enabled"`
	LogBatchSize          int    `mapstructure:"log_batch_size"`
	LogBatchFlushMs       int    `mapstructure:"log_batch_flush_ms"`
	LogBatchQueueSize     int    `mapstructure:"log_batch_queue_size"`
}

// AuthConfig 定义 API 鉴权配置。
type AuthConfig struct {
	APIKey                   string `mapstructure:"api_key"`
	CallbackToken            string `mapstructure:"callback_token"`
	RuntimeTaskToken         string `mapstructure:"runtime_task_token"`
	CallbackRequireRequestID bool   `mapstructure:"callback_require_request_id"`
}

// LLMConfig 定义大模型提供方接入参数。
type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
}

// EngineConfig 定义推理引擎运行参数。
type EngineConfig struct {
	ExecutionMode                        string `mapstructure:"execution_mode"`
	WorldLockEnabled                     bool   `mapstructure:"world_lock_enabled"`
	AutonomousSchedulerEnabled           bool   `mapstructure:"autonomous_scheduler_enabled"`
	AutonomousSchedulerIntervalSeconds   int    `mapstructure:"autonomous_scheduler_interval_seconds"`
	AutonomousSchedulerMaxNodesPerWorld  int    `mapstructure:"autonomous_scheduler_max_nodes_per_world"`
	RuntimeTaskGovernanceIntervalSeconds int    `mapstructure:"runtime_task_governance_interval_seconds"`
	RuntimeTaskHeartbeatTimeoutSeconds   int    `mapstructure:"runtime_task_heartbeat_timeout_seconds"`
	RuntimeTaskAutoRequeueEnabled        bool   `mapstructure:"runtime_task_auto_requeue_enabled"`
	RuntimeTaskAutoRequeueLimit          int    `mapstructure:"runtime_task_auto_requeue_limit"`
	RuntimeTaskAutoRequeueDelayMs        int    `mapstructure:"runtime_task_auto_requeue_delay_ms"`
}

// ExternalIntegrationConfig 定义一个可被 Engine 主动调用的外部接入点。
type ExternalIntegrationConfig struct {
	Type              string                        `mapstructure:"type"`
	BaseURL           string                        `mapstructure:"base_url"`
	Path              string                        `mapstructure:"path"`
	TimeoutMs         int                           `mapstructure:"timeout_ms"`
	RetryMaxAttempts  int                           `mapstructure:"retry_max_attempts"`
	RetryBackoffMs    int                           `mapstructure:"retry_backoff_ms"`
	IdempotencyHeader string                        `mapstructure:"idempotency_header"`
	Headers           map[string]string             `mapstructure:"headers"`
	Auth              ExternalIntegrationAuthConfig `mapstructure:"auth"`
}

// ExternalIntegrationAuthConfig 定义外部接入点的认证方式。
type ExternalIntegrationAuthConfig struct {
	Mode       string `mapstructure:"mode"`
	Token      string `mapstructure:"token"`
	HeaderName string `mapstructure:"header_name"`
}

// ExternalInterfaceConfig 定义一个业务接口的正式投递策略。
type ExternalInterfaceConfig struct {
	Category               string `mapstructure:"category"`
	DeliveryMode           string `mapstructure:"delivery_mode"`
	PrimaryTransport       string `mapstructure:"primary_transport"`
	FallbackTransport      string `mapstructure:"fallback_transport"`
	Consumer               string `mapstructure:"consumer"`
	ResumePolicy           string `mapstructure:"resume_policy"`
	CallbackPostProcess    string `mapstructure:"callback_post_process"`
	CallbackMemoryLevel    string `mapstructure:"callback_memory_level"`
	CallbackMemoryTemplate string `mapstructure:"callback_memory_template"`
	MaxAttempts            int    `mapstructure:"max_attempts"`
	TimeoutMs              int    `mapstructure:"timeout_ms"`
}

var Global Config

// Init 读取配置文件并填充全局配置对象。
// 当未显式指定配置文件时，会按约定路径搜索 gameagentengine.conf.yaml。
func Init(configPath string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("gameagentengine.conf")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "gameagentengine.db")
	v.SetDefault("database.migrations_enabled", true)
	v.SetDefault("database.write_retry_enabled", true)
	v.SetDefault("database.write_retry_max_attempts", 3)
	v.SetDefault("database.write_retry_base_delay_ms", 40)
	v.SetDefault("database.write_retry_max_delay_ms", 250)
	v.SetDefault("database.log_batch_enabled", true)
	v.SetDefault("database.log_batch_size", 32)
	v.SetDefault("database.log_batch_flush_ms", 750)
	v.SetDefault("database.log_batch_queue_size", 1024)
	v.SetDefault("auth.api_key", "dev-key")
	v.SetDefault("auth.callback_token", "")
	v.SetDefault("auth.runtime_task_token", "")
	v.SetDefault("auth.callback_require_request_id", false)
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("engine.execution_mode", "full")
	v.SetDefault("engine.world_lock_enabled", true)
	v.SetDefault("engine.autonomous_scheduler_enabled", false)
	v.SetDefault("engine.autonomous_scheduler_interval_seconds", 300)
	v.SetDefault("engine.autonomous_scheduler_max_nodes_per_world", 10)
	v.SetDefault("engine.runtime_task_governance_interval_seconds", 30)
	v.SetDefault("engine.runtime_task_heartbeat_timeout_seconds", 300)
	v.SetDefault("engine.runtime_task_auto_requeue_enabled", false)
	v.SetDefault("engine.runtime_task_auto_requeue_limit", 100)
	v.SetDefault("engine.runtime_task_auto_requeue_delay_ms", 1000)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}
	if err := v.Unmarshal(&Global); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	Global.Database.DSN = resolveDatabaseDSN(Global.Database.Driver, Global.Database.DSN, v.ConfigFileUsed())
	return nil
}

func resolveDatabaseDSN(driver, dsn, configFile string) string {
	if !strings.EqualFold(driver, "sqlite") {
		return dsn
	}
	if dsn == "" || filepath.IsAbs(dsn) || strings.HasPrefix(strings.ToLower(dsn), "file:") {
		return dsn
	}
	if configFile == "" {
		return dsn
	}
	return filepath.Join(filepath.Dir(configFile), dsn)
}

// ExecutionMode 返回当前生效的执行模式。
// 当配置缺失时，默认回退到 full。
func ExecutionMode() string {
	if Global.Engine.ExecutionMode != "" {
		return Global.Engine.ExecutionMode
	}
	return "full"
}
