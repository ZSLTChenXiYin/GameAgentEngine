package workertest

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type DevCLI struct {
	Executable string
	Server     string
	APIKey     string
}

func (c DevCLI) Run(args ...string) ([]byte, error) {
	fullArgs := append([]string{"--server", c.Server, "--key", c.APIKey}, args...)
	cmd := exec.Command(c.Executable, fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("devcli failed: %s\n%s", strings.Join(fullArgs, " "), strings.TrimSpace(string(out)))
	}
	return out, nil
}

func (c DevCLI) RunJSON(out any, args ...string) error {
	raw, err := c.Run(args...)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || out == nil {
		return nil
	}
	return json.Unmarshal([]byte(trimmed), out)
}
