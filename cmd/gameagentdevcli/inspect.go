package main

import (
	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/creatorui"
)

var creatorCmd = &cobra.Command{
	Use:   "creator",
	Short: "打开 GameAgentCreator",
	Run: func(cmd *cobra.Command, args []string) {
		if err := creatorui.Open(); err != nil {
			fail(err)
		}
	},
}
