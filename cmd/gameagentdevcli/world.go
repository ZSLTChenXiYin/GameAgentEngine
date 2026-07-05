package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

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
		var limit *int
		if cmd.Flags().Changed("autonomous-limit") {
			v, _ := cmd.Flags().GetInt("autonomous-limit")
			limit = &v
		}
		tr, err := newClient().AdvanceTickWithAutonomousLimit(args[0], tickType, gameTime, limit)
		if err != nil {
			fail(err)
		}
		printJSON(tr)
	},
}

var worldEventImpactCmd = &cobra.Command{
	Use:   "event-impact <world-id>",
	Short: "评估一个事件对世界的影响",
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
	Short: "重新生成世界未来时间线大纲",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().TimelineReplan(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}


var cloneLock bool

var worldCloneCmd = &cobra.Command{
	Use:   "clone <world-id> [name]",
	Short: "复制世界及其全部数据",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		worldID := args[0]
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		result, err := newClient().CloneWorld(worldID, name, cloneLock)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var worldSnapshotCmd = &cobra.Command{
	Use:   "snapshot <world-id>",
	Short: "输出世界当前的确切运行快照",
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
	Short: "导出为可再次导入的世界配置",
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
	Short: "管理世界级动作策略",
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
	Short: "管理世界级运行设置",
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
		settings := &sdk.WorldSettings{}
		if cmd.Flags().Changed("memory-limit") {
			v, _ := cmd.Flags().GetInt("memory-limit")
			settings.MemoryLimit = v
		}
		if cmd.Flags().Changed("analysis-rounds") {
			v, _ := cmd.Flags().GetInt("analysis-rounds")
			settings.MaxAnalysisRounds = v
		}
		if cmd.Flags().Changed("context-depth") {
			v, _ := cmd.Flags().GetInt("context-depth")
			settings.MaxContextDepth = v
		}
		if cmd.Flags().Changed("auto-apply") {
			v, _ := cmd.Flags().GetBool("auto-apply")
			settings.AutoApply = v
		}
		if cmd.Flags().Changed("review-above") {
			v, _ := cmd.Flags().GetString("review-above")
			settings.RequireReviewAbove = v
		}
		if cmd.Flags().Changed("propagation-max-depth") {
			v, _ := cmd.Flags().GetInt("propagation-max-depth")
			settings.PropagationMaxDepth = v
		}
		if cmd.Flags().Changed("enable-propagation-machine") {
			v, _ := cmd.Flags().GetBool("enable-propagation-machine")
			settings.EnablePropagationMachine = v
		}
		if cmd.Flags().Changed("sub-task-max-retries") {
			v, _ := cmd.Flags().GetInt("sub-task-max-retries")
			settings.SubTaskMaxRetries = v
		}
		if cmd.Flags().Changed("sub-task-timeout-secs") {
			v, _ := cmd.Flags().GetInt("sub-task-timeout-secs")
			settings.SubTaskTimeoutSecs = v
		if cmd.Flags().Changed("pipeline-mode") {
			v, _ := cmd.Flags().GetString("pipeline-mode")
			settings.PipelineMode = v
		}
		}
		result, err := newClient().SetWorldSettings(args[0], settings)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

func init() {
	worldCmd.AddCommand(worldTickCmd, worldEventImpactCmd, worldScopeAdvanceCmd, worldReplanCmd, worldCloneCmd, worldSnapshotCmd, worldExportCmd, worldPolicyCmd, worldSettingsCmd)

	worldPolicyCmd.AddCommand(worldPolicyGetCmd, worldPolicySetCmd)
	worldPolicySetCmd.Flags().StringSlice("blocked", []string{}, "阻塞的动作列表，逗号分隔")
	worldPolicySetCmd.Flags().StringSlice("safe", []string{}, "安全的动作列表，逗号分隔")

	worldSettingsCmd.AddCommand(worldSettingsGetCmd, worldSettingsSetCmd)
	worldSettingsSetCmd.Flags().Int("memory-limit", 0, "单次推理最多加载的记忆条数")
	worldSettingsSetCmd.Flags().Int("analysis-rounds", 0, "LLM 内部轮询最大次数")
	worldSettingsSetCmd.Flags().Int("context-depth", 0, "上下文向上追溯的最大深度")
	worldSettingsSetCmd.Flags().Bool("auto-apply", false, "是否自动执行变更计划")
	worldSettingsSetCmd.Flags().String("review-above", "", "超过此影响等级需要审核 (minor/major/critical)")
	worldSettingsSetCmd.Flags().Int("propagation-max-depth", 0, "记忆沿父链上传的最大层数；0 为不限制")
	worldSettingsSetCmd.Flags().Bool("enable-propagation-machine", false, "是否启用标签传播状态机")
	worldSettingsSetCmd.Flags().Int("sub-task-max-retries", 0, "子任务最大重试次数；0 使用默认值(2)")
	worldSettingsSetCmd.Flags().Int("sub-task-timeout-secs", 0, "子任务超时秒数；0 使用默认值(60)")
	worldSettingsSetCmd.Flags().String("pipeline-mode", "", "管线模式：vertical/polling/full；留空不修改")

	worldTickCmd.Flags().String("type", "manual", "刻推进类型")
	worldTickCmd.Flags().String("time", "dev-cli", "游戏内时间标识")
	worldTickCmd.Flags().Int("autonomous-limit", 10, "本次 tick 最多触发的 world_tick_sync 自主节点数；0 为不触发")

	worldEventImpactCmd.Flags().String("type", "", "事件类型")
	worldEventImpactCmd.Flags().String("scope", "", "事件作用范围节点 ID")
	worldEventImpactCmd.Flags().String("description", "", "事件描述")
	worldEventImpactCmd.Flags().String("severity", "medium", "事件严重程度")

	worldSnapshotCmd.Flags().String("out", "", "输出文件路径；为空时打印到 stdout")
	worldExportCmd.Flags().String("format", "yaml", "导出格式：yaml 或 json")
	worldExportCmd.Flags().String("out", "", "输出文件路径；为空时打印到 stdout")
}
