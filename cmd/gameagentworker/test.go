package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workercli"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run packaged worker-side full-functional scenarios",
}

func init() {
	for _, scenario := range workercli.SupportedTestScenarios() {
		scenario := scenario
		cmd := &cobra.Command{
			Use:   scenario,
			Short: fmt.Sprintf("Run the %s worker test scenario", scenario),
			RunE: func(cmd *cobra.Command, args []string) error {
				return workerRunner.RunNamedTestScenario(scenario)
			},
		}
		workerRunner.BindTestFlags(cmd.Flags())
		testCmd.AddCommand(cmd)
	}
}
