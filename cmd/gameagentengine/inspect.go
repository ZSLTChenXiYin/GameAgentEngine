package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/creatorui"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Open GameAgentCreator",
	Run: func(cmd *cobra.Command, args []string) {
		if err := creatorui.Open(); err != nil {
			fmt.Println(err.Error())
		}
	},
}
