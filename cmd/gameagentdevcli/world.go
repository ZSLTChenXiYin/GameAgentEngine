package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func parseWorldTimeSettingsInput(jsonText, filePath string) (*sdk.WorldTimeSettings, error) {
	if jsonText != "" && filePath != "" {
		return nil, fmt.Errorf("use either --world-time-settings-json or --world-time-settings-file, not both")
	}
	if jsonText == "" && filePath == "" {
		return nil, nil
	}
	var payload []byte
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		payload = data
	} else {
		payload = []byte(jsonText)
	}
	var worldTimeSettings sdk.WorldTimeSettings
	if err := json.Unmarshal(payload, &worldTimeSettings); err != nil {
		return nil, fmt.Errorf("invalid world_time_settings JSON: %w", err)
	}
	return &worldTimeSettings, nil
}

var worldCmd = &cobra.Command{
	Use:   "world",
	Short: "管理世界级运行时操作",
}

var worldTickCmd = &cobra.Command{
	Use:   "tick <world-id>",
	Short: "推进一次世界刻",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tickType, _ := cmd.Flags().GetString("type")
		gameTime, _ := cmd.Flags().GetString("time")
		var requestedTicks *int
		if cmd.Flags().Changed("requested-ticks") {
			v, _ := cmd.Flags().GetInt("requested-ticks")
			requestedTicks = &v
		}
		if err := validateRequestedTicksForWorld(args[0], requestedTicks); err != nil {
			fail(err)
		}
		var limit *int
		if cmd.Flags().Changed("autonomous-limit") {
			v, _ := cmd.Flags().GetInt("autonomous-limit")
			limit = &v
		}
		tr, err := newClient().AdvanceTickWithOptions(args[0], tickType, gameTime, requestedTicks, limit)
		if err != nil {
			fail(err)
		}
		printJSON(tr)
	},
}

