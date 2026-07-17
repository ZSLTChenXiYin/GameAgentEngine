package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "验证世界配置或运行时行为",
}

var verifyImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "导入配置并验证导入结果",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fail(err)
		}
		format := "yaml"
		if strings.HasSuffix(args[0], ".json") {
			format = "json"
		}

		result, err := newClient().CreatorImport(format, string(data), false, false)
		if err != nil {
			fail(fmt.Errorf("import failed: %w", err))
		}
		worlds, err := newClient().GetWorlds()
		if err != nil {
			fail(fmt.Errorf("get worlds: %w", err))
		}
		fmt.Printf("Import verified: %d nodes, %d relations\n", result.NodeCount, result.RelationCount)
		fmt.Printf("Worlds on server: %d\n", len(worlds))
	},
}

func init() {
	verifyCmd.AddCommand(verifyImportCmd)
}
