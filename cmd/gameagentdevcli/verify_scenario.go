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

type verifyScenario struct {
	Name  string               `yaml:"name" json:"name"`
	Steps []verifyScenarioStep `yaml:"steps" json:"steps"`
}

type verifyScenarioStep struct {
	Name   string               `yaml:"name" json:"name"`
	Invoke verifyScenarioInvoke `yaml:"invoke" json:"invoke"`
	Expect verifyScenarioExpect `yaml:"expect" json:"expect"`
}

type verifyScenarioInvoke struct {
	WorldID  string            `yaml:"world_id" json:"world_id"`
	TaskType string            `yaml:"task_type" json:"task_type"`
	NodeID   string            `yaml:"node_id" json:"node_id"`
	Messages []sdk.ChatMessage `yaml:"messages" json:"messages,omitempty"`
}

type verifyScenarioExpect struct {
	ReplyContains      []string `yaml:"reply_contains" json:"reply_contains,omitempty"`
	ReplyNotContains   []string `yaml:"reply_not_contains" json:"reply_not_contains,omitempty"`
	ActionCalls        []string `yaml:"action_calls" json:"action_calls,omitempty"`
	ActionCallsCount   string   `yaml:"action_calls_count" json:"action_calls_count,omitempty"`
	MemoryUpdatesCount string   `yaml:"memory_updates_count" json:"memory_updates_count,omitempty"`
	WorldChangePlan    *bool    `yaml:"world_change_plan" json:"world_change_plan,omitempty"`
}

var verifyScenarioCmd = &cobra.Command{
	Use:   "scenario <file>",
	Short: "运行基于 invoke 的场景验证文件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()

		data, err := os.ReadFile(args[0])
		if err != nil {
			fail(err)
		}
		var sc verifyScenario
		if err := yaml.Unmarshal(data, &sc); err != nil {
			if err2 := json.Unmarshal(data, &sc); err2 != nil {
				fail(fmt.Errorf("parse error: YAML: %v JSON: %v", err, err2))
			}
		}

		fmt.Printf("Scenario: %s (%d steps)\n", sc.Name, len(sc.Steps))
		fmt.Println(strings.Repeat("-", 50))

		resolveScenarioNames(client, &sc)

		passed, failed := 0, 0
		for i, step := range sc.Steps {
			if runScenarioStep(client, i+1, step) {
				passed++
			} else {
				failed++
			}
		}

		fmt.Println(strings.Repeat("-", 50))
		if failed == 0 {
			fmt.Printf("OK: %d/%d passed\n", passed, passed+failed)
			return
		}
		fail(fmt.Errorf("scenario failed: %d/%d passed, %d failed", passed, passed+failed, failed))
	},
}

func init() {
	verifyCmd.AddCommand(verifyScenarioCmd)
}

func resolveScenarioNames(client *sdk.Client, sc *verifyScenario) {
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

func scenarioCheckCount(expect string, actual int) bool {
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

func runScenarioStep(client *sdk.Client, num int, step verifyScenarioStep) bool {
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
		label := "PASS"
		if !passed {
			label = "FAIL"
		}
		fmt.Printf("    %s reply contains %q (got: %s)\n", label, expected, scenarioTruncate(resp.Reply, 60))
	}
	for _, forbidden := range step.Expect.ReplyNotContains {
		passed := !strings.Contains(resp.Reply, forbidden)
		if !passed {
			allPassed = false
		}
		label := "PASS"
		if !passed {
			label = "FAIL"
		}
		fmt.Printf("    %s reply NOT contains %q\n", label, forbidden)
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
		label := "PASS"
		if !found {
			label = "FAIL"
		}
		fmt.Printf("    %s action_call %s found\n", label, expected)
	}
	if step.Expect.ActionCallsCount != "" {
		passed := scenarioCheckCount(step.Expect.ActionCallsCount, len(resp.ActionCalls))
		if !passed {
			allPassed = false
		}
		label := "PASS"
		if !passed {
			label = "FAIL"
		}
		fmt.Printf("    %s action_calls_count %s (got %d)\n", label, step.Expect.ActionCallsCount, len(resp.ActionCalls))
	}
	if step.Expect.MemoryUpdatesCount != "" {
		passed := scenarioCheckCount(step.Expect.MemoryUpdatesCount, len(resp.MemoryUpdates))
		if !passed {
			allPassed = false
		}
		label := "PASS"
		if !passed {
			label = "FAIL"
		}
		fmt.Printf("    %s memory_updates_count %s (got %d)\n", label, step.Expect.MemoryUpdatesCount, len(resp.MemoryUpdates))
	}
	if step.Expect.WorldChangePlan != nil {
		hasPlan := resp.WorldChangePlan != nil
		passed := hasPlan == *step.Expect.WorldChangePlan
		if !passed {
			allPassed = false
		}
		label := "PASS"
		if !passed {
			label = "FAIL"
		}
		fmt.Printf("    %s world_change_plan=%v (hasPlan=%v)\n", label, *step.Expect.WorldChangePlan, hasPlan)
	}
	if len(step.Expect.ReplyContains) == 0 && len(step.Expect.ReplyNotContains) == 0 &&
		len(step.Expect.ActionCalls) == 0 && step.Expect.ActionCallsCount == "" &&
		step.Expect.MemoryUpdatesCount == "" && step.Expect.WorldChangePlan == nil {
		fmt.Printf("    (no assertions, reply: %s)\n", scenarioTruncate(resp.Reply, 60))
	}
	return allPassed
}

func scenarioTruncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
