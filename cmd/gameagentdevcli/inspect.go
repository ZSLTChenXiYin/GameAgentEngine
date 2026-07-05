package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "打开 GameAgentCreator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Open GameAgentCreator at web/GameAgentCreator/index.html in a browser")
	},
}
