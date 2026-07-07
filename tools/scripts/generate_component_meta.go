package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

func main() {
	items := engine.ComponentMetaList()
	payload, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		panic(err)
	}
	content := "window.GAMEAGENT_COMPONENT_META = " + string(payload) + ";\n"
	target := filepath.Join("tools", "source", "web", "GameAgentCreator", "js", "component-meta.js")
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		panic(err)
	}
}
