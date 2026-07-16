package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var (
	serverURL       string
	apiKey          string
	localConfigPath string
)

var rootCmd = &cobra.Command{
	Use:     "GameAgentDevCli",
	Aliases: []string{"gameagentdevcli"},
	Short:   "GameAgentDevCli",
	Long:    "GameAgentDevCli 用于通过 SDK/HTTP 接口管理 GameAgentEngine 世界、节点、组件、记忆与关系实例。",
}

// newClient 创建一个基于当前全局参数的 SDK 客户端。
func newClient() *sdk.Client {
	c := sdk.NewClient(serverURL, apiKey)
	if key, _ := rootCmd.Flags().GetString("idempotency-key"); key != "" {
		c = c.WithIdempotency(key)
	}
	return c
}

func initLocalStore() error {
	if localConfigPath == "" {
		return fmt.Errorf("--config is required for local reset operations")
	}
	if err := config.Init(localConfigPath); err != nil {
		return err
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
	dsn := config.Global.Database.DSN
	if !filepath.IsAbs(dsn) {
		dsn = filepath.Join(filepath.Dir(localConfigPath), dsn)
	}
	return store.Init(config.Global.Database.Driver, dsn)
}

// fail 将错误打印到标准错误并退出进程。
func fail(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// printJSON 以格式化 JSON 的形式输出结果。
func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fail(err)
	}
	fmt.Println(string(data))
}

// writeJSONOutput 将快照以格式化 JSON 输出到 stdout 或文件。
func writeJSONOutput(v any, outPath string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if outPath == "" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(outPath, data, 0o644)
}

// writeStructuredOutput 将结构化配置按指定格式输出到 stdout 或文件。
func writeStructuredOutput(v any, format, outPath string) error {
	var (
		data []byte
		err  error
	)
	if format == "json" {
		data, err = json.MarshalIndent(v, "", "  ")
	} else {
		data, err = yaml.Marshal(v)
	}
	if err != nil {
		return err
	}
	if outPath == "" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(outPath, data, 0o644)
}

// getChangedString 读取字符串 flag，并返回它是否被显式传入。
func getChangedString(cmd *cobra.Command, name string) (string, bool) {
	value, _ := cmd.Flags().GetString(name)
	return value, cmd.Flags().Changed(name)
}

// ptrIfChanged 在 flag 被显式传入时返回对应字符串指针。
func ptrIfChanged(value string, changed bool) *string {
	if !changed {
		return nil
	}
	return &value
}

// valueIfChanged 在 flag 未被传入时返回空字符串，便于兼容现有 SDK 更新签名。
func valueIfChanged(value string, changed bool) string {
	if !changed {
		return ""
	}
	return value
}

func validPropagationModesText() string {
	return "upward / environment_scope / organization_scope / tag_broadcast / targeted / manual"
}

func validRelationTypesText() string {
	return "belongs_to / ally / enemy / subordinate / kinship / located_at / external_parent"
}

func validatePropagationMode(value string) error {
	if value == "" {
		return nil
	}
	if !slices.Contains(sdk.ValidPropagationModes(), value) {
		return fmt.Errorf("invalid propagation mode %q; allowed: %s", value, validPropagationModesText())
	}
	return nil
}

func containsString(items []string, value string) bool {
	return slices.Contains(items, value)
}

// init 注册 GameAgentDevCli 的根命令和全局参数。
func init() {
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://127.0.0.1:8080", "Agent server URL")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "key", "k", "dev-key", "API key")
	rootCmd.PersistentFlags().StringVar(&localConfigPath, "config", "", "Local GameAgentEngine config path for local reset operations")
	rootCmd.PersistentFlags().String("idempotency-key", "", "为本次命令执行设置幂等 key，避免重复提交")
	rootCmd.PersistentFlags().Int("max-analysis-rounds", 0, "LLM 内部轮询最大次数（0 表示使用服务端配置）")
	rootCmd.PersistentFlags().Int("max-context-depth", 0, "上下文向上追溯最大深度（0 表示使用服务端配置）")
	rootCmd.PersistentFlags().Int("memory-limit", 0, "每次推理最多加载的记忆数量（0 表示使用服务端配置）")
	rootCmd.PersistentFlags().Bool("include-related-nodes", false, "是否启用受控关系补充；不会无差别展开所有邻接节点")
}

