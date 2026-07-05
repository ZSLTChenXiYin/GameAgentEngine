package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "管理记忆实例",
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出节点记忆",
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("node")
		if nodeID == "" {
			fail(fmt.Errorf("--node is required"))
		}
		items, err := newClient().GetMemories(nodeID)
		if err != nil {
			fail(err)
		}
		printJSON(items)
	},
}

var memoryGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "获取记忆详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		item, err := newClient().GetMemory(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var memoryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建记忆",
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("node")
		content, _ := cmd.Flags().GetString("content")
		level, _ := cmd.Flags().GetString("level")
		tags, _ := cmd.Flags().GetString("tags")
		if nodeID == "" || content == "" {
			fail(fmt.Errorf("--node and --content are required"))
		}
		id, err := newClient().AddMemory(nodeID, content, level, tags)
		if err != nil {
			fail(err)
		}
		item, err := newClient().GetMemory(id)
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var memoryUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "更新记忆",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content, contentChanged := getChangedString(cmd, "content")
		level, levelChanged := getChangedString(cmd, "level")
		tags, tagsChanged := getChangedString(cmd, "tags")
		clearTags, _ := cmd.Flags().GetBool("clear-tags")

		if !contentChanged && !levelChanged && !tagsChanged && !clearTags {
			fail(fmt.Errorf("no update flags provided"))
		}

		var tagsPtr *string
		if clearTags {
			empty := ""
			tagsPtr = &empty
		} else {
			tagsPtr = ptrIfChanged(tags, tagsChanged)
		}

		item, err := newClient().UpdateMemory(args[0], ptrIfChanged(content, contentChanged), ptrIfChanged(level, levelChanged), tagsPtr)
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var memoryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "删除记忆",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().DeleteMemory(args[0]); err != nil {
			fail(err)
		}
		printJSON(map[string]string{"status": "deleted", "id": args[0]})
	},
}

func init() {
	memoryCmd.AddCommand(memoryListCmd, memoryGetCmd, memoryCreateCmd, memoryUpdateCmd, memoryDeleteCmd)

	memoryListCmd.Flags().String("node", "", "节点 ID")
	memoryCreateCmd.Flags().String("node", "", "节点 ID")
	memoryCreateCmd.Flags().String("content", "", "记忆内容")
	memoryCreateCmd.Flags().String("level", "long_term", "记忆层级")
	memoryCreateCmd.Flags().String("tags", "", "记忆标签，建议逗号分隔")
	memoryUpdateCmd.Flags().String("content", "", "新的记忆内容")
	memoryUpdateCmd.Flags().String("level", "", "新的记忆层级")
	memoryUpdateCmd.Flags().String("tags", "", "新的记忆标签")
	memoryUpdateCmd.Flags().Bool("clear-tags", false, "清空记忆标签")
}
