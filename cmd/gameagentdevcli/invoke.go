package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke <world-id> <node-id>",
	Short: "Invoke one engine reasoning request",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		req, err := buildInvokeRequestFromFlags(cmd, args[0], args[1])
		if err != nil {
			fail(err)
		}
		resp, err := newClient().Invoke(req)
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

func buildInvokeRequestFromFlags(cmd *cobra.Command, worldID, nodeID string) (*sdk.InvokeRequest, error) {
	taskType, _ := cmd.Flags().GetString("task-type")
	if strings.TrimSpace(taskType) == "" {
		return nil, fmt.Errorf("--task-type is required")
	}
	message, _ := cmd.Flags().GetString("message")
	sessionID, _ := cmd.Flags().GetString("session-id")
	req := &sdk.InvokeRequest{
		WorldID:   worldID,
		NodeID:    nodeID,
		TaskType:  taskType,
		SessionID: sessionID,
		Context:   buildInvokeContext(cmd),
	}
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		req.Messages = []sdk.ChatMessage{{Role: "user", Content: trimmed}}
	}
	return req, nil
}

func init() {
	invokeCmd.Flags().String("task-type", "npc_dialogue", "Reasoning task type, such as npc_dialogue or custom")
	invokeCmd.Flags().String("message", "", "Single user message for the invoke request")
	invokeCmd.Flags().String("session-id", "", "Optional conversation session ID")
}
