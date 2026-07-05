package main

import (
    "github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "检查服务连通性并输出基础状态",
    Run: func(cmd *cobra.Command, args []string) {
        client := newClient()
        if err := client.Health(); err != nil {
            fail(err)
        }
        worlds, err := client.GetWorlds()
        if err != nil {
            fail(err)
        }
        printJSON(map[string]any{
            "status": "ok",
            "worlds": worlds,
        })
    },
}