var worldEventImpactCmd = &cobra.Command{
	Use:   "event-impact <world-id>",
	Short: "Evaluate a world event impact",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eventType, _ := cmd.Flags().GetString("type")
		scopeID, _ := cmd.Flags().GetString("scope")
		description, _ := cmd.Flags().GetString("description")
		severity, _ := cmd.Flags().GetString("severity")

		if eventType == "" || description == "" {
			fail(fmt.Errorf("--type and --description are required"))
		}

		resp, err := newClient().EventImpact(args[0], &sdk.WorldEvent{
			EventType:   eventType,
			ScopeID:     scopeID,
			Description: description,
			Severity:    severity,
		})
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var worldScopeAdvanceCmd = &cobra.Command{
	Use:   "scope-advance <world-id> <scope-id>",
	Short: "推进某个局部范围的世界演化",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().ScopeAdvance(args[0], args[1])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var worldReplanCmd = &cobra.Command{
	Use:   "replan <world-id>",
	Short: "Rebuild future world timeline plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().TimelineReplan(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var worldCopyLock bool

var worldForkCmd = &cobra.Command{
	Use:   "fork <world-id> [name]",
	Short: "创建世界工作副本",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		worldID := args[0]
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		result, err := newClient().ForkWorld(worldID, name, worldCopyLock)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldSaveCmd = &cobra.Command{
	Use:   "save <world-id> [name]",
	Short: "创建世界存档快照",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		worldID := args[0]
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		result, err := newClient().CreateWorldSnapshot(worldID, name, worldCopyLock)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldRestoreCmd = &cobra.Command{
	Use:   "restore <snapshot-world-id> [name]",
	Short: "从存档快照恢复新世界",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		worldID := args[0]
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		result, err := newClient().RestoreWorld(worldID, name, worldCopyLock)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldValidateSnapshotCmd = &cobra.Command{
	Use:   "validate-snapshot <snapshot-world-id>",
	Short: "校验快照兼容性与可恢复性",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().ValidateWorldSnapshot(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldSnapshotInfoCmd = &cobra.Command{
	Use:   "snapshot-info <snapshot-world-id>",
	Short: "查看快照元数据详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().GetWorldSnapshotMetadata(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldListSnapshotsCmd = &cobra.Command{
	Use:   "list-snapshots <world-id>",
	Short: "列出世界相关的所有存档快照",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().ListWorldSnapshots(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldDeleteSnapshotCmd = &cobra.Command{
	Use:   "delete-snapshot <snapshot-world-id>",
	Short: "删除快照世界及其元数据",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().DeleteWorldSnapshot(args[0]); err != nil {
			fail(err)
		}
		printJSON(map[string]any{"status": "deleted", "snapshot_world_id": args[0]})
	},
}

var worldSnapshotCmd = &cobra.Command{
	Use:   "snapshot <world-id>",
	Short: "Export the current world runtime snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := buildWorldSnapshot(args[0])
		if err != nil {
			fail(err)
		}
		outPath, _ := cmd.Flags().GetString("out")
		if err := writeJSONOutput(snapshot, outPath); err != nil {
			fail(err)
		}
	},
}

var worldExportCmd = &cobra.Command{
	Use:   "export <world-id>",
	Short: "Export a world config that can be imported again",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := buildWorldSnapshot(args[0])
		if err != nil {
			fail(err)
		}
		cfg := buildImportConfigFromSnapshot(snapshot)
		format, _ := cmd.Flags().GetString("format")
		outPath, _ := cmd.Flags().GetString("out")
		if err := writeStructuredOutput(cfg, format, outPath); err != nil {
			fail(err)
		}
	},
}

var worldPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage world action policy",
}

var worldPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "管理待审批的世界变更计划",
}

var worldPlanPendingCmd = &cobra.Command{
	Use:   "pending [world-id]",
	Short: "List pending world change plans",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		worldID := ""
		if len(args) > 0 {
			worldID = args[0]
		}
		plans, err := newClient().ListPendingPlans(worldID)
		if err != nil {
			fail(err)
		}
		printJSON(plans)
	},
}

var worldPlanApproveCmd = &cobra.Command{
	Use:   "approve <world-id> <plan-id>",
	Short: "批准一条待审批计划",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().ApprovePlan(args[0], args[1])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var worldPlanRejectCmd = &cobra.Command{
	Use:   "reject <world-id> <plan-id>",
	Short: "拒绝一条待审批计划",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().RejectPlan(args[0], args[1])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var worldPolicyGetCmd = &cobra.Command{
	Use:   "get <world-id>",
	Short: "查看世界动作策略",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		policy, err := newClient().GetWorldPolicy(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(policy)
	},
}

var worldPolicySetCmd = &cobra.Command{
	Use:   "set <world-id>",
	Short: "设置世界动作策略",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		blocked, _ := cmd.Flags().GetStringSlice("blocked")
		safe, _ := cmd.Flags().GetStringSlice("safe")
		policy, err := newClient().SetWorldPolicy(args[0], blocked, safe)
		if err != nil {
			fail(err)
		}
		printJSON(policy)
	},
}

var worldSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage world runtime settings",
}

var worldSettingsGetCmd = &cobra.Command{
	Use:   "get <world-id>",
	Short: "查看世界运行设置",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		settings, err := newClient().GetWorldSettings(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(settings)
	},
}

var worldSettingsSetCmd = &cobra.Command{
	Use:   "set <world-id>",
	Short: "设置世界运行设置",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		settings := &sdk.WorldSettingsUpdate{}
		if cmd.Flags().Changed("memory-limit") {
			v, _ := cmd.Flags().GetInt("memory-limit")
			settings.MemoryLimit = &v
		}
		if cmd.Flags().Changed("analysis-rounds") {
			v, _ := cmd.Flags().GetInt("analysis-rounds")
			settings.MaxAnalysisRounds = &v
		}
		if cmd.Flags().Changed("context-depth") {
			v, _ := cmd.Flags().GetInt("context-depth")
			settings.MaxContextDepth = &v
		}
		if cmd.Flags().Changed("auto-apply") {
			v, _ := cmd.Flags().GetBool("auto-apply")
			settings.AutoApply = &v
		}
		if cmd.Flags().Changed("review-above") {
			v, _ := cmd.Flags().GetString("review-above")
			settings.RequireReviewAbove = &v
		}
		if cmd.Flags().Changed("propagation-max-depth") {
			v, _ := cmd.Flags().GetInt("propagation-max-depth")
			settings.PropagationMaxDepth = &v
		}
		if cmd.Flags().Changed("enable-propagation-machine") {
			v, _ := cmd.Flags().GetBool("enable-propagation-machine")
			settings.EnablePropagationMachine = &v
		}
		if cmd.Flags().Changed("sub-task-max-retries") {
			v, _ := cmd.Flags().GetInt("sub-task-max-retries")
			settings.SubTaskMaxRetries = &v
		}
		if cmd.Flags().Changed("sub-task-timeout-secs") {
			v, _ := cmd.Flags().GetInt("sub-task-timeout-secs")
			settings.SubTaskTimeoutSecs = &v
		}
		if cmd.Flags().Changed("pipeline-mode") {
			v, _ := cmd.Flags().GetString("pipeline-mode")
			settings.PipelineMode = &v
		}
		worldTimeSettingsJSON, _ := cmd.Flags().GetString("world-time-settings-json")
		worldTimeSettingsFile, _ := cmd.Flags().GetString("world-time-settings-file")
		if worldTimeSettingsJSON != "" || worldTimeSettingsFile != "" {
			worldTimeSettings, err := parseWorldTimeSettingsInput(worldTimeSettingsJSON, worldTimeSettingsFile)
			if err != nil {
				fail(err)
			}
			settings.WorldTimeSettings = worldTimeSettings
		}
		result, err := newClient().UpdateWorldSettings(args[0], settings)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldUpdateCmd = &cobra.Command{
	Use:   "update <world-id>",
	Short: "更新世界信息",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, changed := getChangedString(cmd, "name")
		if !changed {
			fail(fmt.Errorf("no update flags provided"))
		}
		result, err := newClient().UpdateWorld(args[0], name)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

func init() {
	worldCmd.AddCommand(worldTickCmd, worldEventImpactCmd, worldScopeAdvanceCmd, worldReplanCmd, worldForkCmd, worldSaveCmd, worldRestoreCmd, worldValidateSnapshotCmd, worldSnapshotInfoCmd, worldListSnapshotsCmd, worldDeleteSnapshotCmd, worldSnapshotCmd, worldExportCmd, worldPolicyCmd, worldPlanCmd, worldSettingsCmd, worldUpdateCmd)

	worldPolicyCmd.AddCommand(worldPolicyGetCmd, worldPolicySetCmd)
	worldPlanCmd.AddCommand(worldPlanPendingCmd, worldPlanApproveCmd, worldPlanRejectCmd)
	worldPolicySetCmd.Flags().StringSlice("blocked", []string{}, "阻止的动作列表，逗号分隔")
	worldPolicySetCmd.Flags().StringSlice("safe", []string{}, "安全的动作列表，逗号分隔")

	worldSettingsCmd.AddCommand(worldSettingsGetCmd, worldSettingsSetCmd)
	worldUpdateCmd.Flags().String("name", "", "新的世界名称")
	worldSettingsSetCmd.Flags().Int("memory-limit", 0, "单次推理最多加载的记忆条数")
	worldSettingsSetCmd.Flags().Int("analysis-rounds", 0, "Maximum internal analysis rounds for the LLM")
	worldSettingsSetCmd.Flags().Int("context-depth", 0, "Maximum ancestor context lookup depth")
	worldSettingsSetCmd.Flags().Bool("auto-apply", false, "是否自动执行变更计划")
	worldSettingsSetCmd.Flags().String("review-above", "", "Require review above this impact level")
	worldSettingsSetCmd.Flags().Int("propagation-max-depth", 0, "默认 upward 主父链传播的最大深度；0 表示不限制")
	worldSettingsSetCmd.Flags().Bool("enable-propagation-machine", false, "是否启用标签传播状态机")
	worldSettingsSetCmd.Flags().Int("sub-task-max-retries", 0, "子任务最大重试次数；0 使用默认值 2")
	worldSettingsSetCmd.Flags().Int("sub-task-timeout-secs", 0, "子任务超时秒数；0 使用默认值 60")
	worldSettingsSetCmd.Flags().String("pipeline-mode", "", "管线模式：vertical / polling / full；留空表示不修改")
	worldSettingsSetCmd.Flags().String("world-time-settings-json", "", "JSON string for world_time_settings")
	worldSettingsSetCmd.Flags().String("world-time-settings-file", "", "Read world_time_settings JSON from file")

	worldTickCmd.Flags().String("type", "manual", "Tick request type")
	worldTickCmd.Flags().String("time", "dev-cli", "Game time label for the request")
	worldTickCmd.Flags().Int("requested-ticks", 1, "本次 world tick 请求推进的基础 tick 数量")
	worldTickCmd.Flags().Int("autonomous-limit", 10, "本次 tick 最多触发的 world_tick_sync 自主节点数量；0 表示不触发")

	worldEventImpactCmd.Flags().String("type", "", "事件类型")
	worldEventImpactCmd.Flags().String("scope", "", "事件作用范围节点 ID")
	worldEventImpactCmd.Flags().String("description", "", "事件描述")
	worldEventImpactCmd.Flags().String("severity", "medium", "事件严重程度")

	worldSnapshotCmd.Flags().String("out", "", "Output file path; print to stdout when empty")
	worldExportCmd.Flags().String("format", "yaml", "Export format: yaml or json")
	worldExportCmd.Flags().String("out", "", "Output file path; print to stdout when empty")
	worldForkCmd.Flags().BoolVar(&worldCopyLock, "lock-world", false, "Lock the source world during fork")
	worldSaveCmd.Flags().BoolVar(&worldCopyLock, "lock-world", true, "Lock the source world during snapshot save")
	worldRestoreCmd.Flags().BoolVar(&worldCopyLock, "lock-world", true, "Lock the snapshot world during restore")
}
