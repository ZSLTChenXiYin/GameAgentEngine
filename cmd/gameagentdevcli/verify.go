package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
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

var verifyDemoCmd = &cobra.Command{
	Use:   "demo",
	Short: "导入 Demo 世界并执行运行时行为验证（需要本地引擎与服务器）",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initLocalStore(); err != nil {
			fail(err)
		}
		if err := store.ResetAll(); err != nil {
			fail(fmt.Errorf("reset error: %w", err))
		}

		path, err := findDemoWorldFile()
		if err != nil {
			fail(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fail(fmt.Errorf("read demo world: %w", err))
		}
		client := newClient()

		result, err := client.CreatorImport("yaml", string(data), false, false)
		if err != nil {
			fail(fmt.Errorf("import demo world: %w", err))
		}
		fmt.Printf("Demo world imported: %d nodes\n", result.NodeCount)

		worlds, err := client.GetWorlds()
		if err != nil {
			fail(fmt.Errorf("get worlds: %w", err))
		}
		if len(worlds) == 0 {
			fail(fmt.Errorf("no worlds found after demo import"))
		}
		worldID := worlds[0].ID
		fmt.Printf("World ID: %s\n", worldID)

		tr, err := client.AdvanceTick(worldID, "manual", "verification-day-1")
		if err != nil {
			fail(fmt.Errorf("tick failed: %w", err))
		}
		fmt.Printf("Tick advanced: #%d\n", tr.Tick.TickNumber)

		snapshot, err := buildWorldSnapshot(worldID)
		if err != nil {
			fail(fmt.Errorf("snapshot failed: %w", err))
		}
		fmt.Printf("Snapshot: %d nodes, %d relations\n", len(snapshot.Nodes), len(snapshot.Relations))
		fmt.Println("Demo verification passed")
	},
}

func init() {
	verifyCmd.AddCommand(verifyImportCmd, verifyDemoCmd)
}

func findDemoWorldFile() (string, error) {
	candidates := make([]string, 0, 4)
	if localConfigPath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(localConfigPath), "demo-world.yaml"))
	}
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "demo-world.yaml"))
	}
	candidates = append(candidates,
		"demo-world.yaml",
		filepath.Join("tools", "source", "demo-world.yaml"),
		filepath.Join("docs", "demo-world.yaml"),
	)
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("demo world file not found; checked: %s", strings.Join(candidates, ", "))
}
