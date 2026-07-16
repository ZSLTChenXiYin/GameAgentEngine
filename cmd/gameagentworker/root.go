package main

import (
	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workercli"
)

var workerRunner = workercli.New(workercli.Options{
	CommandName:       "gameagentworker",
	DisplayName:       "GameAgentWorker",
	ShortDescription:  "Deterministic external worker for GameAgentEngine integration and play scenarios",
	DefaultLeaseOwner: "gameagentworker",
	WorkerID:          "gameagentworker",
})

var rootCmd = &cobra.Command{
	Use:   "GameAgentWorker",
	Short: "Deterministic external worker for GameAgentEngine integration and play scenarios",
}

func init() {
	workerRunner.BindCommonFlags(rootCmd.PersistentFlags())
	rootCmd.AddCommand(serveCmd, pushCmd, pullCmd, pullOnceCmd, playCmd, testCmd)
}
