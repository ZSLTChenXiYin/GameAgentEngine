package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "初始化开发项目目录",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "my-agent"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Printf("Project %q initialized\n", name)
	},
}
