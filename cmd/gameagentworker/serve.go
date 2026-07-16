package main

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run both push receiver and pull worker loops",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerRunner.RunServe(true, true)
	},
}

var pushCmd = &cobra.Command{
	Use:   "push-receiver",
	Short: "Run only the push receiver",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerRunner.RunServe(true, false)
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull-worker",
	Short: "Run only the pull worker loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerRunner.RunServe(false, true)
	},
}

var pullOnceCmd = &cobra.Command{
	Use:   "pull-once",
	Short: "Claim, execute, and callback one pull task if present",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerRunner.RunPullOnce()
	},
}
