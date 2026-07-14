package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage runtime external interaction tasks",
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List runtime tasks",
	Run: func(cmd *cobra.Command, args []string) {
		category, _ := cmd.Flags().GetString("category")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		tasks, err := newClient().ListRuntimeTasks(category, status, limit)
		if err != nil {
			fail(err)
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
			return
		}
		for _, t := range tasks {
			fmt.Printf("[%s] %s status=%s category=%s attempts=%d\n",
				shortID(t.TaskID), t.TaskID, t.Status, t.Category, t.AttemptCount)
		}
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get runtime task details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task, err := newClient().GetRuntimeTask(args[0])
		if err != nil {
			fail(err)
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			printJSON(task)
			return
		}
		printRuntimeTaskSummary(task)
	},
}

var taskInspectCmd = &cobra.Command{
	Use:   "inspect <task-id>",
	Short: "Inspect callback/resume related runtime task details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task, err := newClient().GetRuntimeTask(args[0])
		if err != nil {
			fail(err)
		}
		printRuntimeTaskInspection(task)
	},
}

func printRuntimeTaskSummary(task *sdk.RuntimeTask) {
	if task == nil {
		fmt.Println("No task found.")
		return
	}
	lines := []string{
		fmt.Sprintf("Task: %s", task.TaskID),
		fmt.Sprintf("  status=%s category=%s interface=%s", firstNonEmpty(task.Status, "-"), firstNonEmpty(task.Category, "-"), firstNonEmpty(task.InterfaceName, "-")),
		fmt.Sprintf("  delivery=%s consumer=%s transport=%s", firstNonEmpty(task.DeliveryMode, "-"), firstNonEmpty(task.Consumer, "-"), firstNonEmpty(task.Transport, "-")),
		fmt.Sprintf("  callback=%s resume_execution=%s", firstNonEmpty(task.CallbackID, "-"), firstNonEmpty(task.ResumeExecutionID, "-")),
		fmt.Sprintf("  attempts=%d/%d dispatch_attempts=%d heartbeat_timeouts=%d", task.AttemptCount, task.MaxAttempts, task.DispatchAttempts, task.HeartbeatTimeoutCount),
	}
	if task.ErrorMessage != "" {
		lines = append(lines, "  error="+task.ErrorMessage)
	}
	if task.ResultJSON != "" {
		lines = append(lines, "  result="+compactText(task.ResultJSON, 200))
	}
	for _, line := range lines {
		fmt.Println(line)
	}
}

func printRuntimeTaskInspection(task *sdk.RuntimeTask) {
	if task == nil {
		fmt.Println("No task found.")
		return
	}
	printRuntimeTaskSummary(task)
	if task.PayloadJSON != "" {
		fmt.Println("  payload=" + compactText(task.PayloadJSON, 240))
	}
	if task.LastDispatchError != "" || task.LastDispatchDecision != "" || task.LastDispatchFailureClass != "" {
		fmt.Printf("  dispatch_decision=%s failure_class=%s transition=%s\n",
			firstNonEmpty(task.LastDispatchDecision, "-"),
			firstNonEmpty(task.LastDispatchFailureClass, "-"),
			firstNonEmpty(task.LastTransitionReason, "-"),
		)
	}
	if task.DispatchedAt != "" || task.ClaimedAt != "" || task.LastHeartbeatAt != "" || task.CompletedAt != "" {
		fmt.Printf("  dispatched_at=%s claimed_at=%s last_heartbeat_at=%s completed_at=%s\n",
			firstNonEmpty(task.DispatchedAt, "-"),
			firstNonEmpty(task.ClaimedAt, "-"),
			firstNonEmpty(task.LastHeartbeatAt, "-"),
			firstNonEmpty(task.CompletedAt, "-"),
		)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func compactText(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if limit <= 0 || len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit-3] + "..."
}

var taskClaimCmd = &cobra.Command{
	Use:   "claim <task-id>",
	Short: "Claim a pending runtime task",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		consumer, _ := cmd.Flags().GetString("consumer")
		leaseOwner, _ := cmd.Flags().GetString("owner")
		task, err := newClient().ClaimRuntimeTask(args[0], consumer, leaseOwner)
		if err != nil {
			fail(err)
		}
		fmt.Printf("Claimed task %s (lease_token=%s)\n", shortID(task.TaskID), shortID(task.LeaseToken))
	},
}

var taskStartCmd = &cobra.Command{
	Use:   "start <task-id> <lease-token>",
	Short: "Start executing a claimed runtime task",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		task, err := newClient().StartRuntimeTask(args[0], args[1])
		if err != nil {
			fail(err)
		}
		fmt.Printf("Started task %s\n", shortID(task.TaskID))
	},
}

var taskHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat <task-id> <lease-token>",
	Short: "Send heartbeat for a running runtime task",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().HeartbeatRuntimeTask(args[0], args[1]); err != nil {
			fail(err)
		}
		fmt.Println("Heartbeat sent.")
	},
}

var taskReleaseCmd = &cobra.Command{
	Use:   "release <task-id> <lease-token>",
	Short: "Release a claimed or running runtime task",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		reason, _ := cmd.Flags().GetString("reason")
		if err := newClient().ReleaseRuntimeTask(args[0], args[1], reason); err != nil {
			fail(err)
		}
		fmt.Println("Task released.")
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd, taskGetCmd, taskInspectCmd, taskClaimCmd, taskStartCmd, taskHeartbeatCmd, taskReleaseCmd)
	taskListCmd.Flags().String("category", "", "Filter by category")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().Int("limit", 50, "Max results")
	taskGetCmd.Flags().Bool("json", false, "Print raw JSON instead of a summary")
	taskClaimCmd.Flags().String("consumer", "", "Consumer identifier")
	taskClaimCmd.Flags().String("owner", "devcli", "Lease owner identifier")
	taskReleaseCmd.Flags().String("reason", "manual release", "Release reason")
}