// main 执行 GameAgentDevCli 根命令。
func main() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(nodeLegacyListCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(memoryCmd)
	rootCmd.AddCommand(relationCmd)
	rootCmd.AddCommand(actionCmd)
	rootCmd.AddCommand(worldCmd)
	rootCmd.AddCommand(tickCmd)
	rootCmd.AddCommand(stateCmd)
	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(creatorCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(devCliVersionCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(taskCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// buildInvokeContext 从命令行标志构建 InvokeContext，供需要发起推理的命令使用。
func buildInvokeContext(cmd *cobra.Command) *sdk.InvokeContext {
	ctx := &sdk.InvokeContext{}
	anySet := false

	if v, _ := cmd.Flags().GetInt("max-analysis-rounds"); cmd.Flags().Changed("max-analysis-rounds") {
		ctx.MaxAnalysisRounds = v
		anySet = true
	}
	if v, _ := cmd.Flags().GetInt("max-context-depth"); cmd.Flags().Changed("max-context-depth") {
		ctx.MaxDepth = v
		anySet = true
	}
	if v, _ := cmd.Flags().GetInt("memory-limit"); cmd.Flags().Changed("memory-limit") {
		ctx.MemoryLimit = v
		anySet = true
	}
	if v, _ := cmd.Flags().GetBool("include-related-nodes"); cmd.Flags().Changed("include-related-nodes") {
		ctx.IncludeRelatedNodes = v
		anySet = true
	}
	if v, _ := cmd.Flags().GetString("pipeline-mode"); cmd.Flags().Changed("pipeline-mode") {
		ctx.PipelineMode = v
		anySet = true
	}
	if dynamicInterfaces, err := loadDynamicInterfaces(cmd); err != nil {
		fail(err)
	} else if len(dynamicInterfaces) > 0 {
		ctx.DynamicInterfaces = dynamicInterfaces
		anySet = true
	}

	if !anySet {
		return nil
	}
	return ctx
}

func loadDynamicInterfaces(cmd *cobra.Command) ([]sdk.DynamicInterface, error) {
	jsonText, _ := cmd.Flags().GetString("dynamic-interfaces-json")
	filePath, _ := cmd.Flags().GetString("dynamic-interfaces-file")
	if strings.TrimSpace(jsonText) != "" && strings.TrimSpace(filePath) != "" {
		return nil, fmt.Errorf("use either --dynamic-interfaces-json or --dynamic-interfaces-file, not both")
	}
	if strings.TrimSpace(jsonText) == "" && strings.TrimSpace(filePath) == "" {
		return nil, nil
	}
	var payload []byte
	if strings.TrimSpace(filePath) != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		payload = data
	} else {
		payload = []byte(jsonText)
	}
	var interfaces []sdk.DynamicInterface
	if err := json.Unmarshal(payload, &interfaces); err != nil {
		return nil, fmt.Errorf("invalid dynamic_interfaces JSON: %w", err)
	}
	return interfaces, nil
}

func validateRequestedTicksForWorld(worldID string, requestedTicks *int) error {
	if requestedTicks == nil {
		return nil
	}
	if *requestedTicks <= 0 {
		return fmt.Errorf("requested-ticks must be greater than 0")
	}
	settings, err := newClient().GetWorldSettings(worldID)
	if err != nil {
		return err
	}
	if settings == nil || settings.WorldTimeSettings == nil {
		return nil
	}
	if settings.WorldTimeSettings.TickScaleMode == "fixed" && *requestedTicks != 1 {
		return fmt.Errorf("fixed tick scale mode only allows requested-ticks = 1")
	}
	return nil
}
