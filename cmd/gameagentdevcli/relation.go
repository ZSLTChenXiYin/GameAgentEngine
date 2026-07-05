package main

import (
    "fmt"

    "github.com/spf13/cobra"
)

var relationCmd = &cobra.Command{
    Use:   "relation",
    Short: "管理关系实例",
}

var relationListCmd = &cobra.Command{
    Use:   "list",
    Short: "列出关系",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		relationType, _ := cmd.Flags().GetString("type")
		items, err := newClient().GetRelations(worldID, limit, offset, relationType)
		if err != nil {
			fail(err)
		}
        printJSON(items)
    },
}

var relationGetCmd = &cobra.Command{
    Use:   "get <id>",
    Short: "获取关系详情",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        item, err := newClient().GetRelation(args[0])
        if err != nil {
            fail(err)
        }
        printJSON(item)
    },
}

var relationCreateCmd = &cobra.Command{
    Use:   "create",
    Short: "创建关系",
    Run: func(cmd *cobra.Command, args []string) {
        worldID, _ := cmd.Flags().GetString("world")
        sourceID, _ := cmd.Flags().GetString("source")
        targetID, _ := cmd.Flags().GetString("target")
        relationType, _ := cmd.Flags().GetString("type")
        properties, _ := cmd.Flags().GetString("props")
        weight, _ := cmd.Flags().GetInt("weight")

        if worldID == "" || sourceID == "" || targetID == "" || relationType == "" {
            fail(fmt.Errorf("--world, --source, --target and --type are required"))
        }

        id, err := newClient().CreateRelationWithProps(worldID, sourceID, targetID, relationType, weight, properties)
        if err != nil {
            fail(err)
        }
        item, err := newClient().GetRelation(id)
        if err != nil {
            fail(err)
        }
        printJSON(item)
    },
}

var relationUpdateCmd = &cobra.Command{
    Use:   "update <id>",
    Short: "更新关系",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        sourceID, sourceChanged := getChangedString(cmd, "source")
        targetID, targetChanged := getChangedString(cmd, "target")
        relationType, typeChanged := getChangedString(cmd, "type")
        properties, propsChanged := getChangedString(cmd, "props")
        clearProps, _ := cmd.Flags().GetBool("clear-props")

        var weightPtr *int
        if cmd.Flags().Changed("weight") {
            weight, _ := cmd.Flags().GetInt("weight")
            weightPtr = &weight
        }

        if !sourceChanged && !targetChanged && !typeChanged && !propsChanged && !clearProps && weightPtr == nil {
            fail(fmt.Errorf("no update flags provided"))
        }

        var propsPtr *string
        if clearProps {
            empty := ""
            propsPtr = &empty
        } else {
            propsPtr = ptrIfChanged(properties, propsChanged)
        }

        item, err := newClient().UpdateRelation(
            args[0],
            ptrIfChanged(sourceID, sourceChanged),
            ptrIfChanged(targetID, targetChanged),
            ptrIfChanged(relationType, typeChanged),
            propsPtr,
            weightPtr,
        )
        if err != nil {
            fail(err)
        }
        printJSON(item)
    },
}

var relationDeleteCmd = &cobra.Command{
    Use:   "delete <id>",
    Short: "删除关系",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        if err := newClient().DeleteRelation(args[0]); err != nil {
            fail(err)
        }
        printJSON(map[string]string{"status": "deleted", "id": args[0]})
    },
}

func init() {
    relationCmd.AddCommand(relationListCmd, relationGetCmd, relationCreateCmd, relationUpdateCmd, relationDeleteCmd)

	relationListCmd.Flags().String("world", "", "世界 ID")
	relationListCmd.Flags().Int("limit", 0, "返回条数上限（0 为不限制）")
	relationListCmd.Flags().Int("offset", 0, "偏移量")
	relationListCmd.Flags().String("type", "", "按关系类型过滤")
	relationCreateCmd.Flags().String("world", "", "世界 ID")
    relationCreateCmd.Flags().String("source", "", "源节点 ID")
    relationCreateCmd.Flags().String("target", "", "目标节点 ID")
    relationCreateCmd.Flags().String("type", "", "关系类型")
    relationCreateCmd.Flags().Int("weight", 0, "关系权重")
    relationCreateCmd.Flags().String("props", "", "关系属性，建议传 JSON 字符串")
    relationUpdateCmd.Flags().String("source", "", "新的源节点 ID")
    relationUpdateCmd.Flags().String("target", "", "新的目标节点 ID")
    relationUpdateCmd.Flags().String("type", "", "新的关系类型")
    relationUpdateCmd.Flags().Int("weight", 0, "新的关系权重")
    relationUpdateCmd.Flags().String("props", "", "新的关系属性")
    relationUpdateCmd.Flags().Bool("clear-props", false, "清空关系属性")
}
