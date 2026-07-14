package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestLoadDynamicInterfacesFromJSONFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("dynamic-interfaces-json", "", "")
	cmd.Flags().String("dynamic-interfaces-file", "", "")
	if err := cmd.Flags().Set("dynamic-interfaces-json", `[{"id":"scene_facts","kind":"data_request","external_interface":"game_client_request_data","query_types":["node_detail"],"max_queries":2}]`); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	items, err := loadDynamicInterfaces(cmd)
	if err != nil {
		t.Fatalf("load dynamic interfaces: %v", err)
	}
	if len(items) != 1 || items[0].ExternalInterface != "game_client_request_data" || items[0].MaxQueries != 2 {
		t.Fatalf("unexpected dynamic interfaces: %#v", items)
	}
}

func TestLoadDynamicInterfacesFromFileFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dynamic-interfaces.json")
	if err := os.WriteFile(path, []byte(`[{"id":"merchant_ops","kind":"action","external_interface":"npc_trade_action","max_calls":1}]`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("dynamic-interfaces-json", "", "")
	cmd.Flags().String("dynamic-interfaces-file", "", "")
	if err := cmd.Flags().Set("dynamic-interfaces-file", path); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	items, err := loadDynamicInterfaces(cmd)
	if err != nil {
		t.Fatalf("load dynamic interfaces: %v", err)
	}
	if len(items) != 1 || items[0].Kind != sdk.DynamicInterfaceAction || items[0].MaxCalls != 1 {
		t.Fatalf("unexpected dynamic interfaces: %#v", items)
	}
}

func TestBuildInvokeRequestFromFlagsBuildsContextAndMessage(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("task-type", "", "")
	cmd.Flags().String("message", "", "")
	cmd.Flags().String("session-id", "", "")
	cmd.Flags().Int("max-analysis-rounds", 0, "")
	cmd.Flags().Int("max-context-depth", 0, "")
	cmd.Flags().Int("memory-limit", 0, "")
	cmd.Flags().Bool("include-related-nodes", false, "")
	cmd.Flags().String("pipeline-mode", "", "")
	cmd.Flags().String("dynamic-interfaces-json", "", "")
	cmd.Flags().String("dynamic-interfaces-file", "", "")

	_ = cmd.Flags().Set("task-type", "npc_dialogue")
	_ = cmd.Flags().Set("message", "你现在看到了什么？")
	_ = cmd.Flags().Set("session-id", "session-1")
	_ = cmd.Flags().Set("max-analysis-rounds", "4")
	_ = cmd.Flags().Set("pipeline-mode", sdk.PipelineModePolling)
	_ = cmd.Flags().Set("dynamic-interfaces-json", `[{"id":"scene_facts","kind":"data_request","external_interface":"game_client_request_data","query_types":["node_detail"]}]`)

	req, err := buildInvokeRequestFromFlags(cmd, "world-1", "node-1")
	if err != nil {
		t.Fatalf("build invoke request: %v", err)
	}
	if req.WorldID != "world-1" || req.NodeID != "node-1" || req.TaskType != "npc_dialogue" || req.SessionID != "session-1" {
		t.Fatalf("unexpected request: %#v", req)
	}
	if len(req.Messages) != 1 || req.Messages[0].Content != "你现在看到了什么？" {
		t.Fatalf("unexpected messages: %#v", req.Messages)
	}
	if req.Context == nil || req.Context.MaxAnalysisRounds != 4 || req.Context.PipelineMode != sdk.PipelineModePolling {
		t.Fatalf("unexpected context: %#v", req.Context)
	}
	if len(req.Context.DynamicInterfaces) != 1 || req.Context.DynamicInterfaces[0].ID != "scene_facts" {
		t.Fatalf("unexpected dynamic interfaces: %#v", req.Context.DynamicInterfaces)
	}
}
