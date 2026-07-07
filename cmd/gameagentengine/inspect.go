package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Open GameAgentCreator",
	Run: func(cmd *cobra.Command, args []string) {
		for _, p := range []string{"web/GameAgentCreator/index.html", "tools/source/web/GameAgentCreator/index.html"} {
			if _, err := os.Stat(p); err == nil {
				abs, _ := filepath.Abs(p)
				fmt.Println("Opening:", abs)
				switch runtime.GOOS {
				case "windows":
					exec.Command("cmd", "/c", "start", abs).Start()
				case "darwin":
					exec.Command("open", abs).Start()
				default:
					exec.Command("xdg-open", abs).Start()
				}
				return
			}
		}
		fmt.Println("GameAgentCreator not found")
	},
}
