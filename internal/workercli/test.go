package workercli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var supportedTestScenarios = []string{
	"base-data",
	"continuity",
	"runtime-tasks",
	"callback-resume",
	"tooling-smoke",
	"machine-scenario",
	"all",
}

func (a *app) newTestCommand() *cobra.Command {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run packaged worker-side full-functional scenarios",
	}
	for _, scenario := range supportedTestScenarios {
		scenario := scenario
		cmd := &cobra.Command{
			Use:   scenario,
			Short: fmt.Sprintf("Run the %s worker test scenario", scenario),
			RunE: func(cmd *cobra.Command, args []string) error {
				a.cfg.TestScenario = scenario
				return a.runNamedTestScenario(scenario)
			},
		}
		a.bindTestFlags(cmd.Flags())
		testCmd.AddCommand(cmd)
	}
	return testCmd
}

func (a *app) bindTestFlags(flags *pflag.FlagSet) {
	flags.StringVar(&a.cfg.TestEngineExePath, "engine-exe", a.cfg.TestEngineExePath, "Path to GameAgentEngine executable used by worker test scenarios")
	flags.StringVar(&a.cfg.TestDevCLIExePath, "devcli-exe", a.cfg.TestDevCLIExePath, "Path to GameAgentDevCli executable used by worker test scenarios")
	flags.StringVar(&a.cfg.TestWorkerExePath, "worker-exe", a.cfg.TestWorkerExePath, "Optional path to GameAgentWorker executable when a nested worker process is required")
	flags.StringVar(&a.cfg.TestsDir, "tests-dir", a.cfg.TestsDir, "Directory containing packaged YAML/JSON worker test fixtures")
	flags.StringVar(&a.cfg.TestOutFile, "out", a.cfg.TestOutFile, "Optional output path for the scenario result JSON")
	flags.IntVar(&a.cfg.TestEnginePort, "engine-port", a.cfg.TestEnginePort, "Engine port used by worker test scenarios")
	flags.IntVar(&a.cfg.TestPushPort, "push-port", a.cfg.TestPushPort, "Push receiver port used by worker test scenarios")
	flags.BoolVar(&a.cfg.TestKeepTemp, "keep-temp", a.cfg.TestKeepTemp, "Keep temporary files and directories created by worker test scenarios")
	flags.BoolVar(&a.cfg.TestJSON, "json", a.cfg.TestJSON, "Print scenario result as JSON")
}

func (a *app) runNamedTestScenario(scenario string) error {
	trimmed := strings.TrimSpace(scenario)
	if trimmed == "" {
		return fmt.Errorf("test scenario is required")
	}
	switch trimmed {
	case "base-data":
		return a.runBaseDataScenario()
	case "continuity":
		return a.runContinuityScenario()
	}
	return fmt.Errorf("worker test scenario %q is not implemented yet", trimmed)
}

func jsonMarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
