package main

import (
	"fmt"

	"github.com/spf13/cobra"
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
		printJSON(task)
	},
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
	taskCmd.AddCommand(taskListCmd, taskGetCmd, taskClaimCmd, taskStartCmd, taskHeartbeatCmd, taskReleaseCmd)
	taskListCmd.Flags().String("category", "", "Filter by category")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().Int("limit", 50, "Max results")
	taskClaimCmd.Flags().String("consumer", "", "Consumer identifier")
	taskClaimCmd.Flags().String("owner", "devcli", "Lease owner identifier")
	taskReleaseCmd.Flags().String("reason", "manual release", "Release reason")
}
