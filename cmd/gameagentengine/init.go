package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new GameAgentEngine project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "my-agent"
		if len(args) > 0 {
			name = args[0]
		}
		dir := name
		os.MkdirAll(dir, 0755)
		dfl := `server:
  host: "0.0.0.0"
  port: 8080
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
auth:
  api_key: "dev-key"
`
		os.WriteFile(filepath.Join(dir, "gameagentengine.conf.yaml"), []byte(dfl), 0644)
		fmt.Printf("Project %q initialized\n", name)
	},
}
