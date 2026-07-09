package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node instances",
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List nodes",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		nodeType, _ := cmd.Flags().GetString("type")
		nodes, err := newClient().GetNodes(worldID, limit, offset, nodeType)
		if err != nil {
			fail(err)
		}
		printJSON(nodes)
	},
}

var nodeLegacyListCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List nodes",
	Run:   nodeListCmd.Run,
}

var nodeGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get node details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detail, err := newClient().GetNode(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(detail)
	},
}

var nodeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a node",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		name, _ := cmd.Flags().GetString("name")
		nodeType, _ := cmd.Flags().GetString("type")
		parentID, _ := cmd.Flags().GetString("parent")

		if name == "" || nodeType == "" {
			fail(fmt.Errorf("--name and --type are required"))
		}
		id, err := newClient().CreateNode(worldID, name, nodeType, parentID)
		if err != nil {
			fail(err)
		}
		detail, err := newClient().GetNode(id)
		if err != nil {
			fail(err)
		}
		printJSON(detail.Node)
	},
}

var nodeUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, nameChanged := getChangedString(cmd, "name")
		nodeType, typeChanged := getChangedString(cmd, "type")
		parentValue, parentChanged := getChangedString(cmd, "parent")
		clearParent, _ := cmd.Flags().GetBool("clear-parent")

		if !nameChanged && !typeChanged && !parentChanged && !clearParent {
			fail(fmt.Errorf("no update flags provided"))
		}

		var parentID *string
		if clearParent {
			empty := ""
			parentID = &empty
		} else if parentChanged {
			parentID = &parentValue
		}

		node, err := newClient().UpdateNode(args[0], valueIfChanged(name, nameChanged), valueIfChanged(nodeType, typeChanged), parentID)
		if err != nil {
			fail(err)
		}
		printJSON(node)
	},
}

var nodeDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().DeleteNode(args[0]); err != nil {
			fail(err)
		}
		printJSON(map[string]string{"status": "deleted", "id": args[0]})
	},
}

var nodeCopyCmd = &cobra.Command{
	Use:   "copy <id>",
	Short: "Copy a node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		parentValue, parentChanged := getChangedString(cmd, "parent")
		withChildren, _ := cmd.Flags().GetBool("with-children")

		var parentID *string
		if parentChanged {
			parentID = &parentValue
		}

		node, err := newClient().CopyNode(args[0], name, parentID, withChildren)
		if err != nil {
			fail(err)
		}
		printJSON(node)
	},
}

var nodeAutonomousCmd = &cobra.Command{
	Use:   "autonomous",
	Short: "Manage autonomous node configuration",
}

var nodeAutonomousGetCmd = &cobra.Command{
	Use:   "get <node-id>",
	Short: "Get autonomous node configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().GetAutonomousConfig(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousSetCmd = &cobra.Command{
	Use:   "set <node-id>",
	Short: "Create or update autonomous configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		enabled, _ := cmd.Flags().GetBool("enabled")
		trigger, _ := cmd.Flags().GetString("trigger")
		interval, _ := cmd.Flags().GetInt("interval")
		capabilitiesJSON, _ := cmd.Flags().GetString("capabilities")

		var capabilities []sdk.AgentCapability
		if capabilitiesJSON != "" {
			if err := json.Unmarshal([]byte(capabilitiesJSON), &capabilities); err != nil {
				fail(fmt.Errorf("invalid --capabilities json: %w", err))
			}
		}

		resp, err := newClient().SetAutonomousConfig(args[0], &sdk.AutonomousConfig{
			Enabled:         enabled,
			Trigger:         trigger,
			IntervalSeconds: interval,
			Capabilities:    capabilities,
		})
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousDisableCmd = &cobra.Command{
	Use:   "disable <node-id>",
	Short: "Disable autonomous behavior",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		current, _ := newClient().GetAutonomousConfig(args[0])
		cfg := &sdk.AutonomousConfig{Enabled: false, Trigger: "manual"}
		if current != nil && current.Config != nil {
			cfg = current.Config
			cfg.Enabled = false
		}
		resp, err := newClient().SetAutonomousConfig(args[0], cfg)
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousRunCmd = &cobra.Command{
	Use:   "run <node-id>",
	Short: "Run autonomous behavior",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		if worldID == "" {
			detail, err := newClient().GetNode(args[0])
			if err != nil {
				fail(err)
			}
			worldID = detail.Node.WorldID
		}
		resp, err := newClient().RunAutonomousNode(worldID, args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

func init() {
	nodeCmd.AddCommand(nodeListCmd, nodeGetCmd, nodeCreateCmd, nodeUpdateCmd, nodeDeleteCmd, nodeCopyCmd, nodeAutonomousCmd)
	nodeAutonomousCmd.AddCommand(nodeAutonomousGetCmd, nodeAutonomousSetCmd, nodeAutonomousDisableCmd, nodeAutonomousRunCmd)

	nodeListCmd.Flags().String("world", "", "World ID")
	nodeListCmd.Flags().Int("limit", 0, "Maximum number of nodes to return")
	nodeListCmd.Flags().Int("offset", 0, "List offset")
	nodeListCmd.Flags().String("type", "", "Filter by node type")

	nodeCreateCmd.Flags().String("world", "", "World ID; required for non-world nodes")
	nodeCreateCmd.Flags().String("name", "", "Node name")
	nodeCreateCmd.Flags().String("type", "", "Node type")
	nodeCreateCmd.Flags().String("parent", "", "Primary parent node ID; use this for stable identity/ownership, not current location")

	nodeUpdateCmd.Flags().String("name", "", "New node name")
	nodeUpdateCmd.Flags().String("type", "", "New node type")
	nodeUpdateCmd.Flags().String("parent", "", "New primary parent node ID; do not use parent to model temporary location")
	nodeUpdateCmd.Flags().Bool("clear-parent", false, "Clear the parent node")

	nodeCopyCmd.Flags().String("name", "", "Copied node name")
	nodeCopyCmd.Flags().String("parent", "", "Copied primary parent node ID; defaults to the original stable parent")
	nodeCopyCmd.Flags().Bool("with-children", true, "Copy the whole subtree")

	nodeAutonomousSetCmd.Flags().Bool("enabled", false, "Enable autonomous behavior")
	nodeAutonomousSetCmd.Flags().String("trigger", "manual", "Trigger mode: manual / world_tick_sync / scheduled")
	nodeAutonomousSetCmd.Flags().Int("interval", 0, "Scheduled interval in seconds")
	nodeAutonomousSetCmd.Flags().String("capabilities", "[]", "Capabilities whitelist as a JSON array")
	nodeAutonomousRunCmd.Flags().String("world", "", "World ID; auto-detected from node when omitted")
}
