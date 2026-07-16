package main

import "github.com/spf13/cobra"

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Run a single-user text-game REPL backed by Engine invoke",
	RunE: func(cmd *cobra.Command, args []string) error {
		return workerRunner.RunPlay()
	},
}

func init() {
	workerRunner.BindPlayFlags(playCmd.Flags())
}
