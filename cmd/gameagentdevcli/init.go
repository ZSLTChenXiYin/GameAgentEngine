package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultInitConfig = `server:
  host: "0.0.0.0"
  port: 8080
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
  # driver supports sqlite / mysql / postgres
auth:
  api_key: "dev-key"
`

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "?????????",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "my-agent"
		if len(args) > 0 {
			name = args[0]
		}
		dir := name
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fail(err)
		}
		confPath := filepath.Join(dir, "gameagentengine.conf.yaml")
		if err := os.WriteFile(confPath, []byte(defaultInitConfig), 0o644); err != nil {
			fail(err)
		}
		fmt.Printf("Project %q initialized\n", name)
		fmt.Printf("Created %s\n", confPath)
	},
}
