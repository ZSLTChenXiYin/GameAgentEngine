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
	Short: "绠＄悊涓栫晫绾ц繍琛屾椂鎿嶄綔",
}

var worldTickCmd = &cobra.Command{
	Use:   "tick <world-id>",
	Short: "鎺ㄨ繘涓€娆′笘鐣屽埢",
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
	Short: "鎺ㄨ繘鏌愪釜灞€閮ㄨ寖鍥寸殑涓栫晫婕斿寲",
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
	Short: "鍒涘缓涓栫晫宸ヤ綔鍓湰",
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
	Short: "鍒涘缓涓栫晫瀛樻。蹇収",
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
	Short: "浠庡瓨妗ｅ揩鐓ф仮澶嶆柊涓栫晫",
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
	Short: "妤犲矁鐦夌€涙ɑ銆傝箛顐ゅ弾閺勵垰鎯侀崣顖欎簰鐎瑰鍙忛幁銏狀槻",
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
	Short: "閺屻儳婀呯€涙ɑ銆傝箛顐ゅ弾閸忓啯鏆熼幑顔款嚊閹?",
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
	Short: "閸掓鍤弻鎰嚋娑撴牜鏅惃鍕摠濡楋絽鎻╅悡褍鍨悰?",
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
	Short: "閸掔娀娅庣€涙ɑ銆傝箛顐ゅ弾閸欏﹤鍙鹃幍鈧€电懓绨查惃鍕瑯閻ｅ苯澹囬張?",
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
	Short: "绠＄悊寰呭鎵圭殑涓栫晫鍙樻洿璁″垝",
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
	Short: "鎵瑰噯涓€鏉″緟瀹℃壒璁″垝",
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
	Short: "鎷掔粷涓€鏉″緟瀹℃壒璁″垝",
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
	Short: "鏌ョ湅涓栫晫鍔ㄤ綔绛栫暐",
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
	Short: "璁剧疆涓栫晫鍔ㄤ綔绛栫暐",
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
	Short: "鏌ョ湅涓栫晫杩愯璁剧疆",
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
	Short: "璁剧疆涓栫晫杩愯璁剧疆",
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
	Short: "鏇存柊涓栫晫淇℃伅",
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
	worldPolicySetCmd.Flags().StringSlice("blocked", []string{}, "闃诲鐨勫姩浣滃垪琛紝閫楀彿鍒嗛殧")
	worldPolicySetCmd.Flags().StringSlice("safe", []string{}, "瀹夊叏鐨勫姩浣滃垪琛紝閫楀彿鍒嗛殧")

	worldSettingsCmd.AddCommand(worldSettingsGetCmd, worldSettingsSetCmd)
	worldUpdateCmd.Flags().String("name", "", "鏂扮殑涓栫晫鍚嶇О")
	worldSettingsSetCmd.Flags().Int("memory-limit", 0, "鍗曟鎺ㄧ悊鏈€澶氬姞杞界殑璁板繂鏉℃暟")
	worldSettingsSetCmd.Flags().Int("analysis-rounds", 0, "Maximum internal analysis rounds for the LLM")
	worldSettingsSetCmd.Flags().Int("context-depth", 0, "Maximum ancestor context lookup depth")
	worldSettingsSetCmd.Flags().Bool("auto-apply", false, "鏄惁鑷姩鎵ц鍙樻洿璁″垝")
	worldSettingsSetCmd.Flags().String("review-above", "", "Require review above this impact level")
	worldSettingsSetCmd.Flags().Int("propagation-max-depth", 0, "璁板繂娌跨埗閾句笂浼犵殑鏈€澶у眰鏁帮紱0 涓轰笉闄愬埗")
	worldSettingsSetCmd.Flags().Bool("enable-propagation-machine", false, "鏄惁鍚敤鏍囩浼犳挱鐘舵€佹満")
	worldSettingsSetCmd.Flags().Int("sub-task-max-retries", 0, "瀛愪换鍔℃渶澶ч噸璇曟鏁帮紱0 浣跨敤榛樿鍊?2)")
	worldSettingsSetCmd.Flags().Int("sub-task-timeout-secs", 0, "瀛愪换鍔¤秴鏃剁鏁帮紱0 浣跨敤榛樿鍊?60)")
	worldSettingsSetCmd.Flags().String("pipeline-mode", "", "绠＄嚎妯″紡锛歷ertical/polling/full锛涚暀绌轰笉淇敼")
	worldSettingsSetCmd.Flags().String("world-time-settings-json", "", "JSON string for world_time_settings")
	worldSettingsSetCmd.Flags().String("world-time-settings-file", "", "Read world_time_settings JSON from file")

	worldTickCmd.Flags().String("type", "manual", "Tick request type")
	worldTickCmd.Flags().String("time", "dev-cli", "Game time label for the request")
	worldTickCmd.Flags().Int("requested-ticks", 1, "鏈 world tick 璇锋眰鎺ㄨ繘鐨勫熀纭€ tick 鏁伴噺")
	worldTickCmd.Flags().Int("autonomous-limit", 10, "鏈 tick 鏈€澶氳Е鍙戠殑 world_tick_sync 鑷富鑺傜偣鏁帮紱0 涓轰笉瑙﹀彂")

	worldEventImpactCmd.Flags().String("type", "", "浜嬩欢绫诲瀷")
	worldEventImpactCmd.Flags().String("scope", "", "浜嬩欢浣滅敤鑼冨洿鑺傜偣 ID")
	worldEventImpactCmd.Flags().String("description", "", "浜嬩欢鎻忚堪")
	worldEventImpactCmd.Flags().String("severity", "medium", "浜嬩欢涓ラ噸绋嬪害")

	worldSnapshotCmd.Flags().String("out", "", "Output file path; print to stdout when empty")
	worldExportCmd.Flags().String("format", "yaml", "Export format: yaml or json")
	worldExportCmd.Flags().String("out", "", "Output file path; print to stdout when empty")
	worldForkCmd.Flags().BoolVar(&worldCopyLock, "lock-world", false, "Lock the source world during fork")
	worldSaveCmd.Flags().BoolVar(&worldCopyLock, "lock-world", true, "Lock the source world during snapshot save")
	worldRestoreCmd.Flags().BoolVar(&worldCopyLock, "lock-world", true, "Lock the snapshot world during restore")
}
