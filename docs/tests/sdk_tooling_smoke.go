package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func main() {
	server := flag.String("server", "http://127.0.0.1:18084", "Engine base URL")
	key := flag.String("key", "dev-key", "API key")
	worldID := flag.String("world", "", "World ID")
	nodeID := flag.String("node", "", "Node ID")
	flag.Parse()

	if strings.TrimSpace(*worldID) == "" {
		fail("--world is required")
	}
	if strings.TrimSpace(*nodeID) == "" {
		fail("--node is required")
	}

	client := sdk.NewClient(*server, *key)

	nodes, err := client.GetNodes(*worldID, 50, 0, "")
	if err != nil {
		fail("sdk get nodes: %v", err)
	}
	if len(nodes) < 2 {
		fail("expected at least 2 nodes, got %d", len(nodes))
	}

	pendingTasks, err := client.ListPendingRuntimeTasks("bridge", 20)
	if err != nil {
		fail("sdk list pending runtime tasks: %v", err)
	}
	if len(pendingTasks) == 0 {
		fail("expected at least one pending runtime task")
	}
	if pendingTasks[0].InterfaceName != "npc_trade_action" {
		fail("expected first pending runtime task interface npc_trade_action, got %s", pendingTasks[0].InterfaceName)
	}

	stats, err := client.GetRuntimeTaskStats()
	if err != nil {
		fail("sdk get runtime task stats: %v", err)
	}
	if stats == nil || stats.Total < 1 {
		fail("expected runtime task stats total >= 1, got %+v", stats)
	}

	lat, err := client.GetLatestTimeline(*worldID)
	if err != nil {
		fail("sdk get latest timeline: %v", err)
	}
	if lat == nil || lat.Timeline.TickNumber < 1 {
		fail("expected latest timeline tick >= 1, got %+v", lat)
	}

	states, err := client.GetStateComponents(*worldID)
	if err != nil {
		fail("sdk get state components: %v", err)
	}
	if states == nil || len(states.Components) < 3 {
		fail("expected at least 3 state components, got %+v", states)
	}

	bundle, err := client.GetContinuityBundle(*worldID, &sdk.ContinuityBundleOptions{LogLimit: 20, TraceLimit: 10})
	if err != nil {
		fail("sdk get continuity bundle: %v", err)
	}
	if bundle == nil || bundle.LatestTimeline == nil {
		fail("expected continuity bundle latest timeline, got %+v", bundle)
	}
	if len(bundle.Logs) == 0 {
		fail("expected continuity bundle logs")
	}
	if len(bundle.Traces) == 0 {
		fail("expected continuity bundle traces")
	}

	traces, err := client.GetDebugTraces(*worldID, 10)
	if err != nil {
		fail("sdk get debug traces: %v", err)
	}
	if traces == nil || traces.Count < 1 {
		fail("expected at least one trace, got %+v", traces)
	}

	logs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: *worldID, TaskType: "world_tick", Limit: 20})
	if err != nil {
		fail("sdk get logs by query: %v", err)
	}
	if len(logs) == 0 {
		fail("expected world_tick logs")
	}

	result := map[string]any{
		"node_count":           len(nodes),
		"pending_task_id":      pendingTasks[0].TaskID,
		"pending_task_status":  pendingTasks[0].Status,
		"pending_task_iface":   pendingTasks[0].InterfaceName,
		"runtime_task_total":   stats.Total,
		"latest_tick_number":   lat.Timeline.TickNumber,
		"state_component_count": len(states.Components),
		"continuity_log_count": len(bundle.Logs),
		"continuity_trace_count": len(bundle.Traces),
		"trace_count":          traces.Count,
		"log_count":            len(logs),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fail("encode result: %v", err)
	}
}
