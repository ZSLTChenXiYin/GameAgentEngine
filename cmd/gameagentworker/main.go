package main

import "github.com/ZSLTChenXiYin/GameAgentEngine/internal/workercli"

func main() {
	workercli.Main(workercli.Options{
		CommandName:       "gameagentworker",
		DisplayName:       "GameAgentWorker",
		ShortDescription:  "Deterministic external worker for GameAgentEngine integration and play scenarios",
		DefaultLeaseOwner: "gameagentworker",
		WorkerID:          "gameagentworker",
	})
}
