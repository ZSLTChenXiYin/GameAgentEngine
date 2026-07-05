package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

// testScenario 描述一个完整的测试场景。
type testScenario struct {
	Name  string     `yaml:"name" json:"name"`
	Steps []testStep `yaml:"steps" json:"steps"`
}

// testStep 描述测试场景中的单个步骤。
type testStep struct {
	Name   string     `yaml:"name" json:"name"`
	Invoke testInvoke `yaml:"invoke" json:"invoke"`
	Expect testExpect `yaml:"expect" json:"expect"`
}

// testInvoke 描述单步测试要调用的推理请求。
type testInvoke struct {
	WorldID  string            `yaml:"world_id" json:"world_id"`
	TaskType string            `yaml:"task_type" json:"task_type"`
	NodeID   string            `yaml:"node_id" json:"node_id"`
	Messages []sdk.ChatMessage `yaml:"messages" json:"messages,omitempty"`
}

// testExpect 描述单步测试的断言条件。
type testExpect struct {
	ReplyContains      []string `yaml:"reply_contains" json:"reply_contains,omitempty"`
	ReplyNotContains   []string `yaml:"reply_not_contains" json:"reply_not_contains,omitempty"`
	ActionCalls        []string `yaml:"action_calls" json:"action_calls,omitempty"`
	ActionCallsCount   string   `yaml:"action_calls_count" json:"action_calls_count,omitempty"`
	MemoryUpdatesCount string   `yaml:"memory_updates_count" json:"memory_updates_count,omitempty"`
	WorldChangePlan    *bool    `yaml:"world_change_plan" json:"world_change_plan,omitempty"`
}

var testServerURL string
var testAPIKey string

var testCmd = &cobra.Command{
	Use:   "test <scenario-file>",
	Short: "Run test scenarios against a running Agent server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := sdk.NewClient(testServerURL, testAPIKey)

		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		var sc testScenario
		if err := yaml.Unmarshal(data, &sc); err != nil {
			if err2 := json.Unmarshal(data, &sc); err2 != nil {
				fmt.Fprintf(os.Stderr, "Parse error: YAML: %v\nJSON: %v\n", err, err2)
				os.Exit(1)
			}
		}

		fmt.Printf("Test: %s (%d steps)\n", sc.Name, len(sc.Steps))
		fmt.Println(strings.Repeat("-", 50))

		resolveNames(client, &sc)

		passed, failed := 0, 0
		for i, step := range sc.Steps {
			if runStep(client, i+1, step) {
				passed++
			} else {
				failed++
			}
		}

		fmt.Println(strings.Repeat("-", 50))
		if failed == 0 {
			fmt.Printf("OK: %d/%d passed\n", passed, passed+failed)
		} else {
			fmt.Printf("FAIL: %d/%d passed, %d failed\n", passed, passed+failed, failed)
			os.Exit(1)
		}
	},
}

// init 注册 test 子命令及其参数。
func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&testServerURL, "server", "s", "http://127.0.0.1:8080", "Agent server URL")
	testCmd.Flags().StringVarP(&testAPIKey, "key", "k", "dev-key", "API key")
}

// resolveNames 将场景中使用的名字解析成真实节点 ID。
func resolveNames(client *sdk.Client, sc *testScenario) {
	nodes, err := client.GetNodes("", 0, 0, "")
	if err != nil {
		return
	}
	nameMap := make(map[string]string)
	for _, n := range nodes {
		nameMap[n.Name] = n.ID
	}
	worlds, _ := client.GetWorlds()
	for _, w := range worlds {
		nameMap[w.Name] = w.ID
	}
	if len(worlds) > 0 {
		nameMap["world"] = worlds[0].ID
	}
	for i := range sc.Steps {
		s := &sc.Steps[i]
		if id, ok := nameMap[s.Invoke.NodeID]; ok {
			s.Invoke.NodeID = id
		}
		if s.Invoke.WorldID == "" && nameMap["world"] != "" {
			s.Invoke.WorldID = nameMap["world"]
		}
		if id, ok := nameMap[s.Invoke.WorldID]; ok {
			s.Invoke.WorldID = id
		}
	}
}

