// Package config 负责加载和持有应用配置。
// 配置默认读取 gameagentengine.conf.yaml，并允许通过环境变量覆盖。
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 是应用的总配置对象。
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Engine   EngineConfig   `mapstructure:"engine"`
}

// ServerConfig 定义 HTTP 服务监听参数。
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig 定义数据库驱动和连接串。
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// AuthConfig 定义 API 鉴权配置。
type AuthConfig struct {
	APIKey string `mapstructure:"api_key"`
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
	ExecutionMode                       string `mapstructure:"execution_mode"`
	AutonomousSchedulerEnabled          bool   `mapstructure:"autonomous_scheduler_enabled"`
	AutonomousSchedulerIntervalSeconds  int    `mapstructure:"autonomous_scheduler_interval_seconds"`
	AutonomousSchedulerMaxNodesPerWorld int    `mapstructure:"autonomous_scheduler_max_nodes_per_world"`
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
	v.SetDefault("auth.api_key", "dev-key")
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("engine.execution_mode", "full")
	v.SetDefault("engine.autonomous_scheduler_enabled", false)
	v.SetDefault("engine.autonomous_scheduler_interval_seconds", 300)
	v.SetDefault("engine.autonomous_scheduler_max_nodes_per_world", 10)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}
	if err := v.Unmarshal(&Global); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

// ExecutionMode 返回当前生效的执行模式。
// 当配置缺失时，默认回退到 full。
func ExecutionMode() string {
	if Global.Engine.ExecutionMode != "" {
		return Global.Engine.ExecutionMode
	}
	return "full"
}
