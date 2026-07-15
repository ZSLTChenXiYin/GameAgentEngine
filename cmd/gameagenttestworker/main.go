package main

import "github.com/ZSLTChenXiYin/GameAgentEngine/internal/workercli"

func main() {
	workercli.Main(workercli.Options{
		CommandName:       "gameagentworker",
		DisplayName:       "GameAgentTestWorker",
		ShortDescription:  "Deprecated compatibility entry for the deterministic GameAgent worker",
		DefaultLeaseOwner: "gameagenttestworker",
		WorkerID:          "gameagenttestworker",
		DeprecatedAlias:   "gameagenttestworker",
	})
}
