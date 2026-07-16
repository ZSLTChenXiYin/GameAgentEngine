package creatorui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var CandidatePaths = []string{
	"web/GameAgentCreator/index.html",
	"tools/source/web/GameAgentCreator/index.html",
}

func Open() error {
	for _, p := range CandidatePaths {
		if _, err := os.Stat(p); err == nil {
			abs, absErr := filepath.Abs(p)
			if absErr != nil {
				return absErr
			}
			fmt.Println("Opening:", abs)
			switch runtime.GOOS {
			case "windows":
				return exec.Command("cmd", "/c", "start", abs).Start()
			case "darwin":
				return exec.Command("open", abs).Start()
			default:
				return exec.Command("xdg-open", abs).Start()
			}
		}
	}
	return fmt.Errorf("GameAgentCreator not found")
}