// checkCount 根据比较表达式判断数量是否满足预期。
func checkCount(expect string, actual int) bool {
	r := regexp.MustCompile(`^(>=|<=|>|<|=)?(\d+)$`)
	m := r.FindStringSubmatch(expect)
	if m == nil {
		return false
	}
	num, _ := strconv.Atoi(m[2])
	switch m[1] {
	case ">=":
		return actual >= num
	case "<=":
		return actual <= num
	case ">":
		return actual > num
	case "<":
		return actual < num
	default:
		return actual == num
	}
}

// runStep 执行单个测试步骤并输出断言结果。
func runStep(client *sdk.Client, num int, step testStep) bool {
	fmt.Printf("  Step %d: %s\n", num, step.Name)

	resp, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  step.Invoke.WorldID,
		TaskType: step.Invoke.TaskType,
		NodeID:   step.Invoke.NodeID,
		Messages: step.Invoke.Messages,
	})
	if err != nil {
		fmt.Printf("    FAIL: invoke error: %v\n", err)
		return false
	}

	allPassed := true

	for _, expected := range step.Expect.ReplyContains {
		passed := strings.Contains(resp.Reply, expected)
		if !passed {
			allPassed = false
		}
		s := "PASS"
		if !passed {
			s = "FAIL"
		}
		fmt.Printf("    %s reply contains %q (got: %s)\n", s, expected, truncate(resp.Reply, 60))
	}

	for _, forbidden := range step.Expect.ReplyNotContains {
		passed := !strings.Contains(resp.Reply, forbidden)
		if !passed {
			allPassed = false
		}
		s := "PASS"
		if !passed {
			s = "FAIL"
		}
		fmt.Printf("    %s reply NOT contains %q\n", s, forbidden)
	}

	for _, expected := range step.Expect.ActionCalls {
		found := false
		for _, ac := range resp.ActionCalls {
			if ac.ActionID == expected {
				found = true
				break
			}
		}
		if !found {
			allPassed = false
		}
		s := "PASS"
		if !found {
			s = "FAIL"
		}
		fmt.Printf("    %s action_call %s found\n", s, expected)
	}

	if step.Expect.ActionCallsCount != "" {
		passed := checkCount(step.Expect.ActionCallsCount, len(resp.ActionCalls))
		if !passed {
			allPassed = false
		}
		s := "PASS"
		if !passed {
			s = "FAIL"
		}
		fmt.Printf("    %s action_calls_count %s (got %d)\n", s, step.Expect.ActionCallsCount, len(resp.ActionCalls))
	}

	if step.Expect.MemoryUpdatesCount != "" {
		passed := checkCount(step.Expect.MemoryUpdatesCount, len(resp.MemoryUpdates))
		if !passed {
			allPassed = false
		}
		s := "PASS"
		if !passed {
			s = "FAIL"
		}
		fmt.Printf("    %s memory_updates_count %s (got %d)\n", s, step.Expect.MemoryUpdatesCount, len(resp.MemoryUpdates))
	}

	if step.Expect.WorldChangePlan != nil {
		hasPlan := resp.WorldChangePlan != nil
		passed := hasPlan == *step.Expect.WorldChangePlan
		if !passed {
			allPassed = false
		}
		s := "PASS"
		if !passed {
			s = "FAIL"
		}
		fmt.Printf("    %s world_change_plan=%v (hasPlan=%v)\n", s, *step.Expect.WorldChangePlan, hasPlan)
	}

	if len(step.Expect.ReplyContains) == 0 && len(step.Expect.ReplyNotContains) == 0 &&
		len(step.Expect.ActionCalls) == 0 && step.Expect.ActionCallsCount == "" &&
		step.Expect.MemoryUpdatesCount == "" && step.Expect.WorldChangePlan == nil {
		fmt.Printf("    (no assertions, reply: %s)\n", truncate(resp.Reply, 60))
	}

	return allPassed
}

// truncate 按字符数截断字符串，便于终端展示。
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
